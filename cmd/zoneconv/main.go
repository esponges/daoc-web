// Command zoneconv extracts one DAoC zone and writes browser-ready data:
// a uint16 heightmap, a stitched terrain colour atlas, and a JSON manifest.
//
//	zoneconv -zone 100 -out web/data
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"daocweb/internal/ground"
	"daocweb/internal/zone"
)

// A DAoC zone spans 65536 world units and is sampled by a 256x256 height
// raster, so one raster cell covers 256 units.
const zoneUnits = zone.Units

type manifest struct {
	Zone         int        `json:"zone"`
	Name         string     `json:"name"`
	Grid         int        `json:"grid"`      // height samples per side
	CellUnits    float64    `json:"cellUnits"` // world units between samples
	ZoneUnits    int        `json:"zoneUnits"`
	ScaleFactor  int        `json:"scaleFactor"`
	OffsetFactor int        `json:"offsetFactor"`
	MinHeight    int        `json:"minHeight"`
	MaxHeight    int        `json:"maxHeight"`
	Heights      string     `json:"heights"` // uint16 LE, grid*grid
	Atlas        string     `json:"atlas"`
	AtlasPx      int        `json:"atlasPx"`
	Orientation  string     `json:"orientation"` // how LOD tiles were laid out
	Waters       []waterOut `json:"waters"`
	Start        [4]int     `json:"start"`
	Fog          fogOut     `json:"fog"`
	Splat        *splatOut  `json:"splat,omitempty"` // ground detail layers; see splat.go
	Sky          *skyOut    `json:"sky,omitempty"`   // the region's skydome; see sky.go
	Grass        *grassOut  `json:"grass,omitempty"` // the region's grass sprites; see grass.go
}

type waterOut struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Height int      `json:"height"`
	Color  int      `json:"color"`
	Left   [][2]int `json:"left"`  // shoreline chain, heightmap cells
	Right  [][2]int `json:"right"` // opposite shore, index-aligned with Left
}

type fogOut struct {
	R      int `json:"r"`
	G      int `json:"g"`
	B      int `json:"b"`
	Clip   int `json:"clip"`
	Amount int `json:"amount"`
}

func main() {
	game := flag.String("game", defaultGamePath, "DAoC install directory")
	zoneNum := flag.Int("zone", 100, "zone number (100 = Vale of Mularn)")
	zoneName := flag.String("name", "", "display name for the manifest")
	out := flag.String("out", "web/data", "output directory")
	flag.Parse()

	if err := run(*game, *zoneNum, *zoneName, *out); err != nil {
		log.Fatalf("zoneconv: %v", err)
	}
}

const defaultGamePath = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

