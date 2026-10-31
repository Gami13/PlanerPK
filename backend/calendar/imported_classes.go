package calendar

import (
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/types"
)

// UserTimetable returns a fresh map/slices: the shared cache must never be mutated.
// Missing or no-longer-eligible source classes are omitted rather than copied stale.
func UserTimetable(all types.Timetable, own int64, overrides []db.Override, groups []types.ColumnGroup) types.GroupTimetable {
	result := make(types.GroupTimetable)
	seen := make(map[string]bool)
	for day, classes := range all[own] {
		result[day] = append([]types.Class(nil), classes...)
		for _, class := range classes {
			seen[class.Hash()] = true
		}
	}
	for _, override := range overrides {
		if !types.CanImportGroup(groups, own, override.SourceColumn) || seen[override.ClassHash] {
			continue
		}
		for day, classes := range all[override.SourceColumn] {
			for _, class := range classes {
				if class.Hash() == override.ClassHash && !seen[override.ClassHash] {
					result[day] = append(result[day], class)
					seen[override.ClassHash] = true
				}
			}
		}
	}
	return result
}
