package extractor

import (
	"regexp"
	"strings"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/types"
)

// Global so it's not recompiled every time the function is called
var splitClassTextRegex = regexp.MustCompile(`(  |\n)+`)

func parseClassText(text string, classStartTime time.Time, classEndTime time.Time, day types.Day) (class types.Class) {
	class.Day = day.WeekDay
	class.StartTime = classStartTime
	class.EndTime = classEndTime

	parts := splitClassTextRegex.Split(text, -1)
	parts = removeEmptyStrings(parts)
	parts = fixClassParts(parts)
	if len(parts) == 0 {
		return class
	}
	class.Name = strings.TrimSpace(parts[0])

	if len(parts) == 1 {
		return class
	}

	if len(parts) == 2 {
		if isClassType(parts[1]) {
			class.Type = strings.TrimSpace(parts[1])
		} else {
			class.Teacher = strings.TrimSpace(parts[1])
		}
		return class
	}

	if isClassType(parts[1]) {
		class.Type = strings.TrimSpace(parts[1])
		class.Teacher = strings.TrimSpace(parts[2])
	} else if isClassType(parts[2]) {
		class.Teacher = strings.TrimSpace(parts[1])
		class.Type = strings.TrimSpace(parts[2])
	} else {
		class.Teacher = strings.TrimSpace(parts[1])
		class.Location = strings.TrimSpace(parts[2])
	}

	if len(parts) > 3 && class.Location == "" {
		class.Location = strings.TrimSpace(parts[3])
	}
	if len(parts) > 4 {
		class.AdditionalInfo = strings.TrimSpace(parts[4])
	}

	return class
}

func isClassType(value string) bool {
	for _, classType := range types.ClassTypes {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(classType)) {
			return true
		}
	}
	return false
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

// Fixes it so that the last part (the one that contains additional info) is merged into one string
func fixClassParts(parts []string) []string {
	if len(parts) > 4 {
		merged := strings.Join(parts[4:], " ")
		return append(parts[:4], merged)
	}
	return parts
}
