package snapshot

import (
	"slices"
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
			if err != nil || got == nil || got.Start.Format(time.RFC3339) != tt.first || got.End.Sub(got.Start) != 90*time.Minute || len(got.Recurrence.Exclusions) != 0 || got.Recurrence.ToRRULE() != wantRule {
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
	until := got.Recurrence.Until
	var dates []string
	for date := got.Start; !date.After(until); date = date.AddDate(0, 0, 7) {
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

func TestRecurrenceClosureExclusions(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	previous := time.Local
	time.Local = zone
	t.Cleanup(func() { time.Local = previous })
	day, start := "MON", 480
	calendar := AcademicCalendar{
		Start: "2026-10-19", End: "2026-11-16", ParityAnchor: "2026-09-28",
		Closures:        []DateRange{{From: "2026-10-19", To: "2026-10-26"}, {From: "2026-10-26", To: "2026-10-26"}, {From: "2026-11-16", To: "2026-11-16"}},
		PartialClosures: []PartialClosure{{Date: "2026-11-02", Start: 480, End: 540}},
	}
	for _, tt := range []struct {
		name, frequency  string
		parity, from, to int
		want             []string
	}{
		{"weekly overlap", frequencyWeekly, 0, 480, 540, []string{"RRULE:FREQ=WEEKLY;UNTIL=20261116T070000Z", "EXDATE:20261019T060000Z", "EXDATE:20261026T070000Z", "EXDATE:20261102T070000Z", "EXDATE:20261116T070000Z"}},
		{"A week", frequencyAlternating, teachingWeekA, 480, 540, []string{"RRULE:FREQ=WEEKLY;INTERVAL=2;UNTIL=20261109T070000Z", "EXDATE:20261026T070000Z"}},
		{"B week fully closed", frequencyAlternating, teachingWeekB, 480, 540, nil},
		{"class ends at closure start", frequencyWeekly, 0, 570, 660, []string{"RRULE:FREQ=WEEKLY;UNTIL=20261116T070000Z", "EXDATE:20261019T060000Z", "EXDATE:20261026T070000Z", "EXDATE:20261116T070000Z"}},
		{"class starts at closure end", frequencyWeekly, 0, 420, 480, []string{"RRULE:FREQ=WEEKLY;UNTIL=20261116T070000Z", "EXDATE:20261019T060000Z", "EXDATE:20261026T070000Z", "EXDATE:20261116T070000Z"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			block := Block{Day: &day, Start: &start, Duration: 90, Frequency: tt.frequency, TeachingWeekParity: &tt.parity}
			calendar.PartialClosures[0].Start, calendar.PartialClosures[0].End = tt.from, tt.to
			got, err := block.WeeklyRecurrence(calendar)
			if err != nil {
				t.Fatal(err)
			}
			if tt.want == nil {
				if got != nil {
					t.Fatalf("fully closed schedule = %+v, want nil", got)
				}
				return
			}
			if got == nil || !slices.Equal(append([]string{got.Recurrence.ToRRULE()}, got.Recurrence.ToEXDATE()...), tt.want) {
				t.Fatalf("recurrence = %+v, want rules %v", got, tt.want)
			}
			wantFirst := "2026-10-19"
			if tt.frequency == frequencyAlternating {
				wantFirst = "2026-10-26"
			}
			if got.Start.Format(time.DateOnly) != wantFirst {
				t.Fatalf("series start changed to %s, want %s", got.Start, wantFirst)
			}
		})
	}
	block := Block{Day: &day, Start: &start, Duration: 90, Frequency: frequencyWeekly}
	calendar.PartialClosures[0].Date = "bad"
	if _, err := block.WeeklyRecurrence(calendar); err == nil {
		t.Fatal("accepted invalid closure date")
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
			if err != nil || got == nil || len(got.Recurrence.Exclusions) != 0 || got.Recurrence.ToRRULE() != wantRule {
				t.Fatalf("recurrence = %+v, error = %v; want %s", got, err, wantRule)
			}
			until, err := time.Parse("20060102T150405Z", tt.until)
			if err != nil {
				t.Fatal(err)
			}
			var dates []string
			for date := got.Start; !date.After(until); date = date.AddDate(0, 0, 14) {
				dates = append(dates, date.Format(time.RFC3339))
			}
			if strings.Join(dates, ",") != tt.dates || got.End.Sub(got.Start) != 90*time.Minute {
				t.Fatalf("alternating recurrence = %+v, dates = %v; want %s", got, dates, tt.dates)
			}
		})
	}
	parity := teachingWeekA
	block.TeachingWeekParity, calendar.ParityAnchor = &parity, "2026-10-05"
	if got, err := block.WeeklyRecurrence(calendar); err != nil || got == nil || got.Start.Format(time.DateOnly) != "2026-10-21" {
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
