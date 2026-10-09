// Package scenery bakes scenery models -- the trees, fences, buildings and
// other props a zone places -- into the flat, textured geometry the viewer
// draws.
//
// Each model is flattened at conversion time: the scene graph is walked, world
// transforms are baked into the vertices, and the result is grouped by
// texture. Props never move, so there is nothing for the viewer to re-pose.
package scenery

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"daocweb/internal/dds"
	"daocweb/internal/mpak"
	"daocweb/internal/nif"
)

// Stride is pos[3] + normal[3] + uv[2], all float32, then the baked
// shade as four bytes, RGBA.
const Stride = 36

// Group is one draw: a run of the index buffer sharing a texture and state.
type Group struct {
	Texture string `json:"texture"`
	Alpha   bool   `json:"alpha"`
	First   int    `json:"first"` // index of the first element in the index buffer
	Count   int    `json:"count"`
	// Additive groups are glows -- flames, coronas -- added to what is
	// behind them, black meaning nothing.
	Additive bool `json:"additive,omitempty"`
	// Billboard groups turn to face the camera. Each vertex's position is
	// then the card's pivot, and its normal holds the corner's offset from
	// it: x across the card, y up it.
	Billboard bool `json:"billboard,omitempty"`
	// Scroll is a texture moving across its card, in texture widths a
	// second, U then V: the flames rise at 1 a second. From a
	// NiUVController, whose curves in these models are all straight ramps.
	Scroll *[2]float32 `json:"scroll,omitempty"`
	// Pulse is a billboard's size over one second, [time, scale] keys,
	// looped: the coronas breathe between 0.91 and 1. From the
	// NiKeyframeController on the card's NiBillboardNode.
	Pulse [][2]float32 `json:"pulse,omitempty"`
}

// Model is a baked model: its bounds, its draws and its emitters.
type Model struct {
	Name   string     `json:"name"`
	File   string     `json:"file"`
	Min    [3]float32 `json:"min"`
	Max    [3]float32 `json:"max"`
	Groups []Group    `json:"groups"`
	// Particles are the model's emitters; see particles.go.
	Particles []Particle `json:"particles,omitempty"`
	// Embedded are the textures the model carries inside itself, by the
	// name its groups and emitters use; see embedded.go.
	Embedded map[string]image.Image `json:"-"`
}

// Vertex is one baked vertex, Stride bytes when written.
type Vertex struct {
	P  [3]float32
	N  [3]float32
	UV [2]float32
	C  [4]uint8 // baked shade; white where a model has none
}

// --- model conversion ----------------------------------------------------

// texState is the render state a node inherits from its ancestors. NIF hangs
// properties off any NiAVObject and children inherit them, so a texture set on
// the model root applies to every shape beneath it.
type texState struct {
	texture  string
	alpha    bool
	additive bool
	// billboard is set under a NiBillboardNode, with the node's origin in
	// model space as the point the card turns about.
	billboard bool
	pivot     [3]float32
	pulse     [][2]float32
	// vertexMode is the NiVertexColorProperty in force: 0 ignores a shape's
	// Vertex colours, 1 makes them emissive, 2 has them scale the lighting.
	// Only 2 is used, as baked shade -- the darker eaves, corners and
	// interiors the buildings were lit with when they were made. Shapes set
	// to 0 do carry colours, and some of those are nonsense: the Jordheim
	// gate's run to -2885, a modelling channel exported by mistake, which
	// the game never reads because of that 0.
	vertexMode uint32
}

// groupKey is what splits a model's geometry into draws.
type groupKey struct {
	texture             string
	additive, billboard bool
	scroll              [2]float32
	pulse               string // the pulse keys, printed, so equal curves share a group
}

// ModelDirs is where a zone's scenery models live, in search order. Most
// zones draw only on zones/Nifs. The frontier zones add frontiers/NIFS,
// which holds their keeps and the like, and still take most of their trees
// and rocks from zones/Nifs.
func ModelDirs(game string, frontier bool) []string {
	z := filepath.Join(game, "zones", "Nifs")
	if frontier {
		return []string{filepath.Join(game, "frontiers", "NIFS"), z}
	}
	return []string{z}
}

