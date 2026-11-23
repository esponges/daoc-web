package world

import (
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// The game's own rates, as the viewer uses them.
const (
	Walk = 85.0  // units a second
	Run  = 191.0 // units a second
	Turn = 2.8   // radians a second
	// StepDT is one fixed movement step.
	StepDT = 1.0 / protocol.StepsPerSecond
)

// Body is where a mover stands. Yaw faces (sin yaw, -cos yaw), as in the viewer.
// VZ, PeakZ, Air and Jumped are the vertical state: a body in the air falls
// under gravity until it meets the ground or water.
type Body struct {
	X, Y, Z, Yaw float64
	VZ           float64 // vertical speed in the air, units a second
	PeakZ        float64 // the highest the body has been in this stretch of air
	Air          bool
	Jumped       bool // the air began with a jump, not a drop
}

// Keys are the movement controls held during a step.
type Keys struct {
	Fwd, Strafe, Turn int
	Run               bool
	// Jump asks to leave the ground; Sprint runs faster. The server sets
	// Sprint from the player's state; the client predicts with the same.
	Jump, Sprint bool
	// Slow is the share of speed a slow or a hold takes away: 0 free, 1 held,
	// when the body neither moves nor turns. The server sets it from the
	// body's effects; the client predicts with the factor it was sent.
	Slow float64
}

// Motion is what a step did, for animation.
type Motion struct {
	Speed float64
	Back  bool
	Side  int
	// Swim: the body is afloat. Air: it is in the air at the end of the
	// step. Landed: it came down in this step, from FallHeight units up;
	// Water says it landed in water.
	Swim, Air, Landed, Water bool
	FallHeight               float64
}

// Mover is the extra check applied to a step's destination: collision. The
// zero value moves freely.
type Mover interface {
	// Slide returns where a body moving from (x0,y0) to (x1,y1) ends up.
	Slide(x0, y0, x1, y1, z float64) (float64, float64)
}

// Step advances a body by one fixed step. It is the single definition of how
// a character moves: the server runs it on inputs, and the browser runs the
// same maths (web/net.js) to predict. Backing up and sidestepping are always
// at a walk, like the viewer's.
func Step(g Ground, mv Mover, b *Body, k Keys) Motion {
	m := Motion{}
	held := k.Slow >= 1
	if !held {
		afloat := !b.Air && swimming(g, b.X, b.Y)
		b.Yaw += float64(k.Turn) * Turn * StepDT
		fx, fy := math.Sin(b.Yaw), -math.Cos(b.Yaw)
		rx, ry := math.Cos(b.Yaw), math.Sin(b.Yaw)
		mx := fx*float64(k.Fwd) + rx*float64(k.Strafe)
		my := fy*float64(k.Fwd) + ry*float64(k.Strafe)
		if l := math.Hypot(mx, my); l > 0 {
			mx, my = mx/l, my/l
			m.Back = k.Fwd < 0
			if k.Fwd == 0 {
				m.Side = k.Strafe
			}
			speed := Walk
			switch {
			case afloat:
				speed = rules.SwimSpeed
			case k.Run && !m.Back && m.Side == 0:
				speed = Run
				if k.Sprint {
					speed = rules.SprintSpeed(Run)
				}
			}
			d := speed * (1 - k.Slow) * StepDT
			nx, ny := g.Clamp(b.X+mx*d, b.Y+my*d)
			if mv != nil {
				nx, ny = mv.Slide(b.X, b.Y, nx, ny, b.Z)
			}
			moved := math.Hypot(nx-b.X, ny-b.Y)
			b.X, b.Y = nx, ny
			m.Speed = moved / StepDT
		}
	}
	vertical(g, b, k.Jump && !held, &m)
	m.Air = b.Air
	return m
}

// swimming says whether a body standing at (x, y) is in water deep enough to
// swim in.
func swimming(g Ground, x, y float64) bool {
	surface, ok := g.WaterAt(x, y)
	return ok && surface-g.GroundAt(x, y, surface) > rules.SwimDepth
}

// vertical is the part of a step that moves a body up and down: it follows
// the ground, floats on water, jumps, and falls when the ground drops away
// faster than it can follow.
func vertical(g Ground, b *Body, jump bool, m *Motion) {
	ground := g.GroundAt(b.X, b.Y, b.Z)
	surface, wet := g.WaterAt(b.X, b.Y)
	deep := wet && surface-ground > rules.SwimDepth
	if !b.Air {
		switch {
		case deep:
			b.Z = surface - rules.SwimDepth
			m.Swim = true
			return
		case jump:
			b.Air, b.Jumped, b.VZ, b.PeakZ = true, true, rules.JumpSpeed, b.Z
		case b.Z-ground > rules.StepDown:
			b.Air, b.Jumped, b.VZ, b.PeakZ = true, false, 0, b.Z
		default:
			b.Z = ground
			return
		}
	}
	b.VZ -= rules.Gravity * StepDT
	b.Z += b.VZ * StepDT
	b.PeakZ = math.Max(b.PeakZ, b.Z)
	switch {
	case deep && b.Z <= surface-rules.SwimDepth:
		b.Z = surface - rules.SwimDepth
		m.Water = true
	case !deep && b.Z <= ground:
		b.Z = ground
	default:
		return
	}
	m.Landed, m.FallHeight = true, b.PeakZ-b.Z
	b.Air, b.Jumped, b.VZ = false, false, 0
}

// KeysFromRun decodes one run of an Input: [count, fwd, strafe, turn, flags].
// Out-of-range values are clamped, so a hostile client gets no more than the
// keys can give.
func KeysFromRun(r [5]int) (count int, k Keys) {
	clamp := func(v int) int {
		if v < -1 {
			return -1
		}
		if v > 1 {
			return 1
		}
		return v
	}
	count = r[0]
	if count < 0 {
		count = 0
	}
	if count > protocol.StepsPerSecond*2 {
		count = protocol.StepsPerSecond * 2
	}
	return count, Keys{Fwd: clamp(r[1]), Strafe: clamp(r[2]), Turn: clamp(r[3]), Run: r[4]&protocol.FlagRun != 0, Jump: r[4]&protocol.FlagJump != 0}
}
