package calendar

import (
	"context"
	"fmt"

	"google.golang.org/api/calendar/v3"
)

// applyEventDiff writes an already reviewed diff to a dedicated calendar.
// It stops at the first failure, with no rollback; callers must not mark that
// sync successful. This helper is not connected to the live importer yet.
func applyEventDiff(ctx context.Context, srv *calendar.Service, calendarID string, diff eventDiff) error {
	if calendarID == "" || calendarID == "primary" {
		return fmt.Errorf("dedicated calendar ID is required")
	}
	if len(diff.Exceptions) != 0 && (len(diff.Changed) != 0 || len(diff.ToDelete) != 0) {
		return fmt.Errorf("refusing updates or deletions while recurring exceptions need review")
	}
	for _, event := range diff.ToAdd {
		if _, err := srv.Events.Insert(calendarID, event).Context(ctx).Do(); err != nil {
			return fmt.Errorf("add event %q: %w", event.Summary, err)
		}
	}
	for _, change := range diff.Changed {
		patch := eventContentsPatch(change.Proposed)
		if _, err := srv.Events.Patch(calendarID, change.Current.Id, patch).Context(ctx).Do(); err != nil {
			return fmt.Errorf("update event %q: %w", change.Current.Id, err)
		}
	}
	for _, event := range diff.ToDelete {
		if err := srv.Events.Delete(calendarID, event.Id).Context(ctx).Do(); err != nil {
			return fmt.Errorf("delete event %q: %w", event.Id, err)
		}
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
		ForceSendFields:    []string{"Summary", "Description", "Location", "Transparency", "ColorId"},
	}
	if len(event.Recurrence) == 0 {
		patch.NullFields = []string{"Recurrence"}
	}
	return patch
}
