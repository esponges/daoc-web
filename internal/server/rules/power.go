package rules

// A caster's power: a pool that spells draw on. No public source gave 1.65's
// pool or regeneration numbers, so the pool is this project's choice from the
// one fact found (it grows with the class's casting stat, piety for the
// Midgard casters we have), and the rule is provisional.

// PowerPoolBase is what every caster has before level and stat.
const PowerPoolBase = 10

// MaxPower is a caster's pool from level and casting stat (acuity); 0 for a
// class that does not cast.
func MaxPower(c Class, level, acuity int) int {
	if !c.Casts {
		return 0
	}
	level = max(1, level)
	return PowerPoolBase + level*(5+acuity/50)
}

// PowerRegenPeriod is the seconds between power ticks.
const PowerRegenPeriod = 3.0

// PowerRegen is the power a tick returns: a fiftieth of the pool out of
// combat, a two-hundredth in it, at least 1 either way. A dead caster gets
// none, which the caller enforces.
func PowerRegen(maxPower int, inCombat bool) int {
	if maxPower <= 0 {
		return 0
	}
	if inCombat {
		return max(1, maxPower/200)
	}
	return max(1, maxPower/50)
}

// SpellPowerCost is the power a spell draws when it lands: its power value is
// a percentage of the pool, and a spell that names a cost costs at least 1.
func SpellPowerCost(power, maxPower int) int {
	if power <= 0 {
		return 0
	}
	return max(1, maxPower*power/100)
}

func init() {
	const c = "rules-power"
	register(
		Rule{ID: "power.pool", Capability: c, Requirement: "The power pool", Status: Provisional, Sources: []string{"project"},
			Scenarios: []string{"A non-caster", "Levelling"}},
		Rule{ID: "power.regen", Capability: c, Requirement: "Power regeneration", Status: Provisional, Sources: []string{"project"},
			Scenarios: []string{"Resting", "Interrupted cast"}},
	)
}
