package types

import (
	"sort"
	"strings"
	"unicode"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
)

const (
	minimumPlaceScore         = 0.70
	minimumPlaceScoreLead     = 0.15
	subBuildingAddressPenalty = 0.35
)

type scoredPlace struct {
	Place   *placespb.Place
	Score   float64
	Reasons []string
}

func choosePlace(canonicalName string, candidates []*placespb.Place) (scoredPlace, []scoredPlace, []string, bool) {
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
		return scoredPlace{}, ranked, []string{"Google Text Search returned no candidates"}, false
	}
	best := ranked[0]
	if best.Score < minimumPlaceScore {
		return best, ranked, append(best.Reasons, "candidate confidence is below 0.70"), false
	}
	if len(ranked) > 1 && best.Score-ranked[1].Score < minimumPlaceScoreLead {
		return best, ranked, append(best.Reasons, "best candidate does not lead runner-up by 0.15"), false
	}
	return best, ranked, best.Reasons, true
}

func scorePlace(canonicalName string, place *placespb.Place) (float64, []string) {
	text := foldPlaceText(placeDisplayName(place) + " " + place.GetFormattedAddress())
	var score float64
	reasons := make([]string, 0, 3)
	if hasBuildingNameEvidence(canonicalName, text) || hasBuildingInitialismEvidence(canonicalName, placeDisplayName(place)) {
		score += 0.70
		reasons = append(reasons, "canonical building-name evidence")
	} else {
		reasons = append(reasons, "no canonical building-name evidence")
	}
	if strings.Contains(text, "politechnika") || strings.Contains(text, " pk ") {
		score += 0.15
		reasons = append(reasons, "Politechnika Krakowska context")
	}
	if strings.Contains(text, "krakow") {
		score += 0.15
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

func hasBuildingNameEvidence(canonicalName, candidateText string) bool {
	for _, token := range buildingNameTokens(canonicalName) {
		if strings.Contains(candidateText, " "+token+" ") {
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
		for _, r := range word {
			if unicode.IsUpper(r) {
				result.WriteRune(r)
			}
			break
		}
	}
	return result.String()
}

func isSubBuildingLocation(canonicalName, address string) bool {
	foldedAddress := foldPlaceText(address)
	for _, token := range buildingNameTokens(canonicalName) {
		for start := strings.Index(foldedAddress, " "+token+" "); start >= 0; {
			remainder := foldedAddress[start+len(token)+2:]
			for _, marker := range []string{" pietro ", " pokoj ", " sala ", " gabinet ", " lokal "} {
				if strings.Contains(remainder, marker) {
					return true
				}
			}
			next := strings.Index(foldedAddress[start+len(token)+2:], " "+token+" ")
			if next < 0 {
				break
			}
			start += len(token) + 2 + next
		}
	}
	return false
}

func buildingNameTokens(canonicalName string) []string {
	ignored := map[string]struct{}{"budynek": {}, "budynku": {}, "building": {}, "ul": {}, "ulica": {}, "warszawska": {}, "krakow": {}, "politechnika": {}, "krakowska": {}, "pk": {}}
	var tokens []string
	for _, token := range strings.Fields(foldPlaceText(canonicalName)) {
		if _, skip := ignored[token]; !skip && len([]rune(token)) >= 3 {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func foldPlaceText(value string) string {
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
