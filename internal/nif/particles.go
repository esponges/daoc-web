package nif

import "fmt"

// Particle systems, animation controllers, lights and cameras.
//
// None of this is rendered. It is read because the format gives no choice: a
// block carries no length, so a type that merely exists in a file cannot be
// stepped over, and one unmodelled block stops the walk dead. The campfires
// and torches in Vale of Mularn are ordinary static meshes with a particle
// system hanging off them, and until the particle blocks can be consumed
// exactly, the logs and the brazier cannot be read either.
//
// So these readers exist to get the byte count right. Most of what they read
// is discarded on purpose; the fields are named anyway, because a named field
// that is never used still documents the layout, and the next person to touch
// this needs the layout more than the value.

// Emitter is the common head of the particle geometry blocks. They inherit
// NiGeometry, exactly as NiTriShape does, so the shape itself is unremarkable
// -- it is the data block that differs.
type Emitter struct {
	AVObject
	Data int32
	Skin int32
}

// ParticleData is NiParticlesData and its two subclasses. The per-particle
// arrays are sized by the vertex count of the NiGeometryData it extends.
type ParticleData struct {
	*ShapeData
	Radius    float32
	NumActive int
	Sizes     []float32
	Rotations [][4]float32
}

// Modifier is the shared head of the NiParticleModifier chain. The modifiers
// form a linked list rather than an array, each naming the next.
type Modifier struct {
	Next       int32
	Controller int32
}

// GrowFade is NiParticleGrowFade: how long a particle takes to fade in at
// birth and out at death.
type GrowFade struct {
	Modifier
	Grow float32
	Fade float32
}

// ColorModifier is NiParticleColorModifier, which tints particles over their
// lifetime from a NiColorData curve.
type ColorModifier struct {
	Modifier
	Data int32
}

// Rotation is NiParticleRotation.
type Rotation struct {
	Modifier
	RandomAxis byte
	Axis       [3]float32
	Speed      float32
}

// Bomb is NiParticleBomb: an impulse applied to live particles.
type Bomb struct {
	Modifier
	Decay    float32
	Duration float32
	DeltaV   float32
	Start    float32
	DecayTy  uint32
	Symmetry uint32
	Position [3]float32
	Dir      [3]float32
}

// ColorKey is one entry of a NiColorData curve.
type ColorKey struct {
	Time  float32
	Value [4]float32
}

// ColorData is NiColorData.
type ColorData struct {
	Keys []ColorKey
}

// UVData is NiUVData: four independent curves, for U and V offset and U and V
// tiling. Scrolling textures -- flowing water, a torch flame -- are animated
// this way rather than with moving geometry.
type UVData struct {
	Groups [4][]FloatKey
}

// UVController is NiUVController.
type UVController struct {
	Controller
	UnknownShort uint16
	Data         int32
}

// LookAtController is NiLookAtController, which turns a node to face another.
type LookAtController struct {
	Controller
	Unknown1 uint16
	LookAt   int32
}

// Controller is the NiTimeController head shared by every controller type.
type Controller struct {
	Next      int32
	Flags     uint16
	Frequency float32
	Phase     float32
	StartTime float32
	StopTime  float32
	Target    int32
}

// ParticleSystemController is NiParticleSystemController, the largest block
// in this file by some margin: emitter geometry, rates, lifetimes and the
// live particle array, all of which this pipeline discards.
type ParticleSystemController struct {
	Controller
	Speed        float32
	SpeedRandom  float32
	VertDir      float32
	VertAngle    float32
	HorizDir     float32
	HorizAngle   float32
	Size         float32
	EmitStart    float32
	EmitStop     float32
	EmitRate     float32
	Lifetime     float32
	LifetimeRand float32
	Emitter      int32
	NumParticles int
	NumValid     int
	Extra        int32
}

// Camera is NiCamera. Some models ship the artist's viewport camera.
type Camera struct {
	AVObject
	Frustum  [6]float32
	Viewport [4]float32
	LODAdj   float32
	Scene    int32
}

// Light is NiDirectionalLight and NiAmbientLight, which inherit NiLight
// through NiDynamicEffect.
type Light struct {
	AVObject
	Affected []int32
	Dimmer   float32
	Ambient  [3]float32
	Diffuse  [3]float32
	Specular [3]float32
}

// readController reads the NiTimeController head.
func (f *File) readController(r *reader, c *Controller) {
	c.Next = r.i32()
	c.Flags = r.u16()
	c.Frequency = r.f32()
	c.Phase = r.f32()
	c.StartTime = r.f32()
	c.StopTime = r.f32()
	c.Target = r.i32()
}

// readModifier reads the NiParticleModifier head. The controller back-pointer
// arrives with 4.0.0.2.
func (f *File) readModifier(r *reader, m *Modifier) {
	m.Next = r.i32()
	if f.Version >= 0x04000002 {
		m.Controller = r.i32()
	}
}

