package calendar

import (
	"fmt"

	"google.golang.org/api/calendar/v3"
)

type eventChange struct {
	Current, Proposed *calendar.Event
}

type eventDiff struct {
	ToDelete, ToAdd, Unchanged, Exceptions []*calendar.Event
	Changed                                []eventChange
}

func getEventOverrideID(event *calendar.Event) (string, bool) {
	if event == nil || event.ExtendedProperties == nil {
		return "", false
	}
	id, ok := event.ExtendedProperties.Private["overrideId"]
	return id, ok
}

// buildCurrentEventsMap maps series and single events by overrideId, collects
// recurring exceptions separately, and rejects nil events or missing/duplicate IDs.
func buildCurrentEventsMap(events []*calendar.Event) (map[string]*calendar.Event, []*calendar.Event, error) {
	series := make(map[string]*calendar.Event)
	var exceptions []*calendar.Event
	for _, event := range events {
		if event == nil {
			return nil, nil, fmt.Errorf("nil event in diff")
		}
		if event.RecurringEventId != "" {
			exceptions = append(exceptions, event)
			continue
		}
		id, ok := getEventOverrideID(event)
		if !ok || id == "" {
			return nil, nil, fmt.Errorf("event %q has no overrideId", event.Id)
		}
		if _, ok := series[id]; ok {
			return nil, nil, fmt.Errorf("duplicate overrideId %q", id)
		}
		series[id] = event
	}
	return series, exceptions, nil
}

// calculateEventDiffs compares existing and proposed events by overrideId and contents,
// returning add, delete, unchanged, and changed lists. Existing exceptions are
// returned separately for later handling; proposed exceptions are not supported.
// This only prepares a diff and does not write to Google Calendar.
func calculateEventDiffs(current, proposed []*calendar.Event) (eventDiff, error) {
	old, exceptions, err := buildCurrentEventsMap(current)
	if err != nil {
		return eventDiff{}, err
	}
	next, proposedExceptions, err := buildCurrentEventsMap(proposed)
	if err != nil {
		return eventDiff{}, err
	}
	if len(proposedExceptions) != 0 {
		return eventDiff{}, fmt.Errorf("proposed exceptions are not supported yet")
	}
	diff := eventDiff{Exceptions: exceptions}
	for _, event := range current {
		id, _ := getEventOverrideID(event)
		if event.RecurringEventId == "" && next[id] == nil {
			diff.ToDelete = append(diff.ToDelete, event)
		}
	}
	for _, event := range proposed {
		id, _ := getEventOverrideID(event)
		previous := old[id]
		switch {
		case previous == nil:
			diff.ToAdd = append(diff.ToAdd, event)
		case eventContentHash(*previous) == eventContentHash(*event):
			diff.Unchanged = append(diff.Unchanged, previous)
		default:
			diff.Changed = append(diff.Changed, eventChange{Current: previous, Proposed: event})
		}
	}
	return diff, nil
}
