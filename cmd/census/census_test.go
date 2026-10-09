package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestNoClassRegresses runs the census and compares it with the committed
// baseline, docs/census/census.json: no class may read fewer items than it
// records, and no fewer outdoor placements may convert. Improvements are
// recorded by re-running the census and committing the new report.
func TestNoClassRegresses(t *testing.T) {
	if testing.Short() {
		t.Skip("the census reads the whole install; skipped under -short")
	}
	game := os.Getenv("DAOC_PATH")
	if game == "" {
		game = defaultGame
	}
	if _, err := os.Stat(game); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", game, err)
	}
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "census", "census.json"))
	if err != nil {
		t.Fatalf("no baseline: %v", err)
	}
	var base Report
	if err := json.Unmarshal(raw, &base); err != nil {
		t.Fatalf("baseline: %v", err)
	}

	now, err := Run(game, runtime.NumCPU())
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]*Class{}
	for _, c := range now.Classes {
		got[c.Name] = c
	}
	for _, b := range base.Classes {
		c := got[b.Name]
		if c == nil {
			t.Errorf("%s: in the baseline but no longer counted", b.Name)
			continue
		}
		if c.Passed < b.Passed {
			t.Errorf("%s: %d read, down from %d in the baseline", b.Name, c.Passed, b.Passed)
		} else if c.Passed > b.Passed {
			t.Logf("%s: %d read, up from %d; re-run the census to record it", b.Name, c.Passed, b.Passed)
		}
	}
	if o, b := now.Zones.Outdoor.Converting, base.Zones.Outdoor.Converting; o < b {
		t.Errorf("outdoor placements: %d convert, down from %d in the baseline", o, b)
	}
}
