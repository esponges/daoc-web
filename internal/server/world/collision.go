package world

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// How a walker meets the scenery. Hulls, and for an interior the visible
// walls, are triangles; what blocks is whatever stands between knee and head
// height above the walker's feet, and anything lower is stepped over.
const (
	BodyRadius = 32.0  // a walker's radius, units
	StepHeight = 60.0  // the highest thing stepped over, units
	BodyHeight = 170.0 // above the feet, the highest thing that blocks
	wallNZ     = 0.5   // a triangle blocks when its normal is this flat or flatter
	floorNZ    = 0.34  // and is a floor when its normal is at least this upright
	binSize    = 256.0
)

type tri struct {
	a, b, c    [3]float32
	zmin, zmax float32
}

type cyl struct {
	x, y, r, zmin, zmax float32
}

// Collision is the static obstacles of a zone, binned over the ground plane.
// It is read-only once built, so every walker in the zone shares it.
type Collision struct {
	tris      []tri
	cyls      []cyl
	x0, y0    float64
	w, h      int
	wallBins  [][]int32 // per bin: indices into tris, then cyls offset by len(tris)
	floorTris []tri
	floorBins [][]int32
	hasFloor  bool
}

// instance is a placed model, as props.json writes it.
type instance struct {
	M    int         `json:"m"`
	P    [3]float64  `json:"p"`
	Yaw  float64     `json:"yaw"`
	S    float64     `json:"s"`
	Axis *[3]float64 `json:"axis"`
}

// place is the zone-space position of a model-space point under an instance:
// the same rotation about the instance's axis, scale and translation as the
// viewer's instancePoint.
func (in *instance) place(v [3]float32) [3]float32 {
	a := [3]float64{0, 0, 1}
	if in.Axis != nil {
		a = *in.Axis
	}
	c, s := math.Cos(in.Yaw), math.Sin(in.Yaw)
	x, y, z := float64(v[0]), float64(v[1]), float64(v[2])
	d := (a[0]*x + a[1]*y + a[2]*z) * (1 - c)
	m0 := x*c + (a[1]*z-a[2]*y)*s + a[0]*d
	m1 := y*c + (a[2]*x-a[0]*z)*s + a[1]*d
	m2 := z*c + (a[0]*y-a[1]*x)*s + a[2]*d
	sc := in.S
	if sc == 0 {
		sc = 1
	}
	return [3]float32{float32(sc*m0 + in.P[0]), float32(sc*m1 + in.P[1]), float32(sc*m2 + in.P[2])}
}

func newTri(a, b, c [3]float32) tri {
	t := tri{a: a, b: b, c: c}
	t.zmin = float32(math.Min(float64(a[2]), math.Min(float64(b[2]), float64(c[2]))))
	t.zmax = float32(math.Max(float64(a[2]), math.Max(float64(b[2]), float64(c[2]))))
	return t
}

// normalZ is |nz| of the triangle's unit normal, and whether it is degenerate.
func normalZ(t tri) (float64, bool) {
	ux, uy, uz := float64(t.b[0]-t.a[0]), float64(t.b[1]-t.a[1]), float64(t.b[2]-t.a[2])
	wx, wy, wz := float64(t.c[0]-t.a[0]), float64(t.c[1]-t.a[1]), float64(t.c[2]-t.a[2])
	nx, ny, nz := uy*wz-uz*wy, uz*wx-ux*wz, ux*wy-uy*wx
	n := math.Sqrt(nx*nx + ny*ny + nz*nz)
	if n == 0 {
		return 0, false
	}
	return math.Abs(nz) / n, true
}

// builder gathers triangles then bins them.
type builder struct {
	walls, floors []tri
	cyls          []cyl
}

func (b *builder) add(t tri, wantFloors bool) {
	nz, ok := normalZ(t)
	if !ok {
		return
	}
	if nz < wallNZ {
		b.walls = append(b.walls, t)
	}
	if wantFloors && nz >= floorNZ {
		b.floors = append(b.floors, t)
	}
}

