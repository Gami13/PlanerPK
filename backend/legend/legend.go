package legend

import "strings"

type Teacher struct {
	Shorthand string
	FullName  string
	Email     string
}

type Location struct {
	Shorthand string
	FullName  string
}

type Legend struct {
	Teachers  map[string]Teacher
	Locations map[string]Location
}

// GetLocation looks up a location by exact shorthand first, then falls back
// to matching with an "s. " prefix (common in timetable cell text).
func (l *Legend) GetLocation(name string) (Location, bool) {
	if loc, ok := l.Locations[name]; ok {
		return loc, true
	}
	if loc, ok := l.Locations["s. "+name]; ok {
		return loc, true
	}
	trimmed := strings.TrimPrefix(name, "s. ")
	if trimmed != name {
		if loc, ok := l.Locations[trimmed]; ok {
			return loc, true
		}
	}
	return Location{}, false
}

func (l *Legend) GetTeacher(shorthand string) (Teacher, bool) {
	teacher, ok := l.Teachers[shorthand]
	return teacher, ok
}
