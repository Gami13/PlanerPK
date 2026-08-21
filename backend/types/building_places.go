package types

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	places "cloud.google.com/go/maps/places/apiv1"
	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/global"
	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

const (
	buildingPlaceExpiry            = 90 * 24 * time.Hour
	buildingPlaceSearchResultCount = 20
)

// BuildingFromPlace converts a durable building-place record into the runtime
// representation used for map presentation.
func BuildingFromPlace(place db.BuildingPlace) Building {
	building := Building{CanonicalName: place.CanonicalName}
	if place.PlaceID != nil {
		building.GooglePlaceID = *place.PlaceID
	}
	if place.DisplayName != nil {
		building.Name = *place.DisplayName
	}
	if place.Address != nil {
		building.Address = *place.Address
	}
	return building
}

// ResolveBuildingPlaces synchronizes the durable classroom/building cache,
// applies hand-edited review decisions, resolves missing places, refreshes
// stale details, and hydrates buildings for calendar event creation.
func ResolveBuildingPlaces(ctx context.Context, queries db.Querier, buildings *Buildings, client *places.Client, reviews *BuildingReviewQueue, now time.Time) error {
	if buildings == nil {
		return nil
	}
	if err := syncLegendBuildingCache(ctx, queries, buildings); err != nil {
		return err
	}

	suppressed := reviews.ApplyDecisions(ctx, queries, now)
	records, err := loadBuildingPlaceRecords(ctx, queries)
	if err != nil {
		return err
	}
	resolveBuildingPlaceCandidates(ctx, queries, buildings, client, reviews, records, suppressed, now)

	records, err = loadBuildingPlaceRecords(ctx, queries)
	if err != nil {
		return err
	}
	hydrateBuildingDetails(buildings, records)
	buildings.reviews = reviews
	return nil
}

func loadBuildingPlaceRecords(ctx context.Context, queries db.Querier) ([]db.BuildingPlace, error) {
	records, err := queries.GetBuildingPlaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("get building places: %w", err)
	}
	return records, nil
}

func resolveBuildingPlaceCandidates(ctx context.Context, queries db.Querier, buildings *Buildings, client *places.Client, reviews *BuildingReviewQueue, records []db.BuildingPlace, suppressed map[string]bool, now time.Time) {
	byKey := make(map[string]db.BuildingPlace, len(records))
	for _, record := range records {
		byKey[record.BuildingKey] = record
	}
	keys := slices.Sorted(maps.Keys(buildings.BuildingsByKey))
	for _, key := range keys {
		if suppressed[key] {
			continue
		}
		record := byKey[key]
		if record.PlaceID == nil || *record.PlaceID == "" {
			resolveMissingPlace(ctx, queries, client, reviews, record, now)
			continue
		}
		if record.ExpiresAt != nil && !record.ExpiresAt.After(now) {
			refreshPlace(ctx, queries, client, record, now)
		}
	}
}

func hydrateBuildingDetails(buildings *Buildings, records []db.BuildingPlace) {
	for _, record := range records {
		if _, exists := buildings.BuildingsByKey[record.BuildingKey]; !exists || record.PlaceID == nil || *record.PlaceID == "" {
			continue
		}
		buildings.BuildingsByKey[record.BuildingKey] = BuildingFromPlace(record)
	}
}

