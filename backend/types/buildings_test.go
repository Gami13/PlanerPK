package types

import "testing"

func TestBuildingsResolveAndLookupUsesLegendOnly(t *testing.T) {
	buildings := BuildBuildingsFromLegend(&Legend{Locations: map[string]Location{
		"s. 114 GIL":         {Shorthand: "s. 114 GIL", FullName: "budynek GIL"},
		"s. 135(L2)":         {Shorthand: "s. 135(L2)", FullName: "budynek Chemii"},
		"s. 136(L1,)":        {Shorthand: "s. 136(L1,)", FullName: "budynek Chemii"},
		"s. 1/15 Działownia": {Shorthand: "s. 1/15 Działownia", FullName: "budynek Działownia"},
		"s. 138":             {Shorthand: "s. 138", FullName: "budynek SJO ul. Warszawska"},
	}})

	if buildingKey := buildings.Classrooms["135"]; buildingKey != "budynek Chemii" {
		t.Errorf("135 building key = %q, want budynek Chemii", buildingKey)
	}
	if buildingKey := buildings.Classrooms["sjo"]; buildingKey != "budynek SJO ul. Warszawska" {
		t.Errorf("SJO building key = %q, want SJO building", buildingKey)
	}
	if keys := buildings.ResolveClassroomKeys("s. SJO"); len(keys) != 1 || keys[0] != "sjo" {
		t.Errorf("SJO resolved keys = %#v, want [sjo]", keys)
	}
	for _, location := range []string{"s. 152 SJO", "s. 139 SJO"} {
		if keys := buildings.ResolveClassroomKeys(location); len(keys) != 1 || keys[0] != "sjo" {
			t.Errorf("%s resolved keys = %#v, want [sjo]", location, keys)
		}
	}
	keys := buildings.ResolveClassroomKeys("s. 114 GIL/Działowia 1/15, 135/136 bis")
	wantKeys := []string{"114 gil", "1 15 działownia", "135", "136"}
	if len(keys) != len(wantKeys) {
		t.Fatalf("resolved %d keys, want %d: %#v", len(keys), len(wantKeys), keys)
	}
	for i, want := range wantKeys {
		if keys[i] != want {
			t.Errorf("key %d = %q, want %q", i, keys[i], want)
		}
	}

	resolvedBuildings := buildings.GetBuildingsByClassroomKeys(keys)
	wantBuildings := []string{"budynek GIL", "budynek Działownia", "budynek Chemii"}
	if len(resolvedBuildings) != len(wantBuildings) {
		t.Fatalf("resolved %d buildings, want %d", len(resolvedBuildings), len(wantBuildings))
	}
	for i, want := range wantBuildings {
		if resolvedBuildings[i].Key != want {
			t.Errorf("building %d = %q, want %q", i, resolvedBuildings[i].Key, want)
		}
	}
}

func TestToEventKeepsRawLocationAndDoesNotUseLegacyPlaceData(t *testing.T) {
	buildings := &Buildings{
		Classrooms: map[string]string{"131": "Houston"},
		Buildings: map[string]Building{
			"Houston": {Key: "Houston"},
		},
	}
	class := Class{Location: "s. 131", ResolvedClassroomKeys: []string{"131"}}
	event := class.ToEvent(buildings)
	if event.Location != "s. 131" {
		t.Errorf("event location = %q, want original location", event.Location)
	}
	if event.Description != "" {
		t.Errorf("event description = %q, want no legacy map link", event.Description)
	}
}
