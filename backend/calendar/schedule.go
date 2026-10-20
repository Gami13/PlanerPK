package calendar

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
	"github.com/woodpkrs/PlanerPK/utils"
)

// shiftClassSchedule handles the bounded RRULE and UTC EXDATE forms emitted by
// snapshot. The original schedule remains unchanged for ghost events.
func shiftClassSchedule(schedule types.ClassSchedule, startHour string) (types.ClassSchedule, error) {
	shifted := schedule
	shifted.Recurrence = slices.Clone(schedule.Recurrence)
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
	for i, rule := range schedule.Recurrence {
		shifted.Recurrence[i], err = shiftRecurrenceRule(rule, newStartTime, schedule.Start.Location())
		if err != nil {
			return types.ClassSchedule{}, err
		}
	}

	return shifted, nil
}

// RRULE describes how a class repeats (for example, every week); its UNTIL
// timestamp is the cutoff for the last meeting. EXDATE identifies one meeting
// to skip by its start timestamp. Both must follow the overridden start time
// so the last meeting stays included and skipped meetings stay excluded.
func shiftRecurrenceRule(rule string, newStartTime time.Time, location *time.Location) (string, error) {
	if timestamp, isExclusion := strings.CutPrefix(rule, types.RecurrenceExDate); isExclusion {
		shifted, err := shiftRecurrenceTimestamp(timestamp, newStartTime, location)
		if err != nil {
			return "", err
		}
		return types.RecurrenceExDate + shifted, nil
	}

	if strings.HasPrefix(rule, types.RecurrenceRulePrefix) {
		recurrence, until, hasUntil := strings.Cut(rule, types.RecurrenceUntil)
		if !hasUntil {
			return "", fmt.Errorf("recurrence rule missing UNTIL: %q", rule)
		}

		shifted, err := shiftRecurrenceTimestamp(until, newStartTime, location)
		if err != nil {
			return "", err
		}
		return recurrence + types.RecurrenceUntil + shifted, nil
	}

	return "", fmt.Errorf("unsupported recurrence rule %q", rule)
}

func shiftRecurrenceTimestamp(timestamp string, newStartTime time.Time, location *time.Location) (string, error) {
	date, err := time.Parse(types.RecurrenceTimeLayout, timestamp)
	if err != nil {
		return "", fmt.Errorf("recurrence timestamp: %w", err)
	}
	localDate := date.In(location)
	shiftedDate := utils.CombineDateAndTime(localDate, newStartTime)

	return shiftedDate.UTC().Format(types.RecurrenceTimeLayout), nil
}
