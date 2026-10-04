package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/woodpkrs/PlanerPK/snapshot"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	state, err := snapshot.Load(http.DefaultClient, "https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/snapshot.php")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Full-day closures applied; partial closures, replacement days, and date exceptions are not applied yet. Times use the process timezone.")
	fmt.Println("block_id\tsubject\tactivity\tdate\tstart\tend")
	meetings, omitted := 0, 0
	var excluded []string
	baseCalendar := state.AcademicCalendar
	baseCalendar.Closures = nil
	for _, block := range state.Blocks {
		occurrences, err := block.Occurrences(state.AcademicCalendar)
		if err != nil {
			return fmt.Errorf("block %s (%s): %w", block.ID, block.Subject, err)
		}
		if len(occurrences) == 0 {
			omitted++
		}
		included := make(map[time.Time]bool, len(occurrences))
		for _, occurrence := range occurrences {
			included[occurrence.Start] = true
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", block.ID, block.Subject, block.Activity,
				occurrence.Start.Format("2006-01-02"), occurrence.Start.Format("15:04"), occurrence.End.Format("15:04"))
			meetings++
		}
		base, err := block.Occurrences(baseCalendar)
		if err != nil {
			return fmt.Errorf("block %s (%s): %w", block.ID, block.Subject, err)
		}
		for _, occurrence := range base {
			if !included[occurrence.Start] {
				excluded = append(excluded, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", block.ID, block.Subject, block.Activity,
					occurrence.Start.Format(time.DateOnly), occurrence.Start.Format("15:04"), occurrence.End.Format("15:04")))
			}
		}
	}
	fmt.Println("\nExcluded classes (full-day closures):")
	fmt.Println("block_id\tsubject\tactivity\tdate\tstart\tend")
	for _, row := range excluded {
		fmt.Println(row)
	}
	fmt.Fprintf(os.Stderr, "%d blocks, %d scheduled classes, %d excluded classes, %d blocks without scheduled classes\n", len(state.Blocks), meetings, len(excluded), omitted)
	return nil
}
