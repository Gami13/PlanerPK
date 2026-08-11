package types

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"github.com/woodpkrs/PlanerPK/gen/db"
	localizedtext "google.golang.org/genproto/googleapis/type/localized_text"
)

func googlePlace(id, name, address string) *placespb.Place {
	return &placespb.Place{
		Id:               id,
		DisplayName:      &localizedtext.LocalizedText{Text: name},
		FormattedAddress: address,
	}
}

func testPlaceQueries(t *testing.T) (*sql.DB, db.Querier) {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	_, err = database.Exec(`
		CREATE TABLE building_places (
			building_key TEXT PRIMARY KEY, canonical_name TEXT NOT NULL, place_id TEXT UNIQUE,
			display_name TEXT, address TEXT, confidence REAL NOT NULL DEFAULT 0,
			verified_at TIMESTAMP, expires_at TIMESTAMP
		);`)
	if err != nil {
		t.Fatal(err)
	}
	return database, db.New(database)
}

func testBuildings() *Buildings {
	return BuildBuildingsFromLegend(&Legend{Locations: map[string]Location{
		"s. 135": {Shorthand: "s. 135", FullName: "budynek Chemii"},
	}})
}

func TestBuildingFromPlace(t *testing.T) {
	placeID, displayName, address := "chem", "Wydział Chemii PK", "Kraków"
	building := BuildingFromPlace(db.BuildingPlace{
		BuildingKey:   "chemia",
		CanonicalName: "budynek Chemii",
		PlaceID:       &placeID,
		DisplayName:   &displayName,
		Address:       &address,
	})
	if building != (Building{CanonicalName: "budynek Chemii", GooglePlaceID: placeID, Name: displayName, Address: address}) {
		t.Fatalf("BuildingFromPlace() = %#v", building)
	}
}

func TestBuildingPlaceSearchRequestReturnsTwentyCandidates(t *testing.T) {
	request := buildingPlaceSearchRequest("budynek Houston, ul. Warszawska")
	if request.GetMaxResultCount() != buildingPlaceSearchResultCount {
		t.Errorf("MaxResultCount = %d, want %d", request.GetMaxResultCount(), buildingPlaceSearchResultCount)
	}
	if request.GetTextQuery() != "budynek Houston, ul. Warszawska Politechnika Krakowska" {
		t.Errorf("TextQuery = %q", request.GetTextQuery())
	}
}

func TestResolveBuildingPlacesHydratesConfirmedCache(t *testing.T) {
	_, queries := testPlaceQueries(t)
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	if err := queries.UpsertBuildingPlace(context.Background(), db.UpsertBuildingPlaceParams{BuildingKey: "budynek Chemii", CanonicalName: "budynek Chemii"}); err != nil {
		t.Fatal(err)
	}
	if err := storePlace(context.Background(), queries, "budynek Chemii", googlePlace("chem", "Wydział Chemii PK", "Kraków"), 1, now); err != nil {
		t.Fatal(err)
	}
	buildings := testBuildings()
	queue := &BuildingReviewQueue{}
	if err := ResolveBuildingPlaces(context.Background(), queries, buildings, nil, queue, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	building := buildings.BuildingsByKey["budynek Chemii"]
	if building.GooglePlaceID != "chem" || building.Name != "Wydział Chemii PK" {
		t.Fatalf("hydrated building = %#v", building)
	}
	if got := buildings.BuildingKeyByClassroomKey["135"]; got != "budynek Chemii" {
		t.Fatalf("classroom mapping = %q", got)
	}
}

func TestStorePlaceRefreshesExpiry(t *testing.T) {
	_, queries := testPlaceQueries(t)
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	if err := queries.UpsertBuildingPlace(context.Background(), db.UpsertBuildingPlaceParams{BuildingKey: "budynek Chemii", CanonicalName: "budynek Chemii"}); err != nil {
		t.Fatal(err)
	}
	if err := storePlace(context.Background(), queries, "budynek Chemii", googlePlace("chem", "New", "New address"), 0.9, now); err != nil {
		t.Fatal(err)
	}
	records, err := queries.GetBuildingPlaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if *records[0].DisplayName != "New" || !records[0].ExpiresAt.After(now) {
		t.Fatalf("stored record = %#v", records[0])
	}
}
