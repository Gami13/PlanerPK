package main

import (
	"fmt"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/downloader"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/extractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/webServer"
)

func main() {

	// Fetch the timetable url
	timetableUrl, err := downloader.FetchTimetableUrl()
	if err != nil {
		fmt.Println(err)
		return
	}

	// If doesn't exist, download
	timetablePath, err := extractor.GetCurrentTimetable()
	if err != nil {
		fmt.Println(err)
		fmt.Println("No timetable present")

		timetablePath, err = downloader.DownloadAndProcessTimetable(timetableUrl)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		didChange, err := downloader.DidChange(timetablePath, timetableUrl)
		if err != nil {
			fmt.Println(err)
			return
		}

		if didChange {
			fmt.Println("Timetable changed, downloading a new timetable")
			timetablePath, err = downloader.DownloadAndProcessTimetable(timetableUrl)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			fmt.Println("Timetable did not change")
			timetablePath = timetablePath + "x"
		}
	}

	// timetable := extractor.GetTimetable(3)

	// fmt.Println(timetable)
	webServer.Start()
	// googleCalendar.Run()
}
