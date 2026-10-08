// Command charshot renders an exported character to a PNG without a browser.
//
// It exists because the WebGL viewer cannot be inspected from a terminal, and
// "the numbers look right" is a weak claim about a mesh. charshot reads exactly
// what the browser reads -- char.json, mesh.bin and the PNG textures -- skins
// it on the CPU, poses it with the same gait, and rasterises it. If the figure
// looks like a clothed man with his legs apart, then the export, the skeleton,
// the inverse bind matrices, the weights and the UVs are all sound, and the
// only thing left untested is the shader plumbing itself.
//
// It is also useful on its own: it previews any figure and shape set from the
// command line, which is much faster than reloading a page.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

// Vertex layout, matching charconv.
const (
	offPos    = 0
	offNormal = 12
	offUV     = 24
	offBone   = 32
	offWeight = 36
)

type bone struct {
	Name   string     `json:"name"`
	Parent int        `json:"parent"`
	T      [3]float64 `json:"t"`
	R      [9]float64 `json:"r"`
	S      float64    `json:"s"`
}

type shape struct {
	Name    string        `json:"name"`
	First   int           `json:"first"`
	Count   int           `json:"count"`
	Bones   []int         `json:"bones"`
	InvBind [][12]float64 `json:"invBind"`
	Texture string        `json:"texture"`
	Diffuse [3]float64    `json:"diffuse"`
}

type manifest struct {
	Name        string     `json:"name"`
	Source      string     `json:"source"`
	VertexCount int        `json:"vertexCount"`
	IndexCount  int        `json:"indexCount"`
	Stride      int        `json:"stride"`
	Height      float64    `json:"height"`
	Min         [3]float64 `json:"min"`
	Max         [3]float64 `json:"max"`
	Bones       []bone     `json:"bones"`
	Shapes      []shape    `json:"shapes"`
	Equip       []struct {
		Slot string `json:"slot"`
		Bone string `json:"bone"`
	} `json:"equip"`
}

// xf is a rigid transform as three rows of four, the same packing the manifest
// and the shader use.
type xf [12]float64

var identity = xf{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0}

func mul(a, b xf) xf {
	var o xf
	for r := 0; r < 3; r++ {
		a0, a1, a2 := a[r*4+0], a[r*4+1], a[r*4+2]
		o[r*4+0] = a0*b[0] + a1*b[4] + a2*b[8]
		o[r*4+1] = a0*b[1] + a1*b[5] + a2*b[9]
		o[r*4+2] = a0*b[2] + a1*b[6] + a2*b[10]
		o[r*4+3] = a0*b[3] + a1*b[7] + a2*b[11] + a[r*4+3]
	}
	return o
}

func (m xf) point(v [3]float64) [3]float64 {
	return [3]float64{
		m[0]*v[0] + m[1]*v[1] + m[2]*v[2] + m[3],
		m[4]*v[0] + m[5]*v[1] + m[6]*v[2] + m[7],
		m[8]*v[0] + m[9]*v[1] + m[10]*v[2] + m[11],
	}
}

func (m xf) dir(v [3]float64) [3]float64 {
	return [3]float64{
		m[0]*v[0] + m[1]*v[1] + m[2]*v[2],
		m[4]*v[0] + m[5]*v[1] + m[6]*v[2],
		m[8]*v[0] + m[9]*v[1] + m[10]*v[2],
	}
}

func rotX(a float64) xf {
	c, s := math.Cos(a), math.Sin(a)
	return xf{1, 0, 0, 0, 0, c, -s, 0, 0, s, c, 0}
}

func rotZ(a float64) xf {
	c, s := math.Cos(a), math.Sin(a)
	return xf{c, -s, 0, 0, s, c, 0, 0, 0, 0, 1, 0}
}

// rotTranspose returns transpose(rotation of a) * b.
func rotTranspose(a, b xf) xf {
	var o xf
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			o[r*4+c] = a[0*4+r]*b[0*4+c] + a[1*4+r]*b[1*4+c] + a[2*4+r]*b[2*4+c]
		}
	}
	return o
}

type skeleton struct {
	bones     []bone
	order     []int
	bindLocal []xf
	bindWorld []xf
	animLocal []xf
	world     []xf
	byName    map[string]int
}

