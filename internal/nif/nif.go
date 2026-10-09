// Package nif reads the NetImmerse / Gamebryo models DAoC ships in figures/
// and items/.
//
// A .nif is a flat array of blocks preceded by a header that names each
// block's type. Blocks reference one another by index, so the file is really a
// serialised object graph: NiNode builds the transform hierarchy, NiTriShape
// hangs geometry off it, and NiSkinInstance binds that geometry to bones.
//
// The catch is that a block carries no length. Walking the file means knowing
// the exact field layout of every type present, and a single wrong field
// desynchronises everything after it. That makes the format unforgiving but
// also self-checking: the footer has to land precisely at end-of-file, which
// is a strong signal that every layout in between was right. Parse enforces
// this.
//
// Layouts are version-gated. DAoC ships two generations:
//
//	figures/*.NIF   Gamebryo File Format, Version 10.1.0.0
//	items/*.nif     NetImmerse File Format, Version 4.2.1.0
//
// Only what this pipeline needs is implemented, which for the character models
// is the complete set of ten block types they use.
package nif

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// Version numbers as packed by the format: 10.1.0.0 is 0x0A010000.
const (
	Ver4010012 = 0x0401000C // NetImmerse 4.1.0.12
	Ver4020100 = 0x04020100 // NetImmerse 4.2.1.0
	Ver4020200 = 0x04020200 // NetImmerse 4.2.2.0
	Ver1001000 = 0x0A000100 // Gamebryo 10.0.1.0
	Ver1001008 = 0x0A000108 // Gamebryo 10.0.1.8
	Ver1010000 = 0x0A010000 // Gamebryo 10.1.0.0
)

// Transform is NIF's native rigid transform: a 3x3 rotation, a translation and
// a uniform scale. Kept in this form rather than a 4x4 because that is how the
// file stores it and how the skinning maths composes.
type Transform struct {
	Rot   [9]float32 // row-major: Rot[r*3+c]
	Trans [3]float32
	Scale float32
}

// Identity is the no-op transform.
var Identity = Transform{Rot: [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1}, Scale: 1}

// Mul returns the transform that applies b first, then a.
func (a Transform) Mul(b Transform) Transform {
	var out Transform
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			var s float32
			for k := 0; k < 3; k++ {
				s += a.Rot[r*3+k] * b.Rot[k*3+c]
			}
			out.Rot[r*3+c] = s
		}
	}
	bt := a.Apply(b.Trans)
	out.Trans = bt
	out.Scale = a.Scale * b.Scale
	return out
}

// Apply maps a point through the transform.
func (a Transform) Apply(v [3]float32) [3]float32 {
	var out [3]float32
	for r := 0; r < 3; r++ {
		s := a.Rot[r*3+0]*v[0] + a.Rot[r*3+1]*v[1] + a.Rot[r*3+2]*v[2]
		out[r] = s*a.Scale + a.Trans[r]
	}
	return out
}

// MaxDiff is the largest absolute difference between two transforms across
// rotation, translation and scale. Used to verify the bind pose.
func (a Transform) MaxDiff(b Transform) float32 {
	d := float32(0)
	upd := func(x, y float32) {
		if v := float32(math.Abs(float64(x - y))); v > d {
			d = v
		}
	}
	for i := range a.Rot {
		upd(a.Rot[i], b.Rot[i])
	}
	for i := range a.Trans {
		upd(a.Trans[i], b.Trans[i])
	}
	upd(a.Scale, b.Scale)
	return d
}

// --- block types ---------------------------------------------------------

// ObjectNET is the common head of every named block.
type ObjectNET struct {
	Name       string
	ExtraData  []int32
	Controller int32
}

// AVObject adds the scene-graph transform. Children inherit it.
type AVObject struct {
	ObjectNET
	Flags      uint16
	Local      Transform
	Properties []int32
	Collision  int32
}

// Node is NiNode: an interior scene-graph node. Bones are plain Nodes; what
// makes one a bone is being referenced by a SkinInstance.
type Node struct {
	AVObject
	Children []int32
	Effects  []int32
}

// TriShape is NiTriShape: geometry plus an optional skin binding.
type TriShape struct {
	AVObject
	Data    int32
	Skin    int32
	Shader  string
	HasShdr bool
	ShdrUnk int32
}

