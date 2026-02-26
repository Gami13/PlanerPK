package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/WoodPKrs/GoogleCalendarTimetable/global"
	"github.com/xuri/excelize/v2"
)

func main() {
	global.LoadEnvVariables()

	xlsxPath, err := findXlsxFile(global.TIMETABLE_CURRENT_DESTINATION)
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

	days := findDays(file)
	fmt.Fprintf(os.Stderr, "Found %d days\n", len(days))

	// Find hour-reference columns to skip (header says "Godziny")
	skipCols := findHourColumns(file, days)
	fmt.Fprintf(os.Stderr, "Skipping %d hour-reference columns\n", len(skipCols))

	var classTexts []string

	dayNames := map[string]bool{
		"PONIEDZIAŁEK": true, "WTOREK": true, "ŚRODA": true,
		"CZWARTEK": true, "PIĄTEK": true, "SOBOTA": true, "NIEDZIELA": true,
	}

	for col := global.HOUR_COLUMN + 1; col <= global.LAST_COLUMN; col++ {
		if skipCols[col] {
			continue
		}
		for _, day := range days {
			for row := day.startRow; row <= day.endRow; row++ {
				text := getCell(file, global.SHEET_NAME, row, col)
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

type dayRange struct {
	startRow int
	endRow   int
}

func findDays(file *excelize.File) []dayRange {
	var days []dayRange
	rows, err := file.GetRows(global.SHEET_NAME)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting rows: %v\n", err)
		return days
	}
	for i, row := range rows {
		if len(row) > global.HOUR_COLUMN {
			if row[global.HOUR_COLUMN] == global.DAY_START {
				days = append(days, dayRange{startRow: i})
			}
			if row[global.HOUR_COLUMN] == global.DAY_END && len(days) > 0 {
				days[len(days)-1].endRow = i
			}
		}
	}
	return days
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

// findHourColumns identifies columns whose header row contains "Godziny" — these
// are hour-reference columns, not class data.
func findHourColumns(file *excelize.File, days []dayRange) map[int]bool {
	skip := make(map[int]bool)
	if len(days) == 0 {
		return skip
	}
	// Check a few rows above the first day's start for "Godziny"
	headerSearch := max(days[0].startRow-5, 0)
	for col := global.HOUR_COLUMN + 1; col <= global.LAST_COLUMN; col++ {
		for row := headerSearch; row < days[0].startRow; row++ {
			val := strings.TrimSpace(getCell(file, global.SHEET_NAME, row, col))
			if strings.EqualFold(val, "Godziny") {
				skip[col] = true
				break
			}
		}
	}
	return skip
}
