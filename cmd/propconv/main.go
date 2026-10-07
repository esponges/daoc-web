// Command propconv converts a zone's scenery -- the trees, fences, buildings
// and other props that fixtures.csv places -- into a single vertex buffer plus
// an instance list the viewer can draw.
//
// Two CSVs inside the zone's dat archive drive it. nifs.csv maps a numeric id
// to a model file in zones/Nifs, and fixtures.csv places that id in the world:
//
//	ID,NIF #,Textual Name,X,Y,Z,A,Scale,...,3D Angle,3D Axis X,3D Axis Y,3D Axis Z
//	1,403,Pine,52736.00,19200.00,4960.00,225,150,...,2.356194,0,0,1
//
// The Z is absolute world height, not an offset: it matches the heightmap this
// pipeline derives from terrain.pcx to the unit, which is a useful check on
// both. Rotation is taken from the axis-angle columns rather than the A
// column, because A is a clockwise heading while the axis-angle pair is a
// plain rotation and needs no convention guessed. Scale is a percentage.
//
// Each model is flattened at conversion time: the scene graph is walked, world
// transforms are baked into the vertices, and the result is grouped by
// texture. Props never move, so there is nothing for the viewer to re-pose.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"daocweb/internal/dds"
	"daocweb/internal/mpak"
	"daocweb/internal/nif"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

// vertStride is pos[3] + normal[3] + uv[2], all float32.
const vertStride = 32

type groupOut struct {
	Texture string `json:"texture"`
	Alpha   bool   `json:"alpha"`
	First   int    `json:"first"` // index of the first element in the index buffer
	Count   int    `json:"count"`
}

type modelOut struct {
	Name   string     `json:"name"`
	File   string     `json:"file"`
	Min    [3]float32 `json:"min"`
	Max    [3]float32 `json:"max"`
	Groups []groupOut `json:"groups"`
}

type instanceOut struct {
	Model int        `json:"m"`
	Pos   [3]float32 `json:"p"`
	Yaw   float32    `json:"yaw"`
	Scale float32    `json:"s"`
}

type manifest struct {
	Zone        int           `json:"zone"`
	Stride      int           `json:"stride"`
	VertexCount int           `json:"vertexCount"`
	IndexCount  int           `json:"indexCount"`
	VertexBytes int           `json:"vertexBytes"`
	IndexBytes  int           `json:"indexBytes"`
	UpAxis      string        `json:"upAxis"`
	Models      []modelOut    `json:"models"`
	Instances   []instanceOut `json:"instances"`
	Skipped     []string      `json:"skipped"`
}

type vertex struct {
	P  [3]float32
	N  [3]float32
	UV [2]float32
}

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	zone := flag.Int("zone", 100, "zone number")
	out := flag.String("out", "web/data", "output directory")
	listOnly := flag.Bool("list", false, "list the models a zone places and stop")
	flag.Parse()

	if err := run(*game, *zone, *out, *listOnly); err != nil {
		fmt.Fprintln(os.Stderr, "propconv:", err)
		os.Exit(1)
	}
}

