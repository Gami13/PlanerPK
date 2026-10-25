package calendar

import (
	"reflect"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

func TestBuildResolvedTimetableEvents(t *testing.T) {
	day, start, date := "WED", 480, "2026-10-14"
	block := snapshot.Block{ID: "weekly", Day: &day, Start: &start, Duration: 90, Frequency: "co_tydzien"}
	state := snapshot.State{Blocks: []snapshot.Block{block}, AcademicCalendar: snapshot.AcademicCalendar{
		Start: "2026-10-05", End: "2026-10-21", ParityAnchor: "2026-09-28",
		Closures: []snapshot.DateRange{{From: "2026-10-07", To: "2026-10-07"}},
	}}
	for _, frequency := range []string{"konkretna_data", "jednorazowo"} {
		other := block
		other.ID, other.Frequency, other.Date = frequency, frequency, &date
		state.Blocks = append(state.Blocks, other)
	}
	class := types.Class{Name: "Excel details", SnapshotBlockID: block.ID}
	one := types.Class{Name: "Concrete", SnapshotBlockID: "konkretna_data"}
	omitted := types.Class{Name: "Omitted", SnapshotBlockID: "jednorazowo"}
	timetable := types.GroupTimetable{time.Wednesday: {class, one, omitted}}
	override := db.Override{StartHour: "10:30", IsTransparent: true}
	schedule, err := block.ClassSchedule(state.AcademicCalendar)
	if err != nil {
		t.Fatal(err)
	}
	want, err := buildResolvedClassEvents(class, schedule, override, nil)
	if err != nil {
		t.Fatal(err)
	}
	oneSchedule, err := state.Blocks[1].ClassSchedule(state.AcademicCalendar)
	if err != nil {
		t.Fatal(err)
	}
	want = append(want, one.ToEvent(nil, *oneSchedule))
	want[len(want)-1].ExtendedProperties.Private["contentHash"] = eventContentHash(want[len(want)-1])
	got, err := buildResolvedTimetableEvents(timetable, &state, map[string]db.Override{class.Hash(): override}, nil)
	if err != nil || len(got) != 3 || !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %+v, error = %v; want %+v", got, err, want)
	}
	assertError := func(events []calendar.Event, err error) {
		t.Helper()
		if err == nil || events != nil {
			t.Fatalf("events = %+v, error = %v; want no partial events and an error", events, err)
		}
	}
	assertError(buildResolvedTimetableEvents(timetable, nil, nil, nil))
	for _, id := range []string{"", "missing"} {
		bad := class
		bad.SnapshotBlockID = id
		assertError(buildResolvedTimetableEvents(types.GroupTimetable{time.Wednesday: {one, bad}}, &state, nil, nil))
	}
	assertError(buildResolvedTimetableEvents(timetable, &state, map[string]db.Override{class.Hash(): {StartHour: "bad"}}, nil))
	state.Blocks[1].Date = new(string)
	assertError(buildResolvedTimetableEvents(timetable, &state, nil, nil))
}
