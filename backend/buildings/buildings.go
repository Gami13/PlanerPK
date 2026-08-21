package buildings

import (
	"log/slog"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/woodpkrs/PlanerPK/classroommatch"
	"github.com/woodpkrs/PlanerPK/legend"
)

type Building struct {
	CanonicalName string
	GooglePlaceID string
	Name          string
	Address       string
}

type Buildings struct {
	BuildingsByKey            map[string]Building
	BuildingKeyByClassroomKey map[string]string
	matcher                   *classroommatch.Matcher
}

// BuildBuildingsFromLegend creates the legend-derived classroom-to-building
// cache for one timetable import. Persistent place details are hydrated later.
func BuildBuildingsFromLegend(legendData *legend.Legend) *Buildings {
	buildings := &Buildings{
		BuildingsByKey:            make(map[string]Building),
		BuildingKeyByClassroomKey: make(map[string]string),
	}
	if legendData == nil {
		return buildings
	}

	locationKeys := slices.Sorted(maps.Keys(legendData.Locations))

	candidates := make([]classroommatch.Candidate, 0, len(locationKeys))
	for _, locationKey := range locationKeys {
		location := legendData.Locations[locationKey]
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

		buildings.BuildingsByKey[buildingKey] = Building{CanonicalName: buildingKey}
		candidates = append(candidates, classroommatch.Candidate{
			ClassroomKey: classroomKey,
			BuildingKey:  buildingKey,
		})
	}

	buildingKeys := slices.Sorted(maps.Keys(buildings.BuildingsByKey))
	var buildingAliases []classroommatch.Candidate
	for _, buildingKey := range buildingKeys {
		if alias, ok := buildingAcronymAlias(buildingKey); ok {
			buildingAliases = append(buildingAliases, classroommatch.Candidate{ClassroomKey: alias, BuildingKey: buildingKey})
		}
	}
	buildings.matcher = classroommatch.New(candidates, buildingAliases)
	buildings.BuildingKeyByClassroomKey = buildings.matcher.ClassroomBuildings()
	return buildings
}

// buildingAcronymAlias returns an explicit building shorthand only when the
// legend names it as an uppercase acronym, for example "budynek SJO".
func buildingAcronymAlias(buildingKey string) (string, bool) {
	fields := strings.Fields(buildingKey)
	if len(fields) < 2 {
		return "", false
	}
	for i, field := range fields[:len(fields)-1] {
		if !strings.EqualFold(strings.Trim(field, "\"'(),."), "budynek") {
			continue
		}
		candidate := strings.Trim(fields[i+1], "\"'(),.")
		if IsUppercaseAcronym(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func IsUppercaseAcronym(value string) bool {
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

// ResolveClassroomKeys performs the only fuzzy matching in the runtime flow:
// while a newly imported timetable is being built.
func (b *Buildings) ResolveClassroomKeys(location string) []string {
	if b == nil || b.matcher == nil {
		return nil
	}
	var keys []string
	seen := make(map[string]struct{})
	for _, match := range b.matcher.Resolve(location) {
		switch match.Kind {
		case classroommatch.MatchExact, classroommatch.MatchCorrected:
			if _, exists := seen[match.ClassroomKey]; !exists {
				seen[match.ClassroomKey] = struct{}{}
				keys = append(keys, match.ClassroomKey)
			}
		case classroommatch.MatchAmbiguous, classroommatch.MatchUnmatched:
			// The matcher logs structured diagnostics; unresolved locations keep
			// their raw text and simply receive no map link.
		}
	}
	return keys
}

// GetBuildingsByClassroomKeys performs an exact lookup of keys that were
// precomputed while parsing a timetable.
func (b *Buildings) GetBuildingsByClassroomKeys(classroomKeys []string) []Building {
	if b == nil || b.BuildingKeyByClassroomKey == nil || b.BuildingsByKey == nil {
		return nil
	}
	var buildings []Building
	seen := make(map[string]struct{})
	for _, classroomKey := range classroomKeys {
		buildingKey, exists := b.BuildingKeyByClassroomKey[classroomKey]
		if !exists {
			continue
		}
		if _, exists := seen[buildingKey]; exists {
			continue
		}
		building, exists := b.BuildingsByKey[buildingKey]
		if !exists {
			continue
		}
		seen[buildingKey] = struct{}{}
		buildings = append(buildings, building)
	}
	return buildings
}
