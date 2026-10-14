package world

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"daocweb/internal/server/protocol"
)

func vwall(x, y0, y1, z0, z1 float32) []tri {
	// A vertical wall along x = const from y0 to y1, two triangles.
	return []tri{
		newTri([3]float32{x, y0, z0}, [3]float32{x, y1, z0}, [3]float32{x, y1, z1}),
		newTri([3]float32{x, y0, z0}, [3]float32{x, y1, z1}, [3]float32{x, y0, z1}),
	}
}

func synth(walls []tri, cyls ...cyl) *Collision {
	b := &builder{cyls: cyls}
	for _, t := range walls {
		b.add(t, false)
	}
	return b.build(0, 0, 10000, 10000)
}

func TestSlideStopsAtAWall(t *testing.T) {
	c := synth(vwall(1000, 0, 2000, 0, 300))
	x, y := c.Slide(960, 1000, 975, 1000, 0)
	if x > 1000-BodyRadius+0.01 || x < 1000-BodyRadius-0.5 || y != 1000 {
		t.Errorf("walking into a wall ended at %.2f,%.2f, want x=%.0f", x, y, 1000-BodyRadius)
	}
	// Moving diagonally slides along it: x is held, y advances.
	x, y = c.Slide(965, 1000, 970, 1004, 0)
	if x > 1000-BodyRadius+0.01 || math.Abs(y-1004) > 1.5 {
		t.Errorf("diagonal into the wall ended at %.2f,%.2f", x, y)
	}
	// A body far from it is untouched, and so is one with the wall behind it.
	if x, y := c.Slide(500, 500, 503, 500, 0); x != 503 || y != 500 {
		t.Errorf("open ground: %.2f,%.2f", x, y)
	}
	// A nil collision moves freely.
	var none *Collision
	if x, y := none.Slide(0, 0, 5, 5, 0); x != 5 || y != 5 {
		t.Error("nil collision blocked")
	}
}

func TestLowThingsAreSteppedOver(t *testing.T) {
	c := synth(vwall(1000, 0, 2000, 0, 40)) // a 40-unit kerb
	if x, _ := c.Slide(990, 1000, 1005, 1000, 0); x != 1005 {
		t.Errorf("a kerb blocked: x=%.1f", x)
	}
	// Standing on a platform, the wall below knee height is ignored too, and
	// a wall entirely above the head is no obstacle.
	c = synth(vwall(1000, 0, 2000, 300, 600))
	if x, _ := c.Slide(990, 1000, 1005, 1000, 0); x != 1005 {
		t.Errorf("a wall over our head blocked: x=%.1f", x)
	}
}

func TestCylindersBlockTrunks(t *testing.T) {
	c := synth(nil, cyl{x: 2000, y: 2000, r: 24, zmin: -50, zmax: 400})
	x, y := c.Slide(2040, 2000, 2050-14, 2000, 0)
	_ = y
	if d := math.Hypot(x-2000, y-2000); d < 24+BodyRadius-0.5 {
		t.Errorf("walked into a trunk: distance %.2f", d)
	}
	// Walking straight at it slides around rather than through, over many steps.
	px, py := 1800.0, 2005.0
	for i := 0; i < 300; i++ {
		px, py = c.Slide(px, py, px+3.2, py, 0)
		if d := math.Hypot(px-2000, py-2000); d < 24+BodyRadius-0.5 {
			t.Fatalf("step %d: inside the trunk at distance %.2f", i, d)
		}
	}
	if px < 2000 {
		t.Errorf("never got past the tree: x=%.1f", px)
	}
}

func TestInstancePlacement(t *testing.T) {
	in := instance{P: [3]float64{1000, 2000, 300}, Yaw: math.Pi / 2, S: 2}
	got := in.place([3]float32{10, 0, 5})
	// Turned a quarter about z: (10,0) -> (0,10); scaled x2; moved.
	if math.Abs(float64(got[0])-1000) > 1e-3 || math.Abs(float64(got[1])-2020) > 1e-3 || math.Abs(float64(got[2])-310) > 1e-3 {
		t.Errorf("placed at %v", got)
	}
}

