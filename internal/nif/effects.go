package nif

// The block types the census found blocking whole files: projected and
// environment textures, particle gravity, texture, alpha, colour, visibility
// and path animation, vertex morphs, fog, and textures stored inside the
// model. Each is read exactly, so the files holding them parse; only
// gravity and the embedded textures change what is drawn.

import "fmt"

// TextureEffect is NiTextureEffect: a texture projected onto, or reflected
// from, the nodes it affects -- the dungeons' wet-stone shine, the crystals'
// environment maps. Read, not drawn.
type TextureEffect struct {
	AVObject
	Affected    []int32
	ProjMatrix  [9]float32
	ProjTrans   [3]float32
	Filter      uint32
	Clamp       uint32
	EffectType  uint32 // 0 projected light, 1 projected shadow, 2 environment map, 3 fog map
	CoordGen    uint32
	Source      int32 // NiSourceTexture
	EnablePlane byte
	Plane       [4]float32 // normal, then constant
}

// Gravity is NiGravity, a particle modifier pulling particles along a
// direction (planar) or towards a point (spherical).
type Gravity struct {
	Modifier
	Decay     float32
	Force     float32
	Type      uint32 // 0 planar, 1 spherical
	Position  [3]float32
	Direction [3]float32
}

// TextureTransformController is NiTextureTransformController: one of a
// texture slot's UV transform members over time. Read, not drawn.
type TextureTransformController struct {
	Controller
	ShaderMap bool
	Slot      uint32 // the texturing property's slot: 0 base, 1 dark, ...
	Operation uint32 // 0 translate U, 1 translate V, 2 rotate, 3 scale U, 4 scale V
	Data      int32  // NiFloatData
}

// AlphaController is NiAlphaController: a material's alpha over time. Read,
// not drawn.
type AlphaController struct {
	Controller
	Data int32 // NiFloatData
}

// MaterialColorController is NiMaterialColorController: one of a material's
// colours over time. Read, not drawn.
type MaterialColorController struct {
	Controller
	TargetColor uint16 // from 10.1.0.0; before, it rides in the flags
	Data        int32  // NiPosData
}

// VisController is NiVisController: a node shown and hidden over time. Read,
// not drawn.
type VisController struct {
	Controller
	Data int32 // NiVisData
}

// PathController is NiPathController: a node moved along a curve. Read, not
// drawn.
type PathController struct {
	Controller
	PathFlags    uint16
	BankDir      int32
	MaxBankAngle float32
	Smoothing    float32
	FollowAxis   int16
	PathData     int32 // NiPosData
	PercentData  int32 // NiFloatData
}

// GeomMorpherController is NiGeomMorpherController: a shape blended between
// vertex targets over time. Read, not drawn; the shape's own vertices are
// its base pose and draw as they are.
type GeomMorpherController struct {
	Controller
	MorpherFlags uint16
	Data         int32 // NiMorphData
	AlwaysUpdate byte
}

// Collider is NiPlanarCollider and NiSphericalCollider: particles bouncing
// off a plane or a sphere. Read, not drawn.
type Collider struct {
	Modifier
	Bounce       float32
	SpawnOnHit   bool
	DieOnHit     bool
	Height       float32 // planar only, as are Width, XAxis, YAxis and Plane
	Width        float32
	Radius       float32 // spherical only
	Position     [3]float32
	XAxis, YAxis [3]float32
	Plane        [4]float32
}

// LightColorController is NiLightColorController: a light's colour over
// time. Read, not drawn.
type LightColorController struct {
	Controller
	TargetColor uint16
	Data        int32 // NiPosData
}

// FlipController is NiFlipController: a texture slot cycling through a list
// of textures. Read, not drawn.
type FlipController struct {
	Controller
	Slot    uint32
	Delta   float32
	Sources []int32 // NiSourceTexture
}

// FloatData is NiFloatData, a float curve.
type FloatData struct{ Keys []FloatKey }

// PosData is NiPosData, a Vector3 curve.
type PosData struct{ Keys []VecKey }

// VisKey is a visibility keyframe.
type VisKey struct {
	Time  float32
	Value byte
}

// VisData is NiVisData, an on/off curve.
type VisData struct{ Keys []VisKey }

// Morph is one NiMorphData target: its weight over time and a vertex offset
// (or position) per vertex.
type Morph struct {
	Keys    []FloatKey
	Vectors [][3]float32
}

// MorphData is NiMorphData.
type MorphData struct {
	NumVertices int
	Relative    byte
	Morphs      []Morph
}

// FogProperty is NiFogProperty. Read, not drawn: the zone's own fog covers
// what it would.
type FogProperty struct {
	ObjectNET
	Flags uint16
	Depth float32
	Color [3]float32
}

// MipMap is one level of an NiPixelData image.
type MipMap struct {
	Width, Height, Offset uint32
}

