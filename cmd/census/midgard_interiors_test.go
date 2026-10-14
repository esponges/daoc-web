package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"daocweb/internal/zone"
)

// midgardInteriors is every Midgard city and dungeon in the install:
// Jordheim and the twelve dungeons.
var midgardInteriors = []int{120, 125, 126, 127, 128, 129, 149, 150, 160, 161, 188, 189, 226}

// interiorProblems says what keeps an interior from looking right,
// measured against the baseline row: nothing for one that is as good. A
// transform that moves more than a cell (128 units), or puts fewer exits
// and lights on the geometry, is reported.
func interiorProblems(got, base *InteriorRow) []string {
	var out []string
	if got.Layout != "ok" {
		return []string{"layout: " + got.Layout}
	}
	for _, f := range got.Failing {
		out = append(out, fmt.Sprintf("%s does not convert (%d placements, %s)", f.File, f.Placements, f.Cause))
	}
	if base == nil {
		return out
	}
	if got.Converting < base.Converting {
		out = append(out, fmt.Sprintf("%d placements convert, down from %d", got.Converting, base.Converting))
	}
	for k := 0; k < 3; k++ {
		if math.Abs(got.Origin[k]-base.Origin[k]) > 128 {
			out = append(out, fmt.Sprintf("transform origin %v, was %v", got.Origin, base.Origin))
			break
		}
	}
	if got.OnSurface < base.OnSurface {
		out = append(out, fmt.Sprintf("%d of %d exits and lights on the geometry, down from %d", got.OnSurface, got.Points, base.OnSurface))
	}
	return out
}

// TestMidgardInteriorsLookRight dry-runs every Midgard interior and fails
// if one has lost a placement or its transform.
func TestMidgardInteriorsLookRight(t *testing.T) {
	if testing.Short() {
		t.Skip("reads 13 interiors from the install; skipped under -short")
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
	baseRow := map[int]*InteriorRow{}
	for _, r := range base.Zones.Interiors {
		baseRow[r.Num] = r
	}
	list, err := zone.List(game)
	if err != nil {
		t.Fatal(err)
	}
	info := map[int]zone.Info{}
	for _, z := range list {
		info[z.Num] = z
	}
	rows := make([]*InteriorRow, len(midgardInteriors))
	parallel(len(midgardInteriors), runtime.NumCPU(), func(i int) {
		if in, ok := info[midgardInteriors[i]]; ok && present(game, in.Num) {
			rows[i] = dryRunInterior(game, in)
		}
	})
	for i, r := range rows {
		if r == nil {
			t.Logf("zone %d is not in this install", midgardInteriors[i])
			continue
		}
		if baseRow[r.Num] == nil {
			t.Errorf("zone %03d %s: not in the baseline census", r.Num, r.Name)
		}
		for _, p := range interiorProblems(r, baseRow[r.Num]) {
			t.Errorf("zone %03d %s: %s", r.Num, r.Name, p)
		}
	}
}

// The check itself: an interior whose transform is forced off, or which
// loses a placement, is reported.
func TestInteriorProblemsReported(t *testing.T) {
	ok := &InteriorRow{Num: 125, Layout: "ok", Placements: 196, Converting: 196,
		Origin: [3]float64{23584, 23904, 16064}, OnSurface: 80, Points: 84}
	if p := interiorProblems(ok, ok); len(p) != 0 {
		t.Fatalf("a complete interior reports %v", p)
	}
	off := *ok
	off.Origin[0] += 512
	off.OnSurface = 12
	if p := interiorProblems(&off, ok); len(p) != 2 {
		t.Errorf("a transform forced off reports %v", p)
	}
	bad := *ok
	bad.Converting = 195
	bad.Failing = []ModelFail{{File: "SPDpillar.nif", Placements: 1, Cause: "corrupt data"}}
	if p := interiorProblems(&bad, ok); len(p) != 2 {
		t.Errorf("a failing model reports %v", p)
	}
	unread := *ok
	unread.Layout = "dungeon.place: no such file"
	if p := interiorProblems(&unread, ok); len(p) != 1 {
		t.Errorf("an unread layout reports %v", p)
	}
}
