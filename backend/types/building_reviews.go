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

type BuildingReviewKind string

const (
	ReviewKindClassroomAmbiguous BuildingReviewKind = "classroom_ambiguous"
	ReviewKindClassroomUnmatched BuildingReviewKind = "classroom_unmatched"
	ReviewKindPlaceResolution    BuildingReviewKind = "place_resolution"
)

type BuildingReviewStatus string

const (
	ReviewStatusPending     BuildingReviewStatus = "pending"
	ReviewStatusResolved    BuildingReviewStatus = "resolved"
	ReviewStatusApproved    BuildingReviewStatus = "approved"
	ReviewStatusRejected    BuildingReviewStatus = "rejected"
	ReviewStatusInvalidated BuildingReviewStatus = "invalidated"
)

type BuildingReviewItem struct {
	ID                string                    `json:"id"`
	Kind              BuildingReviewKind        `json:"kind"`
	OriginalText      string                    `json:"original_text,omitempty"`
	NormalizedParts   []string                  `json:"normalized_parts,omitempty"`
	ClassroomKey      string                    `json:"classroom_key,omitempty"`
	BuildingKey       string                    `json:"building_key,omitempty"`
	Confidence        float64                   `json:"confidence"`
	Candidates        []BuildingReviewCandidate `json:"candidates,omitempty"`
	ValidationReasons []string                  `json:"validation_reasons,omitempty"`
	Status            BuildingReviewStatus      `json:"status"`
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
		return queue, fmt.Errorf("read building review queue: %w", err)
	}
	if len(contents) == 0 {
		return queue, nil
	}
	var items []BuildingReviewItem
	if err := json.Unmarshal(contents, &items); err != nil {
		return queue, fmt.Errorf("parse building review queue: %w", err)
	}
	queue.items = items
	return queue, nil
}

func (q *BuildingReviewQueue) RecordClassroomMatch(match classroommatch.Result) {
	if q == nil {
		return
	}
	var kind BuildingReviewKind
	switch match.Kind {
	case classroommatch.MatchAmbiguous:
		kind = ReviewKindClassroomAmbiguous
	case classroommatch.MatchUnmatched:
		kind = ReviewKindClassroomUnmatched
	default:
		return
	}
	q.record(BuildingReviewItem{Kind: kind, OriginalText: match.Input, NormalizedParts: []string{match.Normalized}, ClassroomKey: match.ClassroomKey, BuildingKey: match.BuildingKey, Confidence: match.Score, Status: ReviewStatusPending})
}

func (q *BuildingReviewQueue) RecordPlaceResolution(buildingKey string, confidence float64, candidates []scoredPlace, reasons []string) {
	if q == nil {
		return
	}
	reviewCandidates := make([]BuildingReviewCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		reviewCandidates = append(reviewCandidates, BuildingReviewCandidate{PlaceID: candidate.Place.GetId(), Name: placeDisplayName(candidate.Place), Address: candidate.Place.GetFormattedAddress(), Score: candidate.Score})
	}
	q.record(BuildingReviewItem{Kind: ReviewKindPlaceResolution, BuildingKey: buildingKey, Confidence: confidence, Candidates: reviewCandidates, ValidationReasons: reasons, Status: ReviewStatusPending})
}

func (q *BuildingReviewQueue) MarkPlaceResolved(buildingKey string) {
	if q == nil {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	changed := false
	for i := range q.items {
		if q.items[i].Kind == ReviewKindPlaceResolution && q.items[i].BuildingKey == buildingKey && q.items[i].Status == ReviewStatusPending {
			q.items[i].Status = ReviewStatusResolved
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
		q.mu.Unlock()
		return
	}
	item.FirstSeen, item.LastSeen, item.Occurrences = now, now, 1
	q.items = append(q.items, item)
	err := q.saveLocked()
	q.mu.Unlock()
	if err != nil {
		slog.Error("building_place_review.save_failed", slog.Any("error", err))
		return
	}
	q.notify(item)
}

func reviewID(kind BuildingReviewKind, text string, normalized []string, buildingKey string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{string(kind), text, strings.Join(normalized, "\x00"), buildingKey}, "\x00")))
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
	client := q.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
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
	changed := false
	for i := range q.items {
		item := &q.items[i]
		if item.BuildingKey == "" {
			continue
		}
		switch item.Status {
		case ReviewStatusRejected:
			suppressed[item.BuildingKey] = true
		case ReviewStatusInvalidated:
			if err := queries.ClearBuildingPlace(ctx, item.BuildingKey); err != nil {
				slog.Warn("building_place_review.clear_failed", slog.String("building_key", item.BuildingKey), slog.Any("error", err))
				continue
			}
			item.Status = ReviewStatusPending
			item.LastSeen = now
			changed = true
		case ReviewStatusApproved:
			candidate, ok := reviewCandidate(*item, item.ApprovedPlaceID)
			if !ok {
				slog.Warn("building_place_review.invalid_approval", slog.String("id", item.ID), slog.String("building_key", item.BuildingKey))
				continue
			}
			if err := storePlace(ctx, queries, item.BuildingKey, candidate, 1, now); err != nil {
				slog.Warn("building_place_review.approval_failed", slog.String("building_key", item.BuildingKey), slog.Any("error", err))
				continue
			}
			item.Status = ReviewStatusResolved
			item.LastSeen = now
			changed = true
		}
	}
	if changed {
		if err := q.saveLocked(); err != nil {
			slog.Error("building_place_review.save_failed", slog.Any("error", err))
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
