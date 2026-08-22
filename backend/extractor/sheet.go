package extractor

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var academicYearSheetPattern = regexp.MustCompile(`^(\d{4})_(\d{4})$`)

// ResolveTimetableSheet selects the worksheet for the academic year implied by
// now. Dates before September 15 belong to the previous/current academic year;
// dates on or after September 15 belong to the current/next academic year.
func ResolveTimetableSheet(file *excelize.File, now time.Time) (string, error) {
	target := academicYearSheetName(now)
	for _, sheet := range file.GetSheetList() {
		if sheet == target && isAcademicYearSheetName(sheet) {
			slog.Info("Resolved timetable sheet", slog.String("sheet", sheet))
			return sheet, nil
		}
	}

	sheets := file.GetSheetList()
	available := "<none>"
	if len(sheets) > 0 {
		available = strings.Join(sheets, ", ")
	}
	return "", fmt.Errorf("timetable sheet %q not found; available sheets: %s", target, available)
}

func academicYearSheetName(now time.Time) string {
	year := now.Year()
	if now.Month() < time.September || (now.Month() == time.September && now.Day() < 15) {
		return fmt.Sprintf("%04d_%04d", year-1, year)
	}
	return fmt.Sprintf("%04d_%04d", year, year+1)
}

func isAcademicYearSheetName(name string) bool {
	matches := academicYearSheetPattern.FindStringSubmatch(name)
	if len(matches) != 3 {
		return false
	}

	startYear, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}
	endYear, err := strconv.Atoi(matches[2])
	if err != nil {
		return false
	}
	return endYear == startYear+1
}
