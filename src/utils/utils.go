package utils

import (
	"fmt"
	"time"
)

type Config struct {
	SheetName  string
	LastRow    int
	HourColumn int
	DayStart   string
	DayEnd     string
}

type Day struct {
	StartingRow int
	EndingRow   int
	WeekDay     time.Weekday
}

type Class struct {
	Name           string
	Type           string
	Day            time.Weekday
	StartTime      time.Time
	EndTime        time.Time
	Location       string
	Teacher        string
	AdditionalInfo string
}

func (c Class) String() string {
	text := fmt.Sprintf("%s (%s) %s-%s | %s | %s", c.Name, c.Type, c.StartTime.Format("15:04"), c.EndTime.Format("15:04"), c.Teacher, c.Location)
	if c.AdditionalInfo != "" {
		text += fmt.Sprintf(" | %s", c.AdditionalInfo)
	}
	return text
}

func NewConfig() *Config {
	return &Config{
		SheetName:  "2025_2026",
		LastRow:    365,
		HourColumn: 1,
		DayStart:   "7.30",
		DayEnd:     "21.00",
	}
}

func GetDateOfNextWeekday(weekday time.Weekday) time.Time {
	today := time.Now()
	daysUntil := (int(weekday) - int(today.Weekday()) + 7) % 7
	next := today.AddDate(0, 0, daysUntil)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
}

//JiPP - Języki i paradygmaty programowania
//WdI - Wstęp do informatyki
//F - Fizyka
//AM - Analiza matematyczna
//J.angielski - Język angielski
//LdI - Logika dla inżynierów
//AzG - Algebra z geometrią
//J. niemiecki - Język niemiecki

//wykład, lab., ćwiczenia

var ClassTypes = []string{
	"wykład",
	"lab.",
	"ćwiczenia",
}
