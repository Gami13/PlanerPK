package main

import (
	"fmt"
	"net/http"
	"os"

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
	fmt.Fprintln(os.Stderr, "Base meeting dates only; holidays, replacement days, and date exceptions are not applied yet. Times use the process timezone.")
	fmt.Println("block_id\tsubject\tactivity\tdate\tstart\tend")
	meetings, omitted := 0, 0
	for _, block := range state.Blocks {
		occurrences, err := block.Occurrences(state.AcademicCalendar)
		if err != nil {
			return fmt.Errorf("block %s (%s): %w", block.ID, block.Subject, err)
		}
		if len(occurrences) == 0 {
			omitted++
		}
		for _, occurrence := range occurrences {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", block.ID, block.Subject, block.Activity,
				occurrence.Start.Format("2006-01-02"), occurrence.Start.Format("15:04"), occurrence.End.Format("15:04"))
			meetings++
		}
	}
	fmt.Fprintf(os.Stderr, "%d blocks, %d meetings, %d blocks without meetings\n", len(state.Blocks), meetings, omitted)
	return nil
}
