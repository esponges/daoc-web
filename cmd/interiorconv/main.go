// Command interiorconv converts a city or dungeon zone -- one with no
// heightmap, assembled from model pieces -- into the scenery format the
// viewer draws, with a zone.json saying what kind it is, how its layout
// sits in zone coordinates, and where its exits and lights are.
//
//	go run ./cmd/interiorconv -zone 125      # Spindelhalla
//
// Reading the layout and tying it to zone coordinates is internal/interior;
// baking each model is internal/scenery. Every Midgard interior's layout is
// mirrored in x against the zone, so each model is baked mirrored once and
// the instances stay plain turns.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"daocweb/internal/interior"
	"daocweb/internal/scenery"
	"daocweb/internal/zone"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

// manifest is an interior's zone.json.
type manifest struct {
	Zone      int              `json:"zone"`
	Name      string           `json:"name"`
	Kind      interior.Kind    `json:"kind"`
	Transform interior.Fit     `json:"transform"`
	Bounds    interior.Box     `json:"bounds"` // zone coordinates
	Entrances []entranceOut    `json:"entrances"`
	Lights    []interior.Light `json:"lights"`
	Fog       *fogOut          `json:"fog,omitempty"`
	Props     string           `json:"props"`
	Counts    map[string]int   `json:"counts"`
}

type entranceOut struct {
	Name string     `json:"name"`
	Min  [3]float32 `json:"min"`
	Max  [3]float32 `json:"max"`
}

type fogOut struct {
	R    int `json:"r"`
	G    int `json:"g"`
	B    int `json:"b"`
	Clip int `json:"clip"`
}

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	num := flag.Int("zone", 125, "zone number")
	name := flag.String("name", "", "display name for the manifest")
	out := flag.String("out", filepath.Join("web", "data"), "output directory")
	flag.Parse()
	if err := run(*game, *num, *name, *out); err != nil {
		fmt.Fprintln(os.Stderr, "interiorconv:", err)
		os.Exit(1)
	}
}

func run(game string, num int, name, out string) error {
	l, err := interior.Read(game, num)
	if err != nil {
		return err
	}
	built := interior.Build(game, l)
	fit := built.Fit(l)
	fmt.Printf("zone %d %s: %d placed (%d skipped); transform origin %.0f, %.0f, %.0f puts %d of %d exits and lights on the geometry, %d on it or inside a piece\n",
		num, l.Kind, len(built.Placed), built.Skipped, fit.Origin[0], fit.Origin[1], fit.Origin[2], fit.OnSurface, fit.Points,
		built.Covered(fit.Transform, l.Points()))

	dirs := interior.ModelDirs(game, l)
	var (
		verts   []scenery.Vertex
		indices []uint32
		models  []scenery.Model
		skipped []string
		slot    = map[string]int{}
	)
	for _, p := range built.Placed {
		if _, ok := slot[p.Model]; ok {
			continue
		}
		v0 := len(verts)
		mo, err := scenery.ConvertModel(dirs, p.Model, p.Model, &verts, &indices)
		if err != nil {
			slot[p.Model] = -1
			skipped = append(skipped, fmt.Sprintf("%s: %v (%d placements)", p.Model, err, built.Uses[p.Model]))
			continue
		}
		mirror(mo, verts, v0, indices)
		slot[p.Model] = len(models)
		models = append(models, *mo)
	}

	var inst []scenery.Instance
	for _, p := range built.Placed {
		s := slot[p.Model]
		if s < 0 {
			continue
		}
		inst = append(inst, instance(s, p.M, fit.Transform))
	}
	dir := filepath.Join(out, fmt.Sprintf("zone%03d", num))
	if err := scenery.WriteProps(filepath.Join(dir, "props"), num, dirs, models, verts, indices, inst, skipped); err != nil {
		return err
	}

	man := manifest{
		Zone: num, Name: name, Kind: l.Kind, Transform: fit,
		Bounds: zoneBox(built.Bounds, fit.Transform),
		Lights: l.Lights, Props: "props/props.json",
		Counts: map[string]int{"placed": len(built.Placed), "converting": len(inst), "skipped": built.Skipped},
	}
	if man.Name == "" {
		man.Name = fmt.Sprintf("Zone %d", num)
	}
	if man.Lights == nil {
		man.Lights = []interior.Light{}
	}
	for _, e := range l.Entrances {
		man.Entrances = append(man.Entrances, entranceOut{e.Name, e.Min, e.Max})
	}
	if z, err := zone.Open(game, num); err == nil {
		if sec, err := z.Sector(); err == nil {
			if r := sec.Int("fog", "red", -1); r >= 0 {
				man.Fog = &fogOut{R: r, G: sec.Int("fog", "green", 0), B: sec.Int("fog", "blue", 0), Clip: sec.Int("fog", "clip", 0)}
			}
		}
	}
	js, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "zone.json"), append(js, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("%d models, %d of %d placements written, %d entrances, %d lights -> %s\n",
		len(models), len(inst), len(built.Placed), len(man.Entrances), len(man.Lights), dir)
	for _, s := range skipped {
		fmt.Printf("  skipped %s\n", s)
	}
	return nil
}

