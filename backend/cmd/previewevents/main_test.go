package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPreviewEvents(t *testing.T) {
	input := `{"snapshot":{"blocks":[{"id":"block","day":"WED","start":480,"duration":90,"frequency":"co_tydzien"}],"academicCalendar":{"start":"2026-10-05","end":"2026-10-21","parityAnchor":"2026-09-28"}},"timetable":{"3":[{"Name":"Subject","SnapshotBlockID":"block"}]}}`
	var output bytes.Buffer
	if err := run(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	var report struct {
		ToAdd []struct {
			Summary    string
			Recurrence []string
		}
	}
	if err := json.Unmarshal(output.Bytes(), &report); err != nil || len(report.ToAdd) != 1 || len(report.ToAdd[0].Recurrence) != 1 || !strings.Contains(report.ToAdd[0].Summary, "Subject") {
		t.Fatalf("incomplete report: %s, error = %v", output.String(), err)
	}
	if !strings.Contains(output.String(), "\n  \"ToAdd\":") {
		t.Fatal("report is not indented")
	}
	for _, bad := range []string{`invalid`, `{}`} {
		output.Reset()
		if err := run(strings.NewReader(bad), &output); err == nil || output.Len() != 0 {
			t.Fatal("invalid input must fail without writing a preview")
		}
	}
}
