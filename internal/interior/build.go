package interior

import (
	"math"

	"daocweb/internal/scenery"
)

// Built is an interior's layout with its models baked: what is placed, what
// would not convert, and where the geometry lies.
type Built struct {
	Placed  []Placed
	Skipped int               // layout rows naming no model or piece
	Failing map[string]error  // model file -> why it does not convert
	Uses    map[string]int    // model file -> placements
	Surface *Surface          // the converting placements' vertices, layout space
	Bounds  Box               // of the same, layout space
	verts   [][3]float32      // kept for the fit
	models  map[string]*model // baked once each
}

type model struct {
	verts [][3]float32
}

// Build bakes each model the layout places, once, in memory, through the
// same scenery code the converter writes with.
func Build(game string, l *Layout) *Built {
	b := &Built{Failing: map[string]error{}, Uses: map[string]int{}, models: map[string]*model{}}
	placed, skipped := l.Placed()
	b.Placed, b.Skipped = placed, l.Skipped+skipped
	dirs := ModelDirs(game, l)
	inf := float32(math.Inf(1))
	b.Bounds = Box{Min: [3]float32{inf, inf, inf}, Max: [3]float32{-inf, -inf, -inf}}
	for _, p := range b.Placed {
		b.Uses[p.Model]++
		m, ok := b.models[p.Model]
		if !ok {
			var v []scenery.Vertex
			var ix []uint32
			if _, err := scenery.ConvertModel(dirs, p.Model, p.Model, &v, &ix); err != nil {
				b.Failing[p.Model] = err
			} else {
				m = &model{}
				for _, x := range v {
					m.verts = append(m.verts, x.P)
				}
			}
			b.models[p.Model] = m
		}
		if m == nil {
			continue
		}
		for _, v := range m.verts {
			w := p.M.Point(v)
			b.verts = append(b.verts, w)
			for k := 0; k < 3; k++ {
				b.Bounds.Min[k] = min(b.Bounds.Min[k], w[k])
				b.Bounds.Max[k] = max(b.Bounds.Max[k], w[k])
			}
		}
	}
	b.Surface = NewSurface(b.verts)
	return b
}

// Converting is how many placements convert.
func (b *Built) Converting() int {
	n := 0
	for _, p := range b.Placed {
		if b.models[p.Model] != nil {
			n++
		}
	}
	return n
}
