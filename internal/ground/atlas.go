package ground

import (
	"fmt"
	"image"
	"image/draw"

	"daocweb/internal/dds"
	"daocweb/internal/zone"
)

// Reference is the game's own pre-blended picture of every sector of a
// zone, from tex<zone>.mpk. The splat check scores ground detail against
// it; it also says how the colour atlas's tiles are laid out, since both
// are pictures of the same sectors.
type Reference struct {
	sectors int
	tiles   map[[2]int]*image.NRGBA // by sector x, y
}

// ReadReference reads every sector's tile of a zone of sectors x sectors.
func ReadReference(z *zone.Zone, sectors int) (*Reference, error) {
	arc, err := z.Archive("tex")
	if err != nil {
		return nil, err
	}
	r := &Reference{sectors: sectors, tiles: map[[2]int]*image.NRGBA{}}
	for sy := 0; sy < sectors; sy++ {
		for sx := 0; sx < sectors; sx++ {
			im, err := readTile(arc, sx, sy)
			if err != nil {
				return nil, err
			}
			r.tiles[[2]int{sx, sy}] = im
		}
	}
	return r, nil
}

// Agreement correlates an atlas, block-averaged, with the reference tiles:
// near 1 when the atlas shows each sector where the game does.
func (r *Reference) Agreement(atlas *image.NRGBA) (float64, error) {
	const blocks = 8 // per sector side
	side := atlas.Bounds().Dx()
	if side%(r.sectors*blocks) != 0 {
		return 0, fmt.Errorf("a %dpx atlas does not divide into %d sectors of %d blocks", side, r.sectors, blocks)
	}
	ap := side / r.sectors / blocks
	var xs, ys []float64
	for key, tile := range r.tiles {
		tp := tile.Bounds().Dx() / blocks
		for by := 0; by < blocks; by++ {
			for bx := 0; bx < blocks; bx++ {
				a := avg(atlas, (key[0]*blocks+bx)*ap, (key[1]*blocks+by)*ap, ap, ap)
				t := avg(tile, bx*tp, by*tp, tp, tp)
				for k := 0; k < 3; k++ {
					xs = append(xs, a[k])
					ys = append(ys, t[k])
				}
			}
		}
	}
	return pearson(xs, ys), nil
}

// LOD is a zone's colour atlas in pieces: lod<zone>.mpk's tiles, one per
// sector, lod<x>-<y>.dds.
type LOD struct {
	Tiles    map[[2]int]*image.NRGBA
	SX, SY   int // sectors across and down
	TileSize int // pixels a tile side
}

// ReadLOD reads a zone's LOD tiles for sx x sy sectors.
func ReadLOD(z *zone.Zone, sx, sy int) (*LOD, error) {
	arc, err := z.Archive("lod")
	if err != nil {
		return nil, err
	}
	l := &LOD{Tiles: map[[2]int]*image.NRGBA{}, SX: sx, SY: sy}
	for a := 0; a < sx; a++ {
		for b := 0; b < sy; b++ {
			name := fmt.Sprintf("lod%02d-%02d.dds", a, b)
			raw, err := arc.Read(name)
			if err != nil {
				return nil, err
			}
			im, err := dds.Decode(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			if l.TileSize == 0 {
				l.TileSize = im.Bounds().Dx()
			} else if im.Bounds().Dx() != l.TileSize {
				return nil, fmt.Errorf("%s is %dpx, expected %dpx", name, im.Bounds().Dx(), l.TileSize)
			}
			l.Tiles[[2]int{a, b}] = im
		}
	}
	return l, nil
}

// Layout is one way the LOD tiles might be laid out: tile (a, b) of the
// archive is sector (a, b), or (b, a) when swapped, counted from either
// edge. The first, Layouts()[0], is the tiles as named.
type Layout struct{ Swap, FlipX, FlipY bool }

func (l Layout) String() string {
	return fmt.Sprintf("swap=%-5v flipX=%-5v flipY=%-5v", l.Swap, l.FlipX, l.FlipY)
}

// Layouts is all eight, the tiles as named first.
func Layouts() []Layout {
	var out []Layout
	for _, swap := range []bool{false, true} {
		for _, fx := range []bool{false, true} {
			for _, fy := range []bool{false, true} {
				out = append(out, Layout{swap, fx, fy})
			}
		}
	}
	return out
}

// Stitch assembles the tiles into one atlas under a layout.
func (l *LOD) Stitch(lay Layout) *image.NRGBA {
	atlasPx := l.TileSize * l.SX
	out := image.NewNRGBA(image.Rect(0, 0, atlasPx, atlasPx))
	for key, im := range l.Tiles {
		cx, cy := key[0], key[1]
		if lay.Swap {
			cx, cy = cy, cx
		}
		if lay.FlipX {
			cx = l.SX - 1 - cx
		}
		if lay.FlipY {
			cy = l.SY - 1 - cy
		}
		r := image.Rect(cx*l.TileSize, cy*l.TileSize, (cx+1)*l.TileSize, (cy+1)*l.TileSize)
		draw.Draw(out, r, im, im.Bounds().Min, draw.Src)
	}
	return out
}

// LayoutMargin is how much better than the tiles as named another layout
// must agree with the reference tiles to be taken instead. The layout is
// the file format's, and every zone the reference settles clearly has the
// tiles as named.
const LayoutMargin = 0.1

// LayoutCheck is how a zone's atlas layout was settled by its reference
// tiles.
type LayoutCheck struct {
	Chosen Layout
	Agree  float64 // the chosen layout's agreement with the reference tiles
	Other  float64 // the best other layout's
	// Agreements, by Layouts() order.
	Agreements []float64
}

// CheckLayout scores every layout of the atlas against the reference tiles
// and settles on one.
func CheckLayout(l *LOD, ref *Reference) (*LayoutCheck, error) {
	c := &LayoutCheck{}
	lays := Layouts()
	for _, lay := range lays {
		r, err := ref.Agreement(l.Stitch(lay))
		if err != nil {
			return nil, err
		}
		c.Agreements = append(c.Agreements, r)
	}
	// The best of the other seven, against the tiles as named.
	other := 1
	for i := 2; i < len(lays); i++ {
		if c.Agreements[i] > c.Agreements[other] {
			other = i
		}
	}
	if c.Agreements[other]-c.Agreements[0] >= LayoutMargin {
		c.Chosen, c.Agree, c.Other = lays[other], c.Agreements[other], c.Agreements[0]
	} else {
		c.Chosen, c.Agree, c.Other = lays[0], c.Agreements[0], c.Agreements[other]
	}
	return c, nil
}
