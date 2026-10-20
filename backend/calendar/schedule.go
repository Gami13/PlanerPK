package calendar

import (
	"fmt"
	"slices"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
	"github.com/woodpkrs/PlanerPK/utils"
)

// shiftClassSchedule moves one class's meetings to a new local start time.
// The original schedule remains unchanged for ghost events.
func shiftClassSchedule(schedule types.ClassSchedule, startHour string) (types.ClassSchedule, error) {
	shifted := schedule
	if schedule.Recurrence != nil {
		recurrence := *schedule.Recurrence
		recurrence.Exclusions = slices.Clone(recurrence.Exclusions)
		shifted.Recurrence = &recurrence
	}
	if startHour == "" {
		return shifted, nil
	}

	newStartTime, err := time.Parse("15:04", startHour)
	if err != nil {
		return types.ClassSchedule{}, fmt.Errorf("start hour: %w", err)
	}

	duration := schedule.End.Sub(schedule.Start)
	shifted.Start = utils.CombineDateAndTime(schedule.Start, newStartTime)
	shifted.End = shifted.Start.Add(duration)

	// Move UNTIL so the last class stays included, and EXDATE so cancelled
	// classes still match the series at its new start time.
	if recurrence := shifted.Recurrence; recurrence != nil {
		location := schedule.Start.Location()
		recurrence.Until = utils.CombineDateAndTime(recurrence.Until.In(location), newStartTime)
		for i, date := range recurrence.Exclusions {
			recurrence.Exclusions[i] = utils.CombineDateAndTime(date.In(location), newStartTime)
		}
	}

	return shifted, nil
}
