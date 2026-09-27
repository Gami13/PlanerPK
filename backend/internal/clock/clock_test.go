package clock

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		value      string
		want       string
		wantParsed bool
	}{
		{value: "07:30", want: "07:30", wantParsed: true},
		{value: "7:30", want: "07:30", wantParsed: true},
		{value: "08.15", want: "08:15", wantParsed: true},
		{value: "0.3125"},
		{value: "7.3"},
		{value: "25:00"},
	}
	for _, test := range tests {
		parsed, ok := Parse(test.value)
		if ok != test.wantParsed {
			t.Errorf("Parse(%q) ok = %v, want %v", test.value, ok, test.wantParsed)
		}
		if ok && parsed.Format("15:04") != test.want {
			t.Errorf("Parse(%q) = %s, want %s", test.value, parsed.Format("15:04"), test.want)
		}
	}
}
