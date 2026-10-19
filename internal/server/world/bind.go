package world

import (
	"fmt"
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// DefaultBindRadius is how near a bind point a player must stand to bind
// there when the data gives it no radius.
const DefaultBindRadius = 750

// nearestBind is the bind point of the player's realm nearest to a place in a
// zone, and its distance; ok is false when the zone has none.
func (w *World) nearestBind(p *Player, zone int, x, y float64) (b Bind, dist float64, ok bool) {
	dist = math.Inf(1)
	for _, c := range w.Cat.Binds {
		if c.Zone != zone || (c.Realm != 0 && c.Realm != p.Arch.Realm) {
			continue
		}
		if d := math.Hypot(c.X-x, c.Y-y); d < dist {
			b, dist, ok = c, d, true
		}
	}
	return
}

// settleBind gives a character saved before binding a bind point: the one
// nearest to where they stand, or their start when their zone has none.
func (z *Zone) settleBind(p *Player) {
	c := p.Char
	if c.Bind != nil {
		return
	}
	if b, _, ok := z.W.nearestBind(p, z.ID, p.Body.X, p.Body.Y); ok {
		c.Bind = &store.Place{Zone: b.Zone, X: b.X, Y: b.Y, Z: b.Z}
	} else if st, ok := z.W.Cat.StartFor(p.Arch.Realm, p.Arch.Race, p.Arch.Class); ok {
		c.Bind = &store.Place{Zone: st.Zone, X: st.X, Y: st.Y, Z: st.Z, Heading: st.Heading}
	} else {
		return
	}
	if err := z.W.Store.Save(c); err != nil {
		z.W.Log("binding %s: %v", c.Name, err)
	}
}

// BindHere binds a player at the bind point they stand at (/bind).
func (z *Zone) BindHere(p *Player) {
	if p.dead {
		p.errMsg("bind", "You cannot bind while dead.")
		return
	}
	b, d, ok := z.W.nearestBind(p, z.ID, p.Body.X, p.Body.Y)
	r := float64(b.Radius)
	if r <= 0 {
		r = DefaultBindRadius
	}
	if !ok || d > r {
		p.errMsg("bind", "There is nothing to bind to here.")
		return
	}
	p.Char.Bind = &store.Place{Zone: z.ID, X: b.X, Y: b.Y, Z: b.Z, Heading: int(p.Body.Yaw * 1000)}
	z.W.savePlayer(p)
	p.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "sys", Text: "You are bound here. When you die and release, you will return to this place."})
}

// MovedStartLevel is the highest level moved off an unsafe start.
const MovedStartLevel = 5

// leaveUnsafeStart moves a low-level character who stands at, or is bound
// to, a start the start rule replaced: to the replacement, bound there, with
// a line saying why. Once moved they are neither, so it happens once.
func (w *World) leaveUnsafeStart(p *Player) {
	c := p.Char
	if c.Level > MovedStartLevel {
		return
	}
	u, near := w.Cat.UnsafeStartNear(c.Zone, c.X, c.Y)
	if !near && c.Bind != nil {
		u, near = w.Cat.UnsafeStartNear(c.Bind.Zone, c.Bind.X, c.Bind.Y)
	}
	if !near {
		return
	}
	s := u.Instead
	c.Zone, c.X, c.Y, c.Z, c.Heading = s.Zone, s.X, s.Y, s.Z, s.Heading
	c.Bind = &store.Place{Zone: s.Zone, X: s.X, Y: s.Y, Z: s.Z, Heading: s.Heading}
	if err := w.Store.Save(c); err != nil {
		w.Log("moving %s off an unsafe start: %v", c.Name, err)
	}
	p.notice = fmt.Sprintf("Your starting place was beside %s (level %d), too dangerous for a new character: you have been moved, and bound, to a safer start.",
		u.Worst.Name, u.Worst.Level)
}
