package config

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
		"MINUTES_BETWEEN":         "5",
		"TIME_BETWEEN_SYNC_HOURS": "12",
		"TIMETABLE_URL":           "https://example.test/timetable",
		"GOOGLE_PLACES_API_KEY":   "test-key",
		"DISCORD_WEBHOOK_URL":     "https://discord.example/webhook",
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
	if cfg.LogLevel != slog.LevelWarn {
		t.Errorf("LogLevel = %s, want WARN", cfg.LogLevel)
	}
	if cfg.TimetableUpdatePeriod != 5*time.Minute || cfg.TimetableSyncPeriod != 12*time.Hour {
		t.Errorf("unexpected intervals: update=%s sync=%s", cfg.TimetableUpdatePeriod, cfg.TimetableSyncPeriod)
	}
	if cfg.GooglePlacesAPIKey == "" {
		t.Error("Google Places should be configured")
	}
	if cfg.DiscordWebhookURL != "https://discord.example/webhook" {
		t.Errorf("DiscordWebhookURL = %q", cfg.DiscordWebhookURL)
	}
}

func TestParseEnvironmentRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		environment map[string]string
		wantError   string
	}{
		{"invalid log level", map[string]string{"LOG_LEVEL": "verbose"}, "LogLevel"},
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

func TestParseEnvironmentIgnoresObsoleteTimetableLayoutVariables(t *testing.T) {
	_, err := parseEnvironment(map[string]string{
		"HOUR_COLUMN": "not-an-integer",
		"DAY_START":   "not-a-time",
		"DAY_END":     "not-a-time",
		"LAST_COLUMN": "not-an-integer",
	})
	if err != nil {
		t.Fatalf("parseEnvironment() error = %v", err)
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
