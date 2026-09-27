package legacy

import (
	"regexp"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/types"
)

// Global so it's not recompiled every time the function is called
var splitClassTextRegex = regexp.MustCompile(`(  |\n)+`)

// Matches duration patterns like "15h", "30h", "(15h)", "(5h/10h)", "(14h/8h/8h)", "(15h)(5h/10h)"
var totalHoursRegex = regexp.MustCompile(`(?:\(?\d+h(?:/\d+h)*\)?)+`)

// Matches standalone level/group patterns like "C1", "B2", "A1"
var levelRegex = regexp.MustCompile(`^[A-C][12]$`)

// Matches trailing language level in a name like "J.angielski C1"
var trailingLevelRegex = regexp.MustCompile(`\s+([A-C][12])$`)

// ParseClassText parses one class cell from the legacy heuristic format.
func ParseClassText(text string, classStartTime time.Time, classEndTime time.Time, day types.Day, legend *legend.Legend) (class types.Class) {
	class.Day = day.WeekDay
	class.StartTime = classStartTime
	class.EndTime = classEndTime

	parts := splitClassTextRegex.Split(text, -1)
	parts = removeEmptyStrings(parts)
	parts = mergeAcademicTitleFragments(parts)
	if len(parts) == 0 {
		return class
	}

	// --- Step 1: Name (first part), potentially strip level ---
	name := strings.TrimSpace(parts[0])
	if lvl := extractTrailingLevel(name); lvl != "" {
		class.Level = lvl
		name = strings.TrimSpace(strings.TrimSuffix(name, lvl))
	}
	class.Name = name

	// --- Step 2: Iterate remaining parts through heuristic pipeline ---
	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}

		// Check for class type (possibly merged with teacher, e.g. "lab. Michał Drygaś")
		if typeName, remainder := extractClassType(part); typeName != "" {
			if class.Type == "" {
				class.Type = typeName
			}
			if remainder != "" {
				// The remainder after the type could be a teacher or group info
				classifyToken(remainder, legend, &class)
			}
			continue
		}

		// Total hours pattern: "(15h)", "(5h/10h)", "15h"
		if isTotalHoursOnly(part) && !containsAdditionalInfoKeyword(part) {
			class.TotalHours = appendField(class.TotalHours, part)
			continue
		}

		// Additional info keywords
		if containsAdditionalInfoKeyword(part) {
			class.AdditionalInfo = appendField(class.AdditionalInfo, part)
			continue
		}

		// Level / group pattern
		if levelRegex.MatchString(part) {
			class.Level = part
			continue
		}

		// Classify as location, teacher, or fallback
		classifyToken(part, legend, &class)
	}

	return class
}

// classifyToken determines whether a token is a location, teacher, or unknown,
// and assigns it to the appropriate field on the class.
func classifyToken(token string, legend *legend.Legend, class *types.Class) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}

	// Location heuristics
	if isLocation(token, legend) {
		class.Location = appendField(class.Location, token)
		return
	}

	// Teacher heuristics
	if isTeacher(token, legend) {
		class.Teacher = appendField(class.Teacher, token)
		return
	}

	// Total hours as fallback (in case it wasn't caught earlier)
	if isTotalHoursOnly(token) {
		class.TotalHours = appendField(class.TotalHours, token)
		return
	}

	// Fallback: append to AdditionalInfo
	class.AdditionalInfo = appendField(class.AdditionalInfo, token)
}

// isLocation returns true if the token looks like a location.
func isLocation(token string, legend *legend.Legend) bool {
	if strings.HasPrefix(token, "s.") || strings.EqualFold(token, "ZDALNIE") {
		return true
	}
	if legend != nil {
		if hasLocationSegments(token, legend) {
			return true
		}
		if _, ok := legend.GetLocation(token); ok {
			return true
		}
	}
	return false
}

// hasLocationSegments checks if a slash-delimited token has at least half of
// its segments recognized as locations (by legend or heuristic).
func hasLocationSegments(token string, legend *legend.Legend) bool {
	if !strings.Contains(token, "/") {
		return false
	}
	segments := strings.Split(token, "/")
	locationCount := 0
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, "s.") || strings.EqualFold(seg, "ZDALNIE") {
			locationCount++
		} else if _, ok := legend.GetLocation(seg); ok {
			locationCount++
		}
	}
	return locationCount > 0 && locationCount >= len(segments)/2
}

// isTeacher returns true if the token looks like a teacher name.
func isTeacher(token string, legend *legend.Legend) bool {
	if hasAcademicTitle(token) {
		return true
	}
	if legend == nil {
		return false
	}

	segments := strings.Split(token, "/")
	teacherCount := 0
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if _, ok := legend.GetTeacher(seg); ok {
			teacherCount++
		} else if hasAcademicTitle(seg) {
			teacherCount++
		}
	}
	return teacherCount > 0 && teacherCount >= (len(segments)+1)/2
}

