package snapshot

import (
	"os"
	"strings"
	"testing"
)

func TestDecodeSnapshotFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/winter-2026-27.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.State.SchemaVersion != 1 || got.State.PublicationMode != "ST" || len(got.State.Blocks) != 369 || len(got.State.Rooms) != 25 {
		t.Fatalf("unexpected fixture state: schema=%d mode=%q blocks=%d rooms=%d", got.State.SchemaVersion, got.State.PublicationMode, len(got.State.Blocks), len(got.State.Rooms))
	}
	if got.State.AcademicCalendar.Start == "" || got.CalendarExport == nil || len(got.CalendarExport.Events) != 4057 {
		t.Fatalf("fixture missing calendar or oracle events")
	}
	for _, block := range got.State.Blocks {
		if block.Teacher != "" || block.TeacherDisplay != "" {
			t.Fatalf("fixture contains unsanitized teacher on block %q", block.ID)
		}
	}
}

func TestDecodeSnapshotAcceptsOptionalAndAdditiveFields(t *testing.T) {
	got, err := decodeSnapshot([]byte(`{"state":{"schemaVersion":1,"newField":true,"blocks":[{"id":"b1","day":null,"start":null,"additionalCohorts":null,"newBlockField":42}]},"newEnvelopeField":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.CalendarExport != nil || len(got.State.Blocks) != 1 || got.State.Blocks[0].Day != nil || got.State.Blocks[0].Start != nil {
		t.Fatalf("unexpected optional-field decoding: %+v", got)
	}
}

func TestDecodeSnapshotRejectsMissingOrInvalidState(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{"empty input", ``, "decode snapshot"},
		{"missing state", `{"calendarExport":{"events":[]}}`, "missing state object"},
		{"null state", `{"state":null}`, "missing state object"},
		{"null envelope", `null`, "missing state object"},
		{"wrong state type", `{"state":[]}`, "decode snapshot"},
		{"trailing data", `{"state":{}} {}`, "decode snapshot"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeSnapshot([]byte(tt.input))
			if err == nil || !strings.Contains(err.Error(), tt.want) || got.State.SchemaVersion != 0 || got.CalendarExport != nil {
				t.Fatalf("decodeSnapshot(%q) = %+v, %v; want error containing %q and empty response", tt.input, got, err, tt.want)
			}
		})
	}
}
