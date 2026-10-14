package interior

// Tying layout space to zone space. A city or dungeon's pieces are placed
// around an origin of their own, but its exits (ZONEJUMP.csv) and lights
// (lights.csv) are given in zone coordinates, and nothing says how the two
// meet. The data does, though: lights hang on walls and exits stand on
// floors, so a right transform puts them on the geometry.
//
// Measured that way, every Midgard interior uses one convention: layout x
// is mirrored, and the origin lies near (23584, 23904) on the ground. Fitted
// jointly, that one origin at height 16064 puts 584 of the twelve Midgard
// dungeons' 698 exits and lights on their geometry, each zone within a few
// points of its own best; Jordheim takes the same ground position at 8064,
// 36 of its 37. So the orientation is fixed, and each zone's origin is
// searched close to there on the ground and at any height, and recorded with
// how many points it puts on the geometry. Zones with few points -- Darkspire
// has three -- are held near the shared origin rather than fitted loose.

import (
	"math"
)

// Transform takes layout space to zone space: mirror y, turn about z by
// Turn quarter turns, then add Origin. Every Midgard interior is Turn 2
// with MirrorY -- x mirrored -- but the form says what it is.
type Transform struct {
	Origin  [3]float32 `json:"origin"`
	Turn    int        `json:"turn"`    // quarter turns anticlockwise, 0..3
	MirrorY bool       `json:"mirrorY"` // y negated before the turn
}

// Mirrored is the orientation every Midgard interior has: x negated.
var Mirrored = Transform{Turn: 2, MirrorY: true}

// Apply maps a layout point to zone coordinates.
func (t Transform) Apply(p [3]float32) [3]float32 {
	x, y := p[0], p[1]
	if t.MirrorY {
		y = -y
	}
	for i := 0; i < t.Turn; i++ {
		x, y = -y, x
	}
	return [3]float32{x + t.Origin[0], y + t.Origin[1], p[2] + t.Origin[2]}
}

// Inverse maps a zone point back to layout space.
func (t Transform) Inverse(p [3]float32) [3]float32 {
	x, y := p[0]-t.Origin[0], p[1]-t.Origin[1]
	for i := 0; i < t.Turn; i++ {
		x, y = y, -x
	}
	if t.MirrorY {
		y = -y
	}
	return [3]float32{x, y, p[2] - t.Origin[2]}
}

// Box is an axis-aligned box.
type Box struct {
	Min [3]float32 `json:"min"`
	Max [3]float32 `json:"max"`
}

// Contains is whether p lies in the box, grown by margin on every side.
func (b Box) Contains(p [3]float32, margin float32) bool {
	for k := 0; k < 3; k++ {
		if p[k] < b.Min[k]-margin || p[k] > b.Max[k]+margin {
			return false
		}
	}
	return true
}

// Rotation is the 3x3 matrix of an angle about an axis, row-major. A zero
// axis is taken as z.
func Rotation(angle float32, axis [3]float32) [9]float32 {
	x, y, z := float64(axis[0]), float64(axis[1]), float64(axis[2])
	n := math.Sqrt(x*x + y*y + z*z)
	if n < 1e-9 {
		x, y, z, n = 0, 0, 1, 1
	}
	x, y, z = x/n, y/n, z/n
	c, s := math.Cos(float64(angle)), math.Sin(float64(angle))
	t := 1 - c
	return [9]float32{
		float32(t*x*x + c), float32(t*x*y - s*z), float32(t*x*z + s*y),
		float32(t*x*y + s*z), float32(t*y*y + c), float32(t*y*z - s*x),
		float32(t*x*z - s*y), float32(t*y*z + s*x), float32(t*z*z + c),
	}
}

// Point moves a model-space point to where the placement puts it in layout
// space.
func (p Placement) Point(v [3]float32) [3]float32 {
	r := Rotation(p.Angle, p.Axis)
	var w [3]float32
	for k := 0; k < 3; k++ {
		w[k] = (r[k*3]*v[0]+r[k*3+1]*v[1]+r[k*3+2]*v[2])*p.Scale + p.Pos[k]
	}
	return w
}

// Place is a placement's model bounds moved to where it stands in layout
// space.
func Place(p Placement, min, max [3]float32) Box {
	inf := float32(math.Inf(1))
	b := Box{Min: [3]float32{inf, inf, inf}, Max: [3]float32{-inf, -inf, -inf}}
	for c := 0; c < 8; c++ {
		v := min
		if c&1 != 0 {
			v[0] = max[0]
		}
		if c&2 != 0 {
			v[1] = max[1]
		}
		if c&4 != 0 {
			v[2] = max[2]
		}
		w := p.Point(v)
		for k := 0; k < 3; k++ {
			b.Min[k] = float32(math.Min(float64(b.Min[k]), float64(w[k])))
			b.Max[k] = float32(math.Max(float64(b.Max[k]), float64(w[k])))
		}
	}
	return b
}

// Surface is where an interior's geometry lies, as a set of occupied cells
// in layout space, grown by one cell, so a lookup asks whether geometry
// lies within about a cell of a point.
type Surface struct {
	on map[[3]int32]bool
}

// SurfaceCell is the surface map's resolution.
const SurfaceCell = 128

// NewSurface maps the given layout-space vertices.
func NewSurface(verts [][3]float32) *Surface {
	s := &Surface{on: map[[3]int32]bool{}}
	for _, v := range verts {
		c := cellOf(v)
		for dx := int32(-1); dx <= 1; dx++ {
			for dy := int32(-1); dy <= 1; dy++ {
				for dz := int32(-1); dz <= 1; dz++ {
					s.on[[3]int32{c[0] + dx, c[1] + dy, c[2] + dz}] = true
				}
			}
		}
	}
	return s
}

