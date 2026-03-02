package types

import (
	"fmt"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/utils"
	"google.golang.org/api/calendar/v3"
)

type Class struct {
	Name           string
	Type           string
	Day            time.Weekday
	StartTime      time.Time
	EndTime        time.Time
	Location       string
	Teacher        string
	AdditionalInfo string
	Level          string
	TotalHours     string
}

func (c Class) String() string {
	text := fmt.Sprintf("%s (%s)  %s – %s  %s @ %s",
		c.Name, c.Type,
		c.StartTime.Format("15:04"), c.EndTime.Format("15:04"),
		c.Teacher, c.Location)
	if c.Level != "" {
		text += fmt.Sprintf("  [%s]", c.Level)
	}
	if c.TotalHours != "" {
		text += fmt.Sprintf("  (%s)", c.TotalHours)
	}
	if c.AdditionalInfo != "" {
		text += fmt.Sprintf(" — %s", c.AdditionalInfo)
	}
	return text
}

// Collisions between same-name classes on the same day are intentional — they allow the calendar syncer to detect time changes
func (c Class) Id() string {
	return fmt.Sprintf("%s_%s_%s_%s", c.Type, c.Name, c.Day, c.Level)
}

func (c Class) Summary() string {
	summary := ""
	if len(c.Name) > 0 {
		summary += c.Name
	}
	if len(c.Level) > 0 {
		summary += fmt.Sprintf(" %s", c.Level)
	}
	if len(c.Type) > 0 {
		summary += fmt.Sprintf(" (%s)", c.Type)
	}
	if len(c.Teacher) > 0 {
		summary += fmt.Sprintf(" - %s", c.Teacher)
	}
	// if len(c.Location) > 0 {
	// 	summary += fmt.Sprintf(" | %s", c.Location)
	// }
	if len(c.TotalHours) > 0 {
		summary += fmt.Sprintf(" | %s", c.TotalHours)
	}
	if len(c.AdditionalInfo) > 0 {
		summary += fmt.Sprintf(" | %s", c.AdditionalInfo)
	}
	return summary
}

func (c Class) timeOnDay(t time.Time) time.Time {
	dayDate := utils.GetDateOfNextWeekday(c.Day)
	return time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), t.Hour(), t.Minute(), 0, 0, time.Now().Local().Location())
}

func (c Class) StartTimeRFC3339() string {
	return c.timeOnDay(c.StartTime).Format(time.RFC3339)
}

func (c Class) EndTimeRFC3339() string {
	return c.timeOnDay(c.EndTime).Format(time.RFC3339)
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
		Location: c.Location,
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
