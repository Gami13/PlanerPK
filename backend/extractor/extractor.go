package extractor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/global"
	"github.com/WoodPKrs/GoogleCalendarTimetable/types"
	"github.com/xuri/excelize/v2"
)

var timetableCache = types.NewTimetableCache()

func markTimetableCacheStale() {
	timetableCache.MarkStale()
	markLegendCacheStale()
}

func GetTimetable(groupColumn int64) (types.Timetable, error) {
	if timetable, ok := timetableCache.GetIfFresh(groupColumn); ok {
		slog.Info("Using cached timetable", slog.Int64("group_column", groupColumn))
		return timetable, nil
	}

	file, err := openCurrentTimetableFile()
	if err != nil {
		return nil, fmt.Errorf("failed to open timetable file: %w", err)
	}
	defer closeFile(file)

	timetable := make(types.Timetable)
	days := getDaysStartEnd(file)

	slog.Debug("Found days in timetable", slog.Int("count", len(days)))

	for _, day := range days {
		classes := traverseDay(file, day, groupColumn)
		timetable[day.WeekDay] = classes
		slog.Debug("Processed day",
			slog.String("day", day.WeekDay.String()),
			slog.Int("classes", len(classes)))
	}

	timetableCache.SetFresh(groupColumn, timetable)

	return timetable, nil
}

func getDaysStartEnd(file *excelize.File) []types.Day {
	// Find starting rows for each day
	var days []types.Day
	rows, err := file.GetRows(global.SHEET_NAME)
	if err != nil {
		slog.Error("Failed to get rows from timetable", slog.Any("error", err))
		return days
	}
	for number, row := range rows {
		if len(row) > 1 {

			if row[global.HOUR_COLUMN] == global.DAY_START {
				day := types.Day{
					StartingRow: number,
					WeekDay:     time.Weekday((len(days) + 1) % 7),
				}
				days = append(days, day)
			}
			if row[global.HOUR_COLUMN] == global.DAY_END {
				days[len(days)-1].EndingRow = number
			}

		}
	}
	return days
}

const slotDuration = 15 * time.Minute

func traverseDay(file *excelize.File, day types.Day, groupColumn int64) []types.Class {
	var classes []types.Class

	currentTime := global.DAY_START_TIME

	for i := day.StartingRow; i <= day.EndingRow; i++ {
		classText := getCell(file, global.SHEET_NAME, i, int(groupColumn))
		if classText != "" {

			classEndTime := currentTime.Add(slotDuration)

			class := parseClassText(classText, currentTime, classEndTime, day)

			if len(classes) == 0 || class.Name != classes[len(classes)-1].Name {
				classes = append(classes, class)
			} else {
				classes[len(classes)-1].EndTime = classEndTime
			}
		}
		currentTime = currentTime.Add(slotDuration)

	}
	return classes
}

func getCell(file *excelize.File, sheet string, row int, col int) string {
	column, err := excelize.ColumnNumberToName(col + 1)
	if err != nil {
		slog.Error("Failed to convert column number to name", slog.Int("col", col), slog.Any("error", err))
		return ""
	}

	cell, err := file.GetCellValue(sheet, fmt.Sprintf("%s%d", column, row+1))
	if err != nil {
		slog.Error("Failed to get cell value",
			slog.String("sheet", sheet),
			slog.Int("row", row),
			slog.Int("col", col),
			slog.Any("error", err))
		return ""
	}
	return cell
}
