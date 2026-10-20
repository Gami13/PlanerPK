package utils

import "time"

// CombineDateAndTime uses the date and location from date and the wall-clock
// time from clock. The clock's date and location are ignored.
func CombineDateAndTime(date, clock time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), clock.Hour(), clock.Minute(), clock.Second(), clock.Nanosecond(), date.Location())
}
