package main

import (
	"database/sql"
	"fmt"

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
	webServer.Start(db)
}
