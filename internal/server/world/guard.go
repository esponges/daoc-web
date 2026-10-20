package world

import (
	"math"
	"strings"
)

// GuardRange is how far a guard watches when its spawn gives no aggro range.
const GuardRange = 750.0

// guardWatch is how often a calm guard looks around, in seconds.
const guardWatch = 0.5

// isGuard says whether a spawn's class type makes it a town guard.
func isGuard(classType string) bool { return strings.HasSuffix(classType, ".GameGuard") }

// watch is a calm guard looking for a creature to fight: an aggressive one of
// another realm within its range, or one fighting a player of its realm who
// is within its range. It takes the nearest.
func (n *NPC) watch(z *Zone) {
	if z.now < n.watchAt {
		return
	}
	n.watchAt = z.now + guardWatch
	var best *NPC
	bd := math.Inf(1)
	for _, o := range z.npcs {
		if o == n || o.dead() || o.hidden || o.Friendly || o.Aggro <= 0 || o.Spawn.Realm == n.Spawn.Realm {
			continue
		}
		d := math.Hypot(o.Body.X-n.Body.X, o.Body.Y-n.Body.Y)
		if d > n.guardRange && !n.defends(o.target) {
			continue
		}
		if d < bd {
			best, bd = o, d
		}
	}
	if best != nil {
		n.engage(best, z)
	}
}

// defends says whether a guard stands up for this combatant: a player of its
// realm within its range.
func (n *NPC) defends(c combatant) bool {
	p, ok := c.(*Player)
	return ok && p.Arch.Realm == n.Spawn.Realm && math.Hypot(p.Body.X-n.Body.X, p.Body.Y-n.Body.Y) <= n.guardRange
}

// unsharedDamage counts a guard's damage to a creature: it takes its share of
// the kill's experience, and that share goes to no one.
func unsharedDamage(by combatant, t *NPC, amount int) {
	if g, ok := by.(*NPC); ok && g.Guard {
		t.hurtBy[g.ID] += amount
	}
}
