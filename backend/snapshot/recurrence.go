package snapshot

import (
	"fmt"
	"time"
)

const recurrenceTimeLayout = "20060102T150405Z"

type Recurrence struct {
	First Occurrence
	Rules []string
}

// WeeklyRecurrence encodes weekly or A/B schedules, semester halves, and closures.
// First retains the base series start, even if excluded. A nil result means no classes.
func (block Block) WeeklyRecurrence(calendar AcademicCalendar) (*Recurrence, error) {
	rule := "RRULE:FREQ=WEEKLY"
	if block.Frequency == frequencyAlternating {
		rule += fmt.Sprintf(";INTERVAL=%d", weeksPerParityCycle)
	} else if block.Frequency != frequencyWeekly && block.Frequency != frequencyFirstHalf && block.Frequency != frequencySecondHalf {
		return nil, fmt.Errorf("block %s: unsupported weekly recurrence frequency %q", block.ID, block.Frequency)
	}
	occurrences, err := block.baseOccurrences(calendar)
	if err != nil {
		return nil, err
	}
	remaining, err := calendar.excludeClosures(occurrences)
	if err != nil {
		return nil, err
	}
	if len(remaining) == 0 {
		return nil, nil
	}
	included := make(map[time.Time]bool, len(remaining))
	for _, occurrence := range remaining {
		included[occurrence.Start] = true
	}
	last := occurrences[len(occurrences)-1].Start
	rules := []string{rule + ";UNTIL=" + last.UTC().Format(recurrenceTimeLayout)}
	for _, occurrence := range occurrences {
		if !included[occurrence.Start] {
			rules = append(rules, "EXDATE:"+occurrence.Start.UTC().Format(recurrenceTimeLayout))
		}
	}
	return &Recurrence{
		First: occurrences[0],
		Rules: rules,
	}, nil
}
