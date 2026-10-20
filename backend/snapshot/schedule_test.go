package snapshot

import (
	"slices"
	"testing"
	"time"
)

func TestClassSchedule(t *testing.T) {
	previous := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = previous })
	day, start, date, parity := "WED", 480, "2026-10-07", teachingWeekA
	calendar := AcademicCalendar{
		Start: "2026-10-05", End: "2026-10-21", HalfBoundary: "2026-10-12", ParityAnchor: "2026-09-28",
		Closures: []DateRange{{From: date, To: date}},
	}
	for _, tt := range []struct {
		name, frequency, date, first string
		rules                        []string
		parked, wantError            bool
	}{
		{"weekly excluded first", frequencyWeekly, date, date, []string{"RRULE:FREQ=WEEKLY;UNTIL=20261021T080000Z", "EXDATE:20261007T080000Z"}, false, false},
		{"alternating", frequencyAlternating, date, "2026-10-14", []string{"RRULE:FREQ=WEEKLY;INTERVAL=2;UNTIL=20261014T080000Z"}, false, false},
		{"first half fully closed", frequencyFirstHalf, date, "", nil, false, false},
		{"second half", frequencySecondHalf, date, "2026-10-14", []string{"RRULE:FREQ=WEEKLY;UNTIL=20261021T080000Z"}, false, false},
		{"concrete date", frequencyConcreteDate, "2026-10-14", "2026-10-14", nil, false, false},
		{"closed date", frequencyConcreteDate, date, "", nil, false, false},
		{"outside term", frequencyConcreteDate, "2026-10-28", "", nil, false, false},
		{"one off", frequencyOneOff, date, "", nil, false, false},
		{"first October week", frequencyFirstOctoberWeek, date, "", nil, false, false},
		{"parked", frequencyWeekly, date, "", nil, true, false},
		{"unknown frequency", "unknown", date, "", nil, false, true},
		{"invalid date", frequencyConcreteDate, "bad", "", nil, false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			block := Block{Day: &day, Start: &start, Duration: 90, Frequency: tt.frequency, Date: &tt.date, TeachingWeekParity: &parity}
			if tt.parked {
				block.Day = nil
			}
			got, err := block.ClassSchedule(calendar)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tt.wantError)
			}
			if tt.first == "" {
				if got != nil {
					t.Fatalf("schedule = %+v, want nil", got)
				}
				return
			}
			if got == nil || got.Start.Format(time.RFC3339) != tt.first+"T08:00:00Z" || got.End.Format(time.RFC3339) != tt.first+"T09:30:00Z" || !slices.Equal(got.Recurrence, tt.rules) {
				t.Fatalf("schedule = %+v, want %s 08:00–09:30 and rules %v", got, tt.first, tt.rules)
			}
		})
	}
}
