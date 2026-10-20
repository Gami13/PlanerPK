package types

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAudienceIsIncludedInClassOutput(t *testing.T) {
	class := Class{Name: "Subject", Audience: "GL2"}
	if !strings.Contains(class.Summary(), "GL2") {
		t.Errorf("Summary() = %q, want audience", class.Summary())
	}
	if !strings.Contains(class.String(), "GL2") {
		t.Errorf("String() = %q, want audience", class.String())
	}
	event := class.ToEvent(nil, nil)
	if !strings.Contains(event.Summary, "GL2") || !strings.Contains(event.Description, "GL2") || event.ExtendedProperties.Private["audience"] != "GL2" {
		t.Errorf("ToEvent() omitted audience: %+v", event)
	}
}

func TestToEventResolvedSchedule(t *testing.T) {
	class := Class{
		Name: "Subject", Type: "Lecture", Teacher: "Teacher", Audience: "GL2",
		Location: "Room 1", AdditionalInfo: "Scheduling note", Day: time.Monday,
		StartTime: time.Date(2000, 1, 1, 8, 0, 0, 0, time.Local),
		EndTime:   time.Date(2000, 1, 1, 9, 0, 0, 0, time.Local),
	}
	legacy := class.ToEvent(nil, nil)
	for _, rules := range [][]string{nil, {"RRULE:FREQ=WEEKLY;UNTIL=20261028T080000Z", "EXDATE:20261014T080000Z"}} {
		schedule := ClassSchedule{
			Start:      time.Date(2026, 10, 7, 10, 0, 0, 0, time.Local),
			End:        time.Date(2026, 10, 7, 11, 30, 0, 0, time.Local),
			Recurrence: rules,
		}
		event := class.ToEvent(nil, &schedule)
		if event.Start.DateTime != schedule.Start.Format(time.RFC3339) || event.End.DateTime != schedule.End.Format(time.RFC3339) {
			t.Errorf("resolved dates not preserved: %v – %v", event.Start, event.End)
		}
		if !reflect.DeepEqual(event.Recurrence, rules) {
			t.Errorf("Recurrence = %v, want %v", event.Recurrence, rules)
		}
		if len(rules) > 0 {
			event.Recurrence[0] = "changed"
			if schedule.Recurrence[0] == "changed" {
				t.Error("event recurrence aliases schedule")
			}
		}
		event.Start, event.End, event.Recurrence = legacy.Start, legacy.End, legacy.Recurrence
		if !reflect.DeepEqual(event, legacy) {
			t.Error("resolved schedule changed Excel metadata")
		}
	}
}
