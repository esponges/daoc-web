// Package world is the running game: zones, the players and NPCs in them,
// and the rules that move and fight them. It reads the converted zones in
// web/data and the imported world store, and talks to clients only through
// the protocol package.
package world

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// Ground answers where the floor is. Outdoors it is the heightmap; inside a
// city or dungeon it is found in the geometry.
type Ground interface {
	// GroundAt is the height of the floor at (x, y), seeking from height z.
	GroundAt(x, y, z float64) float64
	// IsWet reports whether (x, y) is under a body of water.
	IsWet(x, y float64) bool
	// Clamp keeps a point inside the zone.
	Clamp(x, y float64) (float64, float64)
	// Extent is the zone's size, for sanity checks.
	Extent() float64
}

// Terrain is an outdoor zone's heightmap and lakes, read from its converted
// files. It matches the viewer's own groundAt, isWet and clamp exactly, so a
// predicting client and the server agree on where the ground is.
type Terrain struct {
	Grid    int
	Cell    float64
	extent  float64
	Heights []uint16
	wet     []wetBox
}

type wetBox struct{ x0, x1, y0, y1, h float64 }

type zoneManifest struct {
	Zone     int    `json:"zone"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Grid     int    `json:"grid"`
	CellUnit int    `json:"cellUnits"`
	Heights  string `json:"heights"`
	Waters   []struct {
		Height float64      `json:"height"`
		Left   [][2]float64 `json:"left"`
		Right  [][2]float64 `json:"right"`
	} `json:"waters"`
}

// LoadTerrain reads an outdoor zone from web/data/zoneN.
func LoadTerrain(webData string, zone int) (*Terrain, string, error) {
	dir := zoneDir(webData, zone)
	b, err := os.ReadFile(filepath.Join(dir, "zone.json"))
	if err != nil {
		return nil, "", fmt.Errorf("zone %d is not converted: %w", zone, err)
	}
	var m zoneManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, "", err
	}
	if m.Kind != "" {
		return nil, m.Name, fmt.Errorf("zone %d is a %s, not an outdoor zone", zone, m.Kind)
	}
	hb, err := os.ReadFile(filepath.Join(dir, m.Heights))
	if err != nil {
		return nil, "", err
	}
	if len(hb) != m.Grid*m.Grid*2 {
		return nil, "", fmt.Errorf("zone %d: heights hold %d bytes for a %d grid", zone, len(hb), m.Grid)
	}
	t := &Terrain{Grid: m.Grid, Cell: float64(m.CellUnit), Heights: make([]uint16, m.Grid*m.Grid)}
	for i := range t.Heights {
		t.Heights[i] = binary.LittleEndian.Uint16(hb[i*2:])
	}
	t.extent = float64(m.Grid-1) * t.Cell
	for _, w := range m.Waters {
		pts := append(append([][2]float64{}, w.Left...), w.Right...)
		if len(pts) == 0 {
			continue
		}
		bx := wetBox{h: w.Height, x0: math.Inf(1), y0: math.Inf(1), x1: math.Inf(-1), y1: math.Inf(-1)}
		for _, p := range pts {
			bx.x0, bx.x1 = math.Min(bx.x0, p[0]*t.Cell), math.Max(bx.x1, p[0]*t.Cell)
			bx.y0, bx.y1 = math.Min(bx.y0, p[1]*t.Cell), math.Max(bx.y1, p[1]*t.Cell)
		}
		t.wet = append(t.wet, bx)
	}
	return t, m.Name, nil
}

func (t *Terrain) h(x, y int) float64 {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > t.Grid-1 {
		x = t.Grid - 1
	}
	if y > t.Grid-1 {
		y = t.Grid - 1
	}
	return float64(t.Heights[y*t.Grid+x])
}

// GroundAt interpolates across the same two triangles per cell the viewer
// draws (split along the (x+1,y)-(x,y+1) diagonal).
func (t *Terrain) GroundAt(wx, wy, _ float64) float64 {
	fx, fy := wx/t.Cell, wy/t.Cell
	x0, y0 := math.Floor(fx), math.Floor(fy)
	tx, ty := fx-x0, fy-y0
	ix, iy := int(x0), int(y0)
	h00, h10, h01, h11 := t.h(ix, iy), t.h(ix+1, iy), t.h(ix, iy+1), t.h(ix+1, iy+1)
	if tx+ty <= 1 {
		return h00 + (h10-h00)*tx + (h01-h00)*ty
	}
	return h11 + (h01-h11)*(1-tx) + (h10-h11)*(1-ty)
}

// IsWet is inside a lake's box and below its surface.
func (t *Terrain) IsWet(x, y float64) bool {
	for _, b := range t.wet {
		if x >= b.x0 && x <= b.x1 && y >= b.y0 && y <= b.y1 && t.GroundAt(x, y, 0) < b.h {
			return true
		}
	}
	return false
}

func (t *Terrain) Clamp(x, y float64) (float64, float64) {
	return math.Max(0, math.Min(t.extent, x)), math.Max(0, math.Min(t.extent, y))
}

func (t *Terrain) Extent() float64 { return t.extent }
