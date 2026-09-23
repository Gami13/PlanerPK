// professionally slopped up
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/extractor"
	"github.com/xuri/excelize/v2"
)

func main() {
	config.Load()

	xlsxPath, err := findCurrentTimetableFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Processing timetable: %s\n", xlsxPath)
	file, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", xlsxPath, err)
		os.Exit(1)
	}
	defer file.Close()
	sheetName, err := extractor.ResolveTimetableSheet(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving timetable sheet: %v\n", err)
		os.Exit(1)
	}
	layout, err := extractor.DiscoverTimetableLayout(file, sheetName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting timetable layout: %v\n", err)
		os.Exit(1)
	}

	uniqueClassrooms := make(map[string]struct{})

	// A. Extract from Legend section ("Sale")
	legend := extractor.ExtractLegend(file, sheetName)
	buildings := buildings.BuildBuildingsFromLegend(legend)
	if legend != nil {
		for _, loc := range legend.Locations {
			if loc.Shorthand != "" {
				addClassroom(uniqueClassrooms, loc.Shorthand)
			}
		}
	}

	// B. Extract from timetable grid classes
	groups, err := extractor.BuildAvailableGroups(file, sheetName, layout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building groups for %s: %v\n", xlsxPath, err)
	} else {
		timetable := extractor.BuildTimetable(groups, file, sheetName, legend, buildings, layout)
		for _, groupTimetable := range timetable {
			for _, classes := range groupTimetable {
				for _, class := range classes {
					if class.Location != "" {
						addClassroom(uniqueClassrooms, class.Location)
					}
				}
			}
		}
	}

	// Sort unique classroom names
	classrooms := make([]string, 0, len(uniqueClassrooms))
	for name := range uniqueClassrooms {
		classrooms = append(classrooms, name)
	}
	sort.Strings(classrooms)

	fmt.Fprintf(os.Stderr, "Found %d unique classroom names in current timetable.\n", len(classrooms))

	// Output results to stdout
	for _, name := range classrooms {
		fmt.Println(name)
	}
}

// addClassroom cleans raw location strings and adds unique classroom names
// without splitting by slashes or matching against the legend.
func addClassroom(unique map[string]struct{}, raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}

	// Split by newlines or semicolons if multiple entries are present
	lines := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '\n' || r == ';'
	})

	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		cleaned = strings.Trim(cleaned, " ,;\t\r\n")
		if cleaned != "" {
			unique[cleaned] = struct{}{}
		}
	}
}

// findCurrentTimetableFile locates the single .xlsx timetable file in the current_timetable directory.
func findCurrentTimetableFile() (string, error) {
	dirsToSearch := []string{config.Config.CurrentTimetableDir}

	seen := make(map[string]bool)
	for _, dir := range dirsToSearch {
		if dir == "" {
			continue
		}
		abs, err := filepath.Abs(dir)
		if err != nil {
			abs = dir
		}
		if seen[abs] {
			continue
		}
		seen[abs] = true

		if entries, err := os.ReadDir(abs); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".xlsx") {
					return filepath.Join(abs, entry.Name()), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no .xlsx file found in current_timetable directory")
}