func newSkeleton(bones []bone) (*skeleton, error) {
	s := &skeleton{bones: bones, byName: map[string]int{}}
	n := len(bones)
	s.bindLocal = make([]xf, n)
	s.bindWorld = make([]xf, n)
	s.animLocal = make([]xf, n)
	s.world = make([]xf, n)
	for i, b := range bones {
		var m xf
		for r := 0; r < 3; r++ {
			m[r*4+0] = b.R[r*3+0] * b.S
			m[r*4+1] = b.R[r*3+1] * b.S
			m[r*4+2] = b.R[r*3+2] * b.S
			m[r*4+3] = b.T[r]
		}
		s.bindLocal[i] = m
		s.byName[b.Name] = i
	}
	children := make([][]int, n)
	var roots []int
	for i, b := range bones {
		if b.Parent >= 0 {
			children[b.Parent] = append(children[b.Parent], i)
		} else {
			roots = append(roots, i)
		}
	}
	stack := append([]int{}, roots...)
	for len(stack) > 0 {
		i := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		s.order = append(s.order, i)
		stack = append(stack, children[i]...)
	}
	if len(s.order) != n {
		return nil, fmt.Errorf("%d of %d bones unreachable from a root", n-len(s.order), n)
	}
	for _, i := range s.order {
		if p := bones[i].Parent; p < 0 {
			s.bindWorld[i] = s.bindLocal[i]
		} else {
			s.bindWorld[i] = mul(s.bindWorld[p], s.bindLocal[i])
		}
	}
	return s, nil
}

func (s *skeleton) id(name string) int {
	if i, ok := s.byName[name]; ok {
		return i
	}
	return -1
}

// setJoint rotates one joint by an angle given in the body's frame, by
// conjugating the rotation into the joint's own frame.
//
// Only the rotation of the bind world transform takes part. Conjugating with
// the full transform would leave a translation behind and detach the limb.
func (s *skeleton) setJoint(i int, angle float64, axis byte) {
	if i < 0 || angle == 0 {
		return
	}
	r := rotX(angle)
	if axis == 'z' {
		r = rotZ(angle)
	}
	w := rotOnly(s.bindWorld[i])
	s.animLocal[i] = mul(s.bindLocal[i], mul(rotTranspose(w, r), w))
}

// rotOnly drops a transform's translation, keeping the rotation.
func rotOnly(a xf) xf {
	var o xf
	for r := 0; r < 3; r++ {
		o[r*4+0], o[r*4+1], o[r*4+2] = a[r*4+0], a[r*4+1], a[r*4+2]
	}
	return o
}

func (s *skeleton) resolve() {
	for _, i := range s.order {
		if p := s.bones[i].Parent; p < 0 {
			s.world[i] = s.animLocal[i]
		} else {
			s.world[i] = mul(s.world[p], s.animLocal[i])
		}
	}
}

// How far the arms come down from the authored T-pose, and the resting elbow.
const (
	armDrop  = 1.32
	armSwing = 0.42
	elbow    = 0.34
)

func rotY(a float64) xf {
	c, s := math.Cos(a), math.Sin(a)
	return xf{c, 0, s, 0, 0, 1, 0, 0, -s, 0, c, 0}
}

// setJointRot applies an arbitrary body-frame rotation to one joint.
func (s *skeleton) setJointRot(i int, r xf) {
	if i < 0 {
		return
	}
	w := rotOnly(s.bindWorld[i])
	s.animLocal[i] = mul(s.bindLocal[i], mul(rotTranspose(w, r), w))
}

// armSign is +1 or -1 depending on which way along the lateral axis an arm
// points, so that dropping both arms lowers them rather than raising one.
func (s *skeleton) armSign(j int) float64 {
	if j >= 0 && s.bindWorld[j][3] < 0 {
		return -1
	}
	return 1
}

