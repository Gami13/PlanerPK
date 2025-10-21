package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/WoodPKrs/GoogleCalendarTimetable/src/timetableExtractor"
)

const TIMETABLE_URL = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
const TIMETABLE_DOWNLOAD_DESTINATION = "./downloads/"

// Get the url of the timetable from the PKs website
func FetchTimetableUrl() (string, error) {
	doc, err := htmlquery.LoadURL(TIMETABLE_URL)

	if err != nil {
		return "", fmt.Errorf("failed to get the timetable url: %w", err)
	}

	link := htmlquery.InnerText(htmlquery.FindOne(doc, "//html/body/div[2]/div[2]/div/div/main/article/div/div[1]/a/@href"))

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
	destination, err := filepath.Abs(TIMETABLE_DOWNLOAD_DESTINATION + filepath.Base(url))
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
	// Download the new timetable
	timetablePathXls, err := DownloadTimetable(url)
	if err != nil {
		return "", err
	}

	// Convert .xls to .xlsx
	timetablePathXlsx, err := Xls2Xlsx(timetablePathXls, TIMETABLE_DOWNLOAD_DESTINATION)
	if err != nil {
		return "", err
	}

	// Move the timetable to the current timetable destination
	os.Rename(timetablePathXls, timetableExtractor.CURRENT_TIMETABLE_DESTINATION + filepath.Base(timetablePathXls))
	os.Rename(timetablePathXlsx, timetableExtractor.CURRENT_TIMETABLE_DESTINATION + filepath.Base(timetablePathXlsx))
	return timetableExtractor.CURRENT_TIMETABLE_DESTINATION + filepath.Base(timetablePathXlsx), nil
}
