package calendar

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	_ "modernc.org/sqlite"
)

func TestEnsureUserCalendar(t *testing.T) {
	for _, mode := range []string{"create and reuse", "creation failure", "save failure"} {
		t.Run(mode, func(t *testing.T) {
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
			ctx := context.Background()
			queries := db.New(database)
			email := "user@example.test"
			user, err := queries.UpsertUser(ctx, db.UpsertUserParams{Email: &email, Expiry: time.Now()})
			if err != nil {
				t.Fatal(err)
			}
			if mode == "save failure" {
				if _, err := database.Exec(`CREATE TRIGGER fail_calendar_save BEFORE UPDATE OF calendar_id ON users BEGIN SELECT RAISE(FAIL, 'save failed'); END`); err != nil {
					t.Fatal(err)
				}
			}
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodPost || r.URL.Path != "/calendars" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				w.Header().Set("Content-Type", "application/json")
				if mode == "creation failure" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				fmt.Fprint(w, `{"id":"dedicated-id"}`)
			}))
			defer server.Close()
			srv, err := calendar.NewService(ctx, option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
			if err != nil {
				t.Fatal(err)
			}
			id, err := EnsureUserCalendar(ctx, srv, queries, user)
			want := ""
			if mode == "create and reuse" {
				want = "dedicated-id"
			}
			if id != want || (err != nil) != (want == "") {
				t.Fatalf("ID = %q, error = %v; want %q", id, err, want)
			}
			stored, err := queries.GetUserByEmail(ctx, &email)
			if err != nil || stored.CalendarID != want {
				t.Fatalf("stored user = %+v, error = %v", stored, err)
			}
			if want != "" {
				id, err = EnsureUserCalendar(ctx, srv, queries, stored)
				if err != nil || id != want {
					t.Fatalf("reuse ID = %q, error = %v", id, err)
				}
			}
			if requests != 1 {
				t.Fatalf("Google requests = %d; want 1", requests)
			}
		})
	}
}
