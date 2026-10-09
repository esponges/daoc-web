package main

// Ground detail: the layers the game paints the terrain with up close.
//
// The colour atlas is one 256px tile per 8192-unit sector, 32 units to the
// texel, which is a smear under a character's feet. The game does not draw
// that up close. ter<zone>.mpk carries textures.csv, which lists for every
// sector the ground textures painted on it -- rock, dirt, two grasses,
// stones, snow, cobbles, pine needles, about a dozen -- each with how many
// times it repeats across the zone. Beside it are mask tiles,
// patch<x><y>-<n>.dds, 128px a sector, three layers to a tile in the red,
// green and blue channels: layer i of a sector is mask n = i/3, channel i%3.
// The first layer is solid everywhere and each later one is laid over the
// stack by its mask, the way paint layers are.
//
// That reading is checked rather than assumed. tex<zone>.mpk holds the
// game's own pre-blended picture of every sector, 512px, which is what the
// layers come to from a distance. Compositing the layers here, block-averaged
// to wash out the texture detail, and correlating with those tiles picks the
// mask orientation and the blend rule, and says how well they agree.

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"daocweb/internal/dds"
	"daocweb/internal/zone"
)

const (
	layerPx = 512 // every layer is resampled to this, to share one texture array
	maskPx  = 128 // mask tile size per sector
	maxSlot = 12  // four mask tiles of three channels
)

type splatOut struct {
	Layers  string   `json:"layers"`  // JPEG, layerPx wide, one layer under another
	LayerPx int      `json:"layerPx"` // side of one layer
	Names   []string `json:"names"`   // layer names, in strip order
	Masks   string   `json:"masks"`   // PNG, one plane under another
	MaskPx  int      `json:"maskPx"`  // mask texels per sector side
	Planes  int      `json:"planes"`  // mask tiles per sector
	Sectors int      `json:"sectors"` // sectors per zone side
	// Slots[sy*Sectors+sx] lists the sector's layers in paint order, each
	// as [layer index, repeats across the zone].
	Slots  [][][2]float64 `json:"slots"`
	Layout string         `json:"layout"` // how mask tiles were found to sit
	Agree  float64        `json:"agree"`  // correlation with the game's own tiles
}

type slot struct {
	layer int
	tiles float64
}

// maskLayout is one way a mask tile might sit in its sector.
type maskLayout struct {
	swapName     bool // patch<a><b>: a is y rather than x
	transpose    bool // texel (u, v) is sector (v, u)
	flipU, flipV bool
}

func (l maskLayout) String() string {
	return fmt.Sprintf("swapName=%-5v transpose=%-5v flipU=%-5v flipV=%-5v", l.swapName, l.transpose, l.flipU, l.flipV)
}

