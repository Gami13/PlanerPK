package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/utils"
	"google.golang.org/api/calendar/v3"
)

type Class struct {
	Name      string
	Type      string
	Day       time.Weekday
	StartTime time.Time
	EndTime   time.Time
	Location  string
	// ResolvedClassroomKeys are precomputed while the timetable is imported.
	// Location remains the original text shown in the calendar.
	ResolvedClassroomKeys []string
	Teacher               string
	AdditionalInfo        string
	Level                 string
	TotalHours            string
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

func (c Class) Hash() string {
	return fmt.Sprintf("%s|%d|%s|%s|%s", c.Summary(), c.Day, c.StartTime.Format("15:04"), c.EndTime.Format("15:04"), c.Location)
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

func (c Class) ToEvent(buildings *Buildings) calendar.Event {
	links := make([]string, 0)

	if buildings != nil {
		for _, building := range buildings.GetBuildingsByClassroomKeys(c.ResolvedClassroomKeys) {
			if link := building.MapHTMLAnchor(); link != "" {
				links = append(links, link)
			}
		}
	}

	event := &calendar.Event{
		Summary: c.Summary(),
		Start: &calendar.EventDateTime{
			DateTime: c.StartTimeRFC3339(),
		},
		End: &calendar.EventDateTime{
			DateTime: c.EndTimeRFC3339(),
		},
		Location:    c.Location,
		Description: strings.Join(links, "<br>"),
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"added_by": ADDED_BY,
				"hash":     c.Hash(),
			},
		},
	}
	return *event
}