// Triangle is one face, indexing into a ShapeData's vertex arrays.
type Triangle [3]uint16

// ShapeData is NiTriShapeData: the vertex arrays themselves.
type ShapeData struct {
	GroupID   int32
	Vertices  [][3]float32
	Normals   [][3]float32
	Colors    [][4]float32
	UV        [][][2]float32 // one slice per UV set
	Center    [3]float32
	Radius    float32
	Triangles []Triangle
	// Strips is set only by NiTriStripsData, which stores geometry as
	// triangle strips. Triangles is filled from it either way, so consumers
	// never need to care which topology the artist exported.
	Strips [][]uint16
}

// SkinInstance is NiSkinInstance: ties a TriShape to a bone list.
type SkinInstance struct {
	Data         int32
	Partition    int32
	SkeletonRoot int32
	Bones        []int32 // block indices of the bone Nodes
}

// VertWeight is one vertex's influence from a single bone.
type VertWeight struct {
	Index  uint16
	Weight float32
}

// BoneData is the per-bone half of a skin binding.
type BoneData struct {
	// Skin maps a vertex from the geometry's own space into this bone's
	// local space. Composed with the bone's world transform it reproduces
	// the bind pose, which is the invariant Validate checks.
	Skin    Transform
	Offset  [3]float32
	Radius  float32
	Weights []VertWeight
}

// SkinData is NiSkinData.
type SkinData struct {
	Skin             Transform // geometry space -> skeleton root space
	Partition        int32
	HasVertexWeights bool
	Bones            []BoneData
}

// Partition is one chunk of a NiSkinPartition: vertices grouped so that each
// chunk references few enough bones to fit a fixed-size GPU uniform array.
type Partition struct {
	Bones         []uint16
	VertexMap     []uint16
	Weights       []float32 // NumVertices * WeightsPerVertex
	BoneIndices   []byte    // NumVertices * WeightsPerVertex
	Triangles     []Triangle
	Strips        [][]uint16
	NumVertices   int
	WeightsPerVtx int
}

// SkinPartition is NiSkinPartition.
type SkinPartition struct {
	Parts []Partition
}

// Material is NiMaterialProperty.
type Material struct {
	ObjectNET
	Ambient    [3]float32
	Diffuse    [3]float32
	Specular   [3]float32
	Emissive   [3]float32
	Glossiness float32
	Alpha      float32
}

// VertexColor is NiVertexColorProperty.
type VertexColor struct {
	ObjectNET
	Flags        uint16
	VertexMode   uint32
	LightingMode uint32
}

// KeyframeController is NiKeyframeController: animates one target node.
type KeyframeController struct {
	Next      int32
	Flags     uint16
	Frequency float32
	Phase     float32
	StartTime float32
	StopTime  float32
	Target    int32
	Data      int32
}

// QuatKey is a rotation keyframe. Quaternions are stored w,x,y,z.
type QuatKey struct {
	Time  float32
	Value [4]float32
}

// VecKey is a translation keyframe.
type VecKey struct {
	Time  float32
	Value [3]float32
}

// FloatKey is a scale keyframe.
type FloatKey struct {
	Time  float32
	Value float32
}

// KeyframeData is NiKeyframeData: the sampled channels for one node.
type KeyframeData struct {
	RotationType uint32
	Rotations    []QuatKey
	Translations []VecKey
	Scales       []FloatKey
}

// Unknown stands in for a block type this package does not model. Because
// blocks carry no length, encountering one is fatal rather than skippable.
type Unknown struct{ Type string }

// --- file ----------------------------------------------------------------

// File is a parsed .nif.
type File struct {
	HeaderString string
	Version      uint32
	UserVersion  uint32
	BlockTypes   []string
	Blocks       []any
	TypeOf       []string // per block, for diagnostics
	Roots        []int32
	Extents      []Extent // byte range of each block, for diagnostics
}

// Extent records where a block was found. Only useful when a layout is wrong.
type Extent struct {
	Start, End int
	Type       string
}

