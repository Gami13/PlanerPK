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
	"ćwiczenia/lab.",
	"ćw./lab.",
	"ćw.",
	"P", // project
}

var AcademicTitles = []string{"dr hab. inż.", "dr hab.", "dr inż.", "dr", "mgr inż.", "mgr", "prof."}

var AdditionalInfoKeywords = []string{"UWAGA", "tydzień", "semestr", "wymiennie", "co 2"}

// ColumnGroup represents an available column/group option parsed from the Excel header rows.
type ColumnGroup struct {
	Column   int    `json:"column"`
	Label    string `json:"label"`
	Category string `json:"category"`
}
