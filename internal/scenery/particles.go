package scenery

// Particle systems: the fire and smoke on the log fires, the forge sparks,
// the portals' streams.
//
// A NiParticleSystemController holds everything about an emitter but where
// its particles are: a rate, a lifetime, a speed, a direction cone, a size
// and the box particles are born in, plus a chain of modifiers that grow and
// fade them, tint them over their lives and spin them. The controller points
// at the emitter node, whose transform places and aims the cone, and at the
// particle geometry, whose properties give the texture and blending.
//
// None of it is simulated here. The settings are written, in the model's
// own space, and the viewer runs them.

import (
	"math"

	"daocweb/internal/nif"
)

// Particle is one emitter's settings, for the viewer to run.
type Particle struct {
	Texture  string     `json:"texture"`
	Additive bool       `json:"additive"` // else blended over what is behind
	Origin   [3]float32 `json:"origin"`   // emitter position, model space
	Axes     [9]float32 `json:"axes"`     // emitter rotation, row-major, model space
	Box      [3]float32 `json:"box"`      // half-size of the birth box, emitter space
	Rate     float32    `json:"rate"`     // particles a second
	Life     [2]float32 `json:"life"`     // seconds, base and spread
	Speed    [2]float32 `json:"speed"`    // units a second, base and spread
	// The cone: vertical is the angle from the emitter's +Z (0 straight up,
	// pi straight down), horizontal the turn about Z; each a direction and
	// a half-width, radians.
	Vertical   [2]float32 `json:"vertical"`
	Horizontal [2]float32 `json:"horizontal"`
	Size       float32    `json:"size"`
	Grow       float32    `json:"grow"` // seconds from nothing to full size
	Fade       float32    `json:"fade"` // seconds from full size to nothing, at the end
	Spin       float32    `json:"spin"` // radians a second
	// Colour over the particle's life, [age fraction, r, g, b, a] keys; it
	// multiplies the texture.
	Color [][5]float32 `json:"color,omitempty"`
	// Gravity is the particles' acceleration, model space, units a second
	// squared: the sum of the system's NiGravity modifiers. It is constant,
	// so a particle's offset is half of it times its age squared. Planar
	// gravity is exact. Spherical gravity pulls towards a point and so
	// changes direction as a particle moves; it is taken as it acts at the
	// emitter, which holds for particles that stay near their source.
	// Either kind fades with distance by its decay, also taken at the
	// emitter.
	Gravity *[3]float32 `json:"gravity,omitempty"`
}

// particleSystems lists a model's emitters. Systems whose emitter or
// geometry is missing, or that emit nothing, are skipped.
func particleSystems(tn *texNamer) []Particle {
	f := tn.f
	world := f.WorldTransforms()
	var out []Particle
	for _, blk := range f.Blocks {
		c, ok := blk.(*nif.ParticleSystemController)
		if !ok || c.EmitRate <= 0 || c.Lifetime <= 0 {
			continue
		}
		geo, ok := f.Block(c.Target).(*nif.Emitter)
		if !ok || int(c.Emitter) < 0 || int(c.Emitter) >= len(world) {
			continue
		}
		p := Particle{
			Box:        c.StartRandom,
			Rate:       c.EmitRate,
			Life:       [2]float32{c.Lifetime, c.LifetimeRand},
			Speed:      [2]float32{c.Speed, c.SpeedRandom},
			Vertical:   [2]float32{c.VertDir, c.VertAngle},
			Horizontal: [2]float32{c.HorizDir, c.HorizAngle},
			Size:       c.Size,
		}
		// Modifiers are in the space of the particle geometry's node.
		frame := nif.Identity
		if int(c.Target) >= 0 && int(c.Target) < len(world) {
			frame = world[c.Target]
		}
		w := world[c.Emitter]
		p.Origin = w.Trans
		p.Axes = w.Rot
		// Scale on the emitter's path scales the whole effect.
		if w.Scale > 0 && w.Scale != 1 {
			s := w.Scale
			p.Box = [3]float32{p.Box[0] * s, p.Box[1] * s, p.Box[2] * s}
			p.Speed = [2]float32{p.Speed[0] * s, p.Speed[1] * s}
			p.Size *= s
		}
		for _, ref := range geo.Properties {
			switch b := f.Block(ref).(type) {
			case *nif.Texturing:
				if src, ok := f.Block(b.Base.Source).(*nif.SourceTexture); ok && b.HasBase {
					p.Texture = tn.name(src)
				}
			case *nif.AlphaProperty:
				p.Additive = b.Flags&1 != 0 && (b.Flags>>5)&15 == 0
			}
		}
		if p.Texture == "" {
			continue
		}
		// The modifier chain.
		for m, n := c.Extra, 0; m >= 0 && n < 16; n++ {
			switch mm := f.Block(m).(type) {
			case *nif.GrowFade:
				p.Grow, p.Fade = mm.Grow, mm.Fade
				m = mm.Next
			case *nif.ColorModifier:
				if d, ok := f.Block(mm.Data).(*nif.ColorData); ok {
					for _, k := range d.Keys {
						p.Color = append(p.Color, [5]float32{k.Time, k.Value[0], k.Value[1], k.Value[2], k.Value[3]})
					}
				}
				m = mm.Next
			case *nif.Rotation:
				p.Spin = mm.Speed
				m = mm.Next
			case *nif.Gravity:
				g := gravityAt(mm, frame, p.Origin)
				if p.Gravity == nil {
					p.Gravity = &[3]float32{}
				}
				for k := 0; k < 3; k++ {
					p.Gravity[k] += g[k]
				}
				m = mm.Next
			case *nif.Bomb:
				// An impulse pushing live particles about; not modelled.
				m = mm.Next
			case *nif.Collider:
				// Particles bouncing off a plane or sphere; not modelled.
				m = mm.Next
			default:
				m = -1
			}
		}
		// The colour keys are age fractions; clamp any stray past the ends.
		for i := range p.Color {
			p.Color[i][0] = float32(math.Max(0, math.Min(1, float64(p.Color[i][0]))))
		}
		out = append(out, p)
	}
	return out
}

// gravityAt is one NiGravity's acceleration on a particle at origin, in
// model space; frame takes the modifier's own space there.
func gravityAt(g *nif.Gravity, frame nif.Transform, origin [3]float32) [3]float32 {
	pos := frame.Apply(g.Position)
	var dir [3]float64
	var dist float64
	switch g.Type {
	case 1: // spherical: towards the point
		for k := 0; k < 3; k++ {
			dir[k] = float64(pos[k] - origin[k])
		}
		dist = math.Sqrt(dir[0]*dir[0] + dir[1]*dir[1] + dir[2]*dir[2])
	default: // planar: along the direction, decaying away from its plane
		d := rotate(frame, g.Direction)
		for k := 0; k < 3; k++ {
			dir[k] = float64(d[k])
			dist += float64(origin[k]-pos[k]) * dir[k]
		}
		dist = math.Abs(dist)
	}
	n := math.Sqrt(dir[0]*dir[0] + dir[1]*dir[1] + dir[2]*dir[2])
	if n < 1e-9 {
		return [3]float32{}
	}
	a := float64(g.Force) * math.Exp(-float64(g.Decay)*dist) / n
	return [3]float32{float32(dir[0] * a), float32(dir[1] * a), float32(dir[2] * a)}
}
