package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/calendar"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/constants"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/database"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/downloader"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/webServer"
)

func main() {

	// Fetch the timetable url
	didChange, path, err := downloader.UpdateTimetable()
	if err != nil {
		fmt.Println(err)
		return
	}
	if didChange {
		fmt.Println("Timetable updated, new path: " + path)

	}

	db, err := sql.Open(constants.DATABASE_DRIVER, constants.DATABASE_SOURCE)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = database.InitializeDatabase(db)
	if err != nil {
		panic(err)
	}
	fmt.Println("Initialized database successfully.")

	queries := database.New(db)

	ticker := time.NewTicker(5 * time.Minute)
	done := make(chan bool)
	fmt.Println("Starting timetable update ticker (5 min interval)")
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:

				fmt.Println("Tick at", t)
				updateAllTimetables(queries)
			}
		}
	}()
	fmt.Println("Starting web server")
	webServer.Start(db)
}

func updateAllTimetables(queries *database.Queries) {
	users, err := queries.GetAllUsers()
	if err != nil {
		fmt.Println("Error fetching users:", err)
		return
	}
	didChange, _, err := downloader.UpdateTimetable()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !didChange {
		fmt.Println("No timetable changes detected; skipping updates.")
		return
	}

	for _, user := range users {

		if *user.ChosenColumn < 0 || !*user.IsActive {
			continue
		}

		fmt.Println("Updating timetable for user:", user.Email)
		calendar.ImportTimetable(*user.ChosenColumn, user.Token())
	}
}
