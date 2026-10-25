package calendar

import (
	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

// buildTimetableDiff prepares the complete offline event proposal with overrides
// and schedule exclusions, then compares it with the supplied existing events.
// It fetches nothing, writes nothing, and returns no partial diff on failure.
func buildTimetableDiff(timetable types.GroupTimetable, state *snapshot.State, overrides map[string]db.Override, buildings *buildings.Buildings, current []*calendar.Event) (eventDiff, error) {
	events, err := buildResolvedTimetableEvents(timetable, state, overrides, buildings)
	if err != nil {
		return eventDiff{}, err
	}
	proposed := make([]*calendar.Event, len(events))
	for i := range events {
		proposed[i] = &events[i]
	}
	return compareEventsByOverrideIDAndContent(current, proposed)
}