func TestMularnHasSolidTreesAndBuildings(t *testing.T) {
	web := filepath.Join("..", "..", "..", "web", "data")
	c := LoadOutdoorCollision(web, 100, 65536)
	if c == nil {
		t.Skip("zone 100 has no collision files: run go run ./cmd/propconv -zone 100")
	}
	if len(c.cyls) < 700 {
		t.Errorf("only %d tree trunks; Mularn places 779 pines and evergreens", len(c.cyls))
	}
	if len(c.tris) < 100 {
		t.Errorf("only %d wall triangles", len(c.tris))
	}
	// Walk straight at some tree from 400 units away: never inside it.
	tr := c.cyls[0]
	px, py := float64(tr.x)-400, float64(tr.y)+3
	for i := 0; i < 400; i++ {
		px, py = c.Slide(px, py, px+3.2, py, float64(tr.zmin)+50)
		if d := math.Hypot(px-float64(tr.x), py-float64(tr.y)); d < float64(tr.r)+BodyRadius-1 {
			t.Fatalf("step %d: inside a trunk at %.1f (radius %.1f)", i, d, tr.r)
		}
	}
}

func TestSpindelhallaFloorsAndWalls(t *testing.T) {
	web := filepath.Join("..", "..", "..", "web", "data")
	in, name, err := LoadInterior(web, 125)
	if err != nil {
		t.Skip("Spindelhalla is not converted:", err)
	}
	if name != "Spindelhalla" || !in.HasFloor() {
		t.Fatalf("name %q floor %v", name, in.HasFloor())
	}
	x, y, z, ok := in.SpawnAt()
	if !ok {
		t.Fatal("no floor found at the entrance")
	}
	// Run in every direction for ten seconds. The floor must hold: never a
	// drop through it to another level, never a climb over a wall, and the
	// walker stays inside the dungeon's bounds.
	for k := 0; k < 16; k++ {
		yaw := float64(k) / 16 * 2 * math.Pi
		b := Body{X: x, Y: y, Z: z, Yaw: yaw}
		lastZ := b.Z
		for i := 0; i < 600; i++ {
			Step(in, in, &b, Keys{Fwd: 1, Run: true})
			if dz := lastZ - b.Z; dz > StepHeight*2 || dz < -StepHeight-1 {
				t.Fatalf("heading %d step %d: z went %.1f -> %.1f", k, i, lastZ, b.Z)
			}
			lastZ = b.Z
		}
		if b.X < in.min[0] || b.X > in.max[0] || b.Y < in.min[1] || b.Y > in.max[1] {
			t.Errorf("heading %d: left the bounds at %.0f,%.0f", k, b.X, b.Y)
		}
	}
}

func TestSlideNeverEndsInsideWallsInAdjacentCells(t *testing.T) {
	// A wall exactly on a bin boundary must still be felt from both sides.
	c := synth(vwall(256, 0, 2000, 0, 300))
	for _, x0 := range []float64{200, 240, 300} {
		x, _ := c.Slide(x0, 500, 256, 500, 0)
		if math.Abs(x-256) < BodyRadius-0.5 {
			t.Errorf("from %v: ended %.2f from the wall", x0, math.Abs(x-256))
		}
	}
}

// --- vectors shared with web/collision.test.mjs ---

type colScene struct {
	Name  string       `json:"name"`
	Tris  [][9]float32 `json:"tris"`
	Cyls  [][5]float32 `json:"cyls"`
	Start [3]float64   `json:"start"`
	Step  [2]float64   `json:"step"` // movement per step
	Steps int          `json:"steps"`
	End   [2]float64   `json:"end"`
}

func (s *colScene) collision() *Collision {
	b := &builder{}
	for _, t := range s.Tris {
		b.add(newTri([3]float32{t[0], t[1], t[2]}, [3]float32{t[3], t[4], t[5]}, [3]float32{t[6], t[7], t[8]}), false)
	}
	for _, c := range s.Cyls {
		b.cyls = append(b.cyls, cyl{x: c[0], y: c[1], r: c[2], zmin: c[3], zmax: c[4]})
	}
	return b.build(0, 0, 10000, 10000)
}

func (s *colScene) run() [2]float64 {
	c := s.collision()
	x, y := s.Start[0], s.Start[1]
	for i := 0; i < s.Steps; i++ {
		x, y = c.Slide(x, y, x+s.Step[0], y+s.Step[1], s.Start[2])
	}
	return [2]float64{x, y}
}

func wallArr(x0, y0, x1, y1, z0, z1 float32) [][9]float32 {
	return [][9]float32{
		{x0, y0, z0, x1, y1, z0, x1, y1, z1},
		{x0, y0, z0, x1, y1, z1, x0, y0, z1},
	}
}

