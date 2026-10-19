package world

import (
	"math"

	"daocweb/internal/server/protocol"
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
type Body struct{ X, Y, Z, Yaw float64 }

// Keys are the movement controls held during a step.
type Keys struct {
	Fwd, Strafe, Turn int
	Run               bool
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
	if k.Slow >= 1 {
		b.Z = g.GroundAt(b.X, b.Y, b.Z)
		return Motion{}
	}
	b.Yaw += float64(k.Turn) * Turn * StepDT
	fx, fy := math.Sin(b.Yaw), -math.Cos(b.Yaw)
	rx, ry := math.Cos(b.Yaw), math.Sin(b.Yaw)
	mx := fx*float64(k.Fwd) + rx*float64(k.Strafe)
	my := fy*float64(k.Fwd) + ry*float64(k.Strafe)
	m := Motion{}
	if l := math.Hypot(mx, my); l > 0 {
		mx, my = mx/l, my/l
		m.Back = k.Fwd < 0
		if k.Fwd == 0 {
			m.Side = k.Strafe
		}
		speed := Walk
		if k.Run && !m.Back && m.Side == 0 {
			speed = Run
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
	b.Z = g.GroundAt(b.X, b.Y, b.Z)
	return m
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
	return count, Keys{Fwd: clamp(r[1]), Strafe: clamp(r[2]), Turn: clamp(r[3]), Run: r[4]&protocol.FlagRun != 0}
}
