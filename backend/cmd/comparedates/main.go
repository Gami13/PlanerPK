package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/extractor"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "../data"
	}
	config.Config.CurrentTimetableDir = filepath.Join(dataDir, "current_timetable")
	file, err := extractor.OpenCurrentTimetableFile()
	if err != nil {
		return err
	}
	defer extractor.CloseFile(file)
	sheet, err := extractor.ResolveTimetableSheet(file)
	if err != nil {
		return err
	}
	layout, err := extractor.DiscoverTimetableLayout(file, sheet)
	if err != nil {
		return err
	}
	groups, err := extractor.BuildAvailableGroups(file, sheet, layout)
	if err != nil {
		return err
	}
	state, err := snapshot.Load(http.DefaultClient, "https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/snapshot.php")
	if err != nil {
		return err
	}
	blocks := make(map[string]*snapshot.Block, len(state.Blocks))
	for i := range state.Blocks {
		blocks[state.Blocks[i].ID] = &state.Blocks[i]
	}
	legend := extractor.ExtractLegend(file, sheet)
	timetable := extractor.BuildTimetable(groups, file, sheet, legend, nil, layout, snapshot.NewSubjectIndex(state.Blocks))
	fmt.Fprintln(os.Stderr, "Old: next weekday at run time. New: all snapshot class dates, with closures. Process timezone; no user overrides, deferred date exceptions, or calendar writes.")
	for _, group := range groups {
		for _, day := range layout.Days {
			for _, class := range timetable[int64(group.Column)][day.WeekDay] {
				if err := writeComparison(os.Stdout, group.Label, class, blocks[class.SnapshotBlockID], state.AcademicCalendar); err != nil {
					return fmt.Errorf("group %s, class %s: %w", group.Label, class.Name, err)
				}
			}
		}
	}
	return nil
}

func writeComparison(out io.Writer, group string, class types.Class, block *snapshot.Block, calendar snapshot.AcademicCalendar) error {
	newClasses := "    unresolved (no unique snapshot match)"
	if block != nil {
		occurrences, err := block.Occurrences(calendar)
		if err != nil {
			return err
		}
		newClasses = "    omitted (no scheduled classes)"
		var dates []string
		for _, occurrence := range occurrences {
			dates = append(dates, "    "+formatClassTimes(occurrence.Start, occurrence.End))
		}
		if len(dates) > 0 {
			newClasses = strings.Join(dates, "\n")
		}
	}
	old := class.ToEvent(nil, nil)
	oldStart, err := time.Parse(time.RFC3339, old.Start.DateTime)
	if err != nil {
		return err
	}
	oldEnd, err := time.Parse(time.RFC3339, old.End.DateTime)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s | %s (%s)\n  Block: %s | Audience: %s\n  Old: %s\n  New:\n%s\n\n",
		group, class.Name, class.Type, class.SnapshotBlockID, class.Audience, formatClassTimes(oldStart, oldEnd), newClasses)
	return err
}

func formatClassTimes(start, end time.Time) string {
	endFormat := "15:04"
	if start.Format(time.DateOnly) != end.Format(time.DateOnly) {
		endFormat = "2006-01-02 15:04"
	}
	return start.Format("Mon 2006-01-02 15:04") + " – " + end.Format(endFormat)
}
