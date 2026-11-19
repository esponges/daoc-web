package rules

import "math"

// Spell types the server casts, as the spell table names them.
const (
	SpellDirectDamage   = "DirectDamage"
	SpellDDNoVariance   = "DirectDamageNoVariance"
	SpellBolt           = "Bolt"
	SpellDamageOverTime = "DamageOverTime"
	SpellSpeedDecrease  = "SpeedDecrease"
	SpellMesmerize      = "Mesmerize"
	SpellStun           = "Stun"
	SpellLifedrain      = "Lifedrain"
	SpellHeal           = "Heal"
	SpellResurrect      = "Resurrect"
	SpellOffensiveProc  = "OffensiveProc"

	// Spells only players cast in this slice: buffs and damage added to blows.
	SpellStrConBuff   = "StrengthConstitutionBuff"
	SpellDexQuiBuff   = "DexterityQuicknessBuff"
	SpellStrengthBuff = "StrengthBuff"
	SpellArmourBuff   = "BaseArmorFactorBuff"
	SpellDamageAdd    = "DamageAdd"

	// Songs and chants: speed, health regeneration and the six resist buffs.
	SpellSpeedEnhancement = "SpeedEnhancement"
	SpellHealthRegenBuff  = "HealthRegenBuff"
	SpellBodyResistBuff   = "BodyResistBuff"
	SpellColdResistBuff   = "ColdResistBuff"
	SpellEnergyResistBuff = "EnergyResistBuff"
	SpellHeatResistBuff   = "HeatResistBuff"
	SpellMatterResistBuff = "MatterResistBuff"
	SpellSpiritResistBuff = "SpiritResistBuff"
)

// The magic damage types, as the spell data numbers them.
const (
	DamageBody   = 10
	DamageCold   = 11
	DamageEnergy = 12
	DamageHeat   = 13
	DamageMatter = 14
	DamageSpirit = 15
)

// ImplementedPlayerSpell says whether a player can cast a spell type: what a
// creature casts, plus the buffs and the damage add.
func ImplementedPlayerSpell(spellType string) bool {
	switch spellType {
	case SpellResurrect, SpellStrConBuff, SpellDexQuiBuff, SpellStrengthBuff, SpellArmourBuff, SpellDamageAdd,
		SpellSpeedEnhancement, SpellHealthRegenBuff, SpellBodyResistBuff, SpellColdResistBuff, SpellEnergyResistBuff,
		SpellHeatResistBuff, SpellMatterResistBuff, SpellSpiritResistBuff:
		return true
	}
	return ImplementedSpell(spellType) && spellType != SpellOffensiveProc
}

// BuffStatCap is the most spell buffs add to one stat, by level; the items'
// cap is separate. Provisional.
func BuffStatCap(level int) int { return level * 5 / 4 }

// BuffAFCap is the most armour factor buffs add, by level. Provisional: no
// source states a flat cap (see spell.buff-af).
func BuffAFCap(level int) int { return level * 2 }

// BuffedStats is the stats a buff spell raises and by how much: its value on
// each of the stats its type names.
func BuffedStats(spellType string, value float64) map[Stat]int {
	v := int(value)
	switch spellType {
	case SpellStrConBuff:
		return map[Stat]int{STR: v, CON: v}
	case SpellDexQuiBuff:
		return map[Stat]int{DEX: v, QUI: v}
	case SpellStrengthBuff:
		return map[Stat]int{STR: v}
	}
	return nil
}

// ImplementedSpell says whether the server can cast a spell type.
func ImplementedSpell(spellType string) bool {
	switch spellType {
	case SpellDirectDamage, SpellDDNoVariance, SpellBolt, SpellDamageOverTime, SpellSpeedDecrease,
		SpellMesmerize, SpellStun, SpellLifedrain, SpellHeal, SpellOffensiveProc:
		return true
	}
	return false
}

// SpellHitChance is the chance a hostile spell lands, from the caster's and
// target's levels.
func SpellHitChance(casterLevel, targetLevel int) float64 {
	pct := 85 + float64(casterLevel-targetLevel)/2
	return math.Max(50, math.Min(99, pct)) / 100
}

// SpellResistCap is the most of a spell's damage a resist takes away.
const SpellResistCap = 70.0

// ResistFactor is what is left of spell damage after a resist, in percent.
func ResistFactor(resist float64) float64 {
	return 1 - math.Max(0, math.Min(SpellResistCap, resist))/100
}