// Parse reads a .nif. It fails if any block layout desynchronises, which it
// detects by requiring the footer to land exactly at end-of-file.
func Parse(b []byte) (*File, error) {
	nl := bytes.IndexByte(b, '\n')
	if nl < 0 || nl > 128 {
		return nil, fileError(BadHeader, "", 0, "no header string")
	}
	f := &File{HeaderString: string(b[:nl])}
	if !strings.Contains(f.HeaderString, "File Format") {
		return nil, fileError(BadHeader, "", 0, "bad header string %q", f.HeaderString)
	}
	r := &reader{b: b, p: nl + 1}
	f.Version = r.u32()

	switch {
	case f.Version == Ver1010000:
		// Gamebryo 10.1.0.0: a block-type table in the header, and a
		// zero uint32 ahead of the header body and of every block.
		f.UserVersion = r.u32()
		n := int(r.u32())
		nTypes := int(r.u16())
		if r.err != nil {
			return f, fileError(kindOf(r.err), "", r.p, "header: %w", r.err)
		}
		f.BlockTypes = make([]string, nTypes)
		for i := range f.BlockTypes {
			f.BlockTypes[i] = r.str()
		}
		idx := make([]uint16, n)
		for i := range idx {
			idx[i] = r.u16()
		}
		r.u32() // unknown, always zero in DAoC's files
		if r.err != nil {
			return f, fileError(kindOf(r.err), "", r.p, "header: %w", r.err)
		}
		f.Blocks = make([]any, n)
		f.TypeOf = make([]string, n)
		f.Extents = make([]Extent, n)
		for i := 0; i < n; i++ {
			if int(idx[i]) >= nTypes {
				return f, fileError(Corrupt, "", r.p, "block %d has type index %d of %d", i, idx[i], nTypes)
			}
			typ := f.BlockTypes[idx[i]]
			r.u32() // per-block separator, always zero
			start := r.p
			blk, err := f.parseBlock(r, typ)
			if err != nil {
				return f, blockError(i, typ, start, err)
			}
			f.Blocks[i], f.TypeOf[i] = blk, typ
			f.Extents[i] = Extent{start, r.p, typ}
			if r.err != nil {
				return f, blockError(i, typ, start, r.err)
			}
		}

	case f.Version >= Ver4010012 && f.Version <= Ver4020200:
		// NetImmerse 4.x: each block is prefixed by its own type name and
		// the block count is all the header carries. The zone's scenery
		// spans 4.1.0.12, 4.2.1.0 and 4.2.2.0, which differ only in
		// individual fields, all gated inside the block readers.
		n := int(r.u32())
		if r.err != nil {
			return f, fileError(kindOf(r.err), "", r.p, "header: %w", r.err)
		}
		f.Blocks = make([]any, 0, n)
		f.TypeOf = make([]string, 0, n)
		for i := 0; i < n; i++ {
			typ := r.str()
			if r.err != nil {
				return f, &Error{Kind: kindOf(r.err), Block: i, Offset: r.p, Err: fmt.Errorf("block %d type name: %w", i, r.err)}
			}
			start := r.p
			blk, err := f.parseBlock(r, typ)
			if err != nil {
				return f, blockError(i, typ, start, err)
			}
			f.Blocks = append(f.Blocks, blk)
			f.TypeOf = append(f.TypeOf, typ)
			f.Extents = append(f.Extents, Extent{start, r.p, typ})
			if r.err != nil {
				return f, blockError(i, typ, start, r.err)
			}
		}

	default:
		return f, fileError(Unsupported, "", 0, "unsupported version 0x%08X (%s)", f.Version, f.HeaderString)
	}

	// Footer: the root block list. It has to consume the rest of the file
	// exactly; anything left over means a layout above was wrong.
	nRoots := int(r.u32())
	if r.err != nil {
		return f, fileError(kindOf(r.err), "", r.p, "footer: %w", r.err)
	}
	f.Roots = make([]int32, nRoots)
	for i := range f.Roots {
		f.Roots[i] = r.i32()
	}
	if r.err != nil {
		return f, fileError(kindOf(r.err), "", r.p, "footer roots: %w", r.err)
	}
	if r.p != len(b) {
		return f, fileError(FooterMismatch, "", r.p, "footer ended at %d but file is %d bytes (%d left over); a block layout is wrong",
			r.p, len(b), len(b)-r.p)
	}
	return f, nil
}

