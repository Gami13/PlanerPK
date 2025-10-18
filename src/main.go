package main

import (
	"GoogleCalendarTimetable/src/timetableExtractor"
	"GoogleCalendarTimetable/src/utils"
	"GoogleCalendarTimetable/src/webServer"
)

func main() {
	config := utils.NewConfig()
	// googleCalendar.AuthorizeFromCode("")
	// timetable := timetableExtractor.GetTimetable(config, 3)

	// fmt.Println(timetable)
	_ = timetableExtractor.GetTimetable(config, 3)
	webServer.Start()
	// googleCalendar.Run()
}
