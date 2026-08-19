package buildings

import (
	"testing"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
)

func TestChoosePlace(t *testing.T) {
	tests := []struct {
		name          string
		canonicalName string
		places        []*placespb.Place
		wantID        string
		wantAccepted  bool
	}{
		{
			name:          "strong building match",
			canonicalName: "budynek Chemii",
			places: []*placespb.Place{
				googlePlace("best", "Wydział Chemii Politechniki Krakowskiej", "Warszawska 24, Kraków"),
				googlePlace("other", "Budynek Chemii", "Katowice"),
			},
			wantID:       "best",
			wantAccepted: true,
		},
		{
			name:          "tie",
			canonicalName: "budynek Chemii",
			places: []*placespb.Place{
				googlePlace("a", "Budynek Chemii", "Kraków"),
				googlePlace("b", "Budynek Chemii", "Kraków"),
			},
		},
		{
			name:          "low confidence",
			canonicalName: "budynek Chemii",
			places:        []*placespb.Place{googlePlace("bad", "Unrelated", "Kraków")},
		},
		{
			name:          "building over floor-level unit",
			canonicalName: "budynek Houston, ul. Warszawska",
			places: []*placespb.Place{
				googlePlace("biuro", "Biuro Karier Politechniki Krakowskiej", "Warszawska 24, Kraków, Polska Budynek Houston II piętro, 31-155 Kraków"),
				googlePlace("transport", "Transport i Logistyka", "Warszawska 24, Kraków, Polska Budynek Houston, 31-155 Kraków"),
			},
			wantID:       "transport",
			wantAccepted: true,
		},
		{
			name:          "building initialism",
			canonicalName: "budynek SJO ul. Warszawska",
			places: []*placespb.Place{
				googlePlace("sjo", "Studium Języków Obcych Politechniki Krakowskiej", "Warszawska 24, 31-155 Kraków, Polska"),
				googlePlace("other", "Centrum Sportu Politechniki Krakowskiej", "Warszawska 24, 31-155 Kraków, Polska"),
			},
			wantID:       "sjo",
			wantAccepted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolution := choosePlace(tt.canonicalName, tt.places)
			if resolution.Accepted != tt.wantAccepted {
				t.Fatalf("accepted = %v, want %v; winner = %#v", resolution.Accepted, tt.wantAccepted, resolution.Best)
			}
			if tt.wantID != "" && resolution.Best.Place.GetId() != tt.wantID {
				t.Fatalf("winner ID = %q, want %q", resolution.Best.Place.GetId(), tt.wantID)
			}
			if tt.name == "low confidence" && resolution.Best.Score >= minimumPlaceScore {
				t.Fatalf("low-confidence result scored %.2f, want below %.2f", resolution.Best.Score, minimumPlaceScore)
			}
		})
	}
}

func TestHasBuildingInitialismEvidence(t *testing.T) {
	tests := []struct {
		canonicalName string
		displayName   string
		want          bool
	}{
		{"budynek SJO ul. Warszawska", "Studium Języków Obcych Politechniki Krakowskiej", true},
		{"budynek PK", "Politechnika Krakowska", true},
		{"budynek SJO ul. Warszawska", "Studium Języków Politechniki Krakowskiej", false},
	}
	for _, tt := range tests {
		if got := hasBuildingInitialismEvidence(tt.canonicalName, tt.displayName); got != tt.want {
			t.Errorf("hasBuildingInitialismEvidence(%q, %q) = %v, want %v", tt.canonicalName, tt.displayName, got, tt.want)
		}
	}
}

func TestKatedraContextInfluencesPlaceScore(t *testing.T) {
	withoutKatedra := googlePlace(
		"building",
		"Wydział Architektury Politechniki Krakowskiej",
		"Warszawska 24, Kraków",
	)
	withKatedra := googlePlace(
		"department",
		"Katedra Architektury Politechniki Krakowskiej",
		"Warszawska 24, Kraków",
	)

	unexpectedScore, unexpectedReasons := scorePlace("budynek Architektury", withKatedra)
	ordinaryScore, _ := scorePlace("budynek Architektury", withoutKatedra)
	if unexpectedScore >= ordinaryScore {
		t.Fatalf("unexpected katedra score = %.2f, want below ordinary score %.2f", unexpectedScore, ordinaryScore)
	}
	if !containsReason(unexpectedReasons, "unexpected katedra context") {
		t.Fatalf("unexpected katedra reasons = %v", unexpectedReasons)
	}

	matchingScore, matchingReasons := scorePlace("Katedra Architektury", withKatedra)
	missingScore, _ := scorePlace("Katedra Architektury", withoutKatedra)
	if matchingScore <= missingScore {
		t.Fatalf("matching katedra score = %.2f, want above missing-katedra score %.2f", matchingScore, missingScore)
	}
	if !containsReason(matchingReasons, "matching katedra context") {
		t.Fatalf("matching katedra reasons = %v", matchingReasons)
	}
}

func containsReason(reasons []string, want string) bool {
	for _, reason := range reasons {
		if reason == want {
			return true
		}
	}
	return false
}
