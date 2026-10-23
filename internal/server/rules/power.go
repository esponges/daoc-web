package rules

// A caster's power: a pool that spells draw on. The pool, the regeneration and
// the way a spell's cost is read follow the community servers (reference): no
// published 1.65 account gave the numbers, only that the pool depends on the
// class's casting stat.

// MaxPower is a caster's pool from level and casting stat (acuity): five a
// level plus the stat above 50, at least 5; 0 for a class that does not cast.
func MaxPower(c Class, level, acuity int) int {
	if !c.Casts {
		return 0
	}
	level = max(1, level)
	return max(5, level*5+acuity-50)
}

// PowerRegenPeriod is the seconds between power ticks.
const PowerRegenPeriod = 3.0

// PowerRegen is the power a tick returns: five plus a level over 2.75, in or
// out of combat. A dead caster gets none, which the caller enforces.
// Provisional for combat: the source notes patch 1.87 made combat ticks twice
// as effective, so 1.65 combat regeneration was lower by an unknown amount.
func PowerRegen(level int) float64 {
	return 5 + float64(max(1, level))/2.75
}

// SpellPowerCost is the power a spell draws when it lands. A spell whose power
// is positive costs that much; one whose power is negative costs that share,
// in percent, of the pool, at least 1. The focus bonus that lowers a cost is
// not modelled.
func SpellPowerCost(power, maxPower int) int {
	switch {
	case power > 0:
		return power
	case power < 0:
		return max(1, maxPower*-power/100)
	}
	return 0
}

func init() {
	const c = "rules-power"
	register(
		Rule{ID: "power.pool", Capability: c, Requirement: "The power pool", Status: Reference, Sources: []string{"dolsharp", "opendaoc"},
			Scenarios: []string{"A non-caster", "Levelling"}},
		Rule{ID: "power.regen", Capability: c, Requirement: "Power regeneration", Status: Provisional, Sources: []string{"dolsharp", "opendaoc"},
			Scenarios: []string{"Resting", "Interrupted cast"}},
		Rule{ID: "power.cost", Capability: c, Requirement: "A spell's cost is read from its power", Status: Reference, Sources: []string{"dolsharp", "data"},
			Scenarios: []string{"A fixed cost", "A share of the pool"}},
	)
}
