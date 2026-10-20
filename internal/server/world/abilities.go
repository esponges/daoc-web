package world

import (
	"math"
	"strconv"
	"strings"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// abilityState is what a Warrior currently has set with Guard, Protect,
// Intercept and Engage. The first three hold the entity id of the member, 0
// for none; engage holds the creature's.
type abilityState struct {
	guard, protect, intercept int64
	engage                    int64
	interceptReadyAt          float64
	attackedAt                float64
}

var abilityNames = map[string]string{"guard": rules.AbGuard, "protect": rules.AbProtect,
	"intercept": rules.AbIntercept, "engage": rules.AbEngage}

// UseAbility runs a player's use of Guard, Protect, Intercept or Engage on
// their target. Using one again on the same target ends it.
func (z *Zone) UseAbility(p *Player, name string) {
	key := abilityNames[strings.ToLower(name)]
	if key == "" {
		sys(p, "There is no ability called "+name+".")
		return
	}
	lv := p.sh.abilities[key]
	switch {
	case p.dead:
		sys(p, "You cannot do that while dead.")
		return
	case lv == 0:
		sys(p, "You do not have "+key+".")
		return
	case held(p):
		sys(p, "You cannot act while you are held.")
		return
	}
	a := &p.ab
	if key == rules.AbEngage {
		z.useEngage(p, a)
		return
	}
	t := z.players[p.target]
	if t == nil || t == p || t.dead {
		sys(p, key+": target a member of your group.")
		return
	}
	if !z.W.sameGroup(p.Char.ID, t.Char.ID) {
		sys(p, key+": "+t.Char.Name+" must be in your group.")
		return
	}
	if math.Hypot(t.Body.X-p.Body.X, t.Body.Y-p.Body.Y) > rules.AbilityRange {
		sys(p, key+": "+t.Char.Name+" is too far away.")
		return
	}
	var slot *int64
	switch key {
	case rules.AbGuard:
		slot = &a.guard
		if p.shield == nil {
			sys(p, "Guard needs a shield.")
			return
		}
	case rules.AbProtect:
		slot = &a.protect
	case rules.AbIntercept:
		slot = &a.intercept
	}
	if *slot == t.ID {
		*slot = 0
		sys(p, "You stop "+strings.ToLower(key)+"ing "+t.Char.Name+".")
		return
	}
	if key == rules.AbIntercept {
		if z.now < a.interceptReadyAt {
			sys(p, "Intercept is not ready for another "+strconv.Itoa(int(math.Ceil(a.interceptReadyAt-z.now)))+" seconds.")
			return
		}
		a.interceptReadyAt = z.now + rules.InterceptCooldown
	}
	*slot = t.ID
	sys(p, "You "+strings.ToLower(key)+" "+t.Char.Name+".")
	sys(t, p.Char.Name+" "+strings.ToLower(key)+"s you.")
}

func itoa(n int) string {
	return strings.TrimSpace(strings.Replace(strings.Repeat(" ", 0)+itoaRaw(n), " ", "", -1))
}

func itoaRaw(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}

func (z *Zone) useEngage(p *Player, a *abilityState) {
	n := z.byID[p.target]
	if n == nil || n.dead() {
		sys(p, "Engage: target an enemy.")
		return
	}
	if a.engage == n.ID {
		a.engage = 0
		sys(p, "You stop engaging "+n.title()+".")
		return
	}
	if p.shield == nil {
		sys(p, "Engage needs a shield.")
		return
	}
	if a.attackedAt != 0 && !rules.EngageAllowed(z.now-a.attackedAt) {
		sys(p, "You cannot engage: you were attacked a moment ago.")
		return
	}
	a.engage = n.ID
	sys(p, "You engage "+n.title()+".")
}

// abilityGuarders are the players in the zone who guard, protect or intercept
// for the member, picked by which of the three slots matches.
func (z *Zone) abilityGuarders(member *Player, pick func(*abilityState) int64) []*Player {
	var out []*Player
	for _, o := range z.players {
		if o != member && !o.dead && pick(&o.ab) == member.ID {
			out = append(out, o)
		}
	}
	return out
}

func distance(a, b *Player) float64 { return math.Hypot(a.Body.X-b.Body.X, a.Body.Y-b.Body.Y) }

// interceptFor returns the Warrior who takes this member's next blow, ending
// their Intercept; nil when there is none in reach.
func (z *Zone) interceptFor(member *Player) *Player {
	for _, w := range z.abilityGuarders(member, func(a *abilityState) int64 { return a.intercept }) {
		if rules.InterceptTakes(distance(w, member)) {
			w.ab.intercept = 0
			sys(w, "You intercept the blow for "+member.Char.Name+".")
			sys(member, w.Char.Name+" intercepts the blow for you.")
			return w
		}
	}
	return nil
}

// guardBlocks tells whether a guarding Warrior blocks this blow for the member.
func (z *Zone) guardBlocks(n *NPC, member *Player) bool {
	for _, w := range z.abilityGuarders(member, func(a *abilityState) int64 { return a.guard }) {
		if w.shield == nil {
			continue
		}
		cb := w.combatant(z.W.Cat)
		if rules.GuardBlocks(distance(w, member), rules.BlockChance(cb, true), z.rng.Float64()) {
			z.event(protocol.Event{K: "block", From: w.ID, To: n.ID}, w.Body.X, w.Body.Y)
			sys(w, "You block the blow meant for "+member.Char.Name+".")
			sys(member, w.Char.Name+" blocks the blow for you.")
			return true
		}
	}
	return false
}

// engageBlocks tells whether the Warrior's Engage blocks this blow, taking its
// endurance, and ends the Engage when they run out.
func (z *Zone) engageBlocks(n *NPC, p *Player, inFront bool) bool {
	if p.ab.engage != n.ID {
		return false
	}
	if p.end < rules.EngageCost {
		p.ab.engage = 0
		sys(p, "You are too tired to keep engaging.")
		return false
	}
	if !rules.EngageBlocks(inFront, p.end, z.rng.Float64()) {
		return false
	}
	p.end -= rules.EngageCost
	z.event(protocol.Event{K: "block", From: p.ID, To: n.ID}, p.Body.X, p.Body.Y)
	p.lastDef = defence{kind: "block", at: z.now}
	return true
}

// protectHate draws a creature's anger to the Warriors protecting a member it
// has just damaged.
func (z *Zone) protectHate(n *NPC, member *Player, damage int) {
	for _, w := range z.abilityGuarders(member, func(a *abilityState) int64 { return a.protect }) {
		if distance(w, member) <= rules.ProtectRange {
			n.addHate(w, rules.ProtectHate(w.sh.abilities[rules.AbProtect], damage))
		}
	}
}

// sameGroup says whether two characters are in one group.
func (w *World) sameGroup(a, b int64) bool {
	for _, m := range w.groupOf(a) {
		if m == b {
			return true
		}
	}
	return false
}
