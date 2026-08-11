package classroommatch

import "testing"

func testMatcher() *Matcher {
	return New([]Candidate{
		{ClassroomKey: "s. 1/15 Działownia", BuildingKey: "Działownia"},
		{ClassroomKey: "s. 114 GIL", BuildingKey: "GIL"},
		{ClassroomKey: "s. 135(L2)", BuildingKey: "Chemia"},
		{ClassroomKey: "s. 136(L1,)", BuildingKey: "Chemia"},
		{ClassroomKey: "s. 151(L3)", BuildingKey: "Chemia"},
	})
}

func TestResolveSupportsAliasesCorrectionsAndCompoundLocations(t *testing.T) {
	matcher := testMatcher()
	results := matcher.Resolve("s. 114 GIL/Działowia 1/15, 136 bis/ZDALNIE")

	if len(results) != 4 {
		t.Fatalf("Resolve returned %d results, want 4", len(results))
	}
	wantKinds := []MatchKind{MatchExact, MatchCorrected, MatchExact, MatchIgnored}
	for i, want := range wantKinds {
		if results[i].Kind != want {
			t.Errorf("result %d kind = %q, want %q", i, results[i].Kind, want)
		}
	}
	if results[0].BuildingKey != "GIL" {
		t.Errorf("first building = %q, want GIL", results[0].BuildingKey)
	}
	if results[1].ClassroomKey != "1 15 działownia" {
		t.Errorf("corrected key = %q, want normalized Działownia key", results[1].ClassroomKey)
	}
	if results[2].ClassroomKey != "136" {
		t.Errorf("bis alias key = %q, want 136", results[2].ClassroomKey)
	}
}

func TestResolveRejectsShortFuzzyAndAmbiguousMatches(t *testing.T) {
	matcher := testMatcher()
	if result := matcher.Resolve("s. S4")[0]; result.Kind != MatchUnmatched {
		t.Errorf("short room kind = %q, want unmatched", result.Kind)
	}

	ambiguous := New([]Candidate{
		{ClassroomKey: "sala alfa", BuildingKey: "A"},
		{ClassroomKey: "sala alga", BuildingKey: "B"},
	})
	if result := ambiguous.Resolve("sala alpa")[0]; result.Kind != MatchAmbiguous {
		t.Errorf("ambiguous room kind = %q, want ambiguous", result.Kind)
	}
}

func TestNewExcludesCollidingAliases(t *testing.T) {
	matcher := New([]Candidate{
		{ClassroomKey: "s. 100", BuildingKey: "A"},
		{ClassroomKey: "100", BuildingKey: "B"},
	})
	if _, exists := matcher.ClassroomBuildings()["100"]; exists {
		t.Fatal("colliding classroom key remained in the exact lookup map")
	}
	if result := matcher.Resolve("s. 100")[0]; result.Kind != MatchUnmatched {
		t.Errorf("colliding room kind = %q, want unmatched", result.Kind)
	}
}
