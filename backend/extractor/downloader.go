package extractor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/html"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

var (
	ErrNoXlsxFile        = fmt.Errorf("no .xlsx file found in the destination folder")
	ErrTimetableNotFound = errors.New("timetable link for informatyka (stacjonarne) not found")
)

func isExcelURL(rawURL string) bool {
	cleanURL := strings.TrimSpace(rawURL)
	if cleanURL == "" {
		return false
	}
	target := cleanURL
	if parsed, err := url.Parse(cleanURL); err == nil && parsed.Path != "" {
		target = parsed.Path
	}
	lower := strings.ToLower(target)
	return strings.HasSuffix(lower, ".xls") || strings.HasSuffix(lower, ".xlsx")
}

func findTimetableLink(doc *html.Node, rawBaseURL string) (string, error) {
	if doc == nil {
		return "", ErrTimetableNotFound
	}

	base, _ := url.Parse(rawBaseURL)
	for _, a := range htmlquery.Find(doc, "//a[@href]") {
		href := strings.TrimSpace(htmlquery.SelectAttr(a, "href"))
		content := strings.ToLower(href + " " + htmlquery.InnerText(a) + " " + htmlquery.SelectAttr(a, "title"))

		if isExcelURL(href) && strings.Contains(content, "informatyka") && !strings.Contains(content, "niestacjonarne") {
			if u, err := url.Parse(href); err == nil && base != nil && rawBaseURL != "" {
				return base.ResolveReference(u).String(), nil
			}
			return href, nil
		}
	}

	return "", ErrTimetableNotFound
}

func fetchTimetableUrl() (string, error) {
	doc, err := htmlquery.LoadURL(config.Config.TimetableURL)
	if err != nil {
		return "", fmt.Errorf("failed to get the timetable url: %w", err)
	}
	link, err := findTimetableLink(doc, config.Config.TimetableURL)
	if err != nil {
		return "", fmt.Errorf("failed to get the timetable url: %w", err)
	}

	return link, nil
}

func getCurrentXlsxPath() (filePath string, err error) {
	entries, err := os.ReadDir(config.Config.CurrentTimetableDir)
	if err != nil {
		return "", err
	}
	if len(entries) < 1 {
		return "", ErrNoXlsxFile
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".xlsx") {
			return filepath.Join(config.Config.CurrentTimetableDir, entry.Name()), nil
		}
	}

	return "", ErrNoXlsxFile
}

func setLatestTimetableInfo(lastModified string, contentLength string, filePath string) (stats types.TimetableFileStats, err error) {
	stats = types.TimetableFileStats{
		LastModified:  lastModified,
		ContentLength: contentLength,
		Path:          filePath,
	}

	jsonString, err := json.Marshal(stats)
	if err != nil {
		return stats, err
	}

	err = os.WriteFile(config.Config.TimetableStats, jsonString, 0o644)
	if err != nil {
		return stats, err
	}
	return stats, nil
}

func getLatestTimetableInfo() (stats types.TimetableFileStats, err error) {
	jsonBytes, err := os.ReadFile(config.Config.TimetableStats)
	if err != nil {
		return types.TimetableFileStats{}, err
	}

	err = json.Unmarshal(jsonBytes, &stats)
	if err != nil {
		return types.TimetableFileStats{}, err
	}
	return stats, nil
}

func DoesTimetableNeedUpdating() (hasChanged bool, fileURL string, err error) {
	remoteTimetableUrl, err := fetchTimetableUrl()
	if err != nil {
		return false, "", err
	}

	response, err := httpClient.Head(remoteTimetableUrl)
	if err != nil {
		return false, "", err
	}
	defer logCloseWarn("response body", response.Body.Close())

	if response.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("HTTP %d: %s", response.StatusCode, response.Status)
	}

	localFileInfo, err := getLatestTimetableInfo()
	if err != nil {
		return false, "", err
	}

	remoteLastModified := response.Header.Get("Last-Modified")
	remoteContentLength := response.Header.Get("Content-Length")
	if localFileInfo.LastModified != remoteLastModified || localFileInfo.ContentLength != remoteContentLength {
		slog.Info("Last modified", "local", localFileInfo.LastModified, "remote", remoteLastModified)
		slog.Info("Content length", "local", localFileInfo.ContentLength, "remote", remoteContentLength)
		return true, remoteTimetableUrl, nil
	}

	return false, localFileInfo.Path, nil
}

// Downloads the timetable from PKs website
// Returns types.TimetableFileStats
func DownloadTimetable(url string) (stats types.TimetableFileStats, err error) {
	destination := filepath.Join(config.Config.TimetableDownloadsDir, filepath.Base(url))

	response, err := httpClient.Get(url)
	if err != nil {
		return types.TimetableFileStats{}, fmt.Errorf("failed to download the timetable: %w", err)
	}
	defer func() {
		logCloseWarn("response body", response.Body.Close())
	}()
	if response.StatusCode != http.StatusOK {
		return types.TimetableFileStats{}, fmt.Errorf("HTTP %d: %s", response.StatusCode, response.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return types.TimetableFileStats{}, fmt.Errorf("failed to create the timetable file: %w", err)
	}
	defer func() {
		logCloseWarn("timetable file", file.Close())
	}()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return types.TimetableFileStats{}, fmt.Errorf("failed to save the timetable file: %w", err)
	}

	lastModified := response.Header.Get("Last-Modified")
	contentLength := response.Header.Get("Content-Length")

	stats, err = setLatestTimetableInfo(lastModified, contentLength, destination)
	if err != nil {
		return stats, err
	}
	return stats, nil
}

func ProcessTimetable(stats types.TimetableFileStats) (newFilePath string, err error) {
	origin := stats.Path

	timetableDestination := config.Config.CurrentTimetableDir
	doesExist := true
	currentXlsxPath, err := getCurrentXlsxPath()
	if err != nil {
		if errors.Is(err, ErrNoXlsxFile) {
			doesExist = false
		} else {
			return "", err
		}
	}
	newXlsxPath, err := xls2xlsx(origin, timetableDestination)
	if err != nil {
		return "", err
	}
	if doesExist {

		err = os.Remove(currentXlsxPath)
		if err != nil {
			// !Hopefully neva eva happens
			// TODO: figure out how to handle not removing the old file if the new one was also created
			slog.Error("Failed to remove old timetable file", "error", err)
			return newXlsxPath, err
		}
	}
	return newXlsxPath, nil
}

// Convert an .xls file to an .xlsx file
// Returns the path to the .xlsx file
func xls2xlsx(src string, dest string) (string, error) {
	// Windows
	if runtime.GOOS == "windows" {
		cmd := exec.Command("scalc", "--headless", "--convert-to", "xlsx", "--outdir", dest, src)

		return filepath.Join(dest, filepath.Base(src)+"x"), cmd.Run()
	}
	// Linux
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "xlsx", "--outdir", dest, src)
	return filepath.Join(dest, filepath.Base(src)+"x"), cmd.Run()
}

func OpenCurrentTimetableFile() (*excelize.File, error) {
	timetablePath, err := getCurrentXlsxPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get current timetable: %w", err)
	}

	file, err := excelize.OpenFile(timetablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open timetable file: %w", err)
	}

	return file, nil
}
