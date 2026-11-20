package rules

import "strings"

// Tasks, from the rules-tasks capability. Numbers and the choice of givers are
// OpenDAoC-Core's (provisional: no 1.65 account of them was found).

const (
	TaskMaxLevel   = 20       // tasks are for players up to this level
	TaskSeconds    = 2 * 3600 // the time a task lasts
	TaskAskPause   = 300.0    // seconds to wait after a giver had nothing for you
	DeliveryChance = 0.35     // the chance an ask of a delivery giver succeeds; a kill giver always does
	TaskKillBelow  = 2        // a kill target may be this many levels under the player
)

// MaxTasks is how many tasks a player of a level may finish before they level up
// (the count resets then): their level, none above the cap.
func MaxTasks(level int) int {
	if level < 1 || level > TaskMaxLevel {
		return 0
	}
	return level
}

var killMoney = [TaskMaxLevel]int{28, 57, 77, 105, 140, 190, 257, 347, 470, 632, 735, 852, 987, 1147, 1330, 1542, 1790, 2077, 2407, 2801}
var deliveryMoney = [TaskMaxLevel]int{25, 34, 46, 63, 84, 114, 154, 208, 282, 379, 441, 511, 592, 688, 798, 925, 1074, 1246, 1444, 1830}

// TaskMoney is a finished task's money in copper: the table's value for the
// level, 3 percent either way by roll in 0..1. Zero above the cap.
func TaskMoney(level int, delivery bool, roll float64) int {
	if level < 1 || level > TaskMaxLevel {
		return 0
	}
	base := killMoney[level-1]
	if delivery {
		base = deliveryMoney[level-1]
	}
	spread := base / 100 * 3
	return base - spread + int(roll*float64(2*spread+1))
}

// TaskXP is a finished task's experience: half of the player's current level's
// experience bar, shared over their level less one (at least one), so that
// finishing the tasks a level allows pays about half a bar. OpenDAoC-Core's code
// uses the experience to reach the level in all, which would pay many bars; its
// comment says "50 percent of total xp for level", which is the reading used.
func TaskXP(level int) int64 {
	if level < 1 || level > TaskMaxLevel {
		return 0
	}
	bar := XPToReach(level+1) - XPToReach(level)
	return int64(float64(bar) * 0.5 / float64(max(1, level-1)))
}

// KillTargetLevels is the levels of creature a kill task may name: from two
// under the player to their own.
func KillTargetLevels(level int) (lo, hi int) { return max(1, level-TaskKillBelow), level }

// KillGiver says whether a named NPC of the realm gives kill tasks: a town
// guard figure or soldier by the rule's name patterns. guild is the NPC's guild
// name, empty for none.
func KillGiver(name, guild string) bool {
	name = strings.TrimSpace(name)
	has := func(p string) bool { return strings.HasPrefix(name, p) }
	end := func(s string) bool { return strings.HasSuffix(name, s) }
	switch {
	case strings.Contains(name, "Guard"):
		for _, no := range []string{"Guardian", "Guardian Sergeant", "Guard", "Guardsman", "Guard's Armorer"} {
			if name == no {
				return false
			}
		}
		return !end("Guardian") && !end("Guard") && !end("Guardsman") && !has("Guardian of the")
	case has("Sir ") || has("Captain ") || has("Lady "):
		return guild == ""
	case has("Jarl ") || has("Soldier ") || has("Soldat "):
		return true
	case has("Sentinel "):
		return !end("Runes") && !end("Kynon")
	case strings.Contains(name, "Viking"):
		return !end("Archer") && !end("Dreng") && !end("Huscarl") && !end("Jarl")
	case has("Huntress ") || has("Bork "):
		return true
	case strings.Contains(name, "Flayer"):
		return end("Kegnar") // the one named Flayer Kegnar
	}
	return false
}

// DeliveryGiver says whether an NPC gives delivery tasks: a trainer, a merchant
// or a town crier.
func DeliveryGiver(trainer, merchant bool, name string) bool {
	return trainer || merchant || strings.Contains(name, "Crier")
}

func init() {
	const c = "rules-tasks"
	register(
		Rule{ID: "task.limits", Capability: c, Requirement: "Town NPCs offer tasks",
			Status: Provisional, Sources: []string{"opendaoc"}, Scenarios: []string{"Task limits", "Task givers"}},
		Rule{ID: "task.reward", Capability: c, Requirement: "A finished task pays experience and money",
			Status: Provisional, Sources: []string{"opendaoc"}, Scenarios: []string{"Reward at level 10", "Rewards by level band"}},
	)
}
