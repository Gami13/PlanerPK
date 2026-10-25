package extractor

import (
	"log/slog"
	"time"

	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

type dayTraverser func(day types.Day, column int64) []types.Class

func newDayTraverser(format cellFormat, file *excelize.File, sheet string, legend *legend.Legend, slotDuration time.Duration) dayTraverser {
	if format == structuredFormat {
		return func(day types.Day, column int64) []types.Class {
			return traverseStructuredDay(file, sheet, day, column, slotDuration)
		}
	}
	return func(day types.Day, column int64) []types.Class {
		return traverseLegacyDay(file, sheet, day, column, legend, slotDuration)
	}
}

func traverseStructuredDay(file *excelize.File, sheet string, day types.Day, column int64, slotDuration time.Duration) []types.Class {
	var classes []types.Class
	previousText := ""
	for row := day.StartingRow; row <= day.EndingRow; row++ {
		text := getCell(file, sheet, row, int(column))
		if text == "" {
			previousText = ""
			continue
		}
		// Excelize repeats a merged cell's anchor value on every covered row.
		if text == previousText {
			continue
		}
		previousText = text

		parsed := parseStructuredClassText(text, day)
		if len(parsed) == 0 {
			slog.Warn("Skipping malformed structured class cell", slog.Int("row", row), slog.Int64("column", column))
			continue
		}
		gridTime := day.StartingTime.Add(time.Duration(row-day.StartingRow) * slotDuration)
		if !containsTime(parsed, gridTime) {
			slog.Warn("Structured class time differs from grid", slog.Int("row", row), slog.String("grid_time", gridTime.Format("15:04")), slog.String("record_time", parsed[0].StartTime.Format("15:04")))
		}
		classes = append(classes, parsed...)
	}
	return classes
}

func traverseLegacyDay(file *excelize.File, sheet string, day types.Day, column int64, legend *legend.Legend, slotDuration time.Duration) []types.Class {
	var classes []types.Class
	for row := day.StartingRow; row <= day.EndingRow; row++ {
		text := getCell(file, sheet, row, int(column))
		if text == "" {
			continue
		}
		start := day.StartingTime.Add(time.Duration(row-day.StartingRow) * slotDuration)
		end := start.Add(slotDuration)
		class := ParseClassText(text, start, end, day, legend)
		if len(classes) == 0 || !sameLegacyClass(classes[len(classes)-1], class) {
			classes = append(classes, class)
		} else {
			classes[len(classes)-1].EndTime = end
		}
	}

	var split []types.Class
	for _, class := range classes {
		split = append(split, SplitSlashClasses(class)...)
	}
	return split
}

func sameLegacyClass(a, b types.Class) bool {
	return a.Name == b.Name && a.Level == b.Level && a.Teacher == b.Teacher && a.Location == b.Location
}

func containsTime(classes []types.Class, target time.Time) bool {
	for _, class := range classes {
		if !target.Before(class.StartTime) && target.Before(class.EndTime) {
			return true
		}
	}
	return false
}
