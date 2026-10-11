package extractor

import (
	"regexp"
	"slices"
	"strings"

	"github.com/woodpkrs/PlanerPK/types"
)

var labGroupLabel = regexp.MustCompile(`^GL[0-9]+$`)
var exerciseLabGroups = regexp.MustCompile(`^C[0-9]+ \((GL[0-9]+(?:\+GL[0-9]+)*)\)$`)

// classesForGroup filters explicit GL audiences in regular GL columns.
// Shared/unknown audiences and supplemental columns keep their existing behavior.
func classesForGroup(classes []types.Class, label string) []types.Class {
	if !labGroupLabel.MatchString(label) {
		return classes
	}
	var included []types.Class
	for _, class := range classes {
		audience := strings.TrimSpace(class.Audience)
		if groups := exerciseLabGroups.FindStringSubmatch(audience); len(groups) != 0 {
			audience = groups[1]
		} else if !labGroupLabel.MatchString(audience) {
			included = append(included, class)
			continue
		}
		if slices.Contains(strings.Split(audience, "+"), label) {
			included = append(included, class)
		}
	}
	return included
}
