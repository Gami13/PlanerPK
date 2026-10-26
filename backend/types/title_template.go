package types

import "regexp"

const DefaultTitleTemplate = "{ClassTitle} ({ClassType}) - {ClassTeacher}"

var titlePlaceholder = regexp.MustCompile(`\{([^{}]+)\}`)

// FormatTitle substitutes known fields once, preserving punctuation and unknown
// placeholders exactly as typed. Values are literal, never recursively expanded.
func (c Class) FormatTitle(template string) string {
	values := map[string]string{
		"{ClassTitle}":      c.Name,
		"{ClassType}":       c.Type,
		"{ClassTeacher}":    c.Teacher,
		"{ClassAudience}":   c.Audience,
		"{ClassRecurrence}": c.AdditionalInfo,
	}
	return titlePlaceholder.ReplaceAllStringFunc(template, func(placeholder string) string {
		if value, ok := values[placeholder]; ok {
			return value
		}
		return placeholder
	})
}
