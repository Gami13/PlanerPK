package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func TestCreateDedicatedCalendar(t *testing.T) {
	for _, mode := range []string{"create", "create failure", "missing ID"} {
		t.Run(mode, func(t *testing.T) {
			creates := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method != http.MethodPost || r.URL.Path != "/calendars" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				creates++
				var body calendar.Calendar
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Summary != "PlanerPK" || body.TimeZone != "Europe/Warsaw" {
					t.Errorf("invalid creation body: %+v, %v", body, err)
				}
				if mode == "create failure" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				if mode == "missing ID" {
					fmt.Fprint(w, `{}`)
					return
				}
				fmt.Fprint(w, `{"id":"created"}`)
			}))
			defer server.Close()
			ctx := context.Background()
			srv, err := calendar.NewService(ctx, option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if mode == "create" {
				want = "created"
			}
			got, err := createDedicatedCalendar(ctx, srv)
			if got != want || (err != nil) != (want == "") {
				t.Fatalf("ID = %q, error = %v; want %q", got, err, want)
			}
			if creates != 1 {
				t.Fatalf("creation requests = %d; want 1", creates)
			}
		})
	}
}
