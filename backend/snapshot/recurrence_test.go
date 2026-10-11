package snapshot

import (
	"strings"
	"testing"
	"time"
)

func TestWeeklyRecurrence(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	previous := time.Local
	time.Local = zone
	t.Cleanup(func() { time.Local = previous })
	day, start := "MON", 480
	calendar := AcademicCalendar{Start: "2026-10-20", End: "2026-11-10", HalfBoundary: "2026-11-02"}
	for _, tt := range []struct {
		frequency, first, until string
	}{
		{frequencyWeekly, "2026-10-26T08:00:00+01:00", "20261109T070000Z"},
		{frequencyFirstHalf, "2026-10-26T08:00:00+01:00", "20261026T070000Z"},
		{frequencySecondHalf, "2026-11-02T08:00:00+01:00", "20261109T070000Z"},
	} {
		t.Run(tt.frequency, func(t *testing.T) {
			block := Block{Day: &day, Start: &start, Duration: 90, Frequency: tt.frequency}
			got, err := block.WeeklyRecurrence(calendar)
			wantRule := "RRULE:FREQ=WEEKLY;UNTIL=" + tt.until
			if err != nil || got == nil || got.First.Start.Format(time.RFC3339) != tt.first || got.First.End.Sub(got.First.Start) != 90*time.Minute || len(got.Rules) != 1 || got.Rules[0] != wantRule {
				t.Fatalf("recurrence = %+v, error = %v; want first %s and %s", got, err, tt.first, wantRule)
			}
		})
	}
	block := Block{Day: &day, Start: &start, Duration: 90, Frequency: frequencyWeekly}
	calendar.Start, calendar.End = "2026-10-19", "2026-11-02"
	got, err := block.WeeklyRecurrence(calendar)
	if err != nil || got == nil {
		t.Fatalf("recurrence = %+v, error = %v", got, err)
	}
	until, err := time.Parse("20060102T150405Z", strings.TrimPrefix(got.Rules[0], "RRULE:FREQ=WEEKLY;UNTIL="))
	if err != nil {
		t.Fatal(err)
	}
	var dates []string
	for date := got.First.Start; !date.After(until); date = date.AddDate(0, 0, 7) {
		dates = append(dates, date.Format(time.RFC3339))
	}
	if strings.Join(dates, ",") != "2026-10-19T08:00:00+02:00,2026-10-26T08:00:00+01:00,2026-11-02T08:00:00+01:00" {
		t.Fatalf("weekly recurrence across DST = %v", dates)
	}
	calendar.End = calendar.Start
	day = "TUE"
	if got, err := block.WeeklyRecurrence(calendar); err != nil || got != nil {
		t.Fatalf("empty schedule: recurrence = %+v, error = %v", got, err)
	}
	calendar.Start = "bad"
	if _, err := block.WeeklyRecurrence(calendar); err == nil {
		t.Fatal("accepted invalid term start")
	}
	block.Frequency = frequencyConcreteDate
	if _, err := block.WeeklyRecurrence(calendar); err == nil {
		t.Fatal("accepted unsupported concrete-date recurrence")
	}
}

func TestAlternatingRecurrence(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	previous := time.Local
	time.Local = zone
	t.Cleanup(func() { time.Local = previous })
	day, start := "WED", 480
	block := Block{Day: &day, Start: &start, Duration: 90, Frequency: frequencyAlternating}
	calendar := AcademicCalendar{Start: "2026-10-13", End: "2026-11-12", ParityAnchor: "2026-09-28"}
	for _, tt := range []struct {
		name, until, dates string
		parity             int
	}{
		{"A", "20261111T070000Z", "2026-10-14T08:00:00+02:00,2026-10-28T08:00:00+01:00,2026-11-11T08:00:00+01:00", teachingWeekA},
		{"B", "20261104T070000Z", "2026-10-21T08:00:00+02:00,2026-11-04T08:00:00+01:00", teachingWeekB},
	} {
		t.Run(tt.name, func(t *testing.T) {
			block.TeachingWeekParity = &tt.parity
			got, err := block.WeeklyRecurrence(calendar)
			wantRule := "RRULE:FREQ=WEEKLY;INTERVAL=2;UNTIL=" + tt.until
			if err != nil || got == nil || len(got.Rules) != 1 || got.Rules[0] != wantRule {
				t.Fatalf("recurrence = %+v, error = %v; want %s", got, err, wantRule)
			}
			until, err := time.Parse("20060102T150405Z", tt.until)
			if err != nil {
				t.Fatal(err)
			}
			var dates []string
			for date := got.First.Start; !date.After(until); date = date.AddDate(0, 0, 14) {
				dates = append(dates, date.Format(time.RFC3339))
			}
			if strings.Join(dates, ",") != tt.dates || got.First.End.Sub(got.First.Start) != 90*time.Minute {
				t.Fatalf("alternating recurrence = %+v, dates = %v; want %s", got, dates, tt.dates)
			}
		})
	}
	parity := teachingWeekA
	block.TeachingWeekParity, calendar.ParityAnchor = &parity, "2026-10-05"
	if got, err := block.WeeklyRecurrence(calendar); err != nil || got == nil || got.First.Start.Format(time.DateOnly) != "2026-10-21" {
		t.Fatalf("shifted parity anchor: recurrence = %+v, error = %v", got, err)
	}
	invalidParity := 2
	for _, parity := range []*int{nil, &invalidParity} {
		block.TeachingWeekParity = parity
		if _, err := block.WeeklyRecurrence(calendar); err == nil {
			t.Fatal("accepted missing or invalid parity")
		}
	}
	block.TeachingWeekParity, calendar.ParityAnchor = &parity, "bad"
	if _, err := block.WeeklyRecurrence(calendar); err == nil {
		t.Fatal("accepted invalid parity anchor")
	}
}
