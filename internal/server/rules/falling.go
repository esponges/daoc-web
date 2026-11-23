package rules

// Falling at 1.65, from the rules-falling capability. A fall hurt from patch
// 1.38; the Stealth line's Safe Fall softens it. The numbers are this
// project's.

const (
	// SafeFallHeight is the height, in units, a fall may drop with no damage.
	SafeFallHeight = 480.0
	// FallLethalSpan is the height beyond the safe one that costs all health.
	FallLethalSpan = 1920.0
	// SafeFallHeightPerLevel raises the safe height for each Safe Fall level.
	SafeFallHeightPerLevel = 240.0
	// SafeFallRelief is the share of the damage each Safe Fall level takes off.
	SafeFallRelief = 0.15
	// MaxSafeFall is the highest Safe Fall level.
	MaxSafeFall = 5
)

// SafeFallLevel is the Safe Fall a Stealth specialization gives: one level for
// each ten points, at most five.
func SafeFallLevel(stealthSpec int) int {
	return max(0, min(MaxSafeFall, stealthSpec/10))
}

// FallDamage is the hit points a landing costs, from the height fallen, the
// character's maximum health and their Safe Fall level. A landing in water
// costs nothing.
func FallDamage(height float64, maxHP, safeFall int, water bool) int {
	if water {
		return 0
	}
	safeFall = max(0, min(MaxSafeFall, safeFall))
	over := height - (SafeFallHeight + SafeFallHeightPerLevel*float64(safeFall))
	if over <= 0 {
		return 0
	}
	share := min(1, over/FallLethalSpan) * (1 - SafeFallRelief*float64(safeFall))
	return int(float64(maxHP) * share)
}

func init() {
	register(
		Rule{ID: "falling.damage", Capability: "rules-falling", Requirement: "Falling damage", Status: Provisional,
			Sources:   []string{"patch-1.38", "project"},
			Scenarios: []string{"A short drop", "A long drop", "A fall can kill"}},
		Rule{ID: "falling.safe-fall", Capability: "rules-falling", Requirement: "Safe Fall", Status: Provisional,
			Sources: []string{"patch-1.38", "project"}, Scenarios: []string{"A Shadowblade with Stealth 30"}},
		Rule{ID: "falling.water", Capability: "rules-falling", Requirement: "Swimming and water landings", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Into a lake"}},
	)
}
