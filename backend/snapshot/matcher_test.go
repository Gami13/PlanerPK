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

func TestCandidatesSubjectFallback(t *testing.T) {
	day, start := "MON", 765
	class := types.Class{Name: "Full subject", Type: "W", Teacher: "Teacher", Day: time.Monday, StartTime: time.Date(2026, 10, 5, 12, 45, 0, 0, time.UTC)}
	short := Block{ID: "short", Subject: "Short subject", Activity: "W", Teacher: "Teacher", Day: &day, Start: &start}
	exact := short
	exact.ID, exact.Subject = "exact", class.Name
	for _, tt := range []struct {
		name   string
		blocks []Block
		want   string
	}{
		{"missing subject", []Block{short}, "short"},
		{"exact takes priority", []Block{short, exact}, "exact"},
	} {
		got := NewSubjectIndex(tt.blocks).Candidates(class)
		if len(got) != 1 || got[0].ID != tt.want {
			t.Fatalf("%s: candidates = %+v", tt.name, got)
		}
	}
	exact.Teacher = "Other teacher"
	if got := NewSubjectIndex([]Block{exact, short}).Candidates(class); len(got) != 1 || got[0].ID != "short" {
		t.Fatalf("failed exact match did not use fallback: %+v", got)
	}
	other := short
	other.ID, other.Subject = "other", "Another subject"
	if got := NewSubjectIndex([]Block{short, other}).Candidates(class); len(got) != 2 {
		t.Fatalf("fallback must retain ambiguity: %+v", got)
	}
	other.Activity = "L"
	if got := NewSubjectIndex([]Block{other}).Candidates(class); len(got) != 0 {
		t.Fatalf("fallback ignored activity mismatch: %+v", got)
	}
	other.Activity = "W"
	short.Teacher = ""
	if got := NewSubjectIndex([]Block{short}).Candidates(class); len(got) != 0 {
		t.Fatalf("fallback without JSON teacher: %+v", got)
	}
	class.Teacher = ""
	if got := NewSubjectIndex([]Block{other}).Candidates(class); len(got) != 0 {
		t.Fatalf("fallback without Excel teacher: %+v", got)
	}
}

func TestCandidatesCompareMastersExerciseAudiences(t *testing.T) {
	day, start := "THU", 660
	class := types.Class{Name: "Psychologia motywacji", Type: "C", Day: time.Thursday, StartTime: time.Date(2026, 10, 8, 11, 0, 0, 0, time.UTC)}
	blocks := []Block{
		{ID: "cy1", Subject: class.Name, Activity: class.Type, Day: &day, Start: &start, AcademicGroupLabel: "CY_C1"},
		{ID: "cy2", Subject: class.Name, Activity: class.Type, Day: &day, Start: &start, AcademicGroupLabel: "CY_C2"},
		{ID: "ds1", Subject: class.Name, Activity: class.Type, Day: &day, Start: &start, AcademicGroupLabel: "DS_C1"},
	}
	for _, tt := range []struct{ audience, id string }{{"CY_C1", "cy1"}, {"CY_C2", "cy2"}, {"DS_C1", "ds1"}} {
		class.Audience = tt.audience
		got := NewSubjectIndex(blocks).Candidates(class)
		if len(got) != 1 || got[0].ID != tt.id {
			t.Fatalf("audience %q: candidates = %+v", tt.audience, got)
		}
	}
	blocks[1].AcademicGroupLabel = ""
	class.Audience = "CY_C1"
	if got := NewSubjectIndex(blocks).Candidates(class); len(got) != 2 {
		t.Fatalf("missing academic group label must retain ambiguity: %+v", got)
	}
}

func TestCandidatesCompareSchedulingNotes(t *testing.T) {
	day, start, a, b := "MON", 765, 0, 1
	firstDate, secondDate := "2026-10-05", "2026-10-12"
	class := types.Class{Name: "Subject", Type: "P", Day: time.Monday, StartTime: time.Date(2026, 10, 5, 12, 45, 0, 0, time.UTC)}
	blocks := []Block{
		{ID: "a", Subject: class.Name, Activity: class.Type, Day: &day, Start: &start, Frequency: "co_2_tygodnie", TeachingWeekParity: &a},
		{ID: "b", Subject: class.Name, Activity: class.Type, Day: &day, Start: &start, Frequency: "co_2_tygodnie", TeachingWeekParity: &b},
	}
	for _, tt := range []struct{ note, id string }{{"A", "a"}, {"B", "b"}, {"A\nStart: 12.10", "a"}} {
		class.AdditionalInfo = tt.note
		got := NewSubjectIndex(blocks).Candidates(class)
		if len(got) != 1 || got[0].ID != tt.id {
			t.Fatalf("note %q: candidates = %+v", tt.note, got)
		}
	}
	blocks[0].Frequency, blocks[1].Frequency = "konkretna_data", "konkretna_data"
	blocks[0].Date, blocks[1].Date = &firstDate, &secondDate
	for _, tt := range []struct{ note, id string }{{"05.10.2026", "a"}, {"12.10.2026", "b"}} {
		class.AdditionalInfo = tt.note
		got := NewSubjectIndex(blocks).Candidates(class)
		if len(got) != 1 || got[0].ID != tt.id {
			t.Fatalf("date %q: candidates = %+v", tt.note, got)
		}
	}
	for _, note := range []string{"Start: 12.10", "31.02.2026", "co tydz.", ""} {
		class.AdditionalInfo = note
		if got := NewSubjectIndex(blocks).Candidates(class); len(got) != 2 {
			t.Fatalf("uninterpreted note %q discarded candidates: %+v", note, got)
		}
	}
	blocks[0].TeachingWeekParity, blocks[1].TeachingWeekParity = nil, nil
	blocks[0].Date, blocks[1].Date = nil, nil
	for _, note := range []string{"A", "05.10.2026"} {
		class.AdditionalInfo = note
		if got := NewSubjectIndex(blocks).Candidates(class); len(got) != 2 {
			t.Fatalf("missing scheduling metadata must retain ambiguity: %+v", got)
		}
	}
}

