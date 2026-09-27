package types

import (
	"strings"
	"testing"
)

func TestAudienceIsIncludedInClassOutput(t *testing.T) {
	class := Class{Name: "Subject", Audience: "GL2"}
	if !strings.Contains(class.Summary(), "GL2") {
		t.Errorf("Summary() = %q, want audience", class.Summary())
	}
	if !strings.Contains(class.String(), "GL2") {
		t.Errorf("String() = %q, want audience", class.String())
	}
	event := class.ToEvent(nil)
	if !strings.Contains(event.Summary, "GL2") || !strings.Contains(event.Description, "GL2") || event.ExtendedProperties.Private["audience"] != "GL2" {
		t.Errorf("ToEvent() omitted audience: %+v", event)
	}
}
