package types

import (
	"slices"
	"strings"
	"testing"
)

func TestBuildingsResolveAndLookupUsesLegendOnly(t *testing.T) {
	buildings := BuildBuildingsFromLegend(&Legend{Locations: map[string]Location{
		"s. 114 GIL":         {Shorthand: "s. 114 GIL", FullName: "budynek GIL"},
		"s. 135(L2)":         {Shorthand: "s. 135(L2)", FullName: "budynek Chemii"},
		"s. 136(L1,)":        {Shorthand: "s. 136(L1,)", FullName: "budynek Chemii"},
		"s. 1/15 Działownia": {Shorthand: "s. 1/15 Działownia", FullName: "budynek Działownia"},
		"s. 138":             {Shorthand: "s. 138", FullName: "budynek SJO ul. Warszawska"},
	}})

	if buildingKey := buildings.BuildingKeyByClassroomKey["135"]; buildingKey != "budynek Chemii" {
		t.Errorf("135 building key = %q, want budynek Chemii", buildingKey)
	}
	if buildingKey := buildings.BuildingKeyByClassroomKey["sjo"]; buildingKey != "budynek SJO ul. Warszawska" {
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
	if !slices.Equal(keys, wantKeys) {
		t.Fatalf("resolved keys = %#v, want %#v", keys, wantKeys)
	}

	resolvedBuildings := buildings.GetBuildingsByClassroomKeys(keys)
	wantBuildings := []string{"budynek GIL", "budynek Działownia", "budynek Chemii"}
	gotBuildings := make([]string, len(resolvedBuildings))
	for i := range resolvedBuildings {
		gotBuildings[i] = resolvedBuildings[i].CanonicalName
	}
	if !slices.Equal(gotBuildings, wantBuildings) {
		t.Fatalf("resolved buildings = %#v, want %#v", gotBuildings, wantBuildings)
	}
}

func TestToEventKeepsRawLocationAndDoesNotUseLegacyPlaceData(t *testing.T) {
	buildings := &Buildings{
		BuildingKeyByClassroomKey: map[string]string{"131": "Houston"},
		BuildingsByKey: map[string]Building{
			"Houston": {},
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

func TestToEventAddsUniqueConfirmedBuildingLinks(t *testing.T) {
	buildings := &Buildings{
		BuildingKeyByClassroomKey: map[string]string{"135": "Chemia", "136": "Chemia", "114": "GIL"},
		BuildingsByKey: map[string]Building{
			"Chemia": {GooglePlaceID: "chem", Name: "Chemia", Address: "Kraków"},
			"GIL":    {GooglePlaceID: "gil", Name: "GIL", Address: "Kraków"},
		},
	}
	class := Class{Location: "s. 135/136, s. 114", ResolvedClassroomKeys: []string{"135", "136", "114"}}
	event := class.ToEvent(buildings)
	if event.Location != class.Location {
		t.Fatalf("location = %q, want original %q", event.Location, class.Location)
	}
	if strings.Count(event.Description, "query_place_id=chem") != 1 || strings.Count(event.Description, "query_place_id=gil") != 1 {
		t.Fatalf("description = %q, want one link per building", event.Description)
	}
}