func (b *builder) build(minX, minY, maxX, maxY float64) *Collision {
	c := &Collision{tris: b.walls, cyls: b.cyls, floorTris: b.floors, hasFloor: len(b.floors) > 0}
	c.x0, c.y0 = minX, minY
	c.w = int(math.Ceil((maxX-minX)/binSize)) + 1
	c.h = int(math.Ceil((maxY-minY)/binSize)) + 1
	c.wallBins = make([][]int32, c.w*c.h)
	c.floorBins = make([][]int32, c.w*c.h)
	span := func(x0, y0, x1, y1 float64, fn func(bin int)) {
		i0, i1 := int(math.Floor((x0-c.x0)/binSize)), int(math.Floor((x1-c.x0)/binSize))
		j0, j1 := int(math.Floor((y0-c.y0)/binSize)), int(math.Floor((y1-c.y0)/binSize))
		for j := max(j0, 0); j <= min(j1, c.h-1); j++ {
			for i := max(i0, 0); i <= min(i1, c.w-1); i++ {
				fn(j*c.w + i)
			}
		}
	}
	bbox := func(t tri) (float64, float64, float64, float64) {
		return math.Min(float64(t.a[0]), math.Min(float64(t.b[0]), float64(t.c[0]))),
			math.Min(float64(t.a[1]), math.Min(float64(t.b[1]), float64(t.c[1]))),
			math.Max(float64(t.a[0]), math.Max(float64(t.b[0]), float64(t.c[0]))),
			math.Max(float64(t.a[1]), math.Max(float64(t.b[1]), float64(t.c[1])))
	}
	for i, t := range c.tris {
		x0, y0, x1, y1 := bbox(t)
		span(x0-BodyRadius, y0-BodyRadius, x1+BodyRadius, y1+BodyRadius, func(bin int) { c.wallBins[bin] = append(c.wallBins[bin], int32(i)) })
	}
	for i, t := range c.cyls {
		span(float64(t.x-t.r)-BodyRadius, float64(t.y-t.r)-BodyRadius, float64(t.x+t.r)+BodyRadius, float64(t.y+t.r)+BodyRadius,
			func(bin int) { c.wallBins[bin] = append(c.wallBins[bin], int32(len(c.tris)+i)) })
	}
	for i, t := range c.floorTris {
		x0, y0, x1, y1 := bbox(t)
		span(x0, y0, x1, y1, func(bin int) { c.floorBins[bin] = append(c.floorBins[bin], int32(i)) })
	}
	return c
}

// Slide returns where a walker moving from (x0,y0) to (x1,y1) ends up with
// its feet at z: the destination pushed out of whatever it would overlap,
// which slides it along a wall instead of stopping it dead.
func (c *Collision) Slide(x0, y0, x1, y1, z float64) (float64, float64) {
	if c == nil {
		return x1, y1
	}
	px, py := x1, y1
	zlo, zhi := float32(z+StepHeight), float32(z+BodyHeight)
	for iter := 0; iter < 3; iter++ {
		moved := false
		i := int(math.Floor((px - c.x0) / binSize))
		j := int(math.Floor((py - c.y0) / binSize))
		if i < 0 || j < 0 || i >= c.w || j >= c.h {
			break
		}
		for _, idx := range c.wallBins[j*c.w+i] {
			if int(idx) >= len(c.tris) {
				k := c.cyls[int(idx)-len(c.tris)]
				if k.zmax <= zlo || k.zmin >= zhi {
					continue
				}
				dx, dy := px-float64(k.x), py-float64(k.y)
				d := math.Hypot(dx, dy)
				if lim := float64(k.r) + BodyRadius; d < lim {
					if d < 1e-6 {
						dx, dy, d = x0-float64(k.x), y0-float64(k.y), math.Hypot(x0-float64(k.x), y0-float64(k.y))
						if d < 1e-6 {
							continue
						}
					}
					px, py = float64(k.x)+dx/d*lim, float64(k.y)+dy/d*lim
					moved = true
				}
				continue
			}
			t := &c.tris[idx]
			if t.zmax <= zlo || t.zmin >= zhi {
				continue
			}
			edges := [3][2][3]float32{{t.a, t.b}, {t.b, t.c}, {t.c, t.a}}
			for _, e := range edges {
				cx, cyy, d := closestOnSegment(px, py, float64(e[0][0]), float64(e[0][1]), float64(e[1][0]), float64(e[1][1]))
				if d < BodyRadius {
					var nx, ny float64
					if d < 1e-6 {
						// On the wall itself: leave the way we came.
						nx, ny = x0-px, y0-py
						if l := math.Hypot(nx, ny); l > 1e-9 {
							nx, ny = nx/l, ny/l
						} else {
							continue
						}
					} else {
						nx, ny = (px-cx)/d, (py-cyy)/d
					}
					px, py = cx+nx*BodyRadius, cyy+ny*BodyRadius
					moved = true
				}
			}
		}
		if !moved {
			break
		}
	}
	return px, py
}

