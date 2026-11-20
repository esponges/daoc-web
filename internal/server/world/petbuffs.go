package world

import (
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// What a creature's lasting spells do in its own fights. A Spiritmaster or
// Bonedancer puts them on their pet: a damage add and a damage shield on its
// blows, an ablative that eats a share of what hits it. Haste, the mesmerize
// duration and the procs are in refreshNPC, addEffect and creatureProcs.

// strongestSpell is the effect of a kind whose spell has the largest damage,
// or nil.
func strongestSpell(c combatant, kind string) *effect {
	var best *effect
	for _, e := range *effectsOf(c) {
		if e.kind == kind && (best == nil || e.spell.Damage > best.spell.Damage) {
			best = e
		}
	}
	return best
}

// creatureDamageAdd is what the damage-add spells on a creature put on a blow
// of blow damage against a target: the strongest one, rolled for its swing.
func (z *Zone) creatureDamageAdd(n *NPC, target combatant, blow int) int {
	e := strongestSpell(n, effAdd)
	if e == nil {
		return 0
	}
	return rules.DamageAdd(e.spell.Damage, n.Level, n.Swing, blow, z.rng.Float64(), z.spellResist(target, e.spell.DamageType))
}

// creatureWard takes an ablative on a creature's share of a blow, and returns
// what is left of it. The ward ends with its pool.
func creatureWard(t *NPC, dmg int) int {
	list := &t.effects
	for i, e := range *list {
		if e.kind != effAblative {
			continue
		}
		taken, left := rules.AblativeAbsorb(dmg, rules.AblativePercent(e.spell.Damage), e.pool)
		e.pool = left
		if left <= 0 {
			*list = append((*list)[:i], (*list)[i+1:]...)
		}
		return dmg - taken
	}
	return dmg
}

// creatureShieldHits is a damage shield on a creature striking the one that
// has just hit it in melee.
func (z *Zone) creatureShieldHits(t *NPC, attacker combatant) {
	e := strongestSpell(t, effShield)
	if e == nil || !attacker.fightable(z) {
		return
	}
	dmg := rules.ShieldDamage(e.spell.Damage, z.spellResist(attacker, e.spell.DamageType))
	ax, ay := attacker.position()
	z.event(protocol.Event{K: "spell", From: t.ID, To: attacker.entityID(), Text: e.spell.Name, Amount: dmg}, ax, ay)
	z.hurt(attacker, t, dmg)
}
