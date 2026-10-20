package types

import (
	"slices"
	"testing"
	"time"
)

func TestWeeklyRecurrenceEncoding(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	recurrence := WeeklyRecurrence{
		Until: time.Date(2026, 11, 2, 8, 0, 0, 0, zone),
		Exclusions: []time.Time{
			time.Date(2026, 10, 19, 8, 0, 0, 0, zone),
			time.Date(2026, 10, 26, 8, 0, 0, 0, zone),
		},
	}
	for _, tt := range []struct {
		interval int
		want     string
	}{
		{1, "RRULE:FREQ=WEEKLY;UNTIL=20261102T070000Z"},
		{2, "RRULE:FREQ=WEEKLY;INTERVAL=2;UNTIL=20261102T070000Z"},
	} {
		recurrence.Interval = tt.interval
		if got := recurrence.ToRRULE(); got != tt.want {
			t.Errorf("ToRRULE() = %q, want %q", got, tt.want)
		}
	}
	want := []string{"EXDATE:20261019T060000Z", "EXDATE:20261026T070000Z"}
	if got := recurrence.ToEXDATE(); !slices.Equal(got, want) {
		t.Errorf("ToEXDATE() = %v, want %v", got, want)
	}
	recurrence.Exclusions = nil
	if got := recurrence.ToEXDATE(); len(got) != 0 {
		t.Errorf("ToEXDATE() = %v, want no exclusions", got)
	}
}
