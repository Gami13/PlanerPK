package calendar

import (
	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

// buildResolvedClassEvents builds one class's events without fetching or writing.
// A nil schedule means no classes, never a fallback to the next weekday.
func buildResolvedClassEvents(class types.Class, schedule *types.ClassSchedule, override db.Override, buildings *buildings.Buildings) ([]calendar.Event, error) {
	if schedule == nil || (override.IsHidden && !override.IsTransparent) {
		return nil, nil
	}

	var events []calendar.Event
	if !override.IsHidden {
		shifted, err := shiftClassSchedule(*schedule, override.StartHour)
		if err != nil {
			return nil, err
		}
		events = append(events, class.ToEvent(buildings, &shifted))
	}

	if override.IsTransparent && (override.IsHidden || override.StartHour != "") {
		ghost := class.ToEvent(buildings, schedule)
		ghost.Transparency = "transparent"
		ghost.ColorId = "8"
		ghost.ExtendedProperties.Private["hash"] = class.Hash() + "_ghost"
		events = append(events, ghost)
	}

	return events, nil
}
