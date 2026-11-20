package rules

// The left hand, from the rules-melee capability. A player trained in Left Axe
// swings the left hand on every round, for a share of damage; a Hand to Hand
// fighter gets extra blows by chance. OpenDAoC-Core's numbers (provisional).

const ObjHandToHand = 25 // the object type of claws and other hand to hand weapons

// LeftAxeDamage is the share of a blow a left-hand swing does for a Left Axe spec.
func LeftAxeDamage(spec int) float64 { return 0.625 + 0.0034*float64(spec) }

// LeftSwings is how many left-hand swings a round has for a Left Axe fighter: one.
func LeftSwings(spec int) int {
	if spec > 0 {
		return 1
	}
	return 0
}

// HandToHandChances are the chances of two, three and four blows in a round for
// a Hand to Hand spec: a half of the spec in percent for the second, half that
// for the third, a quarter of that for the fourth.
func HandToHandChances(spec int) (double, triple, quad float64) {
	if spec <= 0 {
		return 0, 0, 0
	}
	double = float64(spec) * 0.01 * 0.5
	triple = double * 0.5
	quad = triple * 0.25
	return
}

// HandToHandExtra is the extra blows a round has for a roll in 0..1: 1, 2 or 3.
func HandToHandExtra(spec int, roll float64) int {
	d, t, q := HandToHandChances(spec)
	switch {
	case roll < d:
		return 1
	case roll < d+t:
		return 2
	case roll < d+t+q:
		return 3
	}
	return 0
}

func init() {
	const c = "rules-melee"
	register(
		Rule{ID: "melee.left-hand", Capability: c, Requirement: "A left-hand weapon swings alongside the main hand",
			Status: Provisional, Sources: []string{"opendaoc"}, Scenarios: []string{"A Shadowblade with two axes", "Hand to Hand blows"}},
	)
}
