package global

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseEnvironmentDefaults(t *testing.T) {
	cfg, err := parseEnvironment(map[string]string{})
	if err != nil {
		t.Fatalf("parseEnvironment() error = %v", err)
	}

	wantDataDir, err := filepath.Abs("./data")
	if err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}
	if cfg.DataDir != wantDataDir {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, wantDataDir)
	}
	if cfg.HourColumn != 1 || cfg.LastColumn != 40 {
		t.Errorf("unexpected timetable defaults: %+v", cfg)
	}
	if cfg.TimetableUpdatePeriod != time.Minute || cfg.TimetableSyncPeriod != 72*time.Hour {
		t.Errorf("unexpected intervals: update=%s sync=%s", cfg.TimetableUpdatePeriod, cfg.TimetableSyncPeriod)
	}
	if cfg.DataDir != wantDataDir {
		t.Errorf("DataDir = %q", cfg.DataDir)
	}
	if cfg.StaticDir != "static" {
		t.Errorf("StaticDir = %q", cfg.StaticDir)
	}
	if cfg.GooglePlacesAPIKey != "" {
		t.Error("Google Places should be unconfigured without an API key")
	}
}

func TestParseEnvironmentOverridesAndDerivedPaths(t *testing.T) {
	dataDir := t.TempDir()
	cfg, err := parseEnvironment(map[string]string{
		"DATA_DIR":                dataDir,
		"HOUR_COLUMN":             "2",
		"DAY_START":               "8.15",
		"DAY_END":                 "20.45",
		"LAST_COLUMN":             "41",
		"MINUTES_BETWEEN":         "5",
		"TIME_BETWEEN_SYNC_HOURS": "12",
		"TIMETABLE_URL":           "https://example.test/timetable",
		"GOOGLE_PLACES_API_KEY":   "test-key",
		"LOG_LEVEL":               "warn",
	})
	if err != nil {
		t.Fatalf("parseEnvironment() error = %v", err)
	}

	if cfg.CalendarCredentials != filepath.Join(dataDir, "credentials.json") {
		t.Errorf("CalendarCredentials = %q", cfg.CalendarCredentials)
	}
	if !strings.Contains(cfg.DatabaseSource, filepath.ToSlash(filepath.Join(dataDir, "state", "database.db"))) {
		t.Errorf("DatabaseSource = %q", cfg.DatabaseSource)
	}
	if cfg.DayStartTime.Hour() != 8 || cfg.DayStartTime.Minute() != 15 {
		t.Errorf("DayStartTime = %v", cfg.DayStartTime)
	}
	if cfg.HourColumn != 2 || cfg.LastColumn != 41 {
		t.Errorf("unexpected timetable overrides: %+v", cfg)
	}
	if cfg.LogLevel != slog.LevelWarn {
		t.Errorf("LogLevel = %s, want WARN", cfg.LogLevel)
	}
	if cfg.TimetableUpdatePeriod != 5*time.Minute || cfg.TimetableSyncPeriod != 12*time.Hour {
		t.Errorf("unexpected intervals: update=%s sync=%s", cfg.TimetableUpdatePeriod, cfg.TimetableSyncPeriod)
	}
	if cfg.GooglePlacesAPIKey == "" {
		t.Error("Google Places should be configured")
	}
}

func TestParseEnvironmentRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		environment map[string]string
		wantError   string
	}{
		{"malformed integer", map[string]string{"HOUR_COLUMN": "x"}, "parse environment"},
		{"invalid start clock", map[string]string{"DAY_START": "08:15"}, "DAY_START"},
		{"invalid end clock", map[string]string{"DAY_END": "noon"}, "DAY_END"},
		{"invalid log level", map[string]string{"LOG_LEVEL": "verbose"}, "LogLevel"},
		{"negative hour column", map[string]string{"HOUR_COLUMN": "-1"}, "HOUR_COLUMN"},
		{"last column before hour column", map[string]string{"HOUR_COLUMN": "4", "LAST_COLUMN": "4"}, "LAST_COLUMN"},
		{"zero update interval", map[string]string{"MINUTES_BETWEEN": "0"}, "MINUTES_BETWEEN"},
		{"negative sync interval", map[string]string{"TIME_BETWEEN_SYNC_HOURS": "-1"}, "TIME_BETWEEN_SYNC_HOURS"},
		{"empty data directory", map[string]string{"DATA_DIR": ""}, "DATA_DIR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseEnvironment(test.environment)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("parseEnvironment() error = %v, want error containing %q", err, test.wantError)
			}
		})
	}
}

func TestPrepareFilesystemCreatesOnlyMutableDirectories(t *testing.T) {
	dataDir := t.TempDir()
	cfg, err := parseEnvironment(map[string]string{"DATA_DIR": dataDir})
	if err != nil {
		t.Fatalf("parseEnvironment() error = %v", err)
	}

	if err := PrepareFilesystem(cfg); err != nil {
		t.Fatalf("PrepareFilesystem() error = %v", err)
	}
	for _, path := range []string{
		cfg.StateDir,
		cfg.LogsDir,
		cfg.TimetableDownloadsDir,
		cfg.CurrentTimetableDir,
	} {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Errorf("expected runtime directory %q to exist, err=%v", path, err)
		}
	}
	if _, err := os.Stat(cfg.StaticDir); !os.IsNotExist(err) {
		t.Errorf("static directory should not be created, err=%v", err)
	}
}

func TestPrepareFilesystemReportsErrors(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(dataDir, []byte("file"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg, err := parseEnvironment(map[string]string{"DATA_DIR": dataDir})
	if err != nil {
		t.Fatalf("parseEnvironment() error = %v", err)
	}

	if err := PrepareFilesystem(cfg); err == nil {
		t.Fatal("PrepareFilesystem() error = nil, want an error")
	}
}
