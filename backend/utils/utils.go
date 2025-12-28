package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/global"
	"github.com/lmittmann/tint"
)

func GetDateOfNextWeekday(weekday time.Weekday) time.Time {
	today := time.Now()
	daysUntil := (int(weekday) - int(today.Weekday()) + 7) % 7
	next := today.AddDate(0, 0, daysUntil)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
}

// JiPP - Języki i paradygmaty programowania
// WdI - Wstęp do informatyki
// F - Fizyka
// AM - Analiza matematyczna
// J.angielski - Język angielski
// LdI - Logika dla inżynierów
// AzG - Algebra z geometrią
// J. niemiecki - Język niemiecki

// wykład, lab., ćwiczenia

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
		global.SHEET_NAME = sheetNames
	}
	lastRow := os.Getenv("LAST_ROW")
	if lastRow != "" {
		val, err := strconv.Atoi(lastRow)
		if err == nil {
			global.LAST_ROW = val
		}
	}
	hourColumn := os.Getenv("HOUR_COLUMN")
	if hourColumn != "" {
		val, err := strconv.Atoi(hourColumn)
		if err == nil {
			global.HOUR_COLUMN = val
		}
	}
	dayStart := os.Getenv("DAY_START")
	if dayStart != "" {
		global.DAY_START = dayStart
	}
	dayEnd := os.Getenv("DAY_END")
	if dayEnd != "" {
		global.DAY_END = dayEnd
	}
	databaseSource := os.Getenv("DATABASE_SOURCE")
	if databaseSource != "" {
		global.DATABASE_SOURCE = databaseSource
	}
	databaseDriver := os.Getenv("DATABASE_DRIVER")
	if databaseDriver != "" {
		global.DATABASE_DRIVER = databaseDriver
	}
	minutesBetween := os.Getenv("MINUTES_BETWEEN")
	if minutesBetween != "" {
		val, err := strconv.Atoi(minutesBetween)
		if err == nil {
			global.MINUTES_BETWEEN = val
		}
	}
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort != "" {
		global.SERVER_PORT = serverPort
	}
	generalRateLimit := os.Getenv("GENERAL_RATE_LIMIT_PER_MINUTE")
	if generalRateLimit != "" {
		val, err := strconv.Atoi(generalRateLimit)
		if err == nil {
			global.GENERAL_RATE_LIMIT_PER_MINUTE = val
		}
	}
	timetableCurrentDestination := os.Getenv("TIMETABLE_CURRENT_DESTINATION")
	if timetableCurrentDestination != "" {
		global.TIMETABLE_CURRENT_DESTINATION = timetableCurrentDestination
	}
	timetableURL := os.Getenv("TIMETABLE_URL")
	if timetableURL != "" {
		global.TIMETABLE_URL = timetableURL
	}
	timetableDownloadDestination := os.Getenv("TIMETABLE_DOWNLOAD_DESTINATION")
	if timetableDownloadDestination != "" {
		global.TIMETABLE_DOWNLOAD_DESTINATION = timetableDownloadDestination
	}
	calendarCredentialsPath := os.Getenv("CALENDAR_CREDENTIALS_PATH")
	if calendarCredentialsPath != "" {
		global.CALENDAR_CREDENTIALS_PATH = calendarCredentialsPath
	}
	timeBetweenSyncHours := os.Getenv("TIME_BETWEEN_SYNC_HOURS")
	if timeBetweenSyncHours != "" {
		val, err := strconv.Atoi(timeBetweenSyncHours)
		if err == nil {
			global.TIME_BETWEEN_SYNC_HOURS = val
		}
	}
	lastColummn := os.Getenv("LAST_COLUMN")
	if lastColummn != "" {
		val, err := strconv.Atoi(lastColummn)
		if err == nil {
			global.LAST_COLUMN = val
		}
	}
	logsPath := os.Getenv("LOGS_PATH")
	if logsPath != "" {
		global.LOGS_PATH = logsPath
	}
}

func InitializeLogging(logDir string) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join(logDir, fmt.Sprintf("app_%s.log", timestamp))

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create a multi-writer that writes to both file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Create a handler with nice formatting

	slog.SetDefault(slog.New(tint.NewHandler(multiWriter, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Stamp,
	})))

	slog.Info("Logger initialized",
		slog.String("log_file", logFilePath),
		slog.String("log_level", "INFO"))

	return nil
}
