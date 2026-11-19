package rules

import "fmt"

// Concentration (rules-concentration): a caster's pool of attention that long
// lasting buffs hold while they are on.

// ConcentrationRange is how far from its caster a target may be before the
// buff starts to lapse, in units. Reference: OpenDAoC's default.
const ConcentrationRange = 5000.0

// ConcentrationGrace is how long a target may stay out of range before the
// buff ends, in seconds. Provisional: ours, so that a short run past the edge
// does not cost the buffs.
const ConcentrationGrace = 20.0

// MaxConcentration is a caster's pool from level and casting stat: 4.4 a
// level (4 times 2.2, truncated), moved by the stat as a percent (over 50 it
// adds half the difference, under 50 it takes the whole difference away), then
// halved. A class with no spells has no pool.
func MaxConcentration(c Class, level, castStat int) int {
	if !c.Casts {
		return 0
	}
	base := int(float64(max(1, level)*4) * 2.2)
	factor := castStat - 50
	if castStat > 50 {
		factor = (castStat - 50) / 2
	}
	return max(0, (base+base*factor/100)/2)
}

// ConcentrationFree is what is left of the pool.
func ConcentrationFree(maxPool, used int) int { return max(0, maxPool-used) }

// CanHold says whether a buff that costs cost concentration fits in the free
// part of the pool.
func CanHold(free, cost int) bool { return cost <= free }

// ConcentrationShort is what a caster is told when a buff does not fit.
func ConcentrationShort(cost int) string {
	return fmt.Sprintf("This spell requires %d concentration points to cast!", cost)
}

// OutOfRange says whether a target at this distance is beyond concentration
// range.
func OutOfRange(dist float64) bool { return dist > ConcentrationRange }

// Lapsed says whether a buff whose target went out of range at outSince is to
// end at now.
func Lapsed(outSince, now float64) bool { return now-outSince >= ConcentrationGrace }

func init() {
	const c = "rules-concentration"
	register(
		Rule{ID: "conc.pool", Capability: c, Requirement: "Casters have a concentration pool",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc", "grabbag"}, Scenarios: []string{"A level 20 Healer", "A Warrior"}},
		Rule{ID: "conc.hold", Capability: c, Requirement: "A concentration spell holds its cost while it is on",
			Status: Reference, Sources: []string{"dolsharp", "opendaoc"}, Scenarios: []string{"Room in the pool", "No room", "Recasting on the same target"}},
		Rule{ID: "conc.end", Capability: c, Requirement: "A concentration buff ends with range, death or cancel",
			Status: Provisional, Sources: []string{"opendaoc", "project"}, Scenarios: []string{"The target runs off", "The caster dies"}},
	)
}
