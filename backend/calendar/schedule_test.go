package calendar

import (
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
)

func TestShiftClassSchedule(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	original := types.ClassSchedule{
		Start: time.Date(2026, 10, 19, 8, 0, 0, 0, zone),
		End:   time.Date(2026, 10, 19, 9, 30, 0, 0, zone),
		Recurrence: &types.WeeklyRecurrence{
			Interval: 2,
			Until:    time.Date(2026, 11, 2, 8, 0, 0, 0, zone),
			Exclusions: []time.Time{
				time.Date(2026, 10, 19, 8, 0, 0, 0, zone),
				time.Date(2026, 11, 2, 8, 0, 0, 0, zone),
			},
		},
	}
	wantRules := []string{"RRULE:FREQ=WEEKLY;INTERVAL=2;UNTIL=20261102T093000Z", "EXDATE:20261019T083000Z", "EXDATE:20261102T093000Z"}
	for _, recurrence := range []*types.WeeklyRecurrence{nil, original.Recurrence} {
		schedule := original
		schedule.Recurrence = recurrence
		var before types.WeeklyRecurrence
		if recurrence != nil {
			before = *recurrence
			before.Exclusions = slices.Clone(recurrence.Exclusions)
		}
		got, err := shiftClassSchedule(schedule, "10:30")
		if err != nil || got.Start.Format(time.RFC3339) != "2026-10-19T10:30:00+02:00" || got.End.Format(time.RFC3339) != "2026-10-19T12:00:00+02:00" {
			t.Fatalf("shifted schedule = %+v, error = %v", got, err)
		}
		event := (types.Class{}).ToEvent(nil, got)
		if recurrence != nil && !reflect.DeepEqual(event.Recurrence, wantRules) {
			t.Fatalf("rules = %v, want %v", event.Recurrence, wantRules)
		}
		if recurrence == nil && got.Recurrence != nil {
			t.Fatal("single event gained recurrence")
		}
		unchanged, err := shiftClassSchedule(schedule, "")
		if err != nil || !reflect.DeepEqual(unchanged, schedule) {
			t.Fatal("empty override changed schedule")
		}
		if recurrence != nil {
			unchanged.Recurrence.Until = time.Time{}
			unchanged.Recurrence.Exclusions[0] = time.Time{}
		}
		if recurrence != nil && !reflect.DeepEqual(*recurrence, before) {
			t.Fatal("original schedule changed")
		}
	}
	for _, hour := range []string{"bad", "24:00", "10:60"} {
		if _, err := shiftClassSchedule(original, hour); err == nil {
			t.Errorf("accepted start hour %q", hour)
		}
	}
}
