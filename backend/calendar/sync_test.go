package calendar

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func TestSyncTimetable(t *testing.T) {
	for _, mode := range []string{"concurrent", "list failure", "partial write retry", "bad schedule", "exceptions"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			database, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()
			database.SetMaxOpenConns(1)
			schema, err := os.ReadFile("../schema.sql")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := database.Exec(string(schema)); err != nil {
				t.Fatal(err)
			}
			queries := db.New(database)
			user, err := queries.UpsertUser(ctx, db.UpsertUserParams{AccessToken: "token", Expiry: time.Now()})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := database.Exec(`UPDATE users SET chosen_column=1, last_synced='2000-01-01 00:00:00+00:00'`); err != nil {
				t.Fatal(err)
			}
			day, start := "WED", 480
			state := &snapshot.State{Blocks: []snapshot.Block{{ID: "block", Day: &day, Start: &start, Duration: 90, Frequency: "co_tydzien"}},
				AcademicCalendar: snapshot.AcademicCalendar{Start: "2026-10-05", End: "2026-10-21", ParityAnchor: "2026-09-28"}}
			class := types.Class{Name: "Subject", SnapshotBlockID: "block"}
			all := types.Timetable{1: {time.Wednesday: {class}}}
			if mode == "partial write retry" {
				other := class
				other.Name = "Second"
				all[1][time.Wednesday] = append(all[1][time.Wednesday], other)
			}
			var stored []*calendar.Event
			if mode == "exceptions" {
				events, err := buildTimetableEvents(all[1], state, nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				events[0].Id, events[0].Summary = "series", "Old details"
				stored = []*calendar.Event{events[0], {Id: "exception", RecurringEventId: "series", Status: "cancelled"}}
			}
			if mode == "bad schedule" {
				state.Blocks[0].Frequency = "unknown"
			}
			creates, writes, requests := 0, 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == "POST" && r.URL.Path == "/calendars":
					creates++
					fmt.Fprint(w, `{"id":"dedicated"}`)
				case r.Method == "GET" && r.URL.Path == "/calendars/dedicated/events":
					if mode == "list failure" {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					json.NewEncoder(w).Encode(calendar.Events{Items: stored})
				case r.Method == "POST" && r.URL.Path == "/calendars/dedicated/events":
					writes++
					if mode == "partial write retry" && writes == 2 {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					var event calendar.Event
					if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
						t.Error(err)
						return
					}
					event.Id, event.Transparency = fmt.Sprint(writes), "opaque"
					stored = append(stored, &event)
					json.NewEncoder(w).Encode(event)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()
			srv, err := calendar.NewService(ctx, option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
			if err != nil {
				t.Fatal(err)
			}
			run := func() error { return syncTimetable(ctx, srv, &all, "token", queries, state, nil, nil) }
			if mode == "concurrent" {
				var wg sync.WaitGroup
				for range 2 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := run(); err != nil {
							t.Error(err)
						}
					}()
				}
				wg.Wait()
				if creates != 1 || writes != 1 {
					t.Fatalf("duplicate work: creates=%d writes=%d", creates, writes)
				}
			} else {
				if err := run(); err == nil {
					t.Fatal("expected sync failure")
				}
				failed, err := queries.GetUserByAccessToken(ctx, "token")
				if err != nil || failed.LastSynced.Year() != 2000 {
					t.Fatalf("failed sync marked successful: %+v, %v", failed, err)
				}
				if mode == "partial write retry" {
					if err := run(); err != nil {
						t.Fatal(err)
					}
					if creates != 1 || writes != 3 || len(stored) != 2 {
						t.Fatalf("retry duplicated events: creates=%d writes=%d stored=%d", creates, writes, len(stored))
					}
				}
			}
			if mode == "bad schedule" && requests != 0 {
				t.Fatal("invalid schedule reached Google")
			}
			if mode == "exceptions" && writes != 0 {
				t.Fatal("exceptions should block writes")
			}
			final, err := queries.GetUserByAccessToken(ctx, "token")
			if err != nil || final.ID != user.ID {
				t.Fatal(err)
			}
			if (mode == "concurrent" || mode == "partial write retry") && (final.CalendarID != "dedicated" || final.LastSynced.Year() == 2000) {
				t.Fatalf("successful sync not persisted: %+v", final)
			}
		})
	}
}
