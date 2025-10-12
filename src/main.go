package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type Config struct {
	ExcelPath   string
	SheetName   string
	LastRow     int
	HourColumn  int
	DayStart    string
	DayEnd      string
	GroupColumn int
}

type Day struct {
	StartingRow int
	EndingRow   int
	WeekDay     string
}

type Class struct {
	Name           string
	Type           string
	Day            string
	Time           string
	Location       string
	Teacher        string
	AdditionalInfo string
}

var dayNames = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

func (c Class) String() string {
	return fmt.Sprintf("%s (%s) %s | %s | %s | %s", c.Name, c.Type, c.Time, c.Teacher, c.Location, c.AdditionalInfo)
}

func NewConfig() *Config {
	return &Config{
		ExcelPath:   "C:/Users/gamin/Desktop/Projects/GoogleCalendarTimetable/downloads/10102025_8_INFORMATYKA-zima-rozklad-STACJONARNE-2025_2026.xlsx",
		SheetName:   "2025_2026",
		LastRow:     365,
		HourColumn:  1,
		DayStart:    "7.30",
		DayEnd:      "21.00",
		GroupColumn: 3,
	}
}

func main() {
	config := NewConfig()

	getTimetable(config)
}

func getTimetable(config *Config) map[string][]Class {
	file, err := excelize.OpenFile(config.ExcelPath)
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

	timetable := make(map[string][]Class)
	days := getDaysStartEnd(file, config)
	for i, day := range days {
		classes := traverseDay(file, config, day)
		timetable[indexToDay(i)] = classes

	}

	for _, name := range dayNames {
		classes := timetable[name]
		fmt.Println(name)
		for _, class := range classes {
			fmt.Println("  Class: ", class.String())
		}
	}
	return timetable

}

func getDaysStartEnd(file *excelize.File, config *Config) []Day {
	// Find starting rows for each day
	var days []Day
	rows, err := file.GetRows(config.SheetName)
	if err != nil {
		fmt.Println(err)
		return days
	}
	for number, row := range rows {
		if len(row) > 1 {

			if row[config.HourColumn] == config.DayStart {
				day := Day{
					StartingRow: number,
					WeekDay:     dayNames[len(days)%7],
				}
				days = append(days, day)
			}
			if row[config.HourColumn] == config.DayEnd {
				day := Day{
					EndingRow: number,
				}
				days[len(days)-1].EndingRow = day.EndingRow
			}

		}
	}
	return days
}

func traverseDay(file *excelize.File, config *Config, day Day) []Class {
	var classes []Class
	rows, err := file.GetRows(config.SheetName)
	if err != nil {
		fmt.Println(err)
		return classes
	}
	currentTime := time.Date(0, 0, 0, 7, 30, 0, 0, time.UTC)
	mergedCells, err := file.GetMergeCells(config.SheetName)
	if err != nil {
		fmt.Println(err)
		return classes
	}
	//Here is the issue
	for i := day.StartingRow; i <= day.EndingRow; i++ {
		row := rows[i]

		if len(row) > config.GroupColumn {
			cell := getCell(file, config.SheetName, i, config.GroupColumn)
			classText := cell
			if classText != "" {
				startAxis := fmt.Sprint(toCharStr(config.GroupColumn+1), i+1)
				endAxis := findMergedCell(mergedCells, classText)

				cellStartIndex := axisToRowIndex(startAxis)
				cellEndIndex := axisToRowIndex(endAxis)
				height := cellEndIndex - cellStartIndex + 1

				classEndTime := currentTime.Add(time.Duration(height*15) * time.Minute)

				class := parseClassText(classText, currentTime, classEndTime, day)

				if len(classes) == 0 || class.Name != classes[len(classes)-1].Name {
					classes = append(classes, class)
				}
			}
		}
		currentTime = currentTime.Add(15 * time.Minute)

	}
	return classes
}

func findMergedCell(mergedCells []excelize.MergeCell, text string) string {
	for _, mc := range mergedCells {
		if mc.GetCellValue() == text {

			return mc.GetEndAxis()
		}
	}
	return ""
}

func indexToDay(index int) string {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	if index < 0 || index >= len(days) {
		return ""
	}
	return days[index]
}

func toCharStr(i int) string {
	return string(rune('A' - 1 + i))
}

func axisToRowIndex(axis string) int {
	// axis = B12 or AA32
	//remove letters from string
	var rowIndex int
	for i := len(axis) - 1; i >= 0; i-- {
		if axis[i] < '0' || axis[i] > '9' {
			rowIndex, _ = strconv.Atoi(axis[i+1:])
			break
		}
	}
	return rowIndex
}

// TODO: sometimes the name of the teacher and the type of class are swapped, for example: "JiPP AWyrz lab. s. 143"
func parseClassText(text string, classStartTime time.Time, classEndTime time.Time, day Day) Class {
	var class Class
	regex := regexp.MustCompile(`(  |\n)+`)
	parts := regex.Split(text, -1)
	parts = removeEmptyStrings(parts)
	parts = fixClassParts(parts)
	class.Name = parts[0]
	class.Type = parts[1]
	class.Teacher = parts[2]
	class.Location = parts[3]
	if len(parts) > 4 {
		class.AdditionalInfo = parts[4]
	}

	class.Day = day.WeekDay
	class.Time = fmt.Sprintf("%02d:%02d-%02d:%02d", classStartTime.Hour(), classStartTime.Minute(), classEndTime.Hour(), classEndTime.Minute())

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
