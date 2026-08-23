package extractor

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

// TimetableLayout describes the timetable grid discovered in a worksheet.
// Rows and columns use the zero-based indexes used by the extractor.
type TimetableLayout struct {
	HourColumn   int
	SlotDuration time.Duration
	Days         []types.Day
}

// DiscoverTimetableLayout finds the leftmost worksheet column containing one
// or more timetable-like runs with a consistent positive interval. The
// interval is inferred from the values instead of being configured. A
// non-increasing time starts a new day run; the first run is Monday, followed
// by Tuesday and so on.
func DiscoverTimetableLayout(file *excelize.File, sheetName string) (TimetableLayout, error) {
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return TimetableLayout{}, fmt.Errorf("detect timetable layout in sheet %q: read used range: %w", sheetName, err)
	}

	maxColumns := 0
	for _, row := range rows {
		maxColumns = max(maxColumns, len(row))
	}

	for column := range maxColumns {
		candidate, ok := discoverColumnLayout(rows, column)
		if !ok {
			continue
		}
		slog.Info("Detected timetable layout",
			slog.String("sheet", sheetName),
			slog.Int("hour_column", candidate.HourColumn),
			slog.Duration("slot_duration", candidate.SlotDuration),
			slog.Int("day_count", len(candidate.Days)))
		return candidate, nil
	}

	return TimetableLayout{}, fmt.Errorf(
		"could not detect timetable layout in sheet %q: no column contains multiple consecutive time labels with a consistent interval",
		sheetName,
	)
}

const (
	minimumTimetableRunLength = 2
)

type timetableTimeValue struct {
	rowIndex  int
	clockTime time.Time
}

type timetableDayRun struct {
	values    []timetableTimeValue
	interval  time.Duration
	rowStride int
}

// accept adds value when it continues the run's strictly increasing cadence.
// A non-increasing time or cadence mismatch starts a new day candidate.
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

// Worksheet scanning and candidate construction.
func discoverColumnLayout(rows [][]string, column int) (TimetableLayout, bool) {
	days, slotDuration, ok := extractTimetableDays(scanTimetableColumn(rows, column))
	if !ok {
		return TimetableLayout{}, false
	}

	return TimetableLayout{
		HourColumn:   column,
		SlotDuration: slotDuration,
		Days:         days,
	}, true
}

func scanTimetableColumn(rows [][]string, column int) []timetableDayRun {
	var runs []timetableDayRun
	var current timetableDayRun

	for rowIndex, row := range rows {
		if column >= len(row) {
			continue
		}

		parsed, ok := parseTimetableTime(row[column])
		if !ok {
			continue
		}

		observation := timetableTimeValue{
			rowIndex:  rowIndex,
			clockTime: parsed,
		}
		if !current.accept(observation) {
			runs = append(runs, current)
			current = timetableDayRun{values: []timetableTimeValue{observation}}
		}
	}

	if len(current.values) > 0 {
		runs = append(runs, current)
	}
	return runs
}

// Day/cadence conversion. The first complete day run establishes the cadence;
// later day runs with a different cadence are ignored.
func extractTimetableDays(runs []timetableDayRun) (days []types.Day, slotDuration time.Duration, ok bool) {
	var interval time.Duration
	var stride int

	for _, run := range runs {
		if len(run.values) < minimumTimetableRunLength {
			continue
		}
		if interval == 0 {
			interval = run.interval
			stride = run.rowStride
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

// Time parsing.
func parseTimetableTime(value string) (time.Time, bool) {
	s := strings.Replace(strings.TrimSpace(value), ".", ":", 1)
	if len(s) == 4 && s[1] == ':' {
		s = "0" + s
	}
	t, err := time.Parse("15:04", s)
	return t, err == nil
}
