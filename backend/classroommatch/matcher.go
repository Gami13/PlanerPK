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
	exactAliases    map[string]Candidate
	buildingAliases map[string]Candidate
	ordered         []Candidate
}

// New creates a deterministic matcher with separate classroom and building
// alias indexes. Conflicting aliases are excluded rather than chosen by order.
func New(candidates, buildingAliases []Candidate) *Matcher {
	ordered := slices.Clone(candidates)
	buildingAliases = slices.Clone(buildingAliases)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].ClassroomKey == ordered[j].ClassroomKey {
			return ordered[i].BuildingKey < ordered[j].BuildingKey
		}
		return ordered[i].ClassroomKey < ordered[j].ClassroomKey
	})

	m := &Matcher{
		exactAliases:    make(map[string]Candidate),
		buildingAliases: make(map[string]Candidate),
		ordered:         make([]Candidate, 0, len(ordered)),
	}
	conflicts := make(map[string]struct{})
	for _, candidate := range ordered {
		rawClassroomKey := candidate.ClassroomKey
		candidate.ClassroomKey = Normalize(candidate.ClassroomKey)
		candidate.BuildingKey = strings.TrimSpace(candidate.BuildingKey)
		if candidate.ClassroomKey == "" || candidate.BuildingKey == "" {
			continue
		}
		m.ordered = append(m.ordered, candidate)

		for _, alias := range aliasesForRaw(rawClassroomKey) {
			if _, conflict := conflicts[alias]; conflict {
				continue
			}
			existing, exists := m.exactAliases[alias]
			if !exists || existing.BuildingKey == candidate.BuildingKey {
				m.exactAliases[alias] = candidate
				continue
			}

			delete(m.exactAliases, alias)
			conflicts[alias] = struct{}{}
			slog.Warn("classroom_resolution.collision",
				slog.String("classroom_key", alias),
				slog.String("first_building_key", existing.BuildingKey),
				slog.String("second_building_key", candidate.BuildingKey),
				slog.String("source", "legend"))
		}
	}

	buildingAliasConflicts := make(map[string]struct{})
	sort.Slice(buildingAliases, func(i, j int) bool {
		if buildingAliases[i].ClassroomKey == buildingAliases[j].ClassroomKey {
			return buildingAliases[i].BuildingKey < buildingAliases[j].BuildingKey
		}
		return buildingAliases[i].ClassroomKey < buildingAliases[j].ClassroomKey
	})
	for _, candidate := range buildingAliases {
		alias := Normalize(candidate.ClassroomKey)
		buildingKey := strings.TrimSpace(candidate.BuildingKey)
		if alias == "" || buildingKey == "" {
			continue
		}
		if _, conflict := buildingAliasConflicts[alias]; conflict {
			continue
		}
		if exact, exists := m.exactAliases[alias]; exists && exact.BuildingKey != buildingKey {
			delete(m.buildingAliases, alias)
			buildingAliasConflicts[alias] = struct{}{}
			slog.Warn("classroom_resolution.collision",
				slog.String("classroom_key", alias),
				slog.String("first_building_key", exact.BuildingKey),
				slog.String("second_building_key", buildingKey),
				slog.String("source", "building_alias"))
			continue
		}
		if existing, exists := m.buildingAliases[alias]; exists && existing.BuildingKey != buildingKey {
			delete(m.buildingAliases, alias)
			buildingAliasConflicts[alias] = struct{}{}
			slog.Warn("classroom_resolution.collision",
				slog.String("classroom_key", alias),
				slog.String("first_building_key", existing.BuildingKey),
				slog.String("second_building_key", buildingKey),
				slog.String("source", "building_alias"))
			continue
		}
		m.buildingAliases[alias] = Candidate{ClassroomKey: alias, BuildingKey: buildingKey}
	}

	slog.Info("classroom_resolution.index_built",
		slog.Int("candidates_count", len(m.ordered)),
		slog.Int("classroom_keys_count", len(m.exactAliases)),
		slog.Int("building_aliases_count", len(m.buildingAliases)),
		slog.Int("collisions_count", len(conflicts)+len(buildingAliasConflicts)),
		slog.String("source", "legend"))
	return m
}

// ClassroomBuildings returns the normalized keys emitted by exact classroom
// and building-alias matches.
func (m *Matcher) ClassroomBuildings() map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m.exactAliases)+len(m.buildingAliases))
	for key, candidate := range m.buildingAliases {
		result[key] = candidate.BuildingKey
	}
	for key, candidate := range m.exactAliases {
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

	for _, alias := range aliases(normalized) {
		if candidate, ok := m.exactAliases[alias]; ok {
			result.ClassroomKey = alias
			result.BuildingKey = candidate.BuildingKey
			result.Kind = MatchExact
			result.Score = 1
			return result
		}
	}
	if candidate, found := m.buildingAliasIn(normalized); found {
		result.ClassroomKey = candidate.ClassroomKey
		result.BuildingKey = candidate.BuildingKey
		result.Kind = MatchExact
		result.Score = 1
		slog.Debug("classroom_resolution.building_alias",
			slog.String("input", input),
			slog.String("normalized", normalized),
			slog.String("classroom_key", candidate.ClassroomKey),
			slog.String("building_key", candidate.BuildingKey),
			slog.String("source", "legend"))
		return result
	}

	threshold, canCorrect := fuzzyThreshold(normalized)
	if !canCorrect {
		result.Kind = MatchUnmatched
		logUnmatched(result)
		return result
	}

	bestScore, secondScore := -1.0, -1.0
	var best Candidate
	for _, candidate := range m.ordered {
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

func (m *Matcher) buildingAliasIn(normalized string) (Candidate, bool) {
	var result Candidate
	found := false
	for _, word := range strings.Fields(normalized) {
		candidate, exists := m.buildingAliases[word]
		if !exists {
			continue
		}
		if found && result.BuildingKey != candidate.BuildingKey {
			return Candidate{}, false
		}
		result = candidate
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

func aliases(normalized string) []string {
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

func aliasesForRaw(value string) []string {
	result := aliases(Normalize(value))
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
	if alias := Normalize(withoutParentheses); alias != "" && !slices.Contains(result, alias) {
		result = append(result, alias)
	}
	return result
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
