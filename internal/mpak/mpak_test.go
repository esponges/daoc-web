package mpak_test

import (
	"os"
	"path/filepath"
	"testing"

	"daocweb/internal/mpak"
	"daocweb/internal/pcx"
	"daocweb/internal/sector"
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

func zone100(t *testing.T, kind string) *mpak.Archive {
	t.Helper()
	p := filepath.Join(gamePath(t), "zones", "zone100", kind+"100.mpk")
	a, err := mpak.Open(p)
	if err != nil {
		t.Fatalf("open %s: %v", p, err)
	}
	return a
}

// These values were read out of the shipped archives; they pin the container
// layout so a regression in the header de-obfuscation or directory parsing
// fails loudly rather than producing plausible garbage.
func TestTerrainArchiveLayout(t *testing.T) {
	a := zone100(t, "ter")
	if got, want := a.Name, "ter100.mpk"; got != want {
		t.Errorf("archive name = %q, want %q", got, want)
	}
	if got, want := len(a.Entries), 237; got != want {
		t.Errorf("entry count = %d, want %d", got, want)
	}
	if !a.Has("textures.csv") {
		t.Error("textures.csv missing from ter100.mpk")
	}

	// Entry offsets must be contiguous in both compressed and uncompressed
	// space; that is what proves the two offset fields were identified right.
	var uOff, cOff uint32
	for i, e := range a.Entries {
		if e.UncOff != uOff {
			t.Fatalf("entry %d (%s): UncOff = %d, want %d", i, e.Name, e.UncOff, uOff)
		}
		if e.CompOff != cOff {
			t.Fatalf("entry %d (%s): CompOff = %d, want %d", i, e.Name, e.CompOff, cOff)
		}
		uOff += e.UncSize
		cOff += e.CompSize
	}
}

// Every entry must inflate to exactly the size the directory advertises.
func TestEveryEntryInflates(t *testing.T) {
	for _, kind := range []string{"ter", "dat", "lod", "csv"} {
		a := zone100(t, kind)
		for _, e := range a.Entries {
			if _, err := a.ReadEntry(e); err != nil {
				t.Errorf("%s100.mpk: %v", kind, err)
			}
		}
	}
}

func TestHeightmap(t *testing.T) {
	a := zone100(t, "dat")

	secRaw, err := a.Read("SECTOR.DAT")
	if err != nil {
		t.Fatal(err)
	}
	sec, err := sector.Parse(secRaw)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := sec.ScaleFactor(), 8; got != want {
		t.Errorf("scalefactor = %d, want %d", got, want)
	}
	if got, want := sec.OffsetFactor(), 48; got != want {
		t.Errorf("offsetfactor = %d, want %d", got, want)
	}
	if x, y := sec.SectorSize(); x != 8 || y != 8 {
		t.Errorf("sector grid = %dx%d, want 8x8", x, y)
	}

	waters := sec.Waters()
	if len(waters) != 1 {
		t.Fatalf("got %d water bodies, want 1", len(waters))
	}
	lake := waters[0]
	if lake.Name != "Lake" || lake.Height != 4372 {
		t.Errorf("water = %q at %d, want \"Lake\" at 4372", lake.Name, lake.Height)
	}
	if len(lake.Left) != 9 || len(lake.Right) != 9 {
		t.Errorf("shoreline = %d left / %d right, want 9/9", len(lake.Left), len(lake.Right))
	}
	if len(lake.Left) > 0 && lake.Left[0] != [2]int{81, 108} {
		t.Errorf("left00 = %v, want [81 108]", lake.Left[0])
	}

	// Decode both rasters and check the derived height range.
	load := func(name string) *pcx.Image {
		raw, err := a.Read(name)
		if err != nil {
			t.Fatal(err)
		}
		im, err := pcx.Decode(raw)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return im
	}
	ter, off := load("terrain.pcx"), load("offset.pcx")
	if ter.W != 256 || ter.H != 256 {
		t.Fatalf("terrain.pcx = %dx%d, want 256x256", ter.W, ter.H)
	}

	minH, maxH := 1<<30, 0
	belowLake := 0
	for y := 0; y < ter.H; y++ {
		for x := 0; x < ter.W; x++ {
			h := int(ter.At(x, y))*sec.ScaleFactor() + int(off.At(x, y))*sec.OffsetFactor()
			if h < minH {
				minH = h
			}
			if h > maxH {
				maxH = h
			}
			if h < lake.Height {
				belowLake++
			}
		}
	}
	if minH != 2880 || maxH != 11848 {
		t.Errorf("height range = %d..%d, want 2880..11848", minH, maxH)
	}
	// A single lake in a mountain valley: a few percent of the zone, not most
	// of it. Catches an off-by-a-factor error in the height formula.
	if frac := float64(belowLake) / float64(ter.W*ter.H); frac < 0.01 || frac > 0.15 {
		t.Errorf("%.1f%% of the zone is below the waterline, want 1-15%%", frac*100)
	}
}
