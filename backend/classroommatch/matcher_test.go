package classroommatch

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/woodpkrs/PlanerPK/internal/testutil"
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

func ambiguousMatcher() *Matcher {
	return New([]Candidate{
		{ClassroomKey: "sala alfa", BuildingKey: "A"},
		{ClassroomKey: "sala alga", BuildingKey: "B"},
	}, nil)
}

func assertMatchKinds(t *testing.T, results []Result, want ...MatchKind) {
	t.Helper()
	got := make([]MatchKind, len(results))
	for i := range results {
		got[i] = results[i].Kind
	}
	if !slices.Equal(got, want) {
		t.Errorf("result kinds = %v, want %v", got, want)
	}
}

func TestResolveSupportsAliasesCorrectionsAndCompoundLocations(t *testing.T) {
	matcher := testMatcher()
	results := matcher.Resolve("s. 114 GIL/Działowia 1/15, 136 bis/ZDALNIE")

	if len(results) != 4 {
		t.Fatalf("Resolve returned %d results, want 4", len(results))
	}
	assertMatchKinds(t, results, MatchExact, MatchCorrected, MatchExact, MatchIgnored)
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
	for _, test := range []struct {
		name    string
		matcher func() *Matcher
		input   string
		want    MatchKind
	}{
		{name: "short room", matcher: testMatcher, input: "s. S4", want: MatchUnmatched},
		{name: "ambiguous room", matcher: ambiguousMatcher, input: "sala alpa", want: MatchAmbiguous},
	} {
		t.Run(test.name, func(t *testing.T) {
			if result := test.matcher().Resolve(test.input)[0]; result.Kind != test.want {
				t.Errorf("match kind = %q, want %q", result.Kind, test.want)
			}
		})
	}
}

func TestUnresolvedMatchesAreLoggedAtWarnLevel(t *testing.T) {
	handler := testutil.CaptureLogs(t)

	matcher := testMatcher()
	matcher.Resolve("s. S4")
	ambiguousMatcher().Resolve("sala alpa")

	var messages []string
	for _, record := range handler.Records {
		if record.Level == slog.LevelWarn {
			messages = append(messages, record.Message)
		}
	}
	if !slices.Equal(messages, []string{"classroom_resolution.unmatched", "classroom_resolution.ambiguous"}) {
		t.Fatalf("error messages = %v", messages)
	}
}

func TestSortedCandidatesAndNormalizePreserveExpectedKeys(t *testing.T) {
	candidates := []Candidate{
		{ClassroomKey: "z", BuildingKey: "B"},
		{ClassroomKey: "alpha", BuildingKey: "Z"},
		{ClassroomKey: "alpha", BuildingKey: "A"},
		{ClassroomKey: "z", BuildingKey: "A"},
		{ClassroomKey: "alpha", BuildingKey: "A"},
	}
	wantCandidates := []Candidate{
		{ClassroomKey: "alpha", BuildingKey: "A"},
		{ClassroomKey: "alpha", BuildingKey: "A"},
		{ClassroomKey: "alpha", BuildingKey: "Z"},
		{ClassroomKey: "z", BuildingKey: "A"},
		{ClassroomKey: "z", BuildingKey: "B"},
	}
	if got := sortedCandidates(candidates); !slices.Equal(got, wantCandidates) {
		t.Errorf("sortedCandidates() = %#v, want %#v", got, wantCandidates)
	}
	if candidates[0] != (Candidate{ClassroomKey: "z", BuildingKey: "B"}) {
		t.Errorf("sortedCandidates mutated input: %#v", candidates)
	}

	for _, test := range []struct {
		input string
		want  string
	}{
		{input: "", want: ""},
		{input: "---/()", want: ""},
		{input: " s. Sala Żółć-101 ", want: "101 sala żółć"},
		{input: "s. 114 GIL", want: "114 gil"},
	} {
		if got := Normalize(test.input); got != test.want {
			t.Errorf("Normalize(%q) = %q, want %q", test.input, got, test.want)
		}
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
