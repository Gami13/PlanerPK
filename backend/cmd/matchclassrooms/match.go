package main

import (
	"strings"

	"github.com/agnivade/levenshtein"
)

// ScoreThreshold is the minimum Levenshtein similarity ratio for a fuzzy match.
const ScoreThreshold = 0.7

// MatchKind classifies how a part was matched.
type MatchKind int

const (
	KindExact     MatchKind = iota // Subset-word match.
	KindFuzzy                      // Levenshtein score above threshold.
	KindUnmatched                  // No good candidate found.
)

// PartResult holds the match outcome for a single slash-separated part.
type PartResult struct {
	Input string
	Match string
	Kind  MatchKind
	Score float64 // 0–1, only meaningful for fuzzy matches.
}

// LineResult holds the match outcome for a full input line.
type LineResult struct {
	Input  string       // The original input string.
	Output string       // Matched parts joined by "/".
	Parts  []PartResult // Per-part match details.
}

// candidate is a pre-processed classroom name for efficient matching.
type candidate struct {
	raw   string
	words []string
	norm  string
}

// buildCandidates pre-processes a list of classroom names.
func buildCandidates(classrooms []string) []candidate {
	cands := make([]candidate, len(classrooms))
	for i, name := range classrooms {
		cands[i] = candidate{
			raw:   name,
			words: extractWords(name),
			norm:  tokenSort(name),
		}
	}
	return cands
}

// MatchLine matches an input string against a list of known classroom names.
// It splits the input on '/' boundaries, matches each part independently,
// and returns the combined result.
func MatchLine(input string, classrooms []string) LineResult {
	cands := buildCandidates(classrooms)

	parts := splitBySlash(input)
	results := make([]PartResult, len(parts))
	matched := make([]string, len(parts))

	for i, part := range parts {
		results[i] = matchPart(part, cands)
		matched[i] = results[i].Match
	}

	return LineResult{
		Input:  input,
		Output: strings.Join(matched, "/"),
		Parts:  results,
	}
}

// matchPart finds the best candidate for a single classroom string part.
func matchPart(part string, cands []candidate) PartResult {
	words := extractWords(part)

	// Try exact subset match first (skip if input is just "s").
	if len(words) > 0 && !(len(words) == 1 && words[0] == "s") {
		var best candidate
		found := false
		for _, c := range cands {
			if isSubset(words, c.words) {
				if !found || len(c.words) < len(best.words) {
					best = c
					found = true
				}
			}
		}
		if found {
			return PartResult{Input: part, Match: best.raw, Kind: KindExact, Score: 1.0}
		}
	}

	// Fall back to fuzzy Levenshtein match.
	normInput := tokenSort(part)
	bestScore := -1.0
	bestRaw := part
	for _, c := range cands {
		if s := ratio(normInput, c.norm); s > bestScore {
			bestScore = s
			bestRaw = c.raw
		}
	}
	if bestScore >= ScoreThreshold {
		return PartResult{Input: part, Match: bestRaw, Kind: KindFuzzy, Score: bestScore}
	}
	return PartResult{Input: part, Match: part, Kind: KindUnmatched, Score: bestScore}
}

// ratio returns the Levenshtein similarity ratio between a and b (0–1).
func ratio(a, b string) float64 {
	maxLen := max(len([]rune(a)), len([]rune(b)))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(levenshtein.ComputeDistance(a, b))/float64(maxLen)
}