func (f *File) parseBlock(r *reader, typ string) (any, error) {
	v := f.Version
	switch typ {
	case "NiNode":
		n := &Node{}
		f.readAV(r, &n.AVObject)
		n.Children = r.refs()
		n.Effects = r.refs()
		return n, nil

	case "NiTriShape":
		s := &TriShape{}
		f.readAV(r, &s.AVObject)
		s.Data = r.i32()
		s.Skin = r.i32()
		// A shader name sits here for 10.0.1.0 through 20.0.4; DAoC never
		// sets the flag but the byte is still written.
		if v >= Ver1001000 {
			s.HasShdr = r.boolean()
			if s.HasShdr {
				s.Shader = r.str()
				s.ShdrUnk = r.i32()
			}
		}
		return s, nil

	case "NiTriShapeData":
		return f.readShapeData(r)

	case "NiSkinInstance":
		s := &SkinInstance{}
		s.Data = r.i32()
		// The partition reference lives on NiSkinData up to 10.1.0.0 and
		// moves onto NiSkinInstance after; it is never on both. Reading it
		// here as well desynchronises by four bytes.
		if v > Ver1010000 {
			s.Partition = r.i32()
		}
		s.SkeletonRoot = r.i32()
		s.Bones = r.refs()
		return s, nil

	case "NiSkinData":
		return f.readSkinData(r)

	case "NiSkinPartition":
		return f.readSkinPartition(r)

	case "NiMaterialProperty":
		m := &Material{}
		f.readNET(r, &m.ObjectNET)
		if v <= 0x0A000102 { // flags dropped after 10.0.1.2
			r.u16()
		}
		m.Ambient = r.color3()
		m.Diffuse = r.color3()
		m.Specular = r.color3()
		m.Emissive = r.color3()
		m.Glossiness = r.f32()
		m.Alpha = r.f32()
		return m, nil

	case "NiVertexColorProperty":
		p := &VertexColor{}
		f.readNET(r, &p.ObjectNET)
		p.Flags = r.u16()
		if v <= 0x14000004 {
			p.VertexMode = r.u32()
			p.LightingMode = r.u32()
		}
		return p, nil

	case "NiKeyframeController":
		c := &KeyframeController{}
		c.Next = r.i32()
		c.Flags = r.u16()
		c.Frequency = r.f32()
		c.Phase = r.f32()
		c.StartTime = r.f32()
		c.StopTime = r.f32()
		c.Target = r.i32()
		c.Data = r.i32()
		return c, nil

	case "NiKeyframeData":
		return f.readKeyframeData(r)

	default:
		return f.parseWorldBlock(r, typ)
	}
}

func (f *File) readNET(r *reader, o *ObjectNET) {
	o.Name = r.str()
	if f.Version >= Ver1001000 {
		o.ExtraData = r.refs()
	} else {
		o.ExtraData = []int32{r.i32()}
	}
	o.Controller = r.i32()
}

func (f *File) readAV(r *reader, a *AVObject) {
	f.readNET(r, &a.ObjectNET)
	a.Flags = r.u16()
	// Attachment points -- "Bip01 R Shield", "HELD" -- are exported with
	// an all-NaN translation in sixteen of the figures, BritonF1.NIF among
	// them: the bytes are 7fc00000 (or ffc00000) three times over, in an
	// otherwise sound block. The game positions those nodes from whatever
	// is held, so the value never mattered to it. Accept it here, and only
	// here, and leave the node at its parent's origin.
	r.lax = true
	a.Local.Trans = r.vec3()
	r.lax = false
	for i, c := range a.Local.Trans {
		if math.IsNaN(float64(c)) || math.IsInf(float64(c), 0) {
			a.Local.Trans[i] = 0
		}
	}
	a.Local.Rot = r.mat33()
	a.Local.Scale = r.f32()
	if f.Version <= 0x04020200 {
		r.vec3() // velocity, dropped after 4.2.2.0
	}
	a.Properties = r.refs()
	if f.Version <= 0x04020200 {
		if r.boolean() {
			// bounding volume: type + whichever shape follows
			if r.u32() == 0 {
				r.mat33()
				r.vec3()
			}
		}
	}
	if f.Version >= Ver1001000 {
		a.Collision = r.i32()
	}
}

