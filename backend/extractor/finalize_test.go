package extractor

import (
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legacy"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
)

func TestResolveClassroomLocationPreservesRawLocation(t *testing.T) {
	buildings := buildings.BuildBuildingsFromLegend(&legend.Legend{Locations: map[string]legend.Location{
		"s. 1/15 Działownia": {
			Shorthand: "s. 1/15 Działownia",
			FullName:  "budynek Działownia",
		},
	}})
	class := types.Class{Location: "s. Działowia 1/15"}
	resolved := finalizeClasses([]types.Class{class}, buildings, nil)
	if len(resolved) != 1 {
		t.Fatalf("resolveLocations returned %d classes, want 1", len(resolved))
	}
	class = resolved[0]

	if class.Location != "s. Działowia 1/15" {
		t.Errorf("Location = %q, want original value", class.Location)
	}
	if len(class.ResolvedClassroomKeys) != 1 || class.ResolvedClassroomKeys[0] != "1 15 działownia" {
		t.Errorf("ResolvedClassroomKeys = %#v, want normalized Działownia key", class.ResolvedClassroomKeys)
	}
}

func TestResolveLocationsAfterLegacySlashSplit(t *testing.T) {
	buildings := buildings.BuildBuildingsFromLegend(&legend.Legend{Locations: map[string]legend.Location{
		"s. SJO": {Shorthand: "s. SJO", FullName: "budynek SJO"},
		"s. N":   {Shorthand: "s. N", FullName: "budynek N"},
	}})
	parsed := types.Class{
		Name:     "J. niemiecki/J.angielski",
		Type:     "ćw.",
		Level:    "C1",
		Teacher:  "mgr D. Guzik/mgr E. Szabat",
		Location: "s. SJO/s. N",
	}

	classes := finalizeClasses(legacy.SplitSlashClasses(parsed), buildings, nil)
	if len(classes) != 2 {
		t.Fatalf("got %d classes, want 2", len(classes))
	}
	wantKeys := []string{"sjo", "n"}
	for i, want := range wantKeys {
		if len(classes[i].ResolvedClassroomKeys) != 1 || classes[i].ResolvedClassroomKeys[0] != want {
			t.Errorf("class %d resolved keys = %#v, want [%s]", i, classes[i].ResolvedClassroomKeys, want)
		}
	}
}

func TestFinalizeClassesSnapshotMatch(t *testing.T) {
	day, start := "MON", 765
	block := snapshot.Block{ID: "b1", Subject: "Subject", Activity: "W", Day: &day, Start: &start}
	class := types.Class{Name: "Subject", Type: "W", Day: time.Monday, StartTime: time.Date(2026, 10, 5, 12, 45, 0, 0, time.UTC)}
	other := block
	other.ID = "b2"
	for _, tt := range []struct {
		name   string
		blocks []snapshot.Block
		wantID string
	}{
		{"unique", []snapshot.Block{block}, "b1"},
		{"unmatched", nil, ""},
		{"ambiguous", []snapshot.Block{block, other}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			classes := finalizeClasses([]types.Class{class}, nil, snapshot.NewSubjectIndex(tt.blocks))
			if len(classes) != 1 || classes[0].SnapshotBlockID != tt.wantID || classes[0].Hash() != class.Hash() {
				t.Fatalf("classes = %+v, want block ID %q and unchanged hash", classes, tt.wantID)
			}
		})
	}
}
