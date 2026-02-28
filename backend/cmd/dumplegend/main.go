package main

import (
	"fmt"
	"os"

	"github.com/WoodPKrs/GoogleCalendarTimetable/extractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/global"
)

func main() {
	global.LoadEnvVariables()

	legend, err := extractor.GetLegend()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Teachers found: %d\n", len(legend.Teachers))
	for shorthand, t := range legend.Teachers {
		fmt.Printf("  %s -> %s (%s)\n", shorthand, t.FullName, t.Email)
	}

	fmt.Printf("\nLocations found: %d\n", len(legend.Locations))
	for shorthand, l := range legend.Locations {
		fmt.Printf("  %s -> %s\n", shorthand, l.FullName)
	}
}
