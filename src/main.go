package main

import (
	"GoogleCalendarTimetable/src/googleCalendar"
	"GoogleCalendarTimetable/src/timetableExtractor"
	"GoogleCalendarTimetable/src/utils"
	"GoogleCalendarTimetable/src/webServer"
	"fmt"
)

func main() {
	config := utils.NewConfig()

	timetable := timetableExtractor.GetTimetable(config)

	fmt.Println(timetable)
	go webServer.Start()
	googleCalendar.Run()
}
