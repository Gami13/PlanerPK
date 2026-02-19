package extractor

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/WoodPKrs/GoogleCalendarTimetable/global"
	"github.com/xuri/excelize/v2"
)

func logCloseWarn(resource string, err error) {
	if err != nil {
		slog.Warn("Failed to close resource",
			slog.String("resource", resource),
			slog.Any("error", err))
	}
}

func moveFileOrWarn(src, dst, label string) {
	if err := os.Rename(src, dst); err != nil {
		slog.Warn("Failed to move file",
			slog.String("label", label),
			slog.String("source", src),
			slog.String("destination", dst),
			slog.Any("error", err))
	}
}

func removeFileOrWarn(path, label string) {
	if err := os.Remove(path); err != nil {
		slog.Warn("Failed to remove file",
			slog.String("label", label),
			slog.String("path", path),
			slog.Any("error", err))
	}
}

func openCurrentTimetableFile() (*excelize.File, error) {
	timetablePath, err := GetCurrentXlsxTimetablePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get current timetable: %w", err)
	}

	file, err := excelize.OpenFile(timetablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open timetable file: %w", err)
	}

	return file, nil
}

func closeFile(file *excelize.File) {
	if err := file.Close(); err != nil {
		slog.Warn("Failed to close file", slog.Any("error", err))
	}
}

func GetCurrentXlsxTimetablePath() (string, error) {
	timetablePath, err := getCurrentTimetablePath()
	if err != nil {
		return "", err
	}
	return timetablePath + "x", nil
}

// Get the currently used .xls timetable (under the CURRENT_TIMETABLE_DESTINATION folder)
func getCurrentTimetablePath() (string, error) {
	entries, err := os.ReadDir(global.TIMETABLE_CURRENT_DESTINATION)
	if err != nil {
		return "", err
	}

	// If there are multiple .xls files then get the newest one by date and warn about it
	var xlsFiles []fs.FileInfo

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".xls") {
			fileInfo, err := entry.Info()
			if err != nil {
				slog.Error("Skipping possible current timetable file.",
					slog.String("entry", entry.Name()),
					slog.Any("error", err))
			}

			xlsFiles = append(xlsFiles, fileInfo)
		}
	}

	if len(xlsFiles) == 1 {
		return global.TIMETABLE_CURRENT_DESTINATION + xlsFiles[0].Name(), nil
	} else if len(xlsFiles) > 1 {
		slog.Warn("Multiple current timetable files. Getting the newest one by modification date.",
			slog.String("timetable_current_destination", global.TIMETABLE_CURRENT_DESTINATION))

		sort.Slice(xlsFiles, func(i, j int) bool {
			return xlsFiles[i].ModTime().After(xlsFiles[j].ModTime())
		})

		return global.TIMETABLE_CURRENT_DESTINATION + xlsFiles[0].Name(), nil
	}

	return "", fmt.Errorf("no .xls file found in %s", global.TIMETABLE_CURRENT_DESTINATION)
}