// AreaHits says whether a creature dist units from the centre of an area
// spell is inside its radius. There is no falloff: whoever is inside takes the
// spell's full roll, resisted by its own resist, at the centre and at the edge
// alike (the community servers differ; provisional).
func AreaHits(dist, radius float64) bool { return radius > 0 && dist <= radius }

// SpellVariance is the damage multiplier for a roll r in [0, 1): uniform
// between 0.75 and 1.25, or 1 for a spell of the no-variance type.
func SpellVariance(spellType string, r float64) float64 {
	if spellType == SpellDDNoVariance {
		return 1
	}
	return 0.75 + 0.5*r
}

// SpellDamage is a landed damage spell's damage: its base, varied, after the
// resist, rounded down.
func SpellDamage(base, variance, resist float64) int {
	return int(math.Floor(base * variance * ResistFactor(resist)))
}

// BoltBlocked says whether a bolt that landed is blocked: the target's block
// chance against a blow from the front, rolled with r.
func BoltBlocked(blockChance, r float64) bool { return r < blockChance }

// DoTTicks is how many times a damage-over-time spell hits and how many
// seconds apart, from its duration in seconds and frequency in tenths of a
// second.
func DoTTicks(durationS, frequencyTenths int) (ticks int, every float64) {
	if frequencyTenths <= 0 || durationS <= 0 {
		return 0, 0
	}
	every = float64(frequencyTenths) / 10
	return int(float64(durationS)/every + 1e-9), every
}

// SlowFactor is the share of normal speed a speed-decrease spell of a value
// leaves.
func SlowFactor(value float64) float64 { return math.Max(0, math.Min(1, 1-value/100)) }

// BoostFactor is the speed multiplier of a speed-enhancement spell: its value
// read as a percent of normal speed (144 runs at 1.44 times), never below 1.
func BoostFactor(value float64) float64 { return math.Max(1, value/100) }

// WithBoost applies a speed enhancement to the speed factor the slows and
// holds leave: they multiply, and a held body stays held.
func WithBoost(factor, boost float64) float64 {
	if factor <= 0 {
		return 0
	}
	return factor * math.Max(1, boost)
}

// BoostActive says whether a speed enhancement holds: it does not take hold on
// a body in combat, and ends when its holder enters combat.
func BoostActive(inCombat bool) bool { return !inCombat }

// HealthRegenBuffed is the hit points a health tick returns with a
// regeneration buff of buff on: the tick's own amount plus the buff, none in
// combat, where no tick returns anything.
func HealthRegenBuffed(level int, inCombat bool, buff int) int {
	if inCombat {
		return 0
	}
	return HealthRegen(level, false) + max(0, buff)
}

// BuffResistCap is the most spell buffs add to one resist, in percent.
const BuffResistCap = 24.0

// ResistBuff is one resist buff's damage type and strength in percent.
type ResistBuff struct {
	DamageType int
	Value      float64
}

// ResistBuffType is the damage type a resist-buff spell type raises.
func ResistBuffType(spellType string) (int, bool) {
	switch spellType {
	case SpellBodyResistBuff:
		return DamageBody, true
	case SpellColdResistBuff:
		return DamageCold, true
	case SpellEnergyResistBuff:
		return DamageEnergy, true
	case SpellHeatResistBuff:
		return DamageHeat, true
	case SpellMatterResistBuff:
		return DamageMatter, true
	case SpellSpiritResistBuff:
		return DamageSpirit, true
	}
	return 0, false
}

// ResistBuffBonus is what the buffs add to the resist to one damage type: the
// strongest of that type, at most BuffResistCap. The caller adds it to the
// race's and items' resist; the total is capped where damage is taken.
func ResistBuffBonus(damageType int, buffs []ResistBuff) float64 {
	best := 0.0
	for _, b := range buffs {
		if b.DamageType == damageType {
			best = math.Max(best, b.Value)
		}
	}
	return math.Min(best, BuffResistCap)
}

// SpeedFactor is a body's speed multiplier from what holds it: 0 while
// stunned or mesmerized, else the strongest slow, else 1.
func SpeedFactor(held bool, slows ...float64) float64 {
	if held {
		return 0
	}
	f := 1.0
	for _, s := range slows {
		f = math.Min(f, s)
	}
	return f
}

