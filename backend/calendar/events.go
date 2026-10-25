package calendar

import (
	"fmt"
	"time"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

// buildTimetableEvents prepares events offline from one loaded snapshot.
// Any failure returns no events, never a partial preview or guessed dates.
func buildTimetableEvents(timetable types.GroupTimetable, state *snapshot.State, overrides map[string]db.Override, buildings *buildings.Buildings) ([]*calendar.Event, error) {
	if state == nil {
		return nil, fmt.Errorf("snapshot state is missing")
	}
	blocks := make(map[string]snapshot.Block, len(state.Blocks))
	for _, block := range state.Blocks {
		blocks[block.ID] = block
	}
	var events []*calendar.Event
	for day := time.Sunday; day <= time.Saturday; day++ {
		for _, class := range timetable[day] {
			block, ok := blocks[class.SnapshotBlockID]
			if class.SnapshotBlockID == "" || !ok {
				return nil, fmt.Errorf("class %q: snapshot block %q not found", class.Name, class.SnapshotBlockID)
			}
			schedule, err := block.ClassSchedule(state.AcademicCalendar)
			if err != nil {
				return nil, fmt.Errorf("class %q, block %q: %w", class.Name, block.ID, err)
			}
			classEvents, err := buildClassEvents(class, schedule, overrides[class.Hash()], buildings)
			if err != nil {
				return nil, fmt.Errorf("class %q, block %q: %w", class.Name, block.ID, err)
			}
			events = append(events, classEvents...)
		}
	}
	return events, nil
}

// buildClassEvents builds one class's events without fetching or writing.
// A nil schedule means no classes, never a fallback to the next weekday.
func buildClassEvents(class types.Class, schedule *types.ClassSchedule, override db.Override, buildings *buildings.Buildings) ([]*calendar.Event, error) {
	if schedule == nil || (override.IsHidden && !override.IsTransparent) {
		return nil, nil
	}

	var events []*calendar.Event
	if !override.IsHidden {
		shifted, err := shiftClassSchedule(*schedule, override.StartHour)
		if err != nil {
			return nil, err
		}
		event := class.ToEvent(buildings, shifted)
		events = append(events, &event)
	}

	if override.IsTransparent && (override.IsHidden || override.StartHour != "") {
		ghost := class.ToEvent(buildings, *schedule)
		ghost.Transparency = "transparent"
		ghost.ColorId = "8"
		ghost.ExtendedProperties.Private["overrideId"] = class.Hash() + "_ghost"
		events = append(events, &ghost)
	}

	return events, nil
}
