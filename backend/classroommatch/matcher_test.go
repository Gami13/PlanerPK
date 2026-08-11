package classroommatch

import (
	"slices"
	"testing"
)

func testMatcher() *Matcher {
	return New([]Candidate{
		{ClassroomKey: "s. 1/15 Działownia", BuildingKey: "Działownia"},
		{ClassroomKey: "s. 114 GIL", BuildingKey: "GIL"},
		{ClassroomKey: "s. 135(L2)", BuildingKey: "Chemia"},
		{ClassroomKey: "s. 136(L1,)", BuildingKey: "Chemia"},
		{ClassroomKey: "s. 151(L3)", BuildingKey: "Chemia"},
	}, nil)
}

func TestResolveSupportsAliasesCorrectionsAndCompoundLocations(t *testing.T) {
	matcher := testMatcher()
	results := matcher.Resolve("s. 114 GIL/Działowia 1/15, 136 bis/ZDALNIE")

	if len(results) != 4 {
		t.Fatalf("Resolve returned %d results, want 4", len(results))
	}
	wantKinds := []MatchKind{MatchExact, MatchCorrected, MatchExact, MatchIgnored}
	gotKinds := make([]MatchKind, len(results))
	for i := range results {
		gotKinds[i] = results[i].Kind
	}
	if !slices.Equal(gotKinds, wantKinds) {
		t.Errorf("result kinds = %v, want %v", gotKinds, wantKinds)
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
	}, nil)
	if result := ambiguous.Resolve("sala alpa")[0]; result.Kind != MatchAmbiguous {
		t.Errorf("ambiguous room kind = %q, want ambiguous", result.Kind)
	}
}

func TestNewExcludesCollidingAliases(t *testing.T) {
	matcher := New([]Candidate{
		{ClassroomKey: "s. 100", BuildingKey: "A"},
		{ClassroomKey: "100", BuildingKey: "B"},
	}, nil)
	if _, exists := matcher.ClassroomBuildings()["100"]; exists {
		t.Fatal("colliding classroom key remained in the exact lookup map")
	}
	if result := matcher.Resolve("s. 100")[0]; result.Kind != MatchUnmatched {
		t.Errorf("colliding room kind = %q, want unmatched", result.Kind)
	}
}

func TestBuildingAliasesDoNotDependOnExactAliasInsertionOrder(t *testing.T) {
	matcher := New(
		[]Candidate{{ClassroomKey: "s. SJO", BuildingKey: "SJO"}},
		[]Candidate{{ClassroomKey: "SJO", BuildingKey: "SJO"}},
	)
	result := matcher.Resolve("s. 152 SJO")[0]
	if result.Kind != MatchExact || result.BuildingKey != "SJO" {
		t.Fatalf("same-building acronym result = %#v, want exact SJO match", result)
	}
	if got := matcher.ClassroomBuildings()["sjo"]; got != "SJO" {
		t.Fatalf("SJO lookup = %q, want SJO", got)
	}

	conflicting := New(
		[]Candidate{{ClassroomKey: "s. SJO", BuildingKey: "Other"}},
		[]Candidate{{ClassroomKey: "SJO", BuildingKey: "SJO"}},
	)
	if result := conflicting.Resolve("s. 152 SJO")[0]; result.Kind == MatchExact {
		t.Fatalf("conflicting building acronym was accepted: %#v", result)
	}
}
