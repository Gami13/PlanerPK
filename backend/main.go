package main

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/calendar"
	"github.com/WoodPKrs/GoogleCalendarTimetable/database"
	"github.com/WoodPKrs/GoogleCalendarTimetable/downloader"
	"github.com/WoodPKrs/GoogleCalendarTimetable/global"

	"github.com/WoodPKrs/GoogleCalendarTimetable/utils"
	"github.com/WoodPKrs/GoogleCalendarTimetable/webServer"
)

func main() {
	// Initialize logger first
	if err := utils.InitializeLogging(global.LOGS_PATH); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	utils.LoadEnvVariables()
	slog.Info("Environment variables loaded")

	// Fetch the timetable url
	_, _, err := downloader.UpdateTimetable()
	if err != nil {
		slog.Error("Failed to update timetable", slog.Any("error", err))
	}

	db, err := sql.Open(global.DATABASE_DRIVER, global.DATABASE_SOURCE)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", slog.Any("error", err))
		}
	}()

	err = database.InitializeDatabase(db)
	if err != nil {
		slog.Error("Failed to initialize database", slog.Any("error", err))
		panic(err)
	}
	slog.Info("Database initialized successfully")

	queries := database.New(db)

	ticker := time.NewTicker(time.Duration(global.MINUTES_BETWEEN) * time.Minute)
	done := make(chan bool)
	slog.Info("Starting web server", slog.String("port", global.SERVER_PORT))
	go webServer.Start(db)
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

func updateAllTimetables(queries *database.Queries) {
	users, err := queries.GetAllUsers(context.Background())
	if err != nil {
		slog.Error("Failed to fetch users", slog.Any("error", err))
		return
	}
	didChange, _, err := downloader.UpdateTimetable()
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
