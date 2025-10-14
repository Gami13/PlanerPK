package utils

import (
	"fmt"
	"time"
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
	StartTime      time.Time
	EndTime        time.Time
	Location       string
	Teacher        string
	AdditionalInfo string
}

var DayNames = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

func (c Class) String() string {
	return fmt.Sprintf("%s (%s) %s-%s | %s | %s | %s", c.Name, c.Type, c.StartTime.Format("15:04"), c.EndTime.Format("15:04"), c.Teacher, c.Location, c.AdditionalInfo)
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
