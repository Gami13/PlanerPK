package types

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"unicode"

	places "cloud.google.com/go/maps/places/apiv1"
	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"github.com/woodpkrs/PlanerPK/classroommatch"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/global"
	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

type Building struct {
	Key           string
	CanonicalName string
	GooglePlaceID string
	Name          string
	Address       string
}

type Buildings struct {
	Buildings  map[string]Building
	Classrooms map[string]string
	matcher    *classroommatch.Matcher
}

// BuildBuildingsFromLegend creates the in-memory classroom-to-building cache
// for one timetable import. It intentionally does not read legacy database
// mappings.
func BuildBuildingsFromLegend(legend *Legend) *Buildings {
	result := &Buildings{
		Buildings:  make(map[string]Building),
		Classrooms: make(map[string]string),
	}
	if legend == nil {
		return result
	}

	locationKeys := make([]string, 0, len(legend.Locations))
	for key := range legend.Locations {
		locationKeys = append(locationKeys, key)
	}
	sort.Strings(locationKeys)

	candidates := make([]classroommatch.Candidate, 0, len(locationKeys))
	buildingAliases := make(map[string]struct{})
	for _, locationKey := range locationKeys {
		location := legend.Locations[locationKey]
		classroomKey := strings.TrimSpace(location.Shorthand)
		if classroomKey == "" {
			classroomKey = locationKey
		}
		buildingKey := strings.TrimSpace(location.FullName)
		if classroomKey == "" || buildingKey == "" {
			slog.Warn("classroom_resolution.invalid_legend_entry",
				slog.String("classroom", classroomKey),
				slog.String("building_key", buildingKey),
				slog.String("source", "legend"))
			continue
		}
		if classroommatch.Normalize(classroomKey) == "zdalnie" {
			continue
		}

		result.Buildings[buildingKey] = Building{
			Key:           buildingKey,
			CanonicalName: buildingKey,
		}
		candidates = append(candidates, classroommatch.Candidate{
			ClassroomKey: classroomKey,
			BuildingKey:  buildingKey,
		})
		for _, alias := range buildingAcronymAliases(buildingKey) {
			aliasKey := alias + "\x00" + buildingKey
			if _, exists := buildingAliases[aliasKey]; exists {
				continue
			}
			buildingAliases[aliasKey] = struct{}{}
			candidates = append(candidates, classroommatch.Candidate{
				ClassroomKey:  alias,
				BuildingKey:   buildingKey,
				BuildingAlias: true,
			})
		}
	}

	result.matcher = classroommatch.New(candidates)
	result.Classrooms = result.matcher.ClassroomBuildings()
	buildingKeys := make([]string, 0, len(result.Buildings))
	for buildingKey := range result.Buildings {
		buildingKeys = append(buildingKeys, buildingKey)
	}
	sort.Strings(buildingKeys)
	slog.Debug("classroom_resolution.index_data",
		slog.Any("classrooms", result.Classrooms),
		slog.Any("building_keys", buildingKeys),
		slog.String("source", "legend"))
	return result
}

// buildingAcronymAliases adds an explicit building shorthand only when the
// legend names it as an uppercase acronym, for example "budynek SJO".
func buildingAcronymAliases(buildingKey string) []string {
	fields := strings.Fields(buildingKey)
	if len(fields) < 2 {
		return nil
	}
	for i, field := range fields[:len(fields)-1] {
		if !strings.EqualFold(strings.Trim(field, "\"'(),."), "budynek") {
			continue
		}
		candidate := strings.Trim(fields[i+1], "\"'(),.")
		if isUppercaseAcronym(candidate) {
			return []string{candidate}
		}
	}
	return nil
}

