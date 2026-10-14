package world

import (
	"encoding/json"
	"flag"
	"math"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/move_vectors.json")

// plane is a synthetic slope both this test and web/net.test.mjs can build
// from its formula alone.
type plane struct{}

func (plane) GroundAt(x, y, _ float64) float64 { return 4000 + 0.1*x + 0.05*y }
func (plane) IsWet(x, y float64) bool          { return false }
func (plane) Clamp(x, y float64) (float64, float64) {
	return math.Max(0, math.Min(65280, x)), math.Max(0, math.Min(65280, y))
}
func (plane) Extent() float64 { return 65280 }

type vector struct {
	Name  string     `json:"name"`
	Start [4]float64 `json:"start"` // x, y, z, yaw
	Runs  [][5]int   `json:"runs"`
	End   [4]float64 `json:"end"`
}

func runVector(v *vector) [4]float64 {
	b := Body{X: v.Start[0], Y: v.Start[1], Z: v.Start[2], Yaw: v.Start[3]}
	for _, r := range v.Runs {
		n, k := KeysFromRun(r)
		for i := 0; i < n; i++ {
			Step(plane{}, nil, &b, k)
		}
	}
	return [4]float64{b.X, b.Y, b.Z, b.Yaw}
}

func vectors() []vector {
	return []vector{
		{Name: "run forward one second facing north", Start: [4]float64{30000, 30000, 0, 0}, Runs: [][5]int{{60, 1, 0, 0, 1}}},
		{Name: "walk forward one second facing east", Start: [4]float64{30000, 30000, 0, math.Pi / 2}, Runs: [][5]int{{60, 1, 0, 0, 0}}},
		{Name: "back up is a walk even when running", Start: [4]float64{30000, 30000, 0, 0}, Runs: [][5]int{{60, -1, 0, 0, 1}}},
		{Name: "strafe right is a walk", Start: [4]float64{30000, 30000, 0, 0}, Runs: [][5]int{{60, 0, 1, 0, 1}}},
		{Name: "diagonal is not faster", Start: [4]float64{30000, 30000, 0, 0.3}, Runs: [][5]int{{60, 1, 1, 0, 1}}},
		{Name: "turn while running", Start: [4]float64{30000, 30000, 0, 1}, Runs: [][5]int{{30, 1, 0, 1, 1}, {30, 1, 0, -1, 1}, {15, 1, 0, 1, 0}}},
		{Name: "stopped at the zone edge", Start: [4]float64{100, 100, 0, -math.Pi / 2}, Runs: [][5]int{{120, 1, 0, 0, 1}}},
		{Name: "clamped input", Start: [4]float64{30000, 30000, 0, 0}, Runs: [][5]int{{10, 7, -9, 5, 3}}},
	}
}

func TestMoveVectors(t *testing.T) {
	path := filepath.Join("testdata", "move_vectors.json")
	vs := vectors()
	for i := range vs {
		vs[i].End = runVector(&vs[i])
	}
	if *update {
		b, _ := json.MarshalIndent(vs, "", "  ")
		os.MkdirAll("testdata", 0o755)
		if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var want []vector
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != len(vs) {
		t.Fatalf("%d vectors on disk, %d in the test; run with -update", len(want), len(vs))
	}
	for i := range want {
		got := runVector(&want[i])
		for j := range got {
			if math.Abs(got[j]-want[i].End[j]) > 1e-6 {
				t.Errorf("%s: end[%d] = %v, file says %v", want[i].Name, j, got[j], want[i].End[j])
			}
		}
	}
}

func TestMoveRates(t *testing.T) {
	// One second of running covers about 191 units; walking 85.
	b := Body{X: 30000, Y: 30000}
	for i := 0; i < 60; i++ {
		Step(plane{}, nil, &b, Keys{Fwd: 1, Run: true})
	}
	if d := math.Hypot(b.X-30000, b.Y-30000); math.Abs(d-Run) > 1 {
		t.Errorf("a second of running moved %.2f, want %v", d, Run)
	}
	b = Body{X: 30000, Y: 30000}
	for i := 0; i < 60; i++ {
		Step(plane{}, nil, &b, Keys{Fwd: -1, Run: true})
	}
	if d := math.Hypot(b.X-30000, b.Y-30000); math.Abs(d-Walk) > 1 {
		t.Errorf("a second of backing up moved %.2f, want %v", d, Walk)
	}
	// Diagonals are not faster than a straight run (the viewer runs them).
	b = Body{X: 30000, Y: 30000}
	for i := 0; i < 60; i++ {
		Step(plane{}, nil, &b, Keys{Fwd: 1, Strafe: 1, Run: true})
	}
	if d := math.Hypot(b.X-30000, b.Y-30000); d > Run+1 {
		t.Errorf("forward+strafe moved %.2f, faster than a straight run", d)
	}
	// The ground is followed.
	if want := (plane{}).GroundAt(b.X, b.Y, 0); math.Abs(b.Z-want) > 1e-9 {
		t.Errorf("z = %v, want %v", b.Z, want)
	}
}

func TestKeysClamped(t *testing.T) {
	n, k := KeysFromRun([5]int{100000, 9, -9, 4, 1})
	if n != 120 || k.Fwd != 1 || k.Strafe != -1 || k.Turn != 1 || !k.Run {
		t.Errorf("n=%d keys=%+v", n, k)
	}
	if n, _ := KeysFromRun([5]int{-5, 0, 0, 0, 0}); n != 0 {
		t.Errorf("negative count = %d", n)
	}
}

func TestTerrainAgainstZone100(t *testing.T) {
	web := filepath.Join("..", "..", "..", "web", "data")
	tr, name, err := LoadTerrain(web, 100)
	if err != nil {
		t.Skip("zone 100 is not converted:", err)
	}
	if name != "Vale of Mularn" || tr.Grid != 256 {
		t.Errorf("zone = %q grid %d", name, tr.Grid)
	}
	// At a sample, the ground is that sample; the lake is wet.
	if got, want := tr.GroundAt(100*256, 120*256, 0), float64(tr.Heights[120*256+100]); got != want {
		t.Errorf("ground at a sample = %v, want %v", got, want)
	}
	if tr.Extent() != 255*256 {
		t.Errorf("extent = %v", tr.Extent())
	}
	wet := 0
	for y := 0.0; y < tr.Extent(); y += 512 {
		for x := 0.0; x < tr.Extent(); x += 512 {
			if tr.IsWet(x, y) {
				wet++
			}
		}
	}
	if wet == 0 {
		t.Error("Mularn's lake is nowhere wet")
	}
}