// poseWalk mirrors web/skeleton.js so that what this renders is what the
// browser shows.
func (s *skeleton) poseWalk(phase, intensity float64) {
	copy(s.animLocal, s.bindLocal)
	sn, cs := math.Sin(phase), math.Cos(phase)

	// Arms come down out of the T-pose before any swing is applied; swinging
	// about the lateral axis does nothing to an arm lying along it.
	lArm, rArm := s.id("Bip01 L UpperArm"), s.id("Bip01 R UpperArm")
	s.setJointRot(lArm, mul(rotX(armSwing*intensity*sn), rotY(armDrop*s.armSign(lArm))))
	s.setJointRot(rArm, mul(rotX(-armSwing*intensity*sn), rotY(armDrop*s.armSign(rArm))))
	s.setJoint(s.id("Bip01 L Forearm"), -elbow-0.14*intensity*cs, 'x')
	s.setJoint(s.id("Bip01 R Forearm"), -elbow+0.14*intensity*cs, 'x')

	if intensity > 0.001 {
		hip, knee := 0.62*intensity, 0.95*intensity
		s.setJoint(s.id("Bip01 L Thigh"), -hip*sn, 'x')
		s.setJoint(s.id("Bip01 R Thigh"), hip*sn, 'x')
		s.setJoint(s.id("Bip01 L Calf"), knee*math.Max(0, -math.Sin(phase-0.6)), 'x')
		s.setJoint(s.id("Bip01 R Calf"), knee*math.Max(0, -math.Sin(phase+math.Pi-0.6)), 'x')
		s.setJoint(s.id("Bip01 L Foot"), -0.35*intensity*math.Max(0, -sn), 'x')
		s.setJoint(s.id("Bip01 R Foot"), -0.35*intensity*math.Max(0, sn), 'x')
		s.setJoint(s.id("Bip01 Spine"), 0.055*intensity*cs, 'z')
	}
	s.resolve()
}

func main() {
	dir := flag.String("char", filepath.Join("web", "data", "char", "norseman"), "exported character directory")
	out := flag.String("out", "char.png", "output PNG")
	size := flag.Int("size", 520, "image height in pixels")
	phase := flag.Float64("phase", 0, "gait phase in radians")
	intensity := flag.Float64("gait", 0, "gait intensity, 0 for the bind pose")
	yaw := flag.Float64("yaw", 215, "camera azimuth in degrees")
	elev := flag.Float64("elev", 8, "camera elevation in degrees")
	zoom := flag.Float64("zoom", 1, "magnification; 3 frames roughly a hand and what it holds")
	look := flag.String("at", "", "with -zoom, centre on this bone instead of the figure")
	flag.Parse()

	if err := run(*dir, *out, *size, *phase, *intensity, *yaw, *elev, *zoom, *look); err != nil {
		fmt.Fprintln(os.Stderr, "charshot:", err)
		os.Exit(1)
	}
}

// model is one converted directory: a character, or an item it holds.
type model struct {
	man      manifest
	blob     []byte
	textures map[string]image.Image
}

func (m *model) f32(off int) float64 {
	return float64(math.Float32frombits(binary.LittleEndian.Uint32(m.blob[off:])))
}

func (m *model) index(i int) int {
	return int(binary.LittleEndian.Uint32(m.blob[m.man.VertexCount*m.man.Stride+i*4:]))
}

func loadModel(dir string) (*model, error) {
	jsRaw, err := os.ReadFile(filepath.Join(dir, "char.json"))
	if err != nil {
		return nil, err
	}
	m := &model{textures: map[string]image.Image{}}
	if err := json.Unmarshal(jsRaw, &m.man); err != nil {
		return nil, err
	}
	if m.blob, err = os.ReadFile(filepath.Join(dir, "mesh.bin")); err != nil {
		return nil, err
	}
	vertBytes := m.man.VertexCount * m.man.Stride
	if len(m.blob) != vertBytes+m.man.IndexCount*4 {
		return nil, fmt.Errorf("mesh.bin is %d bytes, manifest implies %d",
			len(m.blob), vertBytes+m.man.IndexCount*4)
	}
	for _, s := range m.man.Shapes {
		if s.Texture == "" || m.textures[s.Texture] != nil {
			continue
		}
		f, err := os.Open(filepath.Join(dir, "tex", s.Texture))
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s.Texture, err)
		}
		m.textures[s.Texture] = img
	}
	return m, nil
}

