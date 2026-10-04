package snapshot

import (
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
)

func TestCandidates(t *testing.T) {
	day, start := "MON", 765
	block := Block{ID: "b1", Subject: "Subject", Activity: "W", Day: &day, Start: &start}
	index := NewSubjectIndex([]Block{block})
	class := types.Class{Name: "Subject", Type: "W", Day: time.Monday, StartTime: time.Date(2026, 10, 5, 12, 45, 0, 0, time.UTC)}
	for _, tt := range []struct {
		name string
		edit func(*types.Class)
		want int
	}{
		{"exact slot", func(c *types.Class) {}, 1},
		{"different subject", func(c *types.Class) { c.Name = "subject" }, 0},
		{"different activity", func(c *types.Class) { c.Type = "L" }, 0},
		{"different day", func(c *types.Class) { c.Day = time.Tuesday }, 0},
		{"different start", func(c *types.Class) { c.StartTime = c.StartTime.Add(time.Minute) }, 0},
		{"unsupported day", func(c *types.Class) { c.Day = time.Sunday }, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := class
			tt.edit(&input)
			if got := index.Candidates(input); len(got) != tt.want {
				t.Fatalf("candidates = %v; want %d", got, tt.want)
			}
		})
	}
	other := block
	other.ID, other.Frequency = "b2", "co_2_tygodnie"
	noDay, noStart := block, block
	noDay.ID, noDay.Day = "no-day", nil
	noStart.ID, noStart.Start = "no-start", nil
	index = NewSubjectIndex([]Block{block, other, noDay, noStart})
	got := index.Candidates(class)
	if len(got) != 2 || got[0].ID != "b1" || got[1].ID != "b2" {
		t.Fatalf("must retain ambiguity and exclude parked blocks: %+v", got)
	}
}
