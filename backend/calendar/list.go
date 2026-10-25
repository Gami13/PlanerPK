package calendar

import (
	"context"
	"fmt"

	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

// getCurrentEvents reads managed series and exceptions, never expanded instances.
// No rolling cutoff: a series may start before today and still need updating.
func getCurrentEvents(ctx context.Context, srv *calendar.Service, calendarID string) ([]*calendar.Event, error) {
	if calendarID == "" || calendarID == "primary" {
		return nil, fmt.Errorf("dedicated calendar ID is required")
	}
	var all []*calendar.Event
	err := srv.Events.List(calendarID).SingleEvents(false).ShowDeleted(true).Pages(ctx, func(result *calendar.Events) error {
		all = append(all, result.Items...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list managed events: %w", err)
	}
	managed := make(map[string]bool)
	for _, event := range all {
		if event.ExtendedProperties != nil && event.ExtendedProperties.Private["added_by"] == types.ADDED_BY {
			managed[event.Id] = true
		}
	}
	var events []*calendar.Event
	for _, event := range all {
		// Cancelled exceptions can lack private properties; use the parent series.
		if event.Status == "cancelled" && event.RecurringEventId == "" {
			continue // Deleted series/single events are not live sync targets.
		}
		if managed[event.Id] || (event.RecurringEventId != "" && managed[event.RecurringEventId]) {
			events = append(events, event)
		}
	}
	return events, nil
}
