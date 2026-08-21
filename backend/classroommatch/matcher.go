// Package classroommatch resolves timetable classroom text against the
// classroom entries published in an Excel legend.
package classroommatch

import (
	"log/slog"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/agnivade/levenshtein"
)

// Candidate identifies one canonical classroom and the building it belongs to.
type Candidate struct {
	ClassroomKey string
	BuildingKey  string
}

// MatchKind describes the outcome for one physical location part.
type MatchKind string

const (
	MatchExact     MatchKind = "exact"
	MatchCorrected MatchKind = "corrected"
	MatchAmbiguous MatchKind = "ambiguous"
	MatchUnmatched MatchKind = "unmatched"
	MatchIgnored   MatchKind = "ignored"
)

// Result contains the resolution outcome for a single location part.
type Result struct {
	Input        string
	Normalized   string
	ClassroomKey string
	BuildingKey  string
	Kind         MatchKind
	Score        float64
}

// Matcher indexes legend classrooms and resolves raw timetable location text.
type Matcher struct {
	exactByAlias    map[string]Candidate
	buildingByAlias map[string]buildingAlias
	fuzzyCandidates []Candidate
}

type buildingAlias struct {
	ClassroomAlias string
	BuildingKey    string
}

// New creates a deterministic matcher with separate classroom and building
// alias indexes. Conflicting aliases are excluded rather than chosen by order.
func New(candidates, buildingAliases []Candidate) *Matcher {
	exactByAlias, fuzzyCandidates, classroomCollisions := buildExactClassroomIndex(candidates)
	buildingByAlias, buildingAliasCollisions := buildBuildingAliasIndex(buildingAliases, exactByAlias)
	m := &Matcher{
		exactByAlias:    exactByAlias,
		buildingByAlias: buildingByAlias,
		fuzzyCandidates: fuzzyCandidates,
	}

	slog.Info("classroom_resolution.index_built",
		slog.Int("candidates_count", len(m.fuzzyCandidates)),
		slog.Int("classroom_keys_count", len(m.exactByAlias)),
		slog.Int("building_aliases_count", len(m.buildingByAlias)),
		slog.Int("collisions_count", classroomCollisions+buildingAliasCollisions),
		slog.String("source", "legend"))
	return m
}

func buildExactClassroomIndex(candidates []Candidate) (map[string]Candidate, []Candidate, int) {
	ordered := sortedCandidates(candidates)
	exactByAlias := make(map[string]Candidate)
	fuzzyCandidates := make([]Candidate, 0, len(ordered))
	conflicts := make(map[string]struct{})

	for _, candidate := range ordered {
		rawClassroomKey := candidate.ClassroomKey
		candidate.ClassroomKey = Normalize(candidate.ClassroomKey)
		candidate.BuildingKey = strings.TrimSpace(candidate.BuildingKey)
		if candidate.ClassroomKey == "" || candidate.BuildingKey == "" {
			continue
		}
		fuzzyCandidates = append(fuzzyCandidates, candidate)

		for _, alias := range classroomAliasesForRawText(rawClassroomKey) {
			if _, conflict := conflicts[alias]; conflict {
				continue
			}
			existing, exists := exactByAlias[alias]
			if !exists || existing.BuildingKey == candidate.BuildingKey {
				exactByAlias[alias] = candidate
				continue
			}

			delete(exactByAlias, alias)
			conflicts[alias] = struct{}{}
			slog.Warn("classroom_resolution.collision",
				slog.String("classroom_key", alias),
				slog.String("first_building_key", existing.BuildingKey),
				slog.String("second_building_key", candidate.BuildingKey),
				slog.String("source", "legend"))
		}
	}
	return exactByAlias, fuzzyCandidates, len(conflicts)
}

func buildBuildingAliasIndex(candidates []Candidate, exactByAlias map[string]Candidate) (map[string]buildingAlias, int) {
	buildingAliases := make(map[string]buildingAlias)
	conflicts := make(map[string]struct{})
	for _, candidate := range sortedCandidates(candidates) {
		alias := Normalize(candidate.ClassroomKey)
		buildingKey := strings.TrimSpace(candidate.BuildingKey)
		if alias == "" || buildingKey == "" {
			continue
		}
		if _, conflict := conflicts[alias]; conflict {
			continue
		}
		if exact, exists := exactByAlias[alias]; exists && exact.BuildingKey != buildingKey {
			delete(buildingAliases, alias)
			conflicts[alias] = struct{}{}
			logAliasCollision(alias, exact.BuildingKey, buildingKey, "building_alias")
			continue
		}
		if existing, exists := buildingAliases[alias]; exists && existing.BuildingKey != buildingKey {
			delete(buildingAliases, alias)
			conflicts[alias] = struct{}{}
			logAliasCollision(alias, existing.BuildingKey, buildingKey, "building_alias")
			continue
		}
		buildingAliases[alias] = buildingAlias{ClassroomAlias: alias, BuildingKey: buildingKey}
	}
	return buildingAliases, len(conflicts)
}

