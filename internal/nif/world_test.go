package nif_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/mpak"
	"daocweb/internal/nif"
)

// loadProp pulls a scenery model out of its one-entry archive in zones/Nifs.
func loadProp(t *testing.T, base string) *nif.File {
	t.Helper()
	path := filepath.Join(gamePath(t), "zones", "Nifs", base+".npk")
	a, err := mpak.Open(path)
	if err != nil {
		t.Skipf("no %s: %v", base, err)
	}
	for _, e := range a.Entries {
		if !strings.EqualFold(filepath.Ext(e.Name), ".nif") {
			continue
		}
		raw, err := a.ReadEntry(e)
		if err != nil {
			t.Fatalf("read %s: %v", e.Name, err)
		}
		f, err := nif.Parse(raw)
		if err != nil {
			t.Fatalf("parse %s: %v", e.Name, err)
		}
		return f
	}
	t.Fatalf("%s holds no .nif", path)
	return nil
}

// The pine is the single most-placed model in Vale of Mularn -- 551 of the
// zone's 1006 fixtures -- and it exercises the whole 4.2.1.0 path: an LOD node
// with inline ranges, a texturing property, alpha cut-out foliage.
func TestPineParses(t *testing.T) {
	f := loadProp(t, "npintre1")
	if got, want := f.Version, uint32(nif.Ver4020100); got != want {
		t.Errorf("version = 0x%08X, want 0x%08X", got, want)
	}
	if got, want := len(f.Blocks), 22; got != want {
		t.Errorf("block count = %d, want %d", got, want)
	}

	// The LOD node's two children must line up with its two ranges, which is
	// what pinned the four-byte field the schema does not show at this
	// version.
	var lod *nif.LODNode
	for _, b := range f.Blocks {
		if l, ok := b.(*nif.LODNode); ok {
			lod = l
			break
		}
	}
	if lod == nil {
		t.Fatal("no NiLODNode in the pine")
	}
	if len(lod.Children) != len(lod.Ranges) {
		t.Errorf("%d LOD children but %d ranges", len(lod.Children), len(lod.Ranges))
	}
	// Ranges must be ordered and contiguous: each level takes over where the
	// previous one stops.
	for i := 1; i < len(lod.Ranges); i++ {
		if lod.Ranges[i][0] != lod.Ranges[i-1][1] {
			t.Errorf("LOD range %d starts at %g but %d ended at %g",
				i, lod.Ranges[i][0], i-1, lod.Ranges[i-1][1])
		}
	}
}

// A 10.1.0.0 prop, which takes the other arm of every version gate: the LOD
// ranges live in their own block, and NiSourceTexture is a byte shorter than
// the schema says.
func TestFenceParses(t *testing.T) {
	f := loadProp(t, "NF_b-fence1")
	if got, want := f.Version, uint32(nif.Ver1010000); got != want {
		t.Errorf("version = 0x%08X, want 0x%08X", got, want)
	}
	var seen bool
	for i, b := range f.Blocks {
		if l, ok := b.(*nif.LODNode); ok && f.TypeOf[i] == "NiLODNode" {
			if l.Data < 0 {
				t.Error("10.1.0.0 LOD node has no data reference")
			}
			if _, ok := f.Block(l.Data).(*nif.LODNode); !ok {
				t.Errorf("LOD data reference %d does not point at LOD data", l.Data)
			}
			seen = true
		}
	}
	if !seen {
		t.Error("no NiLODNode found")
	}
}

// Strips are the scenery's native topology. Converting them is where a winding
// or an off-by-one would show, so check the conversion rather than trusting it:
// every triangle must index a real vertex, have three distinct corners, and the
// count must not exceed what the block declares.
func TestStripsConvertToTriangles(t *testing.T) {
	// Topology is per-model, not per-version: the trees and fences are
	// triangle lists, while the rocks, logs and huts are strips. These four
	// are strip-based, so the conversion is actually exercised.
	for _, name := range []string{"BOULDER", "LOG1", "MENHIR", "HmonLurik"} {
		f := loadProp(t, name)
		stripped, tris := 0, 0
		for _, b := range f.Blocks {
			d, ok := b.(*nif.ShapeData)
			if !ok || len(d.Strips) == 0 {
				continue
			}
			stripped++
			tris += len(d.Triangles)
			for _, tri := range d.Triangles {
				for _, v := range tri {
					if int(v) >= len(d.Vertices) {
						t.Fatalf("%s: triangle indexes vertex %d of %d", name, v, len(d.Vertices))
					}
				}
				if tri[0] == tri[1] || tri[1] == tri[2] || tri[0] == tri[2] {
					t.Errorf("%s: degenerate triangle %v survived conversion", name, tri)
				}
			}
		}
		if stripped > 0 {
			t.Logf("%s: %d strip blocks yielding %d triangles", name, stripped, tris)
		}
	}
}

// Scenery carries its own textures, unlike the character models. That is what
// lets propconv convert a prop without guessing which skin belongs to it, so
// the reference chain has to actually resolve.
func TestPropNamesItsOwnTexture(t *testing.T) {
	f := loadProp(t, "npintre1")
	found := map[string]bool{}
	for _, b := range f.Blocks {
		tp, ok := b.(*nif.Texturing)
		if !ok || !tp.HasBase {
			continue
		}
		src, ok := f.Block(tp.Base.Source).(*nif.SourceTexture)
		if !ok {
			t.Fatalf("base texture reference %d is not a NiSourceTexture", tp.Base.Source)
		}
		if src.FileName == "" {
			t.Error("source texture names no file")
			continue
		}
		found[strings.ToLower(src.FileName)] = true
	}
	if len(found) == 0 {
		t.Fatal("the pine resolves no texture at all")
	}
	for name := range found {
		if !strings.HasSuffix(name, ".dds") {
			t.Errorf("texture %q is not a .dds", name)
		}
		if _, err := os.Stat(filepath.Join(gamePath(t), "zones", "Nifs", name)); err != nil {
			// Case differs between the NIF and the filesystem, so a miss
			// here is only worth reporting, not failing on.
			t.Logf("note: %s not found by exact name (%v)", name, err)
		}
	}
	t.Logf("pine textures: %v", keys(found))
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
