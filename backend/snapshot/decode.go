package snapshot

import (
	"encoding/json"
	"fmt"
)

// Decode decodes the source envelope. Scheduling validation happens
// separately; a missing or null state is never a valid snapshot.
func Decode(data []byte) (Response, error) {
	var envelope struct {
		State          *State          `json:"state"`
		CalendarExport *CalendarExport `json:"calendarExport"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return Response{}, fmt.Errorf("decode snapshot: %w", err)
	}
	if envelope.State == nil {
		return Response{}, fmt.Errorf("decode snapshot: missing state object")
	}
	return Response{State: *envelope.State, CalendarExport: envelope.CalendarExport}, nil
}