// readGeomCommon reads NiGeometryData, the base every geometry block shares.
//
// It stops short of NiTriBasedGeomData's triangle count on purpose. The two
// triangle blocks inherit that class and read the count themselves, but
// NiParticlesData extends NiGeometryData directly and has no triangles at all
// -- reading a count there is two bytes the file never wrote, which is what
// desynchronised logfire.nif.
func (f *File) readGeomCommon(r *reader) (d *ShapeData, nVerts int, err error) {
	d = &ShapeData{}
	v := f.Version
	// The published schema puts a Group ID here from 10.1.0.0 onward, but
	// DAoC's 10.1.0.0 files do not carry it: the block opens directly with
	// the vertex count. Verified by hand on NVikingM.NIF, where reading a
	// Group ID first yields 1 vertex for a head mesh and desynchronises,
	// while skipping it yields 116 vertices and keep/compress flags of zero.
	if v > Ver1010000 {
		d.GroupID = r.i32()
	}
	nVerts = int(r.u16())
	if v >= Ver1010000 {
		r.u8() // keep flags
		r.u8() // compress flags
	}
	if r.boolean() {
		d.Vertices = r.vec3s(nVerts)
	}
	// The UV-set count moves. From 10.0.1.0 it sits between the vertices
	// and the normals; at 4.2.1.0 it comes after the vertex colours
	// instead, immediately ahead of the UV arrays it describes.
	//
	// Derived from npintre1.NIF, where a count read here is 22273 and the
	// byte that should follow it is the first of a unit-length normal. Read
	// after the colours it is 2, which matches the two source textures and
	// the "tree_multitex_alpha" material the shape points at.
	nUV := 0
	if v >= Ver1001000 {
		nUV = int(r.u16())
	}
	if r.boolean() {
		d.Normals = r.vec3s(nVerts)
		// Bit 12 of the 10.x UV-set field says a tangent and a bitangent
		// per vertex follow the normals, for normal mapping. The composite
		// NPC bodies (albbody01heada.nif and its siblings) set it; without
		// these two arrays the read lands mid-normal and every float after
		// is skewed by however many bytes it missed.
		if nUV&0x1000 != 0 {
			r.vec3s(nVerts) // tangents
			r.vec3s(nVerts) // bitangents
		}
	}
	d.Center = r.vec3()
	d.Radius = r.f32()
	if r.boolean() {
		d.Colors = make([][4]float32, nVerts)
		for i := range d.Colors {
			d.Colors[i] = [4]float32{r.f32(), r.f32(), r.f32(), r.f32()}
		}
	}
	if v <= 0x04020200 {
		nUV = int(r.u16())
	}
	// The top nibble of the UV-set count carries flags, the tangent bit
	// handled above among them; the low twelve bits are the count.
	sets := nUV & 0x0FFF
	d.UV = make([][][2]float32, sets)
	for s := 0; s < sets; s++ {
		d.UV[s] = make([][2]float32, nVerts)
		for i := range d.UV[s] {
			d.UV[s][i] = [2]float32{r.f32(), r.f32()}
		}
	}
	if v >= Ver1001000 {
		r.u16() // consistency flags
	}
	if r.err != nil {
		return nil, 0, r.err
	}
	return d, nVerts, nil
}

func (f *File) readShapeData(r *reader) (*ShapeData, error) {
	d, nVerts, err := f.readGeomCommon(r)
	if err != nil {
		return nil, err
	}
	v := f.Version
	nTris := int(r.u16()) // NiTriBasedGeomData

	r.u32() // triangle point count, == nTris*3
	hasTris := true
	if v >= Ver1001000 {
		hasTris = r.boolean()
	}
	if hasTris {
		d.Triangles = make([]Triangle, nTris)
		for i := range d.Triangles {
			d.Triangles[i] = Triangle{r.u16(), r.u16(), r.u16()}
		}
	}
	nMatch := int(r.u16())
	for i := 0; i < nMatch; i++ {
		n := int(r.u16())
		for j := 0; j < n; j++ {
			r.u16()
		}
	}
	if r.err != nil {
		return nil, r.err
	}
	// Sanity: indices must address the vertex array.
	for _, t := range d.Triangles {
		for _, i := range t {
			if int(i) >= nVerts {
				return nil, fmt.Errorf("triangle index %d out of range for %d vertices", i, nVerts)
			}
		}
	}
	return d, nil
}

