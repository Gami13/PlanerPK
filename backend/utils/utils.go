package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/WoodPKrs/GoogleCalendarTimetable/gen/db"
	"github.com/lmittmann/tint"
	"golang.org/x/oauth2"
)

func GetDateOfNextWeekday(weekday time.Weekday) time.Time {
	today := time.Now()
	daysUntil := (int(weekday) - int(today.Weekday()) + 7) % 7
	next := today.AddDate(0, 0, daysUntil)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
}

func GetToken(u *db.User) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  u.AccessToken,
		TokenType:    u.TokenType,
		RefreshToken: u.RefreshToken,
		Expiry:       u.Expiry,
		ExpiresIn:    u.ExpiresIn,
	}
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
