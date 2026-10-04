package extractor

import (
	"log/slog"

	"github.com/xuri/excelize/v2"
)

// mergeRange stores parsed information about a single merged cell range.
type mergeRange struct {
	StartRow, StartCol int
	EndRow, EndCol     int
	Value              string
}

func loadMergeRanges(file *excelize.File, sheetName string) ([]mergeRange, error) {
	mergeCells, err := file.GetMergeCells(sheetName)
	if err != nil {
		return nil, err
	}

	ranges := make([]mergeRange, 0, len(mergeCells))
	for _, mergeCell := range mergeCells {
		startCol, startRow, startErr := excelize.CellNameToCoordinates(mergeCell.GetStartAxis())
		endCol, endRow, endErr := excelize.CellNameToCoordinates(mergeCell.GetEndAxis())
		if startErr != nil || endErr != nil {
			slog.Error("Failed to parse merge coordinates",
				slog.String("start", mergeCell.GetStartAxis()),
				slog.String("end", mergeCell.GetEndAxis()),
				slog.Any("startError", startErr),
				slog.Any("endError", endErr))
			continue
		}

		ranges = append(ranges, mergeRange{
			StartRow: startRow - 1, StartCol: startCol - 1,
			EndRow: endRow - 1, EndCol: endCol - 1,
			Value: mergeCell.GetCellValue(),
		})
	}
	return ranges, nil
}

func buildMergeLookup(ranges []mergeRange) map[[2]int]*mergeRange {
	lookup := make(map[[2]int]*mergeRange)

	for index := range ranges {
		merge := &ranges[index]
		for row := merge.StartRow; row <= merge.EndRow; row++ {
			for column := merge.StartCol; column <= merge.EndCol; column++ {
				lookup[[2]int{row, column}] = merge
			}
		}
	}

	return lookup
}
