package buildings

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/internal/testutil"
	"github.com/woodpkrs/PlanerPK/legend"
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
	return BuildBuildingsFromLegend(&legend.Legend{Locations: map[string]legend.Location{
		"s. 135": {Shorthand: "s. 135", FullName: "budynek Chemii"},
	}})
}

var testPlaceTime = time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)

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
	if err := queries.UpsertBuildingPlace(context.Background(), db.UpsertBuildingPlaceParams{BuildingKey: "budynek Chemii", CanonicalName: "budynek Chemii"}); err != nil {
		t.Fatal(err)
	}
	if err := storePlace(context.Background(), queries, "budynek Chemii", googlePlace("chem", "Wydział Chemii PK", "Kraków"), 1, testPlaceTime); err != nil {
		t.Fatal(err)
	}
	buildings := testBuildings()
	if err := ResolveBuildingPlaces(context.Background(), queries, buildings, nil, testPlaceTime.Add(time.Hour)); err != nil {
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
	if err := queries.UpsertBuildingPlace(context.Background(), db.UpsertBuildingPlaceParams{BuildingKey: "budynek Chemii", CanonicalName: "budynek Chemii"}); err != nil {
		t.Fatal(err)
	}
	if err := storePlace(context.Background(), queries, "budynek Chemii", googlePlace("chem", "New", "New address"), 0.9, testPlaceTime); err != nil {
		t.Fatal(err)
	}
	records, err := queries.GetBuildingPlaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if *records[0].DisplayName != "New" || !records[0].ExpiresAt.After(testPlaceTime) {
		t.Fatalf("stored record = %#v", records[0])
	}
}

func TestRejectedPlaceResolutionIsLoggedAtErrorLevel(t *testing.T) {
	handler := testutil.CaptureLogs(t)
	runnerUp := googlePlace("runner-up", "Wydział Architektury PK", "Warszawska 24, Kraków")

	logRejectedPlaceResolution("budynek Chemii", placeResolution{
		Best: scoredPlace{
			Place: googlePlace("best", "Budynek Chemii", "Kraków"),
			Score: 0.85,
		},
		Ranked: []scoredPlace{
			{Place: googlePlace("best", "Budynek Chemii", "Kraków"), Score: 0.85},
			{Place: runnerUp, Score: 0.80},
		},
		Reasons: []string{"candidate confidence is below 0.70"},
	})

	if len(handler.Records) != 1 || handler.Records[0].Level != slog.LevelError || handler.Records[0].Message != "building_place_resolution.rejected" {
		t.Fatalf("records = %#v", handler.Records)
	}
	attrs := make(map[string]slog.Value)
	handler.Records[0].Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value
		return true
	})
	for key, want := range map[string]string{
		"rejected_place_name":     "Budynek Chemii",
		"rejected_place_address":  "Kraków",
		"rejected_place_id":       "best",
		"rejected_place_maps_url": "https://www.google.com/maps/search/?api=1&query=Politechnika+Krakowska&query_place_id=best",
		"runner_up_name":          "Wydział Architektury PK",
		"runner_up_address":       "Warszawska 24, Kraków",
		"runner_up_place_id":      "runner-up",
		"runner_up_maps_url":      "https://www.google.com/maps/search/?api=1&query=Politechnika+Krakowska&query_place_id=runner-up",
	} {
		if got := attrs[key].String(); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if got := attrs["runner_up_confidence"].Float64(); got != 0.80 {
		t.Errorf("runner_up_confidence = %.2f, want 0.80", got)
	}
}
