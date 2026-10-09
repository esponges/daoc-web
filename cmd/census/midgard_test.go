package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"daocweb/internal/zone"
)

// midgard is every Midgard outdoor zone in the install: the classic realm,
// the Shrouded Isles and the realm's four frontier zones.
var midgard = []int{
	100, 101, 102, 103, 104, 105, 106, 107, 108, 111, 112, 113, 115, 116,
	151, 152, 153, 154, 155, 156, 158,
	167, 168, 169, 170,
}

// cannotDraw lists placed models known not to draw, by zone, each with
// why. Every Midgard placement draws today, so it is empty; a model goes
// here only with a reason checked against the game's files.
var cannotDraw = map[int]map[string]string{}

// midgardProblems says what keeps a zone from looking right, measured
// against the baseline row: nothing for a zone that is complete.
func midgardProblems(got, base *ZoneRow) []string {
	var out []string
	if got.Terrain != "ok" {
		out = append(out, "terrain: "+got.Terrain)
	}
	if got.Detail != "ok" {
		out = append(out, "ground detail: "+got.Detail)
	}
	if got.Grass != "ok" && got.Grass != "no grass map" {
		out = append(out, "grass: "+got.Grass)
	}
	if base != nil && base.Grass == "ok" && got.Grass != "ok" {
		out = append(out, "grass lost: "+got.Grass)
	}
	for _, f := range got.Failing {
		if _, ok := cannotDraw[got.Num][f.File]; !ok {
			out = append(out, fmt.Sprintf("%s does not convert (%d placements, %s)", f.File, f.Placements, f.Cause))
		}
	}
	if base != nil && got.Converting < base.Converting {
		out = append(out, fmt.Sprintf("%d placements convert, down from %d", got.Converting, base.Converting))
	}
	return out
}

// TestMidgardLooksRight dry-runs every Midgard zone and fails if one has
// lost its terrain, ground detail, grass or any placement.
func TestMidgardLooksRight(t *testing.T) {
	if testing.Short() {
		t.Skip("reads 25 zones from the install; skipped under -short")
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
	baseRow := map[int]*ZoneRow{}
	for _, r := range base.Zones.Rows {
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

	cache := &bakeCache{done: map[string]*bakeResult{}}
	rows := make([]*ZoneRow, len(midgard))
	parallel(len(midgard), runtime.NumCPU(), func(i int) {
		if in, ok := info[midgard[i]]; ok && present(game, in.Num) {
			rows[i] = dryRun(game, in, cache)
		}
	})
	for i, r := range rows {
		if r == nil {
			t.Logf("zone %d is not in this install", midgard[i])
			continue
		}
		for _, p := range midgardProblems(r, baseRow[r.Num]) {
			t.Errorf("zone %03d %s: %s", r.Num, r.Name, p)
		}
	}
}

// The check itself: a zone whose ground detail is skipped, or which loses
// a placement or its grass, is reported.
func TestMidgardProblemsReported(t *testing.T) {
	ok := &ZoneRow{Num: 100, Terrain: "ok", Detail: "ok", Grass: "ok", Placements: 10, Converting: 10}
	if p := midgardProblems(ok, ok); len(p) != 0 {
		t.Fatalf("a complete zone reports %v", p)
	}
	bad := *ok
	bad.Detail = `"tex01-00.dds" not found in tex102.mpk`
	if p := midgardProblems(&bad, ok); len(p) != 1 {
		t.Errorf("skipped detail reports %v", p)
	}
	bad = *ok
	bad.Grass = "grass maps 512x512 and 512x512, heightmap 256"
	if p := midgardProblems(&bad, ok); len(p) != 2 {
		t.Errorf("lost grass reports %v", p)
	}
	bad = *ok
	bad.Converting = 9
	bad.Failing = []ModelFail{{File: "Ncarcass.nif", Placements: 1, Cause: "corrupt data in NiLODNode"}}
	if p := midgardProblems(&bad, ok); len(p) != 2 {
		t.Errorf("a failing model reports %v", p)
	}
	none := *ok
	none.Grass = "no grass map"
	if p := midgardProblems(&none, &none); len(p) != 0 {
		t.Errorf("a zone with no grass map reports %v", p)
	}
}
