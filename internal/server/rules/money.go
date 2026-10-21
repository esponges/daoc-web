package rules

// KillMoney is the copper a creature of a level carries, before the caller's
// spread. It is this project's own curve until money drops get a rule.
func KillMoney(level int) int64 {
	level = max(1, level)
	return int64(4 + level*level*2 + level*5)
}

func init() {
	register(Rule{ID: "loot.money", Capability: "character-progression", Requirement: "Killing drops loot",
		Status: Provisional, Sources: []string{"project"}})
}

// GenericDropPercent is the chance, in percent, that a kill with no named drop
// leaves one item of about the creature's level: 22 at level 1, one more for
// every two levels, never above 40. No source gives 1.65's generators, so this
// is this project's choice.
func GenericDropPercent(level int) int {
	level = max(1, level)
	return min(40, 22+(level-1)/2)
}

func init() {
	register(Rule{ID: "loot.generic", Capability: "loot-and-release", Requirement: "Loot by creature level",
		Status: Provisional, Sources: []string{"project"}})
}
