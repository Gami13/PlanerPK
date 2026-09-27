package extractor

import (
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

type cellFormat uint8

const (
	legacyFormat cellFormat = iota
	structuredFormat
)

// detectCellFormat selects one parser for the whole workbook. It deliberately
// does not use the legacy parser as a per-cell fallback.
func detectCellFormat(file *excelize.File, sheet string, groups []types.ColumnGroup, days []types.Day) cellFormat {
	for _, day := range days {
		for row := day.StartingRow; row <= day.EndingRow; row++ {
			for _, group := range groups {
				if hasStructuredTimeLine(getCell(file, sheet, row, group.Column)) {
					return structuredFormat
				}
			}
		}
	}
	return legacyFormat
}

func hasStructuredTimeLine(text string) bool {
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if _, _, _, ok := parseTimeLocation(strings.TrimSpace(line), time.Time{}); ok {
			return true
		}
	}
	return false
}
