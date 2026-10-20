package snapshot

import (
	"fmt"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
)

// WeeklyRecurrence encodes weekly or A/B schedules, semester halves, and closures.
// Start retains the base series start, even if excluded. A nil result means no classes.
func (block Block) WeeklyRecurrence(calendar AcademicCalendar) (*types.ClassSchedule, error) {
	recurrence := types.WeeklyRecurrence{Interval: 1}
	if block.Frequency == frequencyAlternating {
		recurrence.Interval = weeksPerParityCycle
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
	recurrence.Until = occurrences[len(occurrences)-1].Start
	for _, occurrence := range occurrences {
		if !included[occurrence.Start] {
			recurrence.Exclusions = append(recurrence.Exclusions, occurrence.Start)
		}
	}
	return &types.ClassSchedule{
		Start:      occurrences[0].Start,
		End:        occurrences[0].End,
		Recurrence: append([]string{recurrence.ToRRULE()}, recurrence.ToEXDATE()...),
	}, nil
}
