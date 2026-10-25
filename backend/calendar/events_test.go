package calendar

import (
	"reflect"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
)

func TestBuildClassEvents(t *testing.T) {
	class := types.Class{Name: "Subject", Teacher: "Teacher", Location: "Room", Audience: "GL2"}
	start := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	schedule := types.ClassSchedule{Start: start, End: start.Add(90 * time.Minute),
		Recurrence: &types.WeeklyRecurrence{Interval: 2, Until: start.AddDate(0, 0, 28), Exclusions: []time.Time{start}}}
	original := class.ToEvent(nil, schedule)
	unchanged := class.ToEvent(nil, schedule)
	ghost := class.ToEvent(nil, schedule)
	ghost.Transparency, ghost.ColorId = "transparent", "8"
	ghost.ExtendedProperties.Private["overrideId"] = class.Hash() + "_ghost"

	delta := 150 * time.Minute
	shiftedSchedule := types.ClassSchedule{Start: schedule.Start.Add(delta), End: schedule.End.Add(delta),
		Recurrence: &types.WeeklyRecurrence{Interval: 2, Until: schedule.Recurrence.Until.Add(delta), Exclusions: []time.Time{start.Add(delta)}}}
	shifted := class.ToEvent(nil, shiftedSchedule)
	for _, event := range []*calendar.Event{&original, &ghost, &shifted} {
		event.ExtendedProperties.Private["contentHash"] = eventContentHash(*event)
	}

	for _, tt := range []struct {
		name      string
		override  db.Override
		want      []*calendar.Event
		wantError bool
	}{
		{"normal", db.Override{}, []*calendar.Event{&original}, false},
		{"hidden", db.Override{IsHidden: true}, nil, false},
		{"hidden ghost ignores shift", db.Override{IsHidden: true, IsTransparent: true, StartHour: "10:30"}, []*calendar.Event{&ghost}, false},
		{"shifted", db.Override{StartHour: "10:30"}, []*calendar.Event{&shifted}, false},
		{"shifted with ghost", db.Override{StartHour: "10:30", IsTransparent: true}, []*calendar.Event{&shifted, &ghost}, false},
		{"transparent without shift", db.Override{IsTransparent: true}, []*calendar.Event{&original}, false},
		{"invalid shift", db.Override{StartHour: "bad"}, nil, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildClassEvents(class, &schedule, tt.override, nil)
			if (err != nil) != tt.wantError || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("events = %+v, error = %v; want %+v, wantError = %v", got, err, tt.want, tt.wantError)
			}
			if after := class.ToEvent(nil, schedule); !reflect.DeepEqual(after, unchanged) {
				t.Fatal("original schedule changed")
			}
		})
	}
	if got, err := buildClassEvents(class, nil, db.Override{}, nil); err != nil || len(got) != 0 {
		t.Fatalf("nil schedule: events = %+v, error = %v", got, err)
	}
}
