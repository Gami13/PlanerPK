package types

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
)

const (
	minimumPlaceScore         = 0.70
	minimumPlaceScoreLead     = 0.15
	buildingEvidenceWeight    = 0.70
	campusContextWeight       = 0.15
	cityContextWeight         = 0.15
	subBuildingAddressPenalty = 0.35
)

var buildingNameIgnoredTokens = map[string]struct{}{
	"budynek":      {},
	"budynku":      {},
	"building":     {},
	"ul":           {},
	"ulica":        {},
	"warszawska":   {},
	"krakow":       {},
	"politechnika": {},
	"krakowska":    {},
	"pk":           {},
}

type scoredPlace struct {
	Place   *placespb.Place
	Score   float64
	Reasons []string
}

type placeResolution struct {
	Best     scoredPlace
	Ranked   []scoredPlace
	Reasons  []string
	Accepted bool
}

func choosePlace(canonicalName string, candidates []*placespb.Place) placeResolution {
	ranked := make([]scoredPlace, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil || candidate.GetId() == "" {
			continue
		}
		score, reasons := scorePlace(canonicalName, candidate)
		ranked = append(ranked, scoredPlace{Place: candidate, Score: score, Reasons: reasons})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].Place.GetId() < ranked[j].Place.GetId()
		}
		return ranked[i].Score > ranked[j].Score
	})
	if len(ranked) == 0 {
		return placeResolution{Ranked: ranked, Reasons: []string{"Google Text Search returned no candidates"}}
	}
	best := ranked[0]
	if best.Score < minimumPlaceScore {
		return placeResolution{Best: best, Ranked: ranked, Reasons: append(best.Reasons, fmt.Sprintf("candidate confidence is below %.2f", minimumPlaceScore))}
	}
	if len(ranked) > 1 && best.Score-ranked[1].Score < minimumPlaceScoreLead {
		return placeResolution{Best: best, Ranked: ranked, Reasons: append(best.Reasons, fmt.Sprintf("best candidate does not lead runner-up by %.2f", minimumPlaceScoreLead))}
	}
	return placeResolution{Best: best, Ranked: ranked, Reasons: best.Reasons, Accepted: true}
}

func scorePlace(canonicalName string, place *placespb.Place) (float64, []string) {
	text := paddedFoldPlaceSearchText(placeDisplayName(place) + " " + place.GetFormattedAddress())
	var score float64
	reasons := make([]string, 0, 3)
	if hasBuildingNameEvidence(canonicalName, text) || hasBuildingInitialismEvidence(canonicalName, placeDisplayName(place)) {
		score += buildingEvidenceWeight
		reasons = append(reasons, "canonical building-name evidence")
	} else {
		reasons = append(reasons, "no canonical building-name evidence")
	}
	if strings.Contains(text, "politechnika") || strings.Contains(text, " pk ") {
		score += campusContextWeight
		reasons = append(reasons, "Politechnika Krakowska context")
	}
	if strings.Contains(text, "krakow") {
		score += cityContextWeight
		reasons = append(reasons, "Krakow context")
	}
	if isSubBuildingLocation(canonicalName, place.GetFormattedAddress()) {
		score -= subBuildingAddressPenalty
		reasons = append(reasons, "candidate identifies a floor or room within the building")
	}
	return score, reasons
}

func placeDisplayName(place *placespb.Place) string {
	if place == nil || place.GetDisplayName() == nil {
		return ""
	}
	return place.GetDisplayName().GetText()
}

// hasBuildingNameEvidence expects candidateSearchText to be padded with one
// space at each end, as returned by paddedFoldPlaceSearchText.
func hasBuildingNameEvidence(canonicalName, candidateSearchText string) bool {
	for _, token := range buildingNameTokens(canonicalName) {
		if strings.Contains(candidateSearchText, " "+token+" ") {
			return true
		}
	}
	return false
}

// hasBuildingInitialismEvidence lets a legend acronym match the initialism of
// a Google display name, such as SJO and Studium Języków Obcych.
func hasBuildingInitialismEvidence(canonicalName, displayName string) bool {
	initialism := displayNameInitialism(displayName)
	if initialism == "" {
		return false
	}
	for _, token := range strings.Fields(canonicalName) {
		acronym := strings.ToUpper(strings.Trim(token, "\"'(),."))
		if isUppercaseAcronym(acronym) && strings.HasPrefix(initialism, acronym) {
			return true
		}
	}
	return false
}

func displayNameInitialism(displayName string) string {
	var result strings.Builder
	for _, word := range strings.Fields(displayName) {
		r, _ := utf8.DecodeRuneInString(word)
		if unicode.IsUpper(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func isSubBuildingLocation(canonicalName, address string) bool {
	foldedAddress := paddedFoldPlaceSearchText(address)
	for _, token := range buildingNameTokens(canonicalName) {
		if start := strings.Index(foldedAddress, " "+token+" "); start >= 0 {
			remainder := foldedAddress[start+len(token)+2:]
			for _, marker := range []string{" pietro ", " pokoj ", " sala ", " gabinet ", " lokal "} {
				if strings.Contains(remainder, marker) {
					return true
				}
			}
		}
	}
	return false
}

func buildingNameTokens(canonicalName string) []string {
	var tokens []string
	for _, token := range strings.Fields(paddedFoldPlaceSearchText(canonicalName)) {
		if _, skip := buildingNameIgnoredTokens[token]; !skip && len([]rune(token)) >= 3 {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

// paddedFoldPlaceSearchText folds case and Polish diacritics, replaces
// punctuation with spaces, and pads the result for whole-token substring
// searches.
func paddedFoldPlaceSearchText(value string) string {
	var result strings.Builder
	for _, r := range strings.ToLower(value) {
		switch r {
		case 'ą':
			r = 'a'
		case 'ć':
			r = 'c'
		case 'ę':
			r = 'e'
		case 'ł':
			r = 'l'
		case 'ń':
			r = 'n'
		case 'ó':
			r = 'o'
		case 'ś':
			r = 's'
		case 'ź', 'ż':
			r = 'z'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	return " " + strings.Join(strings.Fields(result.String()), " ") + " "
}