func (f *File) readSkinData(r *reader) (*SkinData, error) {
	s := &SkinData{}
	s.Skin.Rot = r.mat33()
	s.Skin.Trans = r.vec3()
	s.Skin.Scale = r.f32()
	nBones := int(r.u32())
	if f.Version >= 0x04000002 && f.Version <= Ver1010000 {
		s.Partition = r.i32()
	}
	s.HasVertexWeights = true
	if f.Version >= 0x04020100 {
		s.HasVertexWeights = r.boolean()
	}
	if r.err != nil {
		return nil, r.err
	}
	s.Bones = make([]BoneData, nBones)
	for i := range s.Bones {
		b := &s.Bones[i]
		b.Skin.Rot = r.mat33()
		b.Skin.Trans = r.vec3()
		b.Skin.Scale = r.f32()
		b.Offset = r.vec3()
		b.Radius = r.f32()
		n := int(r.u16())
		if s.HasVertexWeights {
			b.Weights = make([]VertWeight, n)
			for j := range b.Weights {
				b.Weights[j] = VertWeight{r.u16(), r.f32()}
			}
		}
		if r.err != nil {
			return nil, fmt.Errorf("bone %d of %d: %w", i, nBones, r.err)
		}
	}
	return s, nil
}

func (f *File) readSkinPartition(r *reader) (*SkinPartition, error) {
	sp := &SkinPartition{}
	n := int(r.u32())
	if r.err != nil {
		return nil, r.err
	}
	sp.Parts = make([]Partition, n)
	for i := range sp.Parts {
		p := &sp.Parts[i]
		nVerts := int(r.u16())
		nTris := int(r.u16())
		nBones := int(r.u16())
		nStrips := int(r.u16())
		wpv := int(r.u16())
		p.NumVertices, p.WeightsPerVtx = nVerts, wpv
		p.Bones = make([]uint16, nBones)
		for j := range p.Bones {
			p.Bones[j] = r.u16()
		}
		if f.Version < Ver1010000 || r.boolean() {
			p.VertexMap = make([]uint16, nVerts)
			for j := range p.VertexMap {
				p.VertexMap[j] = r.u16()
			}
		}
		if f.Version < Ver1010000 || r.boolean() {
			p.Weights = make([]float32, nVerts*wpv)
			for j := range p.Weights {
				p.Weights[j] = r.f32()
			}
		}
		strip := make([]int, nStrips)
		for j := range strip {
			strip[j] = int(r.u16())
		}
		hasFaces := true
		if f.Version >= Ver1010000 {
			hasFaces = r.boolean()
		}
		if hasFaces {
			if nStrips != 0 {
				p.Strips = make([][]uint16, nStrips)
				for j := range p.Strips {
					p.Strips[j] = make([]uint16, strip[j])
					for k := range p.Strips[j] {
						p.Strips[j][k] = r.u16()
					}
				}
			} else {
				p.Triangles = make([]Triangle, nTris)
				for j := range p.Triangles {
					p.Triangles[j] = Triangle{r.u16(), r.u16(), r.u16()}
				}
			}
		}
		if r.boolean() {
			p.BoneIndices = make([]byte, nVerts*wpv)
			for j := range p.BoneIndices {
				p.BoneIndices[j] = r.u8()
			}
		}
		if r.err != nil {
			return nil, fmt.Errorf("partition %d of %d: %w", i, n, r.err)
		}
	}
	return sp, nil
}

// Key interpolation types, as written in the file.
const (
	keyLinear    = 1
	keyQuadratic = 2
	keyTBC       = 3
	keyXYZ       = 4
)

