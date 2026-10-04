package extractor

import (
	"strings"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

func setWorksheetValue(t *testing.T, file *excelize.File, sheet string, row, column int, value string) {
	t.Helper()
	cell, err := excelize.CoordinatesToCellName(column+1, row+1)
	if err != nil {
		t.Fatalf("CoordinatesToCellName() error = %v", err)
	}
	if err := file.SetCellValue(sheet, cell, value); err != nil {
		t.Fatalf("SetCellValue(%q) error = %v", cell, err)
	}
}

func setTimeSequence(t *testing.T, file *excelize.File, sheet string, column, startRow int, values ...string) {
	t.Helper()
	for row, value := range values {
		setWorksheetValue(t, file, sheet, startRow+row, column, value)
	}
}

func TestDiscoverTimetableLayoutDetectsShiftedColumnAndRepeatedDays(t *testing.T) {
	file := workbookWithSheets(t, "2025_2026")
	defer file.Close()

	setWorksheetValue(t, file, "2025_2026", 0, 0, "Header text")
	setTimeSequence(t, file, "2025_2026", 4, 3,
		"7:30", "07:45", "8.00", "08.15",
		"7.30", "07:45", "8:00",
		"07.30", "7:45")

	layout, err := DiscoverTimetableLayout(file, "2025_2026")
	if err != nil {
		t.Fatalf("DiscoverTimetableLayout() error = %v", err)
	}

	if layout.HourColumn != 4 {
		t.Fatalf("HourColumn = %d, want 4", layout.HourColumn)
	}
	if layout.SlotDuration != 15*time.Minute {
		t.Fatalf("SlotDuration = %s, want 15m", layout.SlotDuration)
	}
	if len(layout.Days) != 3 {
		t.Fatalf("day count = %d, want 3", len(layout.Days))
	}

	wantDays := []struct {
		startRow int
		endRow   int
		weekday  time.Weekday
		hour     int
		minute   int
	}{
		{startRow: 3, endRow: 6, weekday: time.Monday, hour: 7, minute: 30},
		{startRow: 7, endRow: 9, weekday: time.Tuesday, hour: 7, minute: 30},
		{startRow: 10, endRow: 11, weekday: time.Wednesday, hour: 7, minute: 30},
	}
	for index, want := range wantDays {
		day := layout.Days[index]
		if day.StartingRow != want.startRow || day.EndingRow != want.endRow || day.WeekDay != want.weekday {
			t.Errorf("day %d = %+v, want rows %d-%d on %s", index, day, want.startRow, want.endRow, want.weekday)
		}
		if day.StartingTime.Hour() != want.hour || day.StartingTime.Minute() != want.minute {
			t.Errorf("day %d starting time = %s, want %02d:%02d", index, day.StartingTime.Format("15:04"), want.hour, want.minute)
		}
	}
}

func TestDiscoverTimetableLayoutInfersThirtyMinuteSlots(t *testing.T) {
	file := workbookWithSheets(t, "2025_2026")
	defer file.Close()

	setTimeSequence(t, file, "2025_2026", 1, 2, "7.30", "8.00", "8.30", "9.00")

	layout, err := DiscoverTimetableLayout(file, "2025_2026")
	if err != nil {
		t.Fatalf("DiscoverTimetableLayout() error = %v", err)
	}
	if layout.SlotDuration != 30*time.Minute {
		t.Fatalf("SlotDuration = %s, want 30m", layout.SlotDuration)
	}
}

func TestDiscoverTimetableLayoutUsesWeekdayLabelsAcrossIrregularTimes(t *testing.T) {
	file := workbookWithSheets(t, "2026_2027")
	defer file.Close()
	const sheet = "2026_2027"

	setWorksheetValue(t, file, sheet, 1, 2, "PONIEDZIAŁEK")
	setTimeSequence(t, file, sheet, 1, 2, "07:30", "07:45", "08:00")
	setWorksheetValue(t, file, sheet, 5, 2, "WTOREK")
	setTimeSequence(t, file, sheet, 1, 6, "07:30", "07:45", "08:00")
	setWorksheetValue(t, file, sheet, 9, 2, "ŚRODA")
	setTimeSequence(t, file, sheet, 1, 10, "07:30", "07:45", "07:55", "08:00", "08:15")

	layout, err := DiscoverTimetableLayout(file, sheet)
	if err != nil {
		t.Fatal(err)
	}
	if len(layout.Days) != 3 {
		t.Fatalf("day count = %d, want 3", len(layout.Days))
	}
	want := []struct {
		weekday           time.Weekday
		firstRow, lastRow int
	}{
		{time.Monday, 2, 4},
		{time.Tuesday, 6, 8},
		{time.Wednesday, 10, 14},
	}
	for i, expected := range want {
		day := layout.Days[i]
		if day.WeekDay != expected.weekday || day.StartingRow != expected.firstRow || day.EndingRow != expected.lastRow {
			t.Errorf("day %d = %+v, want %s rows %d-%d", i, day, expected.weekday, expected.firstRow, expected.lastRow)
		}
	}
}

func TestDiscoverTimetableLayoutIgnoresBlankRowsBetweenTimeLabels(t *testing.T) {
	file := workbookWithSheets(t, "2025_2026")
	defer file.Close()

	for row, value := range []string{"7.30", "8.00", "8.30", "9.00"} {
		setWorksheetValue(t, file, "2025_2026", 2+row*2, 1, value)
	}

	layout, err := DiscoverTimetableLayout(file, "2025_2026")
	if err != nil {
		t.Fatalf("DiscoverTimetableLayout() error = %v", err)
	}
	if layout.SlotDuration != 15*time.Minute {
		t.Fatalf("SlotDuration = %s, want 15m per worksheet row", layout.SlotDuration)
	}
	if layout.Days[0].StartingRow != 2 || layout.Days[0].EndingRow != 8 {
		t.Errorf("day rows = %d-%d, want 2-8", layout.Days[0].StartingRow, layout.Days[0].EndingRow)
	}
}

func TestDiscoverTimetableLayoutIgnoresNumericExcelTimeValues(t *testing.T) {
	file := workbookWithSheets(t, "2025_2026")
	defer file.Close()

	const sheet = "2025_2026"
	for row, value := range []float64{7.5 / 24, 8.0 / 24, 8.5 / 24, 9.0 / 24} {
		cell, err := excelize.CoordinatesToCellName(2, row+1)
		if err != nil {
			t.Fatalf("CoordinatesToCellName() error = %v", err)
		}
		if err := file.SetCellValue(sheet, cell, value); err != nil {
			t.Fatalf("SetCellValue(%q) error = %v", cell, err)
		}
	}

	_, err := DiscoverTimetableLayout(file, sheet)
	if err == nil {
		t.Fatal("DiscoverTimetableLayout() error = nil, want numeric Excel values to be ignored")
	}
	if !strings.Contains(err.Error(), "no column") {
		t.Errorf("error = %q, want no-column detection error", err)
	}
}

func TestDiscoverTimetableLayoutSelectsLeftmostValidCandidate(t *testing.T) {
	file := workbookWithSheets(t, "2025_2026")
	defer file.Close()

	setTimeSequence(t, file, "2025_2026", 1, 2, "08:00", "08:15")
	setTimeSequence(t, file, "2025_2026", 3, 2, "07.30", "07.45", "08.00")

	layout, err := DiscoverTimetableLayout(file, "2025_2026")
	if err != nil {
		t.Fatalf("DiscoverTimetableLayout() error = %v", err)
	}
	if layout.HourColumn != 1 {
		t.Errorf("HourColumn = %d, want leftmost candidate 1", layout.HourColumn)
	}
}

func TestDiscoverTimetableLayoutIgnoresNoiseMalformedTimesAndIsolatedCells(t *testing.T) {
	file := workbookWithSheets(t, "2025_2026")
	defer file.Close()

	setTimeSequence(t, file, "2025_2026", 0, 0, "7:30")
	setTimeSequence(t, file, "2025_2026", 1, 0, "25:00", "7:5")
	setWorksheetValue(t, file, "2025_2026", 2, 0, " ")
	setTimeSequence(t, file, "2025_2026", 3, 4, "07:30", "07:45", "08:00")

	layout, err := DiscoverTimetableLayout(file, "2025_2026")
	if err != nil {
		t.Fatalf("DiscoverTimetableLayout() error = %v", err)
	}
	if layout.HourColumn != 3 {
		t.Errorf("HourColumn = %d, want 3", layout.HourColumn)
	}
}

func TestDiscoverTimetableLayoutReportsSheetAndDetectionDetailsOnFailure(t *testing.T) {
	file := workbookWithSheets(t, "No timetable")
	defer file.Close()
	setWorksheetValue(t, file, "No timetable", 0, 0, "7:30")

	_, err := DiscoverTimetableLayout(file, "No timetable")
	if err == nil {
		t.Fatal("DiscoverTimetableLayout() error = nil, want detection error")
	}
	for _, want := range []string{"No timetable", "no column", "consistent interval"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
}

func TestBuildExtractionUsesDetectedLayout(t *testing.T) {
	file := workbookWithSheets(t, "2025_2026")
	defer file.Close()

	const sheet = "2025_2026"
	setWorksheetValue(t, file, sheet, 0, 2, "Category")
	setWorksheetValue(t, file, sheet, 1, 2, "Group A")
	setWorksheetValue(t, file, sheet, 0, 8, "Unrelated worksheet data")
	setTimeSequence(t, file, sheet, 1, 2, "07.30", "08.00")
	setWorksheetValue(t, file, sheet, 2, 2, "Mathematics\nwykład")

	layout, err := DiscoverTimetableLayout(file, sheet)
	if err != nil {
		t.Fatalf("DiscoverTimetableLayout() error = %v", err)
	}

	if layout.DataEndColumn != 2 {
		t.Errorf("DataEndColumn = %d, want 2", layout.DataEndColumn)
	}
	if layout.HeaderRow != 1 {
		t.Errorf("HeaderRow = %d, want 1", layout.HeaderRow)
	}

	groups, err := BuildAvailableGroups(file, sheet, layout)
	if err != nil {
		t.Fatalf("BuildAvailableGroups() error = %v", err)
	}
	if len(groups) != 1 || groups[0].Column != 2 {
		t.Fatalf("groups = %+v, want one group in column 2", groups)
	}

	lg := &legend.Legend{
		Teachers:  map[string]legend.Teacher{},
		Locations: map[string]legend.Location{},
	}
	timetable := BuildTimetable(groups, file, sheet, lg, buildings.BuildBuildingsFromLegend(lg), layout, nil)
	classes := timetable[2][time.Monday]
	if len(classes) != 1 {
		t.Fatalf("Monday classes = %+v, want one class", classes)
	}
	if classes[0].StartTime.Hour() != 7 || classes[0].StartTime.Minute() != 30 {
		t.Errorf("class start = %s, want 07:30", classes[0].StartTime.Format("15:04"))
	}
	if classes[0].EndTime.Hour() != 8 || classes[0].EndTime.Minute() != 0 {
		t.Errorf("class end = %s, want 08:00", classes[0].EndTime.Format("15:04"))
	}
	if classes[0].Name != "Mathematics" {
		t.Errorf("class name = %q, want Mathematics", classes[0].Name)
	}
}

func TestDiscoverTimetableLayoutUsesGLGroupRowInsteadOfWeekdayRow(t *testing.T) {
	file := workbookWithSheets(t, "2026_2027")
	defer file.Close()

	const sheet = "2026_2027"
	setWorksheetValue(t, file, sheet, 0, 2, "I st. sem. 1")
	setWorksheetValue(t, file, sheet, 1, 2, "GL1")
	setWorksheetValue(t, file, sheet, 2, 2, "PONIEDZIAŁEK")
	setTimeSequence(t, file, sheet, 1, 3, "07:30", "08:00")
	setWorksheetValue(t, file, sheet, 3, 2, "Matematyka\nwykład")

	layout, err := DiscoverTimetableLayout(file, sheet)
	if err != nil {
		t.Fatalf("DiscoverTimetableLayout() error = %v", err)
	}
	if layout.HeaderRow != 1 {
		t.Fatalf("HeaderRow = %d, want GL group row 1", layout.HeaderRow)
	}

	groups, err := BuildAvailableGroups(file, sheet, layout)
	if err != nil {
		t.Fatalf("BuildAvailableGroups() error = %v", err)
	}
	if len(groups) != 1 || groups[0].Label != "GL1" || groups[0].Category != "I st. sem. 1" {
		t.Errorf("groups = %+v, want GL1 in category I st. sem. 1", groups)
	}
}

func TestBuildTimetableAddsMatchingSupplementalColumns(t *testing.T) {
	file := workbookWithSheets(t, "2026_2027")
	defer file.Close()

	const sheet = "2026_2027"
	setWorksheetValue(t, file, sheet, 2, 2, "Core subject\nwykład")
	setWorksheetValue(t, file, sheet, 2, 3, "Shared elective\nwykład")
	setWorksheetValue(t, file, sheet, 2, 4, "Shared elective\nwykład")
	setWorksheetValue(t, file, sheet, 2, 5, "Other semester subject\nwykład")
	setWorksheetValue(t, file, sheet, 2, 6, "Other semester language\nwykład")

	groups := []types.ColumnGroup{
		{Column: 3, Label: "Językowe (mieszane)", Category: "I st. sem. 3"},
		{Column: 4, Label: "Obieralne (osobne grupy)", Category: "I st. sem. 3"},
		{Column: 6, Label: "Językowe (mieszane)", Category: "I st. sem. 5"},
		{Column: 2, Label: "GL1", Category: "I st. sem. 3"},
		{Column: 5, Label: "GL1", Category: "I st. sem. 5"},
	}
	layout := TimetableLayout{
		Days: []types.Day{{
			StartingRow:  2,
			EndingRow:    2,
			WeekDay:      time.Monday,
			StartingTime: time.Date(2026, 9, 28, 8, 0, 0, 0, time.UTC),
		}},
		SlotDuration: 30 * time.Minute,
	}
	lg := &legend.Legend{Teachers: map[string]legend.Teacher{}, Locations: map[string]legend.Location{}}
	timetable := BuildTimetable(groups, file, sheet, lg, nil, layout, nil)

	semesterThree := timetable[2][time.Monday]
	if len(semesterThree) != 2 {
		t.Fatalf("semester 3 GL1 classes = %+v, want core plus one deduplicated supplemental class", semesterThree)
	}
	if semesterThree[0].Name != "Core subject" || semesterThree[1].Name != "Shared elective" {
		t.Errorf("semester 3 GL1 classes = %+v, want core then shared elective", semesterThree)
	}
	semesterFive := timetable[5][time.Monday]
	if len(semesterFive) != 2 || semesterFive[1].Name != "Other semester language" {
		t.Errorf("semester 5 GL1 classes = %+v, want only matching language supplemental", semesterFive)
	}

	selectable := SelectableGroups(groups)
	if len(selectable) != 2 || selectable[0].Column != 2 || selectable[1].Column != 5 {
		t.Errorf("SelectableGroups() = %+v, want only regular groups", selectable)
	}
}

func TestDiscoverTimetableLayoutIncludesMergedHeaderSpan(t *testing.T) {
	file := workbookWithSheets(t, "2025_2026")
	defer file.Close()

	const sheet = "2025_2026"
	setWorksheetValue(t, file, sheet, 0, 2, "Category")
	setWorksheetValue(t, file, sheet, 1, 2, "Group A")
	if err := file.MergeCell(sheet, "C2", "E2"); err != nil {
		t.Fatalf("MergeCell() error = %v", err)
	}
	setTimeSequence(t, file, sheet, 1, 2, "07.30", "08.00")

	layout, err := DiscoverTimetableLayout(file, sheet)
	if err != nil {
		t.Fatalf("DiscoverTimetableLayout() error = %v", err)
	}
	if layout.DataEndColumn != 4 {
		t.Fatalf("DataEndColumn = %d, want 4", layout.DataEndColumn)
	}

	groups, err := BuildAvailableGroups(file, sheet, layout)
	if err != nil {
		t.Fatalf("BuildAvailableGroups() error = %v", err)
	}
	if len(groups) != 3 {
		t.Fatalf("groups = %+v, want three columns from merged header", groups)
	}
	for index, group := range groups {
		if group.Column != index+2 {
			t.Errorf("groups[%d].Column = %d, want %d", index, group.Column, index+2)
		}
	}
}