// closestOnSegment is the point on segment ab nearest p, and the distance.
func closestOnSegment(px, py, ax, ay, bx, by float64) (float64, float64, float64) {
	dx, dy := bx-ax, by-ay
	l2 := dx*dx + dy*dy
	t := 0.0
	if l2 > 0 {
		t = math.Max(0, math.Min(1, ((px-ax)*dx+(py-ay)*dy)/l2))
	}
	cx, cy := ax+t*dx, ay+t*dy
	return cx, cy, math.Hypot(px-cx, py-cy)
}

// FloorBelow is the highest floor under (x, y) no more than a step above
// fromZ, and whether there is one.
func (c *Collision) FloorBelow(x, y, fromZ float64) (float64, bool) {
	if c == nil || !c.hasFloor {
		return 0, false
	}
	i := int(math.Floor((x - c.x0) / binSize))
	j := int(math.Floor((y - c.y0) / binSize))
	if i < 0 || j < 0 || i >= c.w || j >= c.h {
		return 0, false
	}
	top := fromZ + StepHeight
	best, found := 0.0, false
	for _, idx := range c.floorBins[j*c.w+i] {
		t := &c.floorTris[idx]
		ax, ay, bx, by, cx, cy := float64(t.a[0]), float64(t.a[1]), float64(t.b[0]), float64(t.b[1]), float64(t.c[0]), float64(t.c[1])
		d := (by-cy)*(ax-cx) + (cx-bx)*(ay-cy)
		if d == 0 {
			continue
		}
		l1 := ((by-cy)*(x-cx) + (cx-bx)*(y-cy)) / d
		l2 := ((cy-ay)*(x-cx) + (ax-cx)*(y-cy)) / d
		l3 := 1 - l1 - l2
		if l1 < -1e-4 || l2 < -1e-4 || l3 < -1e-4 {
			continue
		}
		z := l1*float64(t.a[2]) + l2*float64(t.b[2]) + l3*float64(t.c[2])
		if z <= top && (!found || z > best) {
			best, found = z, true
		}
	}
	return best, found
}

// --- loading ---

type propsManifest struct {
	Stride      int          `json:"stride"`
	VertexBytes int          `json:"vertexBytes"`
	IndexBytes  int          `json:"indexBytes"`
	Models      []propsModel `json:"models"`
	Instances   []instance   `json:"instances"`
}

type propsModel struct {
	Groups []struct {
		First     int  `json:"first"`
		Count     int  `json:"count"`
		Billboard bool `json:"billboard"`
		Additive  bool `json:"additive"`
	} `json:"groups"`
}

func zoneDir(webData string, zone int) string {
	return filepath.Join(webData, fmt.Sprintf("zone%03d", zone))
}

