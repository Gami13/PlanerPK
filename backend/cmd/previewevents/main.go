package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/woodpkrs/PlanerPK/calendar"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
	gcal "google.golang.org/api/calendar/v3"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run reads captured inputs and prints a full offline diff, without live services.
func run(in io.Reader, out io.Writer) error {
	var input struct {
		Timetable     types.GroupTimetable   `json:"timetable"`
		Snapshot      *snapshot.State        `json:"snapshot"`
		Overrides     map[string]db.Override `json:"overrides"`
		CurrentEvents []*gcal.Event          `json:"currentEvents"`
	}
	if err := json.NewDecoder(in).Decode(&input); err != nil {
		return fmt.Errorf("read preview inputs: %w", err)
	}
	diff, err := calendar.BuildTimetableDiff(input.Timetable, input.Snapshot, input.Overrides, nil, input.CurrentEvents)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(diff)
}
