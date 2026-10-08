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
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"daocweb/internal/dds"
	"daocweb/internal/mpak"
	"daocweb/internal/pcx"
	"daocweb/internal/sector"
)

// A DAoC zone spans 65536 world units and is sampled by a 256x256 height
// raster, so one raster cell covers 256 units.
const zoneUnits = 65536

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
	zoneDir := filepath.Join(game, "zones", fmt.Sprintf("zone%03d", zoneNum))
	if _, err := os.Stat(zoneDir); err != nil {
		return fmt.Errorf("zone directory %s: %w", zoneDir, err)
	}
	outDir := filepath.Join(outRoot, fmt.Sprintf("zone%03d", zoneNum))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	fmt.Printf("zone %d  <- %s\n", zoneNum, zoneDir)

	// --- zone rasters and metadata -----------------------------------------
	datArc, err := mpak.Open(filepath.Join(zoneDir, fmt.Sprintf("dat%03d.mpk", zoneNum)))
	if err != nil {
		return err
	}
	fmt.Printf("  dat%03d.mpk: %d entries\n", zoneNum, len(datArc.Entries))

	secRaw, err := datArc.Read("SECTOR.DAT")
	if err != nil {
		return err
	}
	sec, err := sector.Parse(secRaw)
	if err != nil {
		return err
	}
	sf, of := sec.ScaleFactor(), sec.OffsetFactor()
	sx, sy := sec.SectorSize()
	fmt.Printf("  scalefactor=%d offsetfactor=%d sectors=%dx%d\n", sf, of, sx, sy)

	ter, err := decodePCX(datArc, "terrain.pcx")
	if err != nil {
		return err
	}
	offs, err := decodePCX(datArc, "offset.pcx")
	if err != nil {
		return err
	}
	if ter.W != offs.W || ter.H != offs.H {
		return fmt.Errorf("terrain %dx%d and offset %dx%d disagree", ter.W, ter.H, offs.W, offs.H)
	}
	grid := ter.W
	fmt.Printf("  heightmap: %dx%d\n", ter.W, ter.H)

	heights := make([]uint16, grid*grid)
	minH, maxH := math.MaxInt32, 0
	for y := 0; y < grid; y++ {
		for x := 0; x < grid; x++ {
			h := int(ter.At(x, y))*sf + int(offs.At(x, y))*of
			if h > 0xFFFF {
				return fmt.Errorf("height %d at (%d,%d) overflows uint16", h, x, y)
			}
			heights[y*grid+x] = uint16(h)
			if h < minH {
				minH = h
			}
			if h > maxH {
				maxH = h
			}
		}
	}
	fmt.Printf("  height range: %d .. %d world units\n", minH, maxH)

	// --- terrain colour atlas from the LOD tiles ---------------------------
	lodArc, err := mpak.Open(filepath.Join(zoneDir, fmt.Sprintf("lod%03d.mpk", zoneNum)))
	if err != nil {
		return err
	}
	tiles := map[[2]int]*image.NRGBA{}
	tileSize := 0
	for a := 0; a < sx; a++ {
		for b := 0; b < sy; b++ {
			name := fmt.Sprintf("lod%02d-%02d.dds", a, b)
			raw, err := lodArc.Read(name)
			if err != nil {
				return err
			}
			im, err := dds.Decode(raw)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			if tileSize == 0 {
				tileSize = im.Bounds().Dx()
			} else if im.Bounds().Dx() != tileSize {
				return fmt.Errorf("%s is %dpx, expected %dpx", name, im.Bounds().Dx(), tileSize)
			}
			tiles[[2]int{a, b}] = im
		}
	}
	atlasPx := tileSize * sx
	fmt.Printf("  lod%03d.mpk: %d tiles of %dpx -> %dx%d atlas\n", zoneNum, len(tiles), tileSize, atlasPx, atlasPx)

	// The tile index order and axis directions are not documented, so try all
	// eight layouts and keep whichever best correlates brightness with height
	// (snow caps are bright; low ground and water are dark).
	type candidate struct {
		name string
		corr float64
		img  *image.NRGBA
	}
	var best *candidate
	for _, swap := range []bool{false, true} {
		for _, fx := range []bool{false, true} {
			for _, fy := range []bool{false, true} {
				img := stitch(tiles, sx, sy, tileSize, swap, fx, fy)
				c := &candidate{
					name: fmt.Sprintf("swap=%-5v flipX=%-5v flipY=%-5v", swap, fx, fy),
					corr: correlate(img, heights, grid),
					img:  img,
				}
				fmt.Printf("    layout %s  brightness/height r = %+.4f\n", c.name, c.corr)
				if best == nil || c.corr > best.corr {
					best = c
				}
			}
		}
	}
	fmt.Printf("  chosen layout: %s (r = %+.4f)\n", best.name, best.corr)
	if best.corr < 0.2 {
		fmt.Printf("  WARNING: weak correlation, texture alignment is unverified\n")
	}

	// --- write outputs -----------------------------------------------------
	hBuf := make([]byte, len(heights)*2)
	for i, h := range heights {
		binary.LittleEndian.PutUint16(hBuf[i*2:], h)
	}
	if err := os.WriteFile(filepath.Join(outDir, "heights.u16"), hBuf, 0o644); err != nil {
		return err
	}
	if err := writePNG(filepath.Join(outDir, "atlas.png"), best.img); err != nil {
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
		Orientation:  best.name,
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

func decodePCX(a *mpak.Archive, name string) (*pcx.Image, error) {
	raw, err := a.Read(name)
	if err != nil {
		return nil, err
	}
	im, err := pcx.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return im, nil
}

// stitch assembles the LOD tiles into one atlas under a candidate layout.
func stitch(tiles map[[2]int]*image.NRGBA, sx, sy, tileSize int, swap, flipX, flipY bool) *image.NRGBA {
	atlasPx := tileSize * sx
	out := image.NewNRGBA(image.Rect(0, 0, atlasPx, atlasPx))
	for key, im := range tiles {
		cx, cy := key[0], key[1]
		if swap {
			cx, cy = cy, cx
		}
		if flipX {
			cx = sx - 1 - cx
		}
		if flipY {
			cy = sy - 1 - cy
		}
		r := image.Rect(cx*tileSize, cy*tileSize, (cx+1)*tileSize, (cy+1)*tileSize)
		draw.Draw(out, r, im, im.Bounds().Min, draw.Src)
	}
	return out
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