func run(game string, zone int, out string, listOnly bool) error {
	zoneDir := filepath.Join(game, "zones", fmt.Sprintf("zone%03d", zone))
	datPath := filepath.Join(zoneDir, fmt.Sprintf("dat%03d.mpk", zone))
	arc, err := mpak.Open(datPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", datPath, err)
	}
	nifsCSV, err := readEntry(arc, "nifs.csv")
	if err != nil {
		return err
	}
	fixCSV, err := readEntry(arc, "fixtures.csv")
	if err != nil {
		return err
	}

	models, err := parseNifs(nifsCSV)
	if err != nil {
		return fmt.Errorf("nifs.csv: %w", err)
	}
	fixtures, err := parseFixtures(fixCSV)
	if err != nil {
		return fmt.Errorf("fixtures.csv: %w", err)
	}
	fmt.Printf("zone %d: %d model definitions, %d fixtures\n", zone, len(models), len(fixtures))

	// Only convert what the zone actually places.
	used := map[int]int{}
	for _, f := range fixtures {
		used[f.NifID]++
	}
	ids := make([]int, 0, len(used))
	for id := range used {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	if listOnly {
		for _, id := range ids {
			m, ok := models[id]
			name := "<not in nifs.csv>"
			if ok {
				name = fmt.Sprintf("%-28s %s", m.Name, m.File)
			}
			fmt.Printf("  %4d  %5d placements  %s\n", id, used[id], name)
		}
		return nil
	}

	dir := filepath.Join(out, fmt.Sprintf("zone%03d", zone), "props")
	if err := os.MkdirAll(filepath.Join(dir, "tex"), 0o755); err != nil {
		return err
	}

	var (
		verts    []vertex
		indices  []uint32
		outModel []modelOut
		skipped  []string
		slot     = map[int]int{} // nif id -> index into outModel
		wantTex  = map[string]bool{}
	)
	for _, id := range ids {
		def, ok := models[id]
		if !ok {
			skipped = append(skipped, fmt.Sprintf("%d: no entry in nifs.csv (%d placements)", id, used[id]))
			continue
		}
		mo, err := convertModel(game, def, &verts, &indices)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v (%d placements)", def.File, err, used[id]))
			continue
		}
		for _, g := range mo.Groups {
			if g.Texture != "" {
				wantTex[g.Texture] = true
			}
		}
		slot[id] = len(outModel)
		outModel = append(outModel, *mo)
	}

	var inst []instanceOut
	placed := 0
	for _, f := range fixtures {
		s, ok := slot[f.NifID]
		if !ok {
			continue
		}
		inst = append(inst, instanceOut{
			Model: s,
			Pos:   [3]float32{f.X, f.Y, f.Z},
			Yaw:   f.Yaw,
			Scale: f.Scale,
		})
		placed++
	}

	// Pack both arrays into one file so the viewer makes a single request.
	buf := &bytes.Buffer{}
	for _, v := range verts {
		writeF32(buf, v.P[0], v.P[1], v.P[2], v.N[0], v.N[1], v.N[2], v.UV[0], v.UV[1])
	}
	vertexBytes := buf.Len()
	for _, i := range indices {
		binary.Write(buf, binary.LittleEndian, i)
	}
	if err := os.WriteFile(filepath.Join(dir, "props.bin"), buf.Bytes(), 0o644); err != nil {
		return err
	}

	man := manifest{
		Zone:        zone,
		Stride:      vertStride,
		VertexCount: len(verts),
		IndexCount:  len(indices),
		VertexBytes: vertexBytes,
		IndexBytes:  buf.Len() - vertexBytes,
		UpAxis:      "Z",
		Models:      outModel,
		Instances:   inst,
		Skipped:     skipped,
	}
	js, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "props.json"), js, 0o644); err != nil {
		return err
	}

	if err := extractTextures(game, filepath.Join(dir, "tex"), wantTex); err != nil {
		return err
	}

	fmt.Printf("\n%d models converted, %d skipped\n", len(outModel), len(skipped))
	for _, s := range skipped {
		fmt.Printf("  skipped %s\n", s)
	}
	fmt.Printf("%d of %d fixtures placed (%.1f%%)\n",
		placed, len(fixtures), 100*float64(placed)/float64(len(fixtures)))
	fmt.Printf("%d vertices, %d indices, %d textures -> %s\n",
		len(verts), len(indices), len(wantTex), dir)
	return nil
}

func writeF32(w io.Writer, vals ...float32) {
	var b [4]byte
	for _, v := range vals {
		binary.LittleEndian.PutUint32(b[:], math.Float32bits(v))
		w.Write(b[:])
	}
}

func readEntry(a *mpak.Archive, name string) ([]byte, error) {
	for _, e := range a.Entries {
		if strings.EqualFold(filepath.Base(e.Name), name) {
			return a.ReadEntry(e)
		}
	}
	return nil, fmt.Errorf("%s not found in archive", name)
}

// --- the CSVs ------------------------------------------------------------

type modelDef struct {
	ID   int
	Name string
	File string
}

type fixture struct {
	NifID int
	X     float32
	Y     float32
	Z     float32
	Yaw   float32
	Scale float32
}

// bothCSVs carry two header rows: a banner and the column names.
func records(b []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(b))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 3 {
		return nil, fmt.Errorf("only %d rows", len(rows))
	}
	return rows[2:], nil
}

func parseNifs(b []byte) (map[int]modelDef, error) {
	rows, err := records(b)
	if err != nil {
		return nil, err
	}
	out := map[int]modelDef{}
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(row[0]))
		if err != nil {
			continue
		}
		out[id] = modelDef{ID: id, Name: strings.TrimSpace(row[1]), File: strings.TrimSpace(row[2])}
	}
	return out, nil
}

func parseFixtures(b []byte) ([]fixture, error) {
	rows, err := records(b)
	if err != nil {
		return nil, err
	}
	var out []fixture
	for _, row := range rows {
		if len(row) < 19 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(row[1]))
		if err != nil {
			continue
		}
		f := fixture{NifID: id}
		f.X = atof(row[3])
		f.Y = atof(row[4])
		f.Z = atof(row[5])
		// The A column is a clockwise heading in degrees; the axis-angle
		// columns say the same thing without a convention to guess, so
		// take the angle and let the Z axis sign carry the direction.
		angle := atof(row[15])
		axisZ := atof(row[18])
		if axisZ < 0 {
			angle = -angle
		}
		f.Yaw = angle
		f.Scale = atof(row[7]) / 100
		if f.Scale <= 0 {
			f.Scale = 1
		}
		out = append(out, f)
	}
	return out, nil
}