// BrokenByDamage says whether damage ends an effect of a spell type: a
// mesmerize, not a stun.
func BrokenByDamage(spellType string) bool { return spellType == SpellMesmerize }

// DrainHeal is what a life drain returns to its caster: the damage done
// times the spell's return, a percent.
func DrainHeal(damage, lifeDrainReturn int) int { return damage * lifeDrainReturn / 100 }

// PercentHeal is what a heal whose value is negative restores: that percent of
// the caster's most health, as the community servers read it.
func PercentHeal(value float64, casterMax int) int {
	return int(-value / 100 * float64(casterMax))
}

// Healed is the health after a heal of value, up to the maximum.
func Healed(health, maxHealth, value int) int { return min(maxHealth, health+max(0, value)) }

func init() {
	const c = "rules-spell-effects"
	register(
		Rule{ID: "spell.buff", Capability: "player-spellcasting", Requirement: "Spell types of the first slice",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "data"}, Scenarios: []string{"A buff is capped by level", "A buff does not scale with the caster"}},
		Rule{ID: "spell.buff-af", Capability: "player-spellcasting", Requirement: "The armour-factor buff cap",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"A big armour buff"}},
		Rule{ID: "spell.hit", Capability: c, Requirement: "A hostile spell lands by a hit roll",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Even levels", "A higher caster", "Far below"}},
		Rule{ID: "spell.resist", Capability: c, Requirement: "Resists reduce spell damage",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A resisted fraction"}},
		Rule{ID: "spell.damage", Capability: c, Requirement: "Direct damage and bolts vary",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "grabbag"}, Scenarios: []string{"Middle of the range", "No variance", "A fully specced caster", "A low spec", "The cap"}},
		Rule{ID: "spell.area", Capability: c, Requirement: "Area damage does not fall off",
			Status: Provisional, Sources: []string{"dolsharp", "opendaoc"}, Scenarios: []string{"Centre and edge"}},
		Rule{ID: "spell.speed-enhancement", Capability: c, Requirement: "Speed enhancement raises speed",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "allakhazam"}, Scenarios: []string{"Simple Song of Travel", "The strongest counts", "With a slow", "Combat"}},
		Rule{ID: "spell.health-regen", Capability: c, Requirement: "A health regeneration buff adds to health ticks",
			Status: Provisional, Sources: []string{"dolsharp", "opendaoc", "project"}, Scenarios: []string{"Song of Rest out of combat", "In combat"}},
		Rule{ID: "spell.resist-buff", Capability: c, Requirement: "Resist buffs raise a resist",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "grabbag"}, Scenarios: []string{"A heat chant against a fire bolt", "Another damage type is untouched", "The buff cap"}},
		Rule{ID: "spell.bolt-block", Capability: c, Requirement: "A bolt can be blocked",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Blocked bolt"}},
		Rule{ID: "spell.damage-over-time", Capability: c, Requirement: "Damage over time ticks",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "data"}, Scenarios: []string{"Five ticks", "A player's tick at spec", "A low spec tick", "The tick cap"}},
		Rule{ID: "spell.damage-add", Capability: c, Requirement: "A damage add is a damage per second",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "data"}, Scenarios: []string{"A slow weapon", "The damage add cap", "A percent add"}},
		Rule{ID: "spell.slow", Capability: c, Requirement: "Slows lower speed",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A 40% slow"}},
		Rule{ID: "spell.mesmerize", Capability: c, Requirement: "Mesmerize holds until hurt",
			Status: Reference, Sources: []string{"opendaoc"}},
		Rule{ID: "spell.stun", Capability: c, Requirement: "Stun holds through damage",
			Status: Reference, Sources: []string{"opendaoc"}},
		Rule{ID: "spell.lifedrain", Capability: c, Requirement: "Life drain heals its caster",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"Drain"}},
		Rule{ID: "spell.heal", Capability: c, Requirement: "Heals restore health",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "grabbag"}, Scenarios: []string{"A creature heals itself", "A player heal at spec", "A player heal below spec", "A percent heal"}},
	)
}

// CasterStatFactor is what a caster's casting stat does to a spell's damage:
// 1 at a stat of 75, rising a little over a third of a percent a point.
func CasterStatFactor(stat int) float64 { return float64(stat+200) / 275 }

