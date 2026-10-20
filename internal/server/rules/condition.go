package rules

import "math"

// Item condition is kept in hundredths of a percent: 10000 is whole.
const (
	FullCondition = 10000
	// WearPerBlow is the condition lost to one blow: 0.05 points.
	WearPerBlow = 5
)

// ConditionStart is the condition a new item begins at: whole, or the item
// template's own (0..1 of its maximum) when it gives one below whole.
func ConditionStart(template float64) int {
	if template <= 0 || template >= 1 {
		return FullCondition
	}
	return int(math.Round(template * FullCondition))
}

// Worn is a condition after one more blow, never below 0.
func Worn(cond int) int { return max(0, cond-WearPerBlow) }

// BlowWears says what a blow's result wears on its target: the armour at the
// hit location when it lands, the shield when it is blocked, nothing otherwise.
func BlowWears(r Result) (armour, shield bool) {
	switch r {
	case Hit:
		return true, false
	case Blocked:
		return false, true
	}
	return false, false
}

// Effectiveness is the share of a weapon's damage per second, or an armour
// piece's armour factor, a condition leaves.
func Effectiveness(cond int) float64 {
	return math.Min(1, math.Max(0, float64(cond)/FullCondition))
}

// WarnThresholds are the conditions, in hundredths of a percent, a character
// is told about as an item falls below each.
var WarnThresholds = []int{9000, 8000, 7000}

// WarnCrossed returns the threshold an item has just fallen below, or 0.
func WarnCrossed(before, after int) int {
	for _, t := range WarnThresholds {
		if before >= t && after < t {
			return t
		}
	}
	return 0
}

// RepairPrice is what a blacksmith charges, in copper: the item's price times
// the condition missing (in percent) over 100, halved, rounded up to a copper.
// An item that is whole costs nothing.
func RepairPrice(price int64, cond int) int64 {
	if cond >= FullCondition || price <= 0 {
		return 0
	}
	missing := float64(FullCondition-cond) / 100 // percent
	return int64(math.Ceil(float64(price) * missing / 100 * 0.5))
}

func init() {
	const c = "rules-item-condition"
	register(
		Rule{ID: "condition.start", Capability: c, Requirement: "Items have a condition",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A new item"}},
		Rule{ID: "condition.wear", Capability: c, Requirement: "Combat wears items",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"Two hundred blows", "A block"}},
		Rule{ID: "condition.effectiveness", Capability: c, Requirement: "Wear lowers effectiveness",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A weapon at 90%"}},
		Rule{ID: "condition.warning", Capability: c, Requirement: "Players are warned of wear",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"The 90% warning"}},
		Rule{ID: "condition.repair-price", Capability: c, Requirement: "Blacksmiths repair for money",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"Repair price", "Not enough money"}},
	)
}
