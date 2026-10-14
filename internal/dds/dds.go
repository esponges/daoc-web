// Package dds decodes the DDS textures DAoC ships. Only what this pipeline
// needs is implemented: DXT1/3/5 blocks and plain uncompressed RGB(A).
package dds

import (
	"encoding/binary"
	"fmt"
	"image"
	"math/bits"
)

// Decode parses a DDS file into an RGBA image, using only the top mip level.
func Decode(b []byte) (*image.NRGBA, error) {
	if len(b) < 128 {
		return nil, fmt.Errorf("dds: shorter than header (%d bytes)", len(b))
	}
	if string(b[:4]) != "DDS " {
		return nil, fmt.Errorf("dds: bad magic %q", b[:4])
	}
	le := binary.LittleEndian
	h := int(le.Uint32(b[12:]))
	w := int(le.Uint32(b[16:]))
	pfFlags := le.Uint32(b[80:])
	fourCC := string(b[84:88])
	rgbBits := int(le.Uint32(b[88:]))
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("dds: bad dimensions %dx%d", w, h)
	}
	data := b[128:]
	// NRGBA, not RGBA: DDS alpha is straight, while image.RGBA is defined as
	// premultiplied. Writing straight values into an RGBA would make the PNG
	// encoder divide them out again and blow the colours toward white.
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	const (
		pfFourCC = 0x4
		pfRGB    = 0x40
	)
	switch {
	case pfFlags&pfFourCC != 0 && (fourCC == "DXT1" || fourCC == "DXT3" || fourCC == "DXT5"):
		return decodeDXT(img, data, w, h, fourCC)
	case (pfFlags&pfRGB != 0 || pfFlags == 0x41) && rgbBits == 16:
		masks := [4]uint32{le.Uint32(b[92:]), le.Uint32(b[96:]), le.Uint32(b[100:]), le.Uint32(b[104:])}
		if pfFlags&0x1 == 0 {
			masks[3] = 0 // no alpha
		}
		return decode16(img, data, w, h, masks)
	case pfFlags&pfRGB != 0 || pfFlags == 0x41:
		return decodeRGB(img, data, w, h, rgbBits)
	default:
		return nil, fmt.Errorf("dds: unsupported pixel format (flags=0x%x fourCC=%q bits=%d)", pfFlags, fourCC, rgbBits)
	}
}

func decodeRGB(img *image.NRGBA, data []byte, w, h, bits int) (*image.NRGBA, error) {
	bytesPP := bits / 8
	if bytesPP != 3 && bytesPP != 4 {
		return nil, fmt.Errorf("dds: unsupported uncompressed depth %d", bits)
	}
	if need := w * h * bytesPP; len(data) < need {
		return nil, fmt.Errorf("dds: want %d bytes of pixels, have %d", need, len(data))
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s := (y*w + x) * bytesPP
			o := img.PixOffset(x, y)
			// DDS stores BGRA.
			img.Pix[o+0] = data[s+2]
			img.Pix[o+1] = data[s+1]
			img.Pix[o+2] = data[s+0]
			if bytesPP == 4 {
				img.Pix[o+3] = data[s+3]
			} else {
				img.Pix[o+3] = 0xFF
			}
		}
	}
	return img, nil
}

