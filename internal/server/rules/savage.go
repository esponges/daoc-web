package rules

// The Savage's own spells (the Savagery line) and the Taunt spell. The buffs
// are cast on the Savage alone and are paid for in hit points, not power: the
// community data gives each a negative power value, and the 1.65H test notes
// say Savage buffs "subtract a percent of your maximum health". Their stats are
// the data's numbers read as percents; that reading is this project's.
const (
	SpellSavageCrush   = "SavageCrushResistanceBuff"
	SpellSavageSlash   = "SavageSlashResistanceBuff"
	SpellSavageThrust  = "SavageThrustResistanceBuff"
	SpellSavageDPS     = "SavageDPSBuff"
	SpellSavageEvade   = "SavageEvadeBuff"
	SpellSavageParry   = "SavageParryBuff"
	SpellSavageSpeed   = "SavageCombatSpeedBuff"
	SpellSavageEndHeal = "SavageEnduranceHeal"
	SpellTaunt         = "Taunt"
	TauntHatePerPoint  = 10.0 // hate a Taunt adds for each point of its value. Provisional.
)

// PaysHealth says whether a spell type is paid for in hit points.
func PaysHealth(spellType string) bool {
	switch spellType {
	case SpellSavageCrush, SpellSavageSlash, SpellSavageThrust, SpellSavageDPS, SpellSavageEvade, SpellSavageParry,
		SpellSavageSpeed, SpellSavageEndHeal:
		return true
	}
	return false
}

// HealthCost is the hit points a Savage spell takes: the spell's power, read as
// the share of maximum health when it is negative, at least 1. A spell with no
// power costs nothing.
func HealthCost(power, maxHP int) int {
	if power >= 0 {
		return 0
	}
	return max(1, maxHP*-power/100)
}

// CanPayHealth says whether a caster at hp can pay a health cost: the cost may
// not take them to nothing.
func CanPayHealth(hp float64, cost int) bool { return hp > float64(cost) }

// SavageResistType is the melee damage type a Savage resistance buff guards
// against.
func SavageResistType(spellType string) (int, bool) {
	switch spellType {
	case SpellSavageCrush:
		return DamageCrush, true
	case SpellSavageSlash:
		return DamageSlash, true
	case SpellSavageThrust:
		return DamageThrus, true
	}
	return 0, false
}

// SavagePercent turns a Savage buff's value into a fraction: the damage bonus
// of the DPS buff, or the evade and parry chance the others add.
func SavagePercent(value float64) float64 { return value / 100 }

// EnduranceHealed is the endurance a Savage endurance heal returns: its value
// as a percent of the full pool.
func EnduranceHealed(value float64, maxEnd int) float64 { return value / 100 * float64(maxEnd) }

// TauntHate is the hate a Taunt adds to the creature's list for the caster.
func TauntHate(value float64) float64 { return value * TauntHatePerPoint }

func init() {
	const c = "rules-spell-effects"
	register(
		Rule{ID: "savage.health-cost", Capability: c, Requirement: "Savage self-buffs",
			Status: Documented, Sources: []string{"data", "herald-patches"}, Scenarios: []string{"Health cost", "Not enough health"}},
		Rule{ID: "savage.buffs", Capability: c, Requirement: "Savage self-buffs",
			Status: Provisional, Sources: []string{"data", "herald-savage", "project"}, Scenarios: []string{"Combat speed"}},
		Rule{ID: "savage.endurance-heal", Capability: c, Requirement: "Savage endurance heal",
			Status: Provisional, Sources: []string{"data", "herald-savage", "project"}, Scenarios: []string{"Endurance restored"}},
		Rule{ID: "savage.taunt", Capability: c, Requirement: "Taunt",
			Status: Provisional, Sources: []string{"data", "project"}, Scenarios: []string{"Pulling hate"}},
	)
}
