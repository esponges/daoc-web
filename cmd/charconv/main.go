// Command charconv extracts one playable character from DAoC's figures/ and
// writes it in a form a WebGL viewer can skin on the GPU.
//
// A race model is not a single mesh. NVikingM.NIF holds 49 skinned shapes:
// ten heads, five torso tiers each in two cuts, several arms, legs, boots,
// gloves and cloaks, plus whole-body meshes for distant rendering. Drawing all
// of them at once would stack every outfit the race can wear on top of itself,
// so a character is assembled by choosing one shape per slot.
//
// Textures are not referenced from the model at all: the shapes carry UVs but
// no NiTexturingProperty, because the game assigns skins at runtime from
// equipment. The mapping here is by slot, following the naming convention in
// figures/skins and figures/Mskins:
//
//	Body1 -> m1_bodym.dds      Arms1  -> m1_arm1.dds
//	LBody1 -> m1_tunic1.dds    Legs1  -> m1_legs1.dds
//	HeadA1 -> nor_m_head01.dds Boots1 -> m1_boot1.dds
//
// Output is <out>/<name>/: mesh.bin (interleaved vertices then indices),
// char.json (skeleton, per-shape draw ranges and inverse bind matrices) and
// tex/*.png.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"daocweb/internal/dds"
	"daocweb/internal/gamedata"
	"daocweb/internal/mpak"
	"daocweb/internal/nif"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

// An outfit is one shape per slot plus the texture to dress it with. The two
// go together: the shape tier decides the silhouette and the texture decides
// what it is made of, and picking them independently gives you plate boots
// painted like cloth.
//
// A model carries its slots in tiers -- Body1..Body5, Boots1..Boots5 -- and
// those tiers are the game's armour categories in order: cloth, leather,
// studded, chain, plate. The heaviest tier a slot has is its plate variant,
// which is why the plate outfit below reaches for the last of each.
type outfit struct {
	desc     string
	figure   string
	shapes   string
	model    int               // monsters.csv row; the race's scale is read from it
	textures map[string]string // shape-name prefix -> .dds; longest prefix wins
	equip    []equip           // items held on the skeleton's sockets
}

var outfits = map[string]outfit{
	// Head A1 is the highest-detail face; the cloak and LOD bodies are
	// deliberately left out of both.
	"norseman": {
		desc:   "Norseman in starter cloth",
		figure: "NVikingM.NIF",
		model:  153, // "Norse Male", drawn at 1.09
		shapes: "HeadA1,Body1,LBody1,Arms1,Legs1,Gloves1,Boots1",
		textures: map[string]string{
			"Head":   "nor_m_head01.dds",
			"Body":   "m1_bodym.dds",
			"LBody":  "m1_tunic1.dds",
			"Arms":   "m1_arm1.dds",
			"Legs":   "m1_legs1.dds",
			"Gloves": "m1_glove1.dds",
			"Boots":  "m1_boot1.dds",
		},
	},
	// Plate is Albion's armour in the live game and a Troll is a Midgard
	// race, so this is a combination the client would never assemble. The
	// meshes and textures are independent of each other, though, so nothing
	// stops it here.
	"troll-plate": {
		desc:   "Troll in full plate",
		figure: "NTrollM.NIF",
		model:  137, // "Troll Male", drawn at 1.30
		shapes: "HeadA1,Body5,Lbody4,Arms6,Legs4,Gloves3,Boots5",
		textures: map[string]string{
			"Head":   "tro_m_Head01.dds",
			"Body":   "pltBody01_04_m.dds",
			"LBody":  "pltBody01_04_m.dds",
			"Arms":   "pltBody01_04_m.dds",
			"Legs":   "pltLegs01_04.dds",
			"Gloves": "pltGloves01_04.dds",
			"Boots":  "pltBoots01_04.dds",
		},
	},
	// A Midgard warrior as the realm would kit one out. Chain is Midgard's
	// heaviest armour -- pskins.csv's realm prefixes run Albion to plate,
	// Hibernia to scale and Midgard to chain -- and a hammer and shield are
	// a warrior's own weapons. The chain is the tier below plate in every
	// slot, dressed in the one chn* texture family that covers them all.
	"troll-warrior": {
		desc:   "Troll warrior in chain, with hammer and shield",
		figure: "NTrollM.NIF",
		model:  137,
		shapes: "HeadA1,Body4,LBody3,Arms4,Legs3,Gloves2,Boots4",
		textures: map[string]string{
			"Head":   "tro_m_Head01.dds",
			"Body":   "chnbody05_10_m.dds",
			"LBody":  "chnbody05_10_m.dds",
			"Arms":   "chnarms05_10.dds",
			"Legs":   "chnlegs05_10.dds",
			"Gloves": "chngloves05_10.dds",
			"Boots":  "chnboots05_10.dds",
		},
		equip: []equip{
			{Slot: "right", Bone: "Bip01 R Held", Item: "N_1h_warhammer01", Name: "norse warhammer"},
			{Slot: "left", Bone: "Bip01 L Shield", Item: "M_Shield_Crescent", Name: "Mid Shield 01"},
		},
	},
}

