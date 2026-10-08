package nif

import "fmt"

// The block set used by the world models in zones/Nifs: trees, buildings,
// fences and the rest of the scenery that fixtures.csv places. These are the
// older NetImmerse 4.2.1.0 generation, and they differ from the character
// models in two ways that matter.
//
// They store geometry as triangle strips rather than triangle lists, and they
// carry their own textures: a NiTexturingProperty naming a NiSourceTexture,
// where the character models leave texturing to the equipment system. So a
// prop can be converted without the slot-name guesswork charconv needs.
//
// Everything else here is render state -- z-buffering, alpha, dithering,
// stencil -- which this pipeline mostly ignores. It still has to be read, and
// read exactly: blocks carry no length, so a type that is present cannot be
// skipped even when nothing downstream wants it.

// Property is the shared head of the simple state blocks. Flags is kept
// because NiAlphaProperty's carries the blend mode, which decides whether a
// prop needs alpha blending or alpha testing.
type Property struct {
	ObjectNET
	Flags uint16
}

// AlphaProperty adds the alpha-test threshold. Trees rely on this: the canopy
// is a handful of quads with a cut-out leaf texture.
type AlphaProperty struct {
	Property
	Threshold byte
}

// StencilProperty is read and discarded; nothing here draws stencilled.
type StencilProperty struct {
	Property
	Enabled  byte
	Func     uint32
	Ref      uint32
	Mask     uint32
	Fail     uint32
	ZFail    uint32
	Pass     uint32
	DrawMode uint32
}

// ZBufferProperty controls depth test and write.
type ZBufferProperty struct {
	Property
	Function uint32
}

// TexDesc is one texture slot of a NiTexturingProperty.
type TexDesc struct {
	Source     int32 // block index of a SourceTexture
	ClampMode  uint32
	FilterMode uint32
	UVSet      uint32
}

// Texturing is NiTexturingProperty. Only the base slot is wired through to the
// converter; the rest are read so the block's length comes out right.
type Texturing struct {
	Property
	ApplyMode uint32
	Count     uint32
	HasBase   bool
	Base      TexDesc
}

// SourceTexture is NiSourceTexture: for DAoC's world models always an external
// reference, naming a .dds that sits beside the .npk in zones/Nifs.
type SourceTexture struct {
	ObjectNET
	UseExternal byte
	FileName    string
	PixelData   int32
	PixelLayout uint32
	UseMipmaps  uint32
	AlphaFormat uint32
	IsStatic    byte
}

// StringExtra is NiStringExtraData, which the exporter uses for tagging.
type StringExtra struct {
	Next   int32
	Length uint32
	Value  string
}

// LODNode is NiLODNode: picks one child by camera distance. The converter
// takes the nearest range, since props are only ever seen from the ground.
type LODNode struct {
	Node
	Center [3]float32
	Ranges [][2]float32
	Data   int32
}

