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
	"strings"
	"sync"
	"time"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"github.com/woodpkrs/PlanerPK/classroommatch"
	"github.com/woodpkrs/PlanerPK/gen/db"
	localizedtext "google.golang.org/genproto/googleapis/type/localized_text"
)

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
	q.record(BuildingReviewItem{Kind: "classroom_" + string(match.Kind), OriginalText: match.Input, NormalizedParts: []string{match.Normalized}, ClassroomKey: match.ClassroomKey, BuildingKey: match.BuildingKey, Confidence: match.Score, Status: "pending"})
}

func (q *BuildingReviewQueue) RecordPlaceResolution(buildingKey string, confidence float64, candidates []scoredPlace, reasons []string) {
	if q == nil {
		return
	}
	reviewCandidates := make([]BuildingReviewCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		reviewCandidates = append(reviewCandidates, BuildingReviewCandidate{PlaceID: candidate.Place.GetId(), Name: placeDisplayName(candidate.Place), Address: candidate.Place.GetFormattedAddress(), Score: candidate.Score})
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
			if err := storePlace(ctx, queries, item.BuildingKey, candidate, 1, now); err != nil {
				slog.Warn("building_place_review.approval_failed", slog.String("building_key", item.BuildingKey), slog.Any("error", err))
			}
		}
	}
	return suppressed
}

func reviewCandidate(item BuildingReviewItem, placeID string) (*placespb.Place, bool) {
	for _, candidate := range item.Candidates {
		if candidate.PlaceID == placeID && placeID != "" {
			return &placespb.Place{Id: candidate.PlaceID, DisplayName: &localizedtext.LocalizedText{Text: candidate.Name}, FormattedAddress: candidate.Address}, true
		}
	}
	return nil, false
}
