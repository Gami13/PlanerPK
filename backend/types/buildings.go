package types

import (
	"log/slog"
	"sort"
	"strings"
	"unicode"

	"github.com/woodpkrs/PlanerPK/classroommatch"
)

type Building struct {
	CanonicalName string
	GooglePlaceID string
	Name          string
	Address       string
}

type Buildings struct {
	Buildings  map[string]Building
	Classrooms map[string]string
	matcher    *classroommatch.Matcher
	reviews    *BuildingReviewQueue
}

// BuildBuildingsFromLegend creates the legend-derived classroom-to-building
// cache for one timetable import. Persistent place details are hydrated later.
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

		result.Buildings[buildingKey] = Building{CanonicalName: buildingKey}
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

// ResolveClassroomKeys performs the only fuzzy matching in the runtime flow:
// while a newly imported timetable is being built.
func (b *Buildings) ResolveClassroomKeys(location string) []string {
	if b == nil || b.matcher == nil {
		return nil
	}
	keys := make([]string, 0)
	seen := make(map[string]struct{})
	for _, match := range b.matcher.Resolve(location) {
		if (match.Kind == classroommatch.MatchAmbiguous || match.Kind == classroommatch.MatchUnmatched) && b.reviews != nil {
			b.reviews.RecordClassroomMatch(match)
		}
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
