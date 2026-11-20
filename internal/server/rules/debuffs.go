package rules

import "math"

// Debuffs, control effects and wards (rules-spell-effects, midgard-to-fifty
// part C). Numbers from findings.md, "Part C".
const (
	SpellStrengthDebuff     = "StrengthDebuff"
	SpellDexterityDebuff    = "DexterityDebuff"
	SpellStrConDebuff       = "StrengthConstitutionDebuff"
	SpellDexQuiDebuff       = "DexterityQuicknessDebuff"
	SpellCombatSpeedDebuff  = "CombatSpeedDebuff"
	SpellBodyResistDebuff   = "BodyResistDebuff"
	SpellColdResistDebuff   = "ColdResistDebuff"
	SpellEnergyResistDebuff = "EnergyResistDebuff"
	SpellHeatResistDebuff   = "HeatResistDebuff"
	SpellMatterResistDebuff = "MatterResistDebuff"
	SpellSpiritResistDebuff = "SpiritResistDebuff"
	SpellDisease            = "Disease"
	SpellNearsight          = "Nearsight"
	SpellConfusion          = "Confusion"
	SpellAmnesia            = "Amnesia"
	SpellBladeturn          = "Bladeturn"
	SpellAblativeArmor      = "AblativeArmor"
	SpellArmorAbsorbBuff    = "ArmorAbsorptionBuff"
	SpellLifeTransfer       = "LifeTransfer"
	SpellDDWithDebuff       = "DirectDamageWithDebuff"
	SpellDamageSpeedDecr    = "DamageSpeedDecrease"
	SpellUnbreakableSlow    = "UnbreakableSpeedDecrease"
	SpellFatigueDebuff      = "FatigueConsumptionDebuff"
	SpellMezDurationBuff    = "MesmerizeDurationBuff"
)

// DebuffedStats is the stats a stat debuff lowers and by how much (positive).
func DebuffedStats(spellType string, value float64) map[Stat]int {
	v := int(math.Abs(value))
	switch spellType {
	case SpellStrengthDebuff:
		return map[Stat]int{STR: v}
	case SpellDexterityDebuff:
		return map[Stat]int{DEX: v}
	case SpellStrConDebuff:
		return map[Stat]int{STR: v, CON: v}
	case SpellDexQuiDebuff:
		return map[Stat]int{DEX: v, QUI: v}
	}
	return nil
}

// ResistDebuffType is the damage type a resist-debuff spell type lowers.
func ResistDebuffType(spellType string) (int, bool) {
	switch spellType {
	case SpellBodyResistDebuff:
		return DamageBody, true
	case SpellColdResistDebuff:
		return DamageCold, true
	case SpellEnergyResistDebuff:
		return DamageEnergy, true
	case SpellHeatResistDebuff:
		return DamageHeat, true
	case SpellMatterResistDebuff:
		return DamageMatter, true
	case SpellSpiritResistDebuff:
		return DamageSpirit, true
	}
	return 0, false
}

// DebuffDuration is how long a debuff lasts, in seconds: the spell's, lengthened
// by the caster's duration bonus (percent), shortened by the target's resist to
// its damage type (percent), at least a second and at most four times the
// spell's.
func DebuffDuration(base, casterBonus, resist float64) float64 {
	d := base * (1 + casterBonus/100)
	d -= d * resist / 100
	return math.Max(1, math.Min(d, base*4))
}

// SlowedSwing is a swing time under a combat-speed debuff of pct percent: the
// attack speed falls by that share, so the swing takes longer. Provisional: the
// sources say "attack speed" without giving the form; at most 90 percent.
func SlowedSwing(swing float64, pct int) float64 {
	if pct <= 0 {
		return swing
	}
	return swing / (1 - math.Min(float64(pct), 90)/100)
}

// Disease: slows its holder, lowers their strength and halves the healing they
// receive; they heal for half from life transfer too.
const (
	DiseaseSlowPct    = 15.0
	DiseaseStrengthBy = 0.075
	DiseaseHealShare  = 0.5
)

