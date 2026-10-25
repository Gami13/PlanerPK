package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

var userSyncLocks sync.Map // user ID -> *sync.Mutex; one application instance.

// ImportTimetable syncs snapshot-resolved events to the user's dedicated calendar.
// A failed read, build, or write never advances the successful-sync timestamp.
func ImportTimetable(all *types.Timetable, token *oauth2.Token, queries db.Querier, state *snapshot.State, buildings *buildings.Buildings) error {
	ctx := context.Background()
	srv, err := createService(ctx, token)
	if err != nil {
		return err
	}
	return syncTimetable(ctx, srv, all, token.AccessToken, queries, state, buildings)
}

// syncTimetable serializes work per user and reloads their settings under the lock.
func syncTimetable(ctx context.Context, srv *calendar.Service, all *types.Timetable, accessToken string, queries db.Querier, state *snapshot.State, buildings *buildings.Buildings) error {
	user, err := queries.GetUserByAccessToken(ctx, accessToken)
	if err != nil {
		return err
	}
	lock, _ := userSyncLocks.LoadOrStore(user.ID, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()
	// Reload after waiting: another sync may have saved the calendar ID, or the
	// user may have changed their group or overrides while this request waited.
	user, err = queries.GetUserByAccessToken(ctx, accessToken)
	if err != nil {
		return err
	}
	if user.IsActive == nil || !*user.IsActive || user.ChosenColumn == nil || *user.ChosenColumn < 0 {
		return nil
	}
	if all == nil {
		return fmt.Errorf("timetable is missing")
	}
	timetable, ok := (*all)[*user.ChosenColumn]
	if !ok {
		return fmt.Errorf("group column %d not found", *user.ChosenColumn)
	}
	overrides, err := queries.GetOverridesForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("read overrides: %w", err)
	}
	byClass := make(map[string]db.Override, len(overrides))
	for _, override := range overrides {
		byClass[override.ClassHash] = override
	}
	events, err := buildTimetableEvents(timetable, state, byClass, buildings)
	if err != nil {
		return err
	}
	if err := syncCalendarEvents(ctx, srv, queries, user, events); err != nil {
		return err
	}
	if err := queries.UpdateUserSyncTime(ctx, accessToken); err != nil {
		return fmt.Errorf("save successful sync time: %w", err)
	}
	slog.Info("Timetable sync completed", slog.Int64("user_id", user.ID))
	return nil
}

// syncCalendarEvents resolves the target, lists current events, and applies the diff.
func syncCalendarEvents(ctx context.Context, srv *calendar.Service, queries db.Querier, user db.User, events []*calendar.Event) error {
	id, err := EnsureUserCalendar(ctx, srv, queries, user)
	if err != nil {
		return err
	}
	current, err := getCurrentEvents(ctx, srv, id)
	if err != nil {
		return err
	}
	diff, err := calculateEventDiffs(current, events)
	if err != nil {
		return err
	}
	return applyEventChanges(ctx, srv, id, diff)
}
