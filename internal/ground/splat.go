// Package ground reads what a zone paints its terrain with up close: the
// ground detail layers and the grass.
package ground

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
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"daocweb/internal/dds"
	"daocweb/internal/zone"
)

const (
	LayerPx = 512 // every layer is resampled to this, to share one texture array
	MaskPx  = 128 // mask tile size per sector
	maxSlot = 12  // four mask tiles of three channels

	// layoutMargin is how much better than the named mask layout another
	// must agree with the game's tiles to be taken instead.
	layoutMargin = 0.1

	layerPx = LayerPx
	maskPx  = MaskPx
)

// Splat is a zone's ground detail, ready to write out.
type Splat struct {
	Names   []string     // layer names, in strip order
	Strip   *image.NRGBA // the layers, LayerPx wide, one under another
	Masks   *image.NRGBA // one plane per mask tile index, sectors stitched in zone order
	Planes  int          // mask tiles per sector
	Sectors int          // sectors per zone side
	// Slots[sy*Sectors+sx] lists the sector's layers in paint order, each
	// as [layer index, repeats across the zone].
	Slots  [][][2]float64
	Layout string  // how mask tiles were found to sit, and the blend rule
	Agree  float64 // correlation with the game's own tiles
	// Weighted is set where the layers come to the game's tiles as an
	// average weighted by their masks, rather than painted one over another.
	Weighted bool
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

// ReadSplat reads and checks zone z's ground detail for a zone of sectors x
// sectors, saying what it finds through logf.
func ReadSplat(game string, z *zone.Zone, sectors int, logf func(string, ...any)) (*Splat, error) {
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
		logf("  WARNING: %d layer rows rotate, offset or scale, which is ignored\n", odd)
	}
	planes := 0
	for _, s := range slots {
		if len(s) > maxSlot {
			return nil, fmt.Errorf("a sector paints %d layers; at most %d fit", len(s), maxSlot)
		}
		planes = max(planes, (len(s)+2)/3)
	}
	logf("  textures.csv: %d distinct layers, up to %d a sector (%d mask tiles)\n", len(names), planes*3, planes)

	// --- the layer textures, all resampled to one size ---
	// A frontier zone's own layers are under frontiers/, and it may borrow
	// the classic ones as well.
	texDirs := []string{filepath.Join(game, "zones", "TerrainTex")}
	if z.Frontier {
		texDirs = append([]string{filepath.Join(game, "frontiers", "zones", "TerrainTex")}, texDirs...)
	}
	layers := make([]*image.NRGBA, len(names))
	for i, n := range names {
		b, err := readFirst(texDirs, n+".dds")
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
				if im.Bounds().Dx() != im.Bounds().Dy() {
					return nil, fmt.Errorf("mask patch%02d%02d-%02d is %v, not square", a, b, n, im.Bounds().Max)
				}
				// Frontier zones' masks are 64px; the viewer wants one size.
				masks[key{a, b, n}] = nearest(im, maskPx)
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
		im, err := readTile(texArc, s.sx, s.sy)
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
	named := map[bool]float64{false: score(maskLayout{}, false), true: score(maskLayout{}, true)}
	logf("  masks: as named and unflipped, painted in order r = %+.4f, weighted r = %+.4f; best other layout %+.4f / %+.4f\n",
		named[false], named[true], second[false], second[true])
	// How the masks sit is the file format's, not each zone's, and every
	// zone that agrees well has them as named. A zone where another layout
	// edges ahead is one the check cannot tell apart (in Midgard, 103 by
	// 0.008, 107 by 0.05 and 158 by 0.013), so it keeps the named layout
	// unless another wins clearly.
	if best != (maskLayout{}) {
		norm := named[true] > named[false]
		if r := named[norm]; bestR-r < layoutMargin {
			logf("  masks: %s fits better by only %+.4f; keeping them as named\n", best, bestR-r)
			best, bestR, bestNorm = maskLayout{}, r, norm
		}
	}
	rule := "painted in order"
	if bestNorm {
		rule = "weighted average"
	}
	logf("  masks: %s, %s  (r = %+.4f against tex%03d.mpk)\n", best, rule, bestR, z.Num)
	if bestR < 0.6 {
		logf("  WARNING: weak agreement with the game's own tiles; ground detail is unverified\n")
	}

	// --- the layers as one strip ---
	strip := image.NewNRGBA(image.Rect(0, 0, layerPx, layerPx*len(layers)))
	for i, im := range layers {
		copy(strip.Pix[i*layerPx*layerPx*4:], im.Pix)
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

	out := &Splat{
		Names: names, Strip: strip, Masks: mstrip, Planes: planes, Sectors: sectors,
		Layout: best.String() + ", " + rule, Agree: math.Round(bestR*1e4) / 1e4,
		Weighted: bestNorm,
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

// nearest scales a square image to n x n by nearest texel, which keeps a
// mask's edges where they are rather than blurring one layer into the next.
func nearest(im *image.NRGBA, n int) *image.NRGBA {
	w := im.Bounds().Dx()
	if w == n {
		return im
	}
	out := image.NewNRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		sy := y * w / n
		for x := 0; x < n; x++ {
			sx := x * w / n
			copy(out.Pix[y*out.Stride+x*4:][:4], im.Pix[sy*im.Stride+sx*4:])
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

// readFirst reads name from the first of dirs that has it.
func readFirst(dirs []string, name string) ([]byte, error) {
	for _, d := range dirs {
		b, err := os.ReadFile(filepath.Join(d, name))
		if err == nil {
			return b, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("%s not in %s", name, strings.Join(dirs, " or "))
}
