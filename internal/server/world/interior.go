package world

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// Interior is a city or dungeon: no heightmap, a floor found in the geometry
// and walls made of it. It is the server's twin of web/interior.js.
type Interior struct {
	col       *Collision
	min, max  [3]float64
	Entrances [][2][3]float64 // each exit's box, min then max
}

type interiorManifest struct {
	Zone   int    `json:"zone"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Bounds struct {
		Min [3]float64 `json:"min"`
		Max [3]float64 `json:"max"`
	} `json:"bounds"`
	Entrances []struct {
		Min [3]float64 `json:"min"`
		Max [3]float64 `json:"max"`
	} `json:"entrances"`
}

// LoadInterior reads a city or dungeon from web/data/zoneN.
func LoadInterior(webData string, zone int) (*Interior, string, error) {
	b, err := os.ReadFile(filepath.Join(zoneDir(webData, zone), "zone.json"))
	if err != nil {
		return nil, "", fmt.Errorf("zone %d is not converted: %w", zone, err)
	}
	var m interiorManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, "", err
	}
	if m.Kind == "" {
		return nil, m.Name, fmt.Errorf("zone %d is outdoors", zone)
	}
	col, err := LoadInteriorGeometry(webData, zone, [2][3]float64{m.Bounds.Min, m.Bounds.Max})
	if err != nil {
		return nil, "", err
	}
	in := &Interior{col: col, min: m.Bounds.Min, max: m.Bounds.Max}
	for _, e := range m.Entrances {
		in.Entrances = append(in.Entrances, [2][3]float64{e.Min, e.Max})
	}
	return in, m.Name, nil
}

// GroundAt is the floor under (x, y) a step above z or lower; with no floor
// there it stays at z, as the viewer does.
func (in *Interior) GroundAt(x, y, z float64) float64 {
	if f, ok := in.col.FloorBelow(x, y, z); ok {
		return f
	}
	return z
}

func (*Interior) IsWet(x, y float64) bool { return false }

func (in *Interior) Clamp(x, y float64) (float64, float64) {
	return math.Max(in.min[0], math.Min(in.max[0], x)), math.Max(in.min[1], math.Min(in.max[1], y))
}

func (in *Interior) Extent() float64 { return math.Max(in.max[0], in.max[1]) }

// MaxDrop is the furthest a walker steps down in one move. A ledge or a
// void beyond it is as solid as a wall: nothing here falls.
const MaxDrop = 110.0

// walkable reports whether there is floor to stand on at (x, y), no more than
// MaxDrop below z.
func (in *Interior) walkable(x, y, z float64) bool {
	f, ok := in.col.FloorBelow(x, y, z)
	return ok && z-f <= MaxDrop
}

// Slide keeps a walker out of the walls, and off ledges and voids: where the
// move would leave the floor it slides along the edge or stops.
func (in *Interior) Slide(x0, y0, x1, y1, z float64) (float64, float64) {
	nx, ny := in.col.Slide(x0, y0, x1, y1, z)
	if !in.walkable(x0, y0, z) {
		return nx, ny // already off the floor: let them find their way back
	}
	if in.walkable(nx, ny, z) {
		return nx, ny
	}
	if in.walkable(nx, y0, z) {
		return nx, y0
	}
	if in.walkable(x0, ny, z) {
		return x0, ny
	}
	return x0, y0
}

// HasFloor reports whether the interior's geometry gave any floor at all.
func (in *Interior) HasFloor() bool { return in.col != nil && in.col.hasFloor }

// SpawnAt finds a place to stand at the first entrance: on the floor under
// the box's middle, or the nearest floor around it, searched in rings.
func (in *Interior) SpawnAt() (x, y, z float64, ok bool) {
	if len(in.Entrances) == 0 {
		return 0, 0, 0, false
	}
	box := in.Entrances[0]
	cx, cy := (box[0][0]+box[1][0])/2, (box[0][1]+box[1][1])/2
	from := box[0][2] + 100
	for r := 0.0; r <= 600; r += 40 {
		steps := 1
		if r > 0 {
			steps = int(math.Ceil(2 * math.Pi * r / 40))
		}
		for k := 0; k < steps; k++ {
			a := float64(k) / float64(steps) * 2 * math.Pi
			px, py := cx+math.Cos(a)*r, cy+math.Sin(a)*r
			if f, found := in.col.FloorBelow(px, py, from); found {
				return px, py, f, true
			}
		}
	}
	return cx, cy, box[0][2], false
}