// ResolveBuildingPlacesForImport is the production entry point. Google and
// review-file failures are intentionally non-fatal so a timetable can still
// be imported without map links.
func ResolveBuildingPlacesForImport(ctx context.Context, queries db.Querier, buildings *Buildings) error {
	reviews, err := LoadBuildingReviewQueue(global.BuildingReviewPath(), global.BuildingReviewWebhookURL())
	if err != nil {
		slog.Error("building_place_review.load_failed", slog.Any("error", err))
	}
	var client *places.Client
	if global.GooglePlacesConfigured() {
		client, err = places.NewClient(ctx, option.WithAPIKey(global.GooglePlacesAPIKey()))
		if err != nil {
			slog.Error("building_place_resolution.client_failed", slog.Any("error", err))
			client = nil
		}
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
	keys := slices.Sorted(maps.Keys(buildings.BuildingsByKey))
	for _, key := range keys {
		if err := queries.UpsertBuildingPlace(ctx, db.UpsertBuildingPlaceParams{
			BuildingKey:   key,
			CanonicalName: buildings.BuildingsByKey[key].CanonicalName,
		}); err != nil {
			return fmt.Errorf("upsert building place %q: %w", key, err)
		}
	}
	return nil
}

func resolveMissingPlace(ctx context.Context, queries db.Querier, client *places.Client, reviews *BuildingReviewQueue, building db.BuildingPlace, now time.Time) {
	if client == nil {
		slog.Debug("building_place_resolution.api_not_configured", slog.String("building_key", building.BuildingKey))
		return
	}
	searchContext := metadata.AppendToOutgoingContext(ctx, "x-goog-fieldmask", "places.id,places.displayName,places.formattedAddress")
	response, err := client.SearchText(searchContext, buildingPlaceSearchRequest(building.CanonicalName))
	if err != nil {
		slog.Error("building_place_resolution.search_failed", slog.String("building_key", building.BuildingKey), slog.Any("error", err))
		return
	}
	resolution := choosePlace(building.CanonicalName, response.GetPlaces())
	if !resolution.Accepted {
		reviews.RecordPlaceResolution(building.BuildingKey, resolution.Best.Score, resolution.Ranked, resolution.Reasons)
		return
	}
	if err := storePlace(ctx, queries, building.BuildingKey, resolution.Best.Place, resolution.Best.Score, now); err != nil {
		slog.Error("building_place_resolution.store_failed", slog.String("building_key", building.BuildingKey), slog.Any("error", err))
		return
	}
	reviews.MarkPlaceResolved(building.BuildingKey)
}

func buildingPlaceSearchRequest(canonicalName string) *placespb.SearchTextRequest {
	return &placespb.SearchTextRequest{
		TextQuery:      canonicalName + " Politechnika Krakowska",
		LanguageCode:   "pl",
		RegionCode:     "PL",
		MaxResultCount: buildingPlaceSearchResultCount,
	}
}

func refreshPlace(ctx context.Context, queries db.Querier, client *places.Client, building db.BuildingPlace, now time.Time) {
	if client == nil || building.PlaceID == nil {
		return
	}
	refreshContext := metadata.AppendToOutgoingContext(ctx, "x-goog-fieldmask", "id,displayName,formattedAddress")
	candidate, err := client.GetPlace(refreshContext, &placespb.GetPlaceRequest{Name: "places/" + *building.PlaceID, LanguageCode: "pl"})
	if err != nil {
		slog.Error("building_place_resolution.refresh_failed", slog.String("building_key", building.BuildingKey), slog.Any("error", err))
		return
	}
	if candidate.GetId() == "" {
		candidate.Id = *building.PlaceID
	}
	if err := storePlace(ctx, queries, building.BuildingKey, candidate, building.Confidence, now); err != nil {
		slog.Error("building_place_resolution.refresh_store_failed", slog.String("building_key", building.BuildingKey), slog.Any("error", err))
	}
}

func storePlace(ctx context.Context, queries db.Querier, buildingKey string, place *placespb.Place, confidence float64, now time.Time) error {
	if place == nil || place.GetId() == "" {
		return errors.New("Google Place has no ID")
	}
	placeID := place.GetId()
	displayName := placeDisplayName(place)
	address := place.GetFormattedAddress()
	expiresAt := now.Add(buildingPlaceExpiry)
	return queries.UpdateBuildingPlace(ctx, db.UpdateBuildingPlaceParams{
		BuildingKey: buildingKey,
		PlaceID:     &placeID,
		DisplayName: &displayName,
		Address:     &address,
		Confidence:  confidence,
		VerifiedAt:  &now,
		ExpiresAt:   &expiresAt,
	})
}
