package extractor

import (
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
)

var (
	testStartTime = time.Date(2026, 1, 5, 8, 0, 0, 0, time.UTC)
	testEndTime   = time.Date(2026, 1, 5, 9, 30, 0, 0, time.UTC)
)

const af40Text = `Przetwarzanie języka naturalnego — L
Książek Wojciech
grupa obieralna 2
15:30–17:45 · s. L5
co tydz.

Programowanie mobilne — L
Olczyk Mateusz
grupa obieralna 1
16:30–18:45 · s. L2
co tydz.

Przetwarzanie języka naturalnego — L
Książek Wojciech
grupa obieralna 3
18:15–20:30 · s. L5
co tydz.

Programowanie mobilne — L
Olczyk Mateusz
grupa obieralna 2
19:00–21:15 · s. L4
co tydz.`

func TestParseStructuredClassTextAF40(t *testing.T) {
	day := types.Day{WeekDay: time.Monday, StartingTime: testStartTime}
	classes := parseStructuredClassText(af40Text, day)
	if len(classes) != 4 {
		t.Fatalf("got %d classes, want 4", len(classes))
	}

	wantStarts := []string{"15:30", "16:30", "18:15", "19:00"}
	wantAudiences := []string{"grupa obieralna 2", "grupa obieralna 1", "grupa obieralna 3", "grupa obieralna 2"}
	for i, class := range classes {
		if got := class.StartTime.Format("15:04"); got != wantStarts[i] {
			t.Errorf("class %d start = %s, want %s", i, got, wantStarts[i])
		}
		if class.Audience != wantAudiences[i] {
			t.Errorf("class %d audience = %q, want %q", i, class.Audience, wantAudiences[i])
		}
		if class.Type != "L" || class.AdditionalInfo != "co tydz." {
			t.Errorf("class %d = %+v, want L with raw cadence", i, class)
		}
	}
	if classes[0].EndTime.Format("15:04") != "17:45" || classes[3].Location != "s. L4" {
		t.Errorf("unexpected boundary fields: first end %s, last location %q", classes[0].EndTime.Format("15:04"), classes[3].Location)
	}
}

func TestParseStructuredClassTextOptionalAndAdditionalNotes(t *testing.T) {
	day := types.Day{WeekDay: time.Wednesday, StartingTime: testStartTime}
	text := "Subject — W/C\nProfessor\nother audience\n07:30-09:00 · ONLINE\n\n" +
		"Other subject — P\nProfessor Two\nGL1\n09:15–10:45 · s. L2\nbez 07.10\nPoczątek: 14.10"

	classes := parseStructuredClassText(text, day)
	if len(classes) != 2 {
		t.Fatalf("got %d classes, want 2", len(classes))
	}
	if classes[0].AdditionalInfo != "" {
		t.Errorf("missing note parsed as %q", classes[0].AdditionalInfo)
	}
	if classes[1].AdditionalInfo != "bez 07.10\nPoczątek: 14.10" {
		t.Errorf("additional info = %q, want both raw note lines", classes[1].AdditionalInfo)
	}
}

func TestParseStructuredClassTextSkipsMalformedRecords(t *testing.T) {
	day := types.Day{WeekDay: time.Monday, StartingTime: testStartTime}
	text := "Missing fields — L\nProfessor\nGL1\nnot a time\nco tydz.\n\n" +
		"Valid — C\nProfessor\nC1 (GL1+GL2)\n11:00–12:30 · s. P1\nA"

	classes := parseStructuredClassText(text, day)
	if len(classes) != 1 || classes[0].Name != "Valid" {
		t.Fatalf("classes = %+v, want only the valid record", classes)
	}
}

func TestBuildTimetableStructuredMergedCellDoesNotMergeRecords(t *testing.T) {
	file := workbookWithSheets(t, "2026_2027")
	defer file.Close()
	setWorksheetValue(t, file, "2026_2027", 0, 1, af40Text)
	if err := file.MergeCell("2026_2027", "B1", "B24"); err != nil {
		t.Fatal(err)
	}

	layout := TimetableLayout{
		Days:         []types.Day{{StartingRow: 0, EndingRow: 23, WeekDay: time.Monday, StartingTime: time.Date(2026, 9, 28, 15, 30, 0, 0, time.UTC)}},
		SlotDuration: 15 * time.Minute,
	}
	timetable := BuildTimetable([]types.ColumnGroup{{Column: 1, Label: "Obieralne"}}, file, "2026_2027", nil, nil, layout, nil)
	classes := timetable[1][time.Monday]
	if len(classes) != 4 {
		t.Fatalf("merged cell produced %d classes, want 4: %+v", len(classes), classes)
	}
	if classes[1].Name != "Programowanie mobilne" || classes[1].StartTime.Format("15:04") != "16:30" {
		t.Errorf("second class = %+v, want independent 16:30 record", classes[1])
	}
}

func TestClassHashIncludesStructuredAudience(t *testing.T) {
	class := types.Class{Name: "Subject", Day: time.Monday, StartTime: testStartTime, EndTime: testEndTime}
	first, second := class, class
	first.Audience = "GL1"
	second.Audience = "GL2"
	if first.Hash() == second.Hash() {
		t.Error("classes for different audiences have identical hashes")
	}
}

func TestStructuredClassNameWithSlashIsNotLegacySplit(t *testing.T) {
	day := types.Day{WeekDay: time.Monday, StartingTime: testStartTime}
	classes := finalizeClasses(parseStructuredClassText(
		"Subject A/Subject B — W\nProfessor\nGL1\n08:00–09:30 · ONLINE",
		day,
	), nil, nil)
	if len(classes) != 1 || classes[0].Name != "Subject A/Subject B" {
		t.Fatalf("classes = %+v, want one unsplit structured class", classes)
	}
}