// parseWorldBlock handles the types the scenery models add on top of the
// character block set. It is the default arm of parseBlock, so an unmodelled
// type still fails loudly here.
func (f *File) parseWorldBlock(r *reader, typ string) (any, error) {
	v := f.Version
	switch typ {
	// NiTriStrips is NiTriShape's sibling: same fields, different topology
	// in the data block it points at.
	case "NiTriStrips":
		s := &TriShape{}
		f.readAV(r, &s.AVObject)
		s.Data = r.i32()
		s.Skin = r.i32()
		if v >= Ver1001000 {
			s.HasShdr = r.boolean()
			if s.HasShdr {
				s.Shader = r.str()
				s.ShdrUnk = r.i32()
			}
		}
		return s, nil

	case "NiTriStripsData":
		return f.readStripsData(r)

	case "NiZBufferProperty":
		p := &ZBufferProperty{}
		f.readNET(r, &p.ObjectNET)
		p.Flags = r.u16()
		// The compare function is a separate field from 4.1.0.12 until the
		// flags absorb it in later versions.
		if v >= 0x0401000C && v <= 0x14000005 {
			p.Function = r.u32()
		}
		return p, nil

	case "NiDitherProperty", "NiSpecularProperty", "NiShadeProperty", "NiWireframeProperty":
		p := &Property{}
		f.readNET(r, &p.ObjectNET)
		p.Flags = r.u16()
		return p, nil

	case "NiAlphaProperty":
		p := &AlphaProperty{}
		f.readNET(r, &p.ObjectNET)
		p.Flags = r.u16()
		p.Threshold = r.u8()
		return p, nil

	case "NiStencilProperty":
		p := &StencilProperty{}
		f.readNET(r, &p.ObjectNET)
		if v <= 0x0A000102 {
			p.Flags = r.u16()
		}
		if v <= 0x14000005 {
			p.Enabled = r.u8()
			p.Func = r.u32()
			p.Ref = r.u32()
			p.Mask = r.u32()
			p.Fail = r.u32()
			p.ZFail = r.u32()
			p.Pass = r.u32()
			p.DrawMode = r.u32()
		}
		return p, nil

	case "NiTexturingProperty":
		return f.readTexturing(r)

	case "NiSourceTexture":
		t := &SourceTexture{}
		f.readNET(r, &t.ObjectNET)
		t.UseExternal = r.u8()
		if t.UseExternal == 1 {
			t.FileName = r.str()
			if v >= Ver1010000 {
				r.i32() // unknown link
			}
		} else {
			if v <= Ver1001000 {
				r.u8() // unknown byte
			}
			if v >= Ver1010000 {
				t.FileName = r.str()
			}
			t.PixelData = r.i32()
		}
		t.PixelLayout = r.u32()
		t.UseMipmaps = r.u32()
		t.AlphaFormat = r.u32()
		t.IsStatic = r.u8()
		// The schema adds a Direct Render flag at 10.1.0.0; DAoC's files do
		// not carry it. In NF_b-fence1.NIF the texture block holding
		// "A_oldwood.dds" ends one byte before the flag would, and reading
		// it swallows the first byte of the next block's separator.
		return t, nil

	case "NiStringExtraData":
		e := &StringExtra{}
		if v >= Ver1001000 {
			e.Value = r.str() // name, then the value below
		} else {
			e.Next = r.i32()
		}
		if v <= 0x04020200 {
			e.Length = r.u32()
		}
		e.Value = r.str()
		return e, nil

	// The integer tags the 10.1.0.0 creature exporter attaches: "Arborist"
	// and the like, a count and that many values. Nothing downstream reads
	// them, but over a hundred figures carry one, and with no block length
	// there is no stepping over them unread.
	case "NiIntegersExtraData":
		f.readExtraPrefix(r)
		n := int(r.u32())
		if !r.need(n * 4) {
			return nil, r.err
		}
		vals := make([]uint32, n)
		for i := range vals {
			vals[i] = r.u32()
		}
		return vals, nil

	case "NiIntegerExtraData":
		f.readExtraPrefix(r)
		return r.u32(), nil

	case "NiBooleanExtraData":
		f.readExtraPrefix(r)
		return r.boolean(), nil

	case "NiFloatExtraData":
		f.readExtraPrefix(r)
		return r.f32(), nil

	case "NiColorExtraData":
		f.readExtraPrefix(r)
		return [4]float32{r.f32(), r.f32(), r.f32(), r.f32()}, nil

	// Text keys mark moments in an animation: "start", "hit", "end".
	case "NiTextKeyExtraData":
		f.readExtraPrefix(r)
		n := int(r.u32())
		keys := make(map[float32]string, n)
		for i := 0; i < n && r.err == nil; i++ {
			t := r.f32()
			keys[t] = r.str()
		}
		return keys, nil

	// A billboard always faces the camera. Before 10.1.0.0 the mode is
	// implicit, so the block is exactly a NiNode.
	case "NiBillboardNode":
		n := &Node{}
		f.readAV(r, &n.AVObject)
		n.Children = r.refs()
		n.Effects = r.refs()
		if v >= Ver1010000 {
			r.u16() // billboard mode
		}
		return n, nil

	case "NiLODNode":
		l := &LODNode{}
		f.readAV(r, &l.Node.AVObject)
		l.Node.Children = r.refs()
		l.Node.Effects = r.refs()
		// NiSwitchNode sits between NiLODNode and NiNode. Its flags are
		// 10.1.0.0 and later, but the active-child index is present at
		// every version -- including 4.2.1.0, where the schema's version
		// gating reads as though the whole class were absent.
		//
		// Derived from npintre1.NIF, whose LOD node has two children
		// (blocks 5 and 13) and two ranges, 0..6575.7 and 6575.7..100000.
		// Those only line up if the centre begins four bytes further on.
		if v >= Ver1010000 {
			r.u16() // flags
		}
		r.i32() // active child index
		// From 10.1.0.0 the ranges move out into their own block and only
		// a reference is left here. Zone 100 uses both generations: the
		// trees are 4.x with inline ranges, the fences are 10.1.0.0.
		if v >= Ver1010000 {
			l.Data = r.i32()
			return l, nil
		}
		l.Center = r.vec3()
		n := int(r.u32())
		if n < 0 || n > 1<<16 {
			return nil, fmt.Errorf("implausible LOD level count %d", n)
		}
		l.Ranges = make([][2]float32, n)
		for i := range l.Ranges {
			l.Ranges[i] = [2]float32{r.f32(), r.f32()}
		}
		return l, nil

	// NiScreenLODData picks a level by how much of the screen the prop
	// covers rather than by distance. The proportions are read and ignored:
	// this pipeline always draws the nearest level.
	case "NiScreenLODData":
		l := &LODNode{}
		l.Center = r.vec3()
		r.f32()  // bound radius
		r.vec3() // world centre
		r.f32()  // world radius
		n := int(r.u32())
		if n < 0 || n > 1<<16 {
			return nil, fmt.Errorf("implausible LOD proportion count %d", n)
		}
		l.Ranges = make([][2]float32, n)
		for i := range l.Ranges {
			l.Ranges[i] = [2]float32{r.f32(), 0}
		}
		return l, nil

	// NiRangeLODData holds what NiLODNode used to carry inline.
	case "NiRangeLODData":
		l := &LODNode{}
		l.Center = r.vec3()
		n := int(r.u32())
		if n < 0 || n > 1<<16 {
			return nil, fmt.Errorf("implausible LOD level count %d", n)
		}
		l.Ranges = make([][2]float32, n)
		for i := range l.Ranges {
			l.Ranges[i] = [2]float32{r.f32(), r.f32()}
		}
		return l, nil

	default:
		return f.parseEffectBlock(r, typ)
	}
}

