package buildings

import "testing"

func TestBuildingMapPresentation(t *testing.T) {
	tests := []struct {
		name       string
		building   Building
		wantLabel  string
		wantURL    string
		wantAnchor string
	}{
		{
			name:       "confirmed place with name and address",
			building:   Building{GooglePlaceID: "chem", Name: "Wydział Chemii", Address: "Kraków"},
			wantLabel:  "Wydział Chemii, Kraków",
			wantURL:    "https://www.google.com/maps/search/?api=1&query=Politechnika+Krakowska&query_place_id=chem",
			wantAnchor: "<a href=\"https://www.google.com/maps/search/?api=1&amp;query=Politechnika+Krakowska&amp;query_place_id=chem\">Wydział Chemii, Kraków</a>",
		},
		{
			name:       "escapes place ID and HTML label",
			building:   Building{GooglePlaceID: "id &/", Name: "Chemia & <lab>"},
			wantLabel:  "Chemia & <lab>",
			wantURL:    "https://www.google.com/maps/search/?api=1&query=Politechnika+Krakowska&query_place_id=id+%26%2F",
			wantAnchor: "<a href=\"https://www.google.com/maps/search/?api=1&amp;query=Politechnika+Krakowska&amp;query_place_id=id+%26%2F\">Chemia &amp; &lt;lab&gt;</a>",
		},
		{
			name:       "unconfirmed place",
			building:   Building{Name: "Chemia"},
			wantLabel:  "Chemia",
			wantURL:    "",
			wantAnchor: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.building.MapLabel(); got != test.wantLabel {
				t.Errorf("MapLabel() = %q, want %q", got, test.wantLabel)
			}
			if got := test.building.MapURL(); got != test.wantURL {
				t.Errorf("MapURL() = %q, want %q", got, test.wantURL)
			}
			if got := test.building.MapHTMLAnchor(); got != test.wantAnchor {
				t.Errorf("MapHTMLAnchor() = %q, want %q", got, test.wantAnchor)
			}
		})
	}
}
