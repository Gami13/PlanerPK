package utils

import (
	"testing"
	"time"
)

func TestCombineDateAndTime(t *testing.T) {
	zone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	clock := time.Date(2000, 1, 1, 10, 30, 15, 123, time.UTC)
	for _, tt := range []struct {
		day, offset int
	}{
		{19, 2 * 60 * 60},
		{26, 60 * 60},
	} {
		date := time.Date(2026, 10, tt.day, 23, 45, 0, 0, zone)
		got := CombineDateAndTime(date, clock)
		_, offset := got.Zone()
		if got.Year() != 2026 || got.Month() != time.October || got.Day() != tt.day || got.Location() != zone || offset != tt.offset {
			t.Errorf("date or timezone changed: %v", got)
		}
		if got.Hour() != 10 || got.Minute() != 30 || got.Second() != 15 || got.Nanosecond() != 123 {
			t.Errorf("clock not preserved: %v", got)
		}
	}
}
