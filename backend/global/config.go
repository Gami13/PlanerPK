package global

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	DatabaseDriver    = "sqlite"
	databaseFilename  = "database.db"
	databaseArguments = "mode=rwc&_busy_timeout=5000&_time_format=sqlite&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)"

	// ServerPort is deliberately fixed. Deployments configure the host-facing
	// port in their reverse proxy or container runtime.
	ServerPort = "8080"
)

// Configuration contains parsed settings and paths derived from DATA_DIR.
type Configuration struct {
	DataDir               string     `env:"DATA_DIR" envDefault:"./data"`
	HourColumn            int        `env:"HOUR_COLUMN" envDefault:"1"`
	DayStart              string     `env:"DAY_START" envDefault:"7.30"`
	DayEnd                string     `env:"DAY_END" envDefault:"21.00"`
	LastColumn            int        `env:"LAST_COLUMN" envDefault:"40"`
	MinutesBetween        int        `env:"MINUTES_BETWEEN" envDefault:"1"`
	TimeBetweenSyncHours  int        `env:"TIME_BETWEEN_SYNC_HOURS" envDefault:"72"`
	TimetableURL          string     `env:"TIMETABLE_URL" envDefault:"https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"`
	GooglePlacesAPIKey    string     `env:"GOOGLE_PLACES_API_KEY"`
	LogLevel              slog.Level `env:"LOG_LEVEL" envDefault:"DEBUG"`
	DayStartTime          time.Time
	TimetableUpdatePeriod time.Duration
	TimetableSyncPeriod   time.Duration
	StateDir              string
	LogsDir               string
	StaticDir             string
	TimetableDownloadsDir string
	CurrentTimetableDir   string
	CalendarCredentials   string
	OAuthConfig           *oauth2.Config
	AllowedOrigins        []string
	TimetableStats        string
	DatabaseSource        string
}

type oauthCredentials struct {
	Web struct {
		JavaScriptOrigins []string `json:"javascript_origins"`
	} `json:"web"`
}

// Config is populated by Load and then treated as read-only.
var Config Configuration

// Load reads, validates, stores the process configuration, and prepares its
// mutable directories. It panics when startup configuration is invalid or the
// filesystem cannot be prepared.
func Load() {
	cfg, err := parseEnvironment(env.ToMap(os.Environ()))
	if err != nil {
		panic(fmt.Sprintf("load configuration: %v", err))
	}
	if err := PrepareFilesystem(cfg); err != nil {
		panic(fmt.Sprintf("prepare runtime directories: %v", err))
	}
	if err := loadOAuthCredentials(&cfg); err != nil {
		panic(fmt.Sprintf("load OAuth credentials: %v", err))
	}
	Config = cfg
}

func parseEnvironment(environment map[string]string) (Configuration, error) {
	if dataDir, isSet := environment["DATA_DIR"]; isSet && strings.TrimSpace(dataDir) == "" {
		return Configuration{}, fmt.Errorf("DATA_DIR must not be empty")
	}

	cfg, err := env.ParseAsWithOptions[Configuration](env.Options{Environment: environment})
	if err != nil {
		return Configuration{}, fmt.Errorf("parse environment: %w", err)
	}

	dayStartTime, err := time.Parse("15.04", cfg.DayStart)
	if err != nil {
		return Configuration{}, fmt.Errorf("DAY_START must use HH.MM format: %w", err)
	}
	if _, err := time.Parse("15.04", cfg.DayEnd); err != nil {
		return Configuration{}, fmt.Errorf("DAY_END must use HH.MM format: %w", err)
	}
	if cfg.HourColumn < 0 {
		return Configuration{}, fmt.Errorf("HOUR_COLUMN must be non-negative")
	}
	if cfg.LastColumn <= cfg.HourColumn {
		return Configuration{}, fmt.Errorf("LAST_COLUMN must be greater than HOUR_COLUMN")
	}
	if cfg.MinutesBetween <= 0 {
		return Configuration{}, fmt.Errorf("MINUTES_BETWEEN must be greater than zero")
	}
	if cfg.TimeBetweenSyncHours < 0 {
		return Configuration{}, fmt.Errorf("TIME_BETWEEN_SYNC_HOURS must be non-negative")
	}

	dataDir, err := filepath.Abs(filepath.Clean(cfg.DataDir))
	if err != nil {
		return Configuration{}, fmt.Errorf("resolve DATA_DIR: %w", err)
	}
	cfg.DataDir = dataDir
	cfg.StateDir = filepath.Join(cfg.DataDir, "state")
	cfg.LogsDir = filepath.Join(cfg.DataDir, "logs")
	cfg.StaticDir = "static"
	cfg.TimetableDownloadsDir = filepath.Join(cfg.DataDir, "downloads")
	cfg.CurrentTimetableDir = filepath.Join(cfg.DataDir, "current_timetable")
	cfg.CalendarCredentials = filepath.Join(cfg.DataDir, "credentials.json")
	cfg.TimetableStats = filepath.Join(cfg.StateDir, "current_timetable_stats.json")
	databasePath := filepath.Join(cfg.StateDir, databaseFilename)
	cfg.DatabaseSource = fmt.Sprintf("file:%s?%s", filepath.ToSlash(databasePath), databaseArguments)
	cfg.DayStartTime = time.Date(0, 0, 0, dayStartTime.Hour(), dayStartTime.Minute(), 0, 0, time.UTC)
	cfg.TimetableUpdatePeriod = time.Duration(cfg.MinutesBetween) * time.Minute
	cfg.TimetableSyncPeriod = time.Duration(cfg.TimeBetweenSyncHours) * time.Hour
	return cfg, nil
}

func loadOAuthCredentials(cfg *Configuration) error {
	credentials, err := os.ReadFile(cfg.CalendarCredentials)
	if err != nil {
		return fmt.Errorf("read credentials file: %w", err)
	}

	oauthConfig, err := google.ConfigFromJSON(credentials, calendar.CalendarScope, "email", "profile")
	if err != nil {
		return fmt.Errorf("parse OAuth config: %w", err)
	}

	var parsedCredentials oauthCredentials
	if err := json.Unmarshal(credentials, &parsedCredentials); err != nil {
		return fmt.Errorf("parse allowed origins: %w", err)
	}
	cfg.OAuthConfig = oauthConfig
	cfg.AllowedOrigins = parsedCredentials.Web.JavaScriptOrigins
	return nil
}

// PrepareFilesystem creates mutable runtime directories. Static assets are
// prepared by the deployment and are intentionally not created here.
func PrepareFilesystem(cfg Configuration) error {
	for _, directory := range []string{
		cfg.StateDir,
		cfg.LogsDir,
		cfg.TimetableDownloadsDir,
		cfg.CurrentTimetableDir,
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("create runtime directory %q: %w", directory, err)
		}
	}
	return nil
}
