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
	ClassroomKey  string
	BuildingKey   string
	BuildingAlias bool
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
	byAlias map[string]Candidate
	ordered []Candidate
}

// New creates a deterministic matcher. Aliases that belong to different
// buildings are excluded so they can never silently select an arbitrary one.
func New(candidates []Candidate) *Matcher {
	ordered := slices.Clone(candidates)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].ClassroomKey == ordered[j].ClassroomKey {
			return ordered[i].BuildingKey < ordered[j].BuildingKey
		}
		return ordered[i].ClassroomKey < ordered[j].ClassroomKey
	})

	m := &Matcher{
		byAlias: make(map[string]Candidate),
		ordered: make([]Candidate, 0, len(ordered)),
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
			existing, exists := m.byAlias[alias]
			if !exists || existing.BuildingKey == candidate.BuildingKey {
				m.byAlias[alias] = candidate
				continue
			}

			delete(m.byAlias, alias)
			conflicts[alias] = struct{}{}
			slog.Warn("classroom_resolution.collision",
				slog.String("classroom_key", alias),
				slog.String("first_building_key", existing.BuildingKey),
				slog.String("second_building_key", candidate.BuildingKey),
				slog.String("source", "legend"))
		}
	}

	slog.Info("classroom_resolution.index_built",
		slog.Int("candidates_count", len(m.ordered)),
		slog.Int("classroom_keys_count", len(m.byAlias)),
		slog.Int("collisions_count", len(conflicts)),
		slog.String("source", "legend"))
	return m
}

// ClassroomBuildings returns the exact normalized classroom-to-building map.
func (m *Matcher) ClassroomBuildings() map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m.byAlias))
	for key, candidate := range m.byAlias {
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
		if candidate, ok := m.byAlias[alias]; ok {
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
		candidate, exists := m.byAlias[word]
		if !exists || !candidate.BuildingAlias {
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
