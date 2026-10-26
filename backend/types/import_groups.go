package types

import (
	"regexp"
	"strconv"
	"strings"
)

var semesterCategory = regexp.MustCompile(`(?i)^(.*?)\bsem\.?\s*(\d+)(.*)$`)

func sameAcademicYear(a, b string) bool {
	left, right := semesterCategory.FindStringSubmatch(a), semesterCategory.FindStringSubmatch(b)
	if left == nil || right == nil {
		// No reliable year information: do not guess cross-group eligibility.
		return false
	}
	first, _ := strconv.Atoi(left[2])
	second, _ := strconv.Atoi(right[2])
	return first > 0 && second > 0 && (first+1)/2 == (second+1)/2 &&
		strings.EqualFold(strings.TrimSpace(left[1]), strings.TrimSpace(right[1])) &&
		strings.EqualFold(strings.TrimSpace(left[3]), strings.TrimSpace(right[3]))
}

func CanImportGroup(groups []ColumnGroup, own, source int64) bool {
	if own == source || source <= 0 {
		return false
	}
	var base, other *ColumnGroup
	for i := range groups {
		if int64(groups[i].Column) == own {
			base = &groups[i]
		}
		if int64(groups[i].Column) == source {
			other = &groups[i]
		}
	}
	return base != nil && other != nil && sameAcademicYear(base.Category, other.Category)
}
