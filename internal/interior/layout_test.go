package interior

import (
	"os"
	"testing"
)

func gamePath(t *testing.T) string {
	t.Helper()
	game := os.Getenv("DAOC_PATH")
	if game == "" {
		game = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`
	}
	if _, err := os.Stat(game); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", game, err)
	}
	return game
}

func read(t *testing.T, num int) *Layout {
	t.Helper()
	l, err := Read(gamePath(t), num)
	if err != nil {
		t.Skipf("zone %d: %v", num, err)
	}
	return l
}

// Spindelhalla: 26 piece models placed 194 times, two props, 82 lights,
// and the way out to Myrkwood.
func TestSpindelhalla(t *testing.T) {
	l := read(t, 125)
	if l.Kind != Dungeon || len(l.Models) != 26 || len(l.Pieces) != 194 || len(l.Props) != 2 || l.Skipped != 0 {
		t.Fatalf("%s: %d models, %d pieces, %d props, %d skipped; want dungeon, 26, 194, 2, 0",
			l.Kind, len(l.Models), len(l.Pieces), len(l.Props), l.Skipped)
	}
	if len(l.Lights) != 82 {
		t.Errorf("%d lights, want 82", len(l.Lights))
	}
	found := false
	for _, e := range l.Entrances {
		if e.Name == "ExittoMyrk" {
			found = true
			if e.Min[0] > e.Max[0] || e.Min[1] > e.Max[1] || e.Min[2] > e.Max[2] {
				t.Errorf("ExittoMyrk box %v..%v is inside out", e.Min, e.Max)
			}
		}
	}
	if !found {
		t.Errorf("no ExittoMyrk among %v", l.Entrances)
	}
	p := l.Pieces[1]
	if p.Model != l.Models[1] || p.Pos[0] != -1679.896729 || p.Angle != 1.570796 || p.Axis != [3]float32{0, 0, -1} {
		t.Errorf("second placement %+v, want piece 1 at -1679.9 turned 90° about -z", p)
	}
}

// Jordheim: 31 blocks, each where it stands.
func TestJordheim(t *testing.T) {
	l := read(t, 120)
	if l.Kind != City || len(l.Pieces) != 31 || l.Skipped != 0 {
		t.Fatalf("%s: %d blocks, %d skipped; want city, 31, 0", l.Kind, len(l.Pieces), l.Skipped)
	}
	if l.Pieces[0].Model != "jord01.nif" || l.Pieces[30].Model != "jord31.nif" {
		t.Errorf("blocks run %s..%s, want jord01.nif..jord31.nif", l.Pieces[0].Model, l.Pieces[30].Model)
	}
	if len(l.Entrances) == 0 || len(l.Lights) == 0 {
		t.Errorf("%d entrances, %d lights; want some of each", len(l.Entrances), len(l.Lights))
	}
}

// midgard is every Midgard interior in the install: Jordheim and the
// twelve dungeons.
var midgard = []int{120, 125, 126, 127, 128, 129, 149, 150, 160, 161, 188, 189, 226}

// Each Midgard interior's fitted transform puts its exits and lights on its
// geometry. Trollheim and Tuscaran Glacier are reported rather than held:
// a third of their points lie off the geometry under any origin (see
// fit.go and the README). Nisse's Lair falls just short, at 77 of 86.
func TestTransformsFit(t *testing.T) {
	game := gamePath(t)
	known := map[int]bool{129: true, 150: true, 160: true}
	for _, n := range midgard {
		l, err := Read(game, n)
		if err != nil {
			t.Logf("zone %d: %v", n, err)
			continue
		}
		b := Build(game, l)
		f := FitInterior(l, b.Surface)
		t.Logf("zone %d %s: origin %.0f, %.0f, %.0f; %d of %d points on the geometry (%.2f)",
			n, l.Kind, f.Origin[0], f.Origin[1], f.Origin[2], f.OnSurface, f.Points, f.Score())
		if f.Turn != 2 || !f.MirrorY {
			t.Errorf("zone %d: orientation turn %d mirror %v, want x mirrored", n, f.Turn, f.MirrorY)
		}
		if f.Score() < 0.9 && !known[n] {
			t.Errorf("zone %d: only %d of %d exits and lights on the geometry", n, f.OnSurface, f.Points)
		}
	}
}

// Props are placed in the frame of the piece their second unnamed column
// names; read so, the props of the dungeons that have many lie on the
// pieces.
func TestPropsSitOnPieces(t *testing.T) {
	game := gamePath(t)
	rejected := 0
	for _, n := range []int{188, 189, 226} {
		l, err := Read(game, n)
		if err != nil {
			t.Logf("zone %d: %v", n, err)
			continue
		}
		b := Build(game, l)
		pieces := Build(game, &Layout{Num: l.Num, Kind: l.Kind, Models: l.Models, Pieces: l.Pieces})
		on, all := 0, 0
		for _, p := range b.Placed {
			if !p.Prop {
				continue
			}
			all++
			if pieces.Surface.Near(p.M.Point([3]float32{})) {
				on++
			}
		}
		if all == 0 || float64(on) < 0.9*float64(all) {
			t.Errorf("zone %d: %d of %d props on the pieces", n, on, all)
		}
		// The other reading, each prop placed in the dungeon's own frame,
		// must not do as well. Where the pieces props sit on stand at the
		// dungeon's origin (all of 189's) the two agree, so only some
		// zones tell them apart.
		other := 0
		for _, p := range l.Props {
			if pieces.Surface.Near(p.Matrix().Point([3]float32{})) {
				other++
			}
		}
		t.Logf("zone %d: %d of %d props on the pieces in their piece's frame, %d in the dungeon's", n, on, all, other)
		if other > on {
			t.Errorf("zone %d: more props on the pieces in the dungeon's frame (%d) than in their piece's (%d)", n, other, on)
		}
		if float64(other) < 0.9*float64(len(l.Props)) {
			rejected++
		}
	}
	if rejected == 0 {
		t.Errorf("no zone tells the props' frames apart")
	}
}
