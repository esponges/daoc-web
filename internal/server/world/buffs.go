package world

import (
	"fmt"
	"math"
	"sort"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// Effect kinds the Healer and Shaman buffs leave, as the client draws them.
const (
	effHaste    = "haste"    // quicker swings
	effShield   = "shield"   // a damage shield
	effEndRegen = "endregen" // extra endurance on each tick
	effPowRegen = "powregen" // extra power on each tick
)

// strongest is the largest value among a player's effects of a kind: the
// strongest of its kind counts.
func (p *Player) strongest(kind string) int {
	best := 0
	for _, e := range p.effects {
		if e.kind == kind {
			best = max(best, int(e.spell.Value))
		}
	}
	return best
}

// haste is the percent a player's swings are shortened by.
func (p *Player) haste() int { return p.strongest(effHaste) }

// damageShield is the strongest damage shield on a player, nil when they have none.
func (p *Player) damageShield() *effect {
	var best *effect
	for _, e := range p.effects {
		if e.kind == effShield && (best == nil || e.spell.Damage > best.spell.Damage) {
			best = e
		}
	}
	return best
}

// shieldHits is a damage shield on a player striking a creature that has just
// hit them in melee.
func (z *Zone) shieldHits(p *Player, n *NPC) {
	e := p.damageShield()
	if e == nil || n.dead() {
		return
	}
	dmg := rules.ShieldDamage(e.spell.Damage, z.spellResist(n, e.spell.DamageType))
	z.event(protocol.Event{K: "spell", From: p.ID, To: n.ID, Text: e.spell.Name, Amount: dmg}, n.Body.X, n.Body.Y)
	z.hurt(n, p, dmg)
	if n.dead() {
		return
	}
}

// cure ends the effects of the kind a cure spell names on its target.
func (z *Zone) cure(caster, target combatant, sp *SpellDef) {
	list := effectsOf(target)
	kept := (*list)[:0]
	cured := 0
	for _, e := range *list {
		if rules.Cures(sp.Type, e.kind) {
			cured++
			if e.cb != nil {
				z.dropConc(e.cb)
			}
			continue
		}
		kept = append(kept, e)
	}
	*list = kept
	tx, ty := target.position()
	z.event(protocol.Event{K: "spell", From: caster.entityID(), To: target.entityID(), Text: sp.Name}, tx, ty)
	if cured > 0 {
		z.effectsChanged(target)
		return
	}
	if cp, ok := caster.(*Player); ok {
		sys(cp, "There is nothing for "+sp.Name+" to cure.")
	}
}

// spreadHeal heals the caster's group, the amount shared out in proportion to
// what each member is missing.
func (z *Zone) spreadHeal(caster *Player, sp *SpellDef) {
	total := int(sp.Value)
	if sc, ok := z.scaleFor(caster, sp); ok {
		total = rules.PlayerHeal(sp.Value, sc.spec, sc.spellLevel, z.rng.Float64())
	}
	total = sickened(total, caster)
	members := z.castGroupTargets(caster, sp)
	sort.SliceStable(members, func(i, j int) bool { return members[i].entityID() < members[j].entityID() })
	missing := make([]int, len(members))
	for i, m := range members {
		if p, ok := m.(*Player); ok {
			missing[i] = max(0, p.maxHP()-int(math.Ceil(p.hp)))
		}
	}
	for i, share := range rules.SpreadShares(missing, total) {
		if share <= 0 {
			continue
		}
		x, y := members[i].position()
		z.heal(members[i], share)
		z.event(protocol.Event{K: "spell", From: caster.ID, To: members[i].entityID(), Text: sp.Name, Amount: -share}, x, y)
	}
	sys(caster, fmt.Sprintf("Your %s spreads %d hit points among your group.", sp.Name, total))
}
