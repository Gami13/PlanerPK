package types

import (
	"testing"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
)

func TestChoosePlaceUsesGooglePlaceTypes(t *testing.T) {
	best, _, _, ok := choosePlace("budynek Chemii", []*placespb.Place{
		googlePlace("best", "Wydział Chemii Politechniki Krakowskiej", "Warszawska 24, Kraków"),
		googlePlace("other", "Budynek Chemii", "Katowice"),
	})
	if !ok || best.Place.GetId() != "best" {
		t.Fatalf("winner = %#v, accepted = %v", best, ok)
	}

	_, _, _, ok = choosePlace("budynek Chemii", []*placespb.Place{
		googlePlace("a", "Budynek Chemii", "Kraków"),
		googlePlace("b", "Budynek Chemii", "Kraków"),
	})
	if ok {
		t.Fatal("tied candidates were accepted")
	}

	best, _, _, ok = choosePlace("budynek Chemii", []*placespb.Place{googlePlace("bad", "Unrelated", "Kraków")})
	if ok || best.Score >= minimumPlaceScore {
		t.Fatalf("low-confidence result accepted: %#v", best)
	}
}

func TestChoosePlacePrefersBuildingOverFloorLevelUnit(t *testing.T) {
	best, _, _, ok := choosePlace("budynek Houston, ul. Warszawska", []*placespb.Place{
		googlePlace("biuro", "Biuro Karier Politechniki Krakowskiej", "Warszawska 24, Kraków, Polska Budynek Houston II piętro, 31-155 Kraków"),
		googlePlace("transport", "Transport i Logistyka", "Warszawska 24, Kraków, Polska Budynek Houston, 31-155 Kraków"),
	})
	if !ok || best.Place.GetId() != "transport" {
		t.Fatalf("winner = %#v, accepted = %v", best, ok)
	}
}

func TestChoosePlaceRecognizesSJOFullName(t *testing.T) {
	best, _, _, ok := choosePlace("budynek SJO ul. Warszawska", []*placespb.Place{
		googlePlace("sjo", "Studium Języków Obcych Politechniki Krakowskiej", "Warszawska 24, 31-155 Kraków, Polska"),
		googlePlace("other", "Centrum Sportu Politechniki Krakowskiej", "Warszawska 24, 31-155 Kraków, Polska"),
	})
	if !ok || best.Place.GetId() != "sjo" {
		t.Fatalf("winner = %#v, accepted = %v", best, ok)
	}
}

func TestHasBuildingInitialismEvidence(t *testing.T) {
	if !hasBuildingInitialismEvidence("budynek SJO ul. Warszawska", "Studium Języków Obcych Politechniki Krakowskiej") {
		t.Fatal("SJO should match the display-name initialism")
	}
	if !hasBuildingInitialismEvidence("budynek PK", "Politechnika Krakowska") {
		t.Fatal("PK should match the display-name initialism")
	}
	if hasBuildingInitialismEvidence("budynek SJO ul. Warszawska", "Studium Języków Politechniki Krakowskiej") {
		t.Fatal("SJO should not match a different display-name initialism")
	}
}
