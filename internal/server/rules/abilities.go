package rules

// The Warrior's shield abilities. Abilities are used on a group member (Guard,
// Protect, Intercept) or an enemy (Engage).
const (
	AbGuard     = "Guard"
	AbProtect   = "Protect"
	AbIntercept = "Intercept"
	AbEngage    = "Engage"

	AbilityRange      = 2000.0 // most distance from the target when an ability is used
	GuardRange        = 256.0  // a guarding Warrior must stand this close to the member
	InterceptRange    = 128.0
	InterceptCooldown = 60.0 // seconds from use until Intercept can be used again
	ProtectRange      = 1000.0
	EngageBlock       = 0.95
	EngageCost        = 5   // endurance per blow blocked
	EngageQuiet       = 5.0 // seconds without being attacked that Engage needs
)

// GuardBlocks says whether a guarding Warrior at distance blocks a blow for
// the member, given their own block chance and a roll in [0,1).
func GuardBlocks(distance, blockChance, roll float64) bool {
	return distance <= GuardRange && roll < blockChance
}

// InterceptTakes says whether an intercepting Warrior at distance takes the blow.
func InterceptTakes(distance float64) bool { return distance <= InterceptRange }

// ProtectHate is the hate a creature adds for the protecting Warrior when it
// damages the protected member.
func ProtectHate(protectLevel, damage int) float64 {
	return float64(damage) * 0.1 * float64(min(max(protectLevel, 1), 3))
}

// EngageBlocks says whether an engaged enemy's blow is blocked: it comes from
// the front, the Warrior has the endurance, and the roll is under the chance.
func EngageBlocks(inFront bool, endurance float64, roll float64) bool {
	return inFront && endurance >= EngageCost && roll < EngageBlock
}

// EngageAllowed says whether Engage may be used, by the seconds since the
// Warrior was last attacked.
func EngageAllowed(sinceAttacked float64) bool { return sinceAttacked >= EngageQuiet }

func init() {
	const c = "rules-abilities"
	register(
		Rule{ID: "ability.gained", Capability: c, Requirement: "Abilities are gained from the data's tables",
			Status: Documented, Sources: []string{"data"}, Scenarios: []string{"Guard at Shields 5", "Too early"}},
		Rule{ID: "ability.target", Capability: c, Requirement: "Abilities target a group member or an enemy",
			Status: Reference, Sources: []string{"opendaoc", "forum2001"}, Scenarios: []string{"Not grouped"}},
		Rule{ID: "ability.guard", Capability: c, Requirement: "Guard blocks for a member",
			Status: Provisional, Sources: []string{"forum2001", "project"}, Scenarios: []string{"Guarded blow", "Too far"}},
		Rule{ID: "ability.intercept", Capability: c, Requirement: "Intercept takes blows for a member",
			Status: Provisional, Sources: []string{"forum2001", "project"}, Scenarios: []string{"Intercepted"}},
		Rule{ID: "ability.protect", Capability: c, Requirement: "Protect draws the creature's anger",
			Status: Provisional, Sources: []string{"forum2001", "project"}, Scenarios: []string{"Protect II"}},
		Rule{ID: "ability.engage", Capability: c, Requirement: "Engage raises block against one enemy",
			Status: Provisional, Sources: []string{"forum2001", "project"}, Scenarios: []string{"Engaged", "Already in combat"}},
	)
}
