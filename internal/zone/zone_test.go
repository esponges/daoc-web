package zone

import (
	"math"
	"os"
	"testing"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

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

func zone100(t *testing.T) *Zone {
	t.Helper()
	z, err := Open(gamePath(t), 100)
	if err != nil {
		t.Skipf("no zone 100: %v", err)
	}
	return z
}

func zoneCSVs(t *testing.T) (map[int]ModelDef, []Fixture) {
	t.Helper()
	z := zone100(t)
	models, err := z.Models()
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := z.Fixtures()
	if err != nil {
		t.Fatal(err)
	}
	return models, fixtures
}

func TestFixturesParse(t *testing.T) {
	models, fixtures := zoneCSVs(t)
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
	_, fixtures := zoneCSVs(t)
	z := zone100(t)
	sec, err := z.Sector()
	if err != nil {
		t.Fatal(err)
	}
	ter, err := z.Terrain(sec)
	if err != nil {
		t.Fatal(err)
	}
	grid := ter.Grid
	if grid != 256 {
		t.Fatalf("grid %d, want 256", grid)
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
		return float64(ter.Heights[cy*grid+cx])
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

// The older fixtures table, which 54 outdoor zones use: no axis-angle
// columns, and anything on the ground written at Z 0 with the Ground flag.
// Ruins of Atlantis places 1364 fixtures that way; read, every one has a
// yaw within a half turn and, once placed, a height from the terrain.
func TestOlderFixturesTable(t *testing.T) {
	z, err := Open(gamePath(t), 70)
	if err != nil {
		t.Skipf("no zone 70: %v", err)
	}
	raw, err := z.Fixtures()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 1364 {
		t.Fatalf("%d fixtures, want 1364", len(raw))
	}
	ground := 0
	for _, f := range raw {
		if f.OnGround {
			ground++
		}
		if f.Yaw <= -math.Pi-1e-6 || f.Yaw > math.Pi+1e-6 {
			t.Fatalf("yaw %v outside a half turn", f.Yaw)
		}
	}
	placed, err := z.PlacedFixtures()
	if err != nil {
		t.Fatal(err)
	}
	for i, f := range placed {
		if f.OnGround && f.Z <= 0 {
			t.Fatalf("fixture %d stands on the ground but has height %v", i, f.Z)
		}
	}
	if ground == 0 {
		t.Error("no fixture is marked as standing on the ground")
	}
	t.Logf("%d fixtures, %d placed on the terrain", len(raw), ground)
}

func TestHeadingYaw(t *testing.T) {
	for _, c := range []struct{ deg, want float32 }{
		{0, 0}, {15, -0.261799}, {225, 2.356194}, {180, math.Pi}, {90, -math.Pi / 2},
	} {
		if got := headingYaw(c.deg); math.Abs(float64(got-c.want)) > 1e-5 {
			t.Errorf("headingYaw(%v) = %v, want %v", c.deg, got, c.want)
		}
	}
}
