package rules

import "math"

// GroupAward is a group's experience for a kill: the creature's value times
// the share of its damage the group's members dealt together.
func GroupAward(creatureLevel int, share float64) float64 {
	return float64(CreatureXP(creatureLevel)) * share
}

// GroupGrey says whether a kill gives a group nothing: the creature cons grey
// to the highest-level member in range.
func GroupGrey(creatureLevel int, levels []int) bool {
	top := 0
	for _, l := range levels {
		top = max(top, l)
	}
	return top > 0 && Grey(top, creatureLevel)
}

// GroupSplit divides a group's award among the members in range by their
// levels, rounding down, and caps each part at 1.25 times a kill of the
// member's own level. levels are those members' levels; the parts come back
// in the same order. A member at the level cap gets nothing.
func GroupSplit(award float64, levels []int) []int64 {
	out := make([]int64, len(levels))
	total := 0
	for _, l := range levels {
		total += l
	}
	if total == 0 || award <= 0 {
		return out
	}
	for i, l := range levels {
		if l >= MaxLevel {
			continue
		}
		part := math.Floor(award * float64(l) / float64(total))
		out[i] = int64(math.Min(part, XPCap*float64(CreatureXP(l))))
	}
	return out
}

// GroupKillXP is what each member in range gains from a group's kill.
func GroupKillXP(creatureLevel int, share float64, levels []int) []int64 {
	if GroupGrey(creatureLevel, levels) {
		return make([]int64, len(levels))
	}
	return GroupSplit(GroupAward(creatureLevel, share), levels)
}

func init() {
	register(
		Rule{ID: "experience.group-share", Capability: "rules-experience", Requirement: "A group shares a kill as one damager",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"A group and a solo player"}},
		Rule{ID: "experience.group-split", Capability: "rules-experience", Requirement: "A group's award is split by level among members in range",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"Two even members", "Split by level, then capped", "A member far away"}},
		Rule{ID: "experience.group-grey", Capability: "rules-experience", Requirement: "A group kill is grey by its highest member in range",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"A high-level helper"}},
	)
}
