package main

import (
	"fmt"
	"os"

	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/extractor"
)

func main() {
	config.Load()

	file, err := extractor.OpenCurrentTimetableFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	sheetName, err := extractor.ResolveTimetableSheet(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving timetable sheet: %v\n", err)
		os.Exit(1)
	}
	legend := extractor.ExtractLegend(file, sheetName)

	fmt.Printf("Teachers found: %d\n", len(legend.Teachers))
	for shorthand, t := range legend.Teachers {
		fmt.Printf("  %s -> %s (%s)\n", shorthand, t.FullName, t.Email)
	}

	fmt.Printf("\nLocations found: %d\n", len(legend.Locations))
	for shorthand, l := range legend.Locations {
		fmt.Printf("  %s -> %s\n", shorthand, l.FullName)
	}
}
