package types

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
)

func TestBuildingReviewDecisionsApproveRejectAndInvalidate(t *testing.T) {
	_, queries := testPlaceQueries(t)
	ctx := context.Background()
	for _, key := range []string{"approved", "rejected", "invalidated"} {
		if err := queries.UpsertBuildingPlace(ctx, db.UpsertBuildingPlaceParams{BuildingKey: key, CanonicalName: key}); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "reviews.json")
	queue := &BuildingReviewQueue{path: path, items: []BuildingReviewItem{
		{BuildingKey: "approved", Status: ReviewStatusApproved, ApprovedPlaceID: "p1", Candidates: []BuildingReviewCandidate{{PlaceID: "p1", Name: "Approved", Address: "Kraków"}}},
		{BuildingKey: "rejected", Status: ReviewStatusRejected},
		{BuildingKey: "invalidated", Status: ReviewStatusInvalidated},
	}}
	oldID := "old"
	if err := queries.UpdateBuildingPlace(ctx, db.UpdateBuildingPlaceParams{PlaceID: &oldID, BuildingKey: "invalidated"}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	suppressed := queue.ApplyDecisions(ctx, queries, now)
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
	if queue.items[0].Status != ReviewStatusResolved || queue.items[2].Status != ReviewStatusPending {
		t.Fatalf("decision statuses were not consumed: %#v", queue.items)
	}

	if err := storePlace(ctx, queries, "approved", googlePlace("p2", "New approval", "Kraków"), 1, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := storePlace(ctx, queries, "invalidated", googlePlace("fresh", "Fresh result", "Kraków"), 1, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if suppressed := queue.ApplyDecisions(ctx, queries, now.Add(2*time.Hour)); !suppressed["rejected"] {
		t.Fatal("rejected building was not suppressed on the second import")
	}
	records, err = queries.GetBuildingPlaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	byKey = map[string]db.BuildingPlace{}
	for _, record := range records {
		byKey[record.BuildingKey] = record
	}
	if byKey["approved"].PlaceID == nil || *byKey["approved"].PlaceID != "p2" {
		t.Fatalf("consumed approval was reapplied: %#v", byKey["approved"])
	}
	if byKey["invalidated"].PlaceID == nil || *byKey["invalidated"].PlaceID != "fresh" {
		t.Fatalf("consumed invalidation cleared a fresh place: %#v", byKey["invalidated"])
	}

	reloaded, err := LoadBuildingReviewQueue(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.items[0].Status != ReviewStatusResolved || reloaded.items[2].Status != ReviewStatusPending {
		t.Fatalf("consumed statuses were not persisted: %#v", reloaded.items)
	}
}

func TestLoadBuildingReviewQueueReturnsUsableQueueOnParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reviews.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	queue, err := LoadBuildingReviewQueue(path, "")
	if err == nil || queue == nil || queue.httpClient == nil {
		t.Fatalf("LoadBuildingReviewQueue() = (%#v, %v), want initialized queue and parse error", queue, err)
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