func buildSplat(game string, zoneNum, sectors int, outDir string) (*splatOut, error) {
	z, err := zone.Open(game, zoneNum)
	if err != nil {
		return nil, err
	}
	terArc, err := z.Archive("ter")
	if err != nil {
		return nil, err
	}
	raw, err := terArc.Read("textures.csv")
	if err != nil {
		return nil, err
	}

	// --- the layer table ---
	rd := csv.NewReader(bytes.NewReader(raw))
	rd.FieldsPerRecord = -1
	rd.TrimLeadingSpace = true
	rows, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("textures.csv: %w", err)
	}
	index := map[string]int{}
	var names []string
	slots := make([][]slot, sectors*sectors)
	odd := 0
	for _, r := range rows[1:] {
		if len(r) < 11 || strings.TrimSpace(r[2]) == "" {
			continue
		}
		px, e1 := strconv.Atoi(strings.TrimSpace(r[0]))
		py, e2 := strconv.Atoi(strings.TrimSpace(r[1]))
		if e1 != nil || e2 != nil || px < 0 || py < 0 || px >= sectors || py >= sectors {
			continue
		}
		name := strings.TrimSpace(r[2])
		// Rotation, offset and scale are all zero or one in the zones
		// looked at; say so if one is not, rather than draw it wrongly
		// without a word.
		if f(r[3]) != 0 || f(r[4]) != 0 || f(r[5]) != 0 || f(r[6]) != 1 || f(r[7]) != 1 {
			odd++
		}
		if strings.TrimSpace(r[8]) == "0" { // not visible
			continue
		}
		key := strings.ToLower(name)
		li, ok := index[key]
		if !ok {
			li = len(names)
			index[key] = li
			names = append(names, name)
		}
		s := &slots[py*sectors+px]
		*s = append(*s, slot{li, math.Max(1, f(r[9]))})
	}
	if odd > 0 {
		fmt.Printf("  WARNING: %d layer rows rotate, offset or scale, which is ignored\n", odd)
	}
	planes := 0
	for _, s := range slots {
		if len(s) > maxSlot {
			return nil, fmt.Errorf("a sector paints %d layers; at most %d fit", len(s), maxSlot)
		}
		planes = max(planes, (len(s)+2)/3)
	}
	fmt.Printf("  textures.csv: %d distinct layers, up to %d a sector (%d mask tiles)\n", len(names), planes*3, planes)

	// --- the layer textures, all resampled to one size ---
	layers := make([]*image.NRGBA, len(names))
	for i, n := range names {
		b, err := os.ReadFile(filepath.Join(game, "zones", "TerrainTex", n+".dds"))
		if err != nil {
			return nil, fmt.Errorf("layer %q: %w", n, err)
		}
		im, err := dds.Decode(b)
		if err != nil {
			return nil, fmt.Errorf("layer %q: %w", n, err)
		}
		layers[i] = resample(im, layerPx)
	}

	// --- the masks, raw per sector and plane ---
	type key struct{ a, b, n int }
	masks := map[key]*image.NRGBA{}
	for a := 0; a < sectors; a++ {
		for b := 0; b < sectors; b++ {
			for n := 0; n < planes; n++ {
				raw, err := terArc.Read(fmt.Sprintf("patch%02d%02d-%02d.dds", a, b, n))
				if err != nil {
					continue // a sector with fewer layers has fewer tiles
				}
				im, err := dds.Decode(raw)
				if err != nil {
					return nil, err
				}
				if im.Bounds().Dx() != maskPx || im.Bounds().Dy() != maskPx {
					return nil, fmt.Errorf("mask patch%02d%02d-%02d is %v, expected %dpx", a, b, n, im.Bounds().Max, maskPx)
				}
				masks[key{a, b, n}] = im
			}
		}
	}

	// maskAt reads sector (sx, sy)'s weight for slot i at sector texel (u, v)
	// under a layout.
	maskAt := func(l maskLayout, sx, sy, i, u, v int) float64 {
		a, b := sx, sy
		if l.swapName {
			a, b = sy, sx
		}
		im := masks[key{a, b, i / 3}]
		if im == nil {
			return 0
		}
		if l.transpose {
			u, v = v, u
		}
		if l.flipU {
			u = maskPx - 1 - u
		}
		if l.flipV {
			v = maskPx - 1 - v
		}
		return float64(im.Pix[(v*maskPx+u)*4+i%3]) / 255
	}

	// --- check against the game's own composite ---
	texArc, err := z.Archive("tex")
	if err != nil {
		return nil, err
	}
	// Each layer's average colour stands in for the layer: the comparison
	// is block-averaged, and a block spans several repeats of most layers.
	// The zone map (one repeat) is the exception, so it is sampled.
	mean := make([][3]float64, len(layers))
	for i, im := range layers {
		mean[i] = avg(im, 0, 0, layerPx, layerPx)
	}
	const blocks = 16 // per sector side
	type sample struct{ sx, sy int }
	var check []sample
	for sy := 0; sy < sectors; sy += 2 {
		for sx := 1; sx < sectors; sx += 2 {
			check = append(check, sample{sx, sy})
		}
	}
	ref := map[sample]*image.NRGBA{}
	for _, s := range check {
		raw, err := texArc.Read(fmt.Sprintf("tex%02d-%02d.dds", s.sx, s.sy))
		if err != nil {
			return nil, err
		}
		im, err := dds.Decode(raw)
		if err != nil {
			return nil, err
		}
		ref[s] = im
	}
	score := func(l maskLayout, normalise bool) float64 {
		var xs, ys []float64
		for _, s := range check {
			r := ref[s]
			rp := r.Bounds().Dx() / blocks
			for by := 0; by < blocks; by++ {
				for bx := 0; bx < blocks; bx++ {
					// Mask texel at the block centre.
					u := (bx*maskPx + maskPx/2) / blocks
					v := (by*maskPx + maskPx/2) / blocks
					var c [3]float64
					var wsum float64
					for i, sl := range slots[s.sy*sectors+s.sx] {
						w := maskAt(l, s.sx, s.sy, i, u, v)
						lc := mean[sl.layer]
						if sl.tiles <= 1 {
							zu := (float64(s.sx) + (float64(bx)+0.5)/blocks) / float64(sectors)
							zv := (float64(s.sy) + (float64(by)+0.5)/blocks) / float64(sectors)
							lc = px3(layers[sl.layer], zu, zv)
						}
						if normalise {
							for k := range c {
								c[k] += lc[k] * w
							}
							wsum += w
						} else {
							for k := range c {
								c[k] = c[k]*(1-w) + lc[k]*w
							}
						}
					}
					if normalise && wsum > 0 {
						for k := range c {
							c[k] /= wsum
						}
					}
					g := avg(r, bx*rp, by*rp, rp, rp)
					for k := 0; k < 3; k++ {
						xs = append(xs, c[k])
						ys = append(ys, g[k])
					}
				}
			}
		}
		return pearson(xs, ys)
	}
	var best maskLayout
	bestR, bestNorm := math.Inf(-1), false
	// The runner-up under each rule, to show the choice is not a near tie.
	second := map[bool]float64{false: math.Inf(-1), true: math.Inf(-1)}
	for _, sw := range []bool{false, true} {
		for _, tr := range []bool{false, true} {
			for _, fu := range []bool{false, true} {
				for _, fv := range []bool{false, true} {
					l := maskLayout{sw, tr, fu, fv}
					for _, norm := range []bool{false, true} {
						r := score(l, norm)
						if r > bestR {
							best, bestR, bestNorm = l, r, norm
						}
						if l != (maskLayout{}) {
							second[norm] = math.Max(second[norm], r)
						}
					}
				}
			}
		}
	}
	fmt.Printf("  masks: as named and unflipped, painted in order r = %+.4f, weighted r = %+.4f; best other layout %+.4f / %+.4f\n",
		score(maskLayout{}, false), score(maskLayout{}, true), second[false], second[true])
	rule := "painted in order"
	if bestNorm {
		rule = "weighted average"
	}
	fmt.Printf("  masks: %s, %s  (r = %+.4f against tex%03d.mpk)\n", best, rule, bestR, zoneNum)
	if bestNorm {
		return nil, fmt.Errorf("the layers fit best as a weighted average, which the viewer does not draw")
	}
	if bestR < 0.6 {
		fmt.Printf("  WARNING: weak agreement with the game's own tiles; ground detail is unverified\n")
	}

	// --- write: layers as one JPEG strip ---
	strip := image.NewNRGBA(image.Rect(0, 0, layerPx, layerPx*len(layers)))
	for i, im := range layers {
		copy(strip.Pix[i*layerPx*layerPx*4:], im.Pix)
	}
	if err := writeJPEG(filepath.Join(outDir, "layers.jpg"), strip); err != nil {
		return nil, err
	}

	// --- masks: one plane per tile index, sectors stitched in zone order,
	// each already turned to the zone's orientation ---
	side := sectors * maskPx
	mstrip := image.NewNRGBA(image.Rect(0, 0, side, side*planes))
	for n := 0; n < planes; n++ {
		for sy := 0; sy < sectors; sy++ {
			for sx := 0; sx < sectors; sx++ {
				for v := 0; v < maskPx; v++ {
					for u := 0; u < maskPx; u++ {
						o := ((n*side+sy*maskPx+v)*side + sx*maskPx + u) * 4
						for c := 0; c < 3; c++ {
							mstrip.Pix[o+c] = uint8(math.Round(255 * maskAt(best, sx, sy, n*3+c, u, v)))
						}
						mstrip.Pix[o+3] = 255
					}
				}
			}
		}
	}
	if err := writePNG(filepath.Join(outDir, "masks.png"), mstrip); err != nil {
		return nil, err
	}

	out := &splatOut{
		Layers: "layers.jpg", LayerPx: layerPx, Names: names,
		Masks: "masks.png", MaskPx: maskPx, Planes: planes, Sectors: sectors,
		Layout: best.String() + ", " + rule, Agree: math.Round(bestR*1e4) / 1e4,
	}
	for _, s := range slots {
		row := [][2]float64{}
		for _, sl := range s {
			row = append(row, [2]float64{float64(sl.layer), sl.tiles})
		}
		out.Slots = append(out.Slots, row)
	}
	return out, nil
}

