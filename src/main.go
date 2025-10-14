package main

import (
	timetableExtractor "GoogleCalendarTimetable/src/TimetableExtractor"
	"GoogleCalendarTimetable/src/utils"
	"fmt"
)

func main() {
	config := utils.NewConfig()

	timetable := timetableExtractor.GetTimetable(config)

	fmt.Println(timetable)
}
