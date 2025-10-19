package types

import (
	"fmt"
	"time"
)

type Timetable = map[time.Weekday][]Class
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

var ClassTypes = []string{
	"wykład",
	"lab.",
	"ćwiczenia",
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
func (c Class) String() string {
	text := fmt.Sprintf("%s (%s) %s-%s | %s | %s", c.Name, c.Type, c.StartTime.Format("15:04"), c.EndTime.Format("15:04"), c.Teacher, c.Location)
	if c.AdditionalInfo != "" {
		text += fmt.Sprintf(" | %s", c.AdditionalInfo)
	}
	return text
}

type UserInfo struct {
	Sub       string `json:"sub"`
	Name      string `json:"name"`
	GivenName string `json:"given_name"`
	Picture   string `json:"picture"`
	Email     string `json:"email"`
}

func (u UserInfo) String() string {
	// Name: John Doe, Email: john.doe@example.com, GivenName: John, FamilyName: Doe, Profile: https://profiles.google.com/johndoe, Picture: https://lh3.googleusercontent.com/a-/AOh14Gg, EmailVerified: true, Gender: male
	return fmt.Sprintf("Name: %s, Email: %s, GivenName: %s, Picture: %s",
		u.Name, u.Email, u.GivenName, u.Picture)
}
