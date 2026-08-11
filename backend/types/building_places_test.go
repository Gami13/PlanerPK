package types

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
		);
		CREATE TABLE classroom_buildings (
			classroom_key TEXT PRIMARY KEY,
			building_key TEXT NOT NULL REFERENCES building_places(building_key) ON DELETE CASCADE
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
	queue, err := LoadBuildingReviewQueue(filepath.Join(t.TempDir(), "reviews.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := ResolveBuildingPlaces(context.Background(), queries, buildings, nil, queue, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	building := buildings.Buildings["budynek Chemii"]
	if building.GooglePlaceID != "chem" || building.Name != "Wydział Chemii PK" {
		t.Fatalf("hydrated building = %#v", building)
	}
	if got := buildings.Classrooms["135"]; got != "budynek Chemii" {
		t.Fatalf("classroom mapping = %q", got)
	}
}

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

func TestBuildingReviewDecisionsApproveRejectAndInvalidate(t *testing.T) {
	_, queries := testPlaceQueries(t)
	ctx := context.Background()
	for _, key := range []string{"approved", "rejected", "invalidated"} {
		if err := queries.UpsertBuildingPlace(ctx, db.UpsertBuildingPlaceParams{BuildingKey: key, CanonicalName: key}); err != nil {
			t.Fatal(err)
		}
	}
	queue := &BuildingReviewQueue{items: []BuildingReviewItem{
		{BuildingKey: "approved", Status: "approved", ApprovedPlaceID: "p1", Candidates: []BuildingReviewCandidate{{PlaceID: "p1", Name: "Approved", Address: "Kraków"}}},
		{BuildingKey: "rejected", Status: "rejected"},
		{BuildingKey: "invalidated", Status: "invalidated"},
	}}
	oldID := "old"
	if err := queries.UpdateBuildingPlace(ctx, db.UpdateBuildingPlaceParams{PlaceID: &oldID, BuildingKey: "invalidated"}); err != nil {
		t.Fatal(err)
	}
	suppressed := queue.ApplyDecisions(ctx, queries, time.Now().UTC())
	if !suppressed["rejected"] {
		t.Fatal("rejected building was not suppressed")
	}
	records, err := queries.GetBuildingPlaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	byKey := map[string]db.BuildingPlace{}
	for _, record := range records {
		byKey[record.BuildingKey] = record
	}
	if byKey["approved"].PlaceID == nil || *byKey["approved"].PlaceID != "p1" {
		t.Fatalf("approval not applied: %#v", byKey["approved"])
	}
	if byKey["invalidated"].PlaceID != nil {
		t.Fatalf("invalidation not applied: %#v", byKey["invalidated"])
	}
}

func TestResolvedPlaceReviewNoLongerInvalidatesCachedPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reviews.json")
	queue, err := LoadBuildingReviewQueue(path, "")
	if err != nil {
		t.Fatal(err)
	}
	queue.RecordPlaceResolution("budynek Chemii", 0.7, []scoredPlace{{Place: googlePlace("candidate", "Chemia", "Kraków"), Score: 0.7}}, []string{"low confidence"})
	queue.MarkPlaceResolved("budynek Chemii")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var items []BuildingReviewItem
	if err := json.Unmarshal(contents, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Status != "resolved" {
		t.Fatalf("review item = %#v, want resolved", items)
	}
}

func TestReviewWebhookFailureDoesNotPreventQueuePersistence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "reviews.json")
	queue, err := LoadBuildingReviewQueue(path, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	queue.RecordPlaceResolution("budynek Chemii", 0, nil, []string{"no candidates"})
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("queue was not persisted after webhook failure: %v", err)
	}
}
