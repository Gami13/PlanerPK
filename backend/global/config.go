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
	defaultDataDir        = "/data"
	defaultStaticDir      = "/app/static"
	defaultSheetName      = "2025_2026"
	defaultHourColumn     = 1
	defaultDayStart       = "7.30"
	defaultDayEnd         = "21.00"
	defaultLastColumn     = 40
	defaultMinutesBetween = 1
	defaultSyncInterval   = 72
	defaultServerPort     = "8080"
	defaultTimetableURL   = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
	defaultLogLevel       = "DEBUG"
	defaultDatabaseDriver = "sqlite"
	databaseArguments     = "cache=shared&mode=rwc&_busy_timeout=5000&_time_format=sqlite"
	defaultPlacesAPIKey   = "PLEASE_SET_YOUR_GOOGLE_PLACES_API_KEY"
)

type config struct {
	timetable    timetableConfig
	database     databaseConfig
	paths        pathConfig
	server       serverConfig
	integrations integrationConfig
	logging      loggingConfig
}

type timetableConfig struct {
	sheetName      string
	hourColumn     int
	dayStart       string
	dayStartTime   time.Time
	dayEnd         string
	lastColumn     int
	updateInterval time.Duration
	url            string
	syncInterval   time.Duration
}

type databaseConfig struct {
	driver string
	path   string
	source string
}

type pathConfig struct {
	dataDir                 string
	stateDir                string
	logsDir                 string
	staticDir               string
	timetableDownloadsDir   string
	currentTimetableDir     string
	calendarCredentialsPath string
	timetableStatsPath      string
}

type serverConfig struct {
	port string
}

type integrationConfig struct {
	googlePlacesAPIKey string
}

type loggingConfig struct {
	level string
}

var current config

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

func ensureDirectories(paths pathConfig) {
	for _, directory := range []string{
		paths.stateDir,
		paths.logsDir,
		paths.timetableDownloadsDir,
		paths.currentTimetableDir,
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			slog.Error("Failed to create data directory",
				slog.String("directory", directory),
				slog.Any("error", err),
			)
		}
	}
}

