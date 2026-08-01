package global

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

var (
	SHEET_NAME                     = "2025_2026"
	LAST_ROW                       = 365
	HOUR_COLUMN                    = 1
	DAY_START                      = "7.30"
	DAY_START_TIME                 = time.Date(0, 0, 0, 7, 30, 0, 0, time.UTC)
	DAY_END                        = "21.00"
	DATABASE_SOURCE                = "file:../data/database.db?cache=shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite"
	LAST_COLUMN                    = 40
	DATABASE_DRIVER                = "sqlite"
	MINUTES_BETWEEN                = 1
	SERVER_PORT                    = "8080"
	GENERAL_RATE_LIMIT_PER_MINUTE  = 60
	TIMETABLE_CURRENT_DESTINATION  = "../data/current_timetable/"
	TIMETABLE_URL                  = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
	TIMETABLE_DOWNLOAD_DESTINATION = "../data/downloads/"
	CALENDAR_CREDENTIALS_PATH      = "../data/credentials.json"
	LOGS_PATH                      = "../data/logs"
	TIME_BETWEEN_SYNC_HOURS        = 72
	TIMETABLE_CURRENT_STATS_NAME   = "current_timetable_stats.json"
	JSON_INFO_LOCATION             = "../data"
	GOOGLE_PLACES_API_KEY          = "PLEASE_SET_YOUR_GOOGLE_PLACES_API_KEY"
)

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
	sheetNames := os.Getenv("SHEET_NAME")
	if sheetNames != "" {
		SHEET_NAME = sheetNames
	}
	lastRow := os.Getenv("LAST_ROW")
	if lastRow != "" {
		val, err := strconv.Atoi(lastRow)
		if err == nil {
			LAST_ROW = val
		}
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
	databaseSource := os.Getenv("DATABASE_SOURCE")
	if databaseSource != "" {
		DATABASE_SOURCE = databaseSource
	}
	databaseDriver := os.Getenv("DATABASE_DRIVER")
	if databaseDriver != "" {
		DATABASE_DRIVER = databaseDriver
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
	timetableCurrentDestination := os.Getenv("TIMETABLE_CURRENT_DESTINATION")
	if timetableCurrentDestination != "" {
		TIMETABLE_CURRENT_DESTINATION = timetableCurrentDestination
	}
	timetableURL := os.Getenv("TIMETABLE_URL")
	if timetableURL != "" {
		TIMETABLE_URL = timetableURL
	}
	timetableDownloadDestination := os.Getenv("TIMETABLE_DOWNLOAD_DESTINATION")
	if timetableDownloadDestination != "" {
		TIMETABLE_DOWNLOAD_DESTINATION = timetableDownloadDestination
	}
	calendarCredentialsPath := os.Getenv("CALENDAR_CREDENTIALS_PATH")
	if calendarCredentialsPath != "" {
		CALENDAR_CREDENTIALS_PATH = calendarCredentialsPath
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
	logsPath := os.Getenv("LOGS_PATH")
	if logsPath != "" {
		LOGS_PATH = logsPath
	}
	jsonInfoLocation := os.Getenv("JSON_INFO_LOCATION")
	if jsonInfoLocation != "" {
		JSON_INFO_LOCATION = jsonInfoLocation
	}
	googlePlacesAPIKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	if googlePlacesAPIKey != "" {
		GOOGLE_PLACES_API_KEY = googlePlacesAPIKey
	}
}

func PrintEnvVariables() {
	slog.Info("Current configuration",
		slog.String("SHEET_NAME", SHEET_NAME),
		slog.Int("LAST_ROW", LAST_ROW),
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
