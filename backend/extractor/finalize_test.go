package extractor

import (
	"testing"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legacy"
	"github.com/woodpkrs/PlanerPK/legend"
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
	resolved := resolveLocations([]types.Class{class}, buildings)
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

	classes := resolveLocations(legacy.SplitSlashClasses(parsed), buildings)
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