// keyGroup reads the count/interpolation/keys triple that NIF uses wherever an
// animated value appears, for a value of the given width in floats. The
// interpolation word is written only when there is at least one key.
func (f *File) keyGroup(r *reader, width int) ([][5]float32, error) {
	n := int(r.u32())
	if n < 0 || n > 1<<20 {
		return nil, fmt.Errorf("implausible key count %d", n)
	}
	interp := uint32(keyLinear)
	if n != 0 {
		interp = r.u32()
	}
	if r.err != nil {
		return nil, r.err
	}
	out := make([][5]float32, n)
	for i := range out {
		out[i][0] = r.f32() // time
		for k := 0; k < width && k < 4; k++ {
			out[i][k+1] = r.f32()
		}
		switch interp {
		case keyQuadratic:
			// forward and backward tangents, one value wide each
			for k := 0; k < 2*width; k++ {
				r.f32()
			}
		case keyTBC:
			r.f32()
			r.f32()
			r.f32()
		}
	}
	return out, r.err
}

func (f *File) parseEffectBlock(r *reader, typ string) (any, error) {
	v := f.Version
	switch typ {
	// The particle shapes are NiGeometry like any other, so their own
	// layout holds no surprises.
	case "NiAutoNormalParticles", "NiRotatingParticles", "NiParticles":
		e := &Emitter{}
		f.readAV(r, &e.AVObject)
		e.Data = r.i32()
		e.Skin = r.i32()
		if v >= Ver1001000 {
			if r.boolean() {
				r.str()
				r.i32()
			}
		}
		return e, nil

	case "NiAutoNormalParticlesData", "NiRotatingParticlesData", "NiParticlesData":
		return f.readParticleData(r, typ)

	case "NiParticleSystemController", "NiBSPArrayController":
		return f.readParticleController(r)

	case "NiParticleGrowFade":
		g := &GrowFade{}
		f.readModifier(r, &g.Modifier)
		g.Grow = r.f32()
		g.Fade = r.f32()
		return g, nil

	case "NiParticleColorModifier":
		c := &ColorModifier{}
		f.readModifier(r, &c.Modifier)
		c.Data = r.i32()
		return c, nil

	case "NiParticleRotation":
		p := &Rotation{}
		f.readModifier(r, &p.Modifier)
		p.RandomAxis = r.u8()
		p.Axis = r.vec3()
		p.Speed = r.f32()
		return p, nil

	case "NiParticleBomb":
		b := &Bomb{}
		f.readModifier(r, &b.Modifier)
		b.Decay = r.f32()
		b.Duration = r.f32()
		b.DeltaV = r.f32()
		b.Start = r.f32()
		b.DecayTy = r.u32()
		if v >= 0x04000002 {
			b.Symmetry = r.u32()
		}
		b.Position = r.vec3()
		b.Dir = r.vec3()
		return b, nil

	case "NiColorData":
		keys, err := f.keyGroup(r, 4)
		if err != nil {
			return nil, err
		}
		c := &ColorData{Keys: make([]ColorKey, len(keys))}
		for i, k := range keys {
			c.Keys[i] = ColorKey{Time: k[0], Value: [4]float32{k[1], k[2], k[3], k[4]}}
		}
		return c, nil

	case "NiUVData":
		d := &UVData{}
		for i := 0; i < 4; i++ {
			g, err := f.readFloatKeys(r)
			if err != nil {
				return nil, err
			}
			d.Groups[i] = g
		}
		return d, nil

	case "NiUVController":
		c := &UVController{}
		f.readController(r, &c.Controller)
		c.UnknownShort = r.u16()
		c.Data = r.i32()
		return c, nil

	case "NiLookAtController":
		c := &LookAtController{}
		f.readController(r, &c.Controller)
		if v >= Ver1010000 {
			c.Unknown1 = r.u16()
		}
		c.LookAt = r.i32()
		return c, nil

	case "NiCamera":
		c := &Camera{}
		f.readAV(r, &c.AVObject)
		if v >= Ver1010000 {
			r.u16()
		}
		for i := range c.Frustum {
			c.Frustum[i] = r.f32()
		}
		// Two gates run the other way from the schema here, and the
		// viewport is what settles it. In HCelttreetower.NIF the four
		// viewport floats read 0, 1, 1, 0 and the frustum 1.0 to 5000.0
		// only if the orthographic flag is absent and the screen-texture
		// count is present; the schema has the first from 4.2.1.0 and the
		// second only from 10.1.0.0.
		if v >= Ver1010000 {
			r.boolean() // use orthographic projection
		}
		for i := range c.Viewport {
			c.Viewport[i] = r.f32()
		}
		c.LODAdj = r.f32()
		c.Scene = r.i32()
		r.u32() // screen polygon count
		r.u32() // screen texture count
		return c, nil

	case "NiDirectionalLight", "NiAmbientLight", "NiPointLight", "NiSpotLight":
		return f.readLight(r, typ)

	default:
		return nil, fmt.Errorf("unmodelled block type %q; blocks carry no length so parsing cannot continue", typ)
	}
}

