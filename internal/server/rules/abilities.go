package rules

// The Warrior's shield abilities. Abilities are used on a group member (Guard,
// Protect, Intercept) or an enemy (Engage). The numbers follow OpenDAoC's, read
// to understand them (nothing copied); the forum thread of 2001 agrees in
// outline.
const (
	AbGuard     = "Guard"
	AbProtect   = "Protect"
	AbIntercept = "Intercept"
	AbEngage    = "Engage"

	AbilityRange      = 2000.0 // most distance from the target when an ability is used
	GuardRange        = 256.0  // a guarding Warrior must stand this close to the member
	GuardPerLevel     = 0.05   // block chance a guard gains with each Guard level
	InterceptRange    = 128.0
	InterceptChance   = 0.5  // chance an armed Intercept takes a blow within range
	InterceptCooldown = 60.0 // seconds from use until Intercept can be used again
	ProtectRange      = 1000.0
	ProtectPerLevel   = 0.1 // share of the member's hate moved to the protector, per Protect level
	EngageBonus       = 0.65
	EngageCost        = 5   // endurance each time Engage raises a block chance
	EngageQuiet       = 5.0 // seconds an engaged creature must have gone unattacked by an enemy
)

// GuardChance is a guarding Warrior's chance to block a blow for the member:
// their own block chance against it (from their front) and 5% a Guard level.
func GuardChance(blockChance float64, guardLevel int) float64 {
	return blockChance + GuardPerLevel*float64(guardLevel)
}

// GuardBlocks says whether a guarding Warrior at distance blocks a blow for
// the member, given their guard chance and a roll in [0,1).
func GuardBlocks(distance, chance, roll float64) bool {
	return distance <= GuardRange && roll < chance
}

// InterceptTakes says whether an armed Intercept at distance takes the blow,
// by a roll in [0,1).
func InterceptTakes(distance, roll float64) bool {
	return distance <= InterceptRange && roll < InterceptChance
}

// ProtectShift is the hate that moves from a member to their protector when a
// creature's hate for the member rises by hate.
func ProtectShift(protectLevel int, hate float64) float64 {
	return hate * ProtectPerLevel * float64(min(max(protectLevel, 1), 3))
}

// EngageAllowed says whether an engaged creature may be (or stay) engaged, by
// the seconds since an enemy last attacked it.
func EngageAllowed(sinceAttacked float64) bool { return sinceAttacked >= EngageQuiet }

func init() {
	const c = "rules-abilities"
	register(
		Rule{ID: "ability.gained", Capability: c, Requirement: "Abilities are gained from the data's tables",
			Status: Documented, Sources: []string{"data"}, Scenarios: []string{"Guard at Shields 5", "Too early"}},
		Rule{ID: "ability.target", Capability: c, Requirement: "Abilities target a group member or an enemy",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Not grouped"}},
		Rule{ID: "ability.guard", Capability: c, Requirement: "Guard blocks for a member",
			Status: Reference, Sources: []string{"opendaoc", "forum2001"}, Scenarios: []string{"Guarded blow", "Too far"}},
		Rule{ID: "ability.intercept", Capability: c, Requirement: "Intercept takes blows for a member",
			Status: Reference, Sources: []string{"opendaoc", "forum2001"}, Scenarios: []string{"Intercepted", "Intercept fails its roll"}},
		Rule{ID: "ability.protect", Capability: c, Requirement: "Protect moves the creature's anger",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Protect II"}},
		Rule{ID: "ability.engage", Capability: c, Requirement: "Engage raises block against one enemy",
			Status: Reference, Sources: []string{"opendaoc", "forum2001"}, Scenarios: []string{"Engaged", "Already in combat"}},
	)
}