// PixelData is NiPixelData: a texture stored inside the model rather than
// beside it as a file.
type PixelData struct {
	Format        uint32 // 0 RGB, 1 RGBA, 2 palettised, 4 DXT1, 5 DXT5
	Masks         [4]uint32
	BitsPerPixel  byte
	Palette       int32 // NiPalette, for format 2
	Mipmaps       []MipMap
	BytesPerPixel uint32
	Pixels        []byte
}

// Palette is NiPalette: up to 256 RGBA entries.
type Palette struct {
	HasAlpha byte
	Colors   [][4]byte
}

// readDynamicEffect reads NiDynamicEffect's affected-node list where the
// files carry it; see readLight for why 4.1 and 4.2 do not.
func (f *File) readDynamicEffect(r *reader) []int32 {
	if f.Version <= 0x04000002 || f.Version >= Ver1001000 {
		return r.refs()
	}
	return nil
}

func (f *File) parseCensusBlock(r *reader, typ string) (any, error) {
	v := f.Version
	switch typ {
	case "NiTextureEffect":
		e := &TextureEffect{}
		f.readAV(r, &e.AVObject)
		e.Affected = f.readDynamicEffect(r)
		e.ProjMatrix = r.mat33()
		e.ProjTrans = r.vec3()
		e.Filter = r.u32()
		e.Clamp = r.u32()
		e.EffectType = r.u32()
		e.CoordGen = r.u32()
		e.Source = r.i32()
		e.EnablePlane = r.u8()
		for i := range e.Plane {
			e.Plane[i] = r.f32()
		}
		r.u16() // PS2 L
		r.u16() // PS2 K
		if v <= 0x0401000C {
			r.u16() // unknown
		}
		return e, nil

	case "NiGravity":
		g := &Gravity{}
		f.readModifier(r, &g.Modifier)
		if v >= 0x04000002 {
			g.Decay = r.f32()
		}
		g.Force = r.f32()
		g.Type = r.u32()
		g.Position = r.vec3()
		g.Direction = r.vec3()
		return g, nil

	case "NiTextureTransformController":
		c := &TextureTransformController{}
		f.readController(r, &c.Controller)
		c.ShaderMap = r.boolean()
		c.Slot = r.u32()
		c.Operation = r.u32()
		c.Data = r.i32()
		return c, nil

	case "NiAlphaController":
		c := &AlphaController{}
		f.readController(r, &c.Controller)
		c.Data = r.i32()
		return c, nil

	case "NiMaterialColorController":
		c := &MaterialColorController{}
		f.readController(r, &c.Controller)
		if v >= Ver1010000 {
			c.TargetColor = r.u16()
		}
		c.Data = r.i32()
		return c, nil

	case "NiVisController":
		c := &VisController{}
		f.readController(r, &c.Controller)
		c.Data = r.i32()
		return c, nil

	case "NiPathController":
		c := &PathController{}
		f.readController(r, &c.Controller)
		if v >= Ver1010000 {
			c.PathFlags = r.u16()
		}
		c.BankDir = r.i32()
		c.MaxBankAngle = r.f32()
		c.Smoothing = r.f32()
		c.FollowAxis = int16(r.u16())
		c.PathData = r.i32()
		c.PercentData = r.i32()
		return c, nil

	case "NiGeomMorpherController":
		c := &GeomMorpherController{}
		f.readController(r, &c.Controller)
		if v >= 0x0A000102 {
			c.MorpherFlags = r.u16()
		}
		c.Data = r.i32()
		c.AlwaysUpdate = r.u8()
		return c, nil

	case "NiPlanarCollider", "NiSphericalCollider":
		c := &Collider{}
		f.readModifier(r, &c.Modifier)
		c.Bounce = r.f32()
		if v >= 0x04020002 {
			c.SpawnOnHit = r.boolean()
			c.DieOnHit = r.boolean()
		}
		if typ == "NiSphericalCollider" {
			c.Radius = r.f32()
			c.Position = r.vec3()
			return c, nil
		}
		c.Height = r.f32()
		c.Width = r.f32()
		c.Position = r.vec3()
		c.XAxis = r.vec3()
		c.YAxis = r.vec3()
		for i := range c.Plane {
			c.Plane[i] = r.f32()
		}
		return c, nil

	case "NiLightColorController":
		c := &LightColorController{}
		f.readController(r, &c.Controller)
		if v >= Ver1010000 {
			c.TargetColor = r.u16()
		}
		c.Data = r.i32()
		return c, nil

	case "NiFlipController":
		c := &FlipController{}
		f.readController(r, &c.Controller)
		c.Slot = r.u32()
		if v >= Ver1010000 {
			r.f32() // accumulated time; zero in ghostKing01.nif
		}
		c.Delta = r.f32()
		c.Sources = r.refs()
		return c, nil

	case "NiVectorExtraData":
		f.readExtraPrefix(r)
		return [4]float32{r.f32(), r.f32(), r.f32(), r.f32()}, nil

	case "NiBinaryExtraData":
		f.readExtraPrefix(r)
		n := int(r.u32())
		if n < 0 || !r.need(n) {
			if r.err == nil {
				return nil, fmt.Errorf("implausible binary data size %d", n)
			}
			return nil, r.err
		}
		b := append([]byte(nil), r.b[r.p:r.p+n]...)
		r.p += n
		return b, nil

	case "NiFloatData":
		keys, err := f.readFloatKeys(r)
		if err != nil {
			return nil, err
		}
		return &FloatData{Keys: keys}, nil

	case "NiPosData":
		keys, err := f.keyGroup(r, 3)
		if err != nil {
			return nil, err
		}
		d := &PosData{Keys: make([]VecKey, len(keys))}
		for i, k := range keys {
			d.Keys[i] = VecKey{Time: k[0], Value: [3]float32{k[1], k[2], k[3]}}
		}
		return d, nil

	case "NiVisData":
		n := int(r.u32())
		if n < 0 || n > 1<<20 {
			return nil, fmt.Errorf("implausible key count %d", n)
		}
		d := &VisData{Keys: make([]VisKey, n)}
		for i := range d.Keys {
			d.Keys[i] = VisKey{Time: r.f32(), Value: r.u8()}
		}
		return d, nil

	case "NiMorphData":
		return f.readMorphData(r)

	case "NiFogProperty":
		p := &FogProperty{}
		f.readNET(r, &p.ObjectNET)
		p.Flags = r.u16()
		p.Depth = r.f32()
		p.Color = r.color3()
		return p, nil

	case "NiPixelData":
		return f.readPixelData(r)

	case "NiPalette":
		p := &Palette{}
		p.HasAlpha = r.u8()
		n := int(r.u32())
		if n < 0 || n > 256 {
			return nil, fmt.Errorf("implausible palette size %d", n)
		}
		p.Colors = make([][4]byte, n)
		for i := range p.Colors {
			p.Colors[i] = [4]byte{r.u8(), r.u8(), r.u8(), r.u8()}
		}
		return p, nil

	default:
		return nil, errOf(Unsupported, "unmodelled block type %q; blocks carry no length so parsing cannot continue", typ)
	}
}

