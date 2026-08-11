package types

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
			best, _, _, accepted := choosePlace(tt.canonicalName, tt.places)
			if accepted != tt.wantAccepted {
				t.Fatalf("accepted = %v, want %v; winner = %#v", accepted, tt.wantAccepted, best)
			}
			if tt.wantID != "" && best.Place.GetId() != tt.wantID {
				t.Fatalf("winner ID = %q, want %q", best.Place.GetId(), tt.wantID)
			}
			if tt.name == "low confidence" && best.Score >= minimumPlaceScore {
				t.Fatalf("low-confidence result scored %.2f, want below %.2f", best.Score, minimumPlaceScore)
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
