package world

import (
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// castState is a spell a creature is casting: at whom, and when it lands.
type castState struct {
	spell  *SpellDef
	target combatant
	ends   float64
}

// learnSpells gives a creature the spells its template lists that the server
// can cast: offensive procs ride on its blows, the rest it casts.
func (n *NPC) learnSpells(cat *Catalog) {
	for _, s := range cat.CastableSpells(n.Spawn.Template) {
		if s.Type == rules.SpellOffensiveProc {
			n.procs = append(n.procs, s)
		} else {
			n.spells = append(n.spells, s)
		}
	}
}

// startCasting is a creature's combat decision about spells: it begins a
// cast and returns true, or returns false to swing or close in.
func (z *Zone) startCasting(n *NPC) bool {
	if len(n.spells) == 0 || n.target == nil {
		return false
	}
	tx, ty := n.target.position()
	dist := math.Hypot(tx-n.Body.X, ty-n.Body.Y)
	hurt := z.hurtAlly(n, 0)
	opts := make([]rules.CastOption, len(n.spells))
	for i, s := range n.spells {
		opts[i] = rules.CastOption{Heal: s.Type == rules.SpellHeal, Ready: z.now >= n.spellReady[s.ID], Range: s.Range}
		if effectKind(s.Type) != "" {
			opts[i].Already = hasEffect(n.target, s.ID)
		}
	}
	i := rules.ChooseCast(opts, dist, hurt != nil)
	if i < 0 {
		return false
	}
	sp := n.spells[i]
	target := n.target
	if opts[i].Heal {
		if target = z.hurtAlly(n, sp.Range); target == nil {
			return false
		}
	}
	n.casting = &castState{spell: sp, target: target, ends: rules.CastEnds(z.now, sp.CastTime)}
	n.speedNow = 0
	z.event(protocol.Event{K: "cast", From: n.ID, To: target.entityID(), Text: sp.Name, Dur: sp.CastTime}, n.Body.X, n.Body.Y)
	if sp.CastTime <= 0 {
		z.continueCasting(n)
	}
	return true
}

// continueCasting holds a casting creature still, facing its target, and
// lands the spell when its time is up: or fails it if the target is gone.
func (z *Zone) continueCasting(n *NPC) {
	c := n.casting
	n.speedNow = 0
	tx, ty := c.target.position()
	n.Body.Yaw = turnToward(n.Body.Yaw, math.Atan2(tx-n.Body.X, -(ty-n.Body.Y)), 6*TickDT)
	if z.now+1e-9 < c.ends {
		return
	}
	n.casting = nil
	dist := math.Hypot(tx-n.Body.X, ty-n.Body.Y)
	if c.target == combatant(n) {
		dist = 0
	}
	if !rules.CastLands(c.target.fightable(z), dist, c.spell.Range) {
		n.spellReady[c.spell.ID] = z.now + rules.CastRetry
		z.event(protocol.Event{K: "castfail", From: n.ID, Text: c.spell.Name}, n.Body.X, n.Body.Y)
		return
	}
	n.spellReady[c.spell.ID] = z.now + c.spell.Recast
	z.castSpell(n, c.target, c.spell)
}

// interruptCasting stops a creature's spell when it is attacked, unless the
// spell is uninterruptible; the spell is ready again after a moment.
func (z *Zone) interruptCasting(n *NPC) {
	c := n.casting
	if c == nil || !rules.Interrupts(c.spell.Uninterruptible) {
		return
	}
	z.cancelCasting(n)
	z.event(protocol.Event{K: "interrupt", From: n.ID, Text: c.spell.Name}, n.Body.X, n.Body.Y)
}

// cancelCasting drops a cast with no effect.
func (z *Zone) cancelCasting(n *NPC) {
	if c := n.casting; c != nil {
		n.spellReady[c.spell.ID] = z.now + rules.CastRetry
		n.casting = nil
	}
}

// hurtAlly is the creature a healer would heal: itself, or a packmate (same
// name) within reach below the heal line; nil when nobody is. A reach of 0
// takes any distance, for the decision; the heal itself checks its range.
func (z *Zone) hurtAlly(n *NPC, reach float64) combatant {
	if rules.Hurt(n.HP, n.MaxHP) {
		return n
	}
	for _, o := range z.npcs {
		if o == n || o.dead() || o.Name != n.Name || !rules.Hurt(o.HP, o.MaxHP) {
			continue
		}
		if reach > 0 && math.Hypot(o.Body.X-n.Body.X, o.Body.Y-n.Body.Y) > reach {
			continue
		}
		return o
	}
	return nil
}

// hasEffect says whether a spell's lasting effect is on a combatant.
func hasEffect(c combatant, spellID int) bool {
	for _, e := range *effectsOf(c) {
		if e.spell.ID == spellID {
			return true
		}
	}
	return false
}

// creatureProcs fires a creature's offensive procs after one of its blows
// lands: each with its chance (the proc spell's frequency, a percent), casting
// the spell it carries on the target.
func (z *Zone) creatureProcs(n *NPC, target combatant) {
	for _, p := range n.procs {
		sp := z.W.Cat.Spells[int(p.Value)]
		if sp == nil || z.rng.Float64()*100 >= float64(p.Frequency) {
			continue
		}
		z.castSpell(n, target, sp)
	}
}
