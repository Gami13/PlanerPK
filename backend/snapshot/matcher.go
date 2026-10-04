package snapshot

import (
	"time"

	"github.com/woodpkrs/PlanerPK/types"
)

// Candidates returns all exact slot matches. Multiple results remain ambiguous;
// callers must not choose one arbitrarily for scheduling.
func (index SubjectIndex) Candidates(class types.Class) []Block {
	days := map[time.Weekday]string{
		time.Monday: "MON", time.Tuesday: "TUE", time.Wednesday: "WED",
		time.Thursday: "THU", time.Friday: "FRI",
	}
	day, ok := days[class.Day]
	if !ok {
		return nil
	}
	start := class.StartTime.Hour()*60 + class.StartTime.Minute()
	var matches []Block
	for _, block := range index[class.Name] {
		if block.Activity == class.Type && block.Day != nil && *block.Day == day && block.Start != nil && *block.Start == start {
			matches = append(matches, block)
		}
	}
	return matches
}
