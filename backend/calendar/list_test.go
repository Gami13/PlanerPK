package calendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func TestListManagedEvents(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "all pages", true: "second page fails"}[fail], func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				q := r.URL.Query()
				if r.Method != http.MethodGet || r.URL.Path != "/calendars/dedicated/events" ||
					q.Get("singleEvents") != "false" || q.Get("showDeleted") != "true" ||
					q.Has("timeMin") || q.Has("orderBy") || q.Has("privateExtendedProperty") {
					t.Errorf("unexpected listing request: %s %s", r.Method, r.URL)
				}
				w.Header().Set("Content-Type", "application/json")
				result := calendar.Events{}
				if q.Get("pageToken") == "" {
					result.NextPageToken = "next"
					result.Items = []*calendar.Event{
						{Id: "cancelled", Status: "cancelled", RecurringEventId: "series", OriginalStartTime: &calendar.EventDateTime{DateTime: "2026-10-14T08:00:00+02:00"}},
						{Id: "unrelated", Summary: "Personal event"},
						{Id: "foreign exception", RecurringEventId: "foreign"},
					}
				} else {
					if fail {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					result.Items = []*calendar.Event{
						{Id: "series", Recurrence: []string{"RRULE:FREQ=WEEKLY"}, ExtendedProperties: &calendar.EventExtendedProperties{Private: map[string]string{"added_by": types.ADDED_BY}}},
						{Id: "modified", RecurringEventId: "series"},
						{Id: "single", ExtendedProperties: &calendar.EventExtendedProperties{Private: map[string]string{"added_by": types.ADDED_BY}}},
					}
				}
				if err := json.NewEncoder(w).Encode(result); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			ctx := context.Background()
			srv, err := calendar.NewService(ctx, option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
			if err != nil {
				t.Fatal(err)
			}
			got, err := getCurrentEvents(ctx, srv, "dedicated")
			if fail {
				if err == nil || got != nil {
					t.Fatalf("events = %+v, error = %v; want no partial result", got, err)
				}
			} else {
				var ids []string
				for _, event := range got {
					ids = append(ids, event.Id)
				}
				if err != nil || !reflect.DeepEqual(ids, []string{"cancelled", "series", "modified", "single"}) || got[0].OriginalStartTime == nil || len(got[1].Recurrence) != 1 {
					t.Fatalf("events = %+v, error = %v", got, err)
				}
			}
			for _, id := range []string{"", "primary"} {
				if _, err := getCurrentEvents(ctx, srv, id); err == nil {
					t.Errorf("accepted invalid calendar ID %q", id)
				}
			}
			if requests != 2 {
				t.Fatalf("requests = %d; want 2", requests)
			}
		})
	}
}
