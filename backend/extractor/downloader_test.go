package extractor

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/types"
)

func TestDirectXlsxDownloadAndRefresh(t *testing.T) {
	previous := config.Config
	t.Cleanup(func() { config.Config = previous })
	config.Config.TimetableDownloadsDir = t.TempDir()
	config.Config.CurrentTimetableDir = t.TempDir()
	config.Config.TimetableStats = filepath.Join(t.TempDir(), "stats.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="Plan-ST-2026-2027.xlsx"`)
		w.Header().Set("Content-Length", "4")
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte("xlsx"))
		}
	}))
	defer server.Close()
	config.Config.TimetableURL = server.URL + "/download.php"
	stats, err := DownloadTimetable(config.Config.TimetableURL)
	if err != nil || filepath.Base(stats.Path) != "Plan-ST-2026-2027.xlsx" {
		t.Fatalf("download = %+v, %v", stats, err)
	}
	for range 2 {
		path, err := ProcessTimetable(stats)
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != "xlsx" {
			t.Fatalf("processed workbook = %q, %v", data, err)
		}
	}
	changed, url, err := DoesTimetableNeedUpdating()
	if err != nil || !changed || url != config.Config.TimetableURL {
		t.Fatalf("refresh without Last-Modified = %v, %q, %v", changed, url, err)
	}
}

func TestRejectXlsTimetable(t *testing.T) {
	previous := config.Config
	t.Cleanup(func() { config.Config = previous })
	config.Config.TimetableDownloadsDir = t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="old.xls"`)
	}))
	defer server.Close()
	if _, err := DownloadTimetable(server.URL + "/download.php"); err == nil || !strings.Contains(err.Error(), "unsupported timetable format") {
		t.Fatalf("XLS download error = %v", err)
	}
	entries, err := os.ReadDir(config.Config.TimetableDownloadsDir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("rejected download wrote files: %v, %v", entries, err)
	}
	if _, err := ProcessTimetable(types.TimetableFileStats{Path: "old.xls"}); err == nil || !strings.Contains(err.Error(), "unsupported timetable format") {
		t.Fatalf("XLS processing error = %v", err)
	}
}