func (f *File) readTexturing(r *reader) (*Texturing, error) {
	t := &Texturing{}
	v := f.Version
	f.readNET(r, &t.ObjectNET)
	if v <= 0x0A000102 || v >= 0x14010001 {
		t.Flags = r.u16()
	}
	if v <= 0x14000005 {
		t.ApplyMode = r.u32()
	}
	t.Count = r.u32()

	// The slots are a fixed sequence, each an optional TexDesc. Only the
	// base one carries the diffuse map this pipeline wants, but every
	// present slot must still be consumed.
	readSlot := func() TexDesc {
		var d TexDesc
		if r.boolean() {
			d = f.readTexDesc(r)
		}
		return d
	}
	t.HasBase = false
	if r.boolean() {
		t.HasBase = true
		t.Base = f.readTexDesc(r)
	}
	readSlot() // dark
	readSlot() // detail
	readSlot() // gloss
	readSlot() // glow
	if t.Count > 5 {
		if r.boolean() { // bump map
			f.readTexDesc(r)
			r.f32() // luma scale
			r.f32() // luma offset
			for i := 0; i < 4; i++ {
				r.f32() // bump map matrix
			}
		}
	}
	if t.Count > 6 {
		readSlot() // decal 0
	}
	if t.Count > 7 {
		readSlot() // decal 1
	}
	if v >= Ver1001000 {
		n := int(r.u32()) // shader textures
		for i := 0; i < n; i++ {
			if r.boolean() {
				f.readTexDesc(r)
				r.u32()
			}
		}
	}
	if r.err != nil {
		return nil, r.err
	}
	return t, nil
}

