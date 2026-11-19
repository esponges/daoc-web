package world

import (
	"fmt"
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// rezOffer is a resurrection spell waiting for a dead player's answer.
type rezOffer struct {
	from     int64 // the caster's entity id
	fromName string
	spell    *SpellDef
	cost     float64 // the power the caster paid, given back if it is not taken
	at       float64 // when it was offered, on the zone clock
}

// effSick is resurrection sickness on a player's effect list.
const effSick = "sick"

// sickSpell stands for the spell that leaves resurrection sickness: it names
// the effect and carries its length and strength.
func sickSpell(seconds float64, percent int) *SpellDef {
	return &SpellDef{ID: -1, Name: "Resurrection Sickness", Type: "ResurrectionIllness", Target: "Self", Duration: int(math.Round(seconds)), Value: float64(percent)}
}

// sickFactor is what resurrection sickness leaves of a combatant's damage and
// spell strength: 1 when they are not sick.
func sickFactor(c combatant) float64 {
	if p, ok := c.(*Player); ok {
		for _, e := range p.effects {
			if e.kind == effSick {
				return rules.SicknessFactor(int(e.spell.Value))
			}
		}
	}
	return 1
}

// sickened is an amount of damage or healing after the doer's resurrection
// sickness, if they have it; what was something stays at least 1.
func sickened(amount int, doer combatant) int {
	f := sickFactor(doer)
	if f >= 1 || amount <= 0 {
		return amount
	}
	return max(1, int(math.Round(float64(amount)*f)))
}

// corpseOK says whether a player is a body a resurrection can raise.
func (z *Zone) corpseOK(c combatant) bool {
	o, ok := c.(*Player)
	return ok && o.dead && !o.closed.Load() && o.Zone == z
}

// offerResurrection asks a dead player whether they will be raised by a
// caster's spell that has just landed. Only one offer stands at a time.
func (z *Zone) offerResurrection(p *Player, t *Player, sp *SpellDef, cost float64) {
	if t.rez != nil {
		z.refundPower(p, cost)
		sys(p, t.Char.Name+" is already being offered a resurrection.")
		return
	}
	t.rez = &rezOffer{from: p.ID, fromName: p.Char.Name, spell: sp, cost: cost, at: z.now}
	t.send(protocol.RezOffer{T: protocol.TRezOffer, From: p.Char.Name, Health: sp.ResurrectHealth, Power: sp.ResurrectMana, Secs: rules.RezOfferSeconds})
	sys(p, fmt.Sprintf("You offer %s a resurrection.", t.Char.Name))
	sys(t, fmt.Sprintf("%s offers to resurrect you with %d percent of your hit points. Accept or decline.", p.Char.Name, sp.ResurrectHealth))
}

func (z *Zone) refundPower(p *Player, cost float64) {
	if p != nil && !p.closed.Load() {
		p.power = math.Min(p.power+cost, float64(p.sh.maxPower))
	}
}

// withdrawOffer ends an offer that was not taken: the caster gets their power
// back and the player's prompt closes.
func (z *Zone) withdrawOffer(t *Player, why string) {
	o := t.rez
	if o == nil {
		return
	}
	t.rez = nil
	z.refundPower(z.players[o.from], o.cost)
	t.send(protocol.RezOffer{T: protocol.TRezOffer, Secs: 0})
	if why != "" {
		sys(t, why)
	}
}

// AnswerResurrection is the dead player's answer to the offer that stands.
func (z *Zone) AnswerResurrection(p *Player, accept bool) {
	o := p.rez
	if !p.dead || o == nil {
		sys(p, "No one is currently trying to resurrect you.")
		return
	}
	if !accept {
		z.withdrawOffer(p, "You decline to be resurrected.")
		return
	}
	p.rez = nil
	p.send(protocol.RezOffer{T: protocol.TRezOffer, Secs: 0})
	z.raise(p, o)
}

// raise stands a dead player up where the caster is, with the spell's share of
// their hit points and power, and spares them the release.
func (z *Zone) raise(p *Player, o *rezOffer) {
	sp := o.spell
	x, y, zz := p.Body.X, p.Body.Y, p.Body.Z
	if c := z.players[o.from]; c != nil && !c.closed.Load() {
		x, y, zz = c.Body.X, c.Body.Y, c.Body.Z
	}
	p.raised = true
	z.releasePenalty(p) // a raised character skips the second stage of the death
	p.raised = false
	p.dead = false
	p.hp = math.Max(1, float64(p.maxHP())*float64(sp.ResurrectHealth)/100)
	p.power = float64(p.sh.maxPower) * float64(sp.ResurrectMana) / 100
	p.end = float64(rules.MaxEndurance) * float64(sp.ResurrectMana) / 100
	p.combatUntil = 0
	p.Body = Body{X: x, Y: y, Z: z.G.GroundAt(x, y, zz), Yaw: p.Body.Yaw}
	p.correct = true
	secs, pct := rules.Sickness(sp.ResurrectHealth)
	z.addSickness(p, secs, pct)
	z.event(protocol.Event{K: "rez", From: o.from, To: p.ID, Text: sp.Name}, x, y)
	sys(p, fmt.Sprintf("%s resurrects you.", o.fromName))
	z.W.savePlayer(p)
}

// addSickness gives a raised player resurrection sickness.
func (z *Zone) addSickness(p *Player, seconds float64, percent int) {
	sp := sickSpell(seconds, percent)
	p.effects = append(p.effects, &effect{spell: sp, kind: effSick, ends: z.now + seconds})
	z.effectsChanged(p)
}