// LoadEnvVariables loads runtime settings and derives all application paths.
// It should be called once before starting the application or any worker.
func LoadEnvVariables() {
	dataDir := getEnvOrDefault("DATA_DIR", defaultDataDir)

	hourColumn := getIntEnv("HOUR_COLUMN", defaultHourColumn)
	if hourColumn < 0 {
		slog.Warn("HOUR_COLUMN must be non-negative; using default",
			slog.Int("value", hourColumn), slog.Int("default", defaultHourColumn))
		hourColumn = defaultHourColumn
	}

	dayStart, startTime := getClockEnv("DAY_START", defaultDayStart)
	dayStartTime := time.Date(0, 0, 0, startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	dayEnd, _ := getClockEnv("DAY_END", defaultDayEnd)

	lastColumn := getIntEnv("LAST_COLUMN", defaultLastColumn)
	if lastColumn <= hourColumn {
		slog.Warn("LAST_COLUMN must be greater than HOUR_COLUMN; using default",
			slog.Int("value", lastColumn), slog.Int("default", defaultLastColumn))
		lastColumn = defaultLastColumn
		if lastColumn <= hourColumn {
			slog.Warn("HOUR_COLUMN is outside the default timetable range; using default",
				slog.Int("value", hourColumn), slog.Int("default", defaultHourColumn))
			hourColumn = defaultHourColumn
		}
	}

	minutesBetween := getIntEnv("MINUTES_BETWEEN", defaultMinutesBetween)
	if minutesBetween <= 0 {
		slog.Warn("MINUTES_BETWEEN must be greater than zero; using default",
			slog.Int("value", minutesBetween), slog.Int("default", defaultMinutesBetween))
		minutesBetween = defaultMinutesBetween
	}

	syncHours := getIntEnv("TIME_BETWEEN_SYNC_HOURS", defaultSyncInterval)
	if syncHours < 0 {
		slog.Warn("TIME_BETWEEN_SYNC_HOURS must be non-negative; using default",
			slog.Int("value", syncHours), slog.Int("default", defaultSyncInterval))
		syncHours = defaultSyncInterval
	}

	stateDir := filepath.Join(dataDir, "state")
	paths := pathConfig{
		dataDir:                 dataDir,
		stateDir:                stateDir,
		logsDir:                 filepath.Join(dataDir, "logs"),
		staticDir:               defaultStaticDir,
		timetableDownloadsDir:   filepath.Join(dataDir, "downloads"),
		currentTimetableDir:     filepath.Join(dataDir, "current_timetable"),
		calendarCredentialsPath: filepath.Join(dataDir, "credentials.json"),
		timetableStatsPath:      filepath.Join(stateDir, "current_timetable_stats.json"),
	}

	databasePath := filepath.Join(stateDir, "database.db")
	current = config{
		timetable: timetableConfig{
			sheetName:      getEnvOrDefault("SHEET_NAME", defaultSheetName),
			hourColumn:     hourColumn,
			dayStart:       dayStart,
			dayStartTime:   dayStartTime,
			dayEnd:         dayEnd,
			lastColumn:     lastColumn,
			updateInterval: time.Duration(minutesBetween) * time.Minute,
			url:            getEnvOrDefault("TIMETABLE_URL", defaultTimetableURL),
			syncInterval:   time.Duration(syncHours) * time.Hour,
		},
		database: databaseConfig{
			driver: defaultDatabaseDriver,
			path:   databasePath,
			source: fmt.Sprintf("file:%s?%s", filepath.ToSlash(databasePath), databaseArguments),
		},
		paths: paths,
		server: serverConfig{
			// The container listens on a fixed internal port. Change the
			// host-facing port through Docker or Compose.
			port: defaultServerPort,
		},
		integrations: integrationConfig{
			googlePlacesAPIKey: getEnvOrDefault("GOOGLE_PLACES_API_KEY", defaultPlacesAPIKey),
		},
		logging: loggingConfig{
			level: getEnvOrDefault("LOG_LEVEL", defaultLogLevel),
		},
	}

	ensureDirectories(current.paths)
}

func SheetName() string                      { return current.timetable.sheetName }
func HourColumn() int                        { return current.timetable.hourColumn }
func DayStart() string                       { return current.timetable.dayStart }
func DayStartTime() time.Time                { return current.timetable.dayStartTime }
func DayEnd() string                         { return current.timetable.dayEnd }
func LastColumn() int                        { return current.timetable.lastColumn }
func TimetableUpdateInterval() time.Duration { return current.timetable.updateInterval }
func TimetableURL() string                   { return current.timetable.url }
func TimeBetweenSync() time.Duration         { return current.timetable.syncInterval }
func DatabaseDriver() string                 { return current.database.driver }
func DatabasePath() string                   { return current.database.path }
func DatabaseSource() string                 { return current.database.source }
func DataDir() string                        { return current.paths.dataDir }
func StateDir() string                       { return current.paths.stateDir }
func LogsDir() string                        { return current.paths.logsDir }
func StaticDir() string                      { return current.paths.staticDir }
func TimetableDownloadsDir() string          { return current.paths.timetableDownloadsDir }
func CurrentTimetableDir() string            { return current.paths.currentTimetableDir }
func CalendarCredentialsPath() string        { return current.paths.calendarCredentialsPath }
func CurrentTimetableStatsPath() string      { return current.paths.timetableStatsPath }
func ServerPort() string                     { return current.server.port }
func GooglePlacesAPIKey() string             { return current.integrations.googlePlacesAPIKey }
func LogLevel() string                       { return current.logging.level }

func PrintEnvVariables() {
	slog.Info(
		"Current configuration",
		slog.String("SHEET_NAME", SheetName()),
		slog.Int("HOUR_COLUMN", HourColumn()),
		slog.String("DAY_START", DayStart()),
		slog.String("DAY_END", DayEnd()),
		slog.Int("LAST_COLUMN", LastColumn()),
		slog.String("DATABASE_DRIVER", DatabaseDriver()),
		slog.Duration("UPDATE_INTERVAL", TimetableUpdateInterval()),
		slog.String("SERVER_PORT", ServerPort()),
		slog.String("TIMETABLE_URL", TimetableURL()),
		slog.Duration("TIME_BETWEEN_SYNC", TimeBetweenSync()),
		slog.Bool("GOOGLE_PLACES_API_KEY_SET", GooglePlacesAPIKey() != "" && GooglePlacesAPIKey() != defaultPlacesAPIKey),
		slog.String("DATA_DIR", DataDir()),
		slog.String("STATE_DIR", StateDir()),
		slog.String("LOGS_DIR", LogsDir()),
		slog.String("DATABASE_PATH", DatabasePath()),
		slog.String("DATABASE_SOURCE", DatabaseSource()),
		slog.String("STATIC_DIR", StaticDir()),
		slog.String("TIMETABLE_DOWNLOADS_DIR", TimetableDownloadsDir()),
		slog.String("CURRENT_TIMETABLE_DIR", CurrentTimetableDir()),
		slog.String("CALENDAR_CREDENTIALS_PATH", CalendarCredentialsPath()),
		slog.String("CURRENT_TIMETABLE_STATS_PATH", CurrentTimetableStatsPath()),
	)
}
