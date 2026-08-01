package types

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/woodpkrs/PlanerPK/gen/db"
)

type CachedBuilding struct {
	ID            int
	Key           string
	GooglePlaceID string
	Name          string
	Address       string
}

type BuildingsCache struct {
	mu         sync.RWMutex
	buildings  map[int64]CachedBuilding
	classrooms map[string]int64
}

var BuildingsGlobalCache = &BuildingsCache{
	buildings:  make(map[int64]CachedBuilding),
	classrooms: make(map[string]int64),
}

func (c *BuildingsCache) SetBuilding(id int64, building CachedBuilding) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buildings[id] = building
}

func (c *BuildingsCache) SetClassroom(classroomName string, buildingId int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.classrooms[classroomName] = buildingId
}

func (c *BuildingsCache) GetBuilding(id int64) (CachedBuilding, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	building, exists := c.buildings[id]
	return building, exists
}

func (c *BuildingsCache) GetBuildingByClassroom(classroomName string) (CachedBuilding, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	buildingKey, roomExists := c.classrooms[classroomName]
	if !roomExists {
		return CachedBuilding{}, false
	}

	building, buildingExists := c.buildings[buildingKey]
	return building, buildingExists
}

func MapToCachedBuildings(building db.Building) CachedBuilding {
	cachedBuilding := CachedBuilding{
		ID: int(building.ID),
	}

	if building.Key != nil {
		cachedBuilding.Key = *building.Key
	}
	if building.GooglePlaceID != nil {
		cachedBuilding.GooglePlaceID = *building.GooglePlaceID
	}
	if building.Name != nil {
		cachedBuilding.Name = *building.Name
	}
	if building.Address != nil {
		cachedBuilding.Address = *building.Address
	}

	return cachedBuilding
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
