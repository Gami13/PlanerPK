package calendar

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

func findEventsToDelete(currentEvents []*calendar.Event, newEvents map[string]*calendar.Event) []*calendar.Event {
	eventsToDelete := make([]*calendar.Event, 0)
	for _, event := range currentEvents {
		if _, ok := newEvents[event.ExtendedProperties.Private["hash"]]; !ok {
			eventsToDelete = append(eventsToDelete, event)
		}
	}
	return eventsToDelete
}

func calculateEventDiffs(currentEvents []*calendar.Event, newEvents map[string]*calendar.Event) ([]*calendar.Event, []*calendar.Event, []*calendar.Event) {
	eventsToDelete := findEventsToDelete(currentEvents, newEvents)
	eventsToAdd := make([]*calendar.Event, 0)
	unchangedEvents := make([]*calendar.Event, 0)

	for _, newEvent := range newEvents {
		isNew, existingEvent := checkForExistingEvent(*newEvent, currentEvents)
		if isNew {
			eventsToAdd = append(eventsToAdd, newEvent)
			continue
		}

		unchangedEvents = append(unchangedEvents, existingEvent)
	}

	return eventsToDelete, eventsToAdd, unchangedEvents
}

func buildNewEventsMap(timetable types.GroupTimetable, overrides map[string]db.Override) map[string]*calendar.Event {
	newEvents := make(map[string]*calendar.Event)
	for day, entry := range timetable {
		slog.Debug("Processing day", slog.String("day", day.String()))
		for _, class := range entry {
			ovr, hasOverride := overrides[class.Hash()]

			// No override — emit normally.
			if !hasOverride {
				event := class.ToEvent()
				newEvents[event.ExtendedProperties.Private["hash"]] = &event
				continue
			}

			// Hidden + not transparent → drop entirely.
			if ovr.IsHidden && !ovr.IsTransparent {
				slog.Debug("Dropping hidden class", slog.String("hash", class.Hash()))
				continue
			}

			// Hidden + transparent → emit only a ghost at the original position.
			if ovr.IsHidden && ovr.IsTransparent {
				slog.Debug("Adding ghost for hidden class", slog.String("hash", class.Hash()))
				ghost := buildGhostEvent(class)
				newEvents[ghost.ExtendedProperties.Private["hash"]] = &ghost
				continue
			}

			// Not hidden — emit real event, possibly time-shifted.
			realEvent := buildRealEvent(class, ovr.StartHour)
			newEvents[realEvent.ExtendedProperties.Private["hash"]] = &realEvent

			// If transparent, also emit a ghost at the original position.
			if ovr.IsTransparent && ovr.StartHour != "" {
				slog.Debug("Adding ghost for shifted class", slog.String("hash", class.Hash()))
				ghost := buildGhostEvent(class)
				newEvents[ghost.ExtendedProperties.Private["hash"]] = &ghost
			}
		}
	}
	return newEvents
}

// buildGhostEvent builds a visually muted, transparent calendar event at the
// class's original time slot. It uses Google Calendar's Graphite colour (id "8")
// so the ghost is visually distinct from real events on the calendar website.
// The Transparency field marks the slot as "Free" in availability checks.
func buildGhostEvent(class types.Class) calendar.Event {
	event := class.ToEvent()
	event.Transparency = "transparent"
	event.ColorId = "8" // Graphite — muted gray, signals a placeholder
	// Use a distinct hash so GCal treats this as a separate event.
	event.ExtendedProperties.Private["hash"] = class.Hash() + "_ghost"
	return event
}

// buildRealEvent builds the primary calendar event, optionally shifting the start time.
// startHour must be "HH:MM" or empty.
func buildRealEvent(class types.Class, startHour string) calendar.Event {
	if startHour == "" {
		return class.ToEvent()
	}

	var h, m int
	if _, err := fmt.Sscanf(startHour, "%d:%d", &h, &m); err != nil {
		slog.Error("Failed to parse start_hour override, using original time",
			slog.String("start_hour", startHour),
			slog.Any("error", err))
		return class.ToEvent()
	}

	originalStart := class.StartTime
	originalEnd := class.EndTime
	duration := originalEnd.Sub(originalStart)

	shiftedStart := time.Date(
		originalStart.Year(), originalStart.Month(), originalStart.Day(),
		h, m, 0, 0, originalStart.Location(),
	)
	shiftedEnd := shiftedStart.Add(duration)

	// Build a copy of the class with the new times.
	shiftedClass := class
	shiftedClass.StartTime = shiftedStart
	shiftedClass.EndTime = shiftedEnd

	return shiftedClass.ToEvent()
}

// can return that it found one and its identical, or that it found none
// returns: (found bool, different bool, event *calendar.Event)
func checkForExistingEvent(newEvent calendar.Event, currentEvents []*calendar.Event) (isNew bool, event *calendar.Event) {
	for _, event := range currentEvents {
		if event.ExtendedProperties.Private["hash"] == newEvent.ExtendedProperties.Private["hash"] {
			slog.Debug("Event is identical", slog.String("summary", event.Summary))
			return false, event
		}
	}
	slog.Debug("Event not found", slog.String("summary", newEvent.Summary))
	return true, nil
}
