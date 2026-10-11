package snapshot

import (
	"testing"
	"time"
)

func TestWeeklyOccurrences(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	previous := time.Local
	time.Local = zone
	t.Cleanup(func() { time.Local = previous })
	day, start := "MON", 480
	block := Block{Day: &day, Start: &start, Duration: 90, Frequency: "co_tydzien"}
	calendar := AcademicCalendar{Start: "2026-10-19", End: "2026-11-02"}
	got, err := block.Occurrences(calendar)
	want := []string{"2026-10-19T08:00:00+02:00", "2026-10-26T08:00:00+01:00", "2026-11-02T08:00:00+01:00"}
	if err != nil || len(got) != len(want) {
		t.Fatalf("occurrences = %+v, error = %v", got, err)
	}
	for i, occurrence := range got {
		if occurrence.Start.Format(time.RFC3339) != want[i] || occurrence.End.Sub(occurrence.Start) != 90*time.Minute {
			t.Fatalf("occurrence %d = %+v, want %s lasting 90 minutes", i, occurrence, want[i])
		}
	}
	calendar.Start = "2026-10-20"
	if got, err = block.Occurrences(calendar); err != nil || len(got) != 2 {
		t.Fatalf("partial first week: occurrences = %+v, error = %v", got, err)
	}
	for _, invalid := range []AcademicCalendar{{Start: "bad", End: calendar.End}, {Start: calendar.Start, End: "bad"}, {Start: calendar.End, End: calendar.Start}} {
		if _, err := block.Occurrences(invalid); err == nil {
			t.Fatalf("accepted invalid bounds: %+v", invalid)
		}
	}
	block.Day = nil
	if got, err = block.Occurrences(calendar); err != nil || len(got) != 0 {
		t.Fatalf("parked block: occurrences = %+v, error = %v", got, err)
	}
	block.Day, block.Duration = &day, -1
	if _, err := block.Occurrences(calendar); err == nil {
		t.Fatal("accepted negative duration")
	}
}

func TestOccurrencesExcludeFullDayClosures(t *testing.T) {
	day, start := "MON", 480
	block := Block{Day: &day, Start: &start, Duration: 90, Frequency: "co_tydzien"}
	calendar := AcademicCalendar{Start: "2026-10-01", End: "2026-10-31", Closures: []DateRange{
		{From: "2026-10-05", To: "2026-10-05"},
		{From: "2026-10-12", To: "2026-10-19"},
		{From: "2026-10-19", To: "2026-10-20"},
	}}
	got, err := block.Occurrences(calendar)
	if err != nil || len(got) != 1 || got[0].Start.Format(time.DateOnly) != "2026-10-26" {
		t.Fatalf("occurrences = %+v, error = %v, want only 26 October", got, err)
	}
	date := "2026-10-19"
	block.Frequency, block.Date = "konkretna_data", &date
	if got, err = block.Occurrences(calendar); err != nil || len(got) != 0 {
		t.Fatalf("closed concrete date: occurrences = %+v, error = %v", got, err)
	}
	for _, invalid := range []DateRange{{From: "bad", To: date}, {From: date, To: "bad"}, {From: "2026-10-20", To: date}} {
		calendar.Closures = []DateRange{invalid}
		if _, err := block.Occurrences(calendar); err == nil {
			t.Fatalf("accepted invalid closure: %+v", invalid)
		}
	}
}

func TestOccurrencesExcludePartialDayClosures(t *testing.T) {
	day, date := "THU", "2026-12-17"
	calendar := AcademicCalendar{Start: date, End: "2026-12-24", PartialClosures: []PartialClosure{{Date: date, Start: 660, End: 840}}}
	for _, tt := range []struct {
		name            string
		start, duration int
		want            int
	}{
		{"before", 450, 90, 2},
		{"ends at closure start", 570, 90, 2},
		{"overlaps start", 555, 135, 1},
		{"inside", 720, 90, 1},
		{"same interval", 660, 180, 1},
		{"spans closure", 600, 300, 1},
		{"overlaps end", 825, 90, 1},
		{"starts at closure end", 840, 90, 2},
		{"after", 870, 90, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			block := Block{Day: &day, Start: &tt.start, Duration: tt.duration, Frequency: frequencyWeekly}
			got, err := block.Occurrences(calendar)
			if err != nil || len(got) != tt.want || got[len(got)-1].Start.Format(time.DateOnly) != calendar.End {
				t.Fatalf("occurrences = %+v, error = %v, want %d including 24 December", got, err, tt.want)
			}
			for _, occurrence := range got {
				if occurrence.Start.Hour()*minutesPerHour+occurrence.Start.Minute() != tt.start || occurrence.End.Sub(occurrence.Start) != time.Duration(tt.duration)*time.Minute {
					t.Fatalf("class time changed: %+v", occurrence)
				}
			}
		})
	}
	start := 720
	block := Block{Day: &day, Start: &start, Duration: 90, Frequency: frequencyConcreteDate, Date: &date}
	if got, err := block.Occurrences(calendar); err != nil || len(got) != 0 {
		t.Fatalf("closed concrete date: occurrences = %+v, error = %v", got, err)
	}
	for _, invalid := range []PartialClosure{{Date: "bad", Start: 660, End: 840}, {Date: date, Start: -1, End: 840}, {Date: date, Start: 660, End: minutesPerDay + 1}, {Date: date, Start: 840, End: 660}, {Date: date, Start: 660, End: 660}} {
		calendar.PartialClosures = []PartialClosure{invalid}
		if _, err := block.Occurrences(calendar); err == nil {
			t.Fatalf("accepted invalid partial closure: %+v", invalid)
		}
	}
}