func run(game string, zoneNum int, zoneName, outRoot string) error {
	z, err := zone.Open(game, zoneNum)
	if err != nil {
		return err
	}
	outDir := filepath.Join(outRoot, fmt.Sprintf("zone%03d", zoneNum))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	fmt.Printf("zone %d  <- %s\n", zoneNum, z.Dir)

	// --- zone rasters and metadata -----------------------------------------
	fmt.Printf("  dat%03d.mpk: %d entries\n", zoneNum, len(z.Dat.Entries))

	sec, err := z.Sector()
	if err != nil {
		return err
	}
	sf, of := sec.ScaleFactor(), sec.OffsetFactor()
	sx, sy := sec.SectorSize()
	fmt.Printf("  scalefactor=%d offsetfactor=%d sectors=%dx%d\n", sf, of, sx, sy)

	terr, err := z.Terrain(sec)
	if err != nil {
		return err
	}
	grid, heights, minH, maxH := terr.Grid, terr.Heights, terr.Min, terr.Max
	fmt.Printf("  heightmap: %dx%d\n", grid, grid)
	fmt.Printf("  height range: %d .. %d world units\n", minH, maxH)

	// --- terrain colour atlas from the LOD tiles ---------------------------
	lod, err := ground.ReadLOD(z, sx, sy)
	if err != nil {
		return err
	}
	atlasPx := lod.TileSize * sx
	fmt.Printf("  lod%03d.mpk: %d tiles of %dpx -> %dx%d atlas\n", zoneNum, len(lod.Tiles), lod.TileSize, atlasPx, atlasPx)

	// The tile index order and axis directions are not documented, so try
	// all eight layouts. Where the zone has the game's own pre-blended tiles
	// of each sector, those say which is right: the atlas is a picture of
	// the same sectors (see ground.CheckLayout). Without them, keep whichever
	// layout best correlates brightness with height (snow caps are bright;
	// low ground and water are dark), which is weak in flat or watery zones.
	var ref *ground.Reference
	if sx == sy {
		if ref, err = ground.ReadReference(z, sx); err != nil {
			fmt.Printf("  no reference tiles to check the layout against: %v\n", err)
			ref = nil
		}
	}
	var chosen ground.Layout
	if ref != nil {
		c, err := ground.CheckLayout(lod, ref)
		if err != nil {
			return err
		}
		for i, lay := range ground.Layouts() {
			fmt.Printf("    layout %s  tiles r = %+.4f\n", lay, c.Agreements[i])
		}
		chosen = c.Chosen
		fmt.Printf("  chosen layout: %s (tiles r = %+.4f; best other %+.4f)\n", chosen, c.Agree, c.Other)
		if c.Agree < 0.6 {
			fmt.Printf("  WARNING: weak agreement with the reference tiles, texture alignment is unverified\n")
		}
	} else {
		best := -2.0
		for _, lay := range ground.Layouts() {
			r := correlate(lod.Stitch(lay), heights, grid)
			fmt.Printf("    layout %s  brightness/height r = %+.4f\n", lay, r)
			if r > best {
				chosen, best = lay, r
			}
		}
		fmt.Printf("  chosen layout: %s (brightness/height r = %+.4f)\n", chosen, best)
		if best < 0.2 {
			fmt.Printf("  WARNING: weak correlation, texture alignment is unverified\n")
		}
	}
	atlas := lod.Stitch(chosen)

	// --- write outputs -----------------------------------------------------
	hBuf := make([]byte, len(heights)*2)
	for i, h := range heights {
		binary.LittleEndian.PutUint16(hBuf[i*2:], h)
	}
	if err := os.WriteFile(filepath.Join(outDir, "heights.u16"), hBuf, 0o644); err != nil {
		return err
	}
	if err := writePNG(filepath.Join(outDir, "atlas.png"), atlas); err != nil {
		return err
	}

	if zoneName == "" {
		zoneName = fmt.Sprintf("Zone %d", zoneNum)
	}
	man := manifest{
		Zone:         zoneNum,
		Name:         zoneName,
		Grid:         grid,
		CellUnits:    float64(zoneUnits) / float64(grid),
		ZoneUnits:    zoneUnits,
		ScaleFactor:  sf,
		OffsetFactor: of,
		MinHeight:    minH,
		MaxHeight:    maxH,
		Heights:      "heights.u16",
		Atlas:        "atlas.png",
		AtlasPx:      atlasPx,
		Orientation:  chosen.String(),
		Start: [4]int{
			sec.Int("start", "x", 0), sec.Int("start", "y", 0),
			sec.Int("start", "z", 0), sec.Int("start", "a", 0),
		},
		Fog: fogOut{
			R:      sec.Int("fog", "red", 128),
			G:      sec.Int("fog", "green", 128),
			B:      sec.Int("fog", "blue", 128),
			Clip:   sec.Int("fog", "clip", 16000),
			Amount: sec.Int("fog", "amount", 50),
		},
	}
	for _, w := range sec.Waters() {
		man.Waters = append(man.Waters, waterOut{
			Name: w.Name, Type: w.Type, Height: w.Height, Color: w.Color,
			Left: w.Left, Right: w.Right,
		})
		fmt.Printf("  water: %s (%s) at height %d, %d shoreline pairs\n", w.Name, w.Type, w.Height, len(w.Left))
	}

	// Ground detail is an addition, not a requirement: a zone without the
	// layer tables still draws from the atlas alone.
	if sx == sy {
		if sp, err := buildSplat(game, zoneNum, sx, outDir); err != nil {
			fmt.Printf("  ground detail skipped: %v\n", err)
		} else {
			man.Splat = sp
		}
	}
	if sk, err := buildSky(game, zoneNum, outDir); err != nil {
		fmt.Printf("  sky skipped: %v\n", err)
	} else {
		man.Sky = sk
	}
	if gr, err := buildGrass(game, zoneNum, grid, outDir); err != nil {
		fmt.Printf("  grass skipped: %v\n", err)
	} else {
		man.Grass = gr
	}

	mj, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "zone.json"), append(mj, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("  wrote %s (heights.u16, atlas.png, zone.json)\n", outDir)
	return nil
}

// correlate scores a candidate atlas by Pearson correlation between per-cell
// mean brightness and terrain height.
func correlate(img *image.NRGBA, heights []uint16, grid int) float64 {
	step := img.Bounds().Dx() / grid
	if step < 1 {
		step = 1
	}
	var n int
	var sx, sy, sxx, syy, sxy float64
	for gy := 0; gy < grid; gy++ {
		for gx := 0; gx < grid; gx++ {
			var sum, cnt float64
			for py := gy * step; py < (gy+1)*step; py++ {
				for px := gx * step; px < (gx+1)*step; px++ {
					o := img.PixOffset(px, py)
					sum += 0.299*float64(img.Pix[o]) + 0.587*float64(img.Pix[o+1]) + 0.114*float64(img.Pix[o+2])
					cnt++
				}
			}
			if cnt == 0 {
				continue
			}
			x := sum / cnt
			y := float64(heights[gy*grid+gx])
			sx += x
			sy += y
			sxx += x * x
			syy += y * y
			sxy += x * y
			n++
		}
	}
	if n == 0 {
		return 0
	}
	fn := float64(n)
	den := math.Sqrt((sxx - sx*sx/fn) * (syy - sy*sy/fn))
	if den == 0 {
		return 0
	}
	return (sxy - sx*sy/fn) / den
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
