package types

import (
	"context"
	"fmt"
	"log/slog"

	places "cloud.google.com/go/maps/places/apiv1"
	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/global"
	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

type Building struct {
	ID            int64
	Key           string
	GooglePlaceID string
	Name          string
	Address       string
}

type Buildings struct {
	Buildings  map[int64]Building
	Classrooms map[string]int64
}

func (b *Buildings) GetBuilding(id int64) (Building, bool) {
	if b == nil || b.Buildings == nil {
		return Building{}, false
	}
	building, exists := b.Buildings[id]
	return building, exists
}

func (b *Buildings) GetBuildingByClassroom(classroomName string) (Building, bool) {
	if b == nil || b.Classrooms == nil || b.Buildings == nil {
		return Building{}, false
	}
	buildingID, roomExists := b.Classrooms[classroomName]
	if !roomExists {
		return Building{}, false
	}
	building, buildingExists := b.Buildings[buildingID]
	return building, buildingExists
}

func MapToBuilding(building db.Building) Building {
	b := Building{
		ID: building.ID,
	}

	if building.Key != nil {
		b.Key = *building.Key
	}
	if building.GooglePlaceID != nil {
		b.GooglePlaceID = *building.GooglePlaceID
	}
	if building.Name != nil {
		b.Name = *building.Name
	}
	if building.Address != nil {
		b.Address = *building.Address
	}

	return b
}

func LoadBuildings(ctx context.Context, queries db.Querier) (*Buildings, error) {
	slog.Info("Loading buildings and classrooms from database")
	bList, err := queries.GetBuildings(ctx)
	if err != nil {
		slog.Error("Failed to fetch buildings from database", slog.Any("error", err))
		return nil, err
	}
	cList, err := queries.GetClassroomNamesWithBuildingIDs(ctx)
	if err != nil {
		slog.Error("Failed to fetch classroom names with building IDs", slog.Any("error", err))
		return nil, err
	}

	res := &Buildings{
		Buildings:  make(map[int64]Building, len(bList)),
		Classrooms: make(map[string]int64, len(cList)),
	}

	for _, b := range bList {
		res.Buildings[b.ID] = MapToBuilding(b)
	}

	for _, c := range cList {
		if c.Name != nil {
			res.Classrooms[*c.Name] = c.ID
		}
	}

	slog.Info("Successfully loaded buildings and classrooms",
		slog.Int("buildings_count", len(res.Buildings)),
		slog.Int("classrooms_count", len(res.Classrooms)))

	return res, nil
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
	client, err := places.NewClient(ctx, option.WithAPIKey(global.GOOGLE_PLACES_API_KEY))
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
