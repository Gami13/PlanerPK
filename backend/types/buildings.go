package types

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/woodpkrs/PlanerPK/gen/db"
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
	bList, err := queries.GetBuildings(ctx)
	if err != nil {
		return nil, err
	}
	cList, err := queries.GetClassroomNamesWithBuildingIDs(ctx)
	if err != nil {
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

	return res, nil
}

type BuildingDetails struct {
	Name    string
	Address string
}

func fetchBuildingDetails(googlePlaceID string) (*BuildingDetails, error) {
	apiKey := os.Getenv("GOOGLE_PLACES_NEW_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_PLACES_NEW_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("https://places.googleapis.com/v1/places/%s?languageCode=pl", googlePlaceID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "displayName,formattedAddress")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiData struct {
		DisplayName struct {
			Text string `json:"text"`
		} `json:"displayName"`
		FormattedAddress string `json:"formattedAddress"`
	}

	if err := json.Unmarshal(body, &apiData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &BuildingDetails{
		Name:    apiData.DisplayName.Text,
		Address: apiData.FormattedAddress,
	}, nil
}

func UpdateBuildingsDetails(ctx context.Context, queries db.Querier) error {
	buildings, err := queries.GetBuildings(ctx)
	if err != nil {
		return err
	}

	for _, building := range buildings {
		buildingDetails, err := fetchBuildingDetails(*building.GooglePlaceID)
		if err != nil {
			return err
		}

		queries.UpdateBuilding(ctx, db.UpdateBuildingParams{
			ID:      building.ID,
			Name:    &buildingDetails.Name,
			Address: &buildingDetails.Address,
		})
	}

	return err
}
