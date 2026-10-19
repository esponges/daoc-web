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
