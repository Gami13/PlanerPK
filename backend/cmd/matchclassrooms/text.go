package main

import (
	"slices"
	"strings"
	"unicode"
)

// extractWords splits s into lowercase alphanumeric words.
func extractWords(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// isSubset reports whether every word in sub appears in set.
func isSubset(sub, set []string) bool {
	for _, w := range sub {
		if !slices.Contains(set, w) {
			return false
		}
	}
	return true
}

// tokenSort normalizes a string by lowercasing, removing '.', filtering
// the standalone word "s", and sorting the remaining tokens.
func tokenSort(s string) string {
	fields := strings.Fields(strings.ReplaceAll(strings.ToLower(s), ".", ""))
	filtered := slices.DeleteFunc(slices.Clone(fields), func(f string) bool { return f == "s" })
	if len(filtered) > 0 {
		fields = filtered
	}
	slices.Sort(fields)
	return strings.Join(fields, " ")
}

// splitBySlash splits input on '/' unless the token following the slash
// is ≤ 2 characters (e.g. "1/15" stays together).
func splitBySlash(input string) []string {
	raw := strings.Split(input, "/")
	if len(raw) <= 1 {
		return raw
	}

	var result []string
	curr := raw[0]
	for _, part := range raw[1:] {
		fields := strings.Fields(strings.Trim(part, ".,; "))
		if len(fields) > 0 && len([]rune(fields[0])) <= 2 {
			curr += "/" + part
		} else {
			result = append(result, strings.TrimSpace(curr))
			curr = part
		}
	}
	return append(result, strings.TrimSpace(curr))
}
