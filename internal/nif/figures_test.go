package nif_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/nif"
)

// Every creature and NPC body lives as a loose file in figures/. This holds
// the parser to the share of them it reads end to end, so a change that
// quietly loses a block type shows up as a number going down.
//
// The ones still out of reach are specialised: textures embedded as
// NiPixelData, NiTextureEffect projections on the spell-like CSR* models,
// colour and texture-transform controllers, and the two "coco" models.
func TestFiguresParse(t *testing.T) {
	dir := filepath.Join(gamePath(t), "figures")
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("no figures: %v", err)
	}
	ok, total := 0, 0
	for _, e := range ents {
		if !strings.EqualFold(filepath.Ext(e.Name()), ".nif") {
			continue
		}
		total++
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := nif.Parse(raw); err != nil {
			t.Logf("%s: %v", e.Name(), err)
			continue
		}
		ok++
	}
	t.Logf("%d of %d figures parse", ok, total)
	if ok < 617 {
		t.Errorf("%d of %d figures parse, want at least 617", ok, total)
	}
}

// The composite NPC bodies set the tangent-space bit in their UV-set field;
// they only parse if the tangent and bitangent arrays are read.
func TestTangentSpaceGeometry(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(gamePath(t), "figures", "albbody01heada.nif"))
	if err != nil {
		t.Skip(err)
	}
	f, err := nif.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	shapes := 0
	for i := range f.Blocks {
		d, isData := f.Block(int32(i)).(*nif.ShapeData)
		if !isData {
			continue
		}
		shapes++
		for _, n := range d.Normals {
			l := n[0]*n[0] + n[1]*n[1] + n[2]*n[2]
			if l < 0.98 || l > 1.02 {
				t.Fatalf("block %d: normal %v has length² %.3f; read is skewed", i, n, l)
			}
		}
	}
	if shapes == 0 {
		t.Fatal("no shape data")
	}
}
