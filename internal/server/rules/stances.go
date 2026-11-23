package rules

import "math"

// The timed abilities of the Berserker, Savage and the casters of the Mystic
// classes, and the Shadowblade's Stealth-line abilities that need a target on
// the ground. Sources found only say that they existed in 1.65 (Charge arrived
// as a realm ability for Berserkers, Blademasters and Mercenaries in 1.53;
// QuickCast was on the bar in 1.65) and, for Berserk, that it later made a
// Berserker's blows critical; none gives amounts or timers, so every number
// here is this project's choice and provisional.
const (
	AbBerserk     = "Berserk"
	AbCharge      = "Charge"
	AbQuickCast   = "QuickCast"
	AbDistraction = "Distraction"
	AbDangerSense = "Danger Sense"

	BerserkDamagePerLevel = 0.05  // extra melee damage for each Berserk level
	BerserkDefenceLoss    = 0.25  // the share of evade, parry and block lost while Berserk lasts
	BerserkSeconds        = 30.0  // how long it lasts
	BerserkReuse          = 300.0 // seconds before it can be used again

	ChargeFactor  = 1.25  // run speed while Charge lasts
	ChargeSeconds = 15.0  // how long it lasts
	ChargeReuse   = 900.0 // seconds before it can be used again

	QuickCastWindow = 30.0  // seconds the next spell stays instant
	QuickCastReuse  = 180.0 // seconds before it can be used again

	GroundTargetRange = 2000.0 // furthest a ground target may be set from the player
	DistractionRadius = 512.0  // creatures this close to the point turn to it
	DistractionReuse  = 600.0  // seconds before Distraction can be used again
	DangerSenseRange  = 1500.0 // how near a hidden enemy must come to be sensed
	DangerSenseQuiet  = 30.0   // seconds before the same hidden enemy is announced again
)

// BerserkDamage is the melee damage bonus of Berserk at an ability level (1 to 4).
func BerserkDamage(level int) float64 {
	return BerserkDamagePerLevel * float64(min(max(level, 0), 4))
}

// StanceWait is the seconds left before an ability's reuse timer ends, 0 when
// it can be used.
func StanceWait(now, readyAt float64) float64 { return math.Max(0, readyAt-now) }

// GroundTargetOK says whether a ground target at dist from the player may be
// set.
func GroundTargetOK(dist float64) bool { return dist >= 0 && dist <= GroundTargetRange }

// DistractionTurns says whether a creature dist from the ground target turns to
// face it.
func DistractionTurns(dist float64) bool { return dist <= DistractionRadius }

func init() {
	const c = "rules-abilities"
	register(
		Rule{ID: "ability.berserk", Capability: c, Requirement: "Berserk", Status: Provisional,
			Sources: []string{"herald-patches", "project"}, Scenarios: []string{"Berserk raises damage", "Reuse"}},
		Rule{ID: "ability.charge", Capability: c, Requirement: "Charge", Status: Provisional,
			Sources: []string{"herald-patches", "project"}, Scenarios: []string{"Charge in combat"}},
		Rule{ID: "ability.quickcast", Capability: c, Requirement: "QuickCast", Status: Provisional,
			Sources: []string{"herald-patches", "project"}, Scenarios: []string{"Instant next spell"}},
		Rule{ID: "ability.distraction", Capability: c, Requirement: "Distraction", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Creatures look away"}},
		Rule{ID: "ability.danger-sense", Capability: c, Requirement: "Danger Sense", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Warning"}},
		Rule{ID: "ground.target", Capability: "ground-target", Requirement: "Setting a ground target", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Set in range", "Out of range"}},
	)
}
