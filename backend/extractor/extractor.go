package extractor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

func BuildTimetable(groups []types.ColumnGroup, file *excelize.File, sheetName string, legend *legend.Legend, buildings *buildings.Buildings, layout TimetableLayout) types.Timetable {
	timetable := make(types.Timetable)

	for _, group := range groups {
		groupTimetable := buildTimetableForGroup(int64(group.Column), file, sheetName, legend, buildings, layout.Days, layout.SlotDuration)
		timetable[int64(group.Column)] = groupTimetable

	}
	return timetable
}

func buildTimetableForGroup(groupColumn int64, file *excelize.File, sheetName string, legend *legend.Legend, buildings *buildings.Buildings, days []types.Day, slotDuration time.Duration) types.GroupTimetable {
	timetable := make(types.GroupTimetable)

	slog.Debug("Found days in timetable", slog.Int("count", len(days)))

	for _, day := range days {
		classes := traverseDay(file, sheetName, day, groupColumn, legend, buildings, slotDuration)
		timetable[day.WeekDay] = classes
		slog.Debug("Processed day",
			slog.String("day", day.WeekDay.String()),
			slog.Int("classes", len(classes)))
	}

	return timetable
}

func traverseDay(file *excelize.File, sheetName string, day types.Day, groupColumn int64, legend *legend.Legend, buildings *buildings.Buildings, slotDuration time.Duration) []types.Class {
	var classes []types.Class

	currentTime := day.StartingTime

	for i := day.StartingRow; i <= day.EndingRow; i++ {
		classText := getCell(file, sheetName, i, int(groupColumn))
		if classText != "" {

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
	result := finalizeClasses(classes, buildings)

	// Log the final classes for the day
	for _, class := range result {
		var buildingNames []string
		for _, building := range buildings.GetBuildingsByClassroomKeys(class.ResolvedClassroomKeys) {
			buildingNames = append(buildingNames, building.CanonicalName)
		}
		slog.Debug("Final class for day",
			slog.String("name", class.Name),
			slog.String("type", class.Type),
			slog.String("teacher", class.Teacher),
			slog.String("location", class.Location),
			slog.String("additional_info", class.AdditionalInfo),
			slog.String("level", class.Level),
			slog.String("total_hours", class.TotalHours),
			slog.Any("classroom_keys", class.ResolvedClassroomKeys),
			slog.Any("buildings", buildingNames),
			slog.String("start_time", class.StartTime.Format("15:04")),
			slog.String("end_time", class.EndTime.Format("15:04")))
	}
	return result
}

func finalizeClasses(classes []types.Class, buildings *buildings.Buildings) []types.Class {
	var result []types.Class
	for _, class := range classes {
		for _, finalClass := range splitSlashClasses(class) {
			if buildings != nil && finalClass.Location != "" {
				finalClass.ResolvedClassroomKeys = buildings.ResolveClassroomKeys(finalClass.Location)
			}
			result = append(result, finalClass)
		}
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
