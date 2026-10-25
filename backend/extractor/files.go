package extractor

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"

	"github.com/xuri/excelize/v2"
)

// CurrentTimetableHash identifies XLSX bytes independently of download headers.
func CurrentTimetableHash() (string, error) {
	path, err := getCurrentXlsxPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func logCloseWarn(resource string, err error) {
	if err != nil {
		slog.Warn("Failed to close resource",
			slog.String("resource", resource),
			slog.Any("error", err))
	}
}

func CloseFile(file *excelize.File) {
	if err := file.Close(); err != nil {
		slog.Warn("Failed to close file", slog.Any("error", err))
	}
}
