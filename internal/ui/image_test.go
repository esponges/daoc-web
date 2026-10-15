package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeTGAUncompressedBottomUp(t *testing.T) {
	// 2x2, 32 bit, bottom-up: the first stored row is the bottom row.
	b := []byte{0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 2, 0, 32, 8}
	b = append(b,
		1, 2, 3, 4, 5, 6, 7, 8, // bottom row: (B,G,R,A) x2
		9, 10, 11, 12, 13, 14, 15, 16) // top row
	img, err := DecodeTGA(b)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.NRGBAAt(0, 1), (color.NRGBA{3, 2, 1, 4}); got != want {
		t.Errorf("bottom-left = %v, want %v", got, want)
	}
	if got, want := img.NRGBAAt(1, 0), (color.NRGBA{15, 14, 13, 16}); got != want {
		t.Errorf("top-right = %v, want %v", got, want)
	}
}

func TestDecodeTGARunLength(t *testing.T) {
	// 4x1, 24 bit RLE: a run of three blue pixels then one raw red pixel.
	b := []byte{0, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 1, 0, 24, 0x20}
	b = append(b, 0x82, 255, 0, 0, 0x00, 0, 0, 255)
	img, err := DecodeTGA(b)
	if err != nil {
		t.Fatal(err)
	}
	if got := img.NRGBAAt(2, 0); got != (color.NRGBA{0, 0, 255, 255}) {
		t.Errorf("pixel 2 = %v, want blue", got)
	}
	if got := img.NRGBAAt(3, 0); got != (color.NRGBA{255, 0, 0, 255}) {
		t.Errorf("pixel 3 = %v, want red", got)
	}
}

func TestDecodeTGATruncated(t *testing.T) {
	b := []byte{0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 2, 0, 32, 8, 1, 2, 3}
	if _, err := DecodeTGA(b); err == nil {
		t.Error("truncated file decoded")
	}
}

func TestDecodeBMP24(t *testing.T) {
	// 2x1, 24 bit, row padded to 8 bytes.
	b := make([]byte, 54, 62)
	b[0], b[1] = 'B', 'M'
	b[10] = 54
	b[14] = 40
	b[18], b[22], b[26], b[28] = 2, 1, 1, 24
	b = append(b, 10, 20, 30, 40, 50, 60, 0, 0)
	img, err := DecodeBMP(b)
	if err != nil {
		t.Fatal(err)
	}
	if got := img.NRGBAAt(1, 0); got != (color.NRGBA{60, 50, 40, 255}) {
		t.Errorf("pixel = %v", got)
	}
}

const installUI = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

func TestRealAtlas(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(installUI, "ui", "atlantis", "atlantis_01.tga"))
	if err != nil {
		t.Skip("no install:", err)
	}
	img, err := DecodeTGA(b)
	if err != nil {
		t.Fatal(err)
	}
	if r := img.Bounds(); r.Dx() != 256 || r.Dy() != 256 {
		t.Fatalf("atlas is %dx%d, want 256x256", r.Dx(), r.Dy())
	}
	opaque := 0
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 {
			opaque++
		}
	}
	if opaque == 0 {
		t.Error("atlas decoded to fully transparent")
	}
}

func TestRealPaperDollBMP(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(installUI, "data", "skin1", "pperdoll.bmp"))
	if err != nil {
		t.Skip("no install:", err)
	}
	img, err := DecodeBMP(b)
	if err != nil {
		t.Fatal(err)
	}
	if r := img.Bounds(); r.Dx() != 192 || r.Dy() != 448 {
		t.Fatalf("sheet is %dx%d, want 192x448", r.Dx(), r.Dy())
	}
}

func TestRealSkinLoads(t *testing.T) {
	s, missing, err := LoadSkin(filepath.Join(installUI, "ui"), "atlantis")
	if err != nil {
		t.Skip("no install:", err)
	}
	if len(missing) > 0 {
		t.Errorf("missing includes: %v", missing)
	}
	chat := false
	for _, e := range s.Elements {
		if e.Tag == "WindowTemplate" && e.Str("WindowId") == "Chat" {
			chat = e.Int("Width") == 350 && e.Int("Height") == 200 && e.Int("MinWidth") == 312
		}
	}
	if !chat {
		t.Error("chat window not 350x200 (min 312) in atlantis")
	}
}
