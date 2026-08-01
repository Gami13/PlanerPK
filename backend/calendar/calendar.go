package calendar

import (
	"context"
	"log/slog"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/types"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

func ImportTimetable(allTimetables *types.Timetable, groupColumn int64, token *oauth2.Token, queries db.Querier, overrides []db.Override, buildings *types.Buildings) {
	ctx := context.Background()

	slog.Info("Starting timetable import", slog.Int64("group_column", groupColumn))

	timetable, ok := (*allTimetables)[groupColumn]
	if !ok {
		slog.Error("Group column not found in timetable", slog.Int64("group_column", groupColumn))
		return
	}
	srv, err := createService(ctx, token)
	if err != nil {
		slog.Error("Failed to create calendar service", slog.Any("error", err))
		return
	}
	currentEvents, err := getCurrentEvents(srv)
	if err != nil {
		slog.Error("Failed to get current events", slog.Any("error", err))
		return
	}

	// Convert overrides slice to a map keyed by class_hash for O(1) lookup.
	overridesMap := make(map[string]db.Override, len(overrides))
	for _, o := range overrides {
		overridesMap[o.ClassHash] = o
	}

	newEvents := buildNewEventsMap(timetable, overridesMap, buildings)
	eventsToDelete, eventsToAdd, unchangedEvents := calculateEventDiffs(currentEvents, newEvents)

	createdCount, deletedCount := applyEventChanges(srv, eventsToDelete, eventsToAdd)
	skippedCount := len(unchangedEvents)

	err = queries.UpdateUserSyncTime(ctx, token.AccessToken)
	if err != nil {
		slog.Error("Failed to update user sync time", slog.Any("error", err))
	}

	slog.Info("Timetable import completed",
		slog.Int("created", createdCount),
		slog.Int("deleted", deletedCount),
		slog.Int("skipped", skippedCount))
}

func applyEventChanges(srv *calendar.Service, eventsToDelete []*calendar.Event, eventsToAdd []*calendar.Event) (int, int) {
	var deletedCount, createdCount int

	for _, event := range eventsToAdd {
		slog.Info("Creating event", slog.String("summary", event.Summary))
		_, err := srv.Events.Insert("primary", event).Do()
		if err != nil {
			slog.Error("Failed to create event",
				slog.String("summary", event.Summary),
				slog.Any("error", err))
			continue
		}
		createdCount++
	}

	for _, event := range eventsToDelete {
		slog.Info("Deleting event", slog.String("summary", event.Summary))
		err := srv.Events.Delete("primary", event.Id).Do()
		if err != nil {
			slog.Error("Failed to delete event",
				slog.String("summary", event.Summary),
				slog.Any("error", err))
			continue
		}
		deletedCount++
	}

	return createdCount, deletedCount
}

func getCurrentEvents(srv *calendar.Service) ([]*calendar.Event, error) {
	// extendedProperties has a private field "added_by" with value "GoogleCalendarTimetableImporter by Gami & Frytak"
	// set to start of day
	t := time.Now().Truncate(24 * time.Hour).Format(time.RFC3339)
	events, err := srv.Events.List("primary").TimeMin(t).ShowDeleted(false).OrderBy("startTime").SingleEvents(true).Do()
	if err != nil {
		slog.Error("Failed to retrieve calendar events", slog.Any("error", err))
		return nil, err
	}

	// Filter events by extendedProperties
	filteredEvents := make([]*calendar.Event, 0)
	for _, event := range events.Items {
		if event.ExtendedProperties != nil && event.ExtendedProperties.Private != nil {
			if event.ExtendedProperties.Private["added_by"] == types.ADDED_BY {
				filteredEvents = append(filteredEvents, event)
			}
		}
	}
	slog.Debug("Retrieved current events", slog.Int("count", len(filteredEvents)))
	return filteredEvents, nil
}
