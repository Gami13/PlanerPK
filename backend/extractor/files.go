package extractor

import (
	"log/slog"

	"github.com/xuri/excelize/v2"
)

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
