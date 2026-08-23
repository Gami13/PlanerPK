package extractor

import (
	"strings"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legend"
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

func TestParseTimetableTimeUsesDisplayedValueOnly(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		want       string
		wantParsed bool
	}{
		{name: "two digit hour", value: "07:30", want: "07:30", wantParsed: true},
		{name: "single digit hour", value: "7:30", want: "07:30", wantParsed: true},
		{name: "dot separator", value: "08.15", want: "08:15", wantParsed: true},
		{name: "fractional Excel serial", value: "0.3125", wantParsed: false},
		{name: "decimal Excel clock", value: "7.3", wantParsed: false},
		{name: "invalid clock", value: "25:00", wantParsed: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, ok := parseTimetableTime(test.value)
			if ok != test.wantParsed {
				t.Fatalf("parseTimetableTime() ok = %v, want %v", ok, test.wantParsed)
			}
			if ok && parsed.Format("15:04") != test.want {
				t.Errorf("parseTimetableTime() = %s, want %s", parsed.Format("15:04"), test.want)
			}
		})
	}
}

func TestScanTimetableColumnSplitsUnexpectedPositiveGaps(t *testing.T) {
	rows := [][]string{{"07:30"}, {"07:45"}, {"08:15"}, {"08:30"}}

	runs := scanTimetableColumn(rows, 0)
	if len(runs) != 2 {
		t.Fatalf("scanTimetableColumn() returned %d runs, want 2", len(runs))
	}
	if len(runs[0].values) != 2 || len(runs[1].values) != 2 {
		t.Errorf("run lengths = %d and %d, want 2 and 2", len(runs[0].values), len(runs[1].values))
	}
}

func TestScanTimetableColumnSplitsAtNonIncreasingTime(t *testing.T) {
	rows := [][]string{{"07:30"}, {"07:45"}, {"07:30"}, {"07:45"}}

	runs := scanTimetableColumn(rows, 0)
	if len(runs) != 2 {
		t.Fatalf("scanTimetableColumn() returned %d runs, want 2", len(runs))
	}
	if len(runs[0].values) != 2 || len(runs[1].values) != 2 {
		t.Errorf("day run lengths = %d and %d, want 2 and 2", len(runs[0].values), len(runs[1].values))
	}
}

func TestExtractTimetableDaysSkipsIncompleteRuns(t *testing.T) {
	tests := []struct {
		name             string
		runs             []timetableDayRun
		wantDays         int
		wantSlotDuration time.Duration
		wantOK           bool
	}{
		{
			name: "no complete day run",
			runs: []timetableDayRun{
				testTimetableDayRun(0, 1, 15*time.Minute, 450),
			},
		},
		{
			name: "first complete run",
			runs: []timetableDayRun{
				testTimetableDayRun(0, 1, 15*time.Minute, 450),
				testTimetableDayRun(5, 2, 30*time.Minute, 450, 480),
			},
			wantDays:         1,
			wantSlotDuration: 15 * time.Minute,
			wantOK:           true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			days, slotDuration, ok := extractTimetableDays(test.runs)
			if ok != test.wantOK {
				t.Fatalf("extractTimetableDays() ok = %v, want %v", ok, test.wantOK)
			}
			if len(days) != test.wantDays {
				t.Fatalf("extractTimetableDays() returned %d days, want %d", len(days), test.wantDays)
			}
			if slotDuration != test.wantSlotDuration {
				t.Errorf("extractTimetableDays() slot duration = %s, want %s", slotDuration, test.wantSlotDuration)
			}
		})
	}
}

func TestExtractTimetableDaysFiltersIncompatibleRuns(t *testing.T) {
	runs := []timetableDayRun{
		testTimetableDayRun(0, 1, 15*time.Minute, 450, 465, 480),
		testTimetableDayRun(3, 1, 15*time.Minute, 450, 465),
		testTimetableDayRun(10, 1, 30*time.Minute, 600, 630),
	}

	days, slotDuration, ok := extractTimetableDays(runs)
	if !ok {
		t.Fatal("extractTimetableDays() ok = false, want true")
	}
	if slotDuration != 15*time.Minute {
		t.Errorf("extractTimetableDays() slot duration = %s, want 15m", slotDuration)
	}
	if len(days) != 2 {
		t.Fatalf("extractTimetableDays() returned %d days, want 2", len(days))
	}
	if days[0].WeekDay != time.Monday || days[1].WeekDay != time.Tuesday {
		t.Errorf("weekdays = %s and %s, want Monday and Tuesday", days[0].WeekDay, days[1].WeekDay)
	}
	if days[0].StartingRow != 0 || days[0].EndingRow != 2 || days[1].StartingRow != 3 || days[1].EndingRow != 4 {
		t.Errorf("day rows = %d-%d and %d-%d, want 0-2 and 3-4", days[0].StartingRow, days[0].EndingRow, days[1].StartingRow, days[1].EndingRow)
	}
}

func testTimetableDayRun(startRow, rowStride int, interval time.Duration, minutes ...int) timetableDayRun {
	values := make([]timetableTimeValue, len(minutes))
	for index, minuteOfDay := range minutes {
		clockTime := time.Date(0, time.January, 1, minuteOfDay/60, minuteOfDay%60, 0, 0, time.UTC)
		values[index] = timetableTimeValue{
			rowIndex:  startRow + index*rowStride,
			clockTime: clockTime,
		}
	}
	return timetableDayRun{
		values:    values,
		interval:  interval,
		rowStride: rowStride,
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
	timetable := BuildTimetable(groups, file, sheet, lg, buildings.BuildBuildingsFromLegend(lg), layout)
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
