package rules

import (
	"math"
	"sort"
)

// The buffs and heals that come with concentration (rules-spell-effects).
const (
	SpellConstitutionBuff  = "ConstitutionBuff"
	SpellDexterityBuff     = "DexterityBuff"
	SpellAcuityBuff        = "AcuityBuff"
	SpellCombatSpeedBuff   = "CombatSpeedBuff"
	SpellDamageShield      = "DamageShield"
	SpellEnduranceRegen    = "EnduranceRegenBuff"
	SpellPowerRegen        = "PowerRegenBuff"
	SpellCurePoison        = "CurePoison"
	SpellCureDisease       = "CureDisease"
	SpellCureMezz          = "CureMezz"
	SpellCureNearsight     = "CureNearsightCustom"
	SpellSpreadHeal        = "SpreadHeal"
	minHastenedSwing       = 1.0 // a hastened swing is never quicker than this, seconds
	effectKindMesmerize    = "mez"
	effectKindPoison       = "poison"
	effectKindDisease      = "disease"
	effectKindNearsightKnd = "nearsight"
)

// BuffedStatsFor is BuffedStats for a holder of a class: an acuity buff raises
// the stat the class casts with, and nothing for a class that does not cast.
func BuffedStatsFor(c Class, spellType string, value float64) map[Stat]int {
	v := int(value)
	switch spellType {
	case SpellConstitutionBuff:
		return map[Stat]int{CON: v}
	case SpellDexterityBuff:
		return map[Stat]int{DEX: v}
	case SpellAcuityBuff:
		if c.Casts {
			return map[Stat]int{c.Acuity: v}
		}
		return nil
	}
	return BuffedStats(spellType, value)
}

// HastenedSwing is a swing time after a haste of pct percent: shorter by that
// share, but never under a second. Provisional: no source gives a floor.
func HastenedSwing(swing float64, pct int) float64 {
	if pct <= 0 {
		return swing
	}
	return math.Max(minHastenedSwing, swing*(1-float64(pct)/100))
}

// ShieldDamage is what a damage shield does to an attacker: its damage, less
// the attacker's resist to its damage type, at least 1.
func ShieldDamage(damage, resist float64) int {
	return max(1, SpellDamage(damage, 1, resist))
}

// PowerRegenGain is the power a tick returns with a regeneration buff: the
// ordinary tick and the buff's value.
func PowerRegenGain(level, buff int) float64 { return PowerRegen(level) + float64(buff) }

// EnduranceRegenGain is the endurance a tick returns with a regeneration buff.
func EnduranceRegenGain(base float64, buff int) float64 { return base + float64(buff) }

// Cures says whether a cure spell removes an effect of a kind. Poison, disease
// and nearsight are effect kinds the control debuffs of part C bring; until
// then a cure of them finds nothing.
func Cures(spellType, kind string) bool {
	switch spellType {
	case SpellCureMezz:
		return kind == effectKindMesmerize
	case SpellCurePoison:
		return kind == effectKindPoison
	case SpellCureDisease:
		return kind == effectKindDisease
	case SpellCureNearsight:
		return kind == effectKindNearsightKnd
	}
	return false
}

// SpreadShares shares a spread heal's total out among members missing the
// given hit points, in proportion to what each is missing and never more than
// each is missing; a point left by rounding goes to the most hurt first.
func SpreadShares(missing []int, total int) []int {
	out := make([]int, len(missing))
	sum := 0
	for _, m := range missing {
		sum += max(0, m)
	}
	if sum == 0 || total <= 0 {
		return out
	}
	given := 0
	for i, m := range missing {
		out[i] = min(max(0, m), total*max(0, m)/sum)
		given += out[i]
	}
	idx := make([]int, len(missing))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return missing[idx[a]] > missing[idx[b]] })
	for left := min(total, sum) - given; left > 0; {
		moved := false
		for _, i := range idx {
			if left > 0 && out[i] < missing[i] {
				out[i]++
				left--
				moved = true
			}
		}
		if !moved {
			break
		}
	}
	return out
}

func init() {
	const c = "rules-spell-effects"
	register(
		Rule{ID: "spell.buff-cda", Capability: c, Requirement: "Constitution, dexterity and acuity buffs",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "data"}, Scenarios: []string{"Acuity on a Healer"}},
		Rule{ID: "spell.haste", Capability: c, Requirement: "Haste and the damage shield",
			Status: Provisional, Sources: []string{"data", "project"}, Scenarios: []string{"Haste", "Damage shield"}},
		Rule{ID: "spell.regen-buffs", Capability: c, Requirement: "Regeneration buffs, cures, spread heal and life transfer",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "data"}, Scenarios: []string{"Power regeneration", "Cure of a mesmerize", "Spread heal"}},
	)
}
