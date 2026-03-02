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

	legend := extractLegend(file)

	timetable := make(types.Timetable)
	days := getDaysStartEnd(file)

	slog.Debug("Found days in timetable", slog.Int("count", len(days)))

	for _, day := range days {
		classes := traverseDay(file, day, groupColumn, legend)
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

func traverseDay(file *excelize.File, day types.Day, groupColumn int64, legend *types.Legend) []types.Class {
	var classes []types.Class

	currentTime := global.DAY_START_TIME

	for i := day.StartingRow; i <= day.EndingRow; i++ {
		classText := getCell(file, global.SHEET_NAME, i, int(groupColumn))
		if classText != "" {

			// if strings.Contains(classText, "angielski") {
			// 	slog.Debug("Found English class", slog.String("class_text", classText))
			// }
			classEndTime := currentTime.Add(slotDuration)

			class := parseClassText(classText, currentTime, classEndTime, day, legend)

			isNewClass := len(classes) == 0 || class.Name != classes[len(classes)-1].Name || class.Level != classes[len(classes)-1].Level || class.Teacher != classes[len(classes)-1].Teacher || class.Location != classes[len(classes)-1].Location

			if isNewClass {
				// slog.Debug("Adding new class", slog.String("class_name", class.Name))
				classes = append(classes, class)
			} else {
				// slog.Debug("Extending existing class", slog.String("class_name", class.Name))
				classes[len(classes)-1].EndTime = classEndTime
			}
		}
		currentTime = currentTime.Add(slotDuration)

	}

	// Post-process: split any slash-delimited class names into separate classes
	var result []types.Class
	for _, class := range classes {
		result = append(result, splitSlashClasses(class)...)
	}

	// Log the final classes for the day
	for _, class := range result {
		slog.Debug("Final class for day",
			slog.String("name", class.Name),
			slog.String("type", class.Type),
			slog.String("teacher", class.Teacher),
			slog.String("location", class.Location),
			slog.String("additional_info", class.AdditionalInfo),
			slog.String("level", class.Level),
			slog.String("total_hours", class.TotalHours),
			slog.String("start_time", class.StartTime.Format("15:04")),
			slog.String("end_time", class.EndTime.Format("15:04")))
	}
	return result
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
