package ui

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestParseBitmapFontSynthetic(t *testing.T) {
	// Two rows of height 3. Each marker row (y=3, y=7) starts with the one-pixel
	// edge mark; row 1 then has segments at 2-6, row 2 at 3-6 after a 1-pixel gap.
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	on := color.NRGBA{255, 255, 255, 255}
	for _, x := range []int{0, 2, 3, 4} {
		img.SetNRGBA(x, 3, on)
	}
	img.SetNRGBA(0, 7, on)
	// Widen row 3 so it counts as mostly filled.
	for _, x := range []int{5, 6} {
		img.SetNRGBA(x, 3, on)
	}
	for _, x := range []int{3, 4, 5, 6} {
		img.SetNRGBA(x, 7, on)
	}
	f, err := ParseBitmapFont(img)
	if err != nil {
		t.Fatal(err)
	}
	if f.Height != 3 {
		t.Errorf("height = %d, want 3", f.Height)
	}
	want := []Glyph{{0x21, 2, 0, 5}, {0x22, 3, 4, 4}}
	if len(f.Glyphs) != len(want) {
		t.Fatalf("glyphs = %+v, want %+v", f.Glyphs, want)
	}
	for i, g := range want {
		if f.Glyphs[i] != g {
			t.Errorf("glyph %d = %+v, want %+v", i, f.Glyphs[i], g)
		}
	}
}

func TestParseBitmapFontNoMarkers(t *testing.T) {
	if _, err := ParseBitmapFont(image.NewNRGBA(image.Rect(0, 0, 8, 8))); err == nil {
		t.Error("blank sheet parsed")
	}
}

func TestRealFontSheets(t *testing.T) {
	dir := filepath.Join(installUI, "ui", "fonts")
	for _, name := range []string{"Arial09.tga", "Arial11b.tga", "Arial14.tga", "button_10.tga", "title_24.tga"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skip("no install:", err)
		}
		img, err := DecodeTGA(b)
		if err != nil {
			t.Fatal(name, err)
		}
		f, err := ParseBitmapFont(img)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		// 0x21..0xFF is 223 characters; the sheets carry a few more after.
		if len(f.Glyphs) < 0xFF-0x21+1 {
			t.Errorf("%s: %d glyphs, want at least 223", name, len(f.Glyphs))
		}
		adv := map[rune]int{}
		for _, g := range f.Glyphs {
			adv[rune(g.Code)] = g.W
		}
		if adv['l'] >= adv['W'] || adv['i'] >= adv['m'] {
			t.Errorf("%s: proportional widths wrong: l=%d W=%d i=%d m=%d", name, adv['l'], adv['W'], adv['i'], adv['m'])
		}
		var a *Glyph
		for i := range f.Glyphs {
			if f.Glyphs[i].Code == 'A' {
				a = &f.Glyphs[i]
			}
		}
		if a == nil || a.W < 4 || a.W > 20 {
			t.Errorf("%s: glyph A = %+v", name, a)
		}
	}
}
