package legacy

import (
	"testing"
	"time"
)

func TestScanTimetableColumnSplitsUnexpectedPositiveGaps(t *testing.T) {
	runs := scanTimetableColumn([][]string{{"07:30"}, {"07:45"}, {"08:15"}, {"08:30"}}, 0)
	if len(runs) != 2 || len(runs[0].values) != 2 || len(runs[1].values) != 2 {
		t.Fatalf("run lengths = %+v, want two runs of two", runs)
	}
}

func TestScanTimetableColumnSplitsAtNonIncreasingTime(t *testing.T) {
	runs := scanTimetableColumn([][]string{{"07:30"}, {"07:45"}, {"07:30"}, {"07:45"}}, 0)
	if len(runs) != 2 || len(runs[0].values) != 2 || len(runs[1].values) != 2 {
		t.Fatalf("run lengths = %+v, want two runs of two", runs)
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
			runs: []timetableDayRun{testTimetableDayRun(0, 1, 15*time.Minute, 450)},
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
			if ok != test.wantOK || len(days) != test.wantDays || slotDuration != test.wantSlotDuration {
				t.Errorf("got days=%d slot=%s ok=%v, want days=%d slot=%s ok=%v", len(days), slotDuration, ok, test.wantDays, test.wantSlotDuration, test.wantOK)
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
	if !ok || slotDuration != 15*time.Minute || len(days) != 2 {
		t.Fatalf("got days=%d slot=%s ok=%v, want two days at 15m", len(days), slotDuration, ok)
	}
	if days[0].WeekDay != time.Monday || days[1].WeekDay != time.Tuesday ||
		days[0].StartingRow != 0 || days[0].EndingRow != 2 || days[1].StartingRow != 3 || days[1].EndingRow != 4 {
		t.Errorf("unexpected days: %+v", days)
	}
}

func testTimetableDayRun(startRow, rowStride int, interval time.Duration, minutes ...int) timetableDayRun {
	values := make([]timetableTimeValue, len(minutes))
	for i, minute := range minutes {
		values[i] = timetableTimeValue{
			rowIndex:  startRow + i*rowStride,
			clockTime: time.Date(0, time.January, 1, minute/60, minute%60, 0, 0, time.UTC),
		}
	}
	return timetableDayRun{values: values, interval: interval, rowStride: rowStride}
}
