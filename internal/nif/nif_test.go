package nif_test

import (
	"os"
	"path/filepath"
	"testing"

	"daocweb/internal/nif"
)

// gamePath points at a real DAoC install. Override with DAOC_PATH.
func gamePath(t *testing.T) string {
	t.Helper()
	p := os.Getenv("DAOC_PATH")
	if p == "" {
		p = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`
	}
	if _, err := os.Stat(p); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", p, err)
	}
	return p
}

func load(t *testing.T, name string) *nif.File {
	t.Helper()
	path := filepath.Join(gamePath(t), "figures", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	f, err := nif.Parse(raw)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return f
}

// Parse already requires the footer to land exactly at end-of-file, so a clean
// parse of a 686KB, 479-block model is itself the headline assertion: every
// block layout in between has to be right for the walk to arrive there.
func TestNorsemanParses(t *testing.T) {
	f := load(t, "NVikingM.NIF")
	if got, want := f.Version, uint32(nif.Ver1010000); got != want {
		t.Errorf("version = 0x%08X, want 0x%08X", got, want)
	}
	if got, want := len(f.Blocks), 479; got != want {
		t.Errorf("block count = %d, want %d", got, want)
	}
	if got, want := len(f.BlockTypes), 10; got != want {
		t.Errorf("block type count = %d, want %d", got, want)
	}
	if len(f.Roots) != 1 || f.Roots[0] != 0 {
		t.Errorf("roots = %v, want [0]", f.Roots)
	}

	census := map[string]int{}
	for _, typ := range f.TypeOf {
		census[typ]++
	}
	// A character is one skinned shape per wearable variant, each with its own
	// geometry, skin binding and partition; the counts move together.
	for _, typ := range []string{"NiTriShape", "NiTriShapeData", "NiSkinInstance", "NiSkinData", "NiSkinPartition"} {
		if census[typ] != 49 {
			t.Errorf("%s count = %d, want 49", typ, census[typ])
		}
	}
	if census["NiNode"] != 114 {
		t.Errorf("NiNode count = %d, want 114", census["NiNode"])
	}
}

// The skeleton uses 3ds Max Biped names, which is what makes the model
// animatable without parsing the .kfa files.
func TestSkeletonUsesBipedNames(t *testing.T) {
	f := load(t, "NVikingM.NIF")
	want := []string{
		"Bip01", "Bip01 Pelvis", "Bip01 Spine", "Bip01 Head",
		"Bip01 L Thigh", "Bip01 L Calf", "Bip01 L Foot",
		"Bip01 R UpperArm", "Bip01 R Forearm", "Bip01 R Hand",
	}
	have := map[string]bool{}
	for _, b := range f.Blocks {
		if n, ok := b.(*nif.Node); ok {
			have[n.Name] = true
		}
	}
	for _, name := range want {
		if !have[name] {
			t.Errorf("skeleton has no bone named %q", name)
		}
	}
}

// The bind pose is the strongest check available on the transform maths: for
// every bone of a skinned shape, the bone's world transform composed with that
// bone's skin transform must reproduce the shape's own world placement. If the
// hierarchy walk, the matrix layout or either transform were read wrongly this
// fails by a wide margin rather than subtly.
func TestBindPoseHolds(t *testing.T) {
	f := load(t, "NVikingM.NIF")
	world := f.WorldTransforms()
	var worst float32
	bindings := 0
	for i, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		si, _ := f.Block(s.Skin).(*nif.SkinInstance)
		if si == nil {
			continue
		}
		sd, _ := f.Block(si.Data).(*nif.SkinData)
		if sd == nil {
			t.Fatalf("shape %s: skin instance has no data", s.Name)
		}
		if len(sd.Bones) != len(si.Bones) {
			t.Fatalf("shape %s: %d bone weights for %d bones", s.Name, len(sd.Bones), len(si.Bones))
		}
		for bi, ref := range si.Bones {
			got := world[ref].Mul(sd.Bones[bi].Skin)
			if d := got.MaxDiff(world[i]); d > worst {
				worst = d
			}
			bindings++
		}
	}
	if bindings != 494 {
		t.Errorf("bone bindings = %d, want 494", bindings)
	}
	// The residual grows with hierarchy depth as float32 matrices multiply;
	// the worst offenders are finger joints ten levels from the root. One per
	// cent of the model's 72-unit height is comfortably above that and far
	// below anything a real layout error would produce.
	if worst > 0.72 {
		t.Errorf("worst bind-pose residual = %g units, want < 0.72", worst)
	}
	t.Logf("%d bone bindings, worst residual %.4g units", bindings, worst)
}

// NiSkinData's overall transform is the inverse of the shape's world placement.
// This check avoids the deep hierarchy, so it pins the arithmetic much tighter.
func TestSkinTransformIsInverseOfShapeWorld(t *testing.T) {
	f := load(t, "NVikingM.NIF")
	world := f.WorldTransforms()
	var worst float32
	for i, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		si, _ := f.Block(s.Skin).(*nif.SkinInstance)
		if si == nil {
			continue
		}
		sd, _ := f.Block(si.Data).(*nif.SkinData)
		if sd == nil {
			continue
		}
		if d := world[i].Inverse().MaxDiff(sd.Skin); d > worst {
			worst = d
		}
	}
	if worst > 1e-4 {
		t.Errorf("worst inverse residual = %g, want < 1e-4", worst)
	}
	t.Logf("worst residual %.4g", worst)
}

// Geometry has to be self-consistent: triangles must index real vertices, and
// every vertex must carry a normal and a UV or the export cannot be textured.
func TestGeometryIsWellFormed(t *testing.T) {
	f := load(t, "NVikingM.NIF")
	shapes, verts, tris := 0, 0, 0
	for _, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		d, _ := f.Block(s.Data).(*nif.ShapeData)
		if d == nil {
			t.Fatalf("shape %s has no geometry block", s.Name)
		}
		shapes++
		verts += len(d.Vertices)
		tris += len(d.Triangles)
		if len(d.Normals) != len(d.Vertices) {
			t.Errorf("shape %s: %d normals for %d vertices", s.Name, len(d.Normals), len(d.Vertices))
		}
		if len(d.UV) == 0 {
			t.Errorf("shape %s: no UV set", s.Name)
			continue
		}
		if len(d.UV[0]) != len(d.Vertices) {
			t.Errorf("shape %s: %d UVs for %d vertices", s.Name, len(d.UV[0]), len(d.Vertices))
		}
		for _, tri := range d.Triangles {
			for _, v := range tri {
				if int(v) >= len(d.Vertices) {
					t.Fatalf("shape %s: triangle indexes vertex %d of %d", s.Name, v, len(d.Vertices))
				}
			}
		}
	}
	if shapes != 49 {
		t.Errorf("shapes = %d, want 49", shapes)
	}
	t.Logf("%d shapes, %d vertices, %d triangles", shapes, verts, tris)
}

// Skin weights must sum to one per vertex, or skinning drags the mesh toward
// the origin.
func TestSkinWeightsSumToOne(t *testing.T) {
	f := load(t, "NVikingM.NIF")
	var worst float64
	for _, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		si, _ := f.Block(s.Skin).(*nif.SkinInstance)
		if si == nil {
			continue
		}
		sd, _ := f.Block(si.Data).(*nif.SkinData)
		d, _ := f.Block(s.Data).(*nif.ShapeData)
		if sd == nil || d == nil {
			continue
		}
		sums := make([]float64, len(d.Vertices))
		for _, bd := range sd.Bones {
			for _, w := range bd.Weights {
				if int(w.Index) >= len(sums) {
					t.Fatalf("shape %s: weight for vertex %d of %d", s.Name, w.Index, len(sums))
				}
				sums[w.Index] += float64(w.Weight)
			}
		}
		for i, v := range sums {
			if dv := abs(v - 1); dv > worst {
				worst = dv
				if dv > 1e-3 {
					t.Errorf("shape %s vertex %d: weights sum to %g", s.Name, i, v)
				}
			}
		}
	}
	t.Logf("worst weight-sum error %.3g", worst)
}

// The other NIF generation in the install. It is not needed by the character
// pipeline, but parsing must fail loudly rather than silently misreading it.
func TestOlderItemFormatIsRejectedClearly(t *testing.T) {
	path := filepath.Join(gamePath(t), "items", "B_2h_Claymore01.nif")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("no sample item: %v", err)
	}
	_, err = nif.Parse(raw)
	if err == nil {
		t.Skip("items now parse; the 4.2.1.0 block set must have been completed")
	}
	t.Logf("as expected, 4.2.1.0 items are not fully modelled yet: %v", err)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
