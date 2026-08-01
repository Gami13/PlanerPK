package global

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	DATA_DIRECTORY                = "../data"
	GOOGLE_PLACES_API_KEY         = "PLEASE_SET_YOUR_GOOGLE_PLACES_API_KEY"
	SERVER_PORT                   = "8080"
	MINUTES_BETWEEN               = 1
	TIME_BETWEEN_SYNC_HOURS       = 72
	GENERAL_RATE_LIMIT_PER_MINUTE = 60
	TIMETABLE_URL                 = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"

	// Could potentially be handled by the parser to not rely on configs
	SHEET_NAME     = "2025_2026"
	HOUR_COLUMN    = 1
	DAY_START      = "7.30"
	DAY_START_TIME = time.Date(0, 0, 0, 7, 30, 0, 0, time.UTC)
	DAY_END        = "21.00"
	LAST_COLUMN    = 40

	// !STATIC CONFIGS! - These are not loaded from environment variables
	DATABASE_PATH                  = DATA_DIRECTORY + "/database.db"
	DATABASE_ARGUMENTS             = "?cache=shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite"
	DATABASE_DRIVER                = "sqlite"
	DATABASE_SOURCE                = "file:" + DATABASE_PATH + DATABASE_ARGUMENTS
	TIMETABLE_CURRENT_DESTINATION  = "/data/current_timetable/"
	TIMETABLE_DOWNLOAD_DESTINATION = "/data/downloads/"
	CALENDAR_CREDENTIALS_PATH      = "/data/credentials.json"
	LOGS_PATH                      = "/data/logs"
	TIMETABLE_CURRENT_STATS_NAME   = "current_timetable_stats.json"
	JSON_INFO_LOCATION             = "/data"
)

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
	dataDir := os.Getenv("DATA_DIRECTORY")
	if dataDir != "" {
		DATA_DIRECTORY = dataDir
	}

	TIMETABLE_CURRENT_DESTINATION = filepath.Join(DATA_DIRECTORY, "current_timetable")
	TIMETABLE_DOWNLOAD_DESTINATION = filepath.Join(DATA_DIRECTORY, "downloads")
	CALENDAR_CREDENTIALS_PATH = filepath.Join(DATA_DIRECTORY, "credentials.json")
	LOGS_PATH = filepath.Join(DATA_DIRECTORY, "logs")
	DATABASE_PATH = filepath.Join(DATA_DIRECTORY, "database.db")
	DATABASE_SOURCE = "file:" + DATABASE_PATH + DATABASE_ARGUMENTS
	JSON_INFO_LOCATION = DATA_DIRECTORY

	sheetNames := os.Getenv("SHEET_NAME")
	if sheetNames != "" {
		SHEET_NAME = sheetNames
	}

	hourColumn := os.Getenv("HOUR_COLUMN")
	if hourColumn != "" {
		val, err := strconv.Atoi(hourColumn)
		if err == nil {
			HOUR_COLUMN = val
		}
	}
	dayStart := os.Getenv("DAY_START")
	if dayStart != "" {
		DAY_START = dayStart
		timeParsed, err := time.Parse("15.04", dayStart)
		if err != nil {
			slog.Error("Failed to parse DAY_START environment variable", slog.Any("error", err))
			panic(err)
		}
		DAY_START_TIME = time.Date(0, 0, 0, timeParsed.Hour(), timeParsed.Minute(), 0, 0, time.UTC)

	}
	dayEnd := os.Getenv("DAY_END")
	if dayEnd != "" {
		DAY_END = dayEnd
	}

	minutesBetween := os.Getenv("MINUTES_BETWEEN")
	if minutesBetween != "" {
		val, err := strconv.Atoi(minutesBetween)
		if err == nil {
			MINUTES_BETWEEN = val
		}
	}
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort != "" {
		SERVER_PORT = serverPort
	}
	generalRateLimit := os.Getenv("GENERAL_RATE_LIMIT_PER_MINUTE")
	if generalRateLimit != "" {
		val, err := strconv.Atoi(generalRateLimit)
		if err == nil {
			GENERAL_RATE_LIMIT_PER_MINUTE = val
		}
	}

	timetableURL := os.Getenv("TIMETABLE_URL")
	if timetableURL != "" {
		TIMETABLE_URL = timetableURL
	}

	timeBetweenSyncHours := os.Getenv("TIME_BETWEEN_SYNC_HOURS")
	if timeBetweenSyncHours != "" {
		val, err := strconv.Atoi(timeBetweenSyncHours)
		if err == nil {
			TIME_BETWEEN_SYNC_HOURS = val
		}
	}
	lastColumn := os.Getenv("LAST_COLUMN")
	if lastColumn != "" {
		val, err := strconv.Atoi(lastColumn)
		if err == nil {
			LAST_COLUMN = val
		}
	}

	googlePlacesAPIKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	if googlePlacesAPIKey != "" {
		GOOGLE_PLACES_API_KEY = googlePlacesAPIKey
	}
}

func PrintEnvVariables() {
	slog.Info("Current configuration",
		slog.String("DATA_DIRECTORY", DATA_DIRECTORY),
		slog.String("SHEET_NAME", SHEET_NAME),

		slog.Int("HOUR_COLUMN", HOUR_COLUMN),
		slog.String("DAY_START", DAY_START),
		slog.String("DAY_END", DAY_END),
		slog.String("DATABASE_SOURCE", DATABASE_SOURCE),
		slog.String("DATABASE_DRIVER", DATABASE_DRIVER),
		slog.Int("MINUTES_BETWEEN", MINUTES_BETWEEN),
		slog.String("SERVER_PORT", SERVER_PORT),
		slog.String("TIMETABLE_CURRENT_DESTINATION", TIMETABLE_CURRENT_DESTINATION),
		slog.String("TIMETABLE_URL", TIMETABLE_URL),
		slog.String("TIMETABLE_DOWNLOAD_DESTINATION", TIMETABLE_DOWNLOAD_DESTINATION),
		slog.String("CALENDAR_CREDENTIALS_PATH", CALENDAR_CREDENTIALS_PATH),
		slog.Int("TIME_BETWEEN_SYNC_HOURS", TIME_BETWEEN_SYNC_HOURS),
		slog.Int("LAST_COLUMN", LAST_COLUMN),
		slog.String("LOGS_PATH", LOGS_PATH),
		slog.String("JSON_INFO_LOCATION", JSON_INFO_LOCATION),
		slog.String("GOOGLE_PLACES_API_KEY", GOOGLE_PLACES_API_KEY),
	)
}
