package handler

import (
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
)

func TestHasClass(t *testing.T) {
	class := types.Class{Name: "Subject", Day: time.Tuesday}
	timetable := types.GroupTimetable{time.Monday: {}, time.Tuesday: {{Name: "Other"}, class}}
	if !hasClass(timetable, class.Hash()) {
		t.Fatal("class outside the first day/slot was not found")
	}
	for _, schedule := range []types.GroupTimetable{nil, {}, timetable} {
		if hasClass(schedule, "missing") {
			t.Fatal("found a nonexistent class")
		}
	}
	class.Audience = "GL2"
	if hasClass(timetable, class.Hash()) {
		t.Fatal("matched a different class hash")
	}
}