func run(dir, out string, size int, phase, intensity, yawDeg, elevDeg, zoom float64, lookAt string) error {
	ch, err := loadModel(dir)
	if err != nil {
		return err
	}
	man := ch.man

	skel, err := newSkeleton(man.Bones)
	if err != nil {
		return err
	}
	skel.poseWalk(phase, intensity)

	// Held items: each is a one-bone model whose bone is a socket on this
	// skeleton, so it is drawn with that socket's world transform.
	type held struct {
		m      *model
		socket int
	}
	var items []held
	for _, e := range man.Equip {
		m, err := loadModel(filepath.Join(dir, "equip", e.Slot))
		if err != nil {
			return fmt.Errorf("equip %s: %w", e.Slot, err)
		}
		s := skel.id(e.Bone)
		if s < 0 {
			return fmt.Errorf("equip %s: no socket %q", e.Slot, e.Bone)
		}
		items = append(items, held{m, s})
	}

	// Camera: orbit the figure's mid-height at a distance that frames it.
	h := man.Max[2] - man.Min[2]
	target := [3]float64{0, 0, man.Min[2] + h*0.52}
	if lookAt != "" {
		b := skel.id(lookAt)
		if b < 0 {
			return fmt.Errorf("no bone %q", lookAt)
		}
		target = skel.world[b].point([3]float64{})
	}
	dist := h * 2.1
	ya, el := yawDeg*math.Pi/180, elevDeg*math.Pi/180
	eye := [3]float64{
		target[0] + dist*math.Cos(el)*math.Cos(ya),
		target[1] + dist*math.Cos(el)*math.Sin(ya),
		target[2] + dist*math.Sin(el),
	}
	fwd := norm(sub(target, eye))
	right := norm(cross(fwd, [3]float64{0, 0, 1}))
	up := cross(right, fwd)

	W, H := size*3/4, size
	focal := float64(H) * 0.62 * dist / h * h / dist // keeps the figure framed
	focal = float64(H) * 1.05 * zoom
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	depth := make([]float64, W*H)
	for i := range depth {
		depth[i] = math.Inf(1)
	}
	// Background: a flat slate so the silhouette is unambiguous.
	for i := 0; i < W*H; i++ {
		img.Pix[i*4+0] = 24
		img.Pix[i*4+1] = 30
		img.Pix[i*4+2] = 38
		img.Pix[i*4+3] = 255
	}

	type pv struct {
		sx, sy, z float64
		u, v      float64
		shade     float64
		ok        bool
	}
	project := func(p, n [3]float64) pv {
		d := sub(p, eye)
		z := dot(d, fwd)
		if z < 1e-3 {
			return pv{}
		}
		light := norm([3]float64{0.4, -0.75, 0.5})
		sh := math.Abs(dot(norm(n), light))
		return pv{
			sx:    float64(W)/2 + dot(d, right)/z*focal,
			sy:    float64(H)/2 - dot(d, up)/z*focal,
			z:     z,
			shade: 0.42 + 0.68*sh,
			ok:    true,
		}
	}

	drawn, clipped := 0, 0
	// draw rasterises one model; boneWorld gives the world transform of
	// each of its bones.
	draw := func(md *model, boneWorld func(int) xf) {
		for _, s := range md.man.Shapes {
			// Precompute this shape's skinning matrices.
			mats := make([]xf, len(s.Bones))
			for b := range s.Bones {
				mats[b] = mul(boneWorld(s.Bones[b]), xf(s.InvBind[b]))
			}
			tex := md.textures[s.Texture]
			f32, index, blob := md.f32, md.index, md.blob
			skin := func(vi int) ([3]float64, [3]float64, [2]float64) {
				o := vi * md.man.Stride
				p := [3]float64{f32(o + offPos), f32(o + offPos + 4), f32(o + offPos + 8)}
				n := [3]float64{f32(o + offNormal), f32(o + offNormal + 4), f32(o + offNormal + 8)}
				uv := [2]float64{f32(o + offUV), f32(o + offUV + 4)}
				var sp, sn [3]float64
				for k := 0; k < 4; k++ {
					w := f32(o + offWeight + k*4)
					if w <= 0 {
						continue
					}
					m := mats[blob[o+offBone+k]]
					tp, tn := m.point(p), m.dir(n)
					for c := 0; c < 3; c++ {
						sp[c] += w * tp[c]
						sn[c] += w * tn[c]
					}
				}
				return sp, sn, uv
			}

			for i := s.First; i < s.First+s.Count; i += 3 {
				var tri [3]pv
				bad := false
				for k := 0; k < 3; k++ {
					p, n, uv := skin(index(i + k))
					v := project(p, n)
					v.u, v.v = uv[0], uv[1]
					if !v.ok {
						bad = true
					}
					tri[k] = v
				}
				if bad {
					clipped++
					continue
				}
				rasterise(img, depth, W, H, tri[0], tri[1], tri[2], tex, s.Diffuse)
				drawn++
			}
		}
	}
	draw(ch, func(b int) xf { return skel.world[b] })
	for _, it := range items {
		draw(it.m, func(int) xf { return skel.world[it.socket] })
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Printf("%s: %d triangles (%d clipped), %d bones, pose phase=%.2f gait=%.2f\n",
		man.Source, drawn, clipped, len(man.Bones), phase, intensity)
	fmt.Printf("wrote %s (%dx%d)\n", out, W, H)
	return nil
}

// rasterise fills one triangle with a z-buffer, interpolating UV and shade.
func rasterise(img *image.RGBA, depth []float64, W, H int, a, b, c struct {
	sx, sy, z float64
	u, v      float64
	shade     float64
	ok        bool
}, tex image.Image, diffuse [3]float64) {
	minX := int(math.Floor(math.Min(a.sx, math.Min(b.sx, c.sx))))
	maxX := int(math.Ceil(math.Max(a.sx, math.Max(b.sx, c.sx))))
	minY := int(math.Floor(math.Min(a.sy, math.Min(b.sy, c.sy))))
	maxY := int(math.Ceil(math.Max(a.sy, math.Max(b.sy, c.sy))))
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= W {
		maxX = W - 1
	}
	if maxY >= H {
		maxY = H - 1
	}
	area := (b.sx-a.sx)*(c.sy-a.sy) - (b.sy-a.sy)*(c.sx-a.sx)
	if math.Abs(area) < 1e-9 {
		return
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			w0 := ((b.sx-px)*(c.sy-py) - (b.sy-py)*(c.sx-px)) / area
			w1 := ((c.sx-px)*(a.sy-py) - (c.sy-py)*(a.sx-px)) / area
			w2 := 1 - w0 - w1
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			z := w0*a.z + w1*b.z + w2*c.z
			di := y*W + x
			if z >= depth[di] {
				continue
			}
			depth[di] = z
			u := w0*a.u + w1*b.u + w2*c.u
			v := w0*a.v + w1*b.v + w2*c.v
			shade := w0*a.shade + w1*b.shade + w2*c.shade
			r, g, bl := diffuse[0], diffuse[1], diffuse[2]
			if tex != nil {
				tb := tex.Bounds()
				tx := tb.Min.X + int(wrap(u)*float64(tb.Dx()))
				ty := tb.Min.Y + int(wrap(v)*float64(tb.Dy()))
				if tx >= tb.Max.X {
					tx = tb.Max.X - 1
				}
				if ty >= tb.Max.Y {
					ty = tb.Max.Y - 1
				}
				cr, cg, cb, _ := tex.At(tx, ty).RGBA()
				r *= float64(cr) / 65535
				g *= float64(cg) / 65535
				bl *= float64(cb) / 65535
			}
			o := img.PixOffset(x, y)
			img.Pix[o+0] = clamp8(r * shade * 255)
			img.Pix[o+1] = clamp8(g * shade * 255)
			img.Pix[o+2] = clamp8(bl * shade * 255)
			img.Pix[o+3] = 255
		}
	}
}

func wrap(t float64) float64 {
	t = math.Mod(t, 1)
	if t < 0 {
		t += 1
	}
	return t
}

func clamp8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(v)
}

func sub(a, b [3]float64) [3]float64 { return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func dot(a, b [3]float64) float64    { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }

func cross(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func norm(a [3]float64) [3]float64 {
	l := math.Sqrt(dot(a, a))
	if l == 0 {
		return a
	}
	return [3]float64{a[0] / l, a[1] / l, a[2] / l}
}

var _ = color.RGBA{}
var _ = identity
