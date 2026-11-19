package rules

import "math"

// Line of sight (rules-line-of-sight). The rules here are the geometry of one
// question: does the segment between two eye points cross scenery? The world
// package gathers the scenery near the segment from its collision bins and
// asks.

// EyeHeight is how far above a body's feet its eye point is, in units. The
// value is this project's choice (a little below the 170 a walker's head
// reaches); tune it in the browser.
const EyeHeight = 150.0

// SightMessage is what a player is told when a spell cannot be aimed.
const SightMessage = "You can't see your target!"

// GroundStep is how far apart the heightmap is sampled along a segment, in
// units: half a collision bin.
const GroundStep = 128.0

// Eye is the eye point of a body standing at feet.
func Eye(feet [3]float64) [3]float64 { return [3]float64{feet[0], feet[1], feet[2] + EyeHeight} }

// Cylinder is a tree trunk or other standing post.
type Cylinder struct{ X, Y, R, ZMin, ZMax float64 }

// SegmentTriangle says whether the segment a-b crosses triangle t (either
// side). A touch at the very end of the segment does not count: the eye of a
// body standing against a wall is not behind it.
func SegmentTriangle(a, b [3]float64, t [3][3]float64) bool {
	const eps = 1e-9
	d := [3]float64{b[0] - a[0], b[1] - a[1], b[2] - a[2]}
	e1 := [3]float64{t[1][0] - t[0][0], t[1][1] - t[0][1], t[1][2] - t[0][2]}
	e2 := [3]float64{t[2][0] - t[0][0], t[2][1] - t[0][1], t[2][2] - t[0][2]}
	p := cross(d, e2)
	det := dot(e1, p)
	if math.Abs(det) < eps {
		return false // parallel to the plane
	}
	inv := 1 / det
	s := [3]float64{a[0] - t[0][0], a[1] - t[0][1], a[2] - t[0][2]}
	u := dot(s, p) * inv
	if u < 0 || u > 1 {
		return false
	}
	q := cross(s, e1)
	v := dot(d, q) * inv
	if v < 0 || u+v > 1 {
		return false
	}
	at := dot(e2, q) * inv
	return at > 1e-3 && at < 1-1e-3
}

// SegmentCylinder says whether the segment a-b passes through the cylinder.
func SegmentCylinder(a, b [3]float64, c Cylinder) bool {
	dx, dy := b[0]-a[0], b[1]-a[1]
	fx, fy := a[0]-c.X, a[1]-c.Y
	qa := dx*dx + dy*dy
	var lo, hi float64
	if qa < 1e-12 {
		if fx*fx+fy*fy > c.R*c.R {
			return false
		}
		lo, hi = 0, 1
	} else {
		qb := 2 * (fx*dx + fy*dy)
		qc := fx*fx + fy*fy - c.R*c.R
		disc := qb*qb - 4*qa*qc
		if disc < 0 {
			return false
		}
		sq := math.Sqrt(disc)
		lo, hi = math.Max((-qb-sq)/(2*qa), 0), math.Min((-qb+sq)/(2*qa), 1)
		if lo > hi {
			return false
		}
	}
	zlo, zhi := a[2]+lo*(b[2]-a[2]), a[2]+hi*(b[2]-a[2])
	if zlo > zhi {
		zlo, zhi = zhi, zlo
	}
	return zhi >= c.ZMin && zlo <= c.ZMax
}

// SegmentUnderGround says whether the segment dips below the heightmap
// anywhere along its length, sampling every GroundStep units.
func SegmentUnderGround(a, b [3]float64, ground func(x, y float64) float64) bool {
	if ground == nil {
		return false
	}
	n := int(math.Ceil(math.Hypot(b[0]-a[0], b[1]-a[1]) / GroundStep))
	for i := 1; i < n; i++ {
		t := float64(i) / float64(n)
		x, y := a[0]+t*(b[0]-a[0]), a[1]+t*(b[1]-a[1])
		if a[2]+t*(b[2]-a[2]) < ground(x, y) {
			return true
		}
	}
	return false
}

// Scenery is what stands between two eyes: triangles that block (walls, and
// in an interior floors and ceilings), trunks, and the terrain. A zone with
// no collision data has none, and everything sees everything.
type Scenery struct {
	Tris   [][3][3]float64
	Cyls   []Cylinder
	Ground func(x, y float64) float64
}

// Visible says whether a and b, eye points, see each other.
func Visible(a, b [3]float64, s Scenery) bool {
	if SegmentUnderGround(a, b, s.Ground) {
		return false
	}
	for _, c := range s.Cyls {
		if SegmentCylinder(a, b, c) {
			return false
		}
	}
	for _, t := range s.Tris {
		if SegmentTriangle(a, b, t) {
			return false
		}
	}
	return true
}

// NeedsSight says whether a player's spell must see its target: not a spell
// on the caster, on the group, or an area spell centred on the caster.
func NeedsSight(self, group, centredOnCaster bool) bool {
	return !self && !group && !centredOnCaster
}

func cross(a, b [3]float64) [3]float64 {
	return [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}
func dot(a, b [3]float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }

func init() {
	const c = "rules-line-of-sight"
	register(
		Rule{ID: "los.visible", Capability: c, Requirement: "Line of sight is a clear segment between eyes",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"Open ground", "A hill between them", "A dungeon wall between them", "A tree trunk"}},
		Rule{ID: "los.cast", Capability: c, Requirement: "A player's hostile spell needs line of sight at start and at landing",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc"}, Scenarios: []string{"Casting at a creature behind a wall", "The target steps behind cover during the cast", "Area spell around the caster"}},
		Rule{ID: "los.creature", Capability: c, Requirement: "A creature without line of sight closes in instead of casting",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc"}, Scenarios: []string{"Caster creature behind a corner"}},
		Rule{ID: "los.cost", Capability: c, Requirement: "Line of sight is cheap enough for every cast",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"A zone without collision data"}},
	)
}