func decodeDXT(img *image.NRGBA, data []byte, w, h int, fourCC string) (*image.NRGBA, error) {
	blockSize := 16
	if fourCC == "DXT1" {
		blockSize = 8
	}
	bw, bh := (w+3)/4, (h+3)/4
	if need := bw * bh * blockSize; len(data) < need {
		return nil, fmt.Errorf("dds: %s wants %d bytes of blocks, have %d", fourCC, need, len(data))
	}

	var c [4][4]byte // RGBA for the four palette slots
	for by := 0; by < bh; by++ {
		for bx := 0; bx < bw; bx++ {
			blk := data[(by*bw+bx)*blockSize:]
			alpha := blk
			color := blk
			if fourCC != "DXT1" {
				color = blk[8:] // DXT3/5 put 8 bytes of alpha first
			}

			c0 := binary.LittleEndian.Uint16(color[0:])
			c1 := binary.LittleEndian.Uint16(color[2:])
			r0, g0, b0 := rgb565(c0)
			r1, g1, b1 := rgb565(c1)
			c[0] = [4]byte{r0, g0, b0, 0xFF}
			c[1] = [4]byte{r1, g1, b1, 0xFF}
			if c0 > c1 || fourCC != "DXT1" {
				// Four-colour block: two interpolated midpoints.
				c[2] = [4]byte{lerp(r0, r1, 1, 3), lerp(g0, g1, 1, 3), lerp(b0, b1, 1, 3), 0xFF}
				c[3] = [4]byte{lerp(r0, r1, 2, 3), lerp(g0, g1, 2, 3), lerp(b0, b1, 2, 3), 0xFF}
			} else {
				// Three-colour block: one midpoint plus transparent black.
				c[2] = [4]byte{lerp(r0, r1, 1, 2), lerp(g0, g1, 1, 2), lerp(b0, b1, 1, 2), 0xFF}
				c[3] = [4]byte{0, 0, 0, 0}
			}

			bits := binary.LittleEndian.Uint32(color[4:])
			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					x, y := bx*4+px, by*4+py
					if x >= w || y >= h {
						continue
					}
					sel := (bits >> uint(2*(py*4+px))) & 3
					o := img.PixOffset(x, y)
					img.Pix[o+0] = c[sel][0]
					img.Pix[o+1] = c[sel][1]
					img.Pix[o+2] = c[sel][2]
					img.Pix[o+3] = c[sel][3]
					switch fourCC {
					case "DXT3":
						// 4 bits of alpha per pixel, two pixels per byte.
						i := py*4 + px
						nib := alpha[i/2]
						if i%2 == 1 {
							nib >>= 4
						}
						img.Pix[o+3] = (nib & 0x0F) * 17
					case "DXT5":
						img.Pix[o+3] = dxt5Alpha(alpha, py*4+px)
					}
				}
			}
		}
	}
	return img, nil
}

func dxt5Alpha(blk []byte, i int) byte {
	a0, a1 := blk[0], blk[1]
	var a [8]byte
	a[0], a[1] = a0, a1
	if a0 > a1 {
		for k := 1; k <= 6; k++ {
			a[k+1] = byte((int(a0)*(7-k) + int(a1)*k) / 7)
		}
	} else {
		for k := 1; k <= 4; k++ {
			a[k+1] = byte((int(a0)*(5-k) + int(a1)*k) / 5)
		}
		a[6], a[7] = 0, 255
	}
	// 3 bits per pixel packed across the 6 bytes following a0/a1. The last
	// pixels straddle byte 7, so the high byte must be read as zero rather
	// than running on into the colour block that follows.
	bitPos := uint(3 * i)
	lo := uint32(blk[2+bitPos/8])
	var hi uint32
	if n := 2 + bitPos/8 + 1; n < 8 {
		hi = uint32(blk[n])
	}
	idx := (lo | hi<<8) >> (bitPos % 8) & 7
	return a[idx]
}

func rgb565(v uint16) (r, g, b byte) {
	r = byte((v>>11)&0x1F) << 3
	g = byte((v>>5)&0x3F) << 2
	b = byte(v&0x1F) << 3
	// Replicate high bits into the low bits so full-scale values reach 255.
	return r | r>>5, g | g>>6, b | b>>5
}

func lerp(a, b byte, num, den int) byte {
	return byte((int(a)*(den-num) + int(b)*num) / den)
}

// DecodeDXT decodes bare DXT1, DXT3 or DXT5 blocks, as NIFs store them
// inside NiPixelData, without a DDS header around them.
func DecodeDXT(data []byte, w, h int, fourCC string) (*image.NRGBA, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("dds: bad dimensions %dx%d", w, h)
	}
	return decodeDXT(image.NewNRGBA(image.Rect(0, 0, w, h)), data, w, h, fourCC)
}

// decode16 reads 16-bit pixels by the header's channel masks -- 565, 4444
// and 1555 alike; Jordheim's m_chain is one. A channel's bits are widened
// to eight by repeating them, so full stays full. Without an alpha mask the
// pixels are opaque.
func decode16(img *image.NRGBA, data []byte, w, h int, masks [4]uint32) (*image.NRGBA, error) {
	if need := w * h * 2; len(data) < need {
		return nil, fmt.Errorf("dds: want %d bytes of pixels, have %d", need, len(data))
	}
	channel := func(v, mask uint32) byte {
		if mask == 0 {
			return 0xFF
		}
		shift := bits.TrailingZeros32(mask)
		width := bits.OnesCount32(mask)
		c := (v & mask) >> shift
		out, n := uint32(0), 0
		for ; n < 8; n += width {
			out = out<<width | c
		}
		return byte(out >> (n - 8))
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint32(binary.LittleEndian.Uint16(data[(y*w+x)*2:]))
			o := img.PixOffset(x, y)
			img.Pix[o+0] = channel(v, masks[0])
			img.Pix[o+1] = channel(v, masks[1])
			img.Pix[o+2] = channel(v, masks[2])
			img.Pix[o+3] = channel(v, masks[3])
		}
	}
	return img, nil
}
