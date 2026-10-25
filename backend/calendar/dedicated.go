package calendar

import (
	"context"
	"fmt"

	"google.golang.org/api/calendar/v3"
)

// ensureDedicatedCalendar returns an ID for the caller to persist before syncing.
// It is not connected to live sync yet. A saved ID remains authoritative.
func ensureDedicatedCalendar(ctx context.Context, srv *calendar.Service, savedID string) (string, error) {
	if savedID != "" {
		return savedID, nil
	}
	created, err := srv.Calendars.Insert(&calendar.Calendar{Summary: "PlanerPK", TimeZone: "Europe/Warsaw"}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create dedicated calendar: %w", err)
	}
	if created.Id == "" {
		return "", fmt.Errorf("created dedicated calendar has no ID")
	}
	return created.Id, nil
}
