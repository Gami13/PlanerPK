package calendar

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

func getEventHash(event *calendar.Event) (string, bool) {
	if event == nil || event.ExtendedProperties == nil || event.ExtendedProperties.Private == nil {
		return "", false
	}
	hash, ok := event.ExtendedProperties.Private["hash"]
	return hash, ok
}

func buildCurrentEventsMap(currentEvents []*calendar.Event) map[string]*calendar.Event {
	currentMap := make(map[string]*calendar.Event, len(currentEvents))
	for _, event := range currentEvents {
		if hash, ok := getEventHash(event); ok {
			currentMap[hash] = event
		} else {
			slog.Warn("Event missing hash extended property", slog.String("summary", event.Summary))
		}
	}
	return currentMap
}

func calculateEventDiffs(currentEvents []*calendar.Event, newEvents map[string]*calendar.Event) ([]*calendar.Event, []*calendar.Event, []*calendar.Event) {
	currentMap := buildCurrentEventsMap(currentEvents)

	eventsToDelete := make([]*calendar.Event, 0)
	for _, event := range currentEvents {
		hash, ok := getEventHash(event)
		if !ok {
			continue
		}
		if _, ok := newEvents[hash]; !ok {
			eventsToDelete = append(eventsToDelete, event)
		}
	}

	eventsToAdd := make([]*calendar.Event, 0)
	unchangedEvents := make([]*calendar.Event, 0)

	for hash, newEvent := range newEvents {
		if existingEvent, ok := currentMap[hash]; ok {
			slog.Debug("Event is identical", slog.String("summary", existingEvent.Summary))
			unchangedEvents = append(unchangedEvents, existingEvent)
		} else {
			slog.Debug("Event not found", slog.String("summary", newEvent.Summary))
			eventsToAdd = append(eventsToAdd, newEvent)
		}
	}

	return eventsToDelete, eventsToAdd, unchangedEvents
}

func buildNewEventsMap(timetable types.GroupTimetable, overrides map[string]db.Override, buildings *types.Buildings) map[string]*calendar.Event {
	newEvents := make(map[string]*calendar.Event)
	for day, entry := range timetable {
		slog.Debug("Processing day", slog.String("day", day.String()))
		for _, class := range entry {
			ovr, hasOverride := overrides[class.Hash()]

			// No override — emit normally.
			if !hasOverride {
				event := class.ToEvent(buildings)
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
				ghost := buildGhostEvent(class, buildings)
				newEvents[ghost.ExtendedProperties.Private["hash"]] = &ghost
				continue
			}

			// Not hidden — emit real event, possibly time-shifted.
			realEvent := buildRealEvent(class, ovr.StartHour, buildings)
			newEvents[realEvent.ExtendedProperties.Private["hash"]] = &realEvent

			// If transparent, also emit a ghost at the original position.
			if ovr.IsTransparent && ovr.StartHour != "" {
				slog.Debug("Adding ghost for shifted class", slog.String("hash", class.Hash()))
				ghost := buildGhostEvent(class, buildings)
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
func buildGhostEvent(class types.Class, buildings *types.Buildings) calendar.Event {
	event := class.ToEvent(buildings)
	event.Transparency = "transparent"
	event.ColorId = "8" // Graphite — muted gray, signals a placeholder
	// Use a distinct hash so GCal treats this as a separate event.
	event.ExtendedProperties.Private["hash"] = class.Hash() + "_ghost"
	return event
}

// buildRealEvent builds the primary calendar event, optionally shifting the start time.
// startHour must be "HH:MM" or empty.
func buildRealEvent(class types.Class, startHour string, buildings *types.Buildings) calendar.Event {
	if startHour == "" {
		return class.ToEvent(buildings)
	}

	var h, m int
	if _, err := fmt.Sscanf(startHour, "%d:%d", &h, &m); err != nil {
		slog.Error("Failed to parse start_hour override, using original time",
			slog.String("start_hour", startHour),
			slog.Any("error", err))
		return class.ToEvent(buildings)
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

	return shiftedClass.ToEvent(buildings)
}
