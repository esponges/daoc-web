package main

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"daocweb/internal/mpak"
)

func gamePath(t *testing.T) string {
	t.Helper()
	p := os.Getenv("DAOC_PATH")
	if p == "" {
		p = defaultGame
	}
	if _, err := os.Stat(p); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", p, err)
	}
	return p
}

func zoneCSVs(t *testing.T) ([]byte, []byte) {
	t.Helper()
	path := filepath.Join(gamePath(t), "zones", "zone100", "dat100.mpk")
	a, err := mpak.Open(path)
	if err != nil {
		t.Skipf("no %s: %v", path, err)
	}
	nifs, err := readEntry(a, "nifs.csv")
	if err != nil {
		t.Fatalf("nifs.csv: %v", err)
	}
	fix, err := readEntry(a, "fixtures.csv")
	if err != nil {
		t.Fatalf("fixtures.csv: %v", err)
	}
	return nifs, fix
}

func TestFixturesParse(t *testing.T) {
	nifsCSV, fixCSV := zoneCSVs(t)
	models, err := parseNifs(nifsCSV)
	if err != nil {
		t.Fatalf("parseNifs: %v", err)
	}
	fixtures, err := parseFixtures(fixCSV)
	if err != nil {
		t.Fatalf("parseFixtures: %v", err)
	}
	if len(models) != 81 {
		t.Errorf("model definitions = %d, want 81", len(models))
	}
	if len(fixtures) != 1006 {
		t.Errorf("fixtures = %d, want 1006", len(fixtures))
	}

	// Every placement must name a model the zone defines, sit inside the
	// zone, and carry a sane scale. A column misread shows up here first.
	const extent = 65536
	for i, f := range fixtures {
		if _, ok := models[f.NifID]; !ok {
			t.Errorf("fixture %d references undefined model %d", i, f.NifID)
		}
		if f.X < 0 || f.X > extent || f.Y < 0 || f.Y > extent {
			t.Errorf("fixture %d is outside the zone at (%g, %g)", i, f.X, f.Y)
		}
		if f.Scale < 0.1 || f.Scale > 10 {
			t.Errorf("fixture %d has implausible scale %g", i, f.Scale)
		}
		if math.Abs(float64(f.Yaw)) > 2*math.Pi+0.001 {
			t.Errorf("fixture %d has yaw %g, outside one turn", i, f.Yaw)
		}
	}

	// The pine is the zone's signature: more than half of everything placed.
	pine := 0
	for _, f := range fixtures {
		if m, ok := models[f.NifID]; ok && m.File == "npintre1.nif" {
			pine++
		}
	}
	if pine != 551 {
		t.Errorf("pine placements = %d, want 551", pine)
	}
	t.Logf("%d fixtures over %d models, %d of them pines", len(fixtures), len(models), pine)
}

// The strongest check available on the terrain pipeline, and it comes free
// with the scenery.
//
// fixtures.csv carries an absolute world height for every prop, authored by
// the game's own designers against the game's own terrain. This project's
// heightmap is derived independently, from two 8-bit PCX rasters combined by
// factors read out of SECTOR.DAT. Nothing links the two. If the heightmap were
// scaled wrongly, offset wrongly, or flipped on either axis, a thousand trees
// would float or sink and this would say so immediately.
func TestFixtureHeightsMatchTheHeightmap(t *testing.T) {
	_, fixCSV := zoneCSVs(t)
	fixtures, err := parseFixtures(fixCSV)
	if err != nil {
		t.Fatalf("parseFixtures: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "data", "zone100", "heights.u16"))
	if err != nil {
		t.Skipf("no converted heightmap; run zoneconv first (%v)", err)
	}
	const grid = 256
	if len(raw) != grid*grid*2 {
		t.Fatalf("heights.u16 is %d bytes, want %d", len(raw), grid*grid*2)
	}
	height := func(cx, cy int) float64 {
		if cx < 0 {
			cx = 0
		}
		if cy < 0 {
			cy = 0
		}
		if cx > grid-1 {
			cx = grid - 1
		}
		if cy > grid-1 {
			cy = grid - 1
		}
		return float64(binary.LittleEndian.Uint16(raw[(cy*grid+cx)*2:]))
	}
	// Bilinear, matching how the viewer puts the character on the ground.
	ground := func(wx, wy float64) float64 {
		const cell = 256.0
		gx, gy := wx/cell, wy/cell
		x0, y0 := int(math.Floor(gx)), int(math.Floor(gy))
		fx, fy := gx-float64(x0), gy-float64(y0)
		h00, h10 := height(x0, y0), height(x0+1, y0)
		h01, h11 := height(x0, y0+1), height(x0+1, y0+1)
		return (h00*(1-fx)+h10*fx)*(1-fy) + (h01*(1-fx)+h11*fx)*fy
	}

	var worst, sum float64
	worstAt := -1
	exact, n := 0, 0
	for i, f := range fixtures {
		// Props on steep ground or deliberately sunk are not evidence
		// either way; the great majority sit flat on the terrain.
		g := ground(float64(f.X), float64(f.Y))
		d := math.Abs(g - float64(f.Z))
		n++
		sum += d
		if d < 0.5 {
			exact++
		}
		if d > worst {
			worst, worstAt = d, i
		}
	}
	mean := sum / float64(n)
	pct := 100 * float64(exact) / float64(n)
	t.Logf("%d fixtures: %.1f%% match the heightmap exactly, mean error %.2f units, worst %.1f (fixture %d)",
		n, pct, mean, worst, worstAt)

	// A unit is about an inch and the character is 71.8 units tall, so an
	// average error of a few units is nothing. A wrong scale factor or a
	// flipped axis would put this in the hundreds or thousands.
	if mean > 20 {
		t.Errorf("mean height disagreement %.2f units is too large; the heightmap and the fixtures disagree", mean)
	}
	if pct < 80 {
		t.Errorf("only %.1f%% of fixtures sit exactly on the derived terrain, want 80%%+", pct)
	}
}
