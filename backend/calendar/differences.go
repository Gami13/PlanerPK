package calendar

import (
	"log/slog"

	"github.com/gami13/PlanerPK/types"
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

func buildNewEventsMap(timetable types.Timetable) map[string]*calendar.Event {
	newEvents := make(map[string]*calendar.Event)
	for day, entry := range timetable {
		slog.Debug("Processing day", slog.String("day", day.String()))
		for _, class := range entry {
			event := class.ToEvent()
			newEvents[event.ExtendedProperties.Private["hash"]] = &event
		}
	}
	return newEvents
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