// slotTextures is the active outfit's mapping, set from the chosen outfit and
// any -tex overrides.
var slotTextures = map[string]string{}

// maxInfluences is how many bones may move one vertex. Four is what the file
// uses in practice and what fits a compact vertex.
const maxInfluences = 4

// Vertex layout in mesh.bin, in bytes.
const (
	offPos     = 0
	offNormal  = 12
	offUV      = 24
	offBone    = 32
	offWeight  = 36
	vertStride = 52
)

type boneOut struct {
	Name   string     `json:"name"`
	Parent int        `json:"parent"`
	T      [3]float32 `json:"t"`
	R      [9]float32 `json:"r"`
	S      float32    `json:"s"`
}

type shapeOut struct {
	Name      string        `json:"name"`
	First     int           `json:"first"`
	Count     int           `json:"count"`
	Bones     []int         `json:"bones"`   // indices into the bone array
	InvBind   [][12]float32 `json:"invBind"` // per bone, rows of 4
	Texture   string        `json:"texture"`
	Diffuse   [3]float32    `json:"diffuse"`
	Emissive  [3]float32    `json:"emissive"`
	Alpha     float32       `json:"alpha"`
	HasColors bool          `json:"hasColors"`
}

type manifest struct {
	Name        string     `json:"name"`
	Source      string     `json:"source"`
	Format      string     `json:"format"`
	VertexCount int        `json:"vertexCount"`
	IndexCount  int        `json:"indexCount"`
	Stride      int        `json:"stride"`
	Height      float32    `json:"height"`
	Min         [3]float32 `json:"min"`
	Max         [3]float32 `json:"max"`
	UpAxis      string     `json:"upAxis"`
	ForwardY    int        `json:"forwardY"`
	// From the game tables when the model has a row there. Scale is how
	// much larger than authored the game draws it; AnimSet picks its clips.
	Title   string  `json:"title,omitempty"` // display name: "Large Grey Wolf"
	Model   int     `json:"model,omitempty"`
	Scale   float64 `json:"scale"`
	AnimSet int     `json:"animSet,omitempty"`
	// Items held on sockets, each converted beside the character into
	// equip/<slot>/, itself a one-bone character.
	Equip  []equip    `json:"equip,omitempty"`
	Bones  []boneOut  `json:"bones"`
	Shapes []shapeOut `json:"shapes"`
}

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	outfitName := flag.String("outfit", "norseman", "named outfit: "+outfitNames())
	figure := flag.String("figure", "", "model in figures/ (overrides the outfit's)")
	name := flag.String("name", "", "output name (defaults to the outfit name)")
	out := flag.String("out", filepath.Join("web", "data", "char"), "output directory")
	shapeList := flag.String("shapes", "", "comma-separated shapes (overrides the outfit's)")
	texList := flag.String("tex", "", "per-slot texture overrides, e.g. Body=pltBody01_01_m.dds,Legs=...")
	listOnly := flag.Bool("list", false, "list available shapes and exit")
	noTex := flag.Bool("no-tex", false, "skip texture extraction")
	model := flag.Int("model", 0, "convert a creature by its monsters.csv model ID instead of an outfit")
	flag.Parse()

	fail := func(err error) {
		fmt.Fprintln(os.Stderr, "charconv:", err)
		os.Exit(1)
	}
	var o outfit
	if *model > 0 {
		var err error
		if o, err = fromTables(*game, *model); err != nil {
			fail(err)
		}
		if *name == "" {
			*name = slug(o.desc)
		}
	} else {
		var ok bool
		if o, ok = outfits[*outfitName]; !ok {
			fmt.Fprintf(os.Stderr, "charconv: unknown outfit %q; have %s\n", *outfitName, outfitNames())
			os.Exit(1)
		}
		if *name == "" {
			*name = *outfitName
		}
		// An outfit names its race's row only for the scale and clips;
		// its shapes and textures stay the hand-picked ones.
		if o.model > 0 {
			t, err := gamedata.Load(*game)
			if err != nil {
				fail(err)
			}
			m := t.Monsters[o.model]
			meta.model, meta.scale, meta.animSet = m.ID, m.Scale, t.Figures[m.Figure].AnimSet
		}
	}
	if *figure == "" {
		*figure = o.figure
	}
	if *shapeList == "" {
		*shapeList = o.shapes
	}
	meta.equip = o.equip
	for k, v := range o.textures {
		slotTextures[k] = v
	}
	if *texList != "" {
		for _, pair := range strings.Split(*texList, ",") {
			k, v, found := strings.Cut(strings.TrimSpace(pair), "=")
			if !found {
				fmt.Fprintf(os.Stderr, "charconv: -tex wants Slot=file.dds, got %q\n", pair)
				os.Exit(1)
			}
			slotTextures[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	fmt.Printf("%s: %s\n", *name, o.desc)

	if err := run(*game, *figure, *name, *out, *shapeList, *listOnly, *noTex); err != nil {
		fail(err)
	}
}

// meta is what the game tables say about the model being converted, carried
// into char.json for the viewer and animconv.
var meta = struct {
	model, animSet int
	title          string
	equip          []equip
	scale          float64
}{scale: 1}

// fromTables builds an outfit from a creature's row in monsters.csv: its
// figure, and a skin for each slot the row fills.
//
// A creature figure usually has a single shape -- the wolf is one strip mesh
// called "Editable Mesh" -- so the body skin also stands in for any shape
// whose name is not a slot, through the empty fallback prefix.
func fromTables(game string, id int) (outfit, error) {
	t, err := gamedata.Load(game)
	if err != nil {
		return outfit{}, err
	}
	m, ok := t.Monsters[id]
	if !ok {
		return outfit{}, fmt.Errorf("no model %d in monsters.csv", id)
	}
	fig, ok := t.Figures[m.Figure]
	if !ok || fig.File == "" {
		return outfit{}, fmt.Errorf("model %d (%s): figure %d not in monnifs.csv", id, m.Name, m.Figure)
	}
	path, err := findCase(filepath.Join(game, "figures"), fig.File+".nif")
	if err != nil {
		return outfit{}, fmt.Errorf("model %d (%s): %w", id, m.Name, err)
	}
	o := outfit{desc: m.Name, figure: filepath.Base(path), model: id, textures: map[string]string{}}
	for slot, sid := range m.Skins {
		s, ok := t.Skins[sid]
		if !ok {
			return outfit{}, fmt.Errorf("model %d: %s skin %d not in skins.csv", id, slot, sid)
		}
		o.textures[slot] = s.DDS()
		if slot == "Body" {
			o.textures[""] = s.DDS()
		}
	}
	meta.model, meta.scale, meta.animSet = id, m.Scale, fig.AnimSet
	meta.title = m.Name
	fmt.Printf("model %d %q: figure %d %s, scale %.2f, anim set %d\n",
		id, m.Name, fig.ID, o.figure, m.Scale, fig.AnimSet)
	return o, nil
}

// slug turns a table name into a directory name: "Large Grey Wolf" ->
// "large-grey-wolf".
func slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// findCase finds a file regardless of case: the tables write "wolf" for
// figures/Wolf.NIF.
func findCase(dir, name string) (string, error) {
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

func outfitNames() string {
	names := make([]string, 0, len(outfits))
	for k := range outfits {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func run(game, figure, name, out, shapeList string, listOnly, noTex bool) error {
	path := filepath.Join(game, "figures", figure)
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	f, err := nif.Parse(raw)
	if err != nil {
		return err
	}
	fmt.Printf("%s  %s\n", figure, f.HeaderString)
	fmt.Printf("  %d blocks, %d bytes\n", len(f.Blocks), len(raw))

	if listOnly {
		for i, b := range f.Blocks {
			s, ok := b.(*nif.TriShape)
			if !ok {
				continue
			}
			d, _ := f.Block(s.Data).(*nif.ShapeData)
			si, _ := f.Block(s.Skin).(*nif.SkinInstance)
			nb := 0
			if si != nil {
				nb = len(si.Bones)
			}
			fmt.Printf("  %-3d %-24s v=%-5d tri=%-5d bones=%d\n", i, s.Name,
				len(d.Vertices), len(d.Triangles), nb)
		}
		return nil
	}

	// The skeleton is every NiNode, exported whole: it is only a hundred-odd
	// nodes and keeping it intact means the viewer can pose any joint by name.
	bones, boneIndex := buildSkeleton(f)
	fmt.Printf("  skeleton: %d nodes\n", len(bones))

	wanted := map[string]bool{}
	for _, s := range strings.Split(shapeList, ",") {
		if s = strings.TrimSpace(s); s != "" {
			wanted[strings.ToLower(s)] = true
		}
	}
	if len(wanted) == 0 {
		wanted = defaultShapes(f)
	}

	world := f.WorldTransforms()
	var (
		verts     []byte
		indices   []uint32
		shapes    []shapeOut
		worstW    float32
		worstBind float32
		bbMin     = [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
		bbMax     = [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
		texNeed   = map[string]bool{}
		found     = map[string]bool{}
	)

	for i, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		// Shape names carry a trailing ":NN" instance tag; match on the stem.
		stem := strings.ToLower(s.Name)
		if c := strings.IndexByte(stem, ':'); c >= 0 {
			stem = stem[:c]
		}
		if !wanted[stem] {
			continue
		}
		found[stem] = true

		d, _ := f.Block(s.Data).(*nif.ShapeData)
		si, _ := f.Block(s.Skin).(*nif.SkinInstance)
		if d == nil || si == nil {
			return fmt.Errorf("shape %s has no geometry or no skin", s.Name)
		}
		sd, _ := f.Block(si.Data).(*nif.SkinData)
		if sd == nil || len(sd.Bones) != len(si.Bones) {
			return fmt.Errorf("shape %s: skin data does not match bone list", s.Name)
		}

		// Verify the bind pose for this shape before trusting its matrices.
		for bi, ref := range si.Bones {
			got := world[ref].Mul(sd.Bones[bi].Skin)
			if dv := got.MaxDiff(world[i]); dv > worstBind {
				worstBind = dv
			}
		}

		// Invert the file's bone-major weight lists into per-vertex lists.
		influences := make([][]influence, len(d.Vertices))
		for bi, bd := range sd.Bones {
			for _, w := range bd.Weights {
				if int(w.Index) >= len(influences) {
					return fmt.Errorf("shape %s: weight references vertex %d of %d",
						s.Name, w.Index, len(influences))
				}
				influences[w.Index] = append(influences[w.Index], influence{bi, w.Weight})
			}
		}

		base := len(verts) / vertStride
		shapeBones := make([]int, len(si.Bones))
		invBind := make([][12]float32, len(si.Bones))
		for bi, ref := range si.Bones {
			gi, ok := boneIndex[ref]
			if !ok {
				return fmt.Errorf("shape %s: bone block %d is not a node", s.Name, ref)
			}
			shapeBones[bi] = gi
			invBind[bi] = rows12(sd.Bones[bi].Skin)
		}

		for vi := range d.Vertices {
			idx, wts, sum := pack(influences[vi])
			if dv := float32(math.Abs(float64(sum - 1))); dv > worstW {
				worstW = dv
			}
			p := d.Vertices[vi]
			// Track extents in the posed position so the reported height is
			// the character's, not the raw buffer's.
			wp := world[i].Apply(p)
			for k := 0; k < 3; k++ {
				if wp[k] < bbMin[k] {
					bbMin[k] = wp[k]
				}
				if wp[k] > bbMax[k] {
					bbMax[k] = wp[k]
				}
			}
			var n [3]float32
			if vi < len(d.Normals) {
				n = d.Normals[vi]
			}
			var uv [2]float32
			if len(d.UV) > 0 && vi < len(d.UV[0]) {
				uv = d.UV[0][vi]
			}
			verts = appendVertex(verts, p, n, uv, idx, wts)
		}
		first := len(indices)
		for _, t := range d.Triangles {
			indices = append(indices,
				uint32(base)+uint32(t[0]),
				uint32(base)+uint32(t[1]),
				uint32(base)+uint32(t[2]))
		}

		so := shapeOut{
			Name: s.Name, First: first, Count: len(indices) - first,
			Bones: shapeBones, InvBind: invBind,
			Alpha: 1, Diffuse: [3]float32{1, 1, 1},
			HasColors: d.Colors != nil,
		}
		for _, pr := range s.Properties {
			if m, ok := f.Block(pr).(*nif.Material); ok {
				so.Diffuse, so.Emissive, so.Alpha = m.Diffuse, m.Emissive, m.Alpha
			}
		}
		if tex := textureFor(s.Name); tex != "" {
			so.Texture = strings.TrimSuffix(tex, ".dds") + ".png"
			texNeed[tex] = true
		}
		shapes = append(shapes, so)
		fmt.Printf("  + %-12s %4d verts %4d tris  %2d bones  tex %s\n",
			s.Name, len(d.Vertices), len(d.Triangles), len(si.Bones), so.Texture)
	}

	for s := range wanted {
		if !found[s] {
			return fmt.Errorf("no shape named %q in %s (try -list)", s, figure)
		}
	}
	if len(shapes) == 0 {
		return fmt.Errorf("no shapes selected")
	}

	fmt.Printf("  bind pose residual %.4g, worst weight-sum error %.4g\n", worstBind, worstW)
	if worstW > 1e-3 {
		return fmt.Errorf("vertex weights do not sum to 1 (off by %g); skinning would distort", worstW)
	}

	dir := filepath.Join(out, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// mesh.bin is the vertex block followed by the index block.
	blob := make([]byte, 0, len(verts)+4*len(indices))
	blob = append(blob, verts...)
	for _, v := range indices {
		blob = binary.LittleEndian.AppendUint32(blob, v)
	}
	if err := os.WriteFile(filepath.Join(dir, "mesh.bin"), blob, 0o644); err != nil {
		return err
	}

	// Which way the model faces is not recorded anywhere, so derive it: the toes
	// sit forward of the ankle, so the sign of that offset along the body's
	// shallow axis gives the facing direction.
	fwd := forwardSign(f, world)
	fmt.Printf("  facing: model %+dY (derived from toe-vs-ankle offset)\n", fwd)

	man := manifest{
		Name:        name,
		Source:      figure,
		ForwardY:    fwd,
		Title:       meta.title,
		Model:       meta.model,
		Scale:       meta.scale,
		AnimSet:     meta.animSet,
		Equip:       meta.equip,
		Format:      "pos3f,normal3f,uv2f,bone4u8,weight4f; then uint32 indices",
		VertexCount: len(verts) / vertStride,
		IndexCount:  len(indices),
		Stride:      vertStride,
		Min:         bbMin, Max: bbMax,
		Height: bbMax[2] - bbMin[2],
		// DAoC models are authored Z-up; the viewer rotates into its own
		// Y-up scene rather than baking a rotation into the vertices.
		UpAxis: "z",
		Bones:  bones, Shapes: shapes,
	}
	js, err := json.MarshalIndent(man, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "char.json"), js, 0o644); err != nil {
		return err
	}

	// Equipment, after the character so its sockets can be checked.
	for _, e := range meta.equip {
		found := false
		for _, b := range bones {
			found = found || b.Name == e.Bone
		}
		if !found {
			return fmt.Errorf("no socket %q on %s for %s", e.Bone, figure, e.Item)
		}
		if err := convertItem(game, e, filepath.Join(dir, "equip", e.Slot)); err != nil {
			return fmt.Errorf("equip %s: %w", e.Slot, err)
		}
	}

	if !noTex {
		if err := extractTextures(game, filepath.Join(dir, "tex"), texNeed); err != nil {
			return err
		}
	}

	fmt.Printf("  extents x %.1f..%.1f  y %.1f..%.1f  z %.1f..%.1f  (height %.1f units)\n",
		bbMin[0], bbMax[0], bbMin[1], bbMax[1], bbMin[2], bbMax[2], man.Height)
	fmt.Printf("  wrote %s (%d verts, %d tris, %d shapes)\n",
		dir, man.VertexCount, len(indices)/3, len(shapes))
	return nil
}

type influence struct {
	bone int
	w    float32
}

// pack reduces a vertex's influences to the fixed-width set the GPU wants,
// keeping the heaviest and renormalising so the weights still sum to one.
func pack(in []influence) (idx [maxInfluences]byte, wts [maxInfluences]float32, sum float32) {
	sort.SliceStable(in, func(i, j int) bool { return in[i].w > in[j].w })
	if len(in) > maxInfluences {
		in = in[:maxInfluences]
	}
	var total float32
	for _, v := range in {
		total += v.w
	}
	if total == 0 {
		// An unweighted vertex would collapse to the origin; pin it to its
		// first bone instead.
		if len(in) > 0 {
			idx[0] = byte(in[0].bone)
		}
		wts[0] = 1
		return idx, wts, 1
	}
	for i, v := range in {
		idx[i] = byte(v.bone)
		wts[i] = v.w / total
	}
	for _, w := range wts {
		sum += w
	}
	return idx, wts, sum
}

func appendVertex(b []byte, p, n [3]float32, uv [2]float32, idx [maxInfluences]byte, w [maxInfluences]float32) []byte {
	put := func(f float32) { b = binary.LittleEndian.AppendUint32(b, math.Float32bits(f)) }
	for _, v := range p {
		put(v)
	}
	for _, v := range n {
		put(v)
	}
	for _, v := range uv {
		put(v)
	}
	b = append(b, idx[:]...)
	for _, v := range w {
		put(v)
	}
	return b
}

// rows12 flattens a transform into three rows of four, which is what the
// shader reads as a mat3x4 and is enough for a rigid transform with scale.
func rows12(t nif.Transform) [12]float32 {
	var o [12]float32
	for r := 0; r < 3; r++ {
		o[r*4+0] = t.Rot[r*3+0] * t.Scale
		o[r*4+1] = t.Rot[r*3+1] * t.Scale
		o[r*4+2] = t.Rot[r*3+2] * t.Scale
		o[r*4+3] = t.Trans[r]
	}
	return o
}

func buildSkeleton(f *nif.File) ([]boneOut, map[int32]int) {
	parents := f.Parents()
	index := map[int32]int{}
	var bones []boneOut
	for i, b := range f.Blocks {
		n, ok := b.(*nif.Node)
		if !ok {
			continue
		}
		index[int32(i)] = len(bones)
		bones = append(bones, boneOut{
			Name: n.Name, Parent: int(parents[i]),
			T: n.Local.Trans, R: n.Local.Rot, S: n.Local.Scale,
		})
	}
	// Parents were block indices; remap to positions in the bone array.
	for i := range bones {
		if bones[i].Parent < 0 {
			continue
		}
		if p, ok := index[int32(bones[i].Parent)]; ok {
			bones[i].Parent = p
		} else {
			bones[i].Parent = -1
		}
	}
	return bones, index
}

// textureFor picks a slot texture by longest matching shape-name prefix.
func textureFor(shape string) string {
	// An empty prefix is the fallback: it matches every shape, and loses to
	// any real one.
	best, bestLen := "", -1
	for prefix, tex := range slotTextures {
		if len(prefix) > bestLen && strings.HasPrefix(strings.ToLower(shape), strings.ToLower(prefix)) {
			best, bestLen = tex, len(prefix)
		}
	}
	return best
}

// extractTextures hunts the named DDS files through every skin archive and
// writes them out as PNG.
func extractTextures(game, dir string, want map[string]bool) error {
	if len(want) == 0 {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	remaining := map[string]bool{}
	for k := range want {
		remaining[strings.ToLower(k)] = true
	}
	var archives []string
	for _, sub := range []string{"skins", "Mskins", "fig3"} {
		g, _ := filepath.Glob(filepath.Join(game, "figures", sub, "*.mpk"))
		archives = append(archives, g...)
	}
	for _, path := range archives {
		if len(remaining) == 0 {
			break
		}
		a, err := mpak.Open(path)
		if err != nil {
			continue
		}
		for _, e := range a.Entries {
			key := strings.ToLower(filepath.Base(e.Name))
			if !remaining[key] {
				continue
			}
			b, err := a.ReadEntry(e)
			if err != nil {
				return fmt.Errorf("%s/%s: %w", filepath.Base(path), e.Name, err)
			}
			img, err := dds.Decode(b)
			if err != nil {
				return fmt.Errorf("%s: %w", e.Name, err)
			}
			// DAoC's character skins are not alpha-masked. The DXT5 faces
			// carry a smoothly varying alpha channel the engine uses for
			// something else -- nor_m_head01 has ten opaque pixels out of
			// 65536, so reading it as opacity would erase the face. Flatten
			// it and keep the colour, which is a correct skin tone.
			flattened := 0
			for i := 3; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 0xFF {
					img.Pix[i] = 0xFF
					flattened++
				}
			}
			dst := filepath.Join(dir, strings.TrimSuffix(key, ".dds")+".png")
			if err := writePNG(dst, img); err != nil {
				return err
			}
			r := img.Bounds()
			note := ""
			if flattened > 0 {
				note = fmt.Sprintf("  (alpha flattened on %d of %d texels)",
					flattened, r.Dx()*r.Dy())
			}
			fmt.Printf("  tex %-22s %dx%d from %s%s\n",
				key, r.Dx(), r.Dy(), filepath.Base(path), note)
			delete(remaining, key)
		}
	}
	if len(remaining) > 0 {
		missing := make([]string, 0, len(remaining))
		for k := range remaining {
			missing = append(missing, k)
		}
		sort.Strings(missing)
		return fmt.Errorf("textures not found in any skin archive: %s", strings.Join(missing, ", "))
	}
	return nil
}

func writePNG(path string, img image.Image) error {
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

// forwardSign works out which way down the body's shallow axis the figure
// faces. The feet are the reliable landmark: toes lead the ankle.
func forwardSign(f *nif.File, world []nif.Transform) int {
	var foot, toe *nif.Transform
	for i, b := range f.Blocks {
		n, ok := b.(*nif.Node)
		if !ok {
			continue
		}
		switch strings.ToLower(n.Name) {
		case "bip01 l foot":
			foot = &world[i]
		case "bip01 l toe0":
			toe = &world[i]
		}
	}
	if foot == nil || toe == nil {
		return 1
	}
	if toe.Trans[1] < foot.Trans[1] {
		return -1
	}
	return 1
}

// shapeStem drops the ":NN" instance tag the exporter appends to shape names.
func shapeStem(name string) string {
	s := strings.ToLower(name)
	if c := strings.IndexByte(s, ':'); c >= 0 {
		s = s[:c]
	}
	return s
}

// defaultShapes picks what to draw when nothing names the shapes.
//
// A race figure carries every armour tier of every slot, and drawing them all
// stacks each outfit on top of the others. Where the shapes are named by slot
// and tier, take the lightest tier of each slot -- what an unarmoured NPC
// wears -- and the first head. Anything else is a creature, whose shapes are
// all simply parts of it.
func defaultShapes(f *nif.File) map[string]bool {
	var stems []string
	seen := map[string]bool{}
	for _, b := range f.Blocks {
		if s, ok := b.(*nif.TriShape); ok && !seen[shapeStem(s.Name)] {
			seen[shapeStem(s.Name)] = true
			stems = append(stems, shapeStem(s.Name))
		}
	}
	sort.Strings(stems)
	slots := []string{"heada", "body", "lbody", "arms", "legs", "gloves", "boots"}
	pick := map[string]bool{}
	for _, slot := range slots {
		for _, s := range stems {
			if rest, ok := strings.CutPrefix(s, slot); ok && rest != "" && strings.Trim(rest, "0123456789") == "" {
				pick[s] = true // sorted, so this is the lowest tier
				break
			}
		}
	}
	if len(pick) == 0 {
		for _, s := range stems {
			pick[s] = true
		}
	}
	return pick
}