func TestHalfSemesterAndDatedOccurrences(t *testing.T) {
	day, start, date := "MON", 480, "2026-10-07"
	calendar := AcademicCalendar{Start: "2026-10-01", End: "2026-10-31", HalfBoundary: "2026-10-19"}
	for _, tt := range []struct {
		frequency string
		want      []string
	}{
		{"co_tydzien_pierwsza_polowa", []string{"2026-10-05", "2026-10-12"}},
		{"co_tydzien_druga_polowa", []string{"2026-10-19", "2026-10-26"}},
		{"konkretna_data", []string{"2026-10-07"}},
		{"jednorazowo", nil},
		{"pierwszy_tydzien_pazdziernika", nil},
	} {
		t.Run(tt.frequency, func(t *testing.T) {
			block := Block{Day: &day, Start: &start, Duration: 90, Frequency: tt.frequency, Date: &date}
			got, err := block.Occurrences(calendar)
			if err != nil || len(got) != len(tt.want) {
				t.Fatalf("occurrences = %+v, error = %v, want %v", got, err, tt.want)
			}
			for i := range got {
				if got[i].Start.Format(time.DateOnly) != tt.want[i] {
					t.Fatalf("got %s, want %s", got[i].Start, tt.want[i])
				}
			}
		})
	}
	block := Block{Day: &day, Start: &start, Duration: 90, Frequency: "konkretna_data"}
	if _, err := block.Occurrences(calendar); err == nil {
		t.Fatal("accepted missing concrete date")
	}
	block.Date = &date
	date = "2026-11-01"
	if got, err := block.Occurrences(calendar); err != nil || len(got) != 0 {
		t.Fatalf("outside term: occurrences = %+v, error = %v", got, err)
	}
	block.Frequency, calendar.HalfBoundary = "co_tydzien_pierwsza_polowa", "bad"
	if _, err := block.Occurrences(calendar); err == nil {
		t.Fatal("accepted invalid half boundary")
	}
}

func TestAlternatingWeeklyOccurrences(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	previous := time.Local
	time.Local = zone
	t.Cleanup(func() { time.Local = previous })
	day, start := "WED", 480
	block := Block{Day: &day, Start: &start, Duration: 90, Frequency: "co_2_tygodnie"}
	calendar := AcademicCalendar{Start: "2026-10-19", End: "2026-11-08", ParityAnchor: "2026-09-28"}
	for parity, want := range [][]string{{"2026-10-28T08:00:00+01:00"}, {"2026-10-21T08:00:00+02:00", "2026-11-04T08:00:00+01:00"}} {
		block.TeachingWeekParity = &parity
		got, err := block.Occurrences(calendar)
		if err != nil || len(got) != len(want) {
			t.Fatalf("parity %d: occurrences = %+v, error = %v", parity, got, err)
		}
		for i := range got {
			if got[i].Start.Format(time.RFC3339) != want[i] {
				t.Fatalf("parity %d: got %s, want %s", parity, got[i].Start, want[i])
			}
		}
	}
	invalidParity := 2
	for _, parity := range []*int{nil, &invalidParity} {
		block.TeachingWeekParity = parity
		if _, err := block.Occurrences(calendar); err == nil {
			t.Fatal("accepted missing or invalid parity")
		}
	}
	parity := 0
	block.TeachingWeekParity = &parity
	calendar.ParityAnchor = "bad"
	if _, err := block.Occurrences(calendar); err == nil {
		t.Fatal("accepted invalid parity anchor")
	}
}