func sortedCandidates(candidates []Candidate) []Candidate {
	ordered := slices.Clone(candidates)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].ClassroomKey == ordered[j].ClassroomKey {
			return ordered[i].BuildingKey < ordered[j].BuildingKey
		}
		return ordered[i].ClassroomKey < ordered[j].ClassroomKey
	})
	return ordered
}

func logAliasCollision(alias, firstBuildingKey, secondBuildingKey, source string) {
	slog.Warn("classroom_resolution.collision",
		slog.String("classroom_key", alias),
		slog.String("first_building_key", firstBuildingKey),
		slog.String("second_building_key", secondBuildingKey),
		slog.String("source", source))
}

// The two index builders intentionally keep their precedence rules separate:
// exact classroom aliases are indexed first, while a building alias may only
// coexist with an exact alias that points at the same building.

// ClassroomBuildings returns the normalized keys emitted by exact classroom
// and building-alias matches.
func (m *Matcher) ClassroomBuildings() map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m.exactByAlias)+len(m.buildingByAlias))
	for key, alias := range m.buildingByAlias {
		result[key] = alias.BuildingKey
	}
	for key, candidate := range m.exactByAlias {
		result[key] = candidate.BuildingKey
	}
	return result
}

// Resolve splits a compound location and resolves every physical classroom
// part. It never chooses an ambiguous fuzzy candidate.
func (m *Matcher) Resolve(location string) []Result {
	parts := splitLocation(location)
	results := make([]Result, 0, len(parts))
	for _, part := range parts {
		results = append(results, m.resolvePart(part))
	}
	return results
}

func (m *Matcher) resolvePart(input string) Result {
	normalized := Normalize(input)
	result := Result{Input: input, Normalized: normalized}
	if normalized == "" || normalized == "zdalnie" {
		result.Kind = MatchIgnored
		return result
	}
	if m == nil {
		result.Kind = MatchUnmatched
		return result
	}

	if exact, found := m.resolveExact(input, normalized); found {
		return exact
	}
	if building, found := m.resolveBuildingAlias(input, normalized); found {
		return building
	}
	return m.resolveFuzzy(input, normalized)
}

func (m *Matcher) resolveExact(input, normalized string) (Result, bool) {
	for _, alias := range normalizedClassroomAliases(normalized) {
		if candidate, ok := m.exactByAlias[alias]; ok {
			return Result{
				Input:        input,
				Normalized:   normalized,
				ClassroomKey: alias,
				BuildingKey:  candidate.BuildingKey,
				Kind:         MatchExact,
				Score:        1,
			}, true
		}
	}
	return Result{}, false
}

func (m *Matcher) resolveBuildingAlias(input, normalized string) (Result, bool) {
	alias, found := m.buildingAliasIn(normalized)
	if !found {
		return Result{}, false
	}
	result := Result{
		Input:        input,
		Normalized:   normalized,
		ClassroomKey: alias.ClassroomAlias,
		BuildingKey:  alias.BuildingKey,
		Kind:         MatchExact,
		Score:        1,
	}
	slog.Debug("classroom_resolution.building_alias",
		slog.String("input", input),
		slog.String("normalized", normalized),
		slog.String("classroom_key", alias.ClassroomAlias),
		slog.String("building_key", alias.BuildingKey),
		slog.String("source", "legend"))
	return result, true
}

func (m *Matcher) resolveFuzzy(input, normalized string) Result {
	result := Result{Input: input, Normalized: normalized}

	threshold, canCorrect := fuzzyThreshold(normalized)
	if !canCorrect {
		result.Kind = MatchUnmatched
		logUnmatched(result)
		return result
	}

	bestScore, secondScore := -1.0, -1.0
	var best Candidate
	for _, candidate := range m.fuzzyCandidates {
		score := similarity(normalized, candidate.ClassroomKey)
		if score > bestScore {
			secondScore = bestScore
			bestScore = score
			best = candidate
		} else if score > secondScore {
			secondScore = score
		}
	}

	result.Score = bestScore
	if bestScore >= threshold && bestScore-secondScore >= 0.10 {
		result.ClassroomKey = best.ClassroomKey
		result.BuildingKey = best.BuildingKey
		result.Kind = MatchCorrected
		slog.Info("classroom_resolution.corrected",
			slog.String("input", input),
			slog.String("normalized", normalized),
			slog.String("classroom_key", best.ClassroomKey),
			slog.String("building_key", best.BuildingKey),
			slog.Float64("score", bestScore),
			slog.String("source", "legend"))
		return result
	}

	if bestScore >= threshold {
		result.Kind = MatchAmbiguous
		slog.Warn("classroom_resolution.ambiguous",
			slog.String("input", input),
			slog.String("normalized", normalized),
			slog.String("best_candidate", best.ClassroomKey),
			slog.Float64("best_score", bestScore),
			slog.Float64("second_score", secondScore),
			slog.String("source", "legend"))
		return result
	}

	result.Kind = MatchUnmatched
	logUnmatched(result)
	return result
}

