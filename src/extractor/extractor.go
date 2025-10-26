package extractor

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/global"

	"github.com/xuri/excelize/v2"
)

// Get the currently used .xls timetable (under the CURRENT_TIMETABLE_DESTINATION folder)
func GetCurrentTimetablePath() (string, error) {
	entries, err := os.ReadDir(global.TIMETABLE_CURRENT_DESTINATION)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".xls") {
			return global.TIMETABLE_CURRENT_DESTINATION + entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no .xls file found in %s", global.TIMETABLE_CURRENT_DESTINATION)
}

func GetTimetable(groupColumn int) (Timetable, error) {

	timetablePath, err := GetCurrentTimetablePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get current timetable: %v", err)
	}
	timetablePath = timetablePath + "x" //excelize requires .xlsx extension
	file, err := excelize.OpenFile(timetablePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Close the spreadsheet.
		if err := file.Close(); err != nil {
			fmt.Println(err)
			return
		}
	}()

	timetable := make(Timetable)
	days := getDaysStartEnd(file)
	for _, day := range days {
		classes := traverseDay(file, day, groupColumn)
		timetable[day.WeekDay] = classes

	}

	for day, classes := range timetable {
		fmt.Println(day)
		for _, class := range classes {
			fmt.Println("  Class: ", class.String())
		}
	}
	return timetable, nil
}

func getDaysStartEnd(file *excelize.File) []Day {
	// Find starting rows for each day
	var days []Day
	rows, err := file.GetRows(global.SHEET_NAME)
	if err != nil {
		fmt.Println(err)
		return days
	}
	for number, row := range rows {
		if len(row) > 1 {

			if row[global.HOUR_COLUMN] == global.DAY_START {
				day := Day{
					StartingRow: number,
					WeekDay:     time.Weekday((len(days) + 1) % 7)}
				days = append(days, day)
			}
			if row[global.HOUR_COLUMN] == global.DAY_END {
				day := Day{
					EndingRow: number,
				}
				days[len(days)-1].EndingRow = day.EndingRow
			}

		}
	}
	return days
}

func traverseDay(file *excelize.File, day Day, groupColumn int) []Class {
	var classes []Class

	currentTime := time.Date(0, 0, 0, 7, 30, 0, 0, time.UTC)

	for i := day.StartingRow; i <= day.EndingRow; i++ {
		cell := getCell(file, global.SHEET_NAME, i, groupColumn)
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

func parseClassText(text string, classStartTime time.Time, classEndTime time.Time, day Day) Class {
	var class Class
	regex := regexp.MustCompile(`(  |\n)+`)
	parts := regex.Split(text, -1)
	parts = removeEmptyStrings(parts)
	parts = fixClassParts(parts)
	class.Name = strings.TrimSpace(parts[0])

	if slices.Contains(ClassTypes, strings.TrimSpace(parts[1])) {
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
