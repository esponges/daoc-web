package main

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

type grassOut struct {
	Table   string        `json:"table"`
	Atlas   string        `json:"atlas"`
	AtlasPx int           `json:"atlasPx"`
	Map     string        `json:"map"` // per cell: group, then density, one byte each
	Sprites []grassSprite `json:"sprites"`
}

type grassSprite struct {
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

func buildGrass(game string, zoneNum, grid int, outDir string) (*grassOut, error) {
	_, reg, err := zone.RegionOf(game, zoneNum)
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
	out := &grassOut{Table: table, Atlas: "grass/atlas.png", Map: "grass/map.bin"}
	for _, r := range rows {
		if len(r) < 16 {
			continue
		}
		if _, err := strconv.Atoi(strings.TrimSpace(r[0])); err != nil {
			continue // the header rows, and the empty padding at the end
		}
		n := func(i int) float64 { return num(r[i], 0) }
		s := grassSprite{
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

	if err := os.MkdirAll(filepath.Join(outDir, "grass"), 0o755); err != nil {
		return nil, err
	}
	png, err := convertSkyTexture(dir, atlas, filepath.Join(outDir, "grass"))
	if err != nil {
		return nil, err
	}
	if err := os.Rename(filepath.Join(outDir, "grass", png), filepath.Join(outDir, "grass", "atlas.png")); err != nil {
		return nil, err
	}
	out.AtlasPx = 512

	// Where each group grows, and how thickly.
	z, err := zone.Open(game, zoneNum)
	if err != nil {
		return nil, err
	}
	gm, err := z.PCX("grassmap.pcx")
	if err != nil {
		return nil, err
	}
	dm, err := z.PCX("densemap.pcx")
	if err != nil {
		return nil, err
	}
	if gm.W != grid || gm.H != grid || dm.W != grid || dm.H != grid {
		return nil, fmt.Errorf("grass maps %dx%d and %dx%d, heightmap %d", gm.W, gm.H, dm.W, dm.H, grid)
	}
	groups := map[int]bool{}
	for _, s := range out.Sprites {
		groups[s.Group] = true
	}
	cells := make([]byte, grid*grid*2)
	missing := map[int]int{}
	for i := 0; i < grid*grid; i++ {
		v := int(gm.Pal[gm.Pix[i]][0])
		g := v / 10
		if v%10 != 0 || !groups[g] {
			missing[v]++
		}
		cells[i*2] = byte(g)
		cells[i*2+1] = dm.Pal[dm.Pix[i]][0]
	}
	if len(missing) > 0 {
		fmt.Printf("  WARNING: grassmap values with no group in %s: %v\n", table, missing)
	}
	if err := os.WriteFile(filepath.Join(outDir, "grass", "map.bin"), cells, 0o644); err != nil {
		return nil, err
	}
	fmt.Printf("  grass: %s, %d sprites in %d groups, atlas %s\n", table, len(out.Sprites), len(groups), atlas)
	return out, nil
}
