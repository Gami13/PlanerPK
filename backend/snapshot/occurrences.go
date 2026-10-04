package snapshot

import (
	"fmt"
	"math"
	"time"
)

type Occurrence struct {
	Start time.Time
	End   time.Time
}

const (
	minutesPerHour      = 60
	hoursPerDay         = 24
	minutesPerDay       = hoursPerDay * minutesPerHour
	daysPerWeek         = 7
	hoursPerWeek        = daysPerWeek * hoursPerDay
	weeksPerParityCycle = 2
	teachingWeekA       = 0
	teachingWeekB       = 1
)

const (
	frequencyWeekly           = "co_tydzien"
	frequencyAlternating      = "co_2_tygodnie"
	frequencyFirstHalf        = "co_tydzien_pierwsza_polowa"
	frequencySecondHalf       = "co_tydzien_druga_polowa"
	frequencyConcreteDate     = "konkretna_data"
	frequencyOneOff           = "jednorazowo"
	frequencyFirstOctoberWeek = "pierwszy_tydzien_pazdziernika"
)

var weekdays = map[string]time.Weekday{
	"MON": time.Monday,
	"TUE": time.Tuesday,
	"WED": time.Wednesday,
	"THU": time.Thursday,
	"FRI": time.Friday,
}

// Occurrences generates weekly or explicitly dated slots in the process
// timezone. Calendar exceptions are not applied yet.
func (block Block) Occurrences(calendar AcademicCalendar) ([]Occurrence, error) {
	if block.Frequency == frequencyOneOff || block.Frequency == frequencyFirstOctoberWeek {
		return nil, nil
	}
	if block.Day == nil || block.Start == nil {
		return nil, nil
	}
	if *block.Start < 0 || *block.Start >= minutesPerDay || block.Duration <= 0 {
		return nil, fmt.Errorf("block %s: invalid start or duration", block.ID)
	}
	weekday, ok := weekdays[*block.Day]
	if !ok {
		return nil, fmt.Errorf("block %s: invalid weekday %q", block.ID, *block.Day)
	}
	start, end, err := block.occurrenceBounds(calendar)
	if err != nil {
		return nil, err
	}
	switch block.Frequency {
	case frequencyConcreteDate:
		return block.datedOccurrence(start, end)
	case frequencyWeekly, frequencyFirstHalf, frequencySecondHalf:
		return block.weeklyOccurrences(start, end, weekday, time.Time{}), nil
	case frequencyAlternating:
		return block.alternatingOccurrences(calendar, start, end, weekday)
	default:
		return nil, fmt.Errorf("block %s: unsupported frequency %q", block.ID, block.Frequency)
	}
}

func (block Block) weeklyOccurrences(start, end time.Time, weekday time.Weekday, anchor time.Time) []Occurrence {
	first := start.AddDate(0, 0, (int(weekday)-int(start.Weekday())+daysPerWeek)%daysPerWeek)
	var occurrences []Occurrence
	for date := first; !date.After(end); date = date.AddDate(0, 0, daysPerWeek) {
		if !anchor.IsZero() && !matchesWeekParity(date, anchor, *block.TeachingWeekParity) {
			continue
		}
		occurrences = append(occurrences, block.occurrenceOn(date))
	}
	return occurrences
}

func (block Block) datedOccurrence(start, end time.Time) ([]Occurrence, error) {
	if block.Date == nil {
		return nil, fmt.Errorf("block %s: missing concrete date", block.ID)
	}
	date, err := time.ParseInLocation(time.DateOnly, *block.Date, time.Local)
	if err != nil {
		return nil, fmt.Errorf("block %s date: %w", block.ID, err)
	}
	if date.Before(start) || date.After(end) {
		return nil, nil
	}
	return []Occurrence{block.occurrenceOn(date)}, nil
}

func (block Block) alternatingOccurrences(calendar AcademicCalendar, start, end time.Time, weekday time.Weekday) ([]Occurrence, error) {
	if block.TeachingWeekParity == nil || (*block.TeachingWeekParity != teachingWeekA && *block.TeachingWeekParity != teachingWeekB) {
		return nil, fmt.Errorf("block %s: missing or invalid teaching week parity", block.ID)
	}
	anchor, err := time.Parse(time.DateOnly, calendar.ParityAnchor)
	if err != nil {
		return nil, fmt.Errorf("parity anchor: %w", err)
	}
	return block.weeklyOccurrences(start, end, weekday, anchor), nil
}

func (block Block) occurrenceOn(date time.Time) Occurrence {
	start := time.Date(date.Year(), date.Month(), date.Day(), *block.Start/minutesPerHour, *block.Start%minutesPerHour, 0, 0, time.Local)
	return Occurrence{Start: start, End: start.Add(time.Duration(block.Duration) * time.Minute)}
}

func (block Block) occurrenceBounds(calendar AcademicCalendar) (time.Time, time.Time, error) {
	start, err := time.ParseInLocation(time.DateOnly, calendar.Start, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("term start: %w", err)
	}
	end, err := time.ParseInLocation(time.DateOnly, calendar.End, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("term end: %w", err)
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("term end precedes start")
	}
	if block.Frequency == frequencyFirstHalf || block.Frequency == frequencySecondHalf {
		boundary, err := time.ParseInLocation(time.DateOnly, calendar.HalfBoundary, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("half boundary: %w", err)
		}
		if block.Frequency == frequencyFirstHalf && boundary.AddDate(0, 0, -1).Before(end) {
			end = boundary.AddDate(0, 0, -1)
		}
		if block.Frequency == frequencySecondHalf && boundary.After(start) {
			start = boundary
		}
	}
	return start, end, nil
}

func matchesWeekParity(date, anchor time.Time, parity int) bool {
	dateUTC := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	week := int(math.Floor(dateUTC.Sub(anchor).Hours() / (hoursPerWeek)))
	return (week%weeksPerParityCycle+weeksPerParityCycle)%weeksPerParityCycle == parity
}
