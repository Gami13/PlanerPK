package types

import (
	"fmt"
	"time"
)

// iCalendar markers shared by schedule generation and override handling.
const (
	RecurrenceRulePrefix = "RRULE:"
	RecurrenceUntil      = ";UNTIL="
	RecurrenceExDate     = "EXDATE:"
	RecurrenceTimeLayout = "20060102T150405Z"
)

// WeeklyRecurrence describes repetition without duplicating a class's start/end.
// Interval is 1 for weekly classes and 2 for alternating weeks. Until and
// Exclusions refer to class start times, not the end of the calendar day.
type WeeklyRecurrence struct {
	Interval   int
	Until      time.Time
	Exclusions []time.Time
}

// ToRRULE encodes the repetition interval and inclusive cutoff in UTC.
func (r WeeklyRecurrence) ToRRULE() string {
	rule := RecurrenceRulePrefix + "FREQ=WEEKLY"
	if r.Interval > 1 {
		rule += fmt.Sprintf(";INTERVAL=%d", r.Interval)
	}
	return rule + RecurrenceUntil + r.Until.UTC().Format(RecurrenceTimeLayout)
}

// ToEXDATE encodes each excluded class start in UTC.
func (r WeeklyRecurrence) ToEXDATE() []string {
	var rules []string
	for _, start := range r.Exclusions {
		rules = append(rules, RecurrenceExDate+start.UTC().Format(RecurrenceTimeLayout))
	}
	return rules
}