func (f *File) readMorphData(r *reader) (*MorphData, error) {
	nMorphs := int(r.u32())
	nVerts := int(r.u32())
	if nMorphs < 0 || nMorphs > 1<<10 || nVerts < 0 || nVerts > 1<<16 {
		return nil, fmt.Errorf("implausible morph counts %d x %d", nMorphs, nVerts)
	}
	d := &MorphData{NumVertices: nVerts, Relative: r.u8(), Morphs: make([]Morph, nMorphs)}
	for i := range d.Morphs {
		m := &d.Morphs[i]
		// Up to 10.1.0.0 each target carries its own weight curve, with
		// the interpolation word written even when there are no keys.
		n := int(r.u32())
		if n < 0 || n > 1<<20 {
			return nil, fmt.Errorf("implausible key count %d", n)
		}
		interp := r.u32()
		m.Keys = make([]FloatKey, n)
		// The base target's weight keys can be junk: kelpnon.nif's run
		// -FLT_MAX, a denormal and a NaN, with sound vertices after them.
		// The weights are not drawn, so they are read without the
		// non-finite check; the vectors keep it.
		r.lax = true
		for k := range m.Keys {
			m.Keys[k] = FloatKey{Time: r.f32(), Value: r.f32()}
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
		r.lax = false
		m.Vectors = r.vec3s(nVerts)
		if r.err != nil {
			return nil, r.err
		}
	}
	return d, nil
}

func (f *File) readPixelData(r *reader) (*PixelData, error) {
	p := &PixelData{}
	p.Format = r.u32()
	for i := range p.Masks {
		p.Masks[i] = r.u32()
	}
	p.BitsPerPixel = r.u8()
	// Three unknown bytes, then eight fast-compare bytes. Checked on
	// avpier.nif: with both, the palette reference lands on the very next
	// block and the single 128x128 level accounts for every pixel byte.
	for i := 0; i < 3+8; i++ {
		r.u8()
	}
	if f.Version >= Ver1010000 {
		r.u32() // unknown
	}
	p.Palette = r.i32()
	n := int(r.u32())
	if n < 0 || n > 32 {
		return nil, fmt.Errorf("implausible mipmap count %d", n)
	}
	p.BytesPerPixel = r.u32()
	p.Mipmaps = make([]MipMap, n)
	for i := range p.Mipmaps {
		p.Mipmaps[i] = MipMap{r.u32(), r.u32(), r.u32()}
	}
	size := int(r.u32())
	if size < 0 || size > 1<<26 || !r.need(size) {
		if r.err == nil {
			return nil, fmt.Errorf("implausible pixel data size %d", size)
		}
		return nil, r.err
	}
	p.Pixels = append([]byte(nil), r.b[r.p:r.p+size]...)
	r.p += size
	return p, nil
}
