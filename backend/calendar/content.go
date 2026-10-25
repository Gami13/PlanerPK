package calendar

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"time"

	"google.golang.org/api/calendar/v3"
)

// eventContentHash fingerprints importer-owned contents, not Google metadata or
// override identity. Recurrence ordering and equivalent timestamp offsets do not
// change the fingerprint; the recurrence timezone is still compared separately.
func eventContentHash(event calendar.Event) string {
	type eventTime struct{ Date, DateTime, TimeZone string }
	timeFields := func(value *calendar.EventDateTime) eventTime {
		if value == nil {
			return eventTime{}
		}
		dateTime := value.DateTime
		if parsed, err := time.Parse(time.RFC3339Nano, dateTime); err == nil {
			dateTime = parsed.UTC().Format(time.RFC3339Nano)
		}
		zone := value.TimeZone
		if len(event.Recurrence) == 0 {
			zone = "" // A single meeting is identified by its instant, not a recurrence zone.
		}
		return eventTime{value.Date, dateTime, zone}
	}
	rules := append([]string{}, event.Recurrence...)
	slices.Sort(rules)
	transparency := event.Transparency
	if transparency == "" {
		transparency = "opaque"
	}
	content := struct {
		Summary, Description, Location, Transparency, ColorID string
		Start, End                                            eventTime
		Recurrence                                            []string
	}{event.Summary, event.Description, event.Location, transparency, event.ColorId,
		timeFields(event.Start), timeFields(event.End), rules}
	// This value contains only strings and slices of strings; marshaling cannot fail.
	encoded, _ := json.Marshal(content)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