func atof(s string) float32 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 32)
	return float32(v)
}

// --- model conversion ----------------------------------------------------

// texState is the render state a node inherits from its ancestors. NIF hangs
// properties off any NiAVObject and children inherit them, so a texture set on
// the model root applies to every shape beneath it.
type texState struct {
	texture string
	alpha   bool
}

func convertModel(game string, def modelDef, verts *[]vertex, indices *[]uint32) (*modelOut, error) {
	raw, err := loadModel(game, def.File)
	if err != nil {
		return nil, err
	}
	f, err := nif.Parse(raw)
	if err != nil {
		return nil, err
	}
	mo := &modelOut{Name: def.Name, File: def.File}

	// Collect geometry per texture so each group is one draw call.
	byTex := map[string]*struct {
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
			st = applyProps(f, b.Properties, st)
			for _, c := range b.Children {
				walk(c, world, st, depth+1)
			}

		case *nif.LODNode:
			if isCollision(b.Name) {
				return
			}
			world := parent.Mul(b.Local)
			st = applyProps(f, b.Properties, st)
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
			st = applyProps(f, b.Properties, st)
			d, _ := f.Block(b.Data).(*nif.ShapeData)
			if d == nil || len(d.Vertices) == 0 || len(d.Triangles) == 0 {
				return
			}
			g := byTex[st.texture]
			if g == nil {
				g = &struct {
					alpha bool
					idx   []uint32
				}{}
				byTex[st.texture] = g
			}
			g.alpha = g.alpha || st.alpha

			first := uint32(len(*verts))
			for i, p := range d.Vertices {
				v := vertex{P: world.Apply(p)}
				if i < len(d.Normals) {
					v.N = rotate(world, d.Normals[i])
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

	names := make([]string, 0, len(byTex))
	for k := range byTex {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, n := range names {
		g := byTex[n]
		if len(g.idx) == 0 {
			continue
		}
		mo.Groups = append(mo.Groups, groupOut{
			Texture: n,
			Alpha:   g.alpha,
			First:   len(*indices),
			Count:   len(g.idx),
		})
		*indices = append(*indices, g.idx...)
	}
	if len(mo.Groups) == 0 {
		return nil, fmt.Errorf("no drawable geometry")
	}
	mo.Min, mo.Max = min, max
	return mo, nil
}

// isCollision spots the invisible collision hulls the exporter ships beside
// the visible mesh. The fence models are the clearest case: a "Collisionswitch"
// node with a "collidee" branch and a "visible" one.
func isCollision(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "collidee") || strings.Contains(n, "collision_") ||
		strings.HasPrefix(n, "collide")
}

// applyProps resolves a node's property list, inheriting anything it does not
// override from its parent.
func applyProps(f *nif.File, refs []int32, st texState) texState {
	for _, p := range refs {
		switch b := f.Block(p).(type) {
		case *nif.Texturing:
			if !b.HasBase {
				continue
			}
			if src, ok := f.Block(b.Base.Source).(*nif.SourceTexture); ok && src.FileName != "" {
				st.texture = texName(src.FileName)
			}
		case *nif.AlphaProperty:
			// Bit 9 of the flags is the alpha-test enable. Trees need it:
			// the canopy is a few quads with a cut-out leaf texture.
			st.alpha = true
		}
	}
	return st
}

func texName(path string) string {
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

// loadModel finds a scenery model. They live one per .npk in zones/Nifs,
// occasionally loose beside them.
func loadModel(game, file string) ([]byte, error) {
	base := strings.TrimSuffix(file, filepath.Ext(file))
	dir := filepath.Join(game, "zones", "Nifs")
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
	return nil, fmt.Errorf("no model archive %s.npk", base)
}

// extractTextures pulls the scenery textures, which sit loose in zones/Nifs
// rather than in an archive.
func extractTextures(game, dir string, want map[string]bool) error {
	if len(want) == 0 {
		return nil
	}
	src := filepath.Join(game, "zones", "Nifs")
	names := make([]string, 0, len(want))
	for k := range want {
		names = append(names, k)
	}
	sort.Strings(names)

	var missing []string
	for _, n := range names {
		path, err := findCase(src, n+".dds")
		if err != nil {
			missing = append(missing, n)
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			missing = append(missing, n)
			continue
		}
		img, err := dds.Decode(b)
		if err != nil {
			return fmt.Errorf("%s: %w", n, err)
		}
		if err := writePNG(filepath.Join(dir, n+".png"), img); err != nil {
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
