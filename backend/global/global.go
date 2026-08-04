package global

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	SHEET_NAME                    = "2025_2026"
	LAST_ROW                      = 365
	HOUR_COLUMN                   = 1
	DAY_START                     = "7.30"
	DAY_START_TIME                = time.Date(0, 0, 0, 7, 30, 0, 0, time.UTC)
	DAY_END                       = "21.00"
	LAST_COLUMN                   = 40
	DATABASE_DRIVER               = "sqlite"
	MINUTES_BETWEEN               = 1
	SERVER_PORT                   = "8080"
	GENERAL_RATE_LIMIT_PER_MINUTE = 60
	TIMETABLE_URL                 = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
	TIME_BETWEEN_SYNC_HOURS       = 72
	GOOGLE_PLACES_API_KEY         = "PLEASE_SET_YOUR_GOOGLE_PLACES_API_KEY"

	// Override for XDG directories. If set, everything will be put into the specified directory
	APP_BASE_DIR string

	// XDG directories
	XDG_DATA_HOME   string
	XDG_CONFIG_HOME string
	XDG_STATE_HOME  string
	XDG_CACHE_HOME  string

	// Application main directories
	APP_DIR    string
	DATA_DIR   string
	CONFIG_DIR string
	STATE_DIR  string
	STATIC_DIR string

	LOGS_DIR                     string
	DATABASE_SOURCE              string
	TIMETABLE_DOWNLOADS_DIR      string
	CURRENT_TIMETABLE_DIR        string
	CALENDAR_CREDENTIALS_PATH    string
	CURRENT_TIMETABLE_STATS_PATH string
)

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// var LAST_ROW = 365
// var HOUR_COLUMN = 1
// var DAY_START = "7.30"
// var DAY_END = "21.00"
// var DATABASE_SOURCE = "file:database.db?cache=shared&mode=rwc&_busy_timeout=5000"
// var DATABASE_DRIVER = "sqlite"
// var MINUTES_BETWEEN = 5
// var SERVER_PORT = "8080"
// var TIMETABLE_CURRENT_DESTINATION = "./current_timetable/"
// var TIMETABLE_URL = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
// var TIMETABLE_DOWNLOAD_DESTINATION = "./downloads/"
// Loads all environment variables from the system, if they don't exist, uses default ones
func LoadEnvVariables() {
	APP_BASE_DIR = os.Getenv("APP_BASE_DIR")

	if APP_BASE_DIR != "" {
		APP_DIR = APP_BASE_DIR
		DATA_DIR = filepath.Join(APP_DIR, "data")
		CONFIG_DIR = filepath.Join(APP_DIR, "config")
		STATE_DIR = filepath.Join(APP_DIR, "state")
		STATIC_DIR = filepath.Join(APP_DIR, "static")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("Failed to determine the user home directory. You must set the `APP_BASE_DIR` environment variable to specify where data should be stored.", slog.Any("error", err))
			panic(err)
		}

		XDG_DATA_HOME = getEnvOrDefault("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
		XDG_CONFIG_HOME = getEnvOrDefault("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
		XDG_STATE_HOME = getEnvOrDefault("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
		XDG_CACHE_HOME = getEnvOrDefault("XDG_CACHE_HOME", filepath.Join(home, ".cache"))

		APP_DIR = filepath.Join(XDG_DATA_HOME, "planer-pk")
		DATA_DIR = filepath.Join(APP_DIR, "data")
		CONFIG_DIR = filepath.Join(XDG_CONFIG_HOME, "planer-pk")
		STATE_DIR = filepath.Join(XDG_STATE_HOME, "planer-pk")
		STATIC_DIR = filepath.Join(XDG_CACHE_HOME, "planer-pk", "static")
	}

	if val := os.Getenv("SHEET_NAME"); val != "" {
		SHEET_NAME = val
	}

	if val := os.Getenv("LAST_ROW"); val != "" {
		if val, err := strconv.Atoi(val); err == nil {
			LAST_ROW = val
		}
	}

	if val := os.Getenv("HOUR_COLUMN"); val != "" {
		if val, err := strconv.Atoi(val); err == nil {
			HOUR_COLUMN = val
		}
	}

	if val := os.Getenv("DAY_START"); val != "" {
		DAY_START = val
		timeParsed, err := time.Parse("15.04", val)
		if err != nil {
			slog.Error("Failed to parse DAY_START environment variable", slog.Any("error", err))
			panic(err)
		}
		DAY_START_TIME = time.Date(0, 0, 0, timeParsed.Hour(), timeParsed.Minute(), 0, 0, time.UTC)

	}

	if val := os.Getenv("DAY_END"); val != "" {
		DAY_END = val
	}

	if val := os.Getenv("DATABASE_DRIVER"); val != "" {
		DATABASE_DRIVER = val
	}

	if minutesBetween := os.Getenv("MINUTES_BETWEEN"); minutesBetween != "" {
		if minutesBetween, err := strconv.Atoi(minutesBetween); err == nil {
			MINUTES_BETWEEN = minutesBetween
		}
	}

	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		SERVER_PORT = serverPort
	}

	if generalRateLimit := os.Getenv("GENERAL_RATE_LIMIT_PER_MINUTE"); generalRateLimit != "" {
		if generalRateLimit, err := strconv.Atoi(generalRateLimit); err == nil {
			GENERAL_RATE_LIMIT_PER_MINUTE = generalRateLimit
		}
	}

	if timetableURL := os.Getenv("TIMETABLE_URL"); timetableURL != "" {
		TIMETABLE_URL = timetableURL
	}

	if timeBetweenSyncHours := os.Getenv("TIME_BETWEEN_SYNC_HOURS"); timeBetweenSyncHours != "" {
		if timeBetweenSyncHours, err := strconv.Atoi(timeBetweenSyncHours); err == nil {
			TIME_BETWEEN_SYNC_HOURS = timeBetweenSyncHours
		}
	}

	if lastColumn := os.Getenv("LAST_COLUMN"); lastColumn != "" {
		if lastColumn, err := strconv.Atoi(lastColumn); err == nil {
			LAST_COLUMN = lastColumn
		}
	}

	if appDir := os.Getenv("APP_DIR"); appDir != "" {
		APP_DIR = appDir
	} else {
		os.Mkdir(APP_DIR, 0o755)
	}

	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		DATA_DIR = dataDir
	} else {
		os.Mkdir(DATA_DIR, 0o755)
	}

	if configDir := os.Getenv("CONFIG_DIR"); configDir != "" {
		CONFIG_DIR = configDir
	} else {
		os.Mkdir(CONFIG_DIR, 0o755)
	}

	if stateDir := os.Getenv("STATE_DIR"); stateDir != "" {
		STATE_DIR = stateDir
	} else {
		if os.Getenv("LOGS_DIR") == "" {
			os.Mkdir(STATE_DIR, 0o755)
		}
	}

	if val := os.Getenv("LOGS_DIR"); val != "" {
		LOGS_DIR = val
	} else {
		LOGS_DIR = filepath.Join(STATE_DIR, "logs")
		os.Mkdir(LOGS_DIR, 0o755)
	}

	if val := os.Getenv("DATABASE_SOURCE"); val != "" {
		DATABASE_SOURCE = val
	} else {
		DATABASE_SOURCE = fmt.Sprintf("file:%s?cache=shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite", filepath.ToSlash(filepath.Join(DATA_DIR, "database.db")))
	}

	if val := os.Getenv("STATIC_DIR"); val != "" {
		STATIC_DIR = val
	} else {
		if APP_BASE_DIR != "" {
			os.Mkdir(filepath.Join(XDG_CACHE_HOME, "planer-pk"), 0o755)
		}
		os.Mkdir(STATIC_DIR, 0o755)
	}

	if val := os.Getenv("TIMETABLE_DOWNLOADS_DIR"); val != "" {
		TIMETABLE_DOWNLOADS_DIR = val
	} else {
		TIMETABLE_DOWNLOADS_DIR = filepath.Join(DATA_DIR, "downloads")
		os.Mkdir(TIMETABLE_DOWNLOADS_DIR, 0o755)
	}

	if val := os.Getenv("CURRENT_TIMETABLE_DIR"); val != "" {
		CURRENT_TIMETABLE_DIR = val
	} else {
		CURRENT_TIMETABLE_DIR = filepath.Join(DATA_DIR, "current_timetable")
		os.Mkdir(CURRENT_TIMETABLE_DIR, 0o755)
	}

	if calendarCredentialsPath := os.Getenv("CALENDAR_CREDENTIALS_PATH"); calendarCredentialsPath != "" {
		CALENDAR_CREDENTIALS_PATH = calendarCredentialsPath
	} else {
		CALENDAR_CREDENTIALS_PATH = filepath.Join(CONFIG_DIR, "credentials.json")
	}

	if val := os.Getenv("CURRENT_TIMETABLE_STATS_PATH"); val != "" {
		CURRENT_TIMETABLE_STATS_PATH = val
	} else {
		CURRENT_TIMETABLE_STATS_PATH = filepath.Join(DATA_DIR, "current_timetable_stats.json")
	}

	if googlePlacesAPIKey := os.Getenv("GOOGLE_PLACES_API_KEY"); googlePlacesAPIKey != "" {
		GOOGLE_PLACES_API_KEY = googlePlacesAPIKey
	}
}