func colScenes() []colScene {
	corner := append(wallArr(1000, 0, 1000, 2000, 0, 300), wallArr(0, 1000, 2000, 1000, 0, 300)...)
	return []colScene{
		{Name: "walk into a wall", Tris: wallArr(1000, 0, 1000, 2000, 0, 300), Start: [3]float64{500, 500, 0}, Step: [2]float64{3.2, 0}, Steps: 300},
		{Name: "diagonal along a wall", Tris: wallArr(1000, 0, 1000, 2000, 0, 300), Start: [3]float64{500, 500, 0}, Step: [2]float64{2.2, 2.2}, Steps: 300},
		{Name: "into a corner", Tris: corner, Start: [3]float64{500, 500, 0}, Step: [2]float64{2.2, 2.2}, Steps: 400},
		{Name: "a diagonal wall", Tris: wallArr(1000, 500, 1500, 1000, 0, 300), Start: [3]float64{900, 900, 0}, Step: [2]float64{3.2, 0}, Steps: 300},
		{Name: "around a tree", Cyls: [][5]float32{{2000, 2000, 24, -50, 400}}, Start: [3]float64{1700, 2004, 0}, Step: [2]float64{3.2, 0}, Steps: 200},
		{Name: "a wall on a bin edge", Tris: wallArr(1024, 0, 1024, 2000, 0, 300), Start: [3]float64{900, 700, 0}, Step: [2]float64{3.2, 0.4}, Steps: 100},
		{Name: "over a kerb", Tris: wallArr(1000, 0, 1000, 2000, 0, 40), Start: [3]float64{500, 500, 0}, Step: [2]float64{3.2, 0}, Steps: 300},
	}
}

func TestCollisionVectors(t *testing.T) {
	path := filepath.Join("testdata", "collision_vectors.json")
	sc := colScenes()
	for i := range sc {
		sc[i].End = sc[i].run()
	}
	if *update {
		b, _ := json.MarshalIndent(sc, "", "  ")
		if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var want []colScene
	if err := json.Unmarshal(b, &want); err != nil || len(want) != len(sc) {
		t.Fatalf("collision_vectors.json out of date (%v); run with -update", err)
	}
	for i := range want {
		got := want[i].run()
		if math.Abs(got[0]-want[i].End[0]) > 1e-4 || math.Abs(got[1]-want[i].End[1]) > 1e-4 {
			t.Errorf("%s: ends at %v, file says %v", want[i].Name, got, want[i].End)
		}
	}
	// And the scenes mean what their names say.
	if e := sc[0].End; math.Abs(e[0]-(1000-BodyRadius)) > 0.5 {
		t.Errorf("wall: ended at x=%.2f", e[0])
	}
	if e := sc[6].End; e[0] < 1400 {
		t.Errorf("a 40-unit kerb stopped us at x=%.1f", e[0])
	}
}

func TestWalkingOffAnEdgeEntersTheNeighbour(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	cat, err := LoadCatalog(filepath.Join(root, "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store")
	}
	tr, name, err := LoadTerrain(filepath.Join(root, "web", "data"), 105)
	if err != nil {
		t.Skip("zone 105 is not converted")
	}
	r := newRig(t)
	r.w.Cat = cat
	z := NewZone(r.w, 105, name, tr, nil, nil)
	r.z = z
	var neighbours []int
	for dir, at := range map[string][2]float64{"west": {5, 49152}, "north": {16384, 5}} {
		p := r.playerIn(z, "Walker"+dir, at[0], at[1])
		yaw := map[string]float64{"west": -math.Pi / 2, "east": math.Pi / 2, "north": 0, "south": math.Pi}[dir]
		p.Body.Yaw = yaw
		p.QueueInput(&protocol.Input{T: "input", Steps: [][5]int{{12, 1, 0, 0, 1}}})
		r.ticks(6)
		if p.closed.Load() {
			neighbours = append(neighbours, p.Char.Zone)
			if p.Char.Zone == 105 {
				t.Errorf("%s: crossed into the same zone", dir)
			}
		}
	}
	sort.Ints(neighbours)
	if len(neighbours) != 2 || neighbours[0] != 103 || neighbours[1] != 106 {
		t.Errorf("walking off Myrkwood Forest's west and north edges reached zones %v, want Gotar (103) and Skona Ravine (106)", neighbours)
	}
}
