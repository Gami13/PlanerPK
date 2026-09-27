package clock

import (
	"strings"
	"time"
)

// Parse accepts the displayed timetable formats H:MM, HH:MM, and HH.MM.
func Parse(value string) (time.Time, bool) {
	s := strings.Replace(strings.TrimSpace(value), ".", ":", 1)
	if len(s) == 4 && s[1] == ':' {
		s = "0" + s
	}
	t, err := time.Parse("15:04", s)
	return t, err == nil
}
