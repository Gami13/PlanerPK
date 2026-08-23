package extractor

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

// BuildAvailableGroups parses the Excel header rows and returns available column groups.
func BuildAvailableGroups(file *excelize.File, sheetName string, layout TimetableLayout) ([]types.ColumnGroup, error) {
	if len(layout.Days) == 0 {
		return nil, fmt.Errorf("cannot build groups for sheet %q: timetable layout contains no day runs", sheetName)
	}
	if layout.HeaderRow < 0 {
		return nil, fmt.Errorf("cannot build groups for sheet %q: timetable header row was not detected", sheetName)
	}
	dataStartRow := layout.Days[0].StartingRow

	groupRow := layout.HeaderRow
	categoryRow := max(groupRow-1, 0)
	slog.Debug("Header rows found",
		slog.Int("groupRow", groupRow),
		slog.Int("categoryRow", categoryRow),
		slog.Int("dataStartRow", dataStartRow))

	mergeRanges := layout.mergeRanges
	if mergeRanges == nil {
		var err error
		mergeRanges, err = loadMergeRanges(file, sheetName)
		if err != nil {
			slog.Error("Failed to get merge cells", slog.Any("error", err))
		}
	}
	mergeLookup := buildMergeLookup(mergeRanges)
	groups := buildColumnGroups(file, sheetName, groupRow, categoryRow, mergeLookup, layout)

	return groups, nil
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
	firstDataColumn := layout.HourColumn + 1

	slog.Debug("Building column groups",
		slog.Int("groupRow", groupRow),
		slog.Int("categoryRow", categoryRow),
		slog.Int("firstCol", firstDataColumn),
		slog.Int("lastCol", layout.DataEndColumn))

	for col := firstDataColumn; col <= layout.DataEndColumn; col++ {
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
