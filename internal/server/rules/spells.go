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
	SpellOffensiveProc  = "OffensiveProc"
)

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

// Healed is the health after a heal of value, up to the maximum.
func Healed(health, maxHealth, value int) int { return min(maxHealth, health+max(0, value)) }

func init() {
	const c = "rules-spell-effects"
	register(
		Rule{ID: "spell.hit", Capability: c, Requirement: "A hostile spell lands by a hit roll",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Even levels", "A higher caster", "Far below"}},
		Rule{ID: "spell.resist", Capability: c, Requirement: "Resists reduce spell damage",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A resisted fraction"}},
		Rule{ID: "spell.damage", Capability: c, Requirement: "Direct damage and bolts vary",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"Middle of the range", "No variance"}},
		Rule{ID: "spell.bolt-block", Capability: c, Requirement: "A bolt can be blocked",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Blocked bolt"}},
		Rule{ID: "spell.damage-over-time", Capability: c, Requirement: "Damage over time ticks",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"Five ticks"}},
		Rule{ID: "spell.slow", Capability: c, Requirement: "Slows lower speed",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A 40% slow"}},
		Rule{ID: "spell.mesmerize", Capability: c, Requirement: "Mesmerize holds until hurt",
			Status: Reference, Sources: []string{"opendaoc"}},
		Rule{ID: "spell.stun", Capability: c, Requirement: "Stun holds through damage",
			Status: Reference, Sources: []string{"opendaoc"}},
		Rule{ID: "spell.lifedrain", Capability: c, Requirement: "Life drain heals its caster",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"Drain"}},
		Rule{ID: "spell.heal", Capability: c, Requirement: "Heals restore health",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A creature heals itself"}},
	)
}
