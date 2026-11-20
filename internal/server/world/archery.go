package world

import (
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// An archer's shot, from the rules-archery capability: draw at the held target,
// release when the draw ends, use one arrow. The server decides the draw time,
// the checks at release, the arrow and the hit; the client only asks.

const arrowFlight = 1800.0 // units a second, from OpenDAoC-Core

// playerDraw is a shot being drawn.
type playerDraw struct {
	target   *NPC
	kind     string  // "normal", "critical" or "rapid"
	base     float64 // the normal draw, seconds
	ends     float64
	x, y     float64 // where the archer stood, for moving away to interrupt
	rangeCls int
	accCls   int
}

// arrowBits is an arrow's speed/absolute field, which holds its quality classes.
func arrowBits(x *Item) int { return int(math.Round(x.Speed * 10)) }

// quiverArrow is the slot and item of the arrows the archer will fire: the
// first quiver slot holding any.
func (p *Player) quiverArrow(cat *Catalog) (int, *Item) {
	for s := SlotQuiver; s < SlotQuiver+4; s++ {
		if it := p.inv[s]; it != nil && it.Count > 0 {
			if x := cat.Items[it.ID]; x != nil && x.ObjectType == rules.ObjArrow {
				return s, x
			}
		}
	}
	return -1, nil
}

// Shoot is the bow order: see protocol.Shoot.
func (z *Zone) Shoot(p *Player, mode string) {
	if p.dead || p.ride != nil {
		return
	}
	switch mode {
	case "stop":
		p.shooting = false
		z.cancelDraw(p, "You lower your bow.")
		return
	case "rapid":
		if p.sh.abilities[rules.AbRapidFire] == 0 {
			p.errMsg("shoot", "You have not learned Rapid Fire.")
			return
		}
		p.rapid = !p.rapid
		p.critical = false
		if p.rapid {
			sys(p, "You switch to rapid fire mode.")
		} else {
			sys(p, "You stop firing rapidly.")
		}
		return
	case "critical":
		if p.sh.abilities[rules.AbCriticalShot] == 0 {
			p.errMsg("shoot", "You have not learned Critical Shot.")
			return
		}
		p.critical, p.rapid = true, false
	}
	if held(p) {
		p.errMsg("shoot", "You cannot shoot while you are held.")
		return
	}
	if p.casting != nil {
		p.errMsg("shoot", "You are already casting a spell.")
		return
	}
	if p.draw != nil {
		return
	}
	p.attacking, p.pending = false, -1
	if z.beginDraw(p) {
		p.shooting = mode != "critical"
	}
}

// beginDraw checks a shot and starts its draw; false when it is refused.
func (z *Zone) beginDraw(p *Player) bool {
	if p.bow == nil {
		p.errMsg("shoot", "You have no bow in your ranged slot.")
		return false
	}
	n := z.byID[p.target]
	if n == nil || n.dead() || n.gone || n.Friendly || !n.fightable(z) {
		p.errMsg("shoot", "You must select a target to shoot at.")
		return false
	}
	_, arrow := p.quiverArrow(z.W.Cat)
	if arrow == nil {
		p.errMsg("shoot", "You have no ammunition.")
		return false
	}
	bits := arrowBits(arrow)
	rc, ac := rules.ArrowRangeClass(bits), rules.ArrowAccuracyClass(bits)
	if math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y) > rules.BowRange(p.bow.ObjectType, 100, rc) {
		p.errMsg("shoot", "That target is too far away!")
		return false
	}
	if !z.sees(p, n) {
		p.errMsg("shoot", rules.SightMessage)
		return false
	}
	if angleOff(p.Body.Yaw, angleTo(p.Body.X, p.Body.Y, n.Body.X, n.Body.Y)) > frontArc*1.5 {
		p.errMsg("shoot", "You must be facing your target.")
		return false
	}
	base := rules.DrawTime(p.bow.Speed, p.sh.stats[rules.QUI])
	kind, dur := "normal", base
	switch {
	case p.critical:
		kind, dur = "critical", rules.CriticalDraw(base, p.sh.abilities[rules.AbCriticalShot])
	case p.rapid:
		kind, dur = "rapid", rules.RapidDraw(base)
	}
	z.unstealth(p, "You are no longer hidden.")
	p.draw = &playerDraw{target: n, kind: kind, base: base, ends: z.now + dur, x: p.Body.X, y: p.Body.Y, rangeCls: rc, accCls: ac}
	p.combatUntil = z.now + CombatLinger
	z.event(protocol.Event{K: "draw", From: p.ID, To: n.ID, Text: kind, Dur: dur}, p.Body.X, p.Body.Y)
	return true
}

// cancelDraw drops a shot being drawn, with no arrow used.
func (z *Zone) cancelDraw(p *Player, text string) {
	if p.draw == nil {
		return
	}
	p.draw = nil
	p.critical = false
	if text != "" {
		sys(p, text)
	}
	z.event(protocol.Event{K: "drawstop", From: p.ID}, p.Body.X, p.Body.Y)
}

