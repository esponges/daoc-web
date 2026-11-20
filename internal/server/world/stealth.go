package world

import (
	"fmt"
	"math"

	"daocweb/internal/server/rules"
)

// Stealth, from the rules-stealth capability. A hidden player carries an
// effect of kind effStealth with no end; their speed is cut, other players see
// them only inside a detection range, aggressive creatures notice them only
// inside theirs, and attacking, casting, shooting or taking damage ends it.

const effStealth = "stealth" // the holder is hidden

const effCamouflage = "camouflage" // a hidden Hunter's Camouflage: Detect Hidden does not help against it

// hidden says whether a player is in stealth.
func (p *Player) hidden() bool { return hasKind(p, effStealth) }

// stealthSpec is the player's Stealth spec level, 0 for a class without it.
func (p *Player) stealthSpec(cat *Catalog) int {
	if !p.hasSpec(cat, "Stealth") {
		return 0
	}
	return p.combatSpec("Stealth")
}

// stealthFactor is the share of its speed a hidden combatant moves at.
func stealthFactor(c combatant) float64 {
	for _, e := range *effectsOf(c) {
		if e.kind == effStealth {
			return e.spell.Value
		}
	}
	return 1
}

// Stealth is the player's order to hide, or to stop hiding.
func (z *Zone) Stealth(p *Player) {
	if p.dead || p.ride != nil {
		return
	}
	if p.stealthAt != 0 && z.now < p.stealthAt+rules.StealthChangeGap {
		p.errMsg("stealth", fmt.Sprintf("You must wait %d more seconds to change your stealth.", int(math.Ceil(p.stealthAt+rules.StealthChangeGap-z.now))))
		return
	}
	if p.hidden() {
		z.unstealth(p, "You are no longer hidden.")
		return
	}
	spec := p.stealthSpec(z.W.Cat)
	switch {
	case spec == 0:
		p.errMsg("stealth", "You have not learned to hide.")
		return
	case z.now < p.combatUntil || p.attacking:
		p.errMsg("stealth", "You can't hide while in combat!")
		return
	case p.casting != nil || p.draw != nil || p.song != nil:
		p.errMsg("stealth", "You can't hide while casting, shooting or playing.")
		return
	case held(p):
		p.errMsg("stealth", "You can't hide while you are held.")
		return
	}
	for _, n := range z.npcs {
		if n.dead() || n.gone || n.Friendly || !n.fightable(z) {
			continue
		}
		if r := rules.StealthTooClose(n.Level, spec, false); r > 0 && math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y) <= r {
			p.errMsg("stealth", "You are too close to an enemy to hide!")
			return
		}
	}
	p.stealthAt = z.now
	// Hiding ends speed songs and enhancements, and what was fighting you
	// loses you.
	kept := p.effects[:0]
	for _, e := range p.effects {
		if e.kind != effSpeed {
			kept = append(kept, e)
		}
	}
	p.effects = kept
	p.effects = append(p.effects, &effect{spell: &SpellDef{Name: "Stealth", Type: "Stealth", Value: rules.StealthSpeed(spec, p.Char.Level)},
		kind: effStealth, ends: math.Inf(1), caster: p})
	for _, n := range z.npcs {
		if n.target == combatant(p) && !n.dead() {
			n.target, n.hate = nil, nil
			n.state = npcReturn
		}
	}
	sys(p, "You are now hidden.")
	z.effectsChanged(p)
}

// unstealth ends a player's stealth, with a line to them.
func (z *Zone) unstealth(p *Player, text string) {
	list := &p.effects
	kept := (*list)[:0]
	was := false
	for _, e := range *list {
		if e.kind == effCamouflage {
			continue // camouflage ends with the stealth it hides in
		}
		if e.kind == effStealth {
			was = true
			continue
		}
		kept = append(kept, e)
	}
	*list = kept
	if !was {
		return
	}
	p.stealthAt = z.now
	if text != "" {
		sys(p, text)
	}
	z.effectsChanged(p)
}

// seesHidden says whether a player looking at a hidden player can see them: they
// are inside the detection range for the viewer's level and the hider's Stealth.
func (z *Zone) seesHidden(viewer, hider *Player) bool {
	if viewer == hider || z.W.groupShares(viewer, hider) {
		return true
	}
	d := math.Hypot(viewer.Body.X-hider.Body.X, viewer.Body.Y-hider.Body.Y)
	r := rules.PlayerDetectRange(viewer.Char.Level, hider.stealthSpec(z.W.Cat), viewer.sh.abilities["Detect Hidden"] > 0, hider.counters())
	return d <= r
}

// noticesHidden says whether a creature notices a hidden player: the player is
// inside its detection radius.
func (z *Zone) noticesHidden(n *NPC, p *Player) bool {
	d := math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y)
	return d <= rules.CreatureDetectRadius(n.Level, p.stealthSpec(z.W.Cat), false, p.counters())
}

// groupShares says whether two players are in one group (who see each other hidden).
func (w *World) groupShares(a, b *Player) bool {
	for _, id := range w.groupOf(a.Char.ID) {
		if id == b.Char.ID {
			return true
		}
	}
	return false
}

// counters says whether a hidden player has what defeats Detect Hidden: it
// themselves, or Camouflage.
func (p *Player) counters() bool {
	return p.sh.abilities["Detect Hidden"] > 0 || hasKind(p, effCamouflage)
}

// Camouflage turns a hidden Hunter's camouflage on or off.
func (z *Zone) Camouflage(p *Player) {
	if p.sh.abilities["Camouflage"] == 0 {
		p.errMsg("camouflage", "You have not learned Camouflage.")
		return
	}
	if !p.hidden() {
		p.errMsg("camouflage", "You must be hidden to use Camouflage.")
		return
	}
	if hasKind(p, effCamouflage) {
		kept := p.effects[:0]
		for _, e := range p.effects {
			if e.kind != effCamouflage {
				kept = append(kept, e)
			}
		}
		p.effects = kept
		sys(p, "Your camouflage is gone.")
	} else {
		p.effects = append(p.effects, &effect{spell: &SpellDef{Name: "Camouflage", Type: "Camouflage"}, kind: effCamouflage, ends: math.Inf(1), caster: p})
		sys(p, "You are camouflaged.")
	}
	z.effectsChanged(p)
}