// DiseasedHeal is the healing a diseased body receives.
func DiseasedHeal(amount int) int { return int(float64(amount) * DiseaseHealShare) }

// DiseasedStrength is a strength under disease.
func DiseasedStrength(str int) int { return int(float64(str) * (1 - DiseaseStrengthBy)) }

// NearsightRange is a spell or bow range under nearsight of pct percent.
func NearsightRange(rng, pct float64) float64 { return rng * (1 - math.Min(pct, 100)/100) }

// NearsightImmunity is the seconds after nearsight ends in which another does
// not take hold. Provisional: no source gives it.
const NearsightImmunity = 10.0

// Confusion on a creature: every ConfusionPulse seconds it drops its target
// and picks a random enemy within ConfusionRadius.
const (
	ConfusionPulse  = 6.0
	ConfusionRadius = 750.0
)

// ConfusionTakesHold says whether a confusion with value takes hold on a roll r
// in [0, 1): a value of 0 or more is its percent chance; a negative value always.
func ConfusionTakesHold(value, r float64) bool { return value < 0 || r*100 < value }

// ConfusionAttacksAlly says whether a confused creature's pulse may pick one of
// its own side: only for a negative value, with that percent chance.
func ConfusionAttacksAlly(value, r float64) bool { return value < 0 && r*100 < -value }

// AmnesiaChance is the percent chance an amnesia makes a creature forget its
// enemies. The data's spell rows carry no chance, so it always does.
// Provisional.
const AmnesiaChance = 100.0

// LifeTransfer is what a life transfer of value (percent of the caster's maximum
// health) heals and what the caster pays: twice the value in percent of their
// maximum is healed, the caster pays half of that, never going under 1 hit
// point, and a diseased target is healed half of it.
func LifeTransfer(casterMax, casterHP int, value float64, diseased bool) (heal, cost int) {
	heal = int(float64(casterMax) / 100 * math.Abs(value) * 2)
	if heal <= 0 {
		return 0, 0
	}
	if heal/2 >= casterHP {
		heal = (casterHP - 1) * 2
	}
	cost = heal / 2
	if diseased {
		heal /= 2
	}
	return max(0, heal), max(0, cost)
}

// AblativePercent is the share of each hit an ablative absorbs: its damage
// column, 25 when it has none, at most 100.
func AblativePercent(damage float64) float64 {
	if damage <= 0 {
		return 25
	}
	return math.Min(100, damage)
}

// AblativeAbsorb is what an ablative of pct percent and pool points takes from a
// hit, and what is left of the pool.
func AblativeAbsorb(hit int, pct float64, pool int) (absorbed, left int) {
	absorbed = min(pool, int(float64(hit)*pct/100))
	return absorbed, pool - absorbed
}

// FatigueFactor is what an endurance cost is multiplied by under a fatigue
// consumption debuff of value (the data stores it negative).
func FatigueFactor(value float64) float64 { return 1 + math.Abs(value)/100 }

// MezDuration is a mesmerize's length for a caster with a mesmerize duration buff
// of pct percent. Provisional.
func MezDuration(base, pct float64) float64 { return base * (1 + math.Max(0, pct)/100) }

// StrongestOf is the largest of the values: debuffs of one kind do not stack.
func StrongestOf(values ...float64) float64 {
	best := 0.0
	for _, v := range values {
		best = math.Max(best, v)
	}
	return best
}

func init() {
	const c = "rules-spell-effects"
	register(
		Rule{ID: "spell.debuffs", Capability: c, Requirement: "Stat and resist debuffs lower what they name",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"Strength debuff on a creature", "Body resist debuff"}},
		Rule{ID: "spell.control-debuffs", Capability: c, Requirement: "Disease, nearsight, confusion and amnesia",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"Disease", "Amnesia"}},
		Rule{ID: "spell.wards", Capability: c, Requirement: "Bladeturn, ablative and absorption",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"Bladeturn", "Ablative pool"}},
		Rule{ID: "spell.life-transfer", Capability: c, Requirement: "Regeneration buffs, cures, spread heal and life transfer",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"Life transfer"}},
	)
}
