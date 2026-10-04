package snapshot

import (
	"fmt"
	"net/http"
)

// Load fetches and decodes scheduling state compatible with the ST importer.
func Load(client *http.Client, url string) (State, error) {
	data, err := Fetch(client, url)
	if err != nil {
		return State{}, err
	}
	response, err := Decode(data)
	if err != nil {
		return State{}, err
	}
	state := response.State
	if state.SchemaVersion != 1 || state.PublicationMode != "ST" {
		return State{}, fmt.Errorf("unsupported snapshot: schema=%d publication mode=%q", state.SchemaVersion, state.PublicationMode)
	}
	if len(state.Blocks) == 0 {
		return State{}, fmt.Errorf("snapshot contains no blocks")
	}
	return state, nil
}
