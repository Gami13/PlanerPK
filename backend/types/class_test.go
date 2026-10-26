package types

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestClassHashPreservesLegacyFormat(t *testing.T) {
	class := Class{
		Name: "Subject", Level: "B2", Type: "L", Teacher: "Teacher", TotalHours: "30h",
		AdditionalInfo: "A", Location: "Room 1", Day: time.Monday,
		StartTime: time.Date(2000, 1, 1, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2000, 1, 1, 9, 30, 0, 0, time.UTC),
	}
	want := "Subject B2 (L) - Teacher | 30h | A|1|08:00|09:30|Room 1"
	if got := class.Hash(); got != want {
		t.Fatalf("Hash() = %q, want %q", got, want)
	}
	class.Audience = "GL2"
	if got := class.Hash(); got != want+"|GL2" {
		t.Fatalf("Hash() = %q, want audience appended to legacy hash", got)
	}
	if got := (Class{Name: "Subject"}).Hash(); got != "Subject|0|00:00|00:00|" {
		t.Fatalf("empty fields changed hash punctuation: %q", got)
	}
}

func TestCalendarSummaryIncludesWeekAndClassDetails(t *testing.T) {
	class := Class{
		Name: "Matematyka dyskretna", Type: "L", Teacher: "Biskup Piotr",
		Audience: "GL2", AdditionalInfo: "A",
	}
	want := "Matematyka dyskretna (L) - Biskup Piotr"
	if got := class.Summary(); got != want {
		t.Fatalf("Summary() = %q, want %q", got, want)
	}
	if got := class.ToEvent(nil, nil).Summary; got != want {
		t.Fatalf("calendar summary = %q, want %q", got, want)
	}
	if got := class.ToEvent(nil, nil).Description; got != "Grupa: GL2<br>Info: A" {
		t.Fatalf("calendar description = %q, want group and info rows", got)
	}
}

func TestAudienceIsIncludedInClassOutput(t *testing.T) {
	class := Class{Name: "Subject", Audience: "GL2"}
	if strings.Contains(class.Summary(), "GL2") {
		t.Errorf("Summary() = %q, audience should only appear in the description", class.Summary())
	}
	if !strings.Contains(class.String(), "GL2") {
		t.Errorf("String() = %q, want audience", class.String())
	}
	event := class.ToEvent(nil, nil)
	if !strings.Contains(event.Description, "Grupa: GL2") || strings.Contains(event.Description, "Audience:") {
		t.Errorf("calendar description = %q, want Polish group label", event.Description)
	}
	if strings.Contains(event.Summary, "GL2") || !strings.Contains(event.Description, "GL2") || event.ExtendedProperties.Private["audience"] != "GL2" {
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
