package calendar

import (
	"reflect"
	"testing"

	"google.golang.org/api/calendar/v3"
)

func TestResolvedEventDiffs(t *testing.T) {
	makeEvent := func(id string) *calendar.Event {
		return &calendar.Event{Id: "google-" + id, Summary: id,
			Recurrence:         []string{"RRULE:FREQ=WEEKLY"},
			ExtendedProperties: &calendar.EventExtendedProperties{Private: map[string]string{"overrideId": id, "contentHash": "stale"}}}
	}
	keep, changed, removed, ghost := makeEvent("keep"), makeEvent("changed"), makeEvent("removed"), makeEvent("keep_ghost")
	modified := makeEvent("changed")
	modified.Recurrence = append(modified.Recurrence, "EXDATE:20261014T060000Z")
	added := makeEvent("added")
	exception := makeEvent("changed") // Inherited identity must not overwrite the series.
	exception.Id, exception.RecurringEventId, exception.Status = "exception", changed.Id, "cancelled"
	current := []*calendar.Event{exception, keep, changed, removed, ghost}
	proposed := []*calendar.Event{makeEvent("keep"), modified, added, makeEvent("keep_ghost")}
	diff, err := compareEventsByOverrideIDAndContent(current, proposed)
	want := eventDiff{ToDelete: []*calendar.Event{removed}, ToAdd: []*calendar.Event{added},
		Unchanged: []*calendar.Event{keep, ghost}, Exceptions: []*calendar.Event{exception},
		Changed: []eventChange{{Current: changed, Proposed: modified}}}
	if err != nil || !reflect.DeepEqual(diff, want) {
		t.Fatalf("diff = %+v, error = %v; want %+v", diff, err, want)
	}
	if changed.Recurrence[0] != "RRULE:FREQ=WEEKLY" || modified.Id != "google-changed" {
		t.Fatal("diff mutated input events")
	}
	for _, pair := range []struct{ current, proposed []*calendar.Event }{
		{current, []*calendar.Event{added, added}},
		{[]*calendar.Event{keep, keep}, proposed},
		{current, []*calendar.Event{{Summary: "missing ID"}}},
		{[]*calendar.Event{nil}, proposed},
		{current, []*calendar.Event{exception}},
	} {
		got, err := compareEventsByOverrideIDAndContent(pair.current, pair.proposed)
		if err == nil || !reflect.DeepEqual(got, eventDiff{}) {
			t.Fatalf("invalid input returned diff = %+v, error = %v", got, err)
		}
	}
}
