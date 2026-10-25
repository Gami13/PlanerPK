package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woodpkrs/PlanerPK/cache"
	"github.com/woodpkrs/PlanerPK/config"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
)

func TestUnchangedExcelStillFetchesSnapshotAndRetainsCacheOnFailure(t *testing.T) {
	previousConfig, previousClient := config.Config, http.DefaultClient
	t.Cleanup(func() { config.Config, http.DefaultClient = previousConfig, previousClient })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("unchanged Excel must not be downloaded: %s", r.Method)
		}
		w.Header().Set("Last-Modified", "unchanged")
		w.Header().Set("Content-Length", "4")
	}))
	defer server.Close()
	config.Config.TimetableURL = server.URL
	config.Config.TimetableStats = filepath.Join(t.TempDir(), "stats.json")
	if err := os.WriteFile(config.Config.TimetableStats, []byte(`{"LastModified":"unchanged","ContentLength":"4"}`), 0600); err != nil {
		t.Fatal(err)
	}
	requests := 0
	http.DefaultClient = &http.Client{Transport: refreshTransport(func(r *http.Request) (*http.Response, error) {
		requests++
		if !strings.HasSuffix(r.URL.Path, "/snapshot.php") {
			t.Errorf("unexpected request: %s", r.URL)
		}
		return &http.Response{StatusCode: 500, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
	})}
	var cached cache.Cache
	state := &snapshot.State{Blocks: []snapshot.Block{{ID: "previous"}}}
	timetable := &types.Timetable{}
	cached.SetSnapshot(state)
	cached.SetTimetable(timetable)
	cached.SetTimetableHash("previous-workbook")
	if _, err := updateTimetableAndCache(nil, &cached); err == nil {
		t.Fatal("expected snapshot refresh failure")
	}
	if requests != 1 || cached.Snapshot() != state || cached.Timetable() != timetable || cached.TimetableHash() != "previous-workbook" {
		t.Fatal("snapshot was not fetched or failed refresh replaced the cache")
	}
	for _, pending := range []bool{false, true} {
		if got := updateAllTimetables(nil, &cached, pending); got != pending {
			t.Fatal("failed refresh must preserve the pending batch state")
		}
	}
}

type refreshTransport func(*http.Request) (*http.Response, error)

func (transport refreshTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}
