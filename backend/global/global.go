package global

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	defaultDataDir   = "/data"
	defaultStaticDir = "/app/static"
)

var (
	SHEET_NAME              = "2025_2026"
	HOUR_COLUMN             = 1
	DAY_START               = "7.30"
	DAY_START_TIME          = time.Date(0, 0, 0, 7, 30, 0, 0, time.UTC)
	DAY_END                 = "21.00"
	LAST_COLUMN             = 40
	DATABASE_DRIVER         = "sqlite"
	DATABASE_ARGUMENTS      = "cache=shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite"
	MINUTES_BETWEEN         = 1
	// Fixed container port. Change the host-facing port through Docker/Compose.
	SERVER_PORT             = "8080"
	TIMETABLE_URL           = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
	TIME_BETWEEN_SYNC_HOURS = 72
	GOOGLE_PLACES_API_KEY   = "PLEASE_SET_YOUR_GOOGLE_PLACES_API_KEY"
	LOG_LEVEL               = "DEBUG" // OPTIONS: DEBUG, INFO, WARN, ERROR

	// DATA_DIR is the only configurable root for mutable application data.
	DATA_DIR   = defaultDataDir
	STATE_DIR  string
	LOGS_DIR   string
	STATIC_DIR = defaultStaticDir

	DATABASE_PATH                string
	DATABASE_SOURCE              string
	TIMETABLE_DOWNLOADS_DIR      string
	CURRENT_TIMETABLE_DIR        string
	CALENDAR_CREDENTIALS_PATH    string
	CURRENT_TIMETABLE_STATS_PATH string
)

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		slog.Warn("Invalid integer environment variable; using default",
			slog.String("variable", key),
			slog.String("value", value),
			slog.Int("default", fallback),
		)
		return fallback
	}
	return parsed
}

func getClockEnv(key, fallback string) (string, time.Time) {
	value := getEnvOrDefault(key, fallback)
	parsed, err := time.Parse("15.04", value)
	if err != nil {
		slog.Warn("Invalid time environment variable; using default",
			slog.String("variable", key),
			slog.String("value", value),
			slog.String("default", fallback),
		)
		value = fallback
		parsed, _ = time.Parse("15.04", fallback)
	}
	return value, parsed
}

func ensureDirectories() {
	for _, directory := range []string{
		STATE_DIR,
		LOGS_DIR,
		TIMETABLE_DOWNLOADS_DIR,
		CURRENT_TIMETABLE_DIR,
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			slog.Error("Failed to create data directory",
				slog.String("directory", directory),
				slog.Any("error", err),
			)
		}
	}
}

