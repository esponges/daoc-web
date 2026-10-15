package ui

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"strings"

	"daocweb/internal/dds"
)

// DecodeImage reads a DDS, TGA or BMP texture by its file extension.
func DecodeImage(name string, b []byte) (*image.NRGBA, error) {
	switch strings.ToLower(name[strings.LastIndex(name, ".")+1:]) {
	case "dds":
		return dds.Decode(b)
	case "tga":
		return DecodeTGA(b)
	case "bmp":
		return DecodeBMP(b)
	}
	return nil, fmt.Errorf("ui: no reader for %q", name)
}

// DecodeTGA reads uncompressed and run-length true-colour TGA files at 24 or
// 32 bits, and 8-bit greyscale or palette files, which is all the install has.
func DecodeTGA(b []byte) (*image.NRGBA, error) {
	if len(b) < 18 {
		return nil, fmt.Errorf("tga: shorter than header (%d bytes)", len(b))
	}
	idLen, cmType, imgType := int(b[0]), b[1], b[2]
	cmLen := int(binary.LittleEndian.Uint16(b[5:]))
	cmBits := int(b[7])
	w := int(binary.LittleEndian.Uint16(b[12:]))
	h := int(binary.LittleEndian.Uint16(b[14:]))
	bpp := int(b[16])
	desc := b[17]
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("tga: bad dimensions %dx%d", w, h)
	}
	rle := imgType >= 9
	base := imgType &^ 8
	if base != 2 && base != 1 && base != 3 {
		return nil, fmt.Errorf("tga: unsupported image type %d", imgType)
	}
	if base == 2 && bpp != 24 && bpp != 32 {
		return nil, fmt.Errorf("tga: unsupported depth %d", bpp)
	}
	if (base == 1 || base == 3) && bpp != 8 {
		return nil, fmt.Errorf("tga: unsupported depth %d for type %d", bpp, imgType)
	}
	pos := 18 + idLen
	var pal []color.NRGBA
	if cmType == 1 {
		eb := (cmBits + 7) / 8
		if pos+cmLen*eb > len(b) {
			return nil, fmt.Errorf("tga: truncated palette")
		}
		for i := 0; i < cmLen; i++ {
			e := b[pos+i*eb : pos+(i+1)*eb]
			switch eb {
			case 3:
				pal = append(pal, color.NRGBA{e[2], e[1], e[0], 255})
			case 4:
				pal = append(pal, color.NRGBA{e[2], e[1], e[0], e[3]})
			default:
				return nil, fmt.Errorf("tga: unsupported palette entry size %d", eb)
			}
		}
		pos += cmLen * eb
	} else {
		pos += cmLen * ((cmBits + 7) / 8)
	}
	bytesPP := bpp / 8
	n := w * h
	px := make([]byte, 0, n*bytesPP)
	data := b[pos:]
	if !rle {
		if len(data) < n*bytesPP {
			return nil, fmt.Errorf("tga: truncated pixels (%d of %d bytes)", len(data), n*bytesPP)
		}
		px = data[:n*bytesPP]
	} else {
		for i := 0; len(px) < n*bytesPP; {
			if i >= len(data) {
				return nil, fmt.Errorf("tga: truncated run-length data")
			}
			c := int(data[i])
			i++
			cnt := c&0x7f + 1
			if c&0x80 != 0 {
				if i+bytesPP > len(data) {
					return nil, fmt.Errorf("tga: truncated run")
				}
				for k := 0; k < cnt; k++ {
					px = append(px, data[i:i+bytesPP]...)
				}
				i += bytesPP
			} else {
				if i+cnt*bytesPP > len(data) {
					return nil, fmt.Errorf("tga: truncated raw run")
				}
				px = append(px, data[i:i+cnt*bytesPP]...)
				i += cnt * bytesPP
			}
		}
		px = px[:n*bytesPP]
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	topFirst := desc&0x20 != 0
	leftFirst := desc&0x10 == 0
	for y := 0; y < h; y++ {
		dy := y
		if !topFirst {
			dy = h - 1 - y
		}
		for x := 0; x < w; x++ {
			dx := x
			if !leftFirst {
				dx = w - 1 - x
			}
			p := px[(y*w+x)*bytesPP:]
			var c color.NRGBA
			switch {
			case base == 2 && bpp == 32:
				c = color.NRGBA{p[2], p[1], p[0], p[3]}
			case base == 2:
				c = color.NRGBA{p[2], p[1], p[0], 255}
			case base == 1:
				if int(p[0]) >= len(pal) {
					return nil, fmt.Errorf("tga: palette index %d of %d", p[0], len(pal))
				}
				c = pal[p[0]]
			default:
				c = color.NRGBA{p[0], p[0], p[0], 255}
			}
			img.SetNRGBA(dx, dy, c)
		}
	}
	return img, nil
}

// DecodeBMP reads the uncompressed 8-bit palette and 24-bit files the install
// ships. Rows are stored bottom up unless the height is negative.
func DecodeBMP(b []byte) (*image.NRGBA, error) {
	if len(b) < 54 || b[0] != 'B' || b[1] != 'M' {
		return nil, fmt.Errorf("bmp: bad header")
	}
	le := binary.LittleEndian
	off := int(le.Uint32(b[10:]))
	hdr := int(le.Uint32(b[14:]))
	w := int(int32(le.Uint32(b[18:])))
	h := int(int32(le.Uint32(b[22:])))
	bpp := int(le.Uint16(b[28:]))
	comp := le.Uint32(b[30:])
	if comp != 0 {
		return nil, fmt.Errorf("bmp: compression %d unsupported", comp)
	}
	topDown := h < 0
	if topDown {
		h = -h
	}
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("bmp: bad dimensions %dx%d", w, h)
	}
	var pal []color.NRGBA
	switch bpp {
	case 8:
		colors := int(le.Uint32(b[46:]))
		if colors == 0 {
			colors = 256
		}
		p := 14 + hdr
		if p+colors*4 > len(b) {
			return nil, fmt.Errorf("bmp: truncated palette")
		}
		for i := 0; i < colors; i++ {
			e := b[p+i*4:]
			pal = append(pal, color.NRGBA{e[2], e[1], e[0], 255})
		}
	case 24:
	default:
		return nil, fmt.Errorf("bmp: unsupported depth %d", bpp)
	}
	stride := (w*bpp/8 + 3) &^ 3
	if off+stride*h > len(b) {
		return nil, fmt.Errorf("bmp: truncated pixels")
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		dy := y
		if !topDown {
			dy = h - 1 - y
		}
		row := b[off+y*stride:]
		for x := 0; x < w; x++ {
			if bpp == 8 {
				if int(row[x]) >= len(pal) {
					return nil, fmt.Errorf("bmp: palette index %d of %d", row[x], len(pal))
				}
				img.SetNRGBA(x, dy, pal[row[x]])
			} else {
				img.SetNRGBA(x, dy, color.NRGBA{row[x*3+2], row[x*3+1], row[x*3], 255})
			}
		}
	}
	return img, nil
}
