package scenery

// Textures stored inside a model. Most models name their textures and the
// files sit beside them; some carry the pixels themselves in an NiPixelData
// block -- the Avalon pier, the evil mushroom trees, demonRoom3's walls. Those
// are decoded here and given a name of their own, model and block, so they
// are written and loaded like any other texture.

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"

	"daocweb/internal/dds"
	"daocweb/internal/nif"
)

// texNamer names the textures one model uses, decoding embedded ones.
type texNamer struct {
	f        *nif.File
	prefix   string // the model's file name, lower-cased, without extension
	embedded map[string]image.Image
}

func newTexNamer(f *nif.File, file string) *texNamer {
	b := strings.ToLower(filepath.Base(file))
	return &texNamer{f: f, prefix: strings.TrimSuffix(b, filepath.Ext(b))}
}

// name is the texture a source block stands for, or "" for none. An embedded
// texture that will not decode falls back to the file name the block also
// carries, when it has one.
func (t *texNamer) name(src *nif.SourceTexture) string {
	if src.UseExternal != 0 {
		if src.FileName == "" {
			return ""
		}
		return TexName(src.FileName)
	}
	n := fmt.Sprintf("%s_%d", t.prefix, src.PixelData)
	if _, ok := t.embedded[n]; ok {
		return n
	}
	pd, ok := t.f.Block(src.PixelData).(*nif.PixelData)
	if ok {
		if img, err := DecodePixelData(t.f, pd); err == nil {
			if t.embedded == nil {
				t.embedded = map[string]image.Image{}
			}
			t.embedded[n] = img
			return n
		}
	}
	if src.FileName != "" {
		return TexName(src.FileName)
	}
	return ""
}

// DecodePixelData decodes the top level of an NiPixelData: 24- and 32-bit
// colour, 8-bit palettised, and DXT1, DXT3 and DXT5.
func DecodePixelData(f *nif.File, p *nif.PixelData) (*image.NRGBA, error) {
	if len(p.Mipmaps) == 0 {
		return nil, fmt.Errorf("pixel data has no levels")
	}
	m := p.Mipmaps[0]
	w, h := int(m.Width), int(m.Height)
	if w <= 0 || h <= 0 || w > 4096 || h > 4096 {
		return nil, fmt.Errorf("pixel data is %dx%d", w, h)
	}
	data := p.Pixels
	if int(m.Offset) > len(data) {
		return nil, fmt.Errorf("level offset %d past %d bytes", m.Offset, len(data))
	}
	data = data[m.Offset:]
	switch p.Format {
	case 4:
		return dds.DecodeDXT(data, w, h, "DXT1")
	case 5:
		return dds.DecodeDXT(data, w, h, "DXT3")
	case 6:
		return dds.DecodeDXT(data, w, h, "DXT5")
	}

	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	var bpp int
	var pal *nif.Palette
	switch p.Format {
	case 0:
		bpp = 3
	case 1:
		bpp = 4
	case 2:
		bpp = 1
		var ok bool
		if pal, ok = f.Block(p.Palette).(*nif.Palette); !ok || len(pal.Colors) == 0 {
			return nil, fmt.Errorf("palettised pixel data without a palette")
		}
	default:
		return nil, fmt.Errorf("pixel format %d not supported", p.Format)
	}
	if len(data) < w*h*bpp {
		return nil, fmt.Errorf("%dx%d at %d bytes a pixel needs %d bytes, have %d", w, h, bpp, w*h*bpp, len(data))
	}
	for i := 0; i < w*h; i++ {
		o := i * 4
		switch bpp {
		case 1:
			c := [4]byte{0, 0, 0, 255}
			if k := int(data[i]); k < len(pal.Colors) {
				c = pal.Colors[k]
			}
			if pal.HasAlpha == 0 {
				c[3] = 255
			}
			copy(img.Pix[o:o+4], c[:])
		case 3:
			// The masks put red in the low byte: R, G, B in memory order.
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = data[i*3], data[i*3+1], data[i*3+2], 255
		case 4:
			copy(img.Pix[o:o+4], data[i*4:i*4+4])
		}
	}
	return img, nil
}
