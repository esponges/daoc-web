package main

// Held items: a weapon in the hand, a shield on the arm.
//
// An item is a small static model in items/, and it carries its own
// attachment markers -- empty nodes named HELD, BELT, BACK and GROUND -- for
// the ways it can be worn. The character carries the matching sockets on its
// skeleton: "Bip01 R Held" in the right hand, "Bip01 L Shield" on the left
// forearm, "Bip01 L Back" across the shoulders. Wielding is lining the two up:
// the item's HELD marker goes where the socket is.
//
// So the converter bakes every vertex into the HELD marker's frame,
//
//	v' = inverse(HELD world) * shape world * v
//
// and the item's single bone is then simply the socket: the viewer sets it to
// the socket's world transform and draws it with the character's own shader.
// The output is an ordinary character directory -- char.json, mesh.bin, tex/
// -- with one bone, so nothing new is needed to load or draw it.

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"daocweb/internal/dds"
	"daocweb/internal/nif"
)

// An equip entry puts one item from items/ on one socket bone.
type equip struct {
	Slot string `json:"slot"` // directory under equip/: "right", "left"
	Bone string `json:"bone"` // the character's socket, e.g. "Bip01 R Held"
	Item string `json:"item"` // .nif in items/, no extension
	Name string `json:"name"` // what items.csv calls it
}

// convertItem writes one item, baked into its HELD frame, to dir.
func convertItem(game string, e equip, dir string) error {
	path, err := findCase(filepath.Join(game, "items"), e.Item+".nif")
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	f, err := nif.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	world := f.WorldTransforms()

	// The marker to hold the item by.
	grip := -1
	for i, b := range f.Blocks {
		if n, ok := b.(*nif.Node); ok && strings.EqualFold(n.Name, "HELD") {
			grip = i
		}
	}
	if grip < 0 {
		return fmt.Errorf("%s has no HELD marker", filepath.Base(path))
	}
	toGrip := world[grip].Inverse()

	var (
		verts   []byte
		indices []uint32
		shapes  []shapeOut
		texNeed = map[string]bool{}
		bbMin   = [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
		bbMax   = [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
	)
	identity := rows12(nif.Identity)
	one := [maxInfluences]float32{1}
	for i, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		d, _ := f.Block(s.Data).(*nif.ShapeData)
		if d == nil || len(d.Vertices) == 0 || len(d.Triangles) == 0 {
			continue
		}
		m := toGrip.Mul(world[i])
		base := len(verts) / vertStride
		for vi, p := range d.Vertices {
			q := m.Apply(p)
			for k := 0; k < 3; k++ {
				bbMin[k] = min(bbMin[k], q[k])
				bbMax[k] = max(bbMax[k], q[k])
			}
			var n [3]float32
			if vi < len(d.Normals) {
				n = rotateOnly(m, d.Normals[vi])
			}
			var uv [2]float32
			if len(d.UV) > 0 && vi < len(d.UV[0]) {
				uv = d.UV[0][vi]
			}
			verts = appendVertex(verts, q, n, uv, [maxInfluences]byte{}, one)
		}
		first := len(indices)
		for _, t := range d.Triangles {
			indices = append(indices, uint32(base)+uint32(t[0]), uint32(base)+uint32(t[1]), uint32(base)+uint32(t[2]))
		}
		so := shapeOut{
			Name: s.Name, First: first, Count: len(indices) - first,
			Bones: []int{0}, InvBind: [][12]float32{identity},
			Alpha: 1, Diffuse: [3]float32{1, 1, 1},
		}
		for _, pr := range s.Properties {
			if mat, ok := f.Block(pr).(*nif.Material); ok {
				so.Diffuse, so.Emissive, so.Alpha = mat.Diffuse, mat.Emissive, mat.Alpha
			}
		}
		if tex := itemTexture(game, f, s); tex != "" {
			so.Texture = strings.TrimSuffix(tex, ".dds") + ".png"
			texNeed[tex] = true
		}
		shapes = append(shapes, so)
	}
	if len(shapes) == 0 {
		return fmt.Errorf("%s has no geometry", filepath.Base(path))
	}

	if err := os.MkdirAll(filepath.Join(dir, "tex"), 0o755); err != nil {
		return err
	}
	blob := append([]byte(nil), verts...)
	for _, v := range indices {
		blob = binary.LittleEndian.AppendUint32(blob, v)
	}
	if err := os.WriteFile(filepath.Join(dir, "mesh.bin"), blob, 0o644); err != nil {
		return err
	}
	man := manifest{
		Name: e.Slot, Source: filepath.Base(path), Title: e.Name,
		Format:      "pos3f,normal3f,uv2f,bone4u8,weight4f; then uint32 indices",
		VertexCount: len(verts) / vertStride, IndexCount: len(indices), Stride: vertStride,
		Min: bbMin, Max: bbMax, Height: bbMax[2] - bbMin[2], UpAxis: "z", Scale: 1,
		// One bone, the socket. Its rest transform is the identity; the
		// viewer replaces it with the socket's every frame.
		Bones:  []boneOut{{Name: "HELD", Parent: -1, R: [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1}, S: 1}},
		Shapes: shapes,
	}
	js, err := json.MarshalIndent(man, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "char.json"), js, 0o644); err != nil {
		return err
	}
	for tex := range texNeed {
		src, err := findCase(filepath.Join(game, "items"), tex)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		img, err := dds.Decode(b)
		if err != nil {
			return fmt.Errorf("%s: %w", tex, err)
		}
		if err := writePNG(filepath.Join(dir, "tex", strings.TrimSuffix(strings.ToLower(tex), ".dds")+".png"), img); err != nil {
			return err
		}
	}
	fmt.Printf("  item %-6s %-22s on %-15s %4d verts %4d tris  %.1f units long\n",
		e.Slot, man.Source, e.Bone, man.VertexCount, len(indices)/3,
		float32(math.Sqrt(float64(sq(bbMax[0]-bbMin[0])+sq(bbMax[1]-bbMin[1])+sq(bbMax[2]-bbMin[2])))))
	return nil
}

// itemTexture picks the texture an item is drawn with: the first of the
// textures its shape names that exists in items/.
//
// For a weapon that is simply its base texture. A shield names three --
// "cloakpattern01", "symbol_001" and the material, "Shield_Wood" or
// "sh_metal01" -- and the first two are not in the install at all: they are
// the guild's pattern and emblem, which the game composites over the material
// at runtime (items.csv's "Strip Textures" flag marks them). Without a guild,
// the material is what is left.
func itemTexture(game string, f *nif.File, s *nif.TriShape) string {
	var names []string
	for _, b := range f.Blocks {
		if src, ok := b.(*nif.SourceTexture); ok && src.FileName != "" {
			n := strings.ToLower(filepath.Base(strings.ReplaceAll(src.FileName, `\`, "/")))
			names = append(names, strings.TrimSuffix(n, filepath.Ext(n))+".dds")
		}
	}
	for _, n := range names {
		if _, err := findCase(filepath.Join(game, "items"), n); err == nil {
			return n
		}
	}
	return ""
}

// rotateOnly applies a transform's rotation to a direction.
func rotateOnly(t nif.Transform, v [3]float32) [3]float32 {
	var o [3]float32
	for r := 0; r < 3; r++ {
		o[r] = t.Rot[r*3]*v[0] + t.Rot[r*3+1]*v[1] + t.Rot[r*3+2]*v[2]
	}
	return o
}

func sq(x float32) float32 { return x * x }