func f(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

// resample scales a square image to n x n: a box filter going down, bilinear
// going up.
func resample(im *image.NRGBA, n int) *image.NRGBA {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	out := image.NewNRGBA(image.Rect(0, 0, n, n))
	if w >= n && h >= n {
		fx, fy := w/n, h/n
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				c := avg(im, x*fx, y*fy, fx, fy)
				out.SetNRGBA(x, y, color.NRGBA{uint8(c[0] * 255), uint8(c[1] * 255), uint8(c[2] * 255), 255})
			}
		}
		return out
	}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			c := px3(im, (float64(x)+0.5)/float64(n), (float64(y)+0.5)/float64(n))
			out.SetNRGBA(x, y, color.NRGBA{uint8(c[0] * 255), uint8(c[1] * 255), uint8(c[2] * 255), 255})
		}
	}
	return out
}

// avg is the mean colour of a rectangle, 0..1.
func avg(im *image.NRGBA, x0, y0, w, h int) [3]float64 {
	var c [3]float64
	stride := im.Stride
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			o := y*stride + x*4
			c[0] += float64(im.Pix[o])
			c[1] += float64(im.Pix[o+1])
			c[2] += float64(im.Pix[o+2])
		}
	}
	n := float64(w*h) * 255
	return [3]float64{c[0] / n, c[1] / n, c[2] / n}
}

// px3 samples an image at (u, v) in 0..1, nearest texel.
func px3(im *image.NRGBA, u, v float64) [3]float64 {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	x := min(w-1, max(0, int(u*float64(w))))
	y := min(h-1, max(0, int(v*float64(h))))
	o := y*im.Stride + x*4
	return [3]float64{float64(im.Pix[o]) / 255, float64(im.Pix[o+1]) / 255, float64(im.Pix[o+2]) / 255}
}

func pearson(xs, ys []float64) float64 {
	n := float64(len(xs))
	var mx, my float64
	for i := range xs {
		mx += xs[i]
		my += ys[i]
	}
	mx /= n
	my /= n
	var sxy, sxx, syy float64
	for i := range xs {
		dx, dy := xs[i]-mx, ys[i]-my
		sxy += dx * dy
		sxx += dx * dx
		syy += dy * dy
	}
	if sxx == 0 || syy == 0 {
		return 0
	}
	return sxy / math.Sqrt(sxx*syy)
}

func writeJPEG(path string, im image.Image) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	var w io.Writer = fh
	return jpeg.Encode(w, im, &jpeg.Options{Quality: 90})
}
