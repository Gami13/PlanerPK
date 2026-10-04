package main

import (
	"encoding/json"
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
	data, err := snapshot.Fetch(http.DefaultClient, "https://ii.pk.edu.pl/~fkruzel/fk-planer-lti/public/plan-st/snapshot.php")
	if err != nil {
		return err
	}
	response, err := snapshot.DecodeSnapshot(data)
	if err != nil {
		return err
	}
	index := snapshot.IndexBySubject(response.State.Blocks)
	fmt.Fprintf(os.Stderr, "%d blocks indexed under %d subjects\n", len(response.State.Blocks), len(index))
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(index)
}
