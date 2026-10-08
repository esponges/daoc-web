package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMularnGrass checks zone 100's grass: the region's table, every cell's
// group present in it, and the reading of the grass map as ten times the
// group, which puts group 3, the waterside plants, on the lake shore.
func TestMularnGrass(t *testing.T) {
	game := os.Getenv("DAOC_PATH")
	if game == "" {
		game = defaultGamePath
	}
	if _, err := os.Stat(game); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", game, err)
	}
	dir := t.TempDir()
	gr, err := buildGrass(game, 100, 256, dir)
	if err != nil {
		t.Fatal(err)
	}
	if gr.Table != "grass_mid.csv" || len(gr.Sprites) != 64 {
		t.Errorf("table %s with %d sprites, want grass_mid.csv with 64", gr.Table, len(gr.Sprites))
	}
	groups := map[int]string{}
	for _, s := range gr.Sprites {
		groups[s.Group] = s.Name
	}
	cells, err := os.ReadFile(filepath.Join(dir, "grass", "map.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cells) != 256*256*2 {
		t.Fatalf("map is %d bytes", len(cells))
	}
	seen := map[int]int{}
	for i := 0; i < 256*256; i++ {
		seen[int(cells[i*2])]++
	}
	for g := range seen {
		if _, ok := groups[g]; !ok {
			t.Errorf("cells use group %d, which the table does not have", g)
		}
	}
	// Group 3 lies along the lake: its cells' mean height is below the
	// water's 4372, where no other group's is.
	hb, err := os.ReadFile("../../web/data/zone100/heights.u16")
	if err != nil {
		t.Skipf("no converted heightmap: %v", err)
	}
	var sum, n float64
	for i := 0; i < 256*256; i++ {
		if cells[i*2] == 3 {
			sum += float64(uint16(hb[i*2]) | uint16(hb[i*2+1])<<8)
			n++
		}
	}
	if n == 0 || sum/n >= 4372 {
		t.Errorf("group 3 (%s) cells average %.0f high, want below the lake's 4372", groups[3], sum/n)
	}
}
