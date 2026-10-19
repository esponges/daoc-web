package world

import (
	"fmt"
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// GraveIDBase keeps gravestone entity ids apart from NPCs and players: a
// grave's id is this plus its owner's character id.
const GraveIDBase = 2_000_000_000

// grave is a gravestone standing in a zone.
type grave struct {
	owner   int64 // character id
	name    string
	x, y, z float64
	xp      int64
}

func (g *grave) id() int64 { return GraveIDBase + g.owner }

// loadGraves puts the zone's saved gravestones back, as a zone starts.
func (z *Zone) loadGraves() {
	z.graves = map[int64]*grave{}
	if z.W.Store == nil {
		return
	}
	rows, err := z.W.Store.GravesIn(z.ID)
	if err != nil {
		z.W.Log("gravestones in zone %d: %v", z.ID, err)
		return
	}
	for _, r := range rows {
		z.graves[r.CharID] = &grave{owner: r.CharID, name: r.Name, x: r.X, y: r.Y, z: r.Z, xp: r.XP}
	}
}

// deathPenalty takes the first part of a death's cost: experience, at once,
// for a death to a creature from level 6. It remembers what was lost so the
// release can take as much again.
func (z *Zone) deathPenalty(p *Player, by *NPC) {
	c := p.Char
	p.deathAt = [3]float64{p.Body.X, p.Body.Y, p.Body.Z}
	p.lastLoss, p.creatureDeath = 0, by != nil
	if by == nil || c.Level < rules.PenaltyLevel {
		return
	}
	loss := rules.DeathLoss(c.Level, c.XP, c.DeathCount)
	c.DeathCount++
	p.lastLoss = loss
	if loss > 0 {
		c.XP -= loss
		z.event(protocol.Event{K: "xploss", To: p.ID, Amount: int(loss), Text: fmt.Sprintf("You lose %d experience.", loss)}, p.Body.X, p.Body.Y)
	}
	z.W.recalc(p)
	z.W.savePlayer(p)
}

// releasePenalty takes the rest when a character releases: the same
// experience again, into a gravestone where they fell, and constitution.
func (z *Zone) releasePenalty(p *Player) {
	c := p.Char
	if !p.creatureDeath || c.Level < rules.PenaltyLevel {
		p.creatureDeath = false
		return
	}
	p.creatureDeath = false
	loss := min(p.lastLoss, max(0, c.XP-rules.XPToReach(c.Level)))
	if loss > 0 {
		c.XP -= loss
		z.event(protocol.Event{K: "xploss", To: p.ID, Amount: int(loss), Text: fmt.Sprintf("You lose %d more experience; your gravestone holds it.", loss)}, p.Body.X, p.Body.Y)
		z.W.placeGrave(z, p, loss)
	}
	con := rules.ConLoss(c.DeathCount)
	c.ConLost += con
	z.event(protocol.Event{K: "conloss", To: p.ID, Amount: con, Text: fmt.Sprintf("You lose %d constitution.", con)}, p.Body.X, p.Body.Y)
	z.W.recalc(p)
}

// placeGrave puts a character's gravestone where they died, replacing any
// other they had, in this zone or another.
func (w *World) placeGrave(z *Zone, p *Player, xp int64) {
	c := p.Char
	if old := c.Grave; old != nil {
		if old.Zone == z.ID {
			delete(z.graves, c.ID)
		} else if oz := w.runningZone(old.Zone); oz != nil {
			id := c.ID
			oz.Do(func() { delete(oz.graves, id) })
		}
	}
	g := &grave{owner: c.ID, name: c.Name, x: p.deathAt[0], y: p.deathAt[1], z: p.deathAt[2], xp: xp}
	z.graves[c.ID] = g
	c.Grave = &store.Grave{Zone: z.ID, X: g.x, Y: g.y, Z: g.z, XP: xp}
}

// runningZone is a loaded zone, or nil.
func (w *World) runningZone(id int) *Zone {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.zones[id]
}

// Pray starts praying at the player's own gravestone, which they must have
// targeted and be near.
func (z *Zone) Pray(p *Player) {
	if p.dead {
		return
	}
	g := z.graves[p.target-GraveIDBase]
	if p.target < GraveIDBase || g == nil {
		for _, o := range z.graves {
			if o.id() == p.target {
				p.errMsg("pray", "That is not your grave.")
				return
			}
		}
		p.errMsg("pray", "You need to target your gravestone to pray.")
		return
	}
	if g.owner != p.Char.ID {
		p.errMsg("pray", "That is not your grave.")
		return
	}
	if math.Hypot(g.x-p.Body.X, g.y-p.Body.Y) > rules.PrayRadius {
		p.errMsg("pray", "You must get closer to your grave to pray.")
		return
	}
	if p.attacking || z.now < p.combatUntil {
		p.errMsg("pray", "You cannot pray while fighting.")
		return
	}
	p.prayUntil = z.now + rules.PraySeconds
	p.prayFrom = [2]float64{p.Body.X, p.Body.Y}
	p.prayCombat = p.combatUntil
	z.event(protocol.Event{K: "pray", To: p.ID, Text: "started"}, p.Body.X, p.Body.Y)
}

// updatePrayer finishes or cancels a prayer. Run each tick.
func (z *Zone) updatePrayer(p *Player) {
	if p.prayUntil == 0 {
		return
	}
	g := z.graves[p.Char.ID]
	moved := math.Hypot(p.Body.X-p.prayFrom[0], p.Body.Y-p.prayFrom[1]) > 1
	if p.dead || g == nil || moved || p.attacking || p.combatUntil != p.prayCombat {
		p.prayUntil = 0
		z.event(protocol.Event{K: "pray", To: p.ID, Text: "cancelled"}, p.Body.X, p.Body.Y)
		return
	}
	if z.now < p.prayUntil {
		return
	}
	p.prayUntil = 0
	c := p.Char
	c.XP += g.xp
	delete(z.graves, c.ID)
	c.Grave = nil
	z.event(protocol.Event{K: "pray", To: p.ID, Amount: int(g.xp), Text: "done"}, p.Body.X, p.Body.Y)
	z.W.recalc(p)
	z.W.savePlayer(p)
}

// graveEntity is a gravestone as a viewer sees it.
func graveEntity(g *grave, to *Player) protocol.Entity {
	return protocol.Entity{ID: g.id(), K: "g", Name: "grave of " + g.name, X: g.x, Y: g.y, Z: g.z, HP: 1, Own: g.owner == to.Char.ID}
}