// LoadOutdoorCollision reads a zone's collision.json/bin (written by
// propconv) and places every hull with props.json's instances. A model with
// no hull but a trunk radius blocks as a cylinder. It returns nil if the zone
// has no collision files: it then moves freely.
func LoadOutdoorCollision(webData string, zone int, extent float64) *Collision {
	dir := filepath.Join(zoneDir(webData, zone), "props")
	cj, err := os.ReadFile(filepath.Join(dir, "collision.json"))
	if err != nil {
		return nil
	}
	cb, err := os.ReadFile(filepath.Join(dir, "collision.bin"))
	if err != nil {
		return nil
	}
	pj, err := os.ReadFile(filepath.Join(dir, "props.json"))
	if err != nil {
		return nil
	}
	var cm struct {
		VertexBytes int `json:"vertexBytes"`
		Models      []struct {
			V, NV, I, NI int
			R            float32 `json:"r"`
		} `json:"models"`
	}
	var pm propsManifest
	if json.Unmarshal(cj, &cm) != nil || json.Unmarshal(pj, &pm) != nil || len(cm.Models) != len(pm.Models) {
		return nil
	}
	verts := make([][3]float32, cm.VertexBytes/12)
	for i := range verts {
		for k := 0; k < 3; k++ {
			verts[i][k] = math.Float32frombits(binary.LittleEndian.Uint32(cb[i*12+k*4:]))
		}
	}
	idx := func(i int) uint32 { return binary.LittleEndian.Uint32(cb[cm.VertexBytes+i*4:]) }
	b := &builder{}
	for _, in := range pm.Instances {
		if in.M < 0 || in.M >= len(cm.Models) {
			continue
		}
		m := cm.Models[in.M]
		if m.NI > 0 {
			for t := 0; t+2 < m.NI; t += 3 {
				va, vb, vc := verts[m.V+int(idx(m.I+t))], verts[m.V+int(idx(m.I+t+1))], verts[m.V+int(idx(m.I+t+2))]
				b.add(newTri(in.place(va), in.place(vb), in.place(vc)), false)
			}
		} else if m.R > 0 {
			sc := in.S
			if sc == 0 {
				sc = 1
			}
			// A trunk: a cylinder from the ground up, as tall as the player.
			b.cyls = append(b.cyls, cyl{x: float32(in.P[0]), y: float32(in.P[1]), r: m.R * float32(sc),
				zmin: float32(in.P[2]) - 50, zmax: float32(in.P[2]) + 400})
		}
	}
	return b.build(0, 0, extent, extent)
}

// LoadInteriorGeometry gathers the floors and walls of a city or dungeon from
// its visible geometry (props.bin placed by props.json's instances).
func LoadInteriorGeometry(webData string, zone int, bounds [2][3]float64) (*Collision, error) {
	dir := filepath.Join(zoneDir(webData, zone), "props")
	pj, err := os.ReadFile(filepath.Join(dir, "props.json"))
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "props.bin"))
	if err != nil {
		return nil, err
	}
	var pm propsManifest
	if err := json.Unmarshal(pj, &pm); err != nil {
		return nil, err
	}
	stride := pm.Stride
	nv := pm.VertexBytes / stride
	vert := func(i uint32) [3]float32 {
		o := int(i) * stride
		return [3]float32{
			math.Float32frombits(binary.LittleEndian.Uint32(raw[o:])),
			math.Float32frombits(binary.LittleEndian.Uint32(raw[o+4:])),
			math.Float32frombits(binary.LittleEndian.Uint32(raw[o+8:])),
		}
	}
	_ = nv
	perModel := make([][]instance, len(pm.Models))
	for _, in := range pm.Instances {
		if in.M >= 0 && in.M < len(perModel) {
			perModel[in.M] = append(perModel[in.M], in)
		}
	}
	b := &builder{}
	for mi, m := range pm.Models {
		for _, g := range m.Groups {
			if g.Billboard || g.Additive {
				continue
			}
			for t := g.First; t+2 < g.First+g.Count; t += 3 {
				ia := binary.LittleEndian.Uint32(raw[pm.VertexBytes+t*4:])
				ib := binary.LittleEndian.Uint32(raw[pm.VertexBytes+(t+1)*4:])
				ic := binary.LittleEndian.Uint32(raw[pm.VertexBytes+(t+2)*4:])
				va, vb, vc := vert(ia), vert(ib), vert(ic)
				for k := range perModel[mi] {
					in := &perModel[mi][k]
					b.add(newTri(in.place(va), in.place(vb), in.place(vc)), true)
				}
			}
		}
	}
	return b.build(bounds[0][0]-binSize, bounds[0][1]-binSize, bounds[1][0]+binSize, bounds[1][1]+binSize), nil
}

// OutdoorCollision is the Config.Collision hook for outdoor zones.
func OutdoorCollision(webData string, zone int) Mover {
	c := LoadOutdoorCollision(webData, zone, 65536)
	if c == nil {
		return nil
	}
	return c
}
