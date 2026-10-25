package extractor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/woodpkrs/PlanerPK/config"
)

func TestCurrentTimetableHash(t *testing.T) {
	previous := config.Config
	t.Cleanup(func() { config.Config = previous })
	config.Config.CurrentTimetableDir = t.TempDir()
	if _, err := CurrentTimetableHash(); err == nil {
		t.Fatal("missing workbook should fail")
	}
	path := filepath.Join(config.Config.CurrentTimetableDir, "timetable.xlsx")
	var first string
	for i, content := range []string{"first", "first", "other"} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		hash, err := CurrentTimetableHash()
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = hash
		} else if (hash == first) != (i == 1) {
			t.Fatalf("incorrect content comparison for %q", content)
		}
	}
}
