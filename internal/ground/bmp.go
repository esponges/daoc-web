package ground

import (
	"encoding/binary"
	"fmt"
	"image"

	"daocweb/internal/dds"
	"daocweb/internal/mpak"
)

// readTile reads the game's pre-blended picture of sector (sx, sy): DDS in
// most zones, BMP in the older ones (in Midgard, 102 to 107 and 111 to 113).
func readTile(arc *mpak.Archive, sx, sy int) (*image.NRGBA, error) {
	name := fmt.Sprintf("tex%02d-%02d", sx, sy)
	if arc.Has(name + ".dds") {
		raw, err := arc.Read(name + ".dds")
		if err != nil {
			return nil, err
		}
		return dds.Decode(raw)
	}
	raw, err := arc.Read(name + ".bmp")
	if err != nil {
		return nil, fmt.Errorf("neither %s.dds nor %s.bmp in %s", name, name, arc.Name)
	}
	im, err := decodeBMP(raw)
	if err != nil {
		return nil, fmt.Errorf("%s.bmp: %w", name, err)
	}
	return im, nil
}

// decodeBMP decodes the one kind of BMP the game's tile archives hold, all
// 1,216 of them: a 40-byte info header, 8 bits a pixel through a palette,
// uncompressed, rows stored bottom up. Anything else is an error rather
// than a guess.
func decodeBMP(b []byte) (*image.NRGBA, error) {
	le := binary.LittleEndian
	if len(b) < 54 || b[0] != 'B' || b[1] != 'M' {
		return nil, fmt.Errorf("not a BMP")
	}
	off := int(le.Uint32(b[10:]))
	hdr := int(le.Uint32(b[14:]))
	w := int(int32(le.Uint32(b[18:])))
	h := int(int32(le.Uint32(b[22:])))
	bpp := le.Uint16(b[28:])
	comp := le.Uint32(b[30:])
	if hdr != 40 || bpp != 8 || comp != 0 || w <= 0 || h == 0 {
		return nil, fmt.Errorf("unsupported BMP: header %d, %dx%d, %d bits, compression %d", hdr, w, h, bpp, comp)
	}
	topDown := h < 0
	if topDown {
		h = -h
	}
	colours := int(le.Uint32(b[46:]))
	if colours == 0 {
		colours = 256
	}
	pal := 14 + hdr
	stride := (w + 3) &^ 3
	if pal+colours*4 > off || off+stride*h > len(b) {
		return nil, fmt.Errorf("BMP truncated: %d bytes for %dx%d", len(b), w, h)
	}
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		src := y
		if !topDown {
			src = h - 1 - y
		}
		row := b[off+src*stride:]
		for x := 0; x < w; x++ {
			i := int(row[x])
			if i >= colours {
				return nil, fmt.Errorf("pixel index %d past a %d-colour palette", i, colours)
			}
			p := b[pal+i*4:] // blue, green, red, unused
			o := y*im.Stride + x*4
			im.Pix[o], im.Pix[o+1], im.Pix[o+2], im.Pix[o+3] = p[2], p[1], p[0], 255
		}
	}
	return im, nil
}
