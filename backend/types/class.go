package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/buildings"
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
	// Audience is the group named by the record, which may differ from the
	// selected timetable column (notably for languages and electives).
	Audience       string
	AdditionalInfo string
	Level          string
	TotalHours     string

	SnapshotBlockID string
}

func (c Class) String() string {
	text := fmt.Sprintf("%s (%s)  %s – %s  %s @ %s",
		c.Name, c.Type,
		c.StartTime.Format("15:04"), c.EndTime.Format("15:04"),
		c.Teacher, c.Location)
	if c.Audience != "" {
		text += fmt.Sprintf("  [%s]", c.Audience)
	}
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
	hash := fmt.Sprintf("%s|%d|%s|%s|%s", c.summary(false), c.Day, c.StartTime.Format("15:04"), c.EndTime.Format("15:04"), c.Location)
	if c.Audience != "" {
		hash += "|" + c.Audience
	}
	return hash
}

func (c Class) Summary() string {
	return c.summary(true)
}

func (c Class) summary(includeAudience bool) string {
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
	if includeAudience && c.Audience != "" {
		summary += fmt.Sprintf(" [%s]", c.Audience)
	}
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

// ClassSchedule holds resolved dates and optional recurrence rules, without
// coupling event construction to the source of the schedule.
type ClassSchedule struct {
	Start      time.Time
	End        time.Time
	Recurrence *WeeklyRecurrence
}

// ToEvent preserves legacy next-weekday scheduling when schedule is nil.
func (c Class) ToEvent(buildings *buildings.Buildings, schedule *ClassSchedule) calendar.Event {
	var start, end string
	var recurrence []string
	var timezone string
	if schedule == nil {
		start, end = c.StartTimeRFC3339(), c.EndTimeRFC3339()
	} else {
		start, end = schedule.Start.Format(time.RFC3339), schedule.End.Format(time.RFC3339)
		if schedule.Recurrence != nil {
			// Google expands recurring meetings in this zone, not a fixed UTC offset.
			timezone = "Europe/Warsaw"
			recurrence = append([]string{schedule.Recurrence.ToRRULE()}, schedule.Recurrence.ToEXDATE()...)
		}
	}
	var description []string
	if c.Audience != "" {
		description = append(description, "Audience: "+c.Audience)
	}
	for _, building := range buildings.GetBuildingsByClassroomKeys(c.ResolvedClassroomKeys) {
		if link := building.MapHTMLAnchor(); link != "" {
			description = append(description, link)
		}
	}
	private := map[string]string{"added_by": ADDED_BY, "overrideId": c.Hash()}
	if c.Audience != "" {
		private["audience"] = c.Audience
	}

	return calendar.Event{
		Summary: c.Summary(),
		Start: &calendar.EventDateTime{
			DateTime: start,
			TimeZone: timezone,
		},
		End: &calendar.EventDateTime{
			DateTime: end,
			TimeZone: timezone,
		},
		Recurrence:  recurrence,
		Location:    c.Location,
		Description: strings.Join(description, "<br>"),
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: private,
		},
	}
}
