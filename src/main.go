package main

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/calendar"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/database"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/downloader"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/global"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/utils"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/webServer"
)

func main() {
	// Initialize logger first
	if err := utils.InitializeLogging("./logs"); err != nil {
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
	defer db.Close()

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
	webServer.Start(db)
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
	users, err := queries.GetAllUsers()
	if err != nil {
		slog.Error("Failed to fetch users", slog.Any("error", err))
		return
	}
	didChange, _, err := downloader.UpdateTimetable()
	if err != nil {
		slog.Error("Failed to update timetable", slog.Any("error", err))
		return
	}
	if !didChange {
		slog.Info("No timetable changes detected, skipping updates")
		return
	}

	slog.Info("Timetable changed, updating for all active users",
		slog.Int("total_users", len(users)))

	for _, user := range users {
		if *user.ChosenColumn < 0 || !*user.IsActive {
			continue
		}

		slog.Info("Updating timetable for user",
			slog.String("email", *user.Email),
			slog.Int("column", *user.ChosenColumn))
		calendar.ImportTimetable(*user.ChosenColumn, user.Token())
	}
}
