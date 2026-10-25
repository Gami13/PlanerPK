package calendar

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

func TestBuildTimetableDiff(t *testing.T) {
	day, start := "WED", 480
	state := &snapshot.State{Blocks: []snapshot.Block{{ID: "block", Day: &day, Start: &start, Duration: 90, Frequency: "co_tydzien"}},
		AcademicCalendar: snapshot.AcademicCalendar{Start: "2026-10-05", End: "2026-10-21", ParityAnchor: "2026-09-28",
			Closures: []snapshot.DateRange{{From: "2026-10-14", To: "2026-10-14"}}}}
	class := types.Class{Name: "Subject", SnapshotBlockID: "block"}
	timetable := types.GroupTimetable{time.Wednesday: {class}}
	existing, err := buildTimetableEvents(timetable, state, nil, nil)
	if err != nil || len(existing) != 1 {
		t.Fatalf("existing = %+v, error = %v", existing, err)
	}
	removed := &calendar.Event{Id: "obsolete", ExtendedProperties: &calendar.EventExtendedProperties{Private: map[string]string{"overrideId": "obsolete"}}}
	exception := &calendar.Event{Id: "exception", RecurringEventId: "series", Status: "cancelled"}
	current := []*calendar.Event{existing[0], removed, exception}
	overrides := map[string]db.Override{class.Hash(): {StartHour: "10:30", IsTransparent: true}}
	diff, err := BuildTimetableDiff(timetable, state, overrides, nil, current)
	if err != nil || len(diff.Changed) != 1 || len(diff.ToAdd) != 1 || len(diff.Unchanged) != 0 ||
		!reflect.DeepEqual(diff.ToDelete, []*calendar.Event{removed}) || !reflect.DeepEqual(diff.Exceptions, []*calendar.Event{exception}) {
		t.Fatalf("diff = %+v, error = %v", diff, err)
	}
	real, ghost := diff.Changed[0].Proposed, diff.ToAdd[0]
	if diff.Changed[0].Current != existing[0] || real.Start.DateTime == existing[0].Start.DateTime ||
		ghost.Start.DateTime != existing[0].Start.DateTime || ghost.Transparency != "transparent" ||
		!strings.Contains(strings.Join(real.Recurrence, "\n"), "EXDATE:") ||
		real.ExtendedProperties.Private["contentHash"] == ghost.ExtendedProperties.Private["contentHash"] {
		t.Fatalf("incorrect real/ghost proposal: real=%+v ghost=%+v", real, ghost)
	}
	overrides[class.Hash()] = db.Override{IsHidden: true}
	hidden, err := BuildTimetableDiff(timetable, state, overrides, nil, current)
	if err != nil || len(hidden.ToDelete) != 2 || len(hidden.ToAdd) != 0 || len(hidden.Changed) != 0 {
		t.Fatalf("hidden diff = %+v, error = %v", hidden, err)
	}
	failed, err := BuildTimetableDiff(timetable, nil, overrides, nil, current)
	if err == nil || !reflect.DeepEqual(failed, eventDiff{}) {
		t.Fatalf("failed preview = %+v, error = %v", failed, err)
	}
}