// ConvertModel reads a scenery model from the first of dirs that holds it
// and bakes it into verts and indices; see Bake.
func ConvertModel(dirs []string, name, file string, verts *[]Vertex, indices *[]uint32) (*Model, error) {
	raw, err := FindModel(dirs, file)
	if err != nil {
		return nil, err
	}
	f, err := nif.Parse(raw)
	if err != nil {
		return nil, err
	}
	return Bake(f, name, file, verts, indices)
}

// Bake flattens a parsed model, appending its vertices and indices to the
// shared buffers, and describes the draws that address them.
func Bake(f *nif.File, name, file string, verts *[]Vertex, indices *[]uint32) (*Model, error) {
	mo := &Model{Name: name, File: file}
	tn := newTexNamer(f, file)

	// Animation hangs off its target by reference from the controller's
	// side, so index the controllers by what they drive.
	pulses := map[string][][2]float32{}
	uvBy := map[int32]*nif.UVController{}
	kfBy := map[int32]*nif.KeyframeController{}
	for _, blk := range f.Blocks {
		switch c := blk.(type) {
		case *nif.UVController:
			uvBy[c.Target] = c
		case *nif.KeyframeController:
			kfBy[c.Target] = c
		}
	}

	// Collect geometry per texture so each group is one draw call.
	byTex := map[groupKey]*struct {
		alpha bool
		idx   []uint32
	}{}
	min := [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
	max := [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
	base := uint32(len(*verts))
	_ = base

	var walk func(ref int32, parent nif.Transform, st texState, depth int)
	walk = func(ref int32, parent nif.Transform, st texState, depth int) {
		if depth > 64 {
			return
		}
		blk := f.Block(ref)
		if blk == nil {
			return
		}
		switch b := blk.(type) {
		case *nif.Node:
			if isCollision(b.Name) {
				return
			}
			world := parent.Mul(b.Local)
			st = applyProps(tn, b.Properties, st)
			if f.TypeOf[ref] == "NiBillboardNode" {
				st.billboard, st.pivot, st.pulse = true, world.Trans, nil
				if kf := kfBy[ref]; kf != nil {
					if d, ok := f.Block(kf.Data).(*nif.KeyframeData); ok && len(d.Scales) > 1 {
						for _, k := range d.Scales {
							// Rounded, so copies of one curve share a draw group.
							r := func(v float32) float32 { return float32(math.Round(float64(v)*1000) / 1000) }
							st.pulse = append(st.pulse, [2]float32{r(k.Time), r(k.Value)})
						}
					}
				}
			}
			for _, c := range b.Children {
				walk(c, world, st, depth+1)
			}

		case *nif.LODNode:
			if isCollision(b.Name) {
				return
			}
			world := parent.Mul(b.Local)
			st = applyProps(tn, b.Properties, st)
			// Only the nearest level is kept: props are seen from the
			// ground, and drawing every level would stack them.
			if len(b.Children) > 0 {
				walk(b.Children[0], world, st, depth+1)
			}

		case *nif.TriShape:
			if isCollision(b.Name) {
				return
			}
			world := parent.Mul(b.Local)
			st = applyProps(tn, b.Properties, st)
			d, _ := f.Block(b.Data).(*nif.ShapeData)
			if d == nil || len(d.Vertices) == 0 || len(d.Triangles) == 0 {
				return
			}
			k := groupKey{texture: st.texture, additive: st.additive, billboard: st.billboard}
			if uv := uvBy[ref]; uv != nil {
				k.scroll = uvScroll(f, uv)
			}
			if st.billboard && len(st.pulse) > 1 {
				k.pulse = fmt.Sprint(st.pulse)
				pulses[k.pulse] = st.pulse
			}
			g := byTex[k]
			if g == nil {
				g = &struct {
					alpha bool
					idx   []uint32
				}{}
				byTex[k] = g
			}
			g.alpha = g.alpha || st.alpha

			first := uint32(len(*verts))
			var card cardAxes
			if st.billboard {
				card = cardFrame(d.Vertices, world, st.pivot)
			}
			for i, p := range d.Vertices {
				v := Vertex{P: world.Apply(p), C: [4]uint8{255, 255, 255, 255}}
				if st.vertexMode == 2 && i < len(d.Colors) {
					for k := 0; k < 4; k++ {
						v.C[k] = uint8(math.Round(255 * math.Max(0, math.Min(1, float64(d.Colors[i][k])))))
					}
				}
				if i < len(d.Normals) {
					v.N = rotate(world, d.Normals[i])
				}
				if st.billboard {
					// The bounds take the corner where the exporter left
					// it; the Vertex itself becomes the pivot plus an offset.
					for k := 0; k < 3; k++ {
						min[k], max[k] = minf(min[k], v.P[k]), maxf(max[k], v.P[k])
					}
					v.N = card.offset(v.P, st.pivot)
					v.P = st.pivot
				}
				if len(d.UV) > 0 && i < len(d.UV[0]) {
					v.UV = d.UV[0][i]
				}
				for k := 0; k < 3; k++ {
					if v.P[k] < min[k] {
						min[k] = v.P[k]
					}
					if v.P[k] > max[k] {
						max[k] = v.P[k]
					}
				}
				*verts = append(*verts, v)
			}
			for _, t := range d.Triangles {
				g.idx = append(g.idx, first+uint32(t[0]), first+uint32(t[1]), first+uint32(t[2]))
			}
		}
	}
	for _, root := range f.Roots {
		walk(root, nif.Identity, texState{}, 0)
	}

	keys := make([]groupKey, 0, len(byTex))
	for k := range byTex {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.texture != b.texture {
			return a.texture < b.texture
		}
		if a.additive != b.additive {
			return !a.additive
		}
		if a.billboard != b.billboard {
			return !a.billboard
		}
		if a.scroll != b.scroll {
			return a.scroll[0] < b.scroll[0] || (a.scroll[0] == b.scroll[0] && a.scroll[1] < b.scroll[1])
		}
		return a.pulse < b.pulse
	})
	for _, k := range keys {
		g := byTex[k]
		if len(g.idx) == 0 {
			continue
		}
		mo.Groups = append(mo.Groups, Group{
			Texture:   k.texture,
			Alpha:     g.alpha,
			First:     len(*indices),
			Count:     len(g.idx),
			Additive:  k.additive,
			Billboard: k.billboard,
			Pulse:     pulses[k.pulse],
		})
		if k.scroll != ([2]float32{}) {
			s := k.scroll
			mo.Groups[len(mo.Groups)-1].Scroll = &s
		}
		*indices = append(*indices, g.idx...)
	}
	mo.Particles = particleSystems(tn)
	mo.Embedded = tn.embedded
	if len(mo.Groups) == 0 {
		// Some models are nothing but emitters -- the wandering mists,
		// the labyrinth fires. They draw, with no geometry and bounds
		// around their emitters.
		if len(mo.Particles) == 0 {
			return nil, fmt.Errorf("no drawable geometry")
		}
		mo.Groups = []Group{}
		for _, p := range mo.Particles {
			for k := 0; k < 3; k++ {
				min[k], max[k] = minf(min[k], p.Origin[k]-p.Size), maxf(max[k], p.Origin[k]+p.Size)
			}
		}
	}
	mo.Min, mo.Max = min, max
	return mo, nil
}

// isCollision spots the invisible hulls the exporter ships beside the
// visible mesh. The fence models are the clearest case: a "Collisionswitch"
// node with a "collidee" branch and a "visible" one. The newer models add a
// third, "shadowcaster": a coarse, untextured stand-in the game renders only
// into its shadow pass. Drawn, it is a white lump over the real thing -- the
// boulders and the Mularn hall base wore one.
func isCollision(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "collidee") || strings.Contains(n, "collision_") ||
		strings.HasPrefix(n, "collide") || n == "shadowcaster"
}

