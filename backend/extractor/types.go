package extractor

import (
	"fmt"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/utils"
	"google.golang.org/api/calendar/v3"
)

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

type Timetable = map[time.Weekday][]Class

const ADDED_BY = "GoogleCalendarTimetableImporter by Gami & Frytak"

var ClassTypes = []string{
	"wykład",
	"lab.",
	"ćwiczenia",
}

// ColumnGroup represents an available column/group option parsed from the Excel header rows.
type ColumnGroup struct {
	Column   int    `json:"column"`
	Label    string `json:"label"`
	Category string `json:"category"`
}

func (c Class) String() string {
	text := fmt.Sprintf("%s (%s) %s-%s | %s | %s", c.Name, c.Type, c.StartTime.Format("15:04"), c.EndTime.Format("15:04"), c.Teacher, c.Location)
	if c.AdditionalInfo != "" {
		text += fmt.Sprintf(" | %s", c.AdditionalInfo)
	}
	return text
}

func (c Class) Id() string {
	return fmt.Sprintf("%s_%s", c.Type, c.Name)
}

func (c Class) Summary() string {
	summary := ""
	if len(c.Name) > 0 {
		summary += c.Name
	}
	if len(c.Type) > 0 {
		summary += fmt.Sprintf(" (%s)", c.Type)
	}
	if len(c.Teacher) > 0 {
		summary += fmt.Sprintf(" - %s", c.Teacher)
	}
	if len(c.Location) > 0 {
		summary += fmt.Sprintf(" | %s", c.Location)
	}
	if len(c.AdditionalInfo) > 0 {
		summary += fmt.Sprintf(" | %s", c.AdditionalInfo)
	}
	return summary
}

func (c Class) StartTimeRFC3339() string {
	dayDate := utils.GetDateOfNextWeekday(c.Day)

	startTime := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), c.StartTime.Hour(), c.StartTime.Minute(), 0, 0, time.Now().Local().Location())
	return startTime.Format(time.RFC3339)
}

func (c Class) EndTimeRFC3339() string {
	dayDate := utils.GetDateOfNextWeekday(c.Day)
	endTime := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), c.EndTime.Hour(), c.EndTime.Minute(), 0, 0, time.Now().Local().Location())
	return endTime.Format(time.RFC3339)
}

func (c Class) ToEvent() calendar.Event {
	event := &calendar.Event{
		Summary: c.Summary(),
		Start: &calendar.EventDateTime{
			DateTime: c.StartTimeRFC3339(),
		},
		End: &calendar.EventDateTime{
			DateTime: c.EndTimeRFC3339(),
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"added_by": ADDED_BY,
				"event_id": c.Id(),
				"hash":     c.Hash(),
			},
		},
	}
	return *event
}

func (c Class) Hash() string {
	return fmt.Sprintf("%s|%s|%s|%s", c.Summary(), c.StartTimeRFC3339(), c.EndTimeRFC3339(), c.Location)
}
