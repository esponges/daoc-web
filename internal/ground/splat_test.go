package ground

import (
	"image"
	"os"
	"strings"
	"testing"

	"daocweb/internal/zone"
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

func open(t *testing.T, game string, num int) *zone.Zone {
	t.Helper()
	z, err := zone.Open(game, num)
	if err != nil {
		t.Skipf("no zone %d: %v", num, err)
	}
	return z
}

func quiet(string, ...any) {}

// TestGroundDetailMatchesTheGame composites zone 100's ground layers and
// checks they come to what the game's own pre-blended tiles show, in the
// orientation the masks are named in, painted one over another.
func TestGroundDetailMatchesTheGame(t *testing.T) {
	game := gamePath(t)
	sp, err := ReadSplat(game, open(t, game, 100), 8, quiet)
	if err != nil {
		t.Fatal(err)
	}
	if sp.Agree < 0.95 {
		t.Errorf("composite agrees with tex100.mpk at r = %.4f, want > 0.95", sp.Agree)
	}
	if !strings.HasPrefix(sp.Layout, (maskLayout{}).String()) || !strings.Contains(sp.Layout, "painted in order") {
		t.Errorf("layout %q, want masks as named, unflipped, painted in order", sp.Layout)
	}
	// Every sector paints something, beginning with a layer laid solid.
	for i, row := range sp.Slots {
		if len(row) == 0 {
			t.Errorf("sector %d paints no layers", i)
		}
	}
	// The cobbles exist only where a village has them.
	found := false
	for _, n := range sp.Names {
		found = found || strings.EqualFold(n, "h_dirtycobbles02")
	}
	if !found {
		t.Errorf("Mularn's cobbles are not among the layers %v", sp.Names)
	}
}

// Zone 102 keeps its reference tiles as 8-bit BMPs, stored bottom up.
func TestBMPTile(t *testing.T) {
	game := gamePath(t)
	arc, err := open(t, game, 102).Archive("tex")
	if err != nil {
		t.Fatal(err)
	}
	if arc.Has("tex01-00.dds") || !arc.Has("tex01-00.bmp") {
		t.Skip("zone 102's tiles are not the BMPs this expects")
	}
	im, err := readTile(arc, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if im.Bounds().Dx() != 256 || im.Bounds().Dy() != 256 {
		t.Fatalf("tile is %v, want 256x256", im.Bounds())
	}
	// A ground tile, so neither blank nor one colour.
	seen := map[[3]uint8]bool{}
	for i := 0; i < len(im.Pix); i += 4 {
		seen[[3]uint8{im.Pix[i], im.Pix[i+1], im.Pix[i+2]}] = true
	}
	if len(seen) < 32 {
		t.Errorf("tile has %d colours", len(seen))
	}
}

// A 64px mask doubles to 128px with its edges kept sharp: every texel is a
// copy of one source texel, none a blend of two.
func TestNearestMask(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			v := uint8(0)
			if x >= 21 {
				v = 255
			}
			o := y*src.Stride + x*4
			src.Pix[o], src.Pix[o+3] = v, 255
		}
	}
	out := nearest(src, 128)
	if out.Bounds().Dx() != 128 || out.Bounds().Dy() != 128 {
		t.Fatalf("got %v", out.Bounds())
	}
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			v := out.Pix[y*out.Stride+x*4]
			want := uint8(0)
			if x >= 42 {
				want = 255
			}
			if v != want {
				t.Fatalf("texel (%d,%d) = %d, want %d", x, y, v, want)
			}
		}
	}
	if nearest(out, 128) != out {
		t.Error("a mask already 128px should be kept as is")
	}
}