func cellOf(v [3]float32) [3]int32 {
	return [3]int32{
		int32(math.Floor(float64(v[0]) / SurfaceCell)),
		int32(math.Floor(float64(v[1]) / SurfaceCell)),
		int32(math.Floor(float64(v[2]) / SurfaceCell)),
	}
}

// Near is whether geometry lies within about a cell of p.
func (s *Surface) Near(p [3]float32) bool { return s.on[cellOf(p)] }

// Count is how many points t puts on the surface.
func (s *Surface) Count(t Transform, pts [][3]float32) int {
	n := 0
	for _, p := range pts {
		if s.Near(t.Inverse(p)) {
			n++
		}
	}
	return n
}

// Fit is a transform and how well the zone's own points bear it out.
type Fit struct {
	Transform
	OnSurface int `json:"onSurface"` // exits and lights on the geometry
	Points    int `json:"points"`    // exits and lights in all
}

// Score is the share of points on the geometry.
func (f Fit) Score() float64 {
	if f.Points == 0 {
		return 0
	}
	return float64(f.OnSurface) / float64(f.Points)
}

// Points are the zone-space points a fit is judged by: each exit box's
// centre and each light.
func (l *Layout) Points() [][3]float32 {
	var pts [][3]float32
	for _, e := range l.Entrances {
		pts = append(pts, e.Centre())
	}
	for _, li := range l.Lights {
		pts = append(pts, li.Pos)
	}
	return pts
}

// nearOrigin is where every Midgard interior's origin lies on the ground,
// within a few hundred units; heights are searched over the whole zone.
var nearOrigin = [2]float32{23584, 23904}

// FitInterior finds the origin, with the mirrored orientation, that puts
// most of the zone's exits and lights on its geometry: coarsely within 512
// units of nearOrigin on the ground and at every height, then finely around
// the best, taking the middle of the offsets that tie.
func FitInterior(l *Layout, s *Surface) Fit {
	pts := l.Points()
	f := Fit{Transform: Mirrored, Points: len(pts)}
	if len(pts) == 0 {
		f.Origin = [3]float32{nearOrigin[0], nearOrigin[1], 0}
		return f
	}
	best, n := Mirrored, -1
	for dx := float32(-512); dx <= 512; dx += 64 {
		for dy := float32(-512); dy <= 512; dy += 64 {
			for z := float32(0); z <= 32768; z += 128 {
				t := Mirrored
				t.Origin = [3]float32{nearOrigin[0] + dx, nearOrigin[1] + dy, z}
				if k := s.Count(t, pts); k > n {
					best, n = t, k
				}
			}
		}
	}
	n = -1
	var sum [3]float64
	var ties float64
	for dx := float32(-192); dx <= 192; dx += 16 {
		for dy := float32(-192); dy <= 192; dy += 16 {
			for dz := float32(-192); dz <= 192; dz += 16 {
				t := best
				t.Origin = [3]float32{best.Origin[0] + dx, best.Origin[1] + dy, best.Origin[2] + dz}
				k := s.Count(t, pts)
				if k > n {
					n, sum, ties = k, [3]float64{}, 0
				}
				if k == n {
					for i := range sum {
						sum[i] += float64(t.Origin[i])
					}
					ties++
				}
			}
		}
	}
	f.Origin = [3]float32{float32(sum[0] / ties), float32(sum[1] / ties), float32(sum[2] / ties)}
	f.OnSurface = s.Count(f.Transform, pts)
	return f
}

// Mat is a model-to-layout transform, 3x4 row-major: rotation and scale in
// the first three columns, translation in the last.
type Mat [12]float32

// Matrix is a placement's transform.
func (p Placement) Matrix() Mat {
	r := Rotation(p.Angle, p.Axis)
	var m Mat
	for k := 0; k < 3; k++ {
		m[k*4] = r[k*3] * p.Scale
		m[k*4+1] = r[k*3+1] * p.Scale
		m[k*4+2] = r[k*3+2] * p.Scale
		m[k*4+3] = p.Pos[k]
	}
	return m
}

// Mul is a then b: (a.Mul(b)).Point(v) == a.Point(b.Point(v)).
func (a Mat) Mul(b Mat) Mat {
	var m Mat
	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			v := a[i*4]*b[j] + a[i*4+1]*b[4+j] + a[i*4+2]*b[8+j]
			if j == 3 {
				v += a[i*4+3]
			}
			m[i*4+j] = v
		}
	}
	return m
}

// Point applies the transform.
func (m Mat) Point(v [3]float32) [3]float32 {
	var w [3]float32
	for k := 0; k < 3; k++ {
		w[k] = m[k*4]*v[0] + m[k*4+1]*v[1] + m[k*4+2]*v[2] + m[k*4+3]
	}
	return w
}

// Placed is a model where the layout finally puts it.
type Placed struct {
	Model string
	M     Mat
	Prop  bool
}

// Placed lists every model the layout puts down: pieces, then props. A
// prop's position is local to the placed piece its second unnamed column
// names, as the data shows -- read that way all 707 of Abandoned Mines'
// props and all 49 of Darkspire's lie on the geometry, read in dungeon
// space 634 and none. A prop naming no piece is skipped and counted.
func (l *Layout) Placed() (out []Placed, skipped int) {
	for _, p := range l.Pieces {
		out = append(out, Placed{Model: p.Model, M: p.Matrix()})
	}
	for _, p := range l.Props {
		if p.B < 0 || p.B >= len(l.Pieces) {
			skipped++
			continue
		}
		out = append(out, Placed{Model: p.Model, M: l.Pieces[p.B].Matrix().Mul(p.Matrix()), Prop: true})
	}
	return out, skipped
}
