package ground

// Grass: the small plants the game scatters over the ground near the camera.
//
// The region's entry in zones.dat names a sprite table and its atlas
// (region 100: grasscsv=grass_mid.csv, grassmap=grass_mid.dds, both in
// zones/textures). Each table row is one sprite: the group it belongs to,
// its chance of being picked within that group, a shape, a height, length
// and width each with a random spread, a scale spread in percent, its
// rectangle in the atlas, and an optional tint.
//
// The zone says where each group grows: dat<zone>.mpk's grassmap.pcx holds,
// per heightmap cell, ten times the group number -- 30 is group 3, the
// waterside plants, and every 30 lies on the lake shore; 110 is group 11,
// the rocks, and sits high on the slopes. densemap.pcx says how thickly,
// 0..255, and is zero over most of the snow.

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"daocweb/internal/zone"
)

// Sprite is one row of a grass table.
type Sprite struct {
	Name   string     `json:"name"`
	Group  int        `json:"group"`
	Chance float64    `json:"chance"`
	Shape  int        `json:"shape"`
	Height [2]float64 `json:"height"` // base, spread
	Length [2]float64 `json:"length"`
	Width  [2]float64 `json:"width"`
	Scale  float64    `json:"scale"` // spread, percent
	Rect   [4]int     `json:"rect"`  // x1, x2, y1, y2 in atlas pixels
	Tint   *[3]int    `json:"tint,omitempty"`
}

// Grass is a zone's grass, ready to write out.
type Grass struct {
	Table    string // the region's sprite table, in zones/textures
	AtlasDir string // where the sprite atlas is
	Atlas    string // the sprite atlas's file name
	Sprites  []Sprite
	Groups   int
	// Cells is a byte pair per cell, group then density, row by row,
	// CellsPerSide to a row: the heightmap's grid, or a multiple of it.
	Cells        []byte
	CellsPerSide int
}

// ErrNoGrassMap is returned for a zone whose data names no grass map: it
// grows no grass, which is not a failure.
var ErrNoGrassMap = fmt.Errorf("no grass map")

// ReadGrass reads zone z's grass for a heightmap of grid x grid cells,
// saying what it finds through logf.
func ReadGrass(game string, z *zone.Zone, grid int, logf func(string, ...any)) (*Grass, error) {
	_, reg, err := zone.RegionOf(game, z.Num)
	if err != nil {
		return nil, err
	}
	table, atlas := reg["grasscsv"], reg["grassmap"]
	if table == "" || atlas == "" {
		return nil, fmt.Errorf("the region names no grass")
	}
	dir := filepath.Join(game, "zones", "textures")
	raw, err := os.ReadFile(filepath.Join(dir, table))
	if err != nil {
		return nil, err
	}
	rd := csv.NewReader(bytes.NewReader(raw))
	rd.FieldsPerRecord = -1
	rows, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", table, err)
	}
	out := &Grass{Table: table, AtlasDir: dir, Atlas: atlas}
	for _, r := range rows {
		if len(r) < 16 {
			continue
		}
		if _, err := strconv.Atoi(strings.TrimSpace(r[0])); err != nil {
			continue // the header rows, and the empty padding at the end
		}
		n := func(i int) float64 { return num(r[i]) }
		s := Sprite{
			Name: strings.TrimSpace(r[1]), Group: int(n(2)), Chance: n(3), Shape: int(n(4)),
			Height: [2]float64{n(5), n(6)}, Length: [2]float64{n(7), n(8)},
			Width: [2]float64{n(9), n(10)}, Scale: n(11),
			Rect: [4]int{int(n(12)), int(n(13)), int(n(14)), int(n(15))},
		}
		if len(r) >= 19 && strings.TrimSpace(r[16]) != "" {
			t := [3]int{int(n(16)), int(n(17)), int(n(18))}
			s.Tint = &t
		}
		out.Sprites = append(out.Sprites, s)
	}
	if len(out.Sprites) == 0 {
		return nil, fmt.Errorf("%s has no sprites", table)
	}

	// Where each group grows, and how thickly.
	if !z.Dat.Has("grassmap.pcx") {
		return nil, ErrNoGrassMap
	}
	gm, err := z.PCX("grassmap.pcx")
	if err != nil {
		return nil, err
	}
	dm, err := z.PCX("densemap.pcx")
	if err != nil {
		return nil, err
	}
	// Classic zones draw both maps at the heightmap's resolution; the
	// Shrouded Isles and the frontiers at twice it, or (frontier Odin's
	// Gate) one map at each. Cells are kept at the finer of the two, the
	// coarser read by nearest cell.
	side := max(gm.W, dm.W)
	for _, m := range [][2]int{{gm.W, gm.H}, {dm.W, dm.H}} {
		if m[0] != m[1] || m[0] < grid || m[0]%grid != 0 || side%m[0] != 0 {
			return nil, fmt.Errorf("grass maps %dx%d and %dx%d, heightmap %d", gm.W, gm.H, dm.W, dm.H, grid)
		}
	}
	out.CellsPerSide = side
	at := func(w, x, y int) int { return (y*w/side)*w + x*w/side }
	groups := map[int]bool{}
	for _, s := range out.Sprites {
		groups[s.Group] = true
	}
	out.Groups = len(groups)
	cells := make([]byte, side*side*2)
	missing := map[int]int{}
	for y := 0; y < side; y++ {
		for x := 0; x < side; x++ {
			i := y*side + x
			v := int(gm.Pal[gm.Pix[at(gm.W, x, y)]][0])
			// A group owns its band of ten. Classic zones paint the
			// band's foot, ten times the group, with soft edges between
			// (Gotar has every value from 0 to 110); the Isles and the
			// frontiers paint near the middle -- 5, 45, 105, 155 -- and
			// every value in the install falls in a band the table has.
			g := v / 10
			if !groups[g] {
				missing[v]++
			}
			cells[i*2] = byte(g)
			cells[i*2+1] = dm.Pal[dm.Pix[at(dm.W, x, y)]][0]
		}
	}
	if len(missing) > 0 {
		logf("  WARNING: grassmap values with no group in %s: %v\n", table, missing)
	}
	out.Cells = cells
	return out, nil
}

// num parses a table number, 0 when empty or not a number.
func num(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}
