package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/extractor"
	"github.com/woodpkrs/PlanerPK/snapshot"
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
	data, err := snapshot.Fetch(http.DefaultClient, "https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/snapshot.php")
	if err != nil {
		return err
	}
	response, err := snapshot.Decode(data)
	if err != nil {
		return err
	}
	index := snapshot.NewSubjectIndex(response.State.Blocks)
	legend := extractor.ExtractLegend(file, sheet)
	timetable := extractor.BuildTimetable(groups, file, sheet, legend, nil, layout)
	counts := map[string]int{}
	fmt.Println("status\tgroup\tcategory\tday\tstart\tclass\tblock_ids")
	for _, group := range groups {
		for _, day := range layout.Days {
			for _, class := range timetable[int64(group.Column)][day.WeekDay] {
				candidates := index.Candidates(class)
				status := "ambiguous"
				if len(candidates) == 0 {
					status = "unmatched"
				} else if len(candidates) == 1 {
					status = "unique"
				}
				counts[status]++
				if status == "unique" {
					subject := class.Name
					if subject != candidates[0].Subject {
						subject += " -> " + candidates[0].Subject
					}
					fmt.Printf("unique\t%q\t\t\t\t%q\t%s\n", group.Label, subject+" ("+class.Type+")", candidates[0].ID)
					continue
				}
				var ids []string
				for _, block := range candidates {
					ids = append(ids, block.ID)
				}
				fmt.Printf("%s\t%q\t%q\t%s\t%s\t%q\t%s\n", status, group.Label, group.Category, class.Day, class.StartTime.Format("15:04"), class.String(), strings.Join(ids, ","))
			}
		}
	}
	fmt.Fprintf(os.Stderr, "Snapshot: %d blocks, %d subjects. Excel records across all columns (including merged supplemental classes): unique=%d ambiguous=%d unmatched=%d\n", len(response.State.Blocks), len(index), counts["unique"], counts["ambiguous"], counts["unmatched"])
	return nil
}
