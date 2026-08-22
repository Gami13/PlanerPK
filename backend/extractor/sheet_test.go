package extractor

import (
	"strings"
	"testing"
	"time"

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

func TestResolveTimetableSheetSelectsAcademicYearByCalendarHalf(t *testing.T) {
	file := workbookWithSheets(t, "Legenda", "2025_2026", "2026_2027", "Dane")
	defer file.Close()

	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{name: "January", now: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), want: "2025_2026"},
		{name: "July 1", now: time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC), want: "2025_2026"},
		{name: "September 14", now: time.Date(2026, time.September, 14, 23, 59, 0, 0, time.UTC), want: "2025_2026"},
		{name: "September 15", now: time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC), want: "2026_2027"},
		{name: "December", now: time.Date(2026, time.December, 31, 23, 59, 0, 0, time.UTC), want: "2026_2027"},
		{name: "January next year", now: time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC), want: "2026_2027"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveTimetableSheet(file, test.now)
			if err != nil {
				t.Fatalf("ResolveTimetableSheet() error = %v", err)
			}
			if got != test.want {
				t.Errorf("ResolveTimetableSheet() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestResolveTimetableSheetRejectsMissingExpectedSheet(t *testing.T) {
	file := workbookWithSheets(t, "Legenda", "2025_2028", "2025-2026", "2025_27")
	defer file.Close()

	_, err := ResolveTimetableSheet(file, time.Date(2026, time.June, 30, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("ResolveTimetableSheet() error = nil, want missing-sheet error")
	}
	for _, want := range []string{"2025_2026", "Legenda", "2025_2028", "2025-2026", "2025_27"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
}

func TestIsAcademicYearSheetName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "2025_2026", want: true},
		{name: "1999_2000", want: true},
		{name: "2025_2028", want: false},
		{name: "2025-2026", want: false},
		{name: "2025_27", want: false},
		{name: "x2025_2026", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isAcademicYearSheetName(test.name); got != test.want {
				t.Errorf("isAcademicYearSheetName(%q) = %v, want %v", test.name, got, test.want)
			}
		})
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