func (m *Matcher) buildingAliasIn(normalized string) (buildingAlias, bool) {
	var result buildingAlias
	found := false
	for _, word := range strings.Fields(normalized) {
		alias, exists := m.buildingByAlias[word]
		if !exists {
			continue
		}
		if found && result.BuildingKey != alias.BuildingKey {
			return buildingAlias{}, false
		}
		result = alias
		found = true
	}
	return result, found
}

func logUnmatched(result Result) {
	slog.Warn("classroom_resolution.unmatched",
		slog.String("input", result.Input),
		slog.String("normalized", result.Normalized),
		slog.Float64("best_score", result.Score),
		slog.String("source", "legend"))
}

func fuzzyThreshold(normalized string) (float64, bool) {
	length := len([]rune(normalized))
	switch {
	case length <= 3:
		return 0, false
	case length <= 5:
		return 0.90, true
	case length <= 8:
		return 0.82, true
	default:
		return 0.75, true
	}
}

func similarity(left, right string) float64 {
	maxLength := max(len([]rune(left)), len([]rune(right)))
	if maxLength == 0 {
		return 1
	}
	return 1 - float64(levenshtein.ComputeDistance(left, right))/float64(maxLength)
}

// Normalize produces the stable key used by both the index and class records.
func Normalize(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimSpace(strings.TrimPrefix(value, "s."))
	value = strings.TrimSpace(strings.TrimPrefix(value, "s "))

	words := strings.FieldsFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	sort.Strings(words)
	return strings.Join(words, " ")
}

func normalizedClassroomAliases(normalized string) []string {
	if normalized == "" {
		return nil
	}
	result := []string{normalized}
	words := strings.Fields(normalized)
	if len(words) > 1 && words[len(words)-1] == "bis" {
		result = append(result, strings.Join(words[:len(words)-1], " "))
	}
	return result
}

func classroomAliasesForRawText(value string) []string {
	result := normalizedClassroomAliases(Normalize(value))
	for _, variant := range rawClassroomTextVariants(value) {
		if alias := Normalize(variant); alias != "" && !slices.Contains(result, alias) {
			result = append(result, alias)
		}
	}
	return result
}

func rawClassroomTextVariants(value string) []string {
	withoutParentheses := value
	for {
		start := strings.Index(withoutParentheses, "(")
		if start == -1 {
			break
		}
		end := strings.Index(withoutParentheses[start:], ")")
		if end == -1 {
			break
		}
		withoutParentheses = withoutParentheses[:start] + withoutParentheses[start+end+1:]
	}
	if withoutParentheses == value {
		return nil
	}
	return []string{withoutParentheses}
}

func splitLocation(location string) []string {
	var result []string
	var current strings.Builder
	runes := []rune(location)
	flush := func() {
		if part := strings.TrimSpace(current.String()); part != "" {
			result = append(result, part)
		}
		current.Reset()
	}

	for i, r := range runes {
		switch r {
		case ',', ';', '\n', '\r':
			flush()
		case '/':
			if numericRoomFraction(runes, i) {
				current.WriteRune(r)
			} else {
				flush()
			}
		default:
			current.WriteRune(r)
		}
	}
	flush()
	return result
}

func numericRoomFraction(runes []rune, slashIndex int) bool {
	if slashIndex == 0 || slashIndex+1 >= len(runes) || !unicode.IsDigit(runes[slashIndex-1]) || !unicode.IsDigit(runes[slashIndex+1]) {
		return false
	}
	leftLength := 0
	for i := slashIndex - 1; i >= 0 && unicode.IsDigit(runes[i]); i-- {
		leftLength++
	}
	rightLength := 0
	for i := slashIndex + 1; i < len(runes) && unicode.IsDigit(runes[i]); i++ {
		rightLength++
	}
	return leftLength <= 2 && rightLength <= 2
}
