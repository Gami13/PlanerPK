package types

import (
	"encoding/json"
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

func TestRecurringEventTimezoneAcrossDST(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	for _, day := range []int{21, 28} {
		start := time.Date(2026, 10, day, 8, 0, 0, 0, zone)
		schedule := ClassSchedule{Start: start, End: start.Add(90 * time.Minute),
			Recurrence: &WeeklyRecurrence{Interval: 1, Until: start.AddDate(0, 0, 7)}}
		event := (Class{}).ToEvent(nil, &schedule)
		payload, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		var sent struct {
			Start, End struct{ DateTime, TimeZone string }
		}
		if err := json.Unmarshal(payload, &sent); err != nil {
			t.Fatal(err)
		}
		if sent.Start.TimeZone != "Europe/Warsaw" || sent.End.TimeZone != "Europe/Warsaw" ||
			sent.Start.DateTime != start.Format(time.RFC3339) || sent.End.DateTime != schedule.End.Format(time.RFC3339) {
			t.Fatalf("incorrect Google event payload: %s", payload)
		}
		schedule.Recurrence = nil
		single := (Class{}).ToEvent(nil, &schedule)
		if single.Start.TimeZone != "" || single.End.TimeZone != "" {
			t.Fatal("single event timezone behavior changed")
		}
	}
	legacy := (Class{}).ToEvent(nil, nil)
	if legacy.Start.TimeZone != "" || legacy.End.TimeZone != "" {
		t.Fatal("legacy event timezone behavior changed")
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
	recurrence := &WeeklyRecurrence{
		Interval:   1,
		Until:      time.Date(2026, 10, 28, 8, 0, 0, 0, time.UTC),
		Exclusions: []time.Time{time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC)},
	}
	for _, tt := range []struct {
		recurrence *WeeklyRecurrence
		rules      []string
	}{
		{},
		{recurrence, []string{"RRULE:FREQ=WEEKLY;UNTIL=20261028T080000Z", "EXDATE:20261014T080000Z"}},
	} {
		schedule := ClassSchedule{
			Start:      time.Date(2026, 10, 7, 10, 0, 0, 0, time.Local),
			End:        time.Date(2026, 10, 7, 11, 30, 0, 0, time.Local),
			Recurrence: tt.recurrence,
		}
		event := class.ToEvent(nil, &schedule)
		if event.Start.DateTime != schedule.Start.Format(time.RFC3339) || event.End.DateTime != schedule.End.Format(time.RFC3339) {
			t.Errorf("resolved dates not preserved: %v – %v", event.Start, event.End)
		}
		if !reflect.DeepEqual(event.Recurrence, tt.rules) {
			t.Errorf("Recurrence = %v, want %v", event.Recurrence, tt.rules)
		}
		if tt.recurrence != nil {
			event.Recurrence[0] = "changed"
			if schedule.Recurrence.ToRRULE() != tt.rules[0] {
				t.Error("event mutation changed schedule recurrence")
			}
		}
		event.Start, event.End, event.Recurrence = legacy.Start, legacy.End, legacy.Recurrence
		if !reflect.DeepEqual(event, legacy) {
			t.Error("resolved schedule changed Excel metadata")
		}
	}
}
