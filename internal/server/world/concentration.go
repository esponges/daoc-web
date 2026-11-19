package world

import (
	"fmt"
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// concBuff is a lasting buff a caster holds on a player for as long as they
// keep concentration on it: it has no timer, and costs the caster part of
// their pool.
type concBuff struct {
	id       int
	spell    *SpellDef
	caster   *Player
	target   *Player
	eff      *effect
	cost     int
	outSince float64 // when the target left concentration range; 0 while in range
}

// holdsConc says whether a spell is a concentration buff: a lasting buff with
// a concentration cost, as opposed to a timed one. Songs, which pulse, and
// hostile spells are not.
func holdsConc(sp *SpellDef) bool {
	return sp.Concentration > 0 && !sp.pulsing() && !sp.Hostile()
}

// concUsed is the part of a caster's pool their buffs hold.
func (p *Player) concUsed() int {
	n := 0
	for _, cb := range p.concs {
		n += cb.cost
	}
	return n
}

// concHeldOn finds the buff a caster already holds of a spell on a target.
func (p *Player) concHeldOn(sp *SpellDef, target *Player) *concBuff {
	for _, cb := range p.concs {
		if cb.spell.ID == sp.ID && cb.target == target {
			return cb
		}
	}
	return nil
}

// concRefusal is why a caster may not take on a concentration buff now, or "".
// A recast of the same buff on the same target gives back its old cost first.
func (z *Zone) concRefusal(p *Player, sp *SpellDef, target combatant) string {
	if !holdsConc(sp) {
		return ""
	}
	free := rules.ConcentrationFree(p.sh.maxConc, p.concUsed())
	if t, ok := target.(*Player); ok {
		if old := p.concHeldOn(sp, t); old != nil {
			free += old.cost
		}
	}
	if !rules.CanHold(free, sp.Concentration) {
		return rules.ConcentrationShort(sp.Concentration)
	}
	return ""
}

// noteConc records a buff that has just landed as held by its caster, ending
// the one it replaces. e is the effect it left on the target.
func (z *Zone) noteConc(caster, target *Player, e *effect) {
	if old := caster.concHeldOn(e.spell, target); old != nil && old.eff == e {
		return // the same effect refreshed: still held
	}
	caster.concNext++
	cb := &concBuff{id: caster.concNext, spell: e.spell, caster: caster, target: target, eff: e, cost: e.spell.Concentration}
	e.cb = cb
	e.ends = math.Inf(1) // no timer: it ends when the caster lets it go
	caster.concs = append(caster.concs, cb)
	z.concChanged(caster)
}

// dropConc forgets a buff on the caster's side, giving its cost back.
func (z *Zone) dropConc(cb *concBuff) {
	c := cb.caster
	for i, o := range c.concs {
		if o == cb {
			c.concs = append(c.concs[:i], c.concs[i+1:]...)
			break
		}
	}
	z.concChanged(c)
}

// endConc ends a held buff: it leaves its target, and the cost returns to the
// caster.
func (z *Zone) endConc(cb *concBuff, why string) {
	z.dropConc(cb)
	list := &cb.target.effects
	for i, e := range *list {
		if e == cb.eff {
			*list = append((*list)[:i], (*list)[i+1:]...)
			break
		}
	}
	z.effectsChanged(cb.target)
	if why != "" {
		sys(cb.caster, why)
	}
}

// endAllConc ends every buff a caster holds.
func (z *Zone) endAllConc(p *Player) {
	for len(p.concs) > 0 {
		z.endConc(p.concs[0], "")
	}
}

// CancelConc is a caster letting go of one of their buffs.
func (z *Zone) CancelConc(p *Player, id int) {
	for _, cb := range p.concs {
		if cb.id == id {
			z.endConc(cb, fmt.Sprintf("You cancel your %s.", cb.spell.Name))
			return
		}
	}
}

// sweepConc is the once-a-second check on every held buff: a caster who has
// died or left, a target who has left, and a target who has been out of range
// longer than the grace time.
func (z *Zone) sweepConc() {
	for _, p := range z.players {
		if len(p.concs) == 0 {
			continue
		}
		for _, cb := range append([]*concBuff(nil), p.concs...) {
			switch {
			case p.dead || p.closed.Load():
				z.endConc(cb, "")
			case cb.target.closed.Load() || cb.target.Zone != z:
				z.endConc(cb, "")
			case cb.target != p && rules.OutOfRange(math.Hypot(cb.target.Body.X-p.Body.X, cb.target.Body.Y-p.Body.Y)):
				if cb.outSince == 0 {
					cb.outSince = z.now
				}
				if rules.Lapsed(cb.outSince, z.now) {
					z.endConc(cb, fmt.Sprintf("Your %s fades from %s: they are too far away.", cb.spell.Name, cb.target.Char.Name))
				}
			default:
				cb.outSince = 0
			}
		}
	}
}

// concMsg is the caster's concentration as the protocol carries it.
func (p *Player) concMsg() protocol.Conc {
	m := protocol.Conc{T: protocol.TConc, Max: p.sh.maxConc, Used: p.concUsed(), List: []protocol.ConcView{}}
	for _, cb := range p.concs {
		m.List = append(m.List, protocol.ConcView{ID: cb.id, Spell: cb.spell.ID, Name: cb.spell.Name, Target: cb.target.Char.Name, Cost: cb.cost})
	}
	return m
}

// concChanged tells a caster their concentration changed, in the list and in
// the character sheet.
func (z *Zone) concChanged(p *Player) {
	if p.closed.Load() || !p.welcomed {
		return
	}
	if m := p.concMsg(); fmt.Sprint(m) != fmt.Sprint(p.concSent) {
		p.concSent = m
		p.send(m)
	}
	if m := p.statsMsg(z.W.Cat); !sameStats(m, p.statsSent) {
		p.statsSent = m
		p.send(m)
	}
}
