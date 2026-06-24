package extractor

import (
	"log/slog"
	"strings"

	"github.com/woodpkrs/PlanerPK/global"
	"github.com/woodpkrs/PlanerPK/types"
	"github.com/xuri/excelize/v2"
)

// ExtractLegend scans the sheet for the row containing "Legenda" and "Sale"
// column headers, then parses teacher and location entries from the rows below.
// Both sections share the same rows — teachers on the left, locations in the
// Sale column.
func ExtractLegend(file *excelize.File) *types.Legend {
	rows, err := file.GetRows(global.SHEET_NAME)
	if err != nil {
		slog.Error("Failed to get rows for legend extraction", slog.Any("error", err))
		return &types.Legend{
			Teachers:  make(map[string]types.Teacher),
			Locations: make(map[string]types.Location),
		}
	}

	legend := &types.Legend{
		Teachers:  make(map[string]types.Teacher),
		Locations: make(map[string]types.Location),
	}

	headerRow := -1
	saleCol := -1

	// Find the row containing both "Legenda" and "Sale" as column headers
	for i, row := range rows {
		for colIdx, cell := range row {
			if strings.TrimSpace(cell) == "Legenda" {
				headerRow = i
			}
			if strings.TrimSpace(cell) == "Sale" {
				saleCol = colIdx
			}
		}
		if headerRow != -1 {
			break
		}
	}

	if headerRow == -1 {
		slog.Warn("Legenda section not found in timetable")
		return legend
	}

	parseTeachers(rows, headerRow, legend)

	if saleCol != -1 {
		parseLocations(rows, headerRow, saleCol, legend)
	} else {
		slog.Warn("Sale column not found in timetable header row")
	}

	return legend
}

// parseTeachers extracts teacher entries from rows below the "Legenda" header.
// It dynamically finds the "Imię i nazwisko" and "adres email" columns in the
// header row (startingRow+1), then iterates downward from startingRow+2.
func parseTeachers(rows [][]string, startingRow int, legend *types.Legend) {
	if startingRow+1 >= len(rows) {
		return
	}

	headerRow := rows[startingRow+1]
	nameCol := -1
	emailCol := -1

	for col, cell := range headerRow {
		trimmed := strings.TrimSpace(strings.ToLower(cell))
		if strings.Contains(trimmed, "imię i nazwisko") || strings.Contains(trimmed, "imie i nazwisko") {
			nameCol = col
		}
		if strings.Contains(trimmed, "adres email") || strings.Contains(trimmed, "email") {
			emailCol = col
		}
	}

	if nameCol == -1 {
		slog.Warn("Could not find 'Imię i nazwisko' column in Legenda header")
		return
	}

	emptyCount := 0
	for i := startingRow + 2; i < len(rows); i++ {
		row := rows[i]

		// Find the shorthand: first non-empty cell before nameCol
		shorthand := ""
		for col := 0; col < nameCol && col < len(row); col++ {
			if strings.TrimSpace(row[col]) != "" {
				shorthand = strings.TrimSpace(row[col])
				break
			}
		}

		fullName := ""
		if nameCol < len(row) {
			fullName = strings.TrimSpace(row[nameCol])
		}

		email := ""
		if emailCol >= 0 && emailCol < len(row) {
			email = strings.TrimSpace(row[emailCol])
		}

		if shorthand == "" && fullName == "" {
			emptyCount++
			if emptyCount >= 10 {
				break
			}
			continue
		}
		emptyCount = 0

		if shorthand != "" {
			legend.Teachers[shorthand] = types.Teacher{
				Shorthand: shorthand,
				FullName:  fullName,
				Email:     email,
			}
		}
	}
}

// parseLocations extracts location entries from the Sale column, starting one
// row below the header (which contains "Sale"). Each cell is split on " - " to
// produce shorthand and full name. Compound shorthands like
// "s. 135(L2), 136(L1,) 151(L3)" are expanded into individual locations.
func parseLocations(rows [][]string, headerRow int, saleCol int, legend *types.Legend) {
	emptyCount := 0
	for i := headerRow + 1; i < len(rows); i++ {
		row := rows[i]
		if saleCol >= len(row) {
			emptyCount++
			if emptyCount >= 10 {
				break
			}
			continue
		}

		cellText := strings.TrimSpace(row[saleCol])
		if cellText == "" {
			emptyCount++
			if emptyCount >= 10 {
				break
			}
			continue
		}
		emptyCount = 0

		parts := strings.SplitN(cellText, " - ", 2)
		shorthand := strings.TrimSpace(parts[0])
		fullName := ""
		if len(parts) == 2 {
			fullName = strings.TrimSpace(parts[1])
		}

		for _, sub := range splitLocationShorthand(shorthand) {
			legend.Locations[sub] = types.Location{
				Shorthand: sub,
				FullName:  fullName,
			}
		}
	}
}

// splitLocationShorthand expands a compound shorthand like
// "s. 135(L2), 136(L1,) 151(L3)" into individual entries:
// ["s. 135(L2)", "s. 136(L1,)", "s. 151(L3)"].
//
// It splits on ", " outside parentheses, then further splits on ") "
// to handle cases where sub-locations are space-separated after a closing paren.
func splitLocationShorthand(shorthand string) []string {
	// Strip "s. " prefix; we'll re-add it to each sub-location.
	prefix := ""
	body := shorthand
	if strings.HasPrefix(shorthand, "s. ") {
		prefix = "s. "
		body = strings.TrimPrefix(shorthand, "s. ")
	}

	// Split on ", " that is outside parentheses.
	chunks := splitOutsideParens(body, ", ")

	// Further split each chunk on ") " to handle "136(L1,) 151(L3)".
	var subs []string
	for _, chunk := range chunks {
		parts := splitAfterCloseParen(chunk)
		subs = append(subs, parts...)
	}

	// If only one result and it equals the original body, no splitting needed.
	if len(subs) == 1 {
		return []string{shorthand}
	}

	result := make([]string, 0, len(subs))
	for _, s := range subs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		result = append(result, prefix+s)
	}
	return result
}

// splitOutsideParens splits s on the given separator, but only when the
// separator is not inside parentheses.
func splitOutsideParens(s string, sep string) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		}
		if depth == 0 && i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// splitAfterCloseParen splits "136(L1,) 151(L3)" into ["136(L1,)", "151(L3)"]
// by finding ") " (close paren + space) where the space is followed by more text.
func splitAfterCloseParen(s string) []string {
	var parts []string
	for {
		idx := strings.Index(s, ") ")
		if idx == -1 {
			break
		}
		rest := s[idx+2:]
		if strings.TrimSpace(rest) == "" {
			break
		}
		parts = append(parts, s[:idx+1])
		s = rest
	}
	parts = append(parts, s)
	return parts
}
