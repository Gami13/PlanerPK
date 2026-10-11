package snapshot

import "fmt"

type Recurrence struct {
	First Occurrence
	Rules []string
}

// WeeklyRecurrence encodes the base weekly or A/B schedule, including semester halves.
// Closure exclusions are not encoded yet. A nil result means no classes.
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
	if len(occurrences) == 0 {
		return nil, nil
	}
	last := occurrences[len(occurrences)-1].Start
	return &Recurrence{
		First: occurrences[0],
		Rules: []string{rule + ";UNTIL=" + last.UTC().Format("20060102T150405Z")},
	}, nil
}