// LoadEnvVariables loads the small set of settings that are meaningful at
// runtime. All mutable data paths are derived from DATA_DIR so the container
// has one persistent data volume.
func LoadEnvVariables() {
	DATA_DIR = getEnvOrDefault("DATA_DIR", defaultDataDir)

	SHEET_NAME = getEnvOrDefault("SHEET_NAME", "2025_2026")
	HOUR_COLUMN = getIntEnv("HOUR_COLUMN", 1)
	if HOUR_COLUMN < 0 {
		slog.Warn("HOUR_COLUMN must be non-negative; using default", slog.Int("value", HOUR_COLUMN), slog.Int("default", 1))
		HOUR_COLUMN = 1
	}

	var startTime time.Time
	DAY_START, startTime = getClockEnv("DAY_START", "7.30")
	DAY_START_TIME = time.Date(0, 0, 0, startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)

	DAY_END, _ = getClockEnv("DAY_END", "21.00")

	LAST_COLUMN = getIntEnv("LAST_COLUMN", 40)
	if LAST_COLUMN <= HOUR_COLUMN {
		slog.Warn("LAST_COLUMN must be greater than HOUR_COLUMN; using default", slog.Int("value", LAST_COLUMN), slog.Int("default", 40))
		LAST_COLUMN = 40
		if LAST_COLUMN <= HOUR_COLUMN {
			slog.Warn("HOUR_COLUMN is outside the default timetable range; using default", slog.Int("value", HOUR_COLUMN), slog.Int("default", 1))
			HOUR_COLUMN = 1
		}
	}

	MINUTES_BETWEEN = getIntEnv("MINUTES_BETWEEN", 1)
	if MINUTES_BETWEEN <= 0 {
		slog.Warn("MINUTES_BETWEEN must be greater than zero; using default", slog.Int("value", MINUTES_BETWEEN), slog.Int("default", 1))
		MINUTES_BETWEEN = 1
	}

	TIMETABLE_URL = getEnvOrDefault("TIMETABLE_URL", "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/")
	TIME_BETWEEN_SYNC_HOURS = getIntEnv("TIME_BETWEEN_SYNC_HOURS", 72)
	if TIME_BETWEEN_SYNC_HOURS < 0 {
		slog.Warn("TIME_BETWEEN_SYNC_HOURS must be non-negative; using default", slog.Int("value", TIME_BETWEEN_SYNC_HOURS), slog.Int("default", 72))
		TIME_BETWEEN_SYNC_HOURS = 72
	}
	LOG_LEVEL = getEnvOrDefault("LOG_LEVEL", "DEBUG")
	GOOGLE_PLACES_API_KEY = getEnvOrDefault("GOOGLE_PLACES_API_KEY", "PLEASE_SET_YOUR_GOOGLE_PLACES_API_KEY")

	STATE_DIR = filepath.Join(DATA_DIR, "state")
	LOGS_DIR = filepath.Join(DATA_DIR, "logs")
	DATABASE_PATH = filepath.Join(STATE_DIR, "database.db")
	DATABASE_SOURCE = fmt.Sprintf(
		"file:%s?%s",
		filepath.ToSlash(DATABASE_PATH),
		DATABASE_ARGUMENTS,
	)
	TIMETABLE_DOWNLOADS_DIR = filepath.Join(DATA_DIR, "downloads")
	CURRENT_TIMETABLE_DIR = filepath.Join(DATA_DIR, "current_timetable")
	CALENDAR_CREDENTIALS_PATH = filepath.Join(DATA_DIR, "credentials.json")
	CURRENT_TIMETABLE_STATS_PATH = filepath.Join(STATE_DIR, "current_timetable_stats.json")

	ensureDirectories()
}

func PrintEnvVariables() {
	slog.Info(
		"Current configuration",
		slog.String("SHEET_NAME", SHEET_NAME),
		slog.Int("HOUR_COLUMN", HOUR_COLUMN),
		slog.String("DAY_START", DAY_START),
		slog.String("DAY_END", DAY_END),
		slog.Int("LAST_COLUMN", LAST_COLUMN),
		slog.String("DATABASE_DRIVER", DATABASE_DRIVER),
		slog.Int("MINUTES_BETWEEN", MINUTES_BETWEEN),
		slog.String("SERVER_PORT", SERVER_PORT),
		slog.String("TIMETABLE_URL", TIMETABLE_URL),
		slog.Int("TIME_BETWEEN_SYNC_HOURS", TIME_BETWEEN_SYNC_HOURS),
		slog.Bool("GOOGLE_PLACES_API_KEY_SET", GOOGLE_PLACES_API_KEY != "" && GOOGLE_PLACES_API_KEY != "PLEASE_SET_YOUR_GOOGLE_PLACES_API_KEY"),
		slog.String("DATA_DIR", DATA_DIR),
		slog.String("STATE_DIR", STATE_DIR),
		slog.String("LOGS_DIR", LOGS_DIR),
		slog.String("DATABASE_PATH", DATABASE_PATH),
		slog.String("DATABASE_SOURCE", DATABASE_SOURCE),
		slog.String("STATIC_DIR", STATIC_DIR),
		slog.String("TIMETABLE_DOWNLOADS_DIR", TIMETABLE_DOWNLOADS_DIR),
		slog.String("CURRENT_TIMETABLE_DIR", CURRENT_TIMETABLE_DIR),
		slog.String("CALENDAR_CREDENTIALS_PATH", CALENDAR_CREDENTIALS_PATH),
		slog.String("CURRENT_TIMETABLE_STATS_PATH", CURRENT_TIMETABLE_STATS_PATH),
	)
}
