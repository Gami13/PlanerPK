package buildings_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/types"
)

func assertResolvedKeys(t *testing.T, cache *buildings.Buildings, location string, want ...string) []string {
	t.Helper()
	got := cache.ResolveClassroomKeys(location)
	if !slices.Equal(got, want) {
		t.Fatalf("%q resolved keys = %#v, want %#v", location, got, want)
	}
	return got
}

func assertBuildingNames(t *testing.T, got []buildings.Building, want ...string) {
	t.Helper()
	gotNames := make([]string, len(got))
	for i := range got {
		gotNames[i] = got[i].CanonicalName
	}
	if !slices.Equal(gotNames, want) {
		t.Fatalf("resolved buildings = %#v, want %#v", gotNames, want)
	}
}

func TestBuildingsResolveAndLookupUsesLegendOnly(t *testing.T) {
	buildingCache := buildings.BuildBuildingsFromLegend(&legend.Legend{Locations: map[string]legend.Location{
		"s. 114 GIL":         {Shorthand: "s. 114 GIL", FullName: "budynek GIL"},
		"s. 135(L2)":         {Shorthand: "s. 135(L2)", FullName: "budynek Chemii"},
		"s. 136(L1,)":        {Shorthand: "s. 136(L1,)", FullName: "budynek Chemii"},
		"s. 1/15 Działownia": {Shorthand: "s. 1/15 Działownia", FullName: "budynek Działownia"},
		"s. 138":             {Shorthand: "s. 138", FullName: "budynek SJO ul. Warszawska"},
	}})

	if buildingKey := buildingCache.BuildingKeyByClassroomKey["135"]; buildingKey != "budynek Chemii" {
		t.Errorf("135 building key = %q, want budynek Chemii", buildingKey)
	}
	if buildingKey := buildingCache.BuildingKeyByClassroomKey["sjo"]; buildingKey != "budynek SJO ul. Warszawska" {
		t.Errorf("SJO building key = %q, want SJO building", buildingKey)
	}
	assertResolvedKeys(t, buildingCache, "s. SJO", "sjo")
	for _, location := range []string{"s. 152 SJO", "s. 139 SJO"} {
		assertResolvedKeys(t, buildingCache, location, "sjo")
	}
	keys := assertResolvedKeys(t, buildingCache, "s. 114 GIL/Działowia 1/15, 135/136 bis", "114 gil", "1 15 działownia", "135", "136")

	assertBuildingNames(t, buildingCache.GetBuildingsByClassroomKeys(keys), "budynek GIL", "budynek Działownia", "budynek Chemii")
}

func TestToEventKeepsRawLocationAndDoesNotUseLegacyPlaceData(t *testing.T) {
	buildingCache := &buildings.Buildings{
		BuildingKeyByClassroomKey: map[string]string{"131": "Houston"},
		BuildingsByKey: map[string]buildings.Building{
			"Houston": {},
		},
	}
	class := types.Class{Location: "s. 131", ResolvedClassroomKeys: []string{"131"}}
	event := class.ToEvent(buildingCache, types.ClassSchedule{})
	if event.Location != "s. 131" {
		t.Errorf("event location = %q, want original location", event.Location)
	}
	if event.Description != "" {
		t.Errorf("event description = %q, want no legacy map link", event.Description)
	}
}

func TestToEventAddsUniqueConfirmedBuildingLinks(t *testing.T) {
	buildingCache := &buildings.Buildings{
		BuildingKeyByClassroomKey: map[string]string{"135": "Chemia", "136": "Chemia", "114": "GIL"},
		BuildingsByKey: map[string]buildings.Building{
			"Chemia": {GooglePlaceID: "chem", Name: "Chemia", Address: "Kraków"},
			"GIL":    {GooglePlaceID: "gil", Name: "GIL", Address: "Kraków"},
		},
	}
	class := types.Class{Location: "s. 135/136, s. 114", ResolvedClassroomKeys: []string{"135", "136", "114"}}
	event := class.ToEvent(buildingCache, types.ClassSchedule{})
	if event.Location != class.Location {
		t.Fatalf("location = %q, want original %q", event.Location, class.Location)
	}
	if strings.Count(event.Description, "query_place_id=chem") != 1 || strings.Count(event.Description, "query_place_id=gil") != 1 {
		t.Fatalf("description = %q, want one link per building", event.Description)
	}
}
