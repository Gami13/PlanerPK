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
)

const TIMETABLE_URL = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
const TIMETABLE_DOWNLOAD_DESTINATION = "./downloads/"

func FetchTimetableUrl() (string, error) {
	doc, err := htmlquery.LoadURL(TIMETABLE_URL)

	if err != nil {
		return "", fmt.Errorf("failed to get the timetable url: %w", err)
	}

	link := htmlquery.InnerText(htmlquery.FindOne(doc, "//html/body/div[2]/div[2]/div/div/main/article/div/div[1]/a/@href"))

	return link, nil
}

func DidChange(timetablePath string, url string) (bool, error) {
	localName, err := filepath.Abs(timetablePath)
	if err != nil {
		return false, nil
	}

	remoteName, err := filepath.Abs(timetablePath)
	if err != nil {
		return false, nil
	}

	if localName != remoteName && localName != remoteName + "x" {
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

func DownloadTimetable(url string) (string, error) {
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

	os.Chtimes(destination, lastModifiedTime, lastModifiedTime)

	return destination, err
}

func Xls2Xlsx(src string, dest string) (string, error) {
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "xlsx", "--outdir", dest, src)
	return dest + filepath.Base(src) + "x", cmd.Run()
}
