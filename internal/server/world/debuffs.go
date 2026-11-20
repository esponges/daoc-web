package world

import (
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// Effect kinds the debuffs, control spells and wards leave. The kinds named
// "mez", "poison", "disease" and "nearsight" are the ones the cures end.
const (
	effDebuff     = "debuff"     // stat points lowered
	effResDebuff  = "resdebuff"  // a resist lowered
	effCombatSlow = "combatslow" // swings made slower
	effDisease    = "disease"
	effNearsight  = "nearsight"
	effNSImmune   = "nsimmune" // hidden: a nearsight has just ended
	effFatigue    = "fatigue"  // endurance costs raised
	effBlade      = "bladeturn"
	effAblative   = "ablative"
	effAbsorb     = "absorb" // armour absorption raised
	effMezDur     = "mezdur" // the holder's mesmerizes last longer
)

// debuffKinds are the kinds whose duration the target's resist shortens.
func debuffKind(kind string) bool {
	switch kind {
	case effDebuff, effResDebuff, effCombatSlow, effDisease, effNearsight, effFatigue:
		return true
	}
	return false
}

// creatureTakes says whether a creature can hold an effect of a kind.
func creatureTakes(kind string) bool {
	switch kind {
	case effDoT, effSlow, effStun, effMez, effProc, effBuff, effDebuff, effResDebuff, effCombatSlow,
		effDisease, effNearsight, effNSImmune:
		return true
	}
	return false
}

// strongestValue is the largest |value| of a combatant's effects of a kind:
// debuffs and buffs of one kind do not stack.
func strongestValue(c combatant, kind string) float64 {
	best := 0.0
	for _, e := range *effectsOf(c) {
		if e.kind == kind {
			best = math.Max(best, math.Abs(e.spell.Value))
		}
	}
	return best
}

func hasKind(c combatant, kind string) bool {
	for _, e := range *effectsOf(c) {
		if e.kind == kind {
			return true
		}
	}
	return false
}

// debuffStats is what the stat debuffs on a combatant take off each stat, the
// strongest per stat, and a disease takes from strength by its share on top.
func debuffStats(c combatant) map[rules.Stat]int {
	out := map[rules.Stat]int{}
	for _, e := range *effectsOf(c) {
		if e.kind != effDebuff {
			continue
		}
		for s, v := range rules.DebuffedStats(e.spell.Type, e.spell.Value) {
			out[s] = max(out[s], v)
		}
	}
	return out
}

// resistDebuff is what resist debuffs take off a combatant's resist to a
// damage type, in percent.
func resistDebuff(c combatant, damageType int) float64 {
	best := 0.0
	for _, e := range *effectsOf(c) {
		if e.kind != effResDebuff {
			continue
		}
		if dt, ok := rules.ResistDebuffType(e.spell.Type); ok && dt == damageType {
			best = math.Max(best, math.Abs(e.spell.Value))
		} else if e.spell.Type == rules.SpellDDWithDebuff && e.spell.DamageType == damageType {
			best = math.Max(best, math.Abs(e.spell.Value))
		}
	}
	return best
}

// spellRange is how far a combatant's spell reaches: nearsight cuts it.
func spellRange(c combatant, sp *SpellDef) float64 {
	return rules.NearsightRange(sp.Range, strongestValue(c, effNearsight))
}

// refreshNPC works a creature's fighting numbers out again from its level's
// and the buffs and debuffs it holds.
func refreshNPC(n *NPC) {
	if n.baseCB.Level == 0 {
		return
	}
	cb := n.baseCB
	stats := cb.Stats
	buff := map[rules.Stat]int{}
	af := 0
	for _, e := range n.effects {
		if e.kind != effBuff {
			continue
		}
		if e.spell.Type == rules.SpellArmourBuff {
			af = max(af, int(e.spell.Value))
			continue
		}
		for s, v := range rules.BuffedStats(e.spell.Type, e.spell.Value) {
			buff[s] = max(buff[s], v)
		}
	}
	for s := range stats {
		stats[s] += min(buff[rules.Stat(s)], rules.BuffStatCap(n.Level))
	}
	for s, v := range debuffStats(n) {
		stats[s] = max(1, stats[s]-v)
	}
	if hasKind(n, effDisease) {
		stats[rules.STR] = rules.DiseasedStrength(stats[rules.STR])
	}
	cb.Stats = stats
	cb.BuffAF = float64(min(af, rules.BuffAFCap(n.Level)))
	n.cb = cb
	n.Swing = rules.SlowedSwing(rules.CreatureSwing(&n.cb), int(strongestValue(n, effCombatSlow)))
}

// nsImmuneAfter notes that a nearsight has ended on a combatant.
func (z *Zone) nsImmuneAfter(c combatant) {
	*effectsOf(c) = append(*effectsOf(c), &effect{
		spell: &SpellDef{Name: "nearsight immunity", Type: "NearsightImmunity"}, kind: effNSImmune, ends: z.now + rules.NearsightImmunity})
}

// castDebuffing handles the spell types of this file that have more to them
// than a lasting effect; it returns whether the type was one of them, and
// whether the spell landed.
func (z *Zone) castDebuffing(caster, target combatant, sp *SpellDef) (handled, landed bool) {
	tx, ty := target.position()
	switch sp.Type {
	case rules.SpellLifeTransfer:
		heal, cost := rules.LifeTransfer(maxHealthOf(caster), int(math.Ceil(healthOf(caster))), sp.Value, hasKind(target, effDisease))
		if cp, ok := caster.(*Player); ok {
			cp.hp = math.Max(1, cp.hp-float64(cost))
		}
		z.heal(target, heal)
		z.event(protocol.Event{K: "spell", From: caster.entityID(), To: target.entityID(), Text: sp.Name, Amount: -heal}, tx, ty)
		return true, true
	case rules.SpellAmnesia:
		chance := rules.AmnesiaChance
		if sp.AmnesiaChance > 0 {
			chance = float64(sp.AmnesiaChance)
		}
		if z.rng.Float64()*100 >= chance {
			z.event(protocol.Event{K: "resist", From: caster.entityID(), To: target.entityID(), Text: sp.Name}, tx, ty)
			return true, false
		}
		if n, ok := target.(*NPC); ok {
			n.hate, n.target = nil, nil
			z.cancelCasting(n)
			n.state = npcReturn
			z.event(protocol.Event{K: "spell", From: caster.entityID(), To: target.entityID(), Text: sp.Name}, tx, ty)
			return true, true
		}
		return true, false
	case rules.SpellConfusion:
		if !rules.ConfusionTakesHold(sp.Value, z.rng.Float64()) {
			z.event(protocol.Event{K: "resist", From: caster.entityID(), To: target.entityID(), Text: sp.Name}, tx, ty)
			return true, false
		}
		if p, ok := target.(*Player); ok {
			// Against a player it only interrupts what they were doing.
			z.interruptPlayerCast(p, "You are confused and lose your concentration.")
			z.event(protocol.Event{K: "effect", From: caster.entityID(), To: target.entityID(), Text: sp.Name}, tx, ty)
			return true, true
		}
		z.addEffect(target, caster, sp, effConfusion)
		return true, true
	}
	return false, false
}

// effConfusion is a confused creature's effect: it picks a new enemy every few
// seconds. It is a creature-only effect and not a client effect kind.
const effConfusion = "confusion"

// healthOf is a combatant's current health.
func healthOf(c combatant) float64 {
	switch t := c.(type) {
	case *Player:
		return t.hp
	case *NPC:
		return float64(t.HP)
	}
	return 0
}

// confuse runs a confused creature's pulse: it drops its target and picks a
// random enemy within reach.
func (z *Zone) confuse(n *NPC, e *effect) {
	var pool []combatant
	for _, p := range z.players {
		if p.fightable(z) && math.Hypot(p.Body.X-n.Body.X, p.Body.Y-n.Body.Y) <= rules.ConfusionRadius {
			pool = append(pool, p)
		}
	}
	if rules.ConfusionAttacksAlly(e.spell.Value, z.rng.Float64()) {
		for _, o := range z.npcs {
			if o != n && o.fightable(z) && math.Hypot(o.Body.X-n.Body.X, o.Body.Y-n.Body.Y) <= rules.ConfusionRadius {
				pool = append(pool, o)
			}
		}
	}
	if len(pool) == 0 {
		return
	}
	pick := pool[z.rng.Intn(len(pool))]
	n.target = pick
	n.addHate(pick, 1)
	n.state = npcChase
}

// wardBlow is what a bladeturn and an ablative do to a creature's blow that
// has hit a player: it returns the damage that gets through, and whether the
// blow was turned whole.
func (z *Zone) wardBlow(p *Player, n *NPC, dmg int) (int, bool) {
	changed := false
	defer func() {
		if changed {
			z.effectsChanged(p)
		}
	}()
	for i, e := range p.effects {
		if e.kind != effBlade {
			continue
		}
		p.effects = append(p.effects[:i], p.effects[i+1:]...)
		changed = true
		sys(p, "The blow is turned by your "+e.spell.Name+"!")
		return 0, true
	}
	for i, e := range p.effects {
		if e.kind != effAblative {
			continue
		}
		taken, left := rules.AblativeAbsorb(dmg, rules.AblativePercent(e.spell.Damage), e.pool)
		e.pool = left
		dmg -= taken
		if left <= 0 {
			p.effects = append(p.effects[:i], p.effects[i+1:]...)
		}
		changed = true
		break
	}
	return dmg, false
}
