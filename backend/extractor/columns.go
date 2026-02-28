package extractor

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/WoodPKrs/GoogleCalendarTimetable/global"
	"github.com/WoodPKrs/GoogleCalendarTimetable/types"
	"github.com/xuri/excelize/v2"
)

// mergeRange stores parsed information about a single merged cell range.
type mergeRange struct {
	StartRow, StartCol int
	EndRow, EndCol     int
	Value              string
}

var availableGroupsCache = types.NewAvailableGroupsCache()

func markAvailableGroupsStale() {
	availableGroupsCache.MarkStale()
}

// GetAvailableGroups parses the Excel header rows and returns available column groups.
func GetAvailableGroups() ([]types.ColumnGroup, error) {
	if groups, ok := availableGroupsCache.GetIfFresh(); ok {
		slog.Info("Using cached available groups", slog.Int("count", len(groups)))
		return groups, nil
	}

	file, err := openCurrentTimetableFile()
	if err != nil {
		return nil, fmt.Errorf("failed to open timetable file: %w", err)
	}
	defer closeFile(file)
	dataStartRow, err := findDataStartRow(file)
	if err != nil {
		return nil, err
	}

	groupRow, categoryRow := findHeaderRows(file, dataStartRow)
	slog.Debug("Header rows found",
		slog.Int("groupRow", groupRow),
		slog.Int("categoryRow", categoryRow),
		slog.Int("dataStartRow", dataStartRow))

	mergeLookup := buildMergeLookup(file)
	groups := buildColumnGroups(file, groupRow, categoryRow, mergeLookup)

	availableGroupsCache.SetFresh(groups)

	return groups, nil
}

// findDataStartRow finds the first row containing DAY_START in HOUR_COLUMN.
func findDataStartRow(file *excelize.File) (int, error) {
	rows, err := file.GetRows(global.SHEET_NAME)
	if err != nil {
		return 0, fmt.Errorf("failed to get rows: %w", err)
	}

	for i, row := range rows {
		if len(row) > global.HOUR_COLUMN && row[global.HOUR_COLUMN] == global.DAY_START {
			return i, nil
		}
	}

	return 0, fmt.Errorf("could not find row with %q in hour column", global.DAY_START)
}

// findHeaderRows scans upward from dataStartRow to find the group name row
// and the category row directly above it.
func findHeaderRows(file *excelize.File, dataStartRow int) (int, int) {
	for row := dataStartRow - 1; row >= 0; row-- {
		if rowHasDataContent(file, row) {
			return row, max(row-1, 0)
		}
	}
	return max(dataStartRow-1, 0), max(dataStartRow-2, 0)
}

// rowHasDataContent checks if any data column in the row is non-empty.
func rowHasDataContent(file *excelize.File, row int) bool {
	for col := global.HOUR_COLUMN + 1; col <= global.LAST_COLUMN; col++ {
		if strings.TrimSpace(getCell(file, global.SHEET_NAME, row, col)) != "" {
			return true
		}
	}
	return false
}

// buildMergeLookup parses all merge cells and maps each covered (row,col) to its mergeRange.
func buildMergeLookup(file *excelize.File) map[[2]int]*mergeRange {
	lookup := make(map[[2]int]*mergeRange)

	mergeCells, err := file.GetMergeCells(global.SHEET_NAME)
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
func effectiveGroupValue(file *excelize.File, row, col int, lookup map[[2]int]*mergeRange) string {
	if mr, ok := lookup[[2]int{row, col}]; ok {
		if mr.EndCol > mr.StartCol && strings.TrimSpace(mr.Value) != "" {
			return strings.TrimSpace(mr.Value)
		}
	}
	return strings.TrimSpace(getCell(file, global.SHEET_NAME, row, col))
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
func buildColumnGroups(file *excelize.File, groupRow, categoryRow int, lookup map[[2]int]*mergeRange) []types.ColumnGroup {
	var groups []types.ColumnGroup
	gkCounter := 1
	prevYearDigit := byte(0) // first digit of last "grupa XY" — reset GK when it changes

	slog.Debug("Building column groups",
		slog.Int("groupRow", groupRow),
		slog.Int("categoryRow", categoryRow),
		slog.Int("firstCol", global.HOUR_COLUMN+1),
		slog.Int("lastCol", global.LAST_COLUMN))

	for col := global.HOUR_COLUMN + 1; col <= global.LAST_COLUMN; col++ {
		value := effectiveGroupValue(file, groupRow, col, lookup)
		category := effectiveCategory(file, categoryRow, col, lookup)

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
func effectiveCategory(file *excelize.File, categoryRow int, col int, lookup map[[2]int]*mergeRange) string {
	for c := col; c >= 0; c-- {
		var v string
		if mr, ok := lookup[[2]int{categoryRow, c}]; ok && strings.TrimSpace(mr.Value) != "" {
			v = strings.TrimSpace(mr.Value)
		} else {
			v = strings.TrimSpace(getCell(file, global.SHEET_NAME, categoryRow, c))
		}
		if v != "" && !isHourLabel(v) {
			return v
		}
	}
	return ""
}
