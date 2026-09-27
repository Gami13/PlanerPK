package extractor

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/legacy"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

// TimetableLayout describes the timetable grid discovered in a worksheet.
// Rows and columns use the zero-based indexes used by the extractor.
type TimetableLayout struct {
	HourColumn    int
	DataEndColumn int
	HeaderRow     int
	SlotDuration  time.Duration
	Days          []types.Day
	mergeRanges   []mergeRange
}

// DiscoverTimetableLayout finds the leftmost worksheet column containing one
// or more timetable-like runs and infers their interval. Explicit weekday
// labels define day boundaries; workbooks without them retain the legacy
// cadence-based detection.
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
		if days := extractLabeledTimetableDays(file, sheetName, rows, candidate.HourColumn); len(days) > 0 {
			candidate.Days = days
		}
		mergeRanges, err := loadMergeRanges(file, sheetName)
		if err != nil {
			slog.Warn("Failed to inspect timetable header merges", slog.Any("error", err))
		}
		candidate.mergeRanges = mergeRanges
		candidate.HeaderRow, candidate.DataEndColumn = discoverHeaderAndDataEnd(rows, candidate)
		slog.Info("Detected timetable layout",
			slog.String("sheet", sheetName),
			slog.Int("hour_column", candidate.HourColumn),
			slog.Int("data_end_column", candidate.DataEndColumn),
			slog.Int("header_row", candidate.HeaderRow),
			slog.Duration("slot_duration", candidate.SlotDuration),
			slog.Int("day_count", len(candidate.Days)))
		return candidate, nil
	}

	return TimetableLayout{}, fmt.Errorf(
		"could not detect timetable layout in sheet %q: no column contains multiple consecutive time labels with a consistent interval",
		sheetName,
	)
}

// discoverHeaderAndDataEnd finds the timetable header row and its right edge.
// The worksheet may contain legend or auxiliary data farther to the right, so
// the worksheet's absolute used range is not a safe timetable boundary.
func discoverHeaderAndDataEnd(rows [][]string, layout TimetableLayout) (int, int) {
	headerRow, endColumn := -1, layout.HourColumn
	if len(layout.Days) == 0 {
		return headerRow, endColumn
	}

	firstDataColumn := layout.HourColumn + 1
	for row := layout.Days[0].StartingRow - 1; row >= 0; row-- {
		if end := lastDataColumn(rows[row], firstDataColumn); end >= 0 {
			headerRow = row
			endColumn = end
			break
		}
	}
	if headerRow == -1 {
		return headerRow, endColumn
	}

	if groupRow := findGLGroupHeaderRow(rows, layout.HourColumn+1, layout.Days[0].StartingRow); groupRow >= 0 {
		headerRow = groupRow
		endColumn = max(endColumn, lastDataColumn(rows[groupRow], layout.HourColumn+1))
	}

	for _, merge := range layout.mergeRanges {
		if merge.StartCol <= endColumn && merge.StartRow <= headerRow && headerRow <= merge.EndRow &&
			merge.EndCol >= firstDataColumn && strings.TrimSpace(merge.Value) != "" {
			endColumn = max(endColumn, merge.EndCol)
		}
	}

	return headerRow, endColumn
}

// findGLGroupHeaderRow scans upward from the timetable for the nearest row
// containing a GL group label. Weekday-name rows can otherwise be mistaken
// for the group header row.
func findGLGroupHeaderRow(rows [][]string, firstDataColumn, dataStartRow int) int {
	for row := min(dataStartRow-1, len(rows)-1); row >= 0; row-- {
		for column := firstDataColumn; column < len(rows[row]); column++ {
			if strings.Contains(strings.ToUpper(rows[row][column]), "GL") {
				return row
			}
		}
	}
	return -1
}

func lastDataColumn(row []string, firstColumn int) int {
	for column := len(row) - 1; column >= firstColumn; column-- {
		if strings.TrimSpace(row[column]) != "" {
			return column
		}
	}
	return -1
}

// Worksheet scanning and candidate construction.
func discoverColumnLayout(rows [][]string, column int) (TimetableLayout, bool) {
	days, slotDuration, ok := legacy.DiscoverTimetableDays(rows, column)
	if !ok {
		return TimetableLayout{}, false
	}

	return TimetableLayout{
		HourColumn:   column,
		SlotDuration: slotDuration,
		Days:         days,
	}, true
}
