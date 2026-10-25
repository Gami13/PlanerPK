package extractor

import (
	"time"

	"github.com/woodpkrs/PlanerPK/internal/clock"
	"github.com/woodpkrs/PlanerPK/types"
)

const minimumTimetableRunLength = 2

type timetableTimeValue struct {
	rowIndex  int
	clockTime time.Time
}

type timetableDayRun struct {
	values    []timetableTimeValue
	interval  time.Duration
	rowStride int
}

func (run *timetableDayRun) accept(value timetableTimeValue) bool {
	if len(run.values) > 0 {
		previous := run.values[len(run.values)-1]
		diff := value.clockTime.Sub(previous.clockTime)
		rowDiff := value.rowIndex - previous.rowIndex
		if diff <= 0 {
			return false
		}
		if run.interval == 0 {
			run.interval = diff
			run.rowStride = rowDiff
		}
		if diff != run.interval || rowDiff != run.rowStride {
			return false
		}
	}
	run.values = append(run.values, value)
	return true
}

// DiscoverTimetableDays preserves the cadence-based day detection used by
// workbooks without explicit weekday separators.
func DiscoverTimetableDays(rows [][]string, column int) ([]types.Day, time.Duration, bool) {
	return extractTimetableDays(scanTimetableColumn(rows, column))
}

func scanTimetableColumn(rows [][]string, column int) []timetableDayRun {
	var runs []timetableDayRun
	var current timetableDayRun
	for rowIndex, row := range rows {
		if column >= len(row) {
			continue
		}
		clockTime, ok := clock.Parse(row[column])
		if !ok {
			continue
		}
		value := timetableTimeValue{rowIndex: rowIndex, clockTime: clockTime}
		if !current.accept(value) {
			runs = append(runs, current)
			current = timetableDayRun{values: []timetableTimeValue{value}}
		}
	}
	if len(current.values) > 0 {
		runs = append(runs, current)
	}
	return runs
}

func extractTimetableDays(runs []timetableDayRun) (days []types.Day, slotDuration time.Duration, ok bool) {
	var interval time.Duration
	var stride int
	for _, run := range runs {
		if len(run.values) < minimumTimetableRunLength {
			continue
		}
		if interval == 0 {
			interval, stride = run.interval, run.rowStride
		}
		if run.interval != interval || run.rowStride != stride {
			continue
		}
		first, last := run.values[0], run.values[len(run.values)-1]
		days = append(days, types.Day{
			StartingRow:  first.rowIndex,
			EndingRow:    last.rowIndex,
			WeekDay:      time.Monday + time.Weekday(len(days)),
			StartingTime: first.clockTime,
		})
	}
	if interval <= 0 || stride <= 0 {
		return nil, 0, false
	}
	return days, interval / time.Duration(stride), true
}
