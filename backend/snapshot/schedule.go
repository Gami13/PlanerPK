package snapshot

import "github.com/woodpkrs/PlanerPK/types"

// ClassSchedule resolves a block for event construction. A nil result means
// there are no classes to emit; no snapshot fetching or matching happens here.
func (block Block) ClassSchedule(calendar AcademicCalendar) (*types.ClassSchedule, error) {
	if block.Frequency == frequencyOneOff || block.Frequency == frequencyFirstOctoberWeek || block.Day == nil || block.Start == nil {
		return nil, nil
	}
	if block.Frequency == frequencyConcreteDate {
		occurrences, err := block.Occurrences(calendar)
		if err != nil || len(occurrences) == 0 {
			return nil, err
		}
		return &types.ClassSchedule{Start: occurrences[0].Start, End: occurrences[0].End}, nil
	}
	return block.WeeklyRecurrence(calendar)
}
