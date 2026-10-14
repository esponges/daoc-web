// Package interior reads the zones that have no heightmap: cities and
// dungeons, assembled from model pieces.
//
// A dungeon's dat<zone>.mpk lists its piece models in dungeon.chunk, one a
// line, and places them in dungeon.place:
//
//	index, x, y, z, angle, axis x, axis y, axis z, ...
//	1, -1679.896729, -128.000000, 0.000000, 1.570796, 0.000000, 0.000000, -1.000000, 0, 0, 4, 0, 0
//
// index counts from 0 into dungeon.chunk; the rotation is an angle about an
// axis, as in the outdoor fixtures. dungeon.prop places props by the same
// index list, with two more numbers and a scale:
//
//	index, x, y, z, angle, axis x, axis y, axis z, a, b, scale
//
// A city's city.csv gives its block count and then its blocks, jord01.nif
// to jord31.nif for Jordheim; each block is modelled where it stands, so
// there is nothing to place.
//
// Both kinds carry ZONEJUMP.csv, the trigger boxes that lead out, and
// lights.csv, the zone's point lights. Those two are in zone coordinates,
// the layout in a space of its own; see fit.go for how the two are tied.
package interior

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"daocweb/internal/zone"
)

// Kind is what an interior is assembled from.
type Kind string

const (
	Dungeon Kind = "dungeon"
	City    Kind = "city"
)

// Placement is one model placed in layout space.
type Placement struct {
	Model string     // file name, from dungeon.chunk or city.csv
	Pos   [3]float32 // layout space
	Angle float32    // radians about Axis
	Axis  [3]float32
	Scale float32
	// A and B are dungeon.prop's two unnamed columns, which may tie a prop
	// to a placed piece; see fit.go.
	A, B int
}

// Entrance is a ZONEJUMP.csv trigger box, in zone coordinates.
type Entrance struct {
	Name     string
	Min, Max [3]float32
	Target   int // the jump's target id
}

// Centre is the middle of the box.
func (e Entrance) Centre() [3]float32 {
	return [3]float32{(e.Min[0] + e.Max[0]) / 2, (e.Min[1] + e.Max[1]) / 2, (e.Min[2] + e.Max[2]) / 2}
}

// Light is a lights.csv row, in zone coordinates.
type Light struct {
	Pos  [3]float32 `json:"pos"`
	Type int        `json:"type"`
}

// Layout is an interior zone as its files describe it.
type Layout struct {
	Num       int
	Kind      Kind
	Models    []string    // dungeon.chunk, or the city's blocks
	Pieces    []Placement // dungeon pieces, or the city's blocks unmoved
	Props     []Placement
	Entrances []Entrance
	Lights    []Light
	// Skipped counts layout rows that name no model, or do not parse.
	Skipped int
	// Frontier is set for a zone under frontiers/.
	Frontier bool
}

// Read reads an interior zone's layout.
func Read(game string, num int) (*Layout, error) {
	z, err := zone.Open(game, num)
	if err != nil {
		return nil, err
	}
	l := &Layout{Num: num, Frontier: z.Frontier}
	switch {
	case has(z, "dungeon.chunk"):
		l.Kind = Dungeon
		if err := l.readDungeon(z); err != nil {
			return nil, err
		}
	case has(z, "city.csv"):
		l.Kind = City
		if err := l.readCity(z); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("zone %d has neither dungeon.chunk nor city.csv", num)
	}
	if raw, err := z.Entry("ZONEJUMP.csv"); err == nil {
		for _, r := range rows(raw) {
			if len(r) < 8 {
				continue
			}
			x1, y1, x2, y2, z1, z2 := f32(r[2]), f32(r[3]), f32(r[4]), f32(r[5]), f32(r[6]), f32(r[7])
			e := Entrance{
				Name: strings.TrimSpace(r[1]),
				Min:  [3]float32{min(x1, x2), min(y1, y2), min(z1, z2)},
				Max:  [3]float32{max(x1, x2), max(y1, y2), max(z1, z2)},
			}
			if len(r) > 8 {
				e.Target = atoi(r[8])
			}
			l.Entrances = append(l.Entrances, e)
		}
	}
	if raw, err := z.Entry("lights.csv"); err == nil {
		for _, r := range rows(raw) {
			if len(r) < 4 {
				continue
			}
			l.Lights = append(l.Lights, Light{Pos: [3]float32{f32(r[0]), f32(r[1]), f32(r[2])}, Type: atoi(r[3])})
		}
	}
	return l, nil
}