func (f *File) readParticleData(r *reader, typ string) (*ParticleData, error) {
	// The non-finite float check earns its keep elsewhere -- it is how two
	// layout bugs announced themselves -- but it cannot apply here.
	// "AutoNormal" means the engine generates the normals at runtime, so
	// the array the exporter wrote is never read by the game, and in
	// NDwrfville.nif it contains a NaN among other uninitialised bytes. The
	// footer landing at end-of-file remains the real check on this block.
	r.lax = true
	defer func() { r.lax = false }()

	d, nVerts, err := f.readGeomCommon(r)
	if err != nil {
		return nil, err
	}
	v := f.Version
	p := &ParticleData{ShapeData: d}
	if v <= 0x04000002 {
		r.u16() // particle count, before it became the vertex count
	}
	if v <= Ver1001000 {
		p.Radius = r.f32()
	} else {
		if r.boolean() {
			for i := 0; i < nVerts; i++ {
				r.f32() // per-particle radii
			}
		}
	}
	p.NumActive = int(r.u16())
	if r.boolean() {
		p.Sizes = make([]float32, nVerts)
		for i := range p.Sizes {
			p.Sizes[i] = r.f32()
		}
	}
	if v >= Ver1001000 {
		if r.boolean() {
			p.Rotations = make([][4]float32, nVerts)
			for i := range p.Rotations {
				p.Rotations[i] = [4]float32{r.f32(), r.f32(), r.f32(), r.f32()}
			}
		}
	}
	// NiRotatingParticlesData keeps its own quaternion array at the older
	// versions, where NiParticlesData has none.
	if typ == "NiRotatingParticlesData" && v <= 0x04020200 {
		if r.boolean() {
			p.Rotations = make([][4]float32, nVerts)
			for i := range p.Rotations {
				p.Rotations[i] = [4]float32{r.f32(), r.f32(), r.f32(), r.f32()}
			}
		}
	}
	if r.err != nil {
		return nil, r.err
	}
	return p, nil
}

func (f *File) readParticleController(r *reader) (*ParticleSystemController, error) {
	c := &ParticleSystemController{}
	v := f.Version
	f.readController(r, &c.Controller)
	c.Speed = r.f32()
	c.SpeedRandom = r.f32()
	c.VertDir = r.f32()
	c.VertAngle = r.f32()
	c.HorizDir = r.f32()
	c.HorizAngle = r.f32()
	r.vec3() // unknown normal
	r.f32()
	r.f32()
	r.f32()
	r.f32() // unknown colour
	c.Size = r.f32()
	c.EmitStart = r.f32()
	c.EmitStop = r.f32()
	if v >= 0x04000002 {
		r.u8() // unknown byte
	}
	c.EmitRate = r.f32()
	c.Lifetime = r.f32()
	c.LifetimeRand = r.f32()
	if v >= 0x04000002 {
		r.u16() // emit flags
	}
	r.vec3() // start random
	c.Emitter = r.i32()
	if v >= 0x04000002 {
		r.u16() // unknown short
		r.f32() // unknown float
		r.u32()
		r.u32()
		r.u16()
	}
	c.NumParticles = int(r.u16())
	c.NumValid = int(r.u16())
	if c.NumParticles < 0 || c.NumParticles > 1<<20 {
		return nil, fmt.Errorf("implausible particle count %d", c.NumParticles)
	}
	// The live particle array is a snapshot of the system as the artist
	// saved it: 40 bytes each, and of no use to anything here.
	if !r.need(c.NumParticles * 40) {
		return nil, r.err
	}
	r.p += c.NumParticles * 40
	r.i32() // unknown link
	c.Extra = r.i32()
	r.i32() // unknown link 2
	if v >= 0x04000002 {
		r.u8() // trailer
	}
	if r.err != nil {
		return nil, r.err
	}
	return c, nil
}

func (f *File) readLight(r *reader, typ string) (*Light, error) {
	l := &Light{}
	v := f.Version
	f.readAV(r, &l.AVObject)
	// NiDynamicEffect lists the nodes a light touches, but DAoC's 4.1/4.2
	// files carry no such array: in spikes.NIF the "__MAX_Default_Light"
	// block puts 1.0f where the count would be, and that float is the
	// dimmer. The schema's two alternative arrays, gated at 4.0.0.2 from
	// either side, both read as present at these versions; neither is.
	if v <= 0x04000002 || v >= Ver1001000 {
		l.Affected = r.refs()
	}
	l.Dimmer = r.f32()
	l.Ambient = r.color3()
	l.Diffuse = r.color3()
	l.Specular = r.color3()
	if r.err != nil {
		return nil, r.err
	}
	return l, nil
}
