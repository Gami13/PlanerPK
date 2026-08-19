package testutil

import (
	"context"
	"log/slog"
	"testing"
)

// LogCapture records slog records while installed as the default logger.
type LogCapture struct {
	Records []slog.Record
}

func (c *LogCapture) Enabled(context.Context, slog.Level) bool { return true }

func (c *LogCapture) Handle(_ context.Context, record slog.Record) error {
	c.Records = append(c.Records, record)
	return nil
}

func (c *LogCapture) WithAttrs([]slog.Attr) slog.Handler { return c }

func (c *LogCapture) WithGroup(string) slog.Handler { return c }

// CaptureLogs installs a record-capturing default logger and restores the
// previous default logger when the test finishes.
func CaptureLogs(t testing.TB) *LogCapture {
	t.Helper()

	capture := &LogCapture{}
	previous := slog.Default()
	slog.SetDefault(slog.New(capture))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return capture
}
