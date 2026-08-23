package extractor

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

// mergeRange stores parsed information about a single merged cell range.
type mergeRange struct {
	StartRow, StartCol int
	EndRow, EndCol     int
	Value              string
}

// BuildAvailableGroups parses the Excel header rows and returns available column groups.
func BuildAvailableGroups(file *excelize.File, sheetName string, layout TimetableLayout) ([]types.ColumnGroup, error) {
	if len(layout.Days) == 0 {
		return nil, fmt.Errorf("cannot build groups for sheet %q: timetable layout contains no day runs", sheetName)
	}
	dataStartRow := layout.Days[0].StartingRow

	groupRow, categoryRow := findHeaderRows(file, sheetName, dataStartRow, layout)
	slog.Debug("Header rows found",
		slog.Int("groupRow", groupRow),
		slog.Int("categoryRow", categoryRow),
		slog.Int("dataStartRow", dataStartRow))

	mergeLookup := buildMergeLookup(file, sheetName)
	groups := buildColumnGroups(file, sheetName, groupRow, categoryRow, mergeLookup, layout)

	return groups, nil
}

// findHeaderRows scans upward from dataStartRow to find the group name row
// and the category row directly above it.
func findHeaderRows(file *excelize.File, sheetName string, dataStartRow int, layout TimetableLayout) (int, int) {
	for row := dataStartRow - 1; row >= 0; row-- {
		if rowHasDataContent(file, sheetName, row, layout) {
			return row, max(row-1, 0)
		}
	}
	return max(dataStartRow-1, 0), max(dataStartRow-2, 0)
}

// rowHasDataContent checks if any data column in the row is non-empty.
func rowHasDataContent(file *excelize.File, sheetName string, row int, layout TimetableLayout) bool {
	for col := layout.HourColumn + 1; col <= layout.DataEndColumn; col++ {
		if strings.TrimSpace(getCell(file, sheetName, row, col)) != "" {
			return true
		}
	}
	return false
}

// buildMergeLookup parses all merge cells and maps each covered (row,col) to its mergeRange.
func buildMergeLookup(file *excelize.File, sheetName string) map[[2]int]*mergeRange {
	lookup := make(map[[2]int]*mergeRange)

	mergeCells, err := file.GetMergeCells(sheetName)
	if err != nil {
		slog.Error("Failed to get merge cells", slog.Any("error", err))
		return lookup
	}

	for _, mc := range mergeCells {
		startCol, startRow, err := excelize.CellNameToCoordinates(mc.GetStartAxis())
		if err != nil {
			slog.Error("Failed to parse merge start", slog.Any("error", err))
			continue
		}
		endCol, endRow, err := excelize.CellNameToCoordinates(mc.GetEndAxis())
		if err != nil {
			slog.Error("Failed to parse merge end", slog.Any("error", err))
			continue
		}

		// Convert 1-based excelize coords to 0-based
		startCol--
		startRow--
		endCol--
		endRow--

		mr := &mergeRange{
			StartRow: startRow, StartCol: startCol,
			EndRow: endRow, EndCol: endCol,
			Value: mc.GetCellValue(),
		}

		for r := startRow; r <= endRow; r++ {
			for c := startCol; c <= endCol; c++ {
				lookup[[2]int{r, c}] = mr
			}
		}
	}

	return lookup
}

// effectiveGroupValue returns the group name for a cell, preferring the merge
// value when the cell is inside a multi-column merge.
func effectiveGroupValue(file *excelize.File, sheetName string, row, col int, lookup map[[2]int]*mergeRange) string {
	if mr, ok := lookup[[2]int{row, col}]; ok {
		if mr.EndCol > mr.StartCol && strings.TrimSpace(mr.Value) != "" {
			return strings.TrimSpace(mr.Value)
		}
	}
	return strings.TrimSpace(getCell(file, sheetName, row, col))
}

// buildColumnGroups iterates data columns and builds ColumnGroup entries.
//
// Logic:
//  1. Skip empty cells and "Godziny" (hour-reference) columns.
//  2. Determine each group's column span from Excel merges.
//  3. Assign sequential GK labels to every multi-column span AND every
//     standalone cell whose value starts with "grupa".
//  4. Reset the GK counter when the year digit changes (e.g. grupa 1x → 2x).
//  5. Resolve categories from the row above, inheriting leftward when the
//     direct value is empty or "Godziny".
func buildColumnGroups(file *excelize.File, sheetName string, groupRow, categoryRow int, lookup map[[2]int]*mergeRange, layout TimetableLayout) []types.ColumnGroup {
	var groups []types.ColumnGroup
	gkCounter := 1
	prevYearDigit := byte(0) // first digit of last "grupa XY" — reset GK when it changes

	slog.Debug("Building column groups",
		slog.Int("groupRow", groupRow),
		slog.Int("categoryRow", categoryRow),
		slog.Int("firstCol", layout.HourColumn+1),
		slog.Int("lastCol", layout.DataEndColumn))

	for col := layout.HourColumn + 1; col <= layout.DataEndColumn; col++ {
		value := effectiveGroupValue(file, sheetName, groupRow, col, lookup)
		category := effectiveCategory(file, sheetName, categoryRow, col, lookup)

		slog.Debug("Scanning column",
			slog.Int("col", col),
			slog.String("effective", value),
			slog.String("category", category))

		// Skip empty and hour-reference columns.
		if value == "" || isHourLabel(value) {
			if isHourLabel(value) {
				slog.Debug("Skipping hour-reference column", slog.Int("col", col))
			}
			continue
		}

		mr, inMerge := lookup[[2]int{groupRow, col}]
		isMultiColMerge := inMerge && mr.EndCol > mr.StartCol

		// Reset GK counter when the "year digit" changes (e.g. grupa 1x → grupa 2x).
		if yd := groupYearDigit(value); yd != 0 && yd != prevYearDigit {
			if prevYearDigit != 0 {
				slog.Debug("Year digit changed, resetting GK counter",
					slog.String("from", string(prevYearDigit)),
					slog.String("to", string(yd)))
			}
			gkCounter = 1
			prevYearDigit = yd
		}

		// Determine column range: full merge span, or just the current column.
		startCol, endCol := col, col
		if isMultiColMerge {
			startCol, endCol = mr.StartCol, mr.EndCol
		}

		needsGK := isMultiColMerge || strings.HasPrefix(strings.ToLower(value), "grupa")

		for c := startCol; c <= endCol; c++ {
			label := value
			if needsGK {
				label = fmt.Sprintf("%s (GK%d)", value, gkCounter)
				gkCounter++
			}

			slog.Debug("Adding column group",
				slog.Int("col", c),
				slog.String("label", label),
				slog.String("category", category))

			groups = append(groups, types.ColumnGroup{
				Column:   c,
				Label:    label,
				Category: category,
			})
		}

		col = endCol // outer loop's col++ advances past the range
	}

	slog.Info("Built column groups", slog.Int("count", len(groups)))
	return groups
}

// effectiveCategory returns the category for a column. When the direct value
// is empty or an hour label ("Godziny"), it inherits from the nearest
// meaningful category to the left.
func effectiveCategory(file *excelize.File, sheetName string, categoryRow int, col int, lookup map[[2]int]*mergeRange) string {
	for c := col; c >= 0; c-- {
		var v string
		if mr, ok := lookup[[2]int{categoryRow, c}]; ok && strings.TrimSpace(mr.Value) != "" {
			v = strings.TrimSpace(mr.Value)
		} else {
			v = strings.TrimSpace(getCell(file, sheetName, categoryRow, c))
		}
		if v != "" && !isHourLabel(v) {
			return v
		}
	}
	return ""
}
