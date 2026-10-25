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

func TestEventContentHashTimestampOffsets(t *testing.T) {
	makeEvent := func(start, end, zone string) calendar.Event {
		return calendar.Event{
			Start:      &calendar.EventDateTime{DateTime: start, TimeZone: zone},
			End:        &calendar.EventDateTime{DateTime: end, TimeZone: zone},
			Recurrence: []string{"RRULE:FREQ=WEEKLY"},
		}
	}
	local := makeEvent("2026-10-21T08:00:00+02:00", "2026-10-21T09:30:00+02:00", "Europe/Warsaw")
	want := eventContentHash(local)
	for _, tt := range []struct {
		name, start, end, zone string
		same                   bool
	}{
		{"UTC representation", "2026-10-21T06:00:00Z", "2026-10-21T07:30:00Z", "Europe/Warsaw", true},
		{"fractional zeros", "2026-10-21T06:00:00.000Z", "2026-10-21T07:30:00.000Z", "Europe/Warsaw", true},
		{"different time", "2026-10-21T07:00:00Z", "2026-10-21T08:30:00Z", "Europe/Warsaw", false},
		{"different recurrence zone", "2026-10-21T06:00:00Z", "2026-10-21T07:30:00Z", "UTC", false},
		{"invalid time", "invalid", "2026-10-21T07:30:00Z", "Europe/Warsaw", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if same := eventContentHash(makeEvent(tt.start, tt.end, tt.zone)) == want; same != tt.same {
				t.Fatalf("same content = %v; want %v", same, tt.same)
			}
		})
	}
}

func TestEventContentHashGoogleDefaults(t *testing.T) {
	proposed := calendar.Event{Start: &calendar.EventDateTime{DateTime: "2026-10-21T06:00:00Z"}}
	current := calendar.Event{Transparency: "opaque", Start: &calendar.EventDateTime{DateTime: "2026-10-21T08:00:00+02:00", TimeZone: "Europe/Warsaw"}}
	if eventContentHash(proposed) != eventContentHash(current) {
		t.Fatal("Google defaults should not trigger another update")
	}
}