func isUppercaseAcronym(value string) bool {
	runes := []rune(value)
	if len(runes) < 2 || len(runes) > 6 {
		return false
	}
	for _, r := range runes {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func (b *Buildings) GetBuilding(key string) (Building, bool) {
	if b == nil || b.Buildings == nil {
		return Building{}, false
	}
	building, exists := b.Buildings[key]
	return building, exists
}

// ResolveClassroomKeys performs the only fuzzy matching in the runtime flow:
// while a newly imported timetable is being built.
func (b *Buildings) ResolveClassroomKeys(location string) []string {
	if b == nil || b.matcher == nil {
		return nil
	}
	keys := make([]string, 0)
	seen := make(map[string]struct{})
	for _, match := range b.matcher.Resolve(location) {
		if match.Kind != classroommatch.MatchExact && match.Kind != classroommatch.MatchCorrected {
			continue
		}
		if _, exists := seen[match.ClassroomKey]; exists {
			continue
		}
		seen[match.ClassroomKey] = struct{}{}
		keys = append(keys, match.ClassroomKey)
	}
	return keys
}

// GetBuildingsByClassroomKeys performs an exact lookup of keys that were
// precomputed while parsing a timetable.
func (b *Buildings) GetBuildingsByClassroomKeys(classroomKeys []string) []Building {
	if b == nil || b.Classrooms == nil || b.Buildings == nil {
		return nil
	}
	buildings := make([]Building, 0, len(classroomKeys))
	seen := make(map[string]struct{})
	for _, classroomKey := range classroomKeys {
		buildingKey, exists := b.Classrooms[classroomKey]
		if !exists {
			continue
		}
		if _, exists := seen[buildingKey]; exists {
			continue
		}
		building, exists := b.Buildings[buildingKey]
		if !exists {
			continue
		}
		seen[buildingKey] = struct{}{}
		buildings = append(buildings, building)
	}
	return buildings
}

type BuildingDetails struct {
	Name    string
	Address string
}

func fetchBuildingDetails(ctx context.Context, client *places.Client, googlePlaceID string) (*BuildingDetails, error) {
	slog.Debug("Fetching building details from Google Places API", slog.String("google_place_id", googlePlaceID))
	reqCtx := metadata.AppendToOutgoingContext(ctx, "x-goog-fieldmask", "displayName,formattedAddress")

	req := &placespb.GetPlaceRequest{
		Name:         fmt.Sprintf("places/%s", googlePlaceID),
		LanguageCode: "pl",
	}

	resp, err := client.GetPlace(reqCtx, req)
	if err != nil {
		slog.Error("Google Places API call failed", slog.String("google_place_id", googlePlaceID), slog.Any("error", err))
		return nil, fmt.Errorf("Google Places API error for place ID %s: %w", googlePlaceID, err)
	}

	var name string
	if resp.GetDisplayName() != nil {
		name = resp.GetDisplayName().GetText()
	}

	details := &BuildingDetails{
		Name:    name,
		Address: resp.GetFormattedAddress(),
	}
	slog.Debug("Successfully fetched building details from Google Places API",
		slog.String("google_place_id", googlePlaceID),
		slog.String("name", details.Name),
		slog.String("address", details.Address))

	return details, nil
}

func UpdateBuildingsDetails(ctx context.Context, queries db.Querier) error {
	slog.Info("Starting update of building details")
	client, err := places.NewClient(ctx, option.WithAPIKey(global.GooglePlacesAPIKey()))
	if err != nil {
		slog.Error("Failed to create Google Places client", slog.Any("error", err))
		return fmt.Errorf("failed to create Google Places client: %w", err)
	}
	defer client.Close()

	buildings, err := queries.GetBuildings(ctx)
	if err != nil {
		slog.Error("Failed to fetch buildings for details update", slog.Any("error", err))
		return err
	}

	updatedCount := 0
	skippedCount := 0
	for _, building := range buildings {
		if building.GooglePlaceID == nil || *building.GooglePlaceID == "" {
			skippedCount++
			slog.Debug("Skipping building without Google Place ID", slog.Int64("building_id", building.ID))
			continue
		}

		slog.Info("Updating details for building",
			slog.Int64("building_id", building.ID),
			slog.String("google_place_id", *building.GooglePlaceID))

		buildingDetails, err := fetchBuildingDetails(ctx, client, *building.GooglePlaceID)
		if err != nil {
			slog.Error("Failed to fetch building details",
				slog.Int64("building_id", building.ID),
				slog.String("google_place_id", *building.GooglePlaceID),
				slog.Any("error", err))
			return err
		}

		if err := queries.UpdateBuilding(ctx, db.UpdateBuildingParams{
			ID:      building.ID,
			Name:    &buildingDetails.Name,
			Address: &buildingDetails.Address,
		}); err != nil {
			slog.Error("Failed to update building in database",
				slog.Int64("building_id", building.ID),
				slog.Any("error", err))
			return err
		}
		updatedCount++
		slog.Info("Successfully updated building details in database",
			slog.Int64("building_id", building.ID),
			slog.String("name", buildingDetails.Name),
			slog.String("address", buildingDetails.Address))
	}

	slog.Info("Finished updating building details",
		slog.Int("updated_count", updatedCount),
		slog.Int("skipped_count", skippedCount))

	return nil
}