func (l *Layout) readDungeon(z *zone.Zone) error {
	raw, err := z.Entry("dungeon.chunk")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(raw), "\r", ""), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			l.Models = append(l.Models, s)
		}
	}
	read := func(name string, cols int, out *[]Placement) error {
		raw, err := z.Entry(name)
		if err != nil {
			return nil // a dungeon without props has no dungeon.prop
		}
		for _, r := range rows(raw) {
			if len(r) < cols {
				l.Skipped++
				continue
			}
			i, err := strconv.Atoi(strings.TrimSpace(r[0]))
			if err != nil || i < 0 || i >= len(l.Models) {
				l.Skipped++
				continue
			}
			p := Placement{
				Model: l.Models[i],
				Pos:   [3]float32{f32(r[1]), f32(r[2]), f32(r[3])},
				Angle: f32(r[4]),
				Axis:  [3]float32{f32(r[5]), f32(r[6]), f32(r[7])},
				Scale: 1,
			}
			if len(r) >= 11 {
				p.A, p.B = atoi(r[8]), atoi(r[9])
			}
			if name == "dungeon.prop" && len(r) >= 11 {
				if s := f32(r[10]); s > 0 {
					p.Scale = s
				}
			}
			*out = append(*out, p)
		}
		return nil
	}
	if err := read("dungeon.place", 8, &l.Pieces); err != nil {
		return err
	}
	return read("dungeon.prop", 8, &l.Props)
}

func (l *Layout) readCity(z *zone.Zone) error {
	raw, err := z.Entry("city.csv")
	if err != nil {
		return err
	}
	for _, r := range rows(raw) {
		if len(r) < 2 || strings.TrimSpace(r[1]) == "" {
			continue // the count on the first line
		}
		name := strings.TrimSpace(r[1])
		if !strings.EqualFold(filepath.Ext(name), ".nif") {
			l.Skipped++
			continue
		}
		l.Models = append(l.Models, name)
		l.Pieces = append(l.Pieces, Placement{Model: name, Axis: [3]float32{0, 0, 1}, Scale: 1})
	}
	return nil
}

// ModelDirs is where an interior's models live, in search order: a city's
// own nifs folder, the dungeon pieces, and the scenery the dungeons
// borrow for their props.
func ModelDirs(game string, l *Layout) []string {
	var dirs []string
	if l.Kind == City {
		dir, _, err := zone.Dir(game, l.Num)
		if err == nil {
			dirs = append(dirs, filepath.Join(dir, "nifs"))
		}
	}
	return append(dirs,
		filepath.Join(game, "zones", "Dnifs"),
		filepath.Join(game, "frontiers", "dnifs"),
		filepath.Join(game, "zones", "Nifs"),
		filepath.Join(game, "frontiers", "NIFS"),
		filepath.Join(game, "Newtowns", "zones", "Nifs"),
	)
}

func has(z *zone.Zone, name string) bool {
	_, err := z.Entry(name)
	return err == nil
}

func rows(raw []byte) [][]string {
	rd := csv.NewReader(bytes.NewReader(raw))
	rd.FieldsPerRecord = -1
	rd.TrimLeadingSpace = true
	rd.LazyQuotes = true
	out, _ := rd.ReadAll()
	return out
}

func f32(s string) float32 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 32)
	return float32(v)
}

func atoi(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}
