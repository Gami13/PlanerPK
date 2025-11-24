package downloader

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/src/extractor"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/global"
	"github.com/antchfx/htmlquery"
)

// Get the url of the timetable from the PKs website
func FetchTimetableUrl() (string, error) {
	doc, err := htmlquery.LoadURL(global.TIMETABLE_URL)

	if err != nil {
		return "", fmt.Errorf("failed to get the timetable url: %w", err)
	}
	link := htmlquery.InnerText(htmlquery.FindOne(doc, "//html/body/div[2]/div[2]/div/div/main/article/div/ul[1]/li[1]/a/@href"))

	return link, nil
}

// Check if the timetable changed compared to the PKs timetable
func DidChange(timetablePath string, url string) (bool, error) {
	localName, err := filepath.Abs(timetablePath)
	if err != nil {
		return false, nil
	}

	remoteName, err := filepath.Abs(timetablePath)
	if err != nil {
		return false, nil
	}

	if localName != remoteName && localName != remoteName+"x" {
		return true, nil
	}

	resp, err := http.Head(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	remoteLastModified := resp.Header.Get("Last-Modified")
	remoteContentLength := resp.Header.Get("Content-Length")

	localFileInfo, err := os.Stat(timetablePath)
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	localLastModified := localFileInfo.ModTime().UTC().Format(http.TimeFormat)
	localContentLength := fmt.Sprintf("%d", localFileInfo.Size())

	if localLastModified != remoteLastModified || localContentLength != remoteContentLength {
		return true, nil
	}

	return false, nil
}

// Downloads the timetable from PKs website
// Returns filepath.Abs(TIMETABLE_DOWNLOAD_DESTINATION + filepath.Base(url))
func DownloadTimetable(url string) (string, error) {
	// Get the absolute path to the destination
	destination, err := filepath.Abs(global.TIMETABLE_DOWNLOAD_DESTINATION + filepath.Base(url))
	if err != nil {
		return destination, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return destination, fmt.Errorf("failed to download the timetable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return destination, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	lastModified := resp.Header.Get("Last-Modified")
	lastModifiedTime, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		return destination, fmt.Errorf("failed to parse the last modified time: %w", err)
	}

	file, err := os.Create(destination)
	if err != nil {
		return destination, fmt.Errorf("failed to create the timetable file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return destination, fmt.Errorf("failed to save the timetable file: %w", err)
	}

	// Set properties so they are the same as the one from the download link for later comparison of possible changes
	os.Chtimes(destination, lastModifiedTime, lastModifiedTime)

	return destination, err
}

// Convert an .xls file to an .xlsx file
// Returns the path to the .xlsx file
func Xls2Xlsx(src string, dest string) (string, error) {
	if os.PathSeparator == '\\' {

		cmd := exec.Command("scalc", "--headless", "--convert-to", "xlsx", "--outdir", dest, src)
		return dest + filepath.Base(src) + "x", cmd.Run()
	}

	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "xlsx", "--outdir", dest, src)
	return dest + filepath.Base(src) + "x", cmd.Run()
}

// Download a timetable from the PK website, convert it to .xlsx and move to the timetableExtractor.CURRENT_TIMETABLE_DESTINATION
// Returns the path timetableExtractor.CURRENT_TIMETABLE_DESTINATION + filepath.Base(timetablePathXlsx)
func DownloadAndProcessTimetable(url string) (string, error) {
	// Get the current timetable that will be removed after a successfull download
	currentTimetablePathXls, err := extractor.GetCurrentTimetablePath();

	// Download the new timetable
	timetablePathXls, err := DownloadTimetable(url)
	if err != nil {
		return "", err
	}

	// Convert .xls to .xlsx
	timetablePathXlsx, err := Xls2Xlsx(timetablePathXls, global.TIMETABLE_DOWNLOAD_DESTINATION)
	if err != nil {
		return "", err
	}

	// Move the timetable to the current timetable destination
	newTimetablePathXls := global.TIMETABLE_CURRENT_DESTINATION+filepath.Base(timetablePathXls)
	err = os.Rename(timetablePathXls, newTimetablePathXls)
	if err != nil {
		slog.Error("Failed to move the new .xls timetable.",
			slog.String("source", timetablePathXls),
			slog.String("destination", newTimetablePathXls),
			slog.Any("error", err))
	}

	newTimetablePathXlsx := global.TIMETABLE_CURRENT_DESTINATION+filepath.Base(timetablePathXlsx)
	err = os.Rename(timetablePathXlsx, newTimetablePathXlsx)
	if err != nil {
		slog.Error("Failed to move the new .xlsx timetable.",
			slog.String("source", timetablePathXlsx),
			slog.String("destination", newTimetablePathXlsx),
			slog.Any("error", err))
	}

	// Delete the old timetable
	err = os.Remove(currentTimetablePathXls)
	if err != nil {
		slog.Error("Failed to delete the old .xls timetable.",
			slog.String("file", currentTimetablePathXls),
			slog.Any("error", err))
	}

	currentTimetablePathXlsx := currentTimetablePathXls + "x"
	err = os.Remove(currentTimetablePathXlsx)
	if err != nil {
		slog.Error("Failed to delete the old .xlsx timetable.",
			slog.String("file", currentTimetablePathXlsx),
			slog.Any("error", err))
	}

	return newTimetablePathXlsx, nil
}

// Updates the timetable if it changed, using FetchTimetableUrl, DidChange and DownloadAndProcessTimetable
// Returns (didChange bool, timetablePath string, err error)
func UpdateTimetable() (didChange bool, path string, err error) {
	timetableUrl, err := FetchTimetableUrl()
	if err != nil {
		slog.Error("Failed to fetch timetable URL", slog.Any("error", err))
		return false, "", err
	}

	slog.Debug("Fetched timetable URL", slog.String("url", timetableUrl))

	timetablePath, err := extractor.GetCurrentTimetablePath()
	if err != nil {
		slog.Info("No timetable present, downloading new one")

		timetablePath, err = DownloadAndProcessTimetable(timetableUrl)
		if err != nil {
			slog.Error("Failed to download and process timetable", slog.Any("error", err))
			return false, "", err
		}
		slog.Info("Downloaded initial timetable", slog.String("path", timetablePath))
		return true, timetablePath, nil
	}

	didChange, err = DidChange(timetablePath, timetableUrl)
	if err != nil {
		slog.Error("Failed to check if timetable changed", slog.Any("error", err))
		return didChange, timetablePath, err
	}

	if didChange {
		slog.Info("Timetable changed, downloading new version")
		timetablePath, err = DownloadAndProcessTimetable(timetableUrl)
		if err != nil {
			slog.Error("Failed to download and process timetable", slog.Any("error", err))
			return didChange, timetablePath, err
		}
		slog.Info("Downloaded updated timetable", slog.String("path", timetablePath))
		return didChange, timetablePath, nil
	}
	slog.Info("No timetable changes detected")
	return didChange, timetablePath + "x", nil
}
