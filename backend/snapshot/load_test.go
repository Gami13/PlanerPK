package snapshot

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	for _, tt := range []struct {
		name, body, wantError string
		status                int
	}{
		{"supported", `{"state":{"schemaVersion":1,"publicationMode":"ST","blocks":[{"id":"b1"}]}}`, "", 200},
		{"schema", `{"state":{"schemaVersion":2,"publicationMode":"ST","blocks":[{"id":"b1"}]}}`, "unsupported snapshot", 200},
		{"mode", `{"state":{"schemaVersion":1,"publicationMode":"NST","blocks":[{"id":"b1"}]}}`, "unsupported snapshot", 200},
		{"empty", `{"state":{"schemaVersion":1,"publicationMode":"ST","blocks":[]}}`, "no blocks", 200},
		{"missing state", `{}`, "missing state", 200},
		{"malformed", `{`, "decode snapshot", 200},
		{"unavailable", `{}`, "HTTP 503", 503},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()
			state, err := Load(server.Client(), server.URL)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil || len(state.Blocks) != 1 || state.Blocks[0].ID != "b1" {
				t.Fatalf("state = %+v, error = %v", state, err)
			}
		})
	}
}
