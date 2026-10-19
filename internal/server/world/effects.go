package world

import (
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// effect is a lasting spell on a combatant: a damage over time, a slow, a
// stun, a mesmerize or an offensive proc. The zone ticks it.
type effect struct {
	spell       *SpellDef
	kind        string
	caster      combatant // nil once it has gone
	casterLevel int
	ends        float64
	next        float64 // the next damage-over-time tick
	every       float64
}

// Effect kinds, as the client draws them.
const (
	effDoT  = "dot"
	effSlow = "slow"
	effStun = "stun"
	effMez  = "mez"
	effProc = "proc"
)

// effectKind is the kind of lasting effect a spell type leaves, "" for none.
func effectKind(spellType string) string {
	switch spellType {
	case rules.SpellDamageOverTime:
		return effDoT
	case rules.SpellSpeedDecrease:
		return effSlow
	case rules.SpellStun:
		return effStun
	case rules.SpellMesmerize:
		return effMez
	case rules.SpellOffensiveProc:
		return effProc
	}
	return ""
}

// effectsOf is where a combatant keeps its effects.
func effectsOf(c combatant) *[]*effect {
	switch t := c.(type) {
	case *Player:
		return &t.effects
	case *NPC:
		return &t.effects
	}
	return nil
}

func levelOf(c combatant) int {
	switch t := c.(type) {
	case *Player:
		return t.Char.Level
	case *NPC:
		return t.Level
	}
	return 1
}

// held says whether a combatant is stunned or mesmerized.
func held(c combatant) bool {
	for _, e := range *effectsOf(c) {
		if e.kind == effStun || e.kind == effMez {
			return true
		}
	}
	return false
}

// speedFactor is the share of its speed a combatant may move at.
func speedFactor(c combatant) float64 {
	var slows []float64
	for _, e := range *effectsOf(c) {
		if e.kind == effSlow {
			slows = append(slows, rules.SlowFactor(e.spell.Value))
		}
	}
	return rules.SpeedFactor(held(c), slows...)
}

// resistBonus is the item bonus type that raises a resist, by damage type.
var resistBonus = map[int]int{10: 11, 11: 12, rules.DamageCrush: 13, 12: 14, 13: 15, 14: 16, rules.DamageSlash: 17, 15: 18, rules.DamageThrus: 19}

// spellResist is a combatant's resist, in percent, to a damage type: its
// race's, and for a player, what their worn items add.
func (z *Zone) spellResist(c combatant, damageType int) float64 {
	switch t := c.(type) {
	case *NPC:
		return t.resists[damageType]
	case *Player:
		r := z.W.Cat.Resists[t.Char.Race][damageType]
		for s := SlotWeapon; s < SlotCount; s++ {
			if it := t.inv[s]; it != nil {
				if x := z.W.Cat.Items[it.ID]; x != nil {
					for _, b := range x.Bonuses {
						if b.Type == resistBonus[damageType] {
							r += float64(b.Amount)
						}
					}
				}
			}
		}
		return r
	}
	return 0
}

// castSpell lands a spell from a caster on a target: a hostile one by the hit
// roll, then its damage, heal or lasting effect, then any sub-spell. It
// returns whether it landed.
func (z *Zone) castSpell(caster, target combatant, sp *SpellDef) bool {
	if !target.fightable(z) {
		return false
	}
	cx, cy := caster.position()
	tx, ty := target.position()
	if sp.Hostile() {
		if z.rng.Float64() >= rules.SpellHitChance(levelOf(caster), levelOf(target)) {
			z.event(protocol.Event{K: "resist", From: caster.entityID(), To: target.entityID(), Text: sp.Name}, tx, ty)
			return false
		}
	}
	switch sp.Type {
	case rules.SpellHeal:
		z.heal(target, int(sp.Value))
		z.event(protocol.Event{K: "spell", From: caster.entityID(), To: target.entityID(), Text: sp.Name, Amount: -int(sp.Value)}, tx, ty)
	case rules.SpellDirectDamage, rules.SpellDDNoVariance, rules.SpellBolt, rules.SpellLifedrain:
		if sp.Type == rules.SpellBolt {
			if p, ok := target.(*Player); ok {
				inFront := angleOff(p.Body.Yaw, angleTo(p.Body.X, p.Body.Y, cx, cy)) <= frontArc
				if rules.BoltBlocked(rules.BlockChance(p.combatant(z.W.Cat), inFront), z.rng.Float64()) {
					z.event(protocol.Event{K: "block", From: target.entityID(), To: caster.entityID(), Text: sp.Name}, tx, ty)
					return true
				}
			}
		}
		dmg := rules.SpellDamage(sp.Damage, rules.SpellVariance(sp.Type, z.rng.Float64()), z.spellResist(target, sp.DamageType))
		z.event(protocol.Event{K: "spell", From: caster.entityID(), To: target.entityID(), Text: sp.Name, Amount: dmg}, tx, ty)
		z.hurt(target, caster, dmg)
		if sp.Type == rules.SpellLifedrain && caster.fightable(z) {
			z.heal(caster, rules.DrainHeal(dmg, sp.LifeDrain))
		}
	default:
		if kind := effectKind(sp.Type); kind != "" {
			z.addEffect(target, caster, sp, kind)
		}
	}
	if sub := z.W.Cat.Spells[sp.SubSpell]; sub != nil && sub.ID != sp.ID && target.fightable(z) {
		z.castSpell(caster, target, sub)
	}
	return true
}

// addEffect puts a lasting effect on a target, or restarts the same spell's.
func (z *Zone) addEffect(target, caster combatant, sp *SpellDef, kind string) {
	list := effectsOf(target)
	var e *effect
	for _, o := range *list {
		if o.spell.ID == sp.ID {
			e = o
		}
	}
	if e == nil {
		e = &effect{spell: sp, kind: kind}
		*list = append(*list, e)
	}
	e.caster, e.casterLevel = caster, levelOf(caster)
	e.ends = z.now + float64(sp.Duration)
	if kind == effDoT {
		_, e.every = rules.DoTTicks(sp.Duration, sp.Frequency)
		e.next = z.now + e.every
	}
	tx, ty := target.position()
	z.event(protocol.Event{K: "effect", From: caster.entityID(), To: target.entityID(), Text: sp.Name}, tx, ty)
	if kind == effStun || kind == effMez {
		z.stopActing(target)
	}
	z.effectsChanged(target)
}

// stopActing ends what a held combatant was doing.
func (z *Zone) stopActing(c combatant) {
	switch t := c.(type) {
	case *Player:
		t.pending = -1
		t.queue, t.queuedSteps = nil, 0
	case *NPC:
		t.speedNow = 0
	}
}

// heal restores health to a combatant, up to its maximum.
func (z *Zone) heal(c combatant, amount int) {
	switch t := c.(type) {
	case *Player:
		t.hp = math.Min(float64(t.maxHP()), t.hp+float64(max(0, amount)))
	case *NPC:
		t.HP = rules.Healed(t.HP, t.MaxHP, amount)
	}
}

// hurt does spell damage to a combatant: it breaks a mesmerize, makes a
// creature hate the one who did it, and kills at zero.
func (z *Zone) hurt(target, by combatant, amount int) {
	if amount <= 0 || !target.fightable(z) {
		return
	}
	z.damaged(target)
	switch t := target.(type) {
	case *Player:
		t.combatUntil = z.now + CombatLinger
		t.hp -= float64(amount)
		if t.hp <= 0 {
			n, _ := by.(*NPC)
			z.killPlayer(t, n)
		}
	case *NPC:
		t.HP -= amount
		if p, ok := by.(*Player); ok {
			t.hurtBy[p.ID] += amount
			if t.tagger == 0 {
				t.tagger = p.ID
			}
		}
		if by != nil {
			z.provoke(t, by, float64(amount))
		}
		if t.HP <= 0 {
			z.killNPC(t)
		}
	}
}

// damaged ends whatever damage breaks on a combatant: a mesmerize.
func (z *Zone) damaged(c combatant) {
	list := effectsOf(c)
	kept := (*list)[:0]
	changed := false
	for _, e := range *list {
		if rules.BrokenByDamage(e.spell.Type) {
			changed = true
			continue
		}
		kept = append(kept, e)
	}
	*list = kept
	if changed {
		z.effectsChanged(c)
	}
}

// clearEffects removes every effect, as death does.
func (z *Zone) clearEffects(c combatant) {
	if l := effectsOf(c); len(*l) > 0 {
		*l = nil
		z.effectsChanged(c)
	}
}

// tickEffects runs every lasting effect in the zone: damage over time hits,
// and effects that have run their course end.
func (z *Zone) tickEffects() {
	for _, p := range z.players {
		z.tickEffectsOf(p)
	}
	for _, n := range z.npcs {
		if len(n.effects) > 0 {
			z.tickEffectsOf(n)
		}
	}
}

func (z *Zone) tickEffectsOf(c combatant) {
	list := effectsOf(c)
	if len(*list) == 0 {
		return
	}
	changed := false
	for _, e := range *list {
		if e.kind != effDoT {
			continue
		}
		for e.every > 0 && z.now+1e-9 >= e.next && e.next <= e.ends+1e-9 && c.fightable(z) {
			e.next += e.every
			if e.caster != nil && !e.caster.fightable(z) {
				e.caster = nil
			}
			dmg := rules.SpellDamage(e.spell.Damage, 1, z.spellResist(c, e.spell.DamageType))
			x, y := c.position()
			var from int64
			if e.caster != nil {
				from = e.caster.entityID()
			}
			z.event(protocol.Event{K: "spell", From: from, To: c.entityID(), Text: e.spell.Name, Amount: dmg}, x, y)
			z.hurt(c, e.caster, dmg)
		}
	}
	list = effectsOf(c) // a death clears the list
	kept := (*list)[:0]
	for _, e := range *list {
		if z.now+1e-9 >= e.ends {
			changed = true
			continue
		}
		kept = append(kept, e)
	}
	*list = kept
	if changed {
		z.effectsChanged(c)
	}
}

// effectsChanged tells a player what their effects now are.
func (z *Zone) effectsChanged(c combatant) {
	p, ok := c.(*Player)
	if !ok {
		return
	}
	m := protocol.Effects{T: protocol.TEffects, List: []protocol.EffectView{}, Speed: speedFactor(p), Held: held(p)}
	for _, e := range p.effects {
		m.List = append(m.List, protocol.EffectView{Name: e.spell.Name, Kind: e.kind, Icon: e.spell.Icon,
			Left: math.Max(0, e.ends-z.now), Of: float64(e.spell.Duration)})
	}
	p.send(m)
}
