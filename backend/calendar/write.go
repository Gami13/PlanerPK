package calendar

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/api/calendar/v3"
)

// applyEventChanges writes an already reviewed diff to a dedicated calendar.
// It stops at the first failure, with no rollback; callers must not mark that
// sync successful.
func applyEventChanges(ctx context.Context, srv *calendar.Service, calendarID string, diff eventDiff) (err error) {
	if calendarID == "" || calendarID == "primary" {
		return fmt.Errorf("dedicated calendar ID is required")
	}
	created, updated, deleted := 0, 0, 0
	defer func() {
		slog.Info("Calendar sync write summary", slog.String("calendar_id", calendarID),
			slog.Int("created", created), slog.Int("updated", updated),
			slog.Int("deleted", deleted), slog.Int("skipped", len(diff.Unchanged)),
			slog.Bool("success", err == nil))
	}()
	affected := make(map[string]bool)
	for _, change := range diff.Changed {
		affected[change.Current.Id] = true
	}
	for _, event := range diff.ToDelete {
		affected[event.Id] = true
	}
	for _, exception := range diff.Exceptions {
		if affected[exception.RecurringEventId] {
			slog.Warn("Calendar sync blocked by recurring exceptions", slog.String("calendar_id", calendarID),
				slog.String("series_id", exception.RecurringEventId))
			return fmt.Errorf("series %q has recurring exceptions; manual review required", exception.RecurringEventId)
		}
	}
	for _, event := range diff.ToAdd {
		if _, err := srv.Events.Insert(calendarID, event).Context(ctx).Do(); err != nil {
			return fmt.Errorf("add event %q: %w", event.Summary, err)
		}
		created++
	}
	for _, change := range diff.Changed {
		patch := eventContentsPatch(change.Proposed)
		if _, err := srv.Events.Patch(calendarID, change.Current.Id, patch).Context(ctx).Do(); err != nil {
			return fmt.Errorf("update event %q: %w", change.Current.Id, err)
		}
		updated++
	}
	for _, event := range diff.ToDelete {
		if err := srv.Events.Delete(calendarID, event.Id).Context(ctx).Do(); err != nil {
			return fmt.Errorf("delete event %q: %w", event.Id, err)
		}
		deleted++
	}
	return nil
}

// eventContentsPatch changes only importer-owned fields, leaving user reminders
// and Google metadata untouched. Empty fields explicitly clear previous values.
func eventContentsPatch(event *calendar.Event) *calendar.Event {
	patch := &calendar.Event{
		Summary: event.Summary, Description: event.Description, Location: event.Location,
		Start: event.Start, End: event.End, Recurrence: event.Recurrence,
		Transparency: event.Transparency, ColorId: event.ColorId,
		ExtendedProperties: event.ExtendedProperties,
		ForceSendFields:    []string{"Summary", "Description", "Location", "Transparency"},
	}
	if patch.Transparency == "" {
		patch.Transparency = "opaque"
	}
	if patch.ColorId == "" {
		patch.NullFields = append(patch.NullFields, "ColorId")
	}
	if len(event.Recurrence) == 0 {
		patch.NullFields = append(patch.NullFields, "Recurrence")
	}
	return patch
}
