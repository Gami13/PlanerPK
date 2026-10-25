package calendar

import (
	"testing"

	"google.golang.org/api/calendar/v3"
)

func TestEventContentHash(t *testing.T) {
	base := func() calendar.Event {
		return calendar.Event{Summary: "Subject", Start: &calendar.EventDateTime{DateTime: "2026-10-07T08:00:00+02:00", TimeZone: "Europe/Warsaw"},
			End:        &calendar.EventDateTime{DateTime: "2026-10-07T09:30:00+02:00", TimeZone: "Europe/Warsaw"},
			Recurrence: []string{"RRULE:FREQ=WEEKLY", "EXDATE:20261014T060000Z"}}
	}
	want := eventContentHash(base())
	for _, change := range []func(*calendar.Event){
		func(e *calendar.Event) { e.Start.DateTime = "2026-10-08T08:00:00+02:00" },
		func(e *calendar.Event) { e.End.DateTime = "2026-10-07T10:00:00+02:00" },
		func(e *calendar.Event) { e.Start.TimeZone = "UTC" },
		func(e *calendar.Event) { e.Recurrence[0] += ";INTERVAL=2" },
		func(e *calendar.Event) { e.Recurrence[1] = "EXDATE:20261021T060000Z" },
		func(e *calendar.Event) { e.Summary = "Changed" },
		func(e *calendar.Event) { e.Description = "Changed" },
		func(e *calendar.Event) { e.Location = "Changed" },
		func(e *calendar.Event) { e.Transparency = "transparent" },
		func(e *calendar.Event) { e.ColorId = "8" },
	} {
		event := base()
		change(&event)
		if eventContentHash(event) == want {
			t.Errorf("content change not detected: %+v", event)
		}
	}
	event := base()
	event.Id, event.Etag = "google-id", "google-etag"
	event.ExtendedProperties = &calendar.EventExtendedProperties{Private: map[string]string{"overrideId": "identity", "contentHash": "previous"}}
	event.Recurrence[0], event.Recurrence[1] = event.Recurrence[1], event.Recurrence[0]
	if eventContentHash(event) != want {
		t.Fatal("metadata or rule ordering changed the content hash")
	}
}
