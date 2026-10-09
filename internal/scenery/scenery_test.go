package scenery

import (
	"math"
	"os"
	"testing"

	"daocweb/internal/zone"
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

// Every model the zone places must parse and yield geometry. This is the
// regression test for the whole block set: 60 models spanning four format
// versions, triangle lists and strips, particle systems, controllers, lights
// and a camera. Because a block carries no length, any layout error anywhere
// in a file shows up here as a model that will not read at all.
func TestEveryPlacedModelConverts(t *testing.T) {
	game := gamePath(t)
	z, err := zone.Open(game, 100)
	if err != nil {
		t.Skipf("no zone 100: %v", err)
	}
	models, err := z.Models()
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := z.Fixtures()
	if err != nil {
		t.Fatal(err)
	}
	used := map[int]int{}
	for _, f := range fixtures {
		used[f.NifID]++
	}

	var verts []Vertex
	var indices []uint32
	converted, placed, total := 0, 0, 0
	for id, n := range used {
		total += n
		def, ok := models[id]
		if !ok {
			t.Errorf("fixture id %d has no entry in nifs.csv", id)
			continue
		}
		mo, err := ConvertModel(ModelDirs(game, false), def.Name, def.File, &verts, &indices)
		if err != nil {
			t.Errorf("%s (%d placements): %v", def.File, n, err)
			continue
		}
		if len(mo.Groups) == 0 {
			t.Errorf("%s produced no draw groups", def.File)
		}
		converted++
		placed += n
	}
	if placed != total {
		t.Errorf("%d of %d fixtures placed; want all of them", placed, total)
	}
	if len(verts) == 0 || len(indices) == 0 {
		t.Fatal("no geometry produced")
	}
	// Indices must address the shared vertex buffer.
	for _, ix := range indices {
		if int(ix) >= len(verts) {
			t.Fatalf("index %d addresses vertex %d of %d", ix, ix, len(verts))
		}
	}
	t.Logf("%d models converted, %d/%d fixtures placed, %d vertices, %d triangles",
		converted, placed, total, len(verts), len(indices)/3)
}

// TestTorchFlickers converts the Mularn tiki torch and checks its animation
// survives: the flame card rises at one texture height a second, additive
// and turned to the camera, and the corona breathes between 0.91 and 1.
func TestTorchFlickers(t *testing.T) {
	game := gamePath(t)
	var verts []Vertex
	var indices []uint32
	mo, err := ConvertModel(ModelDirs(game, false), "M06_tikitorch", "M06_tikitorch.nif", &verts, &indices)
	if err != nil {
		t.Fatal(err)
	}
	var flame, corona *Group
	for i := range mo.Groups {
		g := &mo.Groups[i]
		switch g.Texture {
		case "fireshell3":
			flame = g
		case "lampglow":
			corona = g
		}
	}
	if flame == nil || corona == nil {
		t.Fatalf("groups %+v: want a fireshell3 flame and a lampglow corona", mo.Groups)
	}
	if !flame.Additive || !flame.Billboard || flame.Scroll == nil || *flame.Scroll != [2]float32{0, 1} {
		t.Errorf("flame: additive %v, billboard %v, scroll %v; want both and a V scroll of 1", flame.Additive, flame.Billboard, flame.Scroll)
	}
	if len(corona.Pulse) != 5 || corona.Pulse[0][1] != 1 || corona.Pulse[3][1] > 0.92 {
		t.Errorf("corona pulse %v, want 5 keys from 1 down to about 0.91", corona.Pulse)
	}
}

// TestLogFireParticles converts the log fire and checks its two emitters
// come through as the model sets them: a fast, short-lived additive flame
// and a slow, long-lived blended smoke that grows as it rises.
func TestLogFireParticles(t *testing.T) {
	game := gamePath(t)
	var verts []Vertex
	var indices []uint32
	mo, err := ConvertModel(ModelDirs(game, false), "Fire", "logfire.nif", &verts, &indices)
	if err != nil {
		t.Fatal(err)
	}
	if len(mo.Particles) != 2 {
		t.Fatalf("%d particle systems, want flame and smoke", len(mo.Particles))
	}
	var flame, smoke *Particle
	for i := range mo.Particles {
		p := &mo.Particles[i]
		switch p.Texture {
		case "flame000":
			flame = p
		case "grysmoke":
			smoke = p
		}
	}
	if flame == nil || smoke == nil {
		t.Fatalf("textures %s and %s, want flame000 and grysmoke", mo.Particles[0].Texture, mo.Particles[1].Texture)
	}
	if !flame.Additive || flame.Rate != 90 || flame.Life[0] < 0.39 || flame.Life[0] > 0.41 || flame.Spin != 1 {
		t.Errorf("flame: additive %v, rate %v, life %v, spin %v", flame.Additive, flame.Rate, flame.Life, flame.Spin)
	}
	if flame.Gravity != nil || smoke.Gravity != nil {
		t.Errorf("the log fire has no gravity; got %v and %v", flame.Gravity, smoke.Gravity)
	}
	if smoke.Additive || smoke.Life[0] != 3 || smoke.Grow != 3 || len(smoke.Color) != 3 {
		t.Errorf("smoke: additive %v, life %v, grow %v, %d colour keys", smoke.Additive, smoke.Life, smoke.Grow, len(smoke.Color))
	}
	// Both spray up: the cone's vertical direction is 0, along the
	// emitter's +Z, and the smoke starts above the flames.
	if flame.Vertical[0] != 0 || smoke.Vertical[0] != 0 || smoke.Origin[2] <= flame.Origin[2] {
		t.Errorf("vertical %v / %v, origins %v / %v", flame.Vertical, smoke.Vertical, flame.Origin, smoke.Origin)
	}
	// The flame never has more alive than the model's own particle array.
	if alive := flame.Rate * (flame.Life[0] + flame.Life[1]); alive > 36.5 {
		t.Errorf("flame keeps %.1f alive; the model holds 36", alive)
	}
}

// TestLabyrinthFireRises converts a particle-only model whose flames spray
// downwards and are pulled back up: each of its two emitters has a single
// spherical NiGravity of force 171 at a point 350 units above the model's
// origin. The model has no geometry and must still convert.
func TestLabyrinthFireRises(t *testing.T) {
	game := gamePath(t)
	var verts []Vertex
	var indices []uint32
	mo, err := ConvertModel(ModelDirs(game, false), "fire", "fire_labyrinth_orange_01.nif", &verts, &indices)
	if err != nil {
		t.Fatal(err)
	}
	if len(mo.Groups) != 0 || len(mo.Particles) != 2 {
		t.Fatalf("%d groups, %d emitters; want none and two", len(mo.Groups), len(mo.Particles))
	}
	for _, p := range mo.Particles {
		if p.Gravity == nil {
			t.Fatal("no gravity")
		}
		g := *p.Gravity
		n := math.Sqrt(float64(g[0]*g[0] + g[1]*g[1] + g[2]*g[2]))
		if math.Abs(n-171) > 0.5 || float64(g[2]) < 0.9*n {
			t.Errorf("gravity %v (|g| %.1f); want 171 pulling up towards the point above", g, n)
		}
		// Sprayed straight down, the flames turn back within a quarter
		// second: speed over acceleration.
		if p.Vertical[0] < 3.1 || p.Speed[0]/float32(n) > 0.3 {
			t.Errorf("vertical %v, speed %v", p.Vertical, p.Speed)
		}
	}
	if mo.Min[2] >= mo.Max[2] {
		t.Errorf("bounds %v..%v are empty", mo.Min, mo.Max)
	}
}

// TestAvalonPierEmbeddedTexture converts a model whose textures are stored
// inside it, palettised, rather than beside it as files: they must decode to
// their stated size and be what the model's draws use.
func TestAvalonPierEmbeddedTexture(t *testing.T) {
	game := gamePath(t)
	var verts []Vertex
	var indices []uint32
	mo, err := ConvertModel(ModelDirs(game, false), "pier", "avpier.nif", &verts, &indices)
	if err != nil {
		t.Fatal(err)
	}
	img, ok := mo.Embedded["avpier_15"]
	if !ok {
		t.Fatalf("embedded textures %v, want avpier_15", len(mo.Embedded))
	}
	if b := img.Bounds(); b.Dx() != 128 || b.Dy() != 128 {
		t.Errorf("avpier_15 is %v, want 128x128", b)
	}
	// A real image, not one flat colour.
	colours := map[[4]uint32]bool{}
	for y := 0; y < 128; y += 8 {
		for x := 0; x < 128; x += 8 {
			r, g, b, a := img.At(x, y).RGBA()
			colours[[4]uint32{r, g, b, a}] = true
		}
	}
	if len(colours) < 8 {
		t.Errorf("only %d distinct colours sampled", len(colours))
	}
	used := false
	for _, g := range mo.Groups {
		if _, ok := mo.Embedded[g.Texture]; ok {
			used = true
		}
	}
	if !used {
		t.Errorf("no draw uses an embedded texture; groups %+v", mo.Groups)
	}
}
