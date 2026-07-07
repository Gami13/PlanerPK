package types

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/utils"
	"google.golang.org/api/calendar/v3"
)

type Class struct {
	Name           string
	Type           string
	Day            time.Weekday
	StartTime      time.Time
	EndTime        time.Time
	Location       string
	Teacher        string
	AdditionalInfo string
	Level          string
	TotalHours     string
}

func (c Class) String() string {
	text := fmt.Sprintf("%s (%s)  %s – %s  %s @ %s",
		c.Name, c.Type,
		c.StartTime.Format("15:04"), c.EndTime.Format("15:04"),
		c.Teacher, c.Location)
	if c.Level != "" {
		text += fmt.Sprintf("  [%s]", c.Level)
	}
	if c.TotalHours != "" {
		text += fmt.Sprintf("  (%s)", c.TotalHours)
	}
	if c.AdditionalInfo != "" {
		text += fmt.Sprintf(" — %s", c.AdditionalInfo)
	}
	return text
}

func (c Class) Hash() string {
	return fmt.Sprintf("%s|%d|%s|%s|%s", c.Summary(), c.Day, c.StartTime.Format("15:04"), c.EndTime.Format("15:04"), c.Location)
}

func (c Class) Summary() string {
	summary := ""
	if len(c.Name) > 0 {
		summary += c.Name
	}
	if len(c.Level) > 0 {
		summary += fmt.Sprintf(" %s", c.Level)
	}
	if len(c.Type) > 0 {
		summary += fmt.Sprintf(" (%s)", c.Type)
	}
	if len(c.Teacher) > 0 {
		summary += fmt.Sprintf(" - %s", c.Teacher)
	}
	// if len(c.Location) > 0 {
	// 	summary += fmt.Sprintf(" | %s", c.Location)
	// }
	if len(c.TotalHours) > 0 {
		summary += fmt.Sprintf(" | %s", c.TotalHours)
	}
	if len(c.AdditionalInfo) > 0 {
		summary += fmt.Sprintf(" | %s", c.AdditionalInfo)
	}
	return summary
}

func (c Class) timeOnDay(t time.Time) time.Time {
	dayDate := utils.GetDateOfNextWeekday(c.Day)
	return time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), t.Hour(), t.Minute(), 0, 0, time.Now().Local().Location())
}

func (c Class) StartTimeRFC3339() string {
	return c.timeOnDay(c.StartTime).Format(time.RFC3339)
}

func (c Class) EndTimeRFC3339() string {
	return c.timeOnDay(c.EndTime).Format(time.RFC3339)
}

type CachedBuilding struct {
	ID            int
	Key           string
	GooglePlaceID string
	Name          string
	Address       string
}

type BuildingsCache struct {
	mu         sync.RWMutex
	buildings     map[int64]CachedBuilding
	classrooms map[string]int64
}

var BuildingsGlobalCache = &BuildingsCache{
	buildings:     make(map[int64]CachedBuilding),
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
        ID:  int(building.ID),
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
	buildings, err := queries.GetBuildings(ctx);
	if err != nil {
		return err
	}

	for _, building := range buildings {
		buildingDetails, err := fetchBuildingDetails(*building.GooglePlaceID)
		if err != nil {
			return err
		}

		queries.UpdateBuilding(ctx, db.UpdateBuildingParams{
			ID: building.ID,
			Name: &buildingDetails.Name,
			Address: &buildingDetails.Address,
		})
	}

	return err
}

func (c Class) ToEvent() calendar.Event {
	description := ""

	building, found := BuildingsGlobalCache.GetBuildingByClassroom(strings.TrimPrefix(c.Location, "s. "))
	if found {
		description = fmt.Sprintf("<a href=\"https://www.google.com/maps/search/?api=1&query=Politechnika+Krakowska&query_place_id=%s\">%s, %s</a>", building.GooglePlaceID, building.Name, building.Address)
	}

	event := &calendar.Event{
		Summary: c.Summary(),
		Start: &calendar.EventDateTime{
			DateTime: c.StartTimeRFC3339(),
		},
		End: &calendar.EventDateTime{
			DateTime: c.EndTimeRFC3339(),
		},
		Location: c.Location,
		Description: description,
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"added_by": ADDED_BY,
				"hash":     c.Hash(),
			},
		},
	}
	return *event
}
