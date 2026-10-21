package world

import (
	"fmt"
	"math"
	"sort"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// bookSpell is a spell a player knows: from a spec line they have trained far
// enough.
type bookSpell struct {
	Spec  string
	Level int // the spec level it opens at
	Spell *SpellDef
}

// playerCast is a spell a player is casting: at whom, and when it lands.
type playerCast struct {
	spell  *SpellDef
	target combatant
	ends   float64
	x, y   float64 // where they stood, for moving away to interrupt
}

// castMoveLimit is how far, in units, a caster may drift before the cast is
// interrupted.
const castMoveLimit = 25.0

// spellSupported says whether the server can apply a spell as a player's: its
// type is in the first slice and its target is one this slice resolves.
func spellSupported(sp *SpellDef) bool {
	if !rules.ImplementedPlayerSpell(sp.Type) {
		return false
	}
	switch sp.Target {
	case "Enemy", "Self", "Realm", "Group":
		return true
	}
	return false
}

// buildBook lists the spells a player knows: for each spec line their class
// trains, the spells that open at or below the spec's level, by spec then level.
func (w *World) buildBook(p *Player) []bookSpell {
	var out []bookSpell
	if !p.sh.class.Casts || w.Cat == nil {
		return nil
	}
	for _, spec := range w.Cat.TrainableSpecs(p.Char.Class) {
		lv := p.specLevel(spec)
		for _, l := range w.Cat.SpellsFor(p.Char.Class, spec) {
			if sp := w.Cat.Spells[l.Spell]; sp != nil && l.Level <= lv {
				out = append(out, bookSpell{Spec: spec, Level: l.Level, Spell: sp})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Spec != out[j].Spec {
			return out[i].Spec < out[j].Spec
		}
		return out[i].Level < out[j].Level
	})
	return out
}

// powerCost is what a spell draws from this caster's pool: a percentage of it,
// or a fixed amount when the data gives a negative power.
func (p *Player) powerCost(sp *SpellDef) int {
	if sp.Power < 0 {
		return -sp.Power
	}
	return rules.SpellPowerCost(sp.Power, p.sh.maxPower)
}

// bookMsg is the spellbook as the protocol carries it.
func (p *Player) bookMsg(cat *Catalog) protocol.Book {
	m := protocol.Book{T: protocol.TBook, Lines: []protocol.BookLine{}}
	for _, b := range p.book {
		if n := len(m.Lines); n == 0 || m.Lines[n-1].Spec != b.Spec {
			m.Lines = append(m.Lines, protocol.BookLine{Spec: b.Spec, Name: cat.SpecName(b.Spec), Spells: []protocol.BookSpell{}})
		}
		l := &m.Lines[len(m.Lines)-1]
		l.Spells = append(l.Spells, protocol.BookSpell{ID: b.Spell.ID, Name: b.Spell.Name, Level: b.Level, Cost: p.powerCost(b.Spell),
			Cast: b.Spell.CastTime, Range: int(b.Spell.Range), Type: b.Spell.Type, Target: b.Spell.Target, OK: spellSupported(b.Spell),
			Icon: b.Spell.Icon, Recast: int(b.Spell.Recast)})
	}
	return m
}

// refreshBook rebuilds the spellbook and tells the player when it changed,
// naming the spells they have just learned.
func (w *World) refreshBook(p *Player) {
	old := map[int]bool{}
	for _, b := range p.book {
		old[b.Spell.ID] = true
	}
	p.book = w.buildBook(p)
	if !p.welcomed {
		return
	}
	m := p.bookMsg(w.Cat)
	if fmt.Sprint(m) == fmt.Sprint(p.bookSent) {
		return
	}
	first := p.bookSent.T == ""
	p.bookSent = m
	p.send(m)
	if first {
		return
	}
	for _, b := range p.book {
		if !old[b.Spell.ID] {
			sys(p, "You learn "+b.Spell.Name+".")
		}
	}
}

// bookHas finds a spell in the player's book.
func (p *Player) bookHas(id int) *SpellDef {
	for _, b := range p.book {
		if b.Spell.ID == id {
			return b.Spell
		}
	}
	return nil
}

// Cast is a player's request to cast a spell from their book, at the target
// they name (or the one they hold, or themselves for a spell on the caster).
func (z *Zone) Cast(p *Player, spellID int, targetID int64) {
	if p.dead || p.ride != nil {
		return
	}
	if held(p) {
		p.errMsg("cast", "You cannot cast while you are held.")
		return
	}
	if p.casting != nil {
		p.errMsg("cast", "You are already casting a spell.")
		return
	}
	sp := p.bookHas(spellID)
	if sp == nil {
		p.errMsg("cast", "You do not know that spell.")
		return
	}
	if !spellSupported(sp) {
		p.errMsg("cast", sp.Name+" does not work yet.")
		return
	}
	if z.now < p.spellReady[sp.ID] {
		p.errMsg("cast", sp.Name+" is not ready yet.")
		return
	}
	target := z.castTarget(p, sp, targetID)
	if target == nil {
		return
	}
	if cost := p.powerCost(sp); p.power < float64(cost) {
		p.errMsg("cast", "You don't have enough power.")
		return
	}
	tx, ty := target.position()
	if target != combatant(p) && math.Hypot(tx-p.Body.X, ty-p.Body.Y) > sp.Range {
		p.errMsg("cast", "Your target is too far away.")
		return
	}
	p.casting = &playerCast{spell: sp, target: target, ends: rules.CastEnds(z.now, sp.CastTime), x: p.Body.X, y: p.Body.Y}
	p.pending = -1
	z.event(protocol.Event{K: "cast", From: p.ID, To: target.entityID(), Text: sp.Name, Dur: sp.CastTime}, p.Body.X, p.Body.Y)
	if sp.CastTime <= 0 {
		z.continuePlayerCast(p)
	}
}

// castTarget resolves who a spell is aimed at, or tells the player why not and
// returns nil. A spell on an enemy needs a living creature; one on the realm
// takes a player or the caster; the rest fall on the caster (and their group).
func (z *Zone) castTarget(p *Player, sp *SpellDef, id int64) combatant {
	if id == 0 {
		id = p.target
	}
	switch sp.Target {
	case "Enemy":
		n := z.byID[id]
		if n == nil || n.dead() || n.gone {
			p.errMsg("cast", "You need an enemy target.")
			return nil
		}
		if n.Friendly {
			p.errMsg("cast", "You cannot attack that.")
			return nil
		}
		return n
	case "Realm":
		if o := z.players[id]; o != nil && o.Arch.Realm == p.Arch.Realm && o.fightable(z) {
			return o
		}
		return p
	}
	return p
}

// castGroupTargets is who a group spell reaches: the caster and the members of
// their group in the zone, alive and within the spell's range.
func (z *Zone) castGroupTargets(p *Player, sp *SpellDef) []combatant {
	out := []combatant{p}
	member := map[int64]bool{}
	for _, id := range z.W.groupOf(p.Char.ID) {
		member[id] = true
	}
	for _, o := range z.players {
		if o == p || !member[o.Char.ID] || !o.fightable(z) {
			continue
		}
		if math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) <= sp.Range {
			out = append(out, o)
		}
	}
	return out
}

// continuePlayerCast runs a cast in progress: it stops if the caster moved
// away or is held, and lands the spell when its time is up.
func (z *Zone) continuePlayerCast(p *Player) {
	c := p.casting
	if c == nil {
		return
	}
	if held(p) {
		z.cancelPlayerCast(p, "")
		return
	}
	if math.Hypot(p.Body.X-c.x, p.Body.Y-c.y) > castMoveLimit {
		z.interruptPlayerCast(p, "You move and your spell is interrupted.")
		return
	}
	if z.now+1e-9 < c.ends {
		return
	}
	p.casting = nil
	sp := c.spell
	cost := p.powerCost(sp)
	tx, ty := c.target.position()
	dist := math.Hypot(tx-p.Body.X, ty-p.Body.Y)
	if c.target == combatant(p) {
		dist = 0
	}
	if !rules.CastLands(c.target.fightable(z), dist, sp.Range) {
		z.event(protocol.Event{K: "castfail", From: p.ID, Text: sp.Name}, p.Body.X, p.Body.Y)
		sys(p, "Your target is no longer there, and "+sp.Name+" fails.")
		return
	}
	if p.power < float64(cost) {
		sys(p, "You don't have enough power.")
		return
	}
	p.power -= float64(cost)
	p.spellReady[sp.ID] = z.now + sp.Recast
	if sp.Hostile() {
		p.combatUntil = z.now + CombatLinger
	}
	targets := []combatant{c.target}
	if sp.Target == "Group" {
		targets = z.castGroupTargets(p, sp)
	}
	for _, t := range targets {
		z.castSpell(p, t, sp)
	}
}

// interruptPlayerCast stops a cast with no effect and no power spent; the
// spell is ready again after a moment.
func (z *Zone) interruptPlayerCast(p *Player, text string) {
	c := p.casting
	if c == nil || !rules.Interrupts(c.spell.Uninterruptible) {
		return
	}
	z.cancelPlayerCast(p, text)
	p.spellReady[c.spell.ID] = z.now + rules.CastRetry
	z.event(protocol.Event{K: "interrupt", From: p.ID, Text: c.spell.Name}, p.Body.X, p.Body.Y)
}

// cancelPlayerCast drops a cast with no effect.
func (z *Zone) cancelPlayerCast(p *Player, text string) {
	if p.casting == nil {
		return
	}
	p.casting = nil
	if text != "" {
		sys(p, text)
	}
}

// buffBonuses is what the spell buffs on a player add to each stat and to
// armour factor. Buffs of one kind do not stack: the strongest counts, as
// spells of one line replace each other.
func (p *Player) buffBonuses() (stats rules.Stats, af int) {
	for _, e := range p.effects {
		if e.kind != effBuff {
			continue
		}
		if e.spell.Type == rules.SpellArmourBuff {
			af = max(af, int(e.spell.Value))
			continue
		}
		for s, v := range rules.BuffedStats(e.spell.Type, e.spell.Value) {
			stats[s] = max(stats[s], v)
		}
	}
	return stats, af
}

// damageAdd is the damage the player's damage-add spells put on each landed
// blow: the strongest one.
func (p *Player) damageAdd() int {
	add := 0
	for _, e := range p.effects {
		if e.kind == effAdd {
			add = max(add, int(e.spell.Damage))
		}
	}
	return add
}

// hasBuff says whether any spell buff is on the player.
func (p *Player) hasBuff() bool {
	for _, e := range p.effects {
		if e.kind == effBuff {
			return true
		}
	}
	return false
}
