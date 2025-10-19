package main

import (
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/downloader"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/timetableExtractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/utils"
	"fmt"
)

func main() {
	config := utils.NewConfig()

	// Fetch the timetable url
	timetableUrl, err := downloader.FetchTimetableUrl()
	if err != nil {
		fmt.Println(err)
		return
	}

	timetablePath := downloader.TIMETABLE_DOWNLOAD_DESTINATION + "17102025_9_INFORMATYKA-zima-rozklad-STACJONARNE-2025_2026.xls" // TODO: Get currently local timetable path, if doesn't exist, download and skip to Xls2Xlsx
	didChange, err := downloader.DidChange(timetablePath, timetableUrl)

	if didChange {
		fmt.Println("Timetable changed, downloading a new timetable")

		// Download the new timetable
		timetablePath, err = downloader.DownloadTimetable(timetableUrl)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Convert xls to xlsx
		timetablePath, err = downloader.Xls2Xlsx(timetablePath, downloader.TIMETABLE_DOWNLOAD_DESTINATION)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("Timetable did not change")
		timetablePath = timetablePath + "x" // TODO: Get currently local timetable path that is converted to xlsx
	}

	timetable := timetableExtractor.GetTimetable(config, timetablePath)

	fmt.Println(timetable)
}
