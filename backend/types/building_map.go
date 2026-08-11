package types

import (
	"fmt"
	"html"
	"net/url"
)

func (b Building) MapURL() string {
	if b.GooglePlaceID == "" {
		return ""
	}
	return "https://www.google.com/maps/search/?api=1&query=Politechnika+Krakowska&query_place_id=" + url.QueryEscape(b.GooglePlaceID)
}

func (b Building) MapLabel() string {
	switch {
	case b.Name != "" && b.Address != "":
		return fmt.Sprintf("%s, %s", b.Name, b.Address)
	case b.Name != "":
		return b.Name
	case b.Address != "":
		return b.Address
	case b.GooglePlaceID != "":
		return "Google Maps"
	default:
		return ""
	}
}

func (b Building) MapHTMLAnchor() string {
	if url := b.MapURL(); url != "" {
		return fmt.Sprintf("<a href=\"%s\">%s</a>", html.EscapeString(url), html.EscapeString(b.MapLabel()))
	}
	return ""
}