// hasAcademicTitle checks if a string starts with a known academic title.
func hasAcademicTitle(s string) bool {
	lower := strings.ToLower(s)
	for _, title := range types.AcademicTitles {
		lt := strings.ToLower(title)
		if strings.HasPrefix(lower, lt+" ") || strings.HasPrefix(lower, lt+",") {
			return true
		}
	}
	return false
}

// extractClassType checks if a part starts with a known ClassType.
// Returns the type name and any remaining text after it.
func extractClassType(part string) (typeName string, remainder string) {
	lowerPart := strings.ToLower(part)
	longestMatch := ""
	for _, ct := range types.ClassTypes {
		if strings.HasPrefix(lowerPart, strings.ToLower(ct)) && len(ct) > len(longestMatch) {
			longestMatch = ct
		}
	}
	if longestMatch == "" {
		return "", ""
	}
	remainder = strings.TrimSpace(part[len(longestMatch):])
	return longestMatch, remainder
}

// extractTrailingLevel strips a trailing language level like "C1", "B2" from a name.
func extractTrailingLevel(name string) string {
	m := trailingLevelRegex.FindStringSubmatch(name)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// isTotalHoursOnly checks if the entire token is a total-hours pattern.
func isTotalHoursOnly(token string) bool {
	stripped := strings.TrimSpace(token)
	if stripped == "" {
		return false
	}
	match := totalHoursRegex.FindString(stripped)
	return match == stripped
}

// containsAdditionalInfoKeyword checks if the token contains frequency/warning keywords.
func containsAdditionalInfoKeyword(token string) bool {
	lower := strings.ToLower(token)
	for _, kw := range types.AdditionalInfoKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// appendField concatenates values with a separator, building up multi-value fields.
func appendField(existing, value string) string {
	if existing == "" {
		return value
	}
	return existing + " " + value
}

// mergeAcademicTitleFragments detects parts that are bare academic titles
// (e.g. "dr inż.") caused by double-space splitting, and merges them with
// the following part to reconstruct the full teacher name.
func mergeAcademicTitleFragments(parts []string) []string {
	var result []string
	for i := 0; i < len(parts); i++ {
		p := strings.TrimSpace(parts[i])
		if isBareAcademicTitle(p) && i+1 < len(parts) {
			// Merge with next part
			result = append(result, p+" "+strings.TrimSpace(parts[i+1]))
			i++ // skip next
		} else {
			result = append(result, p)
		}
	}
	return result
}

// isBareAcademicTitle returns true if the string is just an academic title
// with no name following it (e.g. "dr inż." or "mgr").
func isBareAcademicTitle(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	for _, title := range types.AcademicTitles {
		if strings.ToLower(title) == lower || strings.ToLower(title)+"." == lower {
			return true
		}
	}
	return false
}

// SplitSlashClasses splits a class with a slash-delimited Name into separate classes.
// Fields like Teacher and Location are distributed by position when their segment
// count matches the name count; otherwise the whole field is shared.
// Level is assigned only to the last name (where it was originally extracted from).
func SplitSlashClasses(class types.Class) []types.Class {
	if !strings.Contains(class.Name, "/") {
		return []types.Class{class}
	}

	names := strings.Split(class.Name, "/")
	teachers := splitOrRepeat(class.Teacher, len(names))
	locations := splitOrRepeat(class.Location, len(names))

	classes := make([]types.Class, 0, len(names))
	for i, name := range names {
		c := types.Class{
			Name:           strings.TrimSpace(name),
			Type:           class.Type,
			Day:            class.Day,
			StartTime:      class.StartTime,
			EndTime:        class.EndTime,
			Teacher:        strings.TrimSpace(teachers[i]),
			Audience:       class.Audience,
			Location:       strings.TrimSpace(locations[i]),
			AdditionalInfo: class.AdditionalInfo,
			TotalHours:     class.TotalHours,
		}
		if i == len(names)-1 {
			c.Level = class.Level
		}
		classes = append(classes, c)
	}
	return classes
}

// splitOrRepeat splits a field by "/" if the segment count matches expected;
// otherwise repeats the whole field for each slot.
func splitOrRepeat(field string, count int) []string {
	if field == "" {
		return make([]string, count)
	}
	parts := strings.Split(field, "/")
	if len(parts) == count {
		return parts
	}
	result := make([]string, count)
	for i := range result {
		result[i] = field
	}
	return result
}

// Removes empty strings from a slice
func removeEmptyStrings(slice []string) []string {
	var result []string
	for _, str := range slice {
		if strings.TrimSpace(str) != "" {
			result = append(result, str)
		}
	}
	return result
}
