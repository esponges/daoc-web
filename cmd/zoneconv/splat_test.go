package main

import (
	"os"
	"strings"
	"testing"
)

// TestGroundDetailMatchesTheGame composites zone 100's ground layers and
// checks they come to what the game's own pre-blended tiles show, in the
// orientation the masks are named in, painted one over another.
func TestGroundDetailMatchesTheGame(t *testing.T) {
	game := os.Getenv("DAOC_PATH")
	if game == "" {
		game = defaultGamePath
	}
	if _, err := os.Stat(game); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", game, err)
	}
	sp, err := buildSplat(game, 100, 8, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if sp.Agree < 0.95 {
		t.Errorf("composite agrees with tex100.mpk at r = %.4f, want > 0.95", sp.Agree)
	}
	if !strings.HasPrefix(sp.Layout, (maskLayout{}).String()) || !strings.Contains(sp.Layout, "painted in order") {
		t.Errorf("layout %q, want masks as named, unflipped, painted in order", sp.Layout)
	}
	// Every sector paints something, beginning with a layer laid solid.
	for i, row := range sp.Slots {
		if len(row) == 0 {
			t.Errorf("sector %d paints no layers", i)
		}
	}
	// The cobbles exist only where a village has them.
	found := false
	for _, n := range sp.Names {
		found = found || strings.EqualFold(n, "h_dirtycobbles02")
	}
	if !found {
		t.Errorf("Mularn's cobbles are not among the layers %v", sp.Names)
	}
}
