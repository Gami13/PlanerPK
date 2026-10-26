package types

import "testing"

func TestCanImportGroup(t *testing.T) {
	groups := []ColumnGroup{
		{Column: 1, Category: "I st. sem. 1"}, {Column: 2, Category: "I st. sem. 1"},
		{Column: 3, Category: "I st. sem. 2"}, {Column: 4, Category: "I st. sem. 3"},
		{Column: 5, Category: "II st. sem. 1"}, {Column: 6, Category: "CP I st. sem. 1"},
	}
	for _, tt := range []struct {
		source int64
		want   bool
	}{{1, false}, {2, true}, {3, true}, {4, false}, {5, false}, {6, false}, {99, false}} {
		if got := CanImportGroup(groups, 1, tt.source); got != tt.want {
			t.Errorf("source %d: got %v, want %v", tt.source, got, tt.want)
		}
	}
}

func TestAcademicYearKeepsSpecializationsSeparate(t *testing.T) {
	for _, tt := range []struct {
		a, b string
		want bool
	}{
		{"II st. sem. 1 CY", "II st. sem. 2 CY", true},
		{"II st. sem. 2 CY", "II st. sem. 2 DS", false},
		{"I st. sem. 3", "I st. sem. 4", true},
		{"ERASMUS", "ERASMUS", false},
	} {
		if got := sameAcademicYear(tt.a, tt.b); got != tt.want {
			t.Errorf("%s / %s: got %v", tt.a, tt.b, got)
		}
	}
}
