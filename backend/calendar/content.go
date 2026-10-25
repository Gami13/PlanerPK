package calendar

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"

	"google.golang.org/api/calendar/v3"
)

// eventContentHash fingerprints importer-owned contents, not Google metadata or
// override identity. Recurrence ordering does not change the fingerprint.
func eventContentHash(event calendar.Event) string {
	type eventTime struct{ Date, DateTime, TimeZone string }
	timeFields := func(value *calendar.EventDateTime) eventTime {
		if value == nil {
			return eventTime{}
		}
		return eventTime{value.Date, value.DateTime, value.TimeZone}
	}
	rules := append([]string{}, event.Recurrence...)
	slices.Sort(rules)
	content := struct {
		Summary, Description, Location, Transparency, ColorID string
		Start, End                                            eventTime
		Recurrence                                            []string
	}{event.Summary, event.Description, event.Location, event.Transparency, event.ColorId,
		timeFields(event.Start), timeFields(event.End), rules}
	// This value contains only strings and slices of strings; marshaling cannot fail.
	encoded, _ := json.Marshal(content)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
