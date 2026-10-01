package extractor

import (
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func workbookWithSheets(t *testing.T, sheets ...string) *excelize.File {
	t.Helper()

	file := excelize.NewFile()
	if len(sheets) == 0 {
		return file
	}
	if err := file.SetSheetName("Sheet1", sheets[0]); err != nil {
		t.Fatalf("SetSheetName() error = %v", err)
	}
	for _, sheet := range sheets[1:] {
		if _, err := file.NewSheet(sheet); err != nil {
			t.Fatalf("NewSheet(%q) error = %v", sheet, err)
		}
	}
	return file
}

func TestResolveTimetableSheetSelectsHighestAcademicYear(t *testing.T) {
	tests := []struct {
		name   string
		sheets []string
		want   string
	}{
		{
			name:   "single academic year sheet",
			sheets: []string{"Legenda", "2025_2026", "Dane"},
			want:   "2025_2026",
		},
		{
			name:   "multiple academic year sheets picks highest",
			sheets: []string{"Legenda", "2025_2026", "2026_2027", "Dane"},
			want:   "2026_2027",
		},
		{
			name:   "out of order sheets still picks highest",
			sheets: []string{"2026_2027", "Legenda", "2024_2025", "2025_2026"},
			want:   "2026_2027",
		},
		{
			name:   "only one valid sheet among junk",
			sheets: []string{"2025_2028", "2025-2026", "2025_2026"},
			want:   "2025_2026",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := workbookWithSheets(t, test.sheets...)
			defer file.Close()

			got, err := ResolveTimetableSheet(file)
			if err != nil {
				t.Fatalf("ResolveTimetableSheet() error = %v", err)
			}
			if got != test.want {
				t.Errorf("ResolveTimetableSheet() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestResolveTimetableSheetRejectsNoAcademicYearSheet(t *testing.T) {
	file := workbookWithSheets(t, "Legenda", "2025_2028", "2025-2026", "2025_27")
	defer file.Close()

	_, err := ResolveTimetableSheet(file)
	if err == nil {
		t.Fatal("ResolveTimetableSheet() error = nil, want missing-sheet error")
	}
	for _, want := range []string{"Legenda", "2025_2028", "2025-2026", "2025_27"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
}


func TestExtractLegendUsesResolvedSheet(t *testing.T) {
	file := workbookWithSheets(t, "Default", "2025_2026")
	defer file.Close()

	if err := file.SetCellValue("2025_2026", "A1", "Legenda"); err != nil {
		t.Fatalf("SetCellValue() error = %v", err)
	}
	if err := file.SetCellValue("2025_2026", "B1", "Sale"); err != nil {
		t.Fatalf("SetCellValue() error = %v", err)
	}
	if err := file.SetCellValue("2025_2026", "B2", "Imię i nazwisko"); err != nil {
		t.Fatalf("SetCellValue() error = %v", err)
	}
	if err := file.SetCellValue("2025_2026", "C2", "adres email"); err != nil {
		t.Fatalf("SetCellValue() error = %v", err)
	}
	if err := file.SetCellValue("2025_2026", "A3", "JD"); err != nil {
		t.Fatalf("SetCellValue() error = %v", err)
	}

	got := ExtractLegend(file, "2025_2026")
	if _, ok := got.Teachers["JD"]; !ok {
		t.Fatalf("ExtractLegend() did not read the explicitly selected sheet: %#v", got.Teachers)
	}
}
