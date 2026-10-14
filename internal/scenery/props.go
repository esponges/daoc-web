package scenery

// Writing baked scenery for the viewer: one vertex-and-index buffer,
// props.bin, and props.json, the models and the instances placing them,
// with the textures the models name beside them in tex/. propconv writes a
// zone's fixtures this way and interiorconv a city's or dungeon's pieces.

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"math"
	"os"
	"path/filepath"
)

// Instance places a model: at a position in zone coordinates, turned by
// Yaw radians about Axis -- about up when Axis is absent -- and scaled.
type Instance struct {
	Model int        `json:"m"`
	Pos   [3]float32 `json:"p"`
	Yaw   float32    `json:"yaw"`
	Scale float32    `json:"s"`
	// Axis is written only for a turn that is not about up: a few dungeon
	// props lie tipped over. Yaw is then the angle about it.
	Axis *[3]float32 `json:"axis,omitempty"`
}

// Manifest is props.json.
type Manifest struct {
	Zone        int        `json:"zone"`
	Stride      int        `json:"stride"`
	VertexCount int        `json:"vertexCount"`
	IndexCount  int        `json:"indexCount"`
	VertexBytes int        `json:"vertexBytes"`
	IndexBytes  int        `json:"indexBytes"`
	UpAxis      string     `json:"upAxis"`
	Models      []Model    `json:"models"`
	Instances   []Instance `json:"instances"`
	Skipped     []string   `json:"skipped"`
}

// WriteProps writes props.bin, props.json and the textures the models use
// into dir, looking textures up in texDirs.
func WriteProps(dir string, zone int, texDirs []string, models []Model, verts []Vertex, indices []uint32, inst []Instance, skipped []string) error {
	if err := os.MkdirAll(filepath.Join(dir, "tex"), 0o755); err != nil {
		return err
	}
	// Pack both arrays into one file so the viewer makes a single request.
	buf := &bytes.Buffer{}
	var b [4]byte
	f32 := func(vals ...float32) {
		for _, v := range vals {
			binary.LittleEndian.PutUint32(b[:], math.Float32bits(v))
			buf.Write(b[:])
		}
	}
	for _, v := range verts {
		f32(v.P[0], v.P[1], v.P[2], v.N[0], v.N[1], v.N[2], v.UV[0], v.UV[1])
		buf.Write(v.C[:])
	}
	vertexBytes := buf.Len()
	for _, i := range indices {
		binary.Write(buf, binary.LittleEndian, i)
	}
	if err := os.WriteFile(filepath.Join(dir, "props.bin"), buf.Bytes(), 0o644); err != nil {
		return err
	}
	man := Manifest{
		Zone:        zone,
		Stride:      Stride,
		VertexCount: len(verts),
		IndexCount:  len(indices),
		VertexBytes: vertexBytes,
		IndexBytes:  buf.Len() - vertexBytes,
		UpAxis:      "Z",
		Models:      models,
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

	// Embedded textures are written from the models; the rest are looked
	// up beside them.
	want := map[string]bool{}
	embedded := map[string]image.Image{}
	for _, mo := range models {
		for _, g := range mo.Groups {
			if g.Texture != "" {
				want[g.Texture] = true
			}
		}
		for _, p := range mo.Particles {
			want[p.Texture] = true
		}
		for n, img := range mo.Embedded {
			embedded[n] = img
		}
	}
	external := map[string]bool{}
	for n := range want {
		if img, ok := embedded[n]; ok {
			if err := WritePNG(filepath.Join(dir, "tex", n+".png"), img); err != nil {
				return err
			}
			continue
		}
		external[n] = true
	}
	return ExtractTextures(texDirs, filepath.Join(dir, "tex"), external)
}

// TextureCount is how many distinct textures the models name.
func TextureCount(models []Model) int {
	want := map[string]bool{}
	for _, mo := range models {
		for _, g := range mo.Groups {
			if g.Texture != "" {
				want[g.Texture] = true
			}
		}
		for _, p := range mo.Particles {
			want[p.Texture] = true
		}
	}
	return len(want)
}

// collisionModel locates one model's hull in collision.bin.
type collisionModel struct {
	V  int `json:"v"`  // first vertex
	NV int `json:"nv"` // vertex count
	I  int `json:"i"`  // first index
	NI int `json:"ni"` // index count
	// R is a trunk radius for a model with no hull (0 if none).
	R float32 `json:"r,omitempty"`
}

// WriteCollision writes the models' collision hulls beside props.bin:
// collision.bin holds every hull's float32 x y z vertices (model space) and
// then its uint32 indices, and collision.json says where each model's lie, in
// the same order as props.json's models. A model with no hull has zero counts.
// The server places them with props.json's instances.
func WriteCollision(dir string, models []Model) error {
	var verts [][3]float32
	var idx []uint32
	cm := make([]collisionModel, len(models))
	for i, m := range models {
		cm[i].R = m.Radius
		if m.Hull == nil {
			continue
		}
		cm[i] = collisionModel{R: m.Radius, V: len(verts), NV: len(m.Hull.Verts), I: len(idx), NI: len(m.Hull.Tris) * 3}
		verts = append(verts, m.Hull.Verts...)
		for _, t := range m.Hull.Tris {
			idx = append(idx, t[0], t[1], t[2])
		}
	}
	buf := &bytes.Buffer{}
	for _, v := range verts {
		for _, c := range v {
			binary.Write(buf, binary.LittleEndian, math.Float32bits(c))
		}
	}
	vb := buf.Len()
	for _, i := range idx {
		binary.Write(buf, binary.LittleEndian, i)
	}
	if err := os.WriteFile(filepath.Join(dir, "collision.bin"), buf.Bytes(), 0o644); err != nil {
		return err
	}
	js, err := json.MarshalIndent(map[string]any{"vertexBytes": vb, "models": cm}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "collision.json"), js, 0o644)
}
