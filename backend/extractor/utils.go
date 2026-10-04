package extractor

import "strings"

// isHourLabel returns true for values that represent hour/time columns (e.g. "Godziny").
func isHourLabel(v string) bool {
	return strings.EqualFold(strings.TrimSpace(v), "Godziny")
}

// groupYearDigit extracts the first digit after "grupa " (e.g. "grupa 11" → '1',
// "grupa 21" → '2'). Returns 0 if the value doesn't match the pattern.
func groupYearDigit(value string) byte {
	rest, found := strings.CutPrefix(strings.ToLower(value), "grupa ")
	if !found {
		return 0
	}
	rest = strings.TrimSpace(rest)
	if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
		return rest[0]
	}
	return 0
}