func (f *File) readKeyframeData(r *reader) (*KeyframeData, error) {
	k := &KeyframeData{}
	nRot := int(r.u32())
	if nRot != 0 {
		k.RotationType = r.u32()
	}
	if k.RotationType == keyXYZ {
		// Per-axis float channels rather than quaternions.
		if f.Version <= Ver1010000 {
			r.f32()
		}
		for i := 0; i < 3; i++ {
			if _, err := f.readFloatKeys(r); err != nil {
				return nil, err
			}
		}
	} else {
		k.Rotations = make([]QuatKey, nRot)
		for i := range k.Rotations {
			k.Rotations[i].Time = r.f32()
			k.Rotations[i].Value = [4]float32{r.f32(), r.f32(), r.f32(), r.f32()}
			if k.RotationType == keyTBC {
				r.f32()
				r.f32()
				r.f32()
			}
		}
	}
	// Translation channel.
	nT := int(r.u32())
	interp := uint32(keyLinear)
	if nT != 0 {
		interp = r.u32()
	}
	k.Translations = make([]VecKey, nT)
	for i := range k.Translations {
		k.Translations[i].Time = r.f32()
		k.Translations[i].Value = r.vec3()
		switch interp {
		case keyQuadratic:
			r.vec3()
			r.vec3()
		case keyTBC:
			r.f32()
			r.f32()
			r.f32()
		}
	}
	// Scale channel.
	sc, err := f.readFloatKeys(r)
	if err != nil {
		return nil, err
	}
	k.Scales = sc
	return k, r.err
}

func (f *File) readFloatKeys(r *reader) ([]FloatKey, error) {
	n := int(r.u32())
	interp := uint32(keyLinear)
	if n != 0 {
		interp = r.u32()
	}
	if r.err != nil {
		return nil, r.err
	}
	out := make([]FloatKey, n)
	for i := range out {
		out[i].Time = r.f32()
		out[i].Value = r.f32()
		switch interp {
		case keyQuadratic:
			r.f32()
			r.f32()
		case keyTBC:
			r.f32()
			r.f32()
			r.f32()
		}
	}
	return out, r.err
}

// --- accessors -----------------------------------------------------------

// NodeAt returns the Node at a block index, or nil.
func (f *File) NodeAt(i int32) *Node {
	if i < 0 || int(i) >= len(f.Blocks) {
		return nil
	}
	n, _ := f.Blocks[i].(*Node)
	return n
}

// ShapeAt returns the TriShape at a block index, or nil.
func (f *File) ShapeAt(i int32) *TriShape {
	if i < 0 || int(i) >= len(f.Blocks) {
		return nil
	}
	s, _ := f.Blocks[i].(*TriShape)
	return s
}

// Block returns the block at an index, or nil when out of range.
func (f *File) Block(i int32) any {
	if i < 0 || int(i) >= len(f.Blocks) {
		return nil
	}
	return f.Blocks[i]
}

// Parents maps each block index to its parent Node, or -1. Derived by walking
// NiNode child lists, which is the only place the hierarchy is recorded.
func (f *File) Parents() []int32 {
	p := make([]int32, len(f.Blocks))
	for i := range p {
		p[i] = -1
	}
	for i, b := range f.Blocks {
		n, ok := b.(*Node)
		if !ok {
			continue
		}
		for _, c := range n.Children {
			if c >= 0 && int(c) < len(p) {
				p[c] = int32(i)
			}
		}
	}
	return p
}

// WorldTransforms accumulates each block's transform down the hierarchy.
// Blocks that are not AVObjects, and any left unreachable, keep Identity.
func (f *File) WorldTransforms() []Transform {
	out := make([]Transform, len(f.Blocks))
	for i := range out {
		out[i] = Identity
	}
	var walk func(i int32, acc Transform)
	walk = func(i int32, acc Transform) {
		n := f.NodeAt(i)
		if n == nil {
			if s := f.ShapeAt(i); s != nil {
				out[i] = acc.Mul(s.Local)
			}
			return
		}
		w := acc.Mul(n.Local)
		out[i] = w
		for _, c := range n.Children {
			if c >= 0 && int(c) < len(out) {
				walk(c, w)
			}
		}
	}
	for _, root := range f.Roots {
		walk(root, Identity)
	}
	return out
}

// --- primitive reader ----------------------------------------------------

// reader walks the byte slice, latching the first error so callers can read a
// whole block and check once. Past an error every read yields zero.
type reader struct {
	// lax disables the non-finite float check. Set only while reading a
	// particle data block: see readParticleData.
	lax bool
	b   []byte
	p   int
	err error
}

