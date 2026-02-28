package types

import "time"

type Day struct {
	StartingRow int
	EndingRow   int
	WeekDay     time.Weekday
}
type Timetable = map[time.Weekday][]Class

const ADDED_BY = "GoogleCalendarTimetableImporter by Gami & Frytak"

var ClassTypes = []string{
	"wykład",
	"lab.",
	"ćwiczenia",
	"P", // project
}

// ColumnGroup represents an available column/group option parsed from the Excel header rows.
type ColumnGroup struct {
	Column   int    `json:"column"`
	Label    string `json:"label"`
	Category string `json:"category"`
}
