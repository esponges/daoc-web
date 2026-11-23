package world

import (
	"fmt"
	"math"

	"daocweb/internal/server/rules"
)

// Effect kinds of the timed abilities, as the client draws them.
const (
	effBerserk = "berserk" // more melee damage, less defence
	effCharge  = "charge"  // faster, and no slow takes hold
	effQuick   = "quickcast"
)

// stanceIDs give each ability's effect its own spell number, below nothing the
// data uses.
var stanceIDs = map[string]int{rules.AbBerserk: -101, rules.AbCharge: -102, rules.AbQuickCast: -103}

// groundTarget is a point on the ground a player has set for abilities that
// are aimed there, kept with the zone it was set in.
type groundTarget struct {
	zone    *Zone
	x, y, z float64
}

// active returns the ground target when it is set in the zone the player is in.
func (p *Player) groundPoint() *groundTarget {
	if g := p.ground; g != nil && g.zone == p.Zone {
		return g
	}
	return nil
}

// SetGroundTarget sets or clears (with no point) the player's ground target:
// refused beyond the range, or where they cannot see.
func (z *Zone) SetGroundTarget(p *Player, x, y, zz float64, clear bool) {
	if p.dead {
		return
	}
	if clear {
		p.ground = nil
		return
	}
	if !rules.GroundTargetOK(math.Hypot(x-p.Body.X, y-p.Body.Y)) {
		p.errMsg("groundtarget", "Your ground target is too far away!")
		return
	}
	if !z.canSee(p.Body.X, p.Body.Y, p.Body.Z, x, y, zz) {
		p.errMsg("groundtarget", "You cannot see that spot.")
		return
	}
	p.ground = &groundTarget{zone: z, x: x, y: y, z: zz}
}

// useStance runs Berserk, Charge, QuickCast and Distraction; it returns false
// for any other ability.
func (z *Zone) useStance(p *Player, key string, level int) bool {
	var dur, reuse float64
	var kind, said string
	switch key {
	case rules.AbBerserk:
		dur, reuse, kind, said = rules.BerserkSeconds, rules.BerserkReuse, effBerserk, "You go berserk!"
	case rules.AbCharge:
		dur, reuse, kind, said = rules.ChargeSeconds, rules.ChargeReuse, effCharge, "You charge ahead!"
	case rules.AbQuickCast:
		dur, reuse, kind, said = rules.QuickCastWindow, rules.QuickCastReuse, effQuick, "Your next spell will cast instantly."
	case rules.AbDistraction:
		z.distract(p)
		return true
	default:
		return false
	}
	if w := rules.StanceWait(z.now, p.ab.ready[key]); w > 0 {
		sys(p, fmt.Sprintf("%s is not ready for another %d seconds.", key, int(math.Ceil(w))))
		return true
	}
	if hasKind(p, kind) {
		sys(p, key+" is already in force.")
		return true
	}
	if p.ab.ready == nil {
		p.ab.ready = map[string]float64{}
	}
	p.ab.ready[key] = z.now + reuse
	sp := &SpellDef{ID: stanceIDs[key], Name: key, Type: "Stance", Target: "Self", Duration: int(dur), Value: float64(level)}
	z.addEffect(p, p, sp, kind)
	sys(p, said)
	return true
}

// stanceBonuses puts Berserk on the player's combatant: more damage, and a
// share of evade, parry and block given up.
func (p *Player) stanceBonuses(cb *rules.Combatant) {
	for _, e := range p.effects {
		if e.kind == effBerserk {
			cb.DamageBonus += rules.BerserkDamage(int(e.spell.Value))
			cb.DefencePenalty = rules.BerserkDefenceLoss
		}
	}
}

// distract turns the idle creatures near the ground target to face it.
func (z *Zone) distract(p *Player) {
	g := p.groundPoint()
	if g == nil {
		sys(p, "You must have a ground target to use this!")
		return
	}
	if w := rules.StanceWait(z.now, p.ab.ready[rules.AbDistraction]); w > 0 {
		sys(p, fmt.Sprintf("Distraction is not ready for another %d seconds.", int(math.Ceil(w))))
		return
	}
	if p.ab.ready == nil {
		p.ab.ready = map[string]float64{}
	}
	p.ab.ready[rules.AbDistraction] = z.now + rules.DistractionReuse
	turned := 0
	for _, n := range z.npcs {
		if n.dead() || n.gone || n.Friendly || n.target != nil {
			continue
		}
		if rules.DistractionTurns(math.Hypot(n.Body.X-g.x, n.Body.Y-g.y)) {
			n.Body.Yaw = angleTo(n.Body.X, n.Body.Y, g.x, g.y)
			turned++
		}
	}
	sys(p, fmt.Sprintf("You distract %d creatures.", turned))
}

// dangerSense warns a player with Danger Sense of hidden enemy-realm players
// who come near, once for each until they have been away for a while. Nothing
// of another realm can be in a zone with a player yet, so it is quiet today.
func (z *Zone) dangerSense(p *Player) {
	if p.sh.abilities[rules.AbDangerSense] == 0 || p.dead {
		return
	}
	for _, o := range z.players {
		if o == p || o.Arch.Realm == p.Arch.Realm || !o.hidden() {
			continue
		}
		if math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) > rules.DangerSenseRange {
			continue
		}
		if p.sensed == nil {
			p.sensed = map[int64]float64{}
		}
		if last, seen := p.sensed[o.ID]; !seen || z.now-last >= rules.DangerSenseQuiet {
			sys(p, "You sense danger nearby!")
		}
		p.sensed[o.ID] = z.now
	}
}

// groundMsg is the ground target as the protocol carries it, nil when none is set.
func (p *Player) groundMsg() []float64 {
	if g := p.groundPoint(); g != nil {
		return []float64{math.Round(g.x), math.Round(g.y), math.Round(g.z)}
	}
	return nil
}

// castSeconds is how long a player's cast takes: none for the next spell after
// QuickCast, the spell's own time otherwise.
func (z *Zone) castSeconds(p *Player, sp *SpellDef) float64 {
	if hasKind(p, effQuick) && sp.CastTime > 0 && sp.Target != "Corpse" && sp.Type != rules.SpellResurrect {
		return 0
	}
	return sp.CastTime
}

// spendQuickCast ends QuickCast once a spell has gone off under it.
func (z *Zone) spendQuickCast(p *Player) {
	for i, e := range p.effects {
		if e.kind == effQuick {
			p.effects = append(p.effects[:i], p.effects[i+1:]...)
			z.effectsChanged(p)
			return
		}
	}
}
