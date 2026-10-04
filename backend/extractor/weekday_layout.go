package extractor

import (
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/internal/clock"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

var weekdayNames = map[string]time.Weekday{
	"PONIEDZIAŁEK": time.Monday, "MONDAY": time.Monday,
	"WTOREK": time.Tuesday, "TUESDAY": time.Tuesday,
	"ŚRODA": time.Wednesday, "SRODA": time.Wednesday, "WEDNESDAY": time.Wednesday,
	"CZWARTEK": time.Thursday, "THURSDAY": time.Thursday,
	"PIĄTEK": time.Friday, "PIATEK": time.Friday, "FRIDAY": time.Friday,
	"SOBOTA": time.Saturday, "SATURDAY": time.Saturday,
	"NIEDZIELA": time.Sunday, "SUNDAY": time.Sunday,
}

// extractLabeledTimetableDays scans the first class column for day separators
// and records the first and last time row below each one.
func extractLabeledTimetableDays(file *excelize.File, sheet string, rows [][]string, hourColumn int) []types.Day {
	classColumn := hourColumn + 1
	current := types.Day{StartingRow: -1}
	hasSeparator := false
	var days []types.Day

	for row, values := range rows {
		label := strings.ToUpper(strings.TrimSpace(getCell(file, sheet, row, classColumn)))
		if weekday, ok := weekdayNames[label]; ok {
			if current.StartingRow >= 0 {
				days = append(days, current)
			}
			current = types.Day{StartingRow: -1, WeekDay: weekday}
			hasSeparator = true
			continue
		}

		if !hasSeparator || hourColumn >= len(values) {
			continue
		}
		clockTime, ok := clock.Parse(values[hourColumn])
		if !ok {
			continue
		}
		if current.StartingRow < 0 {
			current.StartingRow = row
			current.StartingTime = clockTime
		}
		current.EndingRow = row
	}

	if current.StartingRow >= 0 {
		days = append(days, current)
	}
	return days
}
