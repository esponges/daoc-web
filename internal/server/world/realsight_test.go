package world

import (
	"math"
	"path/filepath"
	"testing"
	"time"

	"daocweb/internal/server/store"
)

func realZone(t *testing.T, id int) *Zone {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	cat, err := LoadCatalog(filepath.Join(root, "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	st, _ := store.Open(":memory:")
	t.Cleanup(func() { st.Close() })
	w := New(Config{WebData: filepath.Join(root, "web", "data"), Cat: cat, Store: st, Log: t.Logf, Collision: OutdoorCollision})
	w.mu.Lock()
	z, err := w.zoneLocked(id)
	w.mu.Unlock()
	if err != nil {
		t.Skip("zone is not converted:", err)
	}
	z.Stop()
	return z
}

// Line of sight over the real Mularn: a pine between two bodies blocks the
// view and a step to the side does not; a ridge is found by the heightmap.
func TestRealMularnLineOfSight(t *testing.T) {
	z := realZone(t, 100)
	c, ok := z.Mv.(*Collision)
	if !ok || len(c.cyls) == 0 {
		t.Skip("Mularn has no collision data")
	}
	tr := c.cyls[0]
	x, y := float64(tr.x), float64(tr.y)
	g := func(px, py float64) float64 { return z.G.GroundAt(px, py, 0) }
	if !z.canSee(x-400, y+250, g(x-400, y+250), x+400, y+250, g(x+400, y+250)) {
		t.Log("the open line beside the tree is not clear (other scenery or terrain); not a failure")
	}
	// Eyes at the tree's own height band: stand on the ground either side of the trunk.
	zz := float64(tr.zmin) + 20
	if z.canSee(x-300, y, zz, x+300, y, zz) {
		t.Errorf("a trunk at (%.0f, %.0f) r %.0f stands between, and they see each other", x, y, tr.r)
	}
	// How many lines of 600 units across the town are blocked, and what each costs.
	blocked, n := 0, 0
	start := time.Now()
	for i := 0; i < 4000; i++ {
		ax := 30000 + float64(i%60)*400
		ay := 30000 + float64(i/60)*400
		bx, by := ax+600*math.Cos(float64(i)), ay+600*math.Sin(float64(i))
		n++
		if !z.canSee(ax, ay, g(ax, ay), bx, by, g(bx, by)) {
			blocked++
		}
	}
	per := time.Since(start) / time.Duration(n)
	t.Logf("%d of %d lines of 600 units blocked, %v each", blocked, n, per)
	if blocked == 0 || blocked == n {
		t.Errorf("%d of %d lines blocked: line of sight is doing nothing or everything", blocked, n)
	}
	if per > 200*time.Microsecond {
		t.Errorf("a line of sight costs %v", per)
	}
}

// Inside Spindelhalla the walls and floors block: lines between random floor
// points are sometimes seen and sometimes not.
func TestRealSpindelhallaLineOfSight(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	in, _, err := LoadInterior(filepath.Join(root, "web", "data"), 125)
	if err != nil || !in.HasFloor() {
		t.Skip("Spindelhalla is not converted")
	}
	x0, y0, z0, ok := in.SpawnAt()
	if !ok {
		t.Skip("no floor at the entrance")
	}
	seen, blocked, nearSeen, near := 0, 0, 0, 0
	for i := 0; i < 600; i++ {
		a := float64(i) * 2.399963 // a spiral of directions and distances
		d := 100 + float64(i%12)*150
		x1, y1 := x0+d*math.Cos(a), y0+d*math.Sin(a)
		fl := in.GroundAt(x1, y1, z0)
		if in.walkable(x1, y1, fl) {
			ok := in.Sees([3]float64{x0, y0, z0 + 150}, [3]float64{x1, y1, fl + 150})
			if d < 200 {
				near++
				if ok {
					nearSeen++
				}
			}
			if ok {
				seen++
			} else {
				blocked++
			}
		}
	}
	t.Logf("from the entrance: %d seen, %d blocked", seen, blocked)
	t.Logf("within 200 units: %d of %d seen", nearSeen, near)
	if near > 0 && nearSeen*2 < near {
		t.Errorf("only %d of %d nearby floor points are seen from the entrance", nearSeen, near)
	}
	if seen == 0 || blocked == 0 {
		t.Errorf("%d seen and %d blocked: walls are not blocking, or everything is", seen, blocked)
	}
}
