package extractor

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

var academicYearSheetPattern = regexp.MustCompile(`^(\d{4})_(\d{4})$`)

// ResolveTimetableSheet selects the worksheet whose name matches the
// YYYY_YYYY academic-year pattern with the highest start year.
//
// This avoids hard-coding an expected year from the current date: it
// gracefully picks the latest available sheet (e.g. "2025_2026") even
// when the university hasn't yet published the next year's sheet.
func ResolveTimetableSheet(file *excelize.File) (string, error) {
	bestSheet := ""
	bestStart := -1

	for _, sheet := range file.GetSheetList() {
		startYear, ok := academicYearStart(sheet)
		if !ok {
			continue
		}
		if startYear > bestStart {
			bestStart = startYear
			bestSheet = sheet
		}
	}

	if bestSheet != "" {
		slog.Info("resolved timetable sheet", slog.String("sheet", bestSheet))
		return bestSheet, nil
	}

	sheets := file.GetSheetList()
	available := "<none>"
	if len(sheets) > 0 {
		available = strings.Join(sheets, ", ")
	}
	return "", fmt.Errorf("no academic-year sheet (YYYY_YYYY) found; available sheets: %s", available)
}

// academicYearStart returns the start year if name matches the YYYY_YYYY
// academic-year pattern (e.g. "2025_2026" → 2025, true).
func academicYearStart(name string) (int, bool) {
	matches := academicYearSheetPattern.FindStringSubmatch(name)
	if len(matches) != 3 {
		return 0, false
	}

	startYear, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, false
	}
	endYear, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, false
	}
	if endYear != startYear+1 {
		return 0, false
	}
	return startYear, true
}
