package extractor

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

func BuildTimetable(groups []types.ColumnGroup, file *excelize.File, sheetName string, legend *legend.Legend, buildings *buildings.Buildings, layout TimetableLayout, index snapshot.SubjectIndex) types.Timetable {
	traverse := newDayTraverser(detectCellFormat(file, sheetName, groups, layout.Days), file, sheetName, legend, layout.SlotDuration)
	timetable := make(types.Timetable)
	for _, group := range groups {
		column := int64(group.Column)
		timetable[column] = buildTimetableForGroup(group, layout.Days, buildings, traverse, index)
	}
	mergeSupplementalTimetables(timetable, groups)
	return timetable
}

// SelectableGroups excludes language/elective source columns from the group
// picker; those classes are included in the matching regular group's timetable.
func SelectableGroups(groups []types.ColumnGroup) []types.ColumnGroup {
	selectable := make([]types.ColumnGroup, 0, len(groups))
	for _, group := range groups {
		if !isSupplementalGroup(group.Label) {
			selectable = append(selectable, group)
		}
	}
	return selectable
}

func isSupplementalGroup(label string) bool {
	label = strings.ToLower(label)
	return strings.Contains(label, "językowe") || strings.Contains(label, "obieralne")
}

func mergeSupplementalTimetables(timetable types.Timetable, groups []types.ColumnGroup) {
	regularByCategory := make(map[string][]int64)
	for _, group := range groups {
		if group.Category != "" && !isSupplementalGroup(group.Label) {
			regularByCategory[group.Category] = append(regularByCategory[group.Category], int64(group.Column))
		}
	}
	for _, group := range groups {
		if !isSupplementalGroup(group.Label) {
			continue
		}
		for _, destination := range regularByCategory[group.Category] {
			mergeGroupTimetable(timetable[destination], timetable[int64(group.Column)])
		}
	}
}

func mergeGroupTimetable(destination, source types.GroupTimetable) {
	for day, classes := range source {
		merged := destination[day]
		seen := make(map[string]struct{}, len(merged)+len(classes))
		for _, class := range merged {
			seen[class.Hash()] = struct{}{}
		}
		for _, class := range classes {
			if _, exists := seen[class.Hash()]; exists {
				continue
			}
			seen[class.Hash()] = struct{}{}
			merged = append(merged, class)
		}
		sort.SliceStable(merged, func(i, j int) bool {
			return merged[i].StartTime.Before(merged[j].StartTime)
		})
		destination[day] = merged
	}
}

func buildTimetableForGroup(group types.ColumnGroup, days []types.Day, buildings *buildings.Buildings, traverse dayTraverser, index snapshot.SubjectIndex) types.GroupTimetable {
	timetable := make(types.GroupTimetable)

	slog.Debug("Found days in timetable", slog.Int("count", len(days)))

	for _, day := range days {
		classes := classesForGroup(traverse(day, int64(group.Column)), group.Label)
		classes = finalizeClasses(classes, buildings, index)
		timetable[day.WeekDay] = classes
		logClasses(classes, buildings)
		slog.Debug("Processed day",
			slog.String("day", day.WeekDay.String()),
			slog.Int("classes", len(classes)))
	}

	return timetable
}

func logClasses(classes []types.Class, buildings *buildings.Buildings) {
	for _, class := range classes {
		var buildingNames []string
		for _, building := range buildings.GetBuildingsByClassroomKeys(class.ResolvedClassroomKeys) {
			buildingNames = append(buildingNames, building.CanonicalName)
		}
		slog.Debug("Final class for day",
			slog.String("name", class.Name),
			slog.String("type", class.Type),
			slog.String("teacher", class.Teacher),
			slog.String("audience", class.Audience),
			slog.String("location", class.Location),
			slog.String("additional_info", class.AdditionalInfo),
			slog.String("level", class.Level),
			slog.String("total_hours", class.TotalHours),
			slog.Any("classroom_keys", class.ResolvedClassroomKeys),
			slog.Any("buildings", buildingNames),
			slog.String("start_time", class.StartTime.Format("15:04")),
			slog.String("end_time", class.EndTime.Format("15:04")))
	}
}

func finalizeClasses(classes []types.Class, buildings *buildings.Buildings, index snapshot.SubjectIndex) []types.Class {
	for i := range classes {
		if buildings != nil && classes[i].Location != "" {
			classes[i].ResolvedClassroomKeys = buildings.ResolveClassroomKeys(classes[i].Location)
		}
		if index != nil {
			candidates := index.Candidates(classes[i])
			if len(candidates) == 1 {
				classes[i].SnapshotBlockID = candidates[0].ID
			} else {
				slog.Warn("Class has no unique snapshot match", "class", classes[i].String(), "day", classes[i].Day, "candidates", len(candidates))
			}
		}
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
