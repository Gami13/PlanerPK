package types

import "testing"

func TestFormatTitle(t *testing.T) {
	class := Class{Name: "Systemy operacyjne", Type: "P", Teacher: "Wojtas Jan", Audience: "GL2", AdditionalInfo: "A"}
	for _, tt := range []struct {
		name, template, want string
	}{
		{"default", DefaultTitleTemplate, "Systemy operacyjne (P) - Wojtas Jan"},
		{"custom recurrence", "{ClassRecurrence}", "A"},
		{"audience", "[{ClassAudience}]", "[GL2]"},
		{"unknown and repeated", "{Unknown} {ClassTitle}/{ClassTitle}", "{Unknown} Systemy operacyjne/Systemy operacyjne"},
		{"empty", "", ""},
		{"literal", "My class", "My class"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := class.FormatTitle(tt.template); got != tt.want {
				t.Fatalf("FormatTitle() = %q, want %q", got, tt.want)
			}
		})
	}
	if got := (Class{Name: "Algebra"}).FormatTitle(DefaultTitleTemplate); got != "Algebra () - " {
		t.Fatalf("empty fields lost punctuation: %q", got)
	}
	if got := (Class{Name: "$& {ClassTeacher}"}).FormatTitle("{ClassTitle}"); got != "$& {ClassTeacher}" {
		t.Fatalf("replacement wasn't literal: %q", got)
	}
	if got := class.ToEvent(nil, nil).Summary; got != class.FormatTitle(DefaultTitleTemplate) {
		t.Fatalf("Google Calendar bypassed template: %q", got)
	}
}
