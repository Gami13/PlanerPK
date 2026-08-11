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

	"github.com/woodpkrs/PlanerPK/gen/db"
)

type fakePlacesClient struct {
	searchResults []PlaceCandidate
	searchErr     error
	getResult     PlaceCandidate
	getErr        error
	searchCalls   int
	getCalls      int
}

func (f *fakePlacesClient) SearchText(context.Context, string) ([]PlaceCandidate, error) {
	f.searchCalls++
	return f.searchResults, f.searchErr
}

func (f *fakePlacesClient) GetPlace(context.Context, string) (PlaceCandidate, error) {
	f.getCalls++
	return f.getResult, f.getErr
}

func (f *fakePlacesClient) Close() error { return nil }

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

func TestResolveBuildingPlacesPersistsAndHydratesClearWinner(t *testing.T) {
	_, queries := testPlaceQueries(t)
	buildings := testBuildings()
	places := &fakePlacesClient{searchResults: []PlaceCandidate{
		{PlaceID: "chem", Name: "Wydział Chemii Politechniki Krakowskiej", Address: "Warszawska 24, Kraków"},
		{PlaceID: "other", Name: "Budynek Chemii", Address: "Katowice"},
	}}
	queue, err := LoadBuildingReviewQueue(filepath.Join(t.TempDir(), "reviews.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	if err := ResolveBuildingPlaces(context.Background(), queries, buildings, places, queue, now); err != nil {
		t.Fatal(err)
	}
	if places.searchCalls != 1 {
		t.Fatalf("search calls = %d, want 1", places.searchCalls)
	}
	building := buildings.Buildings["budynek Chemii"]
	if building.GooglePlaceID != "chem" || building.Name == "" {
		t.Fatalf("hydrated building = %#v", building)
	}
	if got := buildings.Classrooms["135"]; got != "budynek Chemii" {
		t.Fatalf("classroom mapping = %q", got)
	}

	if err := ResolveBuildingPlaces(context.Background(), queries, testBuildings(), places, queue, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if places.searchCalls != 1 {
		t.Fatalf("fresh cached place searched again: %d calls", places.searchCalls)
	}
}

func TestResolveBuildingPlacesQueuesAmbiguousCandidate(t *testing.T) {
	_, queries := testPlaceQueries(t)
	places := &fakePlacesClient{searchResults: []PlaceCandidate{
		{PlaceID: "a", Name: "Budynek Chemii", Address: "Kraków"},
		{PlaceID: "b", Name: "Budynek Chemii", Address: "Kraków"},
	}}
	path := filepath.Join(t.TempDir(), "reviews.json")
	queue, err := LoadBuildingReviewQueue(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := ResolveBuildingPlaces(context.Background(), queries, testBuildings(), places, queue, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var items []BuildingReviewItem
	if err := json.Unmarshal(contents, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Kind != "place_resolution" {
		t.Fatalf("review items = %#v", items)
	}
}

func TestResolveBuildingPlacesRefreshesExpiredDetails(t *testing.T) {
	_, queries := testPlaceQueries(t)
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	if err := queries.UpsertBuildingPlace(context.Background(), db.UpsertBuildingPlaceParams{BuildingKey: "budynek Chemii", CanonicalName: "budynek Chemii"}); err != nil {
		t.Fatal(err)
	}
	oldID, oldName, oldAddress := "chem", "Old", "Old address"
	expired := now.Add(-time.Hour)
	if err := queries.UpdateBuildingPlace(context.Background(), db.UpdateBuildingPlaceParams{PlaceID: &oldID, DisplayName: &oldName, Address: &oldAddress, Confidence: 0.9, VerifiedAt: &expired, ExpiresAt: &expired, BuildingKey: "budynek Chemii"}); err != nil {
		t.Fatal(err)
	}
	places := &fakePlacesClient{getResult: PlaceCandidate{PlaceID: "chem", Name: "New", Address: "New address"}}
	queue, _ := LoadBuildingReviewQueue(filepath.Join(t.TempDir(), "reviews.json"), "")
	if err := ResolveBuildingPlaces(context.Background(), queries, testBuildings(), places, queue, now); err != nil {
		t.Fatal(err)
	}
	if places.getCalls != 1 || places.searchCalls != 0 {
		t.Fatalf("get/search calls = %d/%d, want 1/0", places.getCalls, places.searchCalls)
	}
	records, err := queries.GetBuildingPlaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if *records[0].DisplayName != "New" || !records[0].ExpiresAt.After(now) {
		t.Fatalf("refreshed record = %#v", records[0])
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

func TestChoosePlaceRejectsLowConfidenceAndSearchFailureDoesNotPersist(t *testing.T) {
	best, _, _, ok := choosePlace("budynek Chemii", []PlaceCandidate{{PlaceID: "bad", Name: "Unrelated", Address: "Kraków"}})
	if ok || best.Score >= minimumPlaceScore {
		t.Fatalf("low-confidence result accepted: %#v", best)
	}
	_, queries := testPlaceQueries(t)
	places := &fakePlacesClient{searchErr: os.ErrDeadlineExceeded}
	queue, _ := LoadBuildingReviewQueue(filepath.Join(t.TempDir(), "reviews.json"), "")
	if err := ResolveBuildingPlaces(context.Background(), queries, testBuildings(), places, queue, time.Now().UTC()); err != nil {
		t.Fatalf("transient Google failure should not fail import: %v", err)
	}
	records, err := queries.GetBuildingPlaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].PlaceID != nil {
		t.Fatalf("failure persisted a place: %#v", records)
	}
}

func TestResolvedPlaceReviewNoLongerInvalidatesCachedPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reviews.json")
	queue, err := LoadBuildingReviewQueue(path, "")
	if err != nil {
		t.Fatal(err)
	}
	queue.RecordPlaceResolution("budynek Chemii", 0.7, []scoredPlace{{PlaceCandidate: PlaceCandidate{PlaceID: "candidate"}, Score: 0.7}}, []string{"low confidence"})
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
