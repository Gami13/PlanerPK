package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/woodpkrs/PlanerPK/gen/db"
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

func InitializeLogging(logDir string, level slog.Level) error {
	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join(logDir, fmt.Sprintf("app_%s.log", timestamp))

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		TimeFormat: time.Stamp,
	})
	fileHandler := tint.NewHandler(logFile, &tint.Options{
		Level:      level,
		TimeFormat: time.Stamp,
		NoColor:    true,
	})
	slog.SetDefault(slog.New(slog.NewMultiHandler(consoleHandler, fileHandler)))

	slog.Info("Logger initialized",
		slog.String("log_file", logFilePath),
		slog.String("log_level", level.String()))

	return nil
}