func PrintEnvVariables() {
	slog.Info(
		"Current configuration",
		slog.String("SHEET_NAME", SHEET_NAME),
		slog.Int("LAST_ROW", LAST_ROW),
		slog.Int("HOUR_COLUMN", HOUR_COLUMN),
		slog.String("DAY_START", DAY_START),
		slog.String("DAY_END", DAY_END),
		slog.Int("LAST_COLUMN", LAST_COLUMN),
		slog.String("DATABASE_DRIVER", DATABASE_DRIVER),
		slog.Int("MINUTES_BETWEEN", MINUTES_BETWEEN),
		slog.String("SERVER_PORT", SERVER_PORT),
		slog.Int("GENERAL_RATE_LIMIT_PER_MINUTE", GENERAL_RATE_LIMIT_PER_MINUTE),
		slog.String("TIMETABLE_URL", TIMETABLE_URL),
		slog.Int("TIME_BETWEEN_SYNC_HOURS", TIME_BETWEEN_SYNC_HOURS),
		slog.String("GOOGLE_PLACES_API_KEY", GOOGLE_PLACES_API_KEY),

		slog.String("APP_BASE_DIR", APP_BASE_DIR),

		slog.String("XDG_DATA_HOME", XDG_DATA_HOME),
		slog.String("XDG_CONFIG_HOME", XDG_CONFIG_HOME),
		slog.String("XDG_STATE_HOME", XDG_STATE_HOME),
		slog.String("XDG_CACHE_HOME", XDG_CACHE_HOME),

		slog.String("APP_DIR", APP_DIR),
		slog.String("DATA_DIR", DATA_DIR),
		slog.String("CONFIG_DIR", CONFIG_DIR),
		slog.String("STATE_DIR", STATE_DIR),

		slog.String("LOGS_DIR", LOGS_DIR),
		slog.String("DATABASE_SOURCE", DATABASE_SOURCE),
		slog.String("STATIC_DIR", STATIC_DIR),
		slog.String("TIMETABLE_DOWNLOADS_DIR", TIMETABLE_DOWNLOADS_DIR),
		slog.String("CURRENT_TIMETABLE_DIR", CURRENT_TIMETABLE_DIR),
		slog.String("CALENDAR_CREDENTIALS_PATH", CALENDAR_CREDENTIALS_PATH),
		slog.String("CURRENT_TIMETABLE_STATS_PATH", CURRENT_TIMETABLE_STATS_PATH),
	)
}
