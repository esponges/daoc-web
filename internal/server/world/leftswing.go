package world

import (
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// leftBlows is how many extra blows the left hand makes this round, and the
// share of damage each does: a Left Axe fighter swings it every round at the
// left axe share; a Hand to Hand fighter with claws in both hands gets extra
// blows by chance (rules-melee).
func (z *Zone) leftBlows(p *Player) (int, float64) {
	l := p.left
	if l == nil || p.weapon == nil {
		return 0, 1
	}
	switch l.ObjectType {
	case rules.ObjLeftAxe:
		spec := p.specLevel("Left Axe")
		if !p.sh.hasLeftAxe() {
			return 0, 1
		}
		return rules.LeftSwings(spec), rules.LeftAxeDamage(spec)
	case rules.ObjHandToHand:
		if p.weapon.ObjectType != rules.ObjHandToHand {
			return 0, 1
		}
		return rules.HandToHandExtra(p.specLevel("Hand to Hand"), z.rng.Float64()), 1
	}
	return 0, 1
}

// hasLeftAxe says whether the class may swing a left axe: it holds the weaponry.
func (s sheet) hasLeftAxe() bool { return s.abilities["Weaponry: Left Axes"] > 0 }

// leftSwings makes the left hand's blows against a creature, each rolled on its
// own for hit, defences and damage.
func (z *Zone) leftSwings(p *Player, n *NPC) {
	count, share := z.leftBlows(p)
	for i := 0; i < count && !n.dead() && !n.gone; i++ {
		z.event(protocol.Event{K: "swing", From: p.ID, To: n.ID}, p.Body.X, p.Body.Y)
		cb := p.combatant(z.W.Cat)
		cb.Unarmed = false
		cb.Weapon = weaponOf(p.left)
		cb.Weapon.Condition = p.condOf(SlotLeft)
		cb.Spec = p.combatSpec(weaponSpec(p.left.ObjectType))
		sw := rules.Swing{Attacker: cb, Defender: &n.cb, Attackers: 1, Effectiveness: share,
			InFront: angleOff(n.Body.Yaw, angleTo(n.Body.X, n.Body.Y, p.Body.X, p.Body.Y)) <= frontArc}
		o := rules.Resolve(sw, z.rng)
		if o.Result != rules.Hit {
			if o.Result == rules.Missed {
				z.event(protocol.Event{K: "miss", From: p.ID, To: n.ID}, n.Body.X, n.Body.Y)
			} else {
				z.event(protocol.Event{K: o.Result.String(), From: n.ID, To: p.ID}, n.Body.X, n.Body.Y)
			}
			z.provoke(n, p, 0)
			continue
		}
		amount := sickened(o.Damage, p)
		z.wearItem(p, SlotLeft)
		z.event(protocol.Event{K: "hit", From: p.ID, To: n.ID, Amount: amount, Crit: o.Critical > 0, Text: "your left hand"}, n.Body.X, n.Body.Y)
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
		z.firePoison(p, n, SlotLeft)
	}
}
