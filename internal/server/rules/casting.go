package rules

// HealBelow is the share of health under which a creature heals.
const HealBelow = 0.75

// CastRetry is the seconds before a failed or interrupted spell is ready again.
const CastRetry = 2.0

// CastOption is one of a creature's spells as its combat decision sees it.
type CastOption struct {
	Heal    bool    // a heal, for itself or a hurt ally
	Ready   bool    // its recast delay has passed
	Range   float64 // how far it reaches
	Already bool    // a lasting effect of it is already on the target
}

// ChooseCast is a creature's combat decision: the spell to cast, or -1 to
// swing or close in. A ready heal comes first when someone it would heal is
// below HealBelow; then the first ready hostile spell that reaches the
// target and is not already on it.
func ChooseCast(spells []CastOption, targetDist float64, someoneHurt bool) int {
	if someoneHurt {
		for i, s := range spells {
			if s.Heal && s.Ready {
				return i
			}
		}
	}
	for i, s := range spells {
		if !s.Heal && s.Ready && !s.Already && targetDist <= s.Range {
			return i
		}
	}
	return -1
}

// Hurt says whether a creature at health of max needs a heal.
func Hurt(health, max int) bool { return max > 0 && float64(health) < HealBelow*float64(max) }

// CastEnds is when a cast begun at now takes effect; an instant spell (cast
// time 0) takes effect at once.
func CastEnds(now, castTime float64) float64 { return now + max(0, castTime) }

// CastLands says whether a cast that has run its time takes effect: the
// target must still be fightable and within the spell's range.
func CastLands(targetAlive bool, targetDist, spellRange float64) bool {
	return targetAlive && targetDist <= spellRange
}

// Interrupts says whether a melee attack on a casting creature stops its
// spell: any attack, hit or not, unless the spell is uninterruptible.
func Interrupts(uninterruptible bool) bool { return !uninterruptible }

func init() {
	const c = "rules-creature-casting"
	register(
		Rule{ID: "cast.decision", Capability: c, Requirement: "A creature in combat casts when a spell is ready",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"Cast, then melee", "Out of range", "A slow already on the target"}},
		Rule{ID: "cast.heal", Capability: c, Requirement: "Creatures heal when hurt",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"Healing a packmate"}},
		Rule{ID: "cast.time", Capability: c, Requirement: "Casting takes the spell's cast time",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"A 3-second cast", "The target ran away"}},
		Rule{ID: "cast.interrupt", Capability: c, Requirement: "Being attacked interrupts casting",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Interrupt", "Uninterruptible"}},
	)
}
