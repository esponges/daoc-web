package world

import (
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// Duel limits.
const (
	DuelRange   = 1500.0 // how close a challenge must be made
	DuelAbandon = 2500.0 // a duel ends when the two are further apart than this
	DuelLapse   = 30.0   // seconds a challenge waits for an answer
)

// duel is two players fighting each other by agreement. Nothing else may harm
// either of them while it lasts, it gives no experience, and the loser is
// left at 1 hit point instead of dying.
type duel struct{ a, b *Player }

// duelOffer is a challenge waiting for its answer.
type duelOffer struct {
	from    *Player
	expires float64
}

func (d *duel) other(p *Player) *Player {
	if d.a == p {
		return d.b
	}
	return d.a
}

// dueling says whether two players are in a duel with each other.
func dueling(a, b *Player) bool {
	return a != nil && b != nil && a != b && a.duel != nil && a.duel == b.duel
}

// mayHarm says whether one combatant may harm another. Creatures and players
// harm each other as always; a player may harm another player only in a duel
// with them, and a player in a duel is safe from everything else.
func (z *Zone) mayHarm(by, victim combatant) bool {
	vp, ok := victim.(*Player)
	if !ok {
		return true
	}
	if owner := ownerOf(by); owner != nil {
		if owner == vp {
			return true // one's own area effects and so on
		}
		_, byIsPlayer := by.(*Player)
		if byIsPlayer {
			return dueling(owner, vp)
		}
		return vp.duel == nil || dueling(owner, vp)
	}
	return true // a creature
}

// Duel challenges the targeted player.
func (z *Zone) Duel(p *Player) {
	o := z.players[p.target]
	switch {
	case p.dead:
		sys(p, "You cannot duel while dead.")
	case o == nil || o == p || o.dead || o.closed.Load():
		sys(p, "Target a player to challenge.")
	case o.Arch.Realm != p.Arch.Realm:
		sys(p, "You cannot duel that player.")
	case p.duel != nil || o.duel != nil:
		sys(p, "One of you is already dueling.")
	case z.now < p.combatUntil || z.now < o.combatUntil:
		sys(p, "You cannot duel while in combat.")
	case math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) > DuelRange:
		sys(p, "You are too far away to duel.")
	case o.duelOffer != nil && z.now < o.duelOffer.expires:
		sys(p, o.Char.Name+" has already been challenged.")
	default:
		o.duelOffer = &duelOffer{from: p, expires: z.now + DuelLapse}
		sys(p, "You challenge "+o.Char.Name+" to a duel.")
		sys(o, p.Char.Name+" challenges you to a duel. Type /accept to fight or /decline.")
		o.send(protocol.DuelAsk{T: protocol.TDuelAsk, From: p.Char.Name, Secs: int(DuelLapse)})
	}
}

// answerDuel handles /accept and /decline for a waiting challenge; it returns
// whether there was one.
func (z *Zone) answerDuel(p *Player, accept bool) bool {
	off := p.duelOffer
	if off == nil {
		return false
	}
	p.duelOffer = nil
	c := off.from
	if z.now >= off.expires || c.closed.Load() || c.dead || z.players[c.ID] != c {
		sys(p, "The challenge has lapsed.")
		return true
	}
	if !accept {
		sys(c, p.Char.Name+" declines your challenge.")
		sys(p, "You decline the challenge.")
		return true
	}
	if p.duel != nil || c.duel != nil || p.dead {
		sys(p, "You cannot duel now.")
		return true
	}
	d := &duel{a: c, b: p}
	c.duel, p.duel = d, d
	for _, q := range []*Player{c, p} {
		q.attacking, q.pending = false, -1
		sys(q, "The duel with "+d.other(q).Char.Name+" has begun!")
	}
	return true
}

// endDuel ends a duel. A winner is named when one side was beaten; the loser
// is left as they are (the caller has set them to 1 hit point).
func (z *Zone) endDuel(d *duel, winner, loser *Player, why string) {
	for _, q := range []*Player{d.a, d.b} {
		if q.duel != d {
			continue
		}
		q.duel = nil
		q.attacking, q.pending = false, -1
		if q.casting != nil {
			z.cancelPlayerCast(q, "")
		}
	}
	if winner != nil && loser != nil {
		msg := winner.Char.Name + " has defeated " + loser.Char.Name + " in a duel!"
		for _, q := range z.players {
			sys(q, msg)
		}
		return
	}
	for _, q := range []*Player{d.a, d.b} {
		sys(q, why)
	}
}

// duelHit is a blow or spell on a duellist that would kill them: it leaves
// them at 1 hit point and ends the duel in the other's favour. It returns
// whether it did.
func (z *Zone) duelDefeat(t *Player) bool {
	d := t.duel
	if d == nil {
		return false
	}
	t.hp = 1
	z.endDuel(d, d.other(t), t, "")
	return true
}