func (r *reader) need(n int) bool {
	if r.err != nil {
		return false
	}
	if r.p+n > len(r.b) {
		r.err = errOf(Overrun, "read %d bytes at %d past end of %d", n, r.p, len(r.b))
		return false
	}
	return true
}

func (r *reader) u8() byte {
	if !r.need(1) {
		return 0
	}
	v := r.b[r.p]
	r.p++
	return v
}

func (r *reader) u16() uint16 {
	if !r.need(2) {
		return 0
	}
	v := binary.LittleEndian.Uint16(r.b[r.p:])
	r.p += 2
	return v
}

func (r *reader) u32() uint32 {
	if !r.need(4) {
		return 0
	}
	v := binary.LittleEndian.Uint32(r.b[r.p:])
	r.p += 4
	return v
}

func (r *reader) i32() int32 { return int32(r.u32()) }

func (r *reader) f32() float32 {
	v := math.Float32frombits(r.u32())
	if r.err == nil && !r.lax && (math.IsNaN(float64(v)) || math.IsInf(float64(v), 0)) {
		r.err = fmt.Errorf("non-finite float at %d", r.p-4)
	}
	return v
}

// boolean is one byte from 4.1.0.1 onward, which covers every file here.
func (r *reader) boolean() bool { return r.u8() != 0 }

func (r *reader) str() string {
	n := int(r.u32())
	if n < 0 || n > 1<<16 {
		if r.err == nil {
			r.err = fmt.Errorf("implausible string length %d at %d", n, r.p-4)
		}
		return ""
	}
	if !r.need(n) {
		return ""
	}
	s := string(r.b[r.p : r.p+n])
	r.p += n
	return s
}

func (r *reader) vec3() [3]float32 { return [3]float32{r.f32(), r.f32(), r.f32()} }

func (r *reader) color3() [3]float32 { return r.vec3() }

func (r *reader) mat33() [9]float32 {
	var m [9]float32
	for i := range m {
		m[i] = r.f32()
	}
	return m
}

func (r *reader) vec3s(n int) [][3]float32 {
	if !r.need(n * 12) {
		return nil
	}
	out := make([][3]float32, n)
	for i := range out {
		out[i] = r.vec3()
	}
	return out
}

// refs reads a length-prefixed block-reference array.
func (r *reader) refs() []int32 {
	n := int(r.u32())
	if n < 0 || n > 1<<20 {
		if r.err == nil {
			r.err = fmt.Errorf("implausible array length %d at %d", n, r.p-4)
		}
		return nil
	}
	if !r.need(n * 4) {
		return nil
	}
	out := make([]int32, n)
	for i := range out {
		out[i] = r.i32()
	}
	return out
}

// Transpose flips the rotation, which is how to reinterpret a matrix stored
// in the opposite major order.
func (a Transform) Transpose() Transform {
	var out Transform
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			out.Rot[r*3+c] = a.Rot[c*3+r]
		}
	}
	out.Trans = a.Trans
	out.Scale = a.Scale
	return out
}

// Inverse returns the inverse of a rigid transform with uniform scale.
func (a Transform) Inverse() Transform {
	var out Transform
	s := float32(1)
	if a.Scale != 0 {
		s = 1 / a.Scale
	}
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			out.Rot[r*3+c] = a.Rot[c*3+r]
		}
	}
	out.Scale = s
	// -R^T * t / scale
	for r := 0; r < 3; r++ {
		v := out.Rot[r*3+0]*a.Trans[0] + out.Rot[r*3+1]*a.Trans[1] + out.Rot[r*3+2]*a.Trans[2]
		out.Trans[r] = -v * s
	}
	return out
}

// String renders a transform compactly for diagnostics.
func (a Transform) String() string {
	return fmt.Sprintf("t=(%.2f %.2f %.2f) s=%.3f rot=[%.3f %.3f %.3f | %.3f %.3f %.3f | %.3f %.3f %.3f]",
		a.Trans[0], a.Trans[1], a.Trans[2], a.Scale,
		a.Rot[0], a.Rot[1], a.Rot[2], a.Rot[3], a.Rot[4], a.Rot[5], a.Rot[6], a.Rot[7], a.Rot[8])
}
