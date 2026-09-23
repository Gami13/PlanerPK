package extractor

import (
	"errors"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/woodpkrs/PlanerPK/config"
	"golang.org/x/net/html"
)

func parseHTML(t *testing.T, htmlContent string) *html.Node {
	t.Helper()
	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	return doc
}

func TestFindTimetableLink_RealPKStructure(t *testing.T) {
	htmlContent := `
	<html>
	<body>
		<main>
			<article>
				<div>
					<h3>Studia stacjonarne</h3>
					<ul>
						<li><a href="https://it.pk.edu.pl/wp-content/uploads/2026/04/24042026_11_INFORMATYKA-lato-rozklad-STACJONARNE-2025_2026.xls">Kierunek: Informatyka I stopien i II stopien - pobierz (aktualizacja 24-04-2026)</a></li>
						<li><a href="https://it.pk.edu.pl/wp-content/uploads/2026/09/17092026_MATEMATYKA-zima-rozklad-2026_27.xls">Kierunek: Matematyka I i II stopien - pobierz (aktualizacja 17-09-2026)</a></li>
						<li><a href="https://it.pk.edu.pl/wp-content/uploads/2026/09/17092026_MATEMATYKA-STOSOWANA-zima-2026_27.xls">Kierunek: Matematyka Stosowana - pobierz (aktualizacja 17-09-2026)</a></li>
					</ul>
					<h3>Studia niestacjonarne</h3>
					<ul>
						<li><a href="https://it.pk.edu.pl/wp-content/uploads/2026/05/26052026_15_INFORMATYKA-lato-rozklad-NIESTACJONARNE-2025_2026.xls">Kierunek: Informatyka I stopien i II stopien - pobierz (aktualizacja 26-05-2026)</a></li>
					</ul>
				</div>
			</article>
		</main>
	</body>
	</html>
	`
	doc := parseHTML(t, htmlContent)
	expected := "https://it.pk.edu.pl/wp-content/uploads/2026/04/24042026_11_INFORMATYKA-lato-rozklad-STACJONARNE-2025_2026.xls"

	link, err := findTimetableLink(doc, "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link != expected {
		t.Errorf("got %q, want %q", link, expected)
	}
}

func TestFindTimetableLink_ReorderedLinks(t *testing.T) {
	// PK puts Matematyka as li[1] and Informatyka as li[2]
	htmlContent := `
	<html>
	<body>
		<ul>
			<li><a href="https://it.pk.edu.pl/uploads/matematyka.xls">Kierunek: Matematyka</a></li>
			<li><a href="https://it.pk.edu.pl/uploads/informatyka-stacjonarne.xlsx">Kierunek: Informatyka</a></li>
		</ul>
	</body>
	</html>
	`
	doc := parseHTML(t, htmlContent)
	expected := "https://it.pk.edu.pl/uploads/informatyka-stacjonarne.xlsx"

	link, err := findTimetableLink(doc, "https://it.pk.edu.pl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link != expected {
		t.Errorf("got %q, want %q", link, expected)
	}
}

func TestFindTimetableLink_ModifiedDOMWrapper(t *testing.T) {
	// Links inside divs and paragraphs without ul/li
	htmlContent := `
	<html>
	<body>
		<div class="schedule-card">
			<a href="https://it.pk.edu.pl/files/informatyka-2026.xlsx">Pobierz rozkład</a>
		</div>
	</body>
	</html>
	`
	doc := parseHTML(t, htmlContent)
	expected := "https://it.pk.edu.pl/files/informatyka-2026.xlsx"

	link, err := findTimetableLink(doc, "https://it.pk.edu.pl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link != expected {
		t.Errorf("got %q, want %q", link, expected)
	}
}

func TestFindTimetableLink_TitleAttrOrText(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "URL has informatyka, text is generic",
			html: `<a href="https://it.pk.edu.pl/files/informatyka_plan.xls">Pobierz</a>`,
			expected: "https://it.pk.edu.pl/files/informatyka_plan.xls",
		},
		{
			name: "URL is generic, text has Informatyka",
			html: `<a href="https://it.pk.edu.pl/files/plan.xlsx">Informatyka stacjonarna</a>`,
			expected: "https://it.pk.edu.pl/files/plan.xlsx",
		},
		{
			name: "URL and text are generic, title attribute has informatyka",
			html: `<a href="https://it.pk.edu.pl/files/plan.xls" title="Rozkład dla kierunku Informatyka">Pobierz</a>`,
			expected: "https://it.pk.edu.pl/files/plan.xls",
		},
		{
			name: "Uppercase extension .XLSX and case-insensitive check",
			html: `<a href="https://it.pk.edu.pl/files/INFORMATYKA.XLSX">PLAN</a>`,
			expected: "https://it.pk.edu.pl/files/INFORMATYKA.XLSX",
		},
		{
			name: "URL with query string and fragment",
			html: `<a href="https://it.pk.edu.pl/files/informatyka.xlsx?download=true#latest">Pobierz</a>`,
			expected: "https://it.pk.edu.pl/files/informatyka.xlsx?download=true#latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			link, err := findTimetableLink(doc, "https://it.pk.edu.pl")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if link != tt.expected {
				t.Errorf("got %q, want %q", link, tt.expected)
			}
		})
	}
}

