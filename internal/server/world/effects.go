package world

import (
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// effect is a lasting spell on a combatant: a damage over time, a slow, a
// stun, a mesmerize or an offensive proc. The zone ticks it.
type effect struct {
	spell       *SpellDef
	kind        string
	caster      combatant // nil once it has gone
	casterLevel int
	ends        float64
	next        float64 // the next damage-over-time tick
	every       float64
	// A player's damage over time ticks with the stat and spec the caster
	// had when it landed.
	scaled     bool
	stat, spec int
	cb         *concBuff // set while a caster holds it by concentration
	pool       int       // what is left of an ablative armour
}

// Effect kinds, as the client draws them.
const (
	effDoT  = "dot"
	effSlow = "slow"
	effStun = "stun"
	effMez  = "mez"
	effProc = "proc"
	effBuff = "buff"
	effAdd  = "dmgadd"

	// What songs and chants leave: a speed boost, extra health on each tick, a
	// resist.
	effSpeed  = "speed"
	effRegen  = "regen"
	effResist = "resist"
)

// effectKind is the kind of lasting effect a spell type leaves, "" for none.
func effectKind(spellType string) string {
	switch spellType {
	case rules.SpellDamageOverTime:
		return effDoT
	case rules.SpellSpeedDecrease:
		return effSlow
	case rules.SpellStun:
		return effStun
	case rules.SpellMesmerize:
		return effMez
	case rules.SpellOffensiveProc:
		return effProc
	case rules.SpellStrConBuff, rules.SpellDexQuiBuff, rules.SpellStrengthBuff, rules.SpellArmourBuff,
		rules.SpellConstitutionBuff, rules.SpellDexterityBuff, rules.SpellAcuityBuff:
		return effBuff
	case rules.SpellCombatSpeedBuff:
		return effHaste
	case rules.SpellDamageShield:
		return effShield
	case rules.SpellEnduranceRegen:
		return effEndRegen
	case rules.SpellPowerRegen:
		return effPowRegen
	case rules.SpellDamageAdd:
		return effAdd
	case rules.SpellSpeedEnhancement:
		return effSpeed
	case rules.SpellHealthRegenBuff:
		return effRegen
	case rules.SpellStrengthDebuff, rules.SpellDexterityDebuff, rules.SpellStrConDebuff, rules.SpellDexQuiDebuff:
		return effDebuff
	case rules.SpellBodyResistDebuff, rules.SpellColdResistDebuff, rules.SpellEnergyResistDebuff,
		rules.SpellHeatResistDebuff, rules.SpellMatterResistDebuff, rules.SpellSpiritResistDebuff, rules.SpellDDWithDebuff:
		return effResDebuff
	case rules.SpellCombatSpeedDebuff:
		return effCombatSlow
	case rules.SpellDisease:
		return effDisease
	case rules.SpellNearsight:
		return effNearsight
	case rules.SpellFatigueDebuff:
		return effFatigue
	case rules.SpellBladeturn:
		return effBlade
	case rules.SpellAblativeArmor:
		return effAblative
	case rules.SpellArmorAbsorbBuff:
		return effAbsorb
	case rules.SpellMezDurationBuff:
		return effMezDur
	case rules.SpellUnbreakableSlow, rules.SpellDamageSpeedDecr:
		return effSlow
	case rules.SpellBodyResistBuff, rules.SpellColdResistBuff, rules.SpellEnergyResistBuff,
		rules.SpellHeatResistBuff, rules.SpellMatterResistBuff, rules.SpellSpiritResistBuff:
		return effResist
	}
	return ""
}

// effectsOf is where a combatant keeps its effects.
func effectsOf(c combatant) *[]*effect {
	switch t := c.(type) {
	case *Player:
		return &t.effects
	case *NPC:
		return &t.effects
	}
	return nil
}

// maxHealthOf is a combatant's most health.
func maxHealthOf(c combatant) int {
	switch t := c.(type) {
	case *Player:
		return t.maxHP()
	case *NPC:
		return t.MaxHP
	}
	return 1
}

func levelOf(c combatant) int {
	switch t := c.(type) {
	case *Player:
		return t.Char.Level
	case *NPC:
		return t.Level
	}
	return 1
}

// held says whether a combatant is stunned or mesmerized.
func held(c combatant) bool {
	for _, e := range *effectsOf(c) {
		if e.kind == effStun || e.kind == effMez {
			return true
		}
	}
	return false
}

// speedFactor is the share of its speed a combatant may move at.
func speedFactor(c combatant) float64 {
	var slows []float64
	boost := 1.0
	for _, e := range *effectsOf(c) {
		switch e.kind {
		case effSlow:
			slows = append(slows, rules.SlowFactor(e.spell.Value))
		case effDisease:
			slows = append(slows, rules.SlowFactor(rules.DiseaseSlowPct))
		case effSpeed:
			boost = math.Max(boost, rules.BoostFactor(e.spell.Value))
		}
	}
	return rules.WithBoost(rules.SpeedFactor(held(c), slows...), boost) * stealthFactor(c)
}

// resistBonus is the item bonus type that raises a resist, by damage type.
var resistBonus = map[int]int{10: 11, 11: 12, rules.DamageCrush: 13, 12: 14, 13: 15, 14: 16, rules.DamageSlash: 17, 15: 18, rules.DamageThrus: 19}

// spellResist is a combatant's resist, in percent, to a damage type: its
// race's, and for a player, what their worn items add.
func (z *Zone) spellResist(c combatant, damageType int) float64 {
	switch t := c.(type) {
	case *NPC:
		return t.resists[damageType] - resistDebuff(t, damageType)
	case *Player:
		r := z.W.Cat.Resists[t.Char.Race][damageType] + rules.ResistBuffBonus(damageType, t.resistBuffs())
		for s := SlotWeapon; s < SlotCount; s++ {
			if it := t.inv[s]; it != nil {
				if x := z.W.Cat.Items[it.ID]; x != nil {
					for _, b := range x.Bonuses {
						if b.Type == resistBonus[damageType] {
							r += float64(b.Amount)
						}
					}
				}
			}
		}
		return r
	}
	return 0
}

// casterScale is what a player's casting stat and their spec in the spell's
// line add to a spell they cast from their book.
type casterScale struct{ stat, spec, spellLevel int }

// scaleFor says how a player's stat and spec scale a spell. A creature, or a
// spell outside the player's book (a weapon proc, a style's spell), is not
// scaled.
func (z *Zone) scaleFor(caster combatant, sp *SpellDef) (casterScale, bool) {
	p, ok := caster.(*Player)
	if !ok || !p.sh.class.Casts || z.W.Cat == nil {
		return casterScale{}, false
	}
	for _, spec := range z.W.Cat.TrainableSpecs(p.Char.Class) {
		for _, l := range z.W.Cat.SpellsFor(p.Char.Class, spec) {
			if l.Spell == sp.ID {
				return casterScale{stat: p.sh.stats[p.sh.class.Acuity], spec: p.specLevel(spec), spellLevel: l.Level}, true
			}
		}
	}
	return casterScale{}, false
}

// castSpell lands a spell from a caster on a target: a hostile one by the hit
// roll, then its damage, heal or lasting effect, then any sub-spell. It
// returns whether it landed.
func (z *Zone) castSpell(caster, target combatant, sp *SpellDef) bool {
	return z.castSpellAs(caster, target, sp, levelOf(caster), true)
}

// castSpellAs is castSpell with the hit rolled as a caster of a given level
// (a weapon proc rolls as its item), or not rolled at all (a style's spell).
func (z *Zone) castSpellAs(caster, target combatant, sp *SpellDef, level int, roll bool) bool {
	if !target.fightable(z) {
		return false
	}
	cx, cy := caster.position()
	tx, ty := target.position()
	if sp.Hostile() && roll {
		if z.rng.Float64() >= rules.SpellHitChance(level, levelOf(target)) {
			z.event(protocol.Event{K: "resist", From: caster.entityID(), To: target.entityID(), Text: sp.Name, FX: sp.ClientEffect}, tx, ty)
			return false
		}
	}
	if handled, landed := z.castDebuffing(caster, target, sp); handled {
		return landed
	}
	if handled, landed := z.castPetSpell(caster, target, sp); handled {
		return landed
	}
	switch sp.Type {
	case rules.SpellHeal:
		amount := int(sp.Value)
		if sp.Value < 0 {
			amount = rules.PercentHeal(sp.Value, maxHealthOf(caster))
		} else if sc, ok := z.scaleFor(caster, sp); ok {
			amount = rules.PlayerHeal(sp.Value, sc.spec, sc.spellLevel, z.rng.Float64())
		}
		amount = sickened(amount, caster)
		z.heal(target, amount)
		z.event(protocol.Event{K: "spell", From: caster.entityID(), To: target.entityID(), Text: sp.Name, FX: sp.ClientEffect, Amount: -amount}, tx, ty)
	case rules.SpellDirectDamage, rules.SpellDDNoVariance, rules.SpellBolt, rules.SpellLifedrain, rules.SpellDDWithDebuff, rules.SpellDamageSpeedDecr:
		if sp.Type == rules.SpellBolt {
			if p, ok := target.(*Player); ok {
				inFront := angleOff(p.Body.Yaw, angleTo(p.Body.X, p.Body.Y, cx, cy)) <= frontArc
				if rules.BoltBlocked(rules.BlockChance(p.combatant(z.W.Cat), inFront), z.rng.Float64()) {
					z.event(protocol.Event{K: "block", From: target.entityID(), To: caster.entityID(), Text: sp.Name, FX: sp.ClientEffect}, tx, ty)
					return true
				}
			}
		}
		dmg := rules.SpellDamage(sp.Damage, rules.SpellVariance(sp.Type, z.rng.Float64()), z.spellResist(target, sp.DamageType))
		if sc, ok := z.scaleFor(caster, sp); ok && sp.Type != rules.SpellDDNoVariance {
			dmg = rules.PlayerSpellDamage(sp.Damage, sc.stat, sc.spec, levelOf(caster), levelOf(target), z.rng.Float64(), z.spellResist(target, sp.DamageType))
		}
		dmg = sickened(dmg, caster)
		z.event(protocol.Event{K: "spell", From: caster.entityID(), To: target.entityID(), Text: sp.Name, FX: sp.ClientEffect, Amount: dmg}, tx, ty)
		z.hurt(target, caster, dmg)
		if sp.Type == rules.SpellLifedrain && caster.fightable(z) {
			z.heal(caster, rules.DrainHeal(dmg, sp.LifeDrain))
		}
		if (sp.Type == rules.SpellDDWithDebuff || sp.Type == rules.SpellDamageSpeedDecr) && target.fightable(z) {
			z.addEffect(target, caster, sp, effectKind(sp.Type))
		}
	case rules.SpellCurePoison, rules.SpellCureDisease, rules.SpellCureMezz, rules.SpellCureNearsight:
		z.cure(caster, target, sp)
	case rules.SpellSpreadHeal:
		if cp, ok := caster.(*Player); ok {
			z.spreadHeal(cp, sp)
		}
	default:
		if kind := effectKind(sp.Type); kind != "" {
			tp, isPlayer := target.(*Player)
			if !isPlayer && !creatureTakes(kind) {
				return false // songs and wards are for players
			}
			if kind == effNearsight && (hasKind(target, effNearsight) || hasKind(target, effNSImmune)) {
				if cp, ok := caster.(*Player); ok {
					sys(cp, "Your target is immune to that now.")
				}
				return false
			}
			if kind == effSpeed && !rules.BoostActive(z.now < tp.combatUntil) {
				return false // a speed song does not take hold in combat
			}
			z.addEffect(target, caster, sp, kind)
		}
	}
	if sub := z.W.Cat.Spells[sp.SubSpell]; sub != nil && sub.ID != sp.ID && target.fightable(z) {
		z.castSpellAs(caster, target, sub, level, false)
	}
	return true
}

// addEffect puts a lasting effect on a target, or restarts the same spell's.
func (z *Zone) addEffect(target, caster combatant, sp *SpellDef, kind string) {
	list := effectsOf(target)
	var e *effect
	for _, o := range *list {
		if o.spell.ID == sp.ID {
			e = o
		}
	}
	refresh := e != nil
	if e == nil {
		e = &effect{spell: sp, kind: kind}
		*list = append(*list, e)
	}
	if cp, ok := caster.(*Player); ok && e.cb != nil && e.cb.caster != cp {
		z.dropConc(e.cb) // another caster's cast takes the buff over
		e.cb = nil
	}
	e.caster, e.casterLevel = caster, levelOf(caster)
	dur := float64(sp.Duration)
	if sp.pulsing() {
		dur = sp.effectSeconds()
	}
	if debuffKind(kind) {
		dur = rules.DebuffDuration(dur, 0, z.spellResist(target, sp.DamageType))
	}
	if kind == effMez {
		dur = rules.MezDuration(dur, strongestValue(target, effMezDur))
	}
	if kind == effAblative {
		e.pool = int(sp.Value)
	}
	e.ends = z.now + dur
	if kind == effDoT {
		_, e.every = rules.DoTTicks(sp.Duration, sp.Frequency)
		e.next = z.now + e.every
		sc, ok := z.scaleFor(caster, sp)
		e.scaled, e.stat, e.spec = ok, sc.stat, sc.spec
	}
	if cp, ok := caster.(*Player); ok && holdsConc(sp) {
		if tp, ok := target.(*Player); ok {
			z.noteConc(cp, tp, e)
		}
	}
	tx, ty := target.position()
	if !(refresh && sp.pulsing()) { // a song pulsing on a member again is not news
		z.event(protocol.Event{K: "effect", From: caster.entityID(), To: target.entityID(), Text: sp.Name, FX: sp.ClientEffect}, tx, ty)
	}
	if kind == effStun || kind == effMez {
		z.stopActing(target)
	}
	if n, ok := target.(*NPC); ok {
		refreshNPC(n)
	}
	if kind == effConfusion {
		e.next = z.now + rules.ConfusionPulse
	}
	z.effectsChanged(target)
}

// stopActing ends what a held combatant was doing.
func (z *Zone) stopActing(c combatant) {
	switch t := c.(type) {
	case *Player:
		t.pending = -1
		t.queue, t.queuedSteps = nil, 0
	case *NPC:
		t.speedNow = 0
		z.cancelCasting(t)
	}
}

// heal restores health to a combatant, up to its maximum.
func (z *Zone) heal(c combatant, amount int) {
	if hasKind(c, effDisease) {
		amount = rules.DiseasedHeal(amount)
	}
	switch t := c.(type) {
	case *Player:
		t.hp = math.Min(float64(t.maxHP()), t.hp+float64(max(0, amount)))
	case *NPC:
		t.HP = rules.Healed(t.HP, t.MaxHP, amount)
	}
}

// hurt does spell damage to a combatant: it breaks a mesmerize, makes a
// creature hate the one who did it, and kills at zero.
func (z *Zone) hurt(target, by combatant, amount int) {
	if amount <= 0 || !target.fightable(z) {
		return
	}
	z.damaged(target)
	switch t := target.(type) {
	case *Player:
		t.combatUntil = z.now + CombatLinger
		t.hp -= float64(amount)
		if t.hp <= 0 {
			n, _ := by.(*NPC)
			z.killPlayer(t, n)
		}
	case *NPC:
		t.HP -= amount
		if p := ownerOf(by); p != nil {
			t.hurtBy[p.ID] += amount
			if t.tagger == 0 {
				t.tagger = p.ID
			}
		}
		unsharedDamage(by, t, amount)
		if by != nil {
			z.provoke(t, by, float64(amount))
		}
		if t.HP <= 0 {
			z.killNPC(t)
		}
	}
}

// damaged ends whatever damage breaks on a combatant: a mesmerize.
func (z *Zone) damaged(c combatant) {
	if p, ok := c.(*Player); ok && p.hidden() {
		z.unstealth(p, "You are no longer hidden.")
	}
	if p, ok := c.(*Player); ok && p.casting != nil {
		z.interruptPlayerCast(p, "You are interrupted.")
	}
	if p, ok := c.(*Player); ok && p.draw != nil {
		z.cancelDraw(p, "You are hit and lose your aim.")
	}
	list := effectsOf(c)
	kept := (*list)[:0]
	changed := false
	for _, e := range *list {
		if rules.BrokenByDamage(e.spell.Type) {
			changed = true
			continue
		}
		kept = append(kept, e)
	}
	*list = kept
	if changed {
		z.effectsChanged(c)
	}
}

// clearEffects removes every effect, as death does.
func (z *Zone) clearEffects(c combatant) {
	if l := effectsOf(c); len(*l) > 0 {
		for _, e := range *l {
			if e.cb != nil {
				z.dropConc(e.cb) // what a caster held on them is let go
			}
		}
		*l = nil
		z.effectsChanged(c)
	}
}

// tickEffects runs every lasting effect in the zone: damage over time hits,
// and effects that have run their course end.
func (z *Zone) tickEffects() {
	for _, p := range z.players {
		z.tickEffectsOf(p)
	}
	for _, n := range z.npcs {
		if len(n.effects) > 0 {
			z.tickEffectsOf(n)
		}
	}
}

func (z *Zone) tickEffectsOf(c combatant) {
	list := effectsOf(c)
	if len(*list) == 0 {
		return
	}
	changed := false
	for _, e := range *list {
		if n, ok := c.(*NPC); ok && e.kind == effConfusion && z.now+1e-9 >= e.next && n.fightable(z) {
			e.next += rules.ConfusionPulse
			z.confuse(n, e)
		}
		if e.kind != effDoT {
			continue
		}
		for e.every > 0 && z.now+1e-9 >= e.next && e.next <= e.ends+1e-9 && c.fightable(z) {
			e.next += e.every
			if e.caster != nil && !e.caster.fightable(z) {
				e.caster = nil
			}
			dmg := rules.SpellDamage(e.spell.Damage, 1, z.spellResist(c, e.spell.DamageType))
			if e.scaled {
				dmg = rules.PlayerDoTDamage(e.spell.Damage, e.stat, e.spec, levelOf(c), z.rng.Float64(), z.spellResist(c, e.spell.DamageType))
			}
			x, y := c.position()
			var from int64
			if e.caster != nil {
				from = e.caster.entityID()
			}
			z.event(protocol.Event{K: "spell", From: from, To: c.entityID(), Text: e.spell.Name, FX: e.spell.ClientEffect, Amount: dmg}, x, y)
			z.hurt(c, e.caster, dmg)
		}
	}
	list = effectsOf(c) // a death clears the list
	inCombat := false
	if p, ok := c.(*Player); ok {
		inCombat = z.now < p.combatUntil
	}
	immune := false
	kept := (*list)[:0]
	for _, e := range *list {
		if z.now+1e-9 >= e.ends || e.kind == effSpeed && !rules.BoostActive(inCombat) {
			changed = true
			if e.kind == effNearsight {
				immune = true
			}
			continue
		}
		kept = append(kept, e)
	}
	*list = kept
	if immune {
		z.nsImmuneAfter(c)
	}
	if changed {
		z.effectsChanged(c)
	}
}

// effectsChanged tells a player what their effects now are.
func (z *Zone) effectsChanged(c combatant) {
	p, ok := c.(*Player)
	if !ok {
		if n, ok := c.(*NPC); ok {
			refreshNPC(n)
		}
		return
	}
	if buffed := p.hasBuff(); buffed || p.buffed {
		p.buffed = buffed
		z.W.recalc(p) // buffs raise stats and armour
	}
	m := protocol.Effects{T: protocol.TEffects, List: []protocol.EffectView{}, Speed: speedFactor(p), Held: held(p), Spr: p.sprint}
	if p.song != nil {
		m.Song = p.song.spell.ID
	}
	for _, e := range p.effects {
		v := protocol.EffectView{Name: e.spell.Name, Kind: e.kind, Icon: e.spell.Icon,
			Left: math.Max(0, e.ends-z.now), Of: e.spell.effectSeconds()}
		if e.cb != nil {
			v.Left, v.Of, v.Conc = 0, 0, true
		} else if math.IsInf(e.ends, 1) {
			v.Left, v.Of = 0, 0 // an effect with no end, such as stealth
		}
		m.List = append(m.List, v)
	}
	p.send(m)
}
