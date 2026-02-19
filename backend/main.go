package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log/slog"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/gen/db"

	"github.com/WoodPKrs/GoogleCalendarTimetable/calendar"
	"github.com/WoodPKrs/GoogleCalendarTimetable/extractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/global"

	"github.com/WoodPKrs/GoogleCalendarTimetable/utils"
	"github.com/WoodPKrs/GoogleCalendarTimetable/webServer"
)

//go:embed schema.sql
var ddl string

// initializeDatabase creates tables if they don't exist.
func initializeDatabase() *db.Queries {
	database, err := sql.Open(global.DATABASE_DRIVER, global.DATABASE_SOURCE)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			slog.Error("Failed to close database", slog.Any("error", err))
		}
	}()

	_, err = database.Exec(ddl)
	// we can panic here
	if err != nil {
		slog.Error("Failed to execute DDL statements", slog.Any("error", err))
		panic(err)
	}

	return db.New(database)
}

func main() {
	// Initialize logger first
	if err := utils.InitializeLogging(global.LOGS_PATH); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	utils.LoadEnvVariables()
	slog.Info("Environment variables loaded")

	// Fetch the timetable url
	_, _, err := extractor.UpdateTimetable()
	if err != nil {
		slog.Error("Failed to update timetable", slog.Any("error", err))
	}

	queries := initializeDatabase()

	slog.Info("Database initialized successfully")

	ticker := time.NewTicker(time.Duration(global.MINUTES_BETWEEN) * time.Minute)
	done := make(chan bool)

	slog.Info("Starting web server", slog.String("port", global.SERVER_PORT))

	go webServer.Start(queries)

	slog.Info("Starting timetable update ticker",
		slog.Int("interval_minutes", global.MINUTES_BETWEEN))
	func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				slog.Info("Timetable update triggered", slog.Time("tick_time", t))
				updateAllTimetables(queries)
			}
		}
	}()
}

func updateAllTimetables(queries *db.Queries) {
	users, err := queries.GetAllUsers(context.Background())
	if err != nil {
		slog.Error("Failed to fetch users", slog.Any("error", err))
		return
	}
	didChange, _, err := extractor.UpdateTimetable()
	if err != nil {
		slog.Error("Failed to update timetable", slog.Any("error", err))
		return
	}

	for _, user := range users {
		if *user.ChosenColumn < 0 || !*user.IsActive {
			slog.Info("Skipping timetable update for inactive user or user without chosen column",
				slog.String("email", *user.Email),
				slog.Int64("column", *user.ChosenColumn))

			continue
		}
		timeSinceSync := time.Since(user.LastSynced)
		if !didChange && timeSinceSync.Hours() < float64(global.TIME_BETWEEN_SYNC_HOURS) {
			slog.Info("Skipping timetable update for user (no changes detected and last sync within 3 days)",
				slog.String("email", *user.Email),
				slog.Int64("column", *user.ChosenColumn))
			continue
		}

		slog.Info("Updating timetable for user",
			slog.String("email", *user.Email),
			slog.Int64("column", *user.ChosenColumn))
		calendar.ImportTimetable(*user.ChosenColumn, utils.GetToken(&user), queries)
	}
}
