package calendar

import (
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/types"
	"testing"
	"time"
)

func TestUserTimetableIncludesOnlyEligibleReferencedClasses(t *testing.T) {
	base := types.Class{Name: "Base", Day: time.Monday}
	other := types.Class{Name: "Other", Day: time.Tuesday}
	wrongYear := types.Class{Name: "Wrong year", Day: time.Wednesday}
	all := types.Timetable{1: {time.Monday: {base}}, 2: {time.Monday: {base}, time.Tuesday: {other}}, 3: {time.Wednesday: {wrongYear}}}
	groups := []types.ColumnGroup{{Column: 1, Category: "I st. sem. 1"}, {Column: 2, Category: "I st. sem. 1"}, {Column: 3, Category: "I st. sem. 3"}}
	overrides := []db.Override{
		{ClassHash: other.Hash(), SourceColumn: 2}, {ClassHash: other.Hash(), SourceColumn: 2},
		{ClassHash: base.Hash(), SourceColumn: 2}, {ClassHash: wrongYear.Hash(), SourceColumn: 3}, {ClassHash: "missing", SourceColumn: 2},
	}
	merged := UserTimetable(all, 1, overrides, groups)
	if len(merged[time.Monday]) != 1 || len(merged[time.Tuesday]) != 1 || len(merged[time.Wednesday]) != 0 {
		t.Fatalf("unexpected merged timetable: %+v", merged)
	}
	merged[time.Monday][0].Name = "changed"
	if all[1][time.Monday][0].Name != "Base" || len(all[1][time.Tuesday]) != 0 {
		t.Fatal("mutated shared cache")
	}
	removed := UserTimetable(all, 1, nil, groups)
	if len(removed[time.Tuesday]) != 0 {
		t.Fatal("removed addition still present")
	}
	switched := UserTimetable(all, 3, overrides, groups)
	if len(switched[time.Tuesday]) != 0 {
		t.Fatal("addition leaked into another academic year")
	}
	// Imported classes run through the exact same override/event generation path.
	hidden := buildNewEventsMap(UserTimetable(all, 1, overrides, groups), map[string]db.Override{other.Hash(): {IsHidden: true}}, nil)
	if len(hidden) != 1 {
		t.Fatalf("hidden imported class was emitted: %d", len(hidden))
	}
}
