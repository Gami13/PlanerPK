package calendar

import (
	"context"
	"fmt"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"google.golang.org/api/calendar/v3"
)

// EnsureUserCalendar prepares and persists the target before any event sync.
// Callers must stop on an error.
func EnsureUserCalendar(ctx context.Context, srv *calendar.Service, queries db.Querier, user db.User) (string, error) {
	if user.CalendarID != "" {
		return user.CalendarID, nil
	}
	id, err := createDedicatedCalendar(ctx, srv)
	if err != nil {
		return "", err
	}
	if err := queries.UpdateUserCalendarID(ctx, db.UpdateUserCalendarIDParams{ID: user.ID, CalendarID: id}); err != nil {
		return "", fmt.Errorf("save dedicated calendar ID for user %d: %w", user.ID, err)
	}
	return id, nil
}

// createDedicatedCalendar returns a new ID for the caller to persist before syncing.
func createDedicatedCalendar(ctx context.Context, srv *calendar.Service) (string, error) {
	created, err := srv.Calendars.Insert(&calendar.Calendar{Summary: "PlanerPK", TimeZone: "Europe/Warsaw"}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create dedicated calendar: %w", err)
	}
	if created.Id == "" {
		return "", fmt.Errorf("created dedicated calendar has no ID")
	}
	return created.Id, nil
}