func TestFindTimetableLink_ExcludesNiestacjonarne(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{
			name: "URL contains niestacjonarne",
			html: `<a href="https://it.pk.edu.pl/files/informatyka-niestacjonarne.xlsx">Kierunek Informatyka</a>`,
		},
		{
			name: "Text contains niestacjonarne",
			html: `<a href="https://it.pk.edu.pl/files/informatyka.xls">Informatyka studia niestacjonarne</a>`,
		},
		{
			name: "Title attribute contains niestacjonarne",
			html: `<a href="https://it.pk.edu.pl/files/informatyka.xls" title="Informatyka niestacjonarne">Pobierz</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			_, err := findTimetableLink(doc, "https://it.pk.edu.pl")
			if !errors.Is(err, ErrTimetableNotFound) {
				t.Fatalf("expected ErrTimetableNotFound, got %v", err)
			}
		})
	}
}

func TestFindTimetableLink_NonExcelFilesIgnored(t *testing.T) {
	htmlContent := `
	<html>
	<body>
		<a href="https://it.pk.edu.pl/files/informatyka.pdf">Informatyka PDF</a>
		<a href="https://it.pk.edu.pl/files/informatyka.docx">Informatyka Word</a>
		<a href="https://it.pk.edu.pl/files/informatyka.html">Informatyka HTML</a>
	</body>
	</html>
	`
	doc := parseHTML(t, htmlContent)
	_, err := findTimetableLink(doc, "https://it.pk.edu.pl")
	if !errors.Is(err, ErrTimetableNotFound) {
		t.Fatalf("expected ErrTimetableNotFound, got %v", err)
	}
}

func TestFindTimetableLink_RelativeURLResolution(t *testing.T) {
	htmlContent := `
	<a href="/uploads/2026/INFORMATYKA-stacjonarne.xls">Kierunek Informatyka</a>
	`
	doc := parseHTML(t, htmlContent)
	expected := "https://it.pk.edu.pl/uploads/2026/INFORMATYKA-stacjonarne.xls"

	link, err := findTimetableLink(doc, "https://it.pk.edu.pl/studenci/rozklady/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link != expected {
		t.Errorf("got %q, want %q", link, expected)
	}
}

func TestFindTimetableLink_NilDoc(t *testing.T) {
	_, err := findTimetableLink(nil, "https://it.pk.edu.pl")
	if !errors.Is(err, ErrTimetableNotFound) {
		t.Fatalf("expected ErrTimetableNotFound, got %v", err)
	}
}

func TestIsExcelURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://it.pk.edu.pl/file.xls", true},
		{"https://it.pk.edu.pl/file.xlsx", true},
		{"https://it.pk.edu.pl/file.XLS", true},
		{"https://it.pk.edu.pl/file.XLSX", true},
		{"https://it.pk.edu.pl/path/file.xls?download=1", true},
		{"https://it.pk.edu.pl/path/file.xlsx#anchor", true},
		{"file.xlsx", true},
		{"/uploads/file.xls", true},
		{"https://it.pk.edu.pl/file.pdf", false},
		{"https://it.pk.edu.pl/file.xlsx.bak", false},
		{"https://it.pk.edu.pl/file.xls-fake", false},
		{"", false},
		{"   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := isExcelURL(tt.url)
			if got != tt.expected {
				t.Errorf("isExcelURL(%q) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestFetchTimetableUrl_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live test in short mode")
	}

	origURL := config.Config.TimetableURL
	if origURL == "" {
		config.Config.TimetableURL = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
		defer func() { config.Config.TimetableURL = origURL }()
	}

	link, err := fetchTimetableUrl()
	if err != nil {
		t.Fatalf("fetchTimetableUrl failed: %v", err)
	}

	lower := strings.ToLower(link)
	if !strings.HasSuffix(lower, ".xls") && !strings.HasSuffix(lower, ".xlsx") {
		t.Errorf("expected link to end with .xls or .xlsx, got %q", link)
	}
	if !strings.Contains(lower, "informatyka") {
		t.Errorf("expected link to contain informatyka, got %q", link)
	}
	if strings.Contains(lower, "niestacjonarne") {
		t.Errorf("expected link not to contain niestacjonarne, got %q", link)
	}
}
