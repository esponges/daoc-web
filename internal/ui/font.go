package ui

import (
	"fmt"
	"image"
)

// Glyph is where one character sits in a font sheet and how far it advances.
type Glyph struct {
	Code int `json:"c"` // Windows-1252 code
	X    int `json:"x"`
	Y    int `json:"y"`
	W    int `json:"w"` // advance; the glyph is drawn from X for W pixels
}

// BitmapFont is a glyph sheet with its table.
type BitmapFont struct {
	Height int     `json:"h"` // height of a glyph row, marker row excluded
	Glyphs []Glyph `json:"g"`
}

// ParseBitmapFont reads the layout of the install's TGA font sheets. Glyphs
// run left to right, row after row, from character 0x21 (the space is not in
// the sheet). Under each row of
// glyphs is a marker row: a one-pixel mark at the left edge, then a line of
// opaque segments with a gap between each, one per glyph, spanning the glyph
// and its spacing. The sheets carry
// no other table, so the markers are the metrics.
func ParseBitmapFont(img *image.NRGBA) (*BitmapFont, error) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	opaque := func(x, y int) bool { return img.NRGBAAt(x, y).A > 127 }

	// The first marker row is mostly filled; a row of letters never is. The
	// rows after it follow at the same pitch, and the last may hold only a few
	// glyphs, so they are found by position and not by how full they are.
	first := -1
	for y := 0; y < h && first < 0; y++ {
		n := 0
		for x := 0; x < w; x++ {
			if opaque(x, y) {
				n++
			}
		}
		if n >= w/2 {
			first = y
		}
	}
	if first < 0 {
		return nil, fmt.Errorf("font: no marker rows")
	}
	f := &BitmapFont{Height: first}
	code := 0x21
	for top := 0; top+f.Height < h; top += f.Height + 1 {
		my := top + f.Height
		any := false
		for x := 0; x < w; x++ {
			if opaque(x, my) {
				any = true
				break
			}
		}
		if !any {
			break
		}
		for x := 0; x < w; {
			if !opaque(x, my) {
				x++
				continue
			}
			start := x
			for x < w && opaque(x, my) {
				x++
			}
			if start == 0 && x == 1 {
				continue // the mark that starts every marker row
			}
			f.Glyphs = append(f.Glyphs, Glyph{Code: code, X: start, Y: top, W: x - start})
			code++
		}
	}
	return f, nil
}
