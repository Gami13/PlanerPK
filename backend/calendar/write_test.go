package calendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func TestApplyEventDiff(t *testing.T) {
	for _, failAt := range []int{0, 1, 2, 3} {
		t.Run(string(rune('0'+failAt)), func(t *testing.T) {
			var methods []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method)
				path := "/calendars/dedicated/events"
				if r.Method != http.MethodPost {
					path += "/existing"
				}
				if r.URL.Path != path {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method == http.MethodPatch {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					for _, key := range []string{"description", "location", "transparency", "colorId"} {
						if value, ok := body[key]; !ok || value != "" {
							t.Errorf("patch must clear %s: %+v", key, body)
						}
					}
					if value, ok := body["recurrence"]; !ok || value != nil || body["reminders"] != nil || body["id"] != nil {
						t.Errorf("incorrect patch: %+v", body)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if failAt == len(methods) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				w.Write([]byte(`{}`))
			}))
			defer server.Close()
			ctx := context.Background()
			srv, err := calendar.NewService(ctx, option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
			if err != nil {
				t.Fatal(err)
			}
			current := &calendar.Event{Id: "existing"}
			proposed := &calendar.Event{Summary: "Subject"}
			diff := eventDiff{ToAdd: []*calendar.Event{proposed}, Changed: []eventChange{{current, proposed}}, ToDelete: []*calendar.Event{current}}
			for _, id := range []string{"", "primary"} {
				if err := applyEventDiff(ctx, srv, id, diff); err == nil {
					t.Error("accepted invalid calendar")
				}
			}
			diff.Exceptions = []*calendar.Event{{RecurringEventId: "existing"}}
			if err := applyEventDiff(ctx, srv, "dedicated", diff); err == nil || len(methods) != 0 {
				t.Fatal("exception guard must fail before any writes")
			}
			diff.Exceptions = nil
			err = applyEventDiff(ctx, srv, "dedicated", diff)
			want := []string{"POST", "PATCH", "DELETE"}
			if failAt != 0 {
				want = want[:failAt]
			}
			if (err != nil) != (failAt != 0) || !reflect.DeepEqual(methods, want) {
				t.Fatalf("requests = %v, error = %v; want %v", methods, err, want)
			}
		})
	}
}
