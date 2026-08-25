package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/extractor"
	"github.com/xuri/excelize/v2"
)

func main() {
	config.Load()

	xlsxPath, err := findXlsxFile(config.Config.CurrentTimetableDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding timetable file: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Opening: %s\n", xlsxPath)

	file, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	sheetName, err := extractor.ResolveTimetableSheet(file, time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving timetable sheet: %v\n", err)
		os.Exit(1)
	}

	layout, err := extractor.DiscoverTimetableLayout(file, sheetName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting timetable layout: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Found %d days\n", len(layout.Days))

	groups, err := extractor.BuildAvailableGroups(file, sheetName, layout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building timetable groups: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Found %d timetable group columns\n", len(groups))

	var classTexts []string

	dayNames := map[string]bool{
		"PONIEDZIAŁEK": true, "WTOREK": true, "ŚRODA": true,
		"CZWARTEK": true, "PIĄTEK": true, "SOBOTA": true, "NIEDZIELA": true,
	}

	for _, group := range groups {
		col := group.Column
		for _, day := range layout.Days {
			for row := day.StartingRow; row <= day.EndingRow; row++ {
				text := getCell(file, sheetName, row, col)
				if text != "" && !dayNames[strings.TrimSpace(text)] {
					classTexts = append(classTexts, text)
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Collected %d classText entries\n", len(classTexts))

	// Write to testdata file
	outDir := filepath.Join("testdata")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join(outDir, "classTexts.txt")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	for _, text := range classTexts {
		fmt.Fprintf(outFile, "%q\n", text)
	}

	fmt.Fprintf(os.Stderr, "Written to %s\n", outPath)
}

func getCell(file *excelize.File, sheet string, row int, col int) string {
	column, err := excelize.ColumnNumberToName(col + 1)
	if err != nil {
		return ""
	}
	cell, err := file.GetCellValue(sheet, fmt.Sprintf("%s%d", column, row+1))
	if err != nil {
		return ""
	}
	return cell
}

func findXlsxFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("reading directory %s: %w", dir, err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".xlsx") {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no .xlsx file found in %s", dir)
}