// checkDuel ends a duel whose players have gone apart or away. Run each tick
// for the first of the pair.
func (z *Zone) checkDuel(p *Player) {
	d := p.duel
	if d == nil || d.a != p {
		return
	}
	switch {
	case d.a.closed.Load() || d.b.closed.Load() || z.players[d.a.ID] != d.a || z.players[d.b.ID] != d.b:
		z.endDuel(d, nil, nil, "The duel is over.")
	case d.a.dead || d.b.dead:
		z.endDuel(d, nil, nil, "The duel is over.")
	case math.Hypot(d.a.Body.X-d.b.Body.X, d.a.Body.Y-d.b.Body.Y) > DuelAbandon:
		z.endDuel(d, nil, nil, "The duel ends: you are too far apart.")
	}
}

// duelInterference ends the duel of a player a creature (or anyone not their
// opponent) has attacked.
func (z *Zone) duelInterference(p *Player) {
	if d := p.duel; d != nil {
		z.endDuel(d, nil, nil, "The duel ends: someone else has joined the fight.")
	}
}

// duelTarget is the opponent when the player has targeted them in a duel.
func (z *Zone) duelTarget(p *Player) *Player {
	if o := z.players[p.target]; dueling(p, o) {
		return o
	}
	return nil
}

// combatDuel runs a duellist's swings at their opponent.
func (z *Zone) combatDuel(p, o *Player) {
	if o.dead || o.closed.Load() {
		p.attacking, p.pending = false, -1
		return
	}
	if math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) > 130 {
		return
	}
	if angleOff(p.Body.Yaw, angleTo(p.Body.X, p.Body.Y, o.Body.X, o.Body.Y)) > frontArc*1.2 {
		return
	}
	if z.now < p.swingAt {
		return
	}
	p.swingAt = z.now + p.swingTime()
	z.duelSwing(p, o)
}

// duelSwing is one blow of one duellist at the other, resolved with the
// defender's own evade, parry, block and armour.
func (z *Zone) duelSwing(p, o *Player) {
	style := z.queuedStyle(p)
	z.unstealth(p, "You are no longer hidden.")
	z.standUp(o)
	p.combatUntil, o.combatUntil = z.now+CombatLinger, z.now+CombatLinger
	ev := protocol.Event{K: "swing", From: p.ID, To: o.ID}
	if style != nil {
		ev.Style, ev.Text = style.ClientStyle(p.weapon != nil && p.weapon.TwoHanded()), style.Name
	}
	z.event(ev, p.Body.X, p.Body.Y)
	cb := o.combatant(z.W.Cat)
	sw := rules.Swing{Attacker: p.combatant(z.W.Cat), Defender: cb, Attackers: 1,
		InFront: angleOff(o.Body.Yaw, angleTo(o.Body.X, o.Body.Y, p.Body.X, p.Body.Y)) <= frontArc}
	if style != nil {
		sw.Style = &rules.Style{GrowthRate: style.GrowthRate, Spec: p.combatSpec(style.Spec), BonusToHit: style.BonusToHit, Location: rules.LocNone}
	}
	out := rules.Resolve(sw, z.rng)
	if out.Result != rules.Hit {
		if style != nil {
			p.end = math.Max(0, p.end-float64(rules.FailedStyleCost(p.weaponSpeed(), style.Endurance)))
		}
		if out.Result == rules.Missed {
			z.event(protocol.Event{K: "miss", From: p.ID, To: o.ID}, o.Body.X, o.Body.Y)
		} else {
			kind := out.Result.String()
			z.event(protocol.Event{K: kind, From: o.ID, To: p.ID}, o.Body.X, o.Body.Y)
			o.lastDef = defence{kind: kind, at: z.now}
		}
		p.lastHit, p.lastHitAt, p.lastStyle = false, z.now, 0
		return
	}
	if style != nil {
		p.end = math.Max(0, p.end-float64(p.styleCost(style)))
		p.lastStyle, p.lastStyleAt = style.ID, z.now
		p.styleDefence = style.BonusToDef
	} else {
		p.lastStyle = 0
	}
	z.wearItem(p, SlotWeapon)
	z.wearBlow(o, out)
	amount := sickened(out.Damage, p)
	p.lastHit, p.lastHitAt = true, z.now
	z.event(protocol.Event{K: "hit", From: p.ID, To: o.ID, Amount: amount, Crit: out.Critical > 0, Text: ev.Text}, o.Body.X, o.Body.Y)
	z.hurt(o, p, amount)
}

// mayHarmPlayer says whether the player may aim a hostile spell at another
// player: only their duel opponent.
func (z *Zone) mayHarmPlayer(p, o *Player) bool { return o != nil && dueling(p, o) }