// applyProps resolves a node's property list, inheriting anything it does not
// override from its parent.
func applyProps(tn *texNamer, refs []int32, st texState) texState {
	f := tn.f
	for _, p := range refs {
		switch b := f.Block(p).(type) {
		case *nif.Texturing:
			if !b.HasBase {
				continue
			}
			if src, ok := f.Block(b.Base.Source).(*nif.SourceTexture); ok {
				if n := tn.name(src); n != "" {
					st.texture = n
				}
			}
		case *nif.VertexColor:
			st.vertexMode = b.VertexMode
		case *nif.AlphaProperty:
			// Bit 9 of the flags is the alpha-test enable. Trees need it:
			// the canopy is a few quads with a cut-out leaf texture.
			st.alpha = true
			// Bit 0 enables blending; bits 1-4 and 5-8 are the source and
			// destination factors. A destination of 0, ONE, adds the
			// texture to what is behind: the flames and glows, whose
			// textures are drawn on black.
			st.additive = b.Flags&1 != 0 && (b.Flags>>5)&15 == 0
		}
	}
	return st
}

// TexName is how a texture is known here: its file's base name, lower-cased,
// without the extension.
func TexName(path string) string {
	b := strings.ToLower(filepath.Base(strings.ReplaceAll(path, `\`, "/")))
	return strings.TrimSuffix(b, filepath.Ext(b))
}

// rotate applies only a transform's rotation, which is what a normal wants.
// The scales here are uniform, so renormalising is enough to undo it.
func rotate(t nif.Transform, v [3]float32) [3]float32 {
	var o [3]float32
	for r := 0; r < 3; r++ {
		o[r] = t.Rot[r*3+0]*v[0] + t.Rot[r*3+1]*v[1] + t.Rot[r*3+2]*v[2]
	}
	n := float32(math.Sqrt(float64(o[0]*o[0] + o[1]*o[1] + o[2]*o[2])))
	if n > 1e-6 {
		o[0], o[1], o[2] = o[0]/n, o[1]/n, o[2]/n
	}
	return o
}

// LoadModel finds a scenery model in dir. They live one per .npk,
// occasionally loose beside them.
func LoadModel(dir, file string) ([]byte, error) {
	base := strings.TrimSuffix(file, filepath.Ext(file))
	pk := filepath.Join(dir, base+".npk")
	if a, err := mpak.Open(pk); err == nil {
		for _, e := range a.Entries {
			if strings.EqualFold(filepath.Ext(e.Name), ".nif") {
				return a.ReadEntry(e)
			}
		}
		return nil, fmt.Errorf("%s holds no .nif", filepath.Base(pk))
	}
	if b, err := os.ReadFile(filepath.Join(dir, file)); err == nil {
		return b, nil
	}
	return nil, fmt.Errorf("%w %s.npk", ErrNoModel, base)
}

// FindModel is LoadModel over several directories: the first that holds
// the model wins, and if none does the error is the first directory's.
func FindModel(dirs []string, file string) ([]byte, error) {
	var first error
	for _, d := range dirs {
		b, err := LoadModel(d, file)
		if err == nil {
			return b, nil
		}
		if first == nil {
			first = err
		}
	}
	return nil, first
}

// ReadTexture decodes a model's texture by name, which sits loose beside
// the models as a .dds; dirs are searched in order.
func ReadTexture(dirs []string, name string) (image.Image, error) {
	var path string
	for _, d := range dirs {
		if p, err := findCase(d, name+".dds"); err == nil {
			path = p
			break
		}
	}
	if path == "" {
		return nil, fmt.Errorf("%w: %s.dds", ErrNoTexture, name)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return dds.Decode(b)
}

// ErrNoModel is a model found in none of the directories searched.
var ErrNoModel = errors.New("no model archive")

// ErrNoTexture is a texture found in none of the directories searched.
var ErrNoTexture = errors.New("texture not found")

// ExtractTextures writes the textures a zone's models use as PNGs into dir.
func ExtractTextures(dirs []string, dir string, want map[string]bool) error {
	if len(want) == 0 {
		return nil
	}
	names := make([]string, 0, len(want))
	for k := range want {
		names = append(names, k)
	}
	sort.Strings(names)

	var missing []string
	for _, n := range names {
		img, err := ReadTexture(dirs, n)
		if errors.Is(err, ErrNoTexture) || errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, n)
			continue
		}
		if err != nil {
			return fmt.Errorf("%s: %w", n, err)
		}
		if err := WritePNG(filepath.Join(dir, n+".png"), img); err != nil {
			return err
		}
	}
	if len(missing) > 0 {
		fmt.Printf("  %d textures not found: %s\n", len(missing), strings.Join(missing, ", "))
	}
	fmt.Printf("  %d of %d textures written\n", len(names)-len(missing), len(names))
	return nil
}

// findCase resolves a name case-insensitively; the CSVs and the files on disk
// disagree about capitalisation.
func findCase(dir, name string) (string, error) {
	p := filepath.Join(dir, name)
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range ents {
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("%s not found in %s", name, dir)
}

// WritePNG writes an image as a PNG file.
func WritePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// cardAxes is how a billboard card's corners are re-expressed for drawing:
// across and up the card, so the viewer can lay them out along the camera's
// right and up instead of wherever the exporter left them facing.
type cardAxes struct {
	across [2]float32 // horizontal direction across an upright card
	flat   bool       // a card lying flat: both axes are horizontal
}

// cardFrame finds a card's axes from its corners. An upright card -- a
// flame -- spans height, and across is the direction its corners spread in
// the ground plane. A flat one -- some coronas -- spans no height, and is
// laid out by its two ground-plane directions instead.
func cardFrame(ps [][3]float32, world nif.Transform, pivot [3]float32) cardAxes {
	var sxx, sxy, syy, zlo, zhi, ext float64
	zlo, zhi = math.Inf(1), math.Inf(-1)
	for _, p := range ps {
		w := world.Apply(p)
		x, y, z := float64(w[0]-pivot[0]), float64(w[1]-pivot[1]), float64(w[2]-pivot[2])
		sxx += x * x
		sxy += x * y
		syy += y * y
		zlo, zhi = math.Min(zlo, z), math.Max(zhi, z)
		ext = math.Max(ext, math.Hypot(x, y))
	}
	// Principal axis of the ground-plane spread.
	a := 0.5 * math.Atan2(2*sxy, sxx-syy)
	c := cardAxes{across: [2]float32{float32(math.Cos(a)), float32(math.Sin(a))}}
	c.flat = zhi-zlo < 0.2*ext
	return c
}

func (c cardAxes) offset(p, pivot [3]float32) [3]float32 {
	x, y, z := p[0]-pivot[0], p[1]-pivot[1], p[2]-pivot[2]
	u := x*c.across[0] + y*c.across[1]
	if c.flat {
		return [3]float32{u, -x*c.across[1] + y*c.across[0], 0}
	}
	return [3]float32{u, z, 0}
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// uvScroll turns a NiUVController's offset curves into a rate, texture
// widths a second, U then V. Every curve in the zone's models is a two-key
// ramp -- the flames' V from 0 to 1 over a second, the bindstone's U over
// ten -- so first to last key is the whole of it.
func uvScroll(f *nif.File, c *nif.UVController) [2]float32 {
	d, ok := f.Block(c.Data).(*nif.UVData)
	if !ok {
		return [2]float32{}
	}
	var out [2]float32
	for i := 0; i < 2; i++ {
		k := d.Groups[i]
		if len(k) < 2 || k[len(k)-1].Time <= k[0].Time {
			continue
		}
		out[i] = (k[len(k)-1].Value - k[0].Value) / (k[len(k)-1].Time - k[0].Time)
	}
	return out
}
