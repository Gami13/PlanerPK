package calendar

import (
	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

// BuildTimetableDiff prepares the complete offline event proposal with overrides
// and schedule exclusions, then compares it with the supplied existing events.
// It fetches nothing, writes nothing, and returns no partial diff on failure.
func BuildTimetableDiff(timetable types.GroupTimetable, state *snapshot.State, overrides map[string]db.Override, buildings *buildings.Buildings, current []*calendar.Event) (eventDiff, error) {
	events, err := buildTimetableEvents(timetable, state, overrides, buildings)
	if err != nil {
		return eventDiff{}, err
	}
	return calculateEventDiffs(current, events)
}
