// Package pcx decodes the 8-bit RLE PCX images DAoC uses for per-zone raster
// data (heightmaps, grass density, shade and shadow masks).
package pcx

import (
	"errors"
	"fmt"
)

// Image is a decoded 8-bit paletted PCX.
type Image struct {
	W, H   int
	Pix    []byte       // W*H palette indices, row-major, top-down
	Pal    [256][3]byte // VGA palette; zero if the file carried none
	HasPal bool
}

// At returns the palette index at (x, y), clamped to the image bounds.
func (im *Image) At(x, y int) byte {
	if x < 0 {
		x = 0
	} else if x >= im.W {
		x = im.W - 1
	}
	if y < 0 {
		y = 0
	} else if y >= im.H {
		y = im.H - 1
	}
	return im.Pix[y*im.W+x]
}

// Decode parses an 8-bit, single-plane, RLE-encoded PCX.
func Decode(b []byte) (*Image, error) {
	if len(b) < 128 {
		return nil, errors.New("pcx: shorter than header")
	}
	if b[0] != 0x0A {
		return nil, fmt.Errorf("pcx: bad manufacturer byte 0x%02x", b[0])
	}
	enc, bpp, planes := b[2], b[3], b[65]
	if enc != 1 {
		return nil, fmt.Errorf("pcx: unsupported encoding %d (want 1, RLE)", enc)
	}
	if bpp != 8 || planes != 1 {
		return nil, fmt.Errorf("pcx: unsupported %d bpp / %d planes (want 8/1)", bpp, planes)
	}

	u16 := func(o int) int { return int(b[o]) | int(b[o+1])<<8 }
	xmin, ymin, xmax, ymax := u16(4), u16(6), u16(8), u16(10)
	w, h := xmax-xmin+1, ymax-ymin+1
	bpl := u16(66) // bytes per scanline, may exceed w as padding
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("pcx: bad dimensions %dx%d", w, h)
	}
	if bpl < w {
		bpl = w
	}

	im := &Image{W: w, H: h, Pix: make([]byte, w*h)}
	p := 128
	for y := 0; y < h; y++ {
		for x := 0; x < bpl; {
			if p >= len(b) {
				return nil, fmt.Errorf("pcx: data ends mid-image at row %d of %d", y, h)
			}
			c := b[p]
			p++
			run, val := 1, c
			if c&0xC0 == 0xC0 { // top two bits set => run-length packet
				run = int(c & 0x3F)
				if p >= len(b) {
					return nil, errors.New("pcx: run-length packet truncated")
				}
				val = b[p]
				p++
			}
			for k := 0; k < run; k++ {
				if x < w { // drop scanline padding
					im.Pix[y*w+x] = val
				}
				x++
			}
		}
	}

	// A 256-entry VGA palette may follow, introduced by a 0x0C marker.
	if len(b)-769 >= p && b[len(b)-769] == 0x0C {
		pal := b[len(b)-768:]
		for i := 0; i < 256; i++ {
			im.Pal[i] = [3]byte{pal[i*3], pal[i*3+1], pal[i*3+2]}
		}
		im.HasPal = true
	}
	return im, nil
}
