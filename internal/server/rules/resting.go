package rules

// Resting at 1.65, from the rules-resting capability. A seated character out
// of combat regains health, power and endurance faster than a standing one.

// SeatedRegenFactor is how many times the out-of-combat tick a seated
// character regains.
const SeatedRegenFactor = 2.0

// SeatedRegen is the amount a regeneration tick returns once the character's
// posture is counted: the base amount, doubled when seated out of combat.
func SeatedRegen(base float64, seated, inCombat bool) float64 {
	if seated && !inCombat {
		return base * SeatedRegenFactor
	}
	return base
}

func init() {
	register(
		Rule{ID: "resting.seated", Capability: "rules-resting", Requirement: "Seated regeneration", Status: Provisional,
			Sources:   []string{"patch-1.87", "project"},
			Scenarios: []string{"Level 20 seated", "Seated caster", "Seated endurance"}},
	)
}