func (f *File) readTexDesc(r *reader) TexDesc {
	var d TexDesc
	v := f.Version
	d.Source = r.i32()
	if v <= 0x14000005 {
		d.ClampMode = r.u32()
		d.FilterMode = r.u32()
		d.UVSet = r.u32()
	}
	if v <= 0x0A040001 {
		r.u16() // PS2 L
		r.u16() // PS2 K
	}
	if v <= 0x0401000C {
		r.u16() // unknown short, pre-4.2
	}
	if v >= Ver1010000 {
		if r.boolean() { // texture transform
			r.f32()
			r.f32() // translation
			r.f32()
			r.f32() // tiling
			r.f32() // w rotation
			r.u32() // transform type
			r.f32()
			r.f32() // centre offset
		}
	}
	return d
}

// readStripsData reads NiTriStripsData and converts the strips to a triangle
// list. A strip alternates winding with every step, so odd-indexed triangles
// are emitted reversed to keep the whole mesh consistently wound; degenerate
// triangles, which strips use to stitch runs together, are dropped.
func (f *File) readStripsData(r *reader) (*ShapeData, error) {
	d, nVerts, err := f.readGeomCommon(r)
	if err != nil {
		return nil, err
	}
	v := f.Version
	nTris := int(r.u16()) // NiTriBasedGeomData
	nStrips := int(r.u16())
	if nStrips < 0 || nStrips > 1<<16 {
		return nil, fmt.Errorf("implausible strip count %d", nStrips)
	}
	lengths := make([]int, nStrips)
	for i := range lengths {
		lengths[i] = int(r.u16())
	}
	hasPoints := true
	if v >= 0x0A000103 {
		hasPoints = r.boolean()
	}
	if r.err != nil {
		return nil, r.err
	}
	if !hasPoints {
		return d, nil
	}
	d.Strips = make([][]uint16, nStrips)
	for i, n := range lengths {
		if !r.need(n * 2) {
			return nil, r.err
		}
		s := make([]uint16, n)
		for j := range s {
			s[j] = r.u16()
			if int(s[j]) >= nVerts {
				return nil, fmt.Errorf("strip %d point %d indexes vertex %d of %d", i, j, s[j], nVerts)
			}
		}
		d.Strips[i] = s
	}
	for _, s := range d.Strips {
		for i := 0; i+2 < len(s); i++ {
			a, b, c := s[i], s[i+1], s[i+2]
			if a == b || b == c || a == c {
				continue // degenerate: a stitch between runs
			}
			if i%2 == 1 {
				b, c = c, b
			}
			d.Triangles = append(d.Triangles, Triangle{a, b, c})
		}
	}
	// The declared count is of real triangles, so it is a lower bound on
	// what the strips yield once degenerates are dropped -- a useful check
	// that the strip data was read at the right offset.
	if nTris > 0 && len(d.Triangles) > nTris {
		return nil, fmt.Errorf("strips yielded %d triangles but the block declares %d", len(d.Triangles), nTris)
	}
	return d, nil
}

// readExtraPrefix reads the NiExtraData fields every extra-data block opens
// with: a name from 10.0.1.0, the next link in the chain before that.
func (f *File) readExtraPrefix(r *reader) {
	if f.Version >= Ver1001000 {
		r.str()
	} else {
		r.i32()
	}
}