// mirror turns a freshly baked model into its mirror image in x: the
// interiors' layouts are mirrored against the zone, and mirroring the
// models once keeps every instance a plain turn. Positions and normals
// change sign in x and triangles wind the other way so their fronts still
// face out. A billboard's normal holds its corner offset on the card, not a
// direction, and stays.
func mirror(mo *scenery.Model, verts []scenery.Vertex, v0 int, indices []uint32) {
	billboard := map[int]bool{}
	for _, g := range mo.Groups {
		for i := g.First; i+2 < g.First+g.Count; i += 3 {
			indices[i+1], indices[i+2] = indices[i+2], indices[i+1]
		}
		if g.Billboard {
			for _, ix := range indices[g.First : g.First+g.Count] {
				billboard[int(ix)] = true
			}
		}
	}
	for i := v0; i < len(verts); i++ {
		verts[i].P[0] = -verts[i].P[0]
		if !billboard[i] {
			verts[i].N[0] = -verts[i].N[0]
		}
	}
	mo.Min[0], mo.Max[0] = -mo.Max[0], -mo.Min[0]
	for i := range mo.Particles {
		p := &mo.Particles[i]
		p.Origin[0] = -p.Origin[0]
		// L A L, with L the mirror: the emitter's frame mirrored too.
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if (r == 0) != (c == 0) {
					p.Axes[r*3+c] = -p.Axes[r*3+c]
				}
			}
		}
		if p.Gravity != nil {
			g := *p.Gravity
			g[0] = -g[0]
			p.Gravity = &g
		}
	}
}

// instance places a mirrored model where the layout matrix m puts the
// original, in zone coordinates. With L the mirror in x, the zone sees
// L m v; the mirrored model holds L v, so the instance's turn is L R L --
// a proper rotation -- and its position the transform of m's.
func instance(model int, m interior.Mat, t interior.Transform) scenery.Instance {
	scale := float32(math.Sqrt(float64(m[0]*m[0] + m[4]*m[4] + m[8]*m[8])))
	if scale == 0 {
		scale = 1
	}
	var r [9]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			v := float64(m[i*4+j] / scale)
			if (i == 0) != (j == 0) {
				v = -v
			}
			r[i*3+j] = v
		}
	}
	angle, axis := axisAngle(r)
	in := scenery.Instance{
		Model: model,
		Pos:   t.Apply([3]float32{m[3], m[7], m[11]}),
		Scale: scale,
	}
	if math.Abs(axis[2]) > 1-1e-3 {
		in.Yaw = float32(angle * math.Copysign(1, axis[2]))
	} else {
		a := [3]float32{float32(axis[0]), float32(axis[1]), float32(axis[2])}
		in.Yaw, in.Axis = float32(angle), &a
	}
	return in
}

// axisAngle reads a rotation matrix as an angle about a unit axis.
func axisAngle(r [9]float64) (float64, [3]float64) {
	c := (r[0] + r[4] + r[8] - 1) / 2
	angle := math.Acos(math.Max(-1, math.Min(1, c)))
	if angle < 1e-6 {
		return 0, [3]float64{0, 0, 1}
	}
	if math.Pi-angle < 1e-4 {
		// Half turn: the axis is the column of r + I with the largest norm.
		var best [3]float64
		bn := -1.0
		for j := 0; j < 3; j++ {
			col := [3]float64{r[j], r[3+j], r[6+j]}
			col[j]++
			n := math.Sqrt(col[0]*col[0] + col[1]*col[1] + col[2]*col[2])
			if n > bn {
				bn, best = n, [3]float64{col[0] / n, col[1] / n, col[2] / n}
			}
		}
		return angle, best
	}
	s := 2 * math.Sin(angle)
	return angle, [3]float64{(r[7] - r[5]) / s, (r[2] - r[6]) / s, (r[3] - r[1]) / s}
}

// zoneBox is a layout box's bounds in zone coordinates.
func zoneBox(b interior.Box, t interior.Transform) interior.Box {
	inf := float32(math.Inf(1))
	out := interior.Box{Min: [3]float32{inf, inf, inf}, Max: [3]float32{-inf, -inf, -inf}}
	for c := 0; c < 8; c++ {
		v := b.Min
		if c&1 != 0 {
			v[0] = b.Max[0]
		}
		if c&2 != 0 {
			v[1] = b.Max[1]
		}
		if c&4 != 0 {
			v[2] = b.Max[2]
		}
		w := t.Apply(v)
		for k := 0; k < 3; k++ {
			out.Min[k] = min(out.Min[k], w[k])
			out.Max[k] = max(out.Max[k], w[k])
		}
	}
	return out
}