func TestCandidatesCompareTeachers(t *testing.T) {
	day, start := "MON", 870
	class := types.Class{Name: "Język obcy", Type: "C", Day: time.Monday, StartTime: time.Date(2026, 10, 5, 14, 30, 0, 0, time.UTC)}
	for _, tt := range []struct {
		name, excel, teacher, display string
		want                          int
	}{
		{"same teacher", "Elżbieta Szabat", "Elżbieta Szabat", "", 1},
		{"different teacher", "Aleksandra Zając", "Elżbieta Szabat", "", 0},
		{"display name", "Elżbieta Szabat", "Szabat Elżbieta", "Elżbieta Szabat", 1},
		{"display only", "Elżbieta Szabat", "", "Elżbieta Szabat", 1},
		{"trimmed name", " Elżbieta Szabat ", "Elżbieta Szabat ", "", 1},
		{"missing Excel name", "", "Elżbieta Szabat", "", 1},
		{"missing JSON names", "Elżbieta Szabat", "", "", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			class.Teacher = tt.excel
			index := NewSubjectIndex([]Block{{ID: "b1", Subject: class.Name, Activity: class.Type, Day: &day, Start: &start, Teacher: tt.teacher, TeacherDisplay: tt.display}})
			if got := index.Candidates(class); len(got) != tt.want {
				t.Fatalf("candidates = %+v; want %d", got, tt.want)
			}
		})
	}
	class.Teacher = "Elżbieta Szabat"
	blocks := []Block{
		{ID: "b1", Subject: class.Name, Activity: class.Type, Day: &day, Start: &start, Teacher: class.Teacher},
		{ID: "b2", Subject: class.Name, Activity: class.Type, Day: &day, Start: &start},
	}
	if got := NewSubjectIndex(blocks).Candidates(class); len(got) != 2 {
		t.Fatalf("missing teacher must retain ambiguity: %+v", got)
	}
}

func TestCandidatesCompareKnownAudiences(t *testing.T) {
	day, start, first, second := "MON", 765, 1, 2
	blocks := []Block{
		{ID: "g1", Subject: "Subject", Activity: "C", Day: &day, Start: &start, GroupNo: &first},
		{ID: "g2", Subject: "Subject", Activity: "C", Day: &day, Start: &start, GroupNo: &second},
	}
	class := types.Class{Name: "Subject", Type: "C", Day: time.Monday, StartTime: time.Date(2026, 10, 5, 12, 45, 0, 0, time.UTC)}
	for _, audience := range []string{"GL1", "C1 (GL1+GL2)", "gr. 1", "grupa obieralna 1"} {
		class.Audience = audience
		got := NewSubjectIndex(blocks).Candidates(class)
		if len(got) != 1 || got[0].ID != "g1" {
			t.Fatalf("audience %q: candidates = %+v", audience, got)
		}
	}
	for _, audience := range []string{"", "CY_GL1", "SIRR_GL1", "cały rocznik", "grupa językowa J1 (mieszana)"} {
		class.Audience = audience
		if got := NewSubjectIndex(blocks).Candidates(class); len(got) != 2 {
			t.Fatalf("unsupported audience %q discarded candidates: %+v", audience, got)
		}
	}
	class.Audience = "GL1"
	blocks[0].GroupNo, blocks[1].GroupNo = nil, nil
	blocks[0].Cohort, blocks[1].Cohort = "I stopień stac sem. 1 / gr. 1", "I stopień stac sem. 1 / gr. 2"
	if got := NewSubjectIndex(blocks).Candidates(class); len(got) != 1 || got[0].ID != "g1" {
		t.Fatalf("cohort group fallback = %+v", got)
	}
	blocks[1].Cohort = ""
	if got := NewSubjectIndex(blocks).Candidates(class); len(got) != 2 {
		t.Fatalf("missing audience metadata must retain ambiguity: %+v", got)
	}
}
