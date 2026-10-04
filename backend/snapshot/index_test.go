package snapshot

import (
	"os"
	"reflect"
	"testing"
)

func TestIndexBySubjectRetainsCandidates(t *testing.T) {
	day, start := "MON", 480
	blocks := []Block{
		{ID: "lecture", Subject: "Subject", Activity: "W", Day: &day, Start: &start},
		{ID: "lab", Subject: "Subject", Activity: "L", Day: &day, Start: &start},
		{ID: "parked", Subject: "Subject"},
		{ID: "other", Subject: "Other"},
		{ID: "distinct", Subject: "subject"},
	}
	got := IndexBySubject(blocks)
	want := map[string][]Block{
		"Subject": blocks[:3], "Other": blocks[3:4], "subject": blocks[4:],
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("index = %#v; want %#v", got, want)
	}
	if len(IndexBySubject(nil)) != 0 {
		t.Fatal("empty input must produce an empty index")
	}
}

func TestIndexBySubjectPreservesFixtureBlocks(t *testing.T) {
	data, err := os.ReadFile("testdata/winter-2026-27.json")
	if err != nil {
		t.Fatal(err)
	}
	response, err := DecodeSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	var indexed []Block
	for subject, blocks := range IndexBySubject(response.State.Blocks) {
		for _, block := range blocks {
			if block.Subject != subject {
				t.Fatalf("block %q indexed under %q", block.ID, subject)
			}
		}
		indexed = append(indexed, blocks...)
	}
	want := make(map[string]Block)
	for _, block := range response.State.Blocks {
		want[block.ID] = block
	}
	got := make(map[string]Block)
	for _, block := range indexed {
		got[block.ID] = block
	}
	if len(indexed) != len(response.State.Blocks) || !reflect.DeepEqual(got, want) {
		t.Fatal("index lost or changed fixture blocks")
	}
}
