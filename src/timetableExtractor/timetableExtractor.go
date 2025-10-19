package timetableExtractor

import (
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/utils"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func GetTimetable(config *utils.Config, timetablePath string, groupColumn int) map[time.Weekday][]utils.Class {
	file, err := excelize.OpenFile(timetablePath)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer func() {
		// Close the spreadsheet.
		if err := file.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	timetable := make(map[time.Weekday][]utils.Class)
	days := getDaysStartEnd(file, config)
	for _, day := range days {
		classes := traverseDay(file, config.SheetName, day, groupColumn)
		timetable[day.WeekDay] = classes

	}

	for day, classes := range timetable {
		fmt.Println(day)
		for _, class := range classes {
			fmt.Println("  Class: ", class.String())
		}
	}
	return timetable

}

func getDaysStartEnd(file *excelize.File, config *utils.Config) []utils.Day {
	// Find starting rows for each day
	var days []utils.Day
	rows, err := file.GetRows(config.SheetName)
	if err != nil {
		fmt.Println(err)
		return days
	}
	for number, row := range rows {
		if len(row) > 1 {

			if row[config.HourColumn] == config.DayStart {
				day := utils.Day{
					StartingRow: number,
					WeekDay:     time.Weekday((len(days) + 1) % 7)}
				days = append(days, day)
			}
			if row[config.HourColumn] == config.DayEnd {
				day := utils.Day{
					EndingRow: number,
				}
				days[len(days)-1].EndingRow = day.EndingRow
			}

		}
	}
	return days
}

func traverseDay(file *excelize.File, sheetName string, day utils.Day, groupColumn int) []utils.Class {
	var classes []utils.Class

	currentTime := time.Date(0, 0, 0, 7, 30, 0, 0, time.UTC)

	for i := day.StartingRow; i <= day.EndingRow; i++ {
		cell := getCell(file, sheetName, i, groupColumn)
		classText := cell
		if classText != "" {

			classEndTime := currentTime.Add(15 * time.Minute)

			class := parseClassText(classText, currentTime, classEndTime, day)
			if len(classes) == 0 || class.Name != classes[len(classes)-1].Name {
				classes = append(classes, class)
			} else if len(classes) != 0 && class.Name == classes[len(classes)-1].Name {
				classes[len(classes)-1].EndTime = classEndTime
			}
		}
		currentTime = currentTime.Add(15 * time.Minute)

	}
	return classes
}

func toCharStr(i int) string {
	return string(rune('A' - 1 + i))
}

func parseClassText(text string, classStartTime time.Time, classEndTime time.Time, day utils.Day) utils.Class {
	var class utils.Class
	regex := regexp.MustCompile(`(  |\n)+`)
	parts := regex.Split(text, -1)
	parts = removeEmptyStrings(parts)
	parts = fixClassParts(parts)
	class.Name = strings.TrimSpace(parts[0])

	if slices.Contains(utils.ClassTypes, strings.TrimSpace(parts[1])) {
		class.Type = strings.TrimSpace(parts[1])
		class.Teacher = strings.TrimSpace(parts[2])
	} else {

		class.Type = strings.TrimSpace(parts[2])
		class.Teacher = strings.TrimSpace(parts[1])
	}
	class.Location = strings.TrimSpace(parts[3])
	if len(parts) > 4 {
		class.AdditionalInfo = strings.TrimSpace(parts[4])
	}

	class.Day = day.WeekDay
	class.StartTime = classStartTime
	class.EndTime = classEndTime

	return class
}

// Removes empty strings from a slice
func removeEmptyStrings(slice []string) []string {
	var result []string
	for _, str := range slice {
		if strings.TrimSpace(str) != "" {
			result = append(result, str)
		}
	}
	return result
}

// Fixes it so that the last part (the one that contains additional info) is merged into one string
func fixClassParts(parts []string) []string {
	if len(parts) > 4 {
		merged := strings.Join(parts[4:], " ")
		return append(parts[:4], merged)
	}
	return parts
}

func getCell(file *excelize.File, sheet string, row int, col int) string {
	cell, err := file.GetCellValue(sheet, fmt.Sprintf("%s%d", toCharStr(col+1), row+1))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return cell
}
