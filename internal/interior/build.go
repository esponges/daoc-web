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
	Surface *Surface          // the converting placements' vertices, layout space; what the fit is made against
	Dense   *Surface          // the same with points across each triangle, for how much a fit covers
	Bounds  Box               // of the same, layout space
	verts   [][3]float32      // kept for the fit
	models  map[string]*model // baked once each
	pieces  []pieceBox        // each placed piece's box, for what lies inside
}

type model struct {
	verts [][3]float32
	tris  [][3]uint32 // vertex indices, three to a triangle
	box   Box         // the model's own bounds
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
	var dense [][3]float32 // the corners, and points across each triangle
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
				infv := float32(math.Inf(1))
				m.box = Box{Min: [3]float32{infv, infv, infv}, Max: [3]float32{-infv, -infv, -infv}}
				for _, x := range v {
					m.verts = append(m.verts, x.P)
					for k := 0; k < 3; k++ {
						m.box.Min[k] = min(m.box.Min[k], x.P[k])
						m.box.Max[k] = max(m.box.Max[k], x.P[k])
					}
				}
				for i := 0; i+2 < len(ix); i += 3 {
					m.tris = append(m.tris, [3]uint32{ix[i], ix[i+1], ix[i+2]})
				}
			}
			b.models[p.Model] = m
		}
		if m == nil {
			continue
		}
		if !p.Prop {
			b.pieces = append(b.pieces, newPieceBox(p.M, m.box))
		}
		for _, v := range m.verts {
			w := p.M.Point(v)
			b.verts = append(b.verts, w)
			for k := 0; k < 3; k++ {
				b.Bounds.Min[k] = min(b.Bounds.Min[k], w[k])
				b.Bounds.Max[k] = max(b.Bounds.Max[k], w[k])
			}
		}
		// A floor or wall is often one large triangle whose corners lie far
		// from its middle; the surface needs points across it, not only at
		// its ends.
		for _, t := range m.tris {
			if int(t[0]) < len(m.verts) && int(t[1]) < len(m.verts) && int(t[2]) < len(m.verts) {
				dense = fill(dense, p.M.Point(m.verts[t[0]]), p.M.Point(m.verts[t[1]]), p.M.Point(m.verts[t[2]]))
			}
		}
	}
	dense = append(dense, b.verts...)
	b.Surface = NewSurface(b.verts)
	b.Dense = NewSurface(dense)
	return b
}

// fillStep is the longest edge a triangle may keep before its middle is
// sampled: well inside a surface cell, so no cell a triangle crosses is missed.
const fillStep = SurfaceCell / 2

// fill appends the points that spread a triangle across the surface map: its
// longest edge is halved until every edge is short.
func fill(out [][3]float32, a, b, c [3]float32) [][3]float32 {
	d := [3]float32{dist(a, b), dist(b, c), dist(c, a)}
	longest := 0
	for i := 1; i < 3; i++ {
		if d[i] > d[longest] {
			longest = i
		}
	}
	if d[longest] <= fillStep {
		return out
	}
	p, q, r := a, b, c // split edge p-q, the third corner r
	switch longest {
	case 1:
		p, q, r = b, c, a
	case 2:
		p, q, r = c, a, b
	}
	m := [3]float32{(p[0] + q[0]) / 2, (p[1] + q[1]) / 2, (p[2] + q[2]) / 2}
	out = append(out, m)
	out = fill(out, p, m, r)
	return fill(out, m, q, r)
}

func dist(a, b [3]float32) float32 {
	return float32(math.Sqrt(float64((a[0]-b[0])*(a[0]-b[0]) + (a[1]-b[1])*(a[1]-b[1]) + (a[2]-b[2])*(a[2]-b[2]))))
}

// pieceBox is a placed piece's own bounds with the way back into its frame: a
// point inside the piece's room is not near its walls, but it lies in the box.
type pieceBox struct {
	m   Mat
	s2  float32 // the squared scale, since the matrix is a turn times a scale
	box Box
}

func newPieceBox(m Mat, box Box) pieceBox {
	s2 := m[0]*m[0] + m[4]*m[4] + m[8]*m[8]
	return pieceBox{m: m, s2: s2, box: box}
}

// contains is whether the layout-space point lies in the piece's box.
func (p pieceBox) contains(v [3]float32) bool {
	if p.s2 <= 0 {
		return false
	}
	d := [3]float32{v[0] - p.m[3], v[1] - p.m[7], v[2] - p.m[11]}
	var local [3]float32
	for k := 0; k < 3; k++ {
		// The transpose of the turn-and-scale, over the squared scale.
		local[k] = (p.m[k]*d[0] + p.m[4+k]*d[1] + p.m[8+k]*d[2]) / p.s2
	}
	return p.box.Contains(local, 0)
}

// Covered counts the points t puts on the geometry or inside a placed piece's
// box: a light hanging in the middle of a cavern is in the piece that makes the
// cavern even where no wall is within a cell of it.
func (b *Built) Covered(t Transform, pts [][3]float32) int {
	n := 0
	for _, p := range pts {
		q := t.Inverse(p)
		if b.Dense.Near(q) {
			n++
			continue
		}
		for _, pc := range b.pieces {
			if pc.contains(q) {
				n++
				break
			}
		}
	}
	return n
}

// Fit finds the zone's transform. The fit is made against the models'
// vertices; a zone that comes out under nine in ten of its points on the
// geometry or inside a piece is fitted again against the surface sampled
// across the triangles, and the better of the two is kept. A zone that fits
// well keeps the first, so its origin does not wander over the flat top of a
// good fit.
func (b *Built) Fit(l *Layout) Fit {
	f := FitInterior(l, b.Surface)
	pts := l.Points()
	if len(pts) == 0 || float64(b.Covered(f.Transform, pts)) >= 0.9*float64(len(pts)) {
		return f
	}
	g := FitInterior(l, b.Dense)
	if b.Covered(g.Transform, pts) > b.Covered(f.Transform, pts) {
		return g
	}
	return f
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
