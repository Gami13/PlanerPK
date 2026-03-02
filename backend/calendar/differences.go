package calendar

import (
	"log/slog"

	"github.com/WoodPKrs/GoogleCalendarTimetable/types"
	"google.golang.org/api/calendar/v3"
)

func findEventsToDelete(currentEvents []*calendar.Event, newEvents map[string]*calendar.Event) []*calendar.Event {
	eventsToDelete := make([]*calendar.Event, 0)
	for _, event := range currentEvents {
		if _, ok := newEvents[event.ExtendedProperties.Private["event_id"]]; !ok {
			eventsToDelete = append(eventsToDelete, event)
		}
	}
	return eventsToDelete
}

func calculateEventDiffs(currentEvents []*calendar.Event, newEvents map[string]*calendar.Event) ([]*calendar.Event, []eventUpdate, []*calendar.Event, []*calendar.Event) {
	eventsToDelete := findEventsToDelete(currentEvents, newEvents)
	eventsToUpdate := make([]eventUpdate, 0)
	eventsToAdd := make([]*calendar.Event, 0)
	unchangedEvents := make([]*calendar.Event, 0)

	for _, newEvent := range newEvents {
		decision, existingEvent := checkForExistingEvent(*newEvent, currentEvents)
		if decision == EVENT_NEW {
			eventsToAdd = append(eventsToAdd, newEvent)
			continue
		}

		if decision == EVENT_UPDATE {
			eventsToUpdate = append(eventsToUpdate, eventUpdate{
				existingEvent: existingEvent,
				newEvent:      newEvent,
			})
			continue
		}

		unchangedEvents = append(unchangedEvents, existingEvent)
	}

	return eventsToDelete, eventsToUpdate, eventsToAdd, unchangedEvents
}

func buildNewEventsMap(timetable types.Timetable) map[string]*calendar.Event {
	newEvents := make(map[string]*calendar.Event)
	for day, entry := range timetable {
		slog.Debug("Processing day", slog.String("day", day.String()))
		for _, class := range entry {
			event := class.ToEvent()
			newEvents[event.ExtendedProperties.Private["event_id"]] = &event
		}
	}
	return newEvents
}

type decision int

const (
	EVENT_UNCHANGED decision = iota
	EVENT_NEW
	EVENT_UPDATE
	EVENT_DELETE
)

// can return that it found one and its identical, or that it found one but it's different, or that it found none, and can return the found event if any
// returns: (found bool, different bool, event *calendar.Event)
func checkForExistingEvent(newEvent calendar.Event, currentEvents []*calendar.Event) (decision decision, event *calendar.Event) {
	for _, event := range currentEvents {
		if event.ExtendedProperties.Private["event_id"] == newEvent.ExtendedProperties.Private["event_id"] {
			slog.Debug("Comparing events", slog.String("Old", event.Summary), slog.String("New", newEvent.Summary))
			if event.ExtendedProperties.Private["hash"] != newEvent.ExtendedProperties.Private["hash"] || event.Summary != newEvent.Summary || event.Start.DateTime != newEvent.Start.DateTime || event.End.DateTime != newEvent.End.DateTime {
				slog.Debug("Event differs",
					slog.String("old_summary", event.Summary),
					slog.String("new_summary", newEvent.Summary))
				return EVENT_UPDATE, event
			}
			slog.Debug("Event is identical", slog.String("summary", event.Summary))
			return EVENT_UNCHANGED, event
		}
	}
	slog.Debug("Event not found", slog.String("summary", newEvent.Summary))
	return EVENT_NEW, nil
}
