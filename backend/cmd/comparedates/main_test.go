package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
)

func TestWriteComparison(t *testing.T) {
	previous := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = previous })
	day, start, date := "WED", 480, "2026-10-07"
	block := snapshot.Block{ID: "dated", Day: &day, Start: &start, Duration: 90, Frequency: "konkretna_data", Date: &date}
	weekly := block
	weekly.Frequency = "co_tydzien"
	class := types.Class{Name: "Subject", Type: "Lecture", Audience: "GL2", AdditionalInfo: "Long scheduling note", Day: time.Wednesday, SnapshotBlockID: block.ID,
		StartTime: time.Date(2000, 1, 1, 8, 0, 0, 0, time.UTC), EndTime: time.Date(2000, 1, 1, 9, 30, 0, 0, time.UTC)}
	calendar := snapshot.AcademicCalendar{Start: "2026-10-01", End: "2026-10-31"}
	for _, tt := range []struct {
		name, want string
		block      *snapshot.Block
		closed     bool
	}{
		{"dated", "Wed 2026-10-07 08:00 – 09:30", &block, false},
		{"weekly with closure", "Wed 2026-10-14 08:00 – 09:30\n    Wed 2026-10-21 08:00 – 09:30\n    Wed 2026-10-28 08:00 – 09:30", &weekly, true},
		{"closed", "omitted (no scheduled classes)", &block, true},
		{"unresolved", "unresolved (no unique snapshot match)", nil, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calendar.Closures = nil
			if tt.closed {
				calendar.Closures = []snapshot.DateRange{{From: date, To: date}}
			}
			var out bytes.Buffer
			if err := writeComparison(&out, "GL2", class, tt.block, calendar); err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("GL2 | Subject (Lecture)\n  Block: dated | Audience: GL2\n  Excel: Wednesday 08:00 – 09:30\n  Snapshot:\n    %s\n\n", tt.want)
			if out.String() != want {
				t.Fatalf("output = %q, want %q", out.String(), want)
			}
		})
	}
}
