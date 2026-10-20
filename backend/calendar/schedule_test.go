package calendar

import (
	"reflect"
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
		Start:      time.Date(2026, 10, 19, 8, 0, 0, 0, zone),
		End:        time.Date(2026, 10, 19, 9, 30, 0, 0, zone),
		Recurrence: []string{"RRULE:FREQ=WEEKLY;INTERVAL=2;UNTIL=20261102T070000Z", "EXDATE:20261019T060000Z", "EXDATE:20261102T070000Z"},
	}
	wantRules := []string{"RRULE:FREQ=WEEKLY;INTERVAL=2;UNTIL=20261102T093000Z", "EXDATE:20261019T083000Z", "EXDATE:20261102T093000Z"}
	for _, rules := range [][]string{nil, original.Recurrence} {
		before := append([]string(nil), rules...)
		schedule := original
		schedule.Recurrence = rules
		got, err := shiftClassSchedule(schedule, "10:30")
		if err != nil || got.Start.Format(time.RFC3339) != "2026-10-19T10:30:00+02:00" || got.End.Format(time.RFC3339) != "2026-10-19T12:00:00+02:00" {
			t.Fatalf("shifted schedule = %+v, error = %v", got, err)
		}
		if rules != nil && !reflect.DeepEqual(got.Recurrence, wantRules) {
			t.Fatalf("rules = %v, want %v", got.Recurrence, wantRules)
		}
		if rules == nil && got.Recurrence != nil {
			t.Fatal("single event gained recurrence")
		}
		unchanged, err := shiftClassSchedule(schedule, "")
		if err != nil || !reflect.DeepEqual(unchanged, schedule) || !reflect.DeepEqual(rules, before) {
			t.Fatal("original schedule changed")
		}
	}
	for _, hour := range []string{"bad", "24:00", "10:60"} {
		if _, err := shiftClassSchedule(original, hour); err == nil {
			t.Errorf("accepted start hour %q", hour)
		}
	}
}