// continueDraw runs a draw in progress: moving or being held stops it, and a
// finished draw releases the arrow, then draws the next shot if the archer is
// still shooting.
func (z *Zone) continueDraw(p *Player) {
	d := p.draw
	if d == nil {
		// A blow that broke the draw leaves the archer shooting: draw again.
		if p.shooting && !p.dead && !held(p) && p.casting == nil && !z.beginDraw(p) {
			p.shooting = false
		}
		return
	}
	if held(p) {
		p.shooting = false
		z.cancelDraw(p, "")
		return
	}
	if math.Hypot(p.Body.X-d.x, p.Body.Y-d.y) > castMoveLimit {
		p.shooting = false
		z.cancelDraw(p, "You move and your shot is interrupted.")
		return
	}
	if z.now+1e-9 < d.ends {
		return
	}
	p.draw = nil
	z.event(protocol.Event{K: "drawstop", From: p.ID}, p.Body.X, p.Body.Y)
	z.releaseShot(p, d)
	p.critical = false
	if p.shooting && !p.dead && !d.target.dead() {
		if !z.beginDraw(p) {
			p.shooting = false
		}
	} else {
		p.shooting = false
	}
}

// releaseShot lets an arrow go: the checks are made again at release, one arrow
// and the endurance are used, and the arrow is rolled like a swing with no parry.
func (z *Zone) releaseShot(p *Player, d *playerDraw) {
	n := d.target
	if p.bow == nil || n.dead() || n.gone || !n.fightable(z) {
		return
	}
	slot, arrow := p.quiverArrow(z.W.Cat)
	if arrow == nil {
		p.errMsg("shoot", "You have no ammunition.")
		return
	}
	dist := math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y)
	if dist > rules.BowRange(p.bow.ObjectType, 100, d.rangeCls) {
		p.errMsg("shoot", "That target is too far away!")
		return
	}
	if !z.sees(p, n) {
		p.errMsg("shoot", rules.SightMessage)
		return
	}
	cost := float64(rules.ShotEndurance(d.kind == "critical", p.sh.abilities[rules.AbRapidFire]))
	if p.end < cost {
		p.errMsg("shoot", "You are too tired to shoot.")
		p.shooting = false
		return
	}
	p.end -= cost
	z.useArrow(p, slot)
	eff := 1.0
	switch d.kind {
	case "critical":
		eff = rules.CriticalShotDamage(p.Char.Level, n.Level)
	case "rapid":
		eff = rules.RapidFireDamage(rules.RapidDraw(d.base), d.base)
	}
	z.event(protocol.Event{K: "arrow", From: p.ID, To: n.ID, Dur: dist / arrowFlight}, p.Body.X, p.Body.Y)
	z.interruptCasting(n)
	cb := p.rangedCombatant(z.W.Cat)
	sw := rules.Swing{Attacker: cb, Defender: &n.cb, Attackers: 1, Ranged: true, Effectiveness: eff,
		MissScale: rules.ArrowMissScale(d.accCls),
		InFront:   angleOff(n.Body.Yaw, angleTo(n.Body.X, n.Body.Y, p.Body.X, p.Body.Y)) <= frontArc}
	o := rules.Resolve(sw, z.rng)
	if o.Result != rules.Hit {
		p.lastHit = false
		if o.Result == rules.Missed {
			z.event(protocol.Event{K: "miss", From: p.ID, To: n.ID}, n.Body.X, n.Body.Y)
		} else {
			z.event(protocol.Event{K: o.Result.String(), From: n.ID, To: p.ID}, n.Body.X, n.Body.Y)
		}
		z.provoke(n, p, 0)
		return
	}
	amount := sickened(o.Damage, p)
	p.lastHit, p.lastHitAt = true, z.now
	z.event(protocol.Event{K: "hit", From: p.ID, To: n.ID, Amount: amount, Crit: o.Critical > 0, Text: d.kindText()}, n.Body.X, n.Body.Y)
	n.HP -= amount
	n.hurtBy[p.ID] += amount
	z.damaged(n)
	if n.tagger == 0 {
		n.tagger = p.ID
	}
	z.provoke(n, p, float64(amount))
	if n.HP <= 0 {
		z.killNPC(n)
		return
	}
	if add := z.damageAdd(p, n, amount); add > 0 {
		z.event(protocol.Event{K: "spell", From: p.ID, To: n.ID, Text: "damage add", Amount: add}, n.Body.X, n.Body.Y)
		z.hurt(n, p, add)
	}
}

func (d *playerDraw) kindText() string {
	switch d.kind {
	case "critical":
		return "Critical Shot"
	case "rapid":
		return "Rapid Fire"
	}
	return ""
}

// rangedCombatant is the archer as a shot sees them: the bow in hand, with the
// composite bow spec.
func (p *Player) rangedCombatant(cat *Catalog) *rules.Combatant {
	cb := p.combatant(cat)
	cb.Unarmed = false
	cb.Weapon = weaponOf(p.bow)
	cb.Weapon.TwoHanded = true
	cb.Weapon.Condition = p.condOf(SlotRanged)
	cb.Spec = p.combatSpec(weaponSpec(p.bow.ObjectType))
	return cb
}

// useArrow takes one arrow from the stack in a quiver slot.
func (z *Zone) useArrow(p *Player, slot int) {
	it := p.inv[slot]
	if it == nil {
		return
	}
	it.Count--
	if it.Count <= 0 {
		z.W.Store.Tx(func(tx *store.Tx) error { return tx.RemoveItem(p.Char.ID, slot) })
		delete(p.inv, slot)
	} else {
		z.W.Store.Tx(func(tx *store.Tx) error { return tx.SetCount(p.Char.ID, slot, it.Count) })
	}
	p.sendInv(z.W.Cat, p.Char.Money)
}
