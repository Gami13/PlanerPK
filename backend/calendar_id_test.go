package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/woodpkrs/PlanerPK/gen/db"
)

func TestUserCalendarIDPersistence(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err := database.Exec(ddl); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	queries := db.New(database)
	email := "user@example.test"
	login := db.UpsertUserParams{Email: &email, AccessToken: "first-token", Expiry: time.Now()}
	user, err := queries.UpsertUser(ctx, login)
	if err != nil || user.CalendarID != "" {
		t.Fatalf("new user = %+v, error = %v", user, err)
	}
	if err := queries.UpdateUserCalendarID(ctx, db.UpdateUserCalendarIDParams{ID: user.ID, CalendarID: "dedicated-id"}); err != nil {
		t.Fatal(err)
	}
	login.AccessToken = "new-token"
	loggedIn, err := queries.UpsertUser(ctx, login)
	if err != nil || loggedIn.ID != user.ID || loggedIn.CalendarID != "dedicated-id" {
		t.Fatalf("calendar ID lost on login: user = %+v, error = %v", loggedIn, err)
	}
	byEmail, err := queries.GetUserByEmail(ctx, &email)
	if err != nil || byEmail.CalendarID != "dedicated-id" {
		t.Fatalf("email lookup = %+v, error = %v", byEmail, err)
	}
	byToken, err := queries.GetUserByAccessToken(ctx, login.AccessToken)
	if err != nil || byToken.CalendarID != "dedicated-id" {
		t.Fatalf("token lookup = %+v, error = %v", byToken, err)
	}
	users, err := queries.GetAllUsers(ctx)
	if err != nil || len(users) != 1 || users[0].CalendarID != "dedicated-id" {
		t.Fatalf("user listing = %+v, error = %v", users, err)
	}
}
