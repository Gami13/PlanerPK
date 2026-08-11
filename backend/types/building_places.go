package types

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	places "cloud.google.com/go/maps/places/apiv1"
	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"github.com/woodpkrs/PlanerPK/classroommatch"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/global"
	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

const (
	buildingPlaceExpiry   = 90 * 24 * time.Hour
	minimumPlaceScore     = 0.70
	minimumPlaceScoreLead = 0.15
)

// PlaceCandidate contains only the Google Place fields needed to choose and
// render a map link. Keeping this separate from the generated client makes
// resolver behavior deterministic and easy to test.
type PlaceCandidate struct {
	PlaceID string `json:"place_id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// PlacesClient is the small Google Places surface used by building resolution.
type PlacesClient interface {
	SearchText(ctx context.Context, query string) ([]PlaceCandidate, error)
	GetPlace(ctx context.Context, placeID string) (PlaceCandidate, error)
	Close() error
}

type googlePlacesClient struct{ client *places.Client }

func newGooglePlacesClient(ctx context.Context) (PlacesClient, error) {
	if global.GooglePlacesAPIKey() == "" || global.GooglePlacesAPIKey() == "PLEASE_SET_YOUR_GOOGLE_PLACES_API_KEY" {
		return nil, nil
	}
	client, err := places.NewClient(ctx, option.WithAPIKey(global.GooglePlacesAPIKey()))
	if err != nil {
		return nil, fmt.Errorf("create Google Places client: %w", err)
	}
	return &googlePlacesClient{client: client}, nil
}

func (c *googlePlacesClient) SearchText(ctx context.Context, query string) ([]PlaceCandidate, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-goog-fieldmask", "places.id,places.displayName,places.formattedAddress")
	response, err := c.client.SearchText(ctx, &placespb.SearchTextRequest{
		TextQuery:      query,
		LanguageCode:   "pl",
		RegionCode:     "PL",
		MaxResultCount: 5,
	})
	if err != nil {
		return nil, err
	}
	result := make([]PlaceCandidate, 0, len(response.GetPlaces()))
	for _, place := range response.GetPlaces() {
		name := ""
		if place.GetDisplayName() != nil {
			name = place.GetDisplayName().GetText()
		}
		result = append(result, PlaceCandidate{PlaceID: place.GetId(), Name: name, Address: place.GetFormattedAddress()})
	}
	return result, nil
}

func (c *googlePlacesClient) GetPlace(ctx context.Context, placeID string) (PlaceCandidate, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-goog-fieldmask", "id,displayName,formattedAddress")
	place, err := c.client.GetPlace(ctx, &placespb.GetPlaceRequest{Name: "places/" + placeID, LanguageCode: "pl"})
	if err != nil {
		return PlaceCandidate{}, err
	}
	name := ""
	if place.GetDisplayName() != nil {
		name = place.GetDisplayName().GetText()
	}
	return PlaceCandidate{PlaceID: place.GetId(), Name: name, Address: place.GetFormattedAddress()}, nil
}

func (c *googlePlacesClient) Close() error { return c.client.Close() }

type BuildingReviewCandidate struct {
	PlaceID string  `json:"place_id"`
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Score   float64 `json:"score"`
}

type BuildingReviewItem struct {
	ID                string                    `json:"id"`
	Kind              string                    `json:"kind"`
	OriginalText      string                    `json:"original_text,omitempty"`
	NormalizedParts   []string                  `json:"normalized_parts,omitempty"`
	ClassroomKey      string                    `json:"classroom_key,omitempty"`
	BuildingKey       string                    `json:"building_key,omitempty"`
	Confidence        float64                   `json:"confidence"`
	Candidates        []BuildingReviewCandidate `json:"candidates,omitempty"`
	ValidationReasons []string                  `json:"validation_reasons,omitempty"`
	Status            string                    `json:"status"`
	ApprovedPlaceID   string                    `json:"approved_place_id,omitempty"`
	FirstSeen         time.Time                 `json:"first_seen"`
	LastSeen          time.Time                 `json:"last_seen"`
	Occurrences       int                       `json:"occurrences"`
}

// BuildingReviewQueue is the hand-editable durable review queue. Status is
// pending, resolved, approved, rejected, or invalidated.
type BuildingReviewQueue struct {
	path       string
	webhookURL string
	httpClient *http.Client
	mu         sync.Mutex
	items      []BuildingReviewItem
}

func LoadBuildingReviewQueue(path, webhookURL string) (*BuildingReviewQueue, error) {
	queue := &BuildingReviewQueue{path: path, webhookURL: webhookURL, httpClient: &http.Client{Timeout: 5 * time.Second}}
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return queue, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read building review queue: %w", err)
	}
	if len(contents) == 0 {
		return queue, nil
	}
	if err := json.Unmarshal(contents, &queue.items); err != nil {
		return nil, fmt.Errorf("parse building review queue: %w", err)
	}
	return queue, nil
}

func (q *BuildingReviewQueue) RecordClassroomMatch(match classroommatch.Result) {
	if q == nil {
		return
	}
	q.record(BuildingReviewItem{
		Kind:            "classroom_" + string(match.Kind),
		OriginalText:    match.Input,
		NormalizedParts: []string{match.Normalized},
		ClassroomKey:    match.ClassroomKey,
		BuildingKey:     match.BuildingKey,
		Confidence:      match.Score,
		Status:          "pending",
	})
}

func (q *BuildingReviewQueue) RecordPlaceResolution(buildingKey string, confidence float64, candidates []scoredPlace, reasons []string) {
	if q == nil {
		return
	}
	reviewCandidates := make([]BuildingReviewCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		reviewCandidates = append(reviewCandidates, BuildingReviewCandidate{PlaceID: candidate.PlaceID, Name: candidate.Name, Address: candidate.Address, Score: candidate.Score})
	}
	q.record(BuildingReviewItem{Kind: "place_resolution", BuildingKey: buildingKey, Confidence: confidence, Candidates: reviewCandidates, ValidationReasons: reasons, Status: "pending"})
}

func (q *BuildingReviewQueue) MarkPlaceResolved(buildingKey string) {
	if q == nil {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	changed := false
	for i := range q.items {
		if q.items[i].Kind == "place_resolution" && q.items[i].BuildingKey == buildingKey && q.items[i].Status == "pending" {
			q.items[i].Status = "resolved"
			q.items[i].LastSeen = time.Now().UTC()
			changed = true
		}
	}
	if changed {
		if err := q.saveLocked(); err != nil {
			slog.Error("building_place_review.save_failed", slog.Any("error", err))
		}
	}
}

func (q *BuildingReviewQueue) record(item BuildingReviewItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	item.ID = reviewID(item.Kind, item.OriginalText, item.NormalizedParts, item.BuildingKey)
	now := time.Now().UTC()
	for i := range q.items {
		if q.items[i].ID != item.ID {
			continue
		}
		q.items[i].LastSeen = now
		q.items[i].Occurrences++
		if err := q.saveLocked(); err != nil {
			slog.Error("building_place_review.save_failed", slog.Any("error", err))
		}
		return
	}
	item.FirstSeen, item.LastSeen, item.Occurrences = now, now, 1
	q.items = append(q.items, item)
	if err := q.saveLocked(); err != nil {
		slog.Error("building_place_review.save_failed", slog.Any("error", err))
		return
	}
	q.notify(item)
}

func reviewID(kind, text string, normalized []string, buildingKey string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{kind, text, strings.Join(normalized, "\x00"), buildingKey}, "\x00")))
	return hex.EncodeToString(sum[:])
}

func (q *BuildingReviewQueue) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(q.path), 0o755); err != nil {
		return err
	}
	contents, err := json.MarshalIndent(q.items, "", "  ")
	if err != nil {
		return err
	}
	temporary := q.path + ".tmp"
	if err := os.WriteFile(temporary, contents, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, q.path)
}

func (q *BuildingReviewQueue) notify(item BuildingReviewItem) {
	if q.webhookURL == "" {
		return
	}
	body, err := json.Marshal(item)
	if err != nil {
		return
	}
	request, err := http.NewRequest(http.MethodPost, q.webhookURL, bytes.NewReader(body))
	if err != nil {
		slog.Warn("building_place_review.webhook_request_failed", slog.Any("error", err))
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := q.httpClient.Do(request)
	if err != nil {
		slog.Warn("building_place_review.webhook_failed", slog.Any("error", err))
		return
	}
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		slog.Warn("building_place_review.webhook_failed", slog.Int("status_code", response.StatusCode))
	}
}

func (q *BuildingReviewQueue) ApplyDecisions(ctx context.Context, queries db.Querier, now time.Time) map[string]bool {
	suppressed := make(map[string]bool)
	if q == nil {
		return suppressed
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, item := range q.items {
		if item.BuildingKey == "" {
			continue
		}
		switch item.Status {
		case "rejected":
			suppressed[item.BuildingKey] = true
		case "invalidated", "pending":
			if err := queries.ClearBuildingPlace(ctx, item.BuildingKey); err != nil {
				slog.Warn("building_place_review.clear_failed", slog.String("building_key", item.BuildingKey), slog.Any("error", err))
			}
		case "approved":
			candidate, ok := reviewCandidate(item, item.ApprovedPlaceID)
			if !ok {
				slog.Warn("building_place_review.invalid_approval", slog.String("id", item.ID), slog.String("building_key", item.BuildingKey))
				continue
			}
			if err := storePlace(ctx, queries, item.BuildingKey, candidate.PlaceCandidate, 1, now); err != nil {
				slog.Warn("building_place_review.approval_failed", slog.String("building_key", item.BuildingKey), slog.Any("error", err))
			}
		}
	}
	return suppressed
}

func reviewCandidate(item BuildingReviewItem, placeID string) (scoredPlace, bool) {
	for _, candidate := range item.Candidates {
		if candidate.PlaceID == placeID && placeID != "" {
			return scoredPlace{PlaceCandidate: PlaceCandidate{PlaceID: candidate.PlaceID, Name: candidate.Name, Address: candidate.Address}, Score: candidate.Score}, true
		}
	}
	return scoredPlace{}, false
}

type scoredPlace struct {
	PlaceCandidate
	Score   float64
	Reasons []string
}

// ResolveBuildingPlaces synchronizes the durable classroom/building cache,
// applies hand-edited review decisions, resolves missing places, refreshes
// stale details, and hydrates buildings for calendar event creation.
func ResolveBuildingPlaces(ctx context.Context, queries db.Querier, buildings *Buildings, client PlacesClient, reviews *BuildingReviewQueue, now time.Time) error {
	if buildings == nil {
		return nil
	}
	if err := syncLegendBuildingCache(ctx, queries, buildings); err != nil {
		return err
	}
	suppressed := reviews.ApplyDecisions(ctx, queries, now)
	records, err := queries.GetBuildingPlaces(ctx)
	if err != nil {
		return fmt.Errorf("get building places: %w", err)
	}
	byKey := make(map[string]db.BuildingPlace, len(records))
	for _, record := range records {
		byKey[record.BuildingKey] = record
	}
	keys := make([]string, 0, len(buildings.Buildings))
	for key := range buildings.Buildings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		record := byKey[key]
		if suppressed[key] {
			continue
		}
		if record.PlaceID == nil || *record.PlaceID == "" {
			resolveMissingPlace(ctx, queries, client, reviews, record, now)
			continue
		}
		if record.ExpiresAt != nil && !record.ExpiresAt.After(now) {
			refreshPlace(ctx, queries, client, reviews, record, now)
		}
	}
	records, err = queries.GetBuildingPlaces(ctx)
	if err != nil {
		return fmt.Errorf("reload building places: %w", err)
	}
	for _, record := range records {
		building, exists := buildings.Buildings[record.BuildingKey]
		if !exists || record.PlaceID == nil || *record.PlaceID == "" {
			continue
		}
		building.GooglePlaceID = *record.PlaceID
		if record.DisplayName != nil {
			building.Name = *record.DisplayName
		}
		if record.Address != nil {
			building.Address = *record.Address
		}
		buildings.Buildings[record.BuildingKey] = building
	}
	buildings.reviews = reviews
	return nil
}

// ResolveBuildingPlacesForImport is the production entry point. Google and
// review-file failures are intentionally non-fatal so a timetable can still
// be imported without map links.
func ResolveBuildingPlacesForImport(ctx context.Context, queries db.Querier, buildings *Buildings) error {
	reviews, err := LoadBuildingReviewQueue(global.BuildingReviewPath(), global.BuildingReviewWebhookURL())
	if err != nil {
		slog.Error("building_place_review.load_failed", slog.Any("error", err))
		reviews = &BuildingReviewQueue{path: global.BuildingReviewPath(), webhookURL: global.BuildingReviewWebhookURL(), httpClient: &http.Client{Timeout: 5 * time.Second}}
	}
	client, err := newGooglePlacesClient(ctx)
	if err != nil {
		slog.Error("building_place_resolution.client_failed", slog.Any("error", err))
		client = nil
	}
	if client != nil {
		defer func() {
			if err := client.Close(); err != nil {
				slog.Warn("building_place_resolution.client_close_failed", slog.Any("error", err))
			}
		}()
	}
	return ResolveBuildingPlaces(ctx, queries, buildings, client, reviews, time.Now().UTC())
}

func syncLegendBuildingCache(ctx context.Context, queries db.Querier, buildings *Buildings) error {
	keys := make([]string, 0, len(buildings.Buildings))
	for key := range buildings.Buildings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := queries.UpsertBuildingPlace(ctx, db.UpsertBuildingPlaceParams{BuildingKey: key, CanonicalName: buildings.Buildings[key].CanonicalName}); err != nil {
			return fmt.Errorf("upsert building place %q: %w", key, err)
		}
	}
	if err := queries.ClearClassroomBuildings(ctx); err != nil {
		return fmt.Errorf("clear classroom buildings: %w", err)
	}
	classrooms := make([]string, 0, len(buildings.Classrooms))
	for classroom := range buildings.Classrooms {
		classrooms = append(classrooms, classroom)
	}
	sort.Strings(classrooms)
	for _, classroom := range classrooms {
		if err := queries.UpsertClassroomBuilding(ctx, db.UpsertClassroomBuildingParams{ClassroomKey: classroom, BuildingKey: buildings.Classrooms[classroom]}); err != nil {
			return fmt.Errorf("upsert classroom building %q: %w", classroom, err)
		}
	}
	return nil
}

func resolveMissingPlace(ctx context.Context, queries db.Querier, client PlacesClient, reviews *BuildingReviewQueue, building db.BuildingPlace, now time.Time) {
	if client == nil {
		slog.Debug("building_place_resolution.api_not_configured", slog.String("building_key", building.BuildingKey))
		return
	}
	candidates, err := client.SearchText(ctx, building.CanonicalName+" Politechnika Krakowska")
	if err != nil {
		slog.Error("building_place_resolution.search_failed", slog.String("building_key", building.BuildingKey), slog.Any("error", err))
		return
	}
	best, ranked, reasons, ok := choosePlace(building.CanonicalName, candidates)
	if !ok {
		reviews.RecordPlaceResolution(building.BuildingKey, best.Score, ranked, reasons)
		return
	}
	if err := storePlace(ctx, queries, building.BuildingKey, best.PlaceCandidate, best.Score, now); err != nil {
		slog.Error("building_place_resolution.store_failed", slog.String("building_key", building.BuildingKey), slog.Any("error", err))
		return
	}
	reviews.MarkPlaceResolved(building.BuildingKey)
}

func refreshPlace(ctx context.Context, queries db.Querier, client PlacesClient, reviews *BuildingReviewQueue, building db.BuildingPlace, now time.Time) {
	if client == nil || building.PlaceID == nil {
		return
	}
	candidate, err := client.GetPlace(ctx, *building.PlaceID)
	if err != nil {
		slog.Error("building_place_resolution.refresh_failed", slog.String("building_key", building.BuildingKey), slog.Any("error", err))
		return
	}
	if candidate.PlaceID == "" {
		candidate.PlaceID = *building.PlaceID
	}
	if err := storePlace(ctx, queries, building.BuildingKey, candidate, building.Confidence, now); err != nil {
		slog.Error("building_place_resolution.refresh_store_failed", slog.String("building_key", building.BuildingKey), slog.Any("error", err))
	}
}

func storePlace(ctx context.Context, queries db.Querier, buildingKey string, candidate PlaceCandidate, confidence float64, now time.Time) error {
	return queries.UpdateBuildingPlace(ctx, db.UpdateBuildingPlaceParams{
		PlaceID: &candidate.PlaceID, DisplayName: &candidate.Name, Address: &candidate.Address,
		Confidence: confidence, VerifiedAt: &now, ExpiresAt: timePointer(now.Add(buildingPlaceExpiry)), BuildingKey: buildingKey,
	})
}

func timePointer(value time.Time) *time.Time { return &value }

func choosePlace(canonicalName string, candidates []PlaceCandidate) (scoredPlace, []scoredPlace, []string, bool) {
	ranked := make([]scoredPlace, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.PlaceID == "" {
			continue
		}
		score, reasons := scorePlace(canonicalName, candidate)
		ranked = append(ranked, scoredPlace{PlaceCandidate: candidate, Score: score, Reasons: reasons})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].PlaceID < ranked[j].PlaceID
		}
		return ranked[i].Score > ranked[j].Score
	})
	if len(ranked) == 0 {
		return scoredPlace{}, ranked, []string{"Google Text Search returned no candidates"}, false
	}
	best := ranked[0]
	if best.Score < minimumPlaceScore {
		return best, ranked, append(best.Reasons, "candidate confidence is below 0.70"), false
	}
	if len(ranked) > 1 && best.Score-ranked[1].Score < minimumPlaceScoreLead {
		return best, ranked, append(best.Reasons, "best candidate does not lead runner-up by 0.15"), false
	}
	return best, ranked, best.Reasons, true
}

func scorePlace(canonicalName string, candidate PlaceCandidate) (float64, []string) {
	text := foldPlaceText(candidate.Name + " " + candidate.Address)
	var score float64
	reasons := make([]string, 0, 3)
	if hasBuildingNameEvidence(canonicalName, text) {
		score += 0.70
		reasons = append(reasons, "canonical building-name evidence")
	} else {
		reasons = append(reasons, "no canonical building-name evidence")
	}
	if strings.Contains(text, "politechnika") || strings.Contains(text, " pk ") {
		score += 0.15
		reasons = append(reasons, "Politechnika Krakowska context")
	}
	if strings.Contains(text, "krakow") {
		score += 0.15
		reasons = append(reasons, "Krakow context")
	}
	return score, reasons
}

func hasBuildingNameEvidence(canonicalName, candidateText string) bool {
	ignored := map[string]struct{}{"budynek": {}, "budynku": {}, "building": {}, "ul": {}, "ulica": {}, "warszawska": {}, "krakow": {}, "politechnika": {}, "krakowska": {}, "pk": {}}
	for _, token := range strings.Fields(foldPlaceText(canonicalName)) {
		if _, skip := ignored[token]; skip {
			continue
		}
		if len([]rune(token)) >= 3 {
			if strings.Contains(candidateText, " "+token+" ") {
				return true
			}
		}
	}
	return false
}

func foldPlaceText(value string) string {
	var result strings.Builder
	for _, r := range strings.ToLower(value) {
		switch r {
		case 'ą':
			r = 'a'
		case 'ć':
			r = 'c'
		case 'ę':
			r = 'e'
		case 'ł':
			r = 'l'
		case 'ń':
			r = 'n'
		case 'ó':
			r = 'o'
		case 'ś':
			r = 's'
		case 'ź', 'ż':
			r = 'z'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	return " " + strings.Join(strings.Fields(result.String()), " ") + " "
}