// PlayerVarianceRange is the damage multiplier range of a player's direct
// damage spell. The low end rises with the caster's spec in the spell's line
// against the target's level, the high end is 1.25; a spec above the target's
// level and a caster above the target's level lift both ends.
func PlayerVarianceRange(specLevel, casterLevel, targetLevel int) (lo, hi float64) {
	lo, hi = 1.25, 1.25
	if targetLevel > 0 {
		lo = 0.25 + float64(specLevel-1)/float64(targetLevel)
	}
	if specLevel-1 > targetLevel {
		over := float64(specLevel-1-targetLevel) * 0.005
		lo, hi = lo+over, hi+over
	}
	diff := 0.02 * float64(casterLevel-targetLevel)
	lo, hi = lo+diff, hi+diff
	hi = math.Max(0.25, hi)
	lo = math.Max(0, math.Min(lo, hi))
	return lo, hi
}

// PlayerSpellCapFactor is the most a damage spell does, in times its base damage.
const PlayerSpellCapFactor = 3.0

// PlayerSpellDamage is a landed direct damage spell's damage from a player:
// the base scaled by the casting stat, a roll r in [0, 1) within the variance
// range, the resist, and the cap, rounded down.
func PlayerSpellDamage(base float64, stat, specLevel, casterLevel, targetLevel int, r, resist float64) int {
	lo, hi := PlayerVarianceRange(specLevel, casterLevel, targetLevel)
	scaled := base * CasterStatFactor(stat)
	dmg := (lo + (hi-lo)*r) * scaled * ResistFactor(resist)
	return int(math.Floor(math.Min(dmg, base*PlayerSpellCapFactor)))
}

// PlayerHealRange is the least and most a player's heal of value restores. The
// most is 1.25 times the value. The least rises with the caster's spec in the
// spell's line toward the spell's own level, and is at most the most.
func PlayerHealRange(value float64, specLevel, spellLevel int) (lo, hi int) {
	hi = max(1, int(value*1.25))
	eff := 0.25
	if spellLevel > 0 {
		eff = math.Min(1.25, eff+float64(max(1, specLevel)-1)/float64(spellLevel))
	}
	lo = min(hi, max(1, int(value*eff)))
	return lo, hi
}

// PlayerHeal is what a player's heal restores for a roll r in [0, 1).
func PlayerHeal(value float64, specLevel, spellLevel int, r float64) int {
	lo, hi := PlayerHealRange(value, specLevel, spellLevel)
	return lo + int(r*float64(hi-lo+1))
}

// PlayerDoTRange is the damage multiplier range of one tick of a player's
// damage-over-time spell: the low end rises with the caster's spec in the
// spell's line against the target's level, the high end is 1.13, and the
// level difference between them does not count.
func PlayerDoTRange(specLevel, targetLevel int) (lo, hi float64) {
	hi = 1.13
	lo = 0.13
	if targetLevel > 0 {
		lo += float64(max(1, specLevel)-1) / float64(targetLevel)
	}
	return math.Max(0, math.Min(lo, hi)), hi
}

// PlayerDoTDamage is one tick of a player's damage-over-time spell for a roll
// r in [0, 1): the base scaled by the casting stat and the range, after the
// resist, at most three times the base, rounded down.
func PlayerDoTDamage(base float64, stat, specLevel, targetLevel int, r, resist float64) int {
	lo, hi := PlayerDoTRange(specLevel, targetLevel)
	dmg := (lo + (hi-lo)*r) * base * CasterStatFactor(stat) * ResistFactor(resist)
	return int(math.Floor(math.Min(dmg, base*PlayerSpellCapFactor)))
}

// DamageAddCap is the most damage per second a damage add counts for, by the
// caster's level.
func DamageAddCap(level int) float64 { return (1.2 + 0.3*float64(level)) * 0.7 }

// DamageAdd is what a damage-add spell puts on one landed blow. A spell damage
// above zero is a damage per second: the lesser of it and the cap, times the
// swing in seconds, times a roll r in [0, 1) mapped to 0.5 to 1. Below zero it
// is that percent of the blow's damage. The resist comes off either way.
func DamageAdd(spellDamage float64, casterLevel int, swingSeconds float64, blow int, r, resist float64) int {
	var dmg float64
	if spellDamage < 0 {
		dmg = float64(blow) * -spellDamage / 100
	} else {
		dmg = math.Min(spellDamage, DamageAddCap(casterLevel)) * swingSeconds * (0.5 + 0.5*r)
	}
	return int(math.Floor(dmg * ResistFactor(resist)))
}
