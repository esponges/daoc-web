package rules

import "math"

// Stealth at 1.65, from the rules-stealth capability. The numbers are
// OpenDAoC-Core's, which cite a critshot forum thread and disorder.dk for the
// detection ranges (provisional: no first-party source).

// StealthChangeGap is the seconds that must pass between hiding and un-hiding.
const StealthChangeGap = 2.0

// StealthSpeed is the share of normal speed a hidden player moves at: 0.3 plus
// 0.3 scaled by their Stealth spec (never counted above their level) over their
// level, each plus 10.
func StealthSpeed(spec, level int) float64 {
	s := math.Min(float64(spec), float64(level))
	return 0.3 + (s+10)*0.3/float64(level+10)
}

// PlayerDetectRange is how far a player looking sees a hidden player: the gap
// between the viewer's level and the hider's Stealth spec (spec counted to 50,
// the gap never below 0) times 20 plus 125; times 50 plus 250 for a viewer with
// Detect Hidden when the hider has neither Detect Hidden nor Camouflage.
func PlayerDetectRange(viewerLevel, hiderStealth int, detectHidden, hiderCounters bool) float64 {
	gap := max(0, viewerLevel-min(50, hiderStealth))
	if detectHidden && !hiderCounters {
		return float64(gap*50 + 250)
	}
	return float64(gap*20 + 125)
}

// CreatureDetectRadius is how far a creature notices a hidden player: 125 plus
// 20 for each level it is above their Stealth spec (down to the floor of 126),
// plus 125 more for a creature that detects the hidden when the hider has
// nothing to counter it.
func CreatureDetectRadius(creatureLevel, hiderStealth int, detectHidden, hiderCounters bool) float64 {
	r := 125 + float64(max(creatureLevel, 1)-min(50, hiderStealth))*20
	if detectHidden && !hiderCounters {
		r += 125
	}
	return math.Max(126, r)
}

// StealthTooClose is how near an enemy may be for a player to hide: 1024 less
// 896 times their Stealth spec (to 50) over the enemy's level; double that
// range for an enemy who detects the hidden. Zero or less means no limit.
func StealthTooClose(enemyLevel, hiderStealth int, detectHidden bool) float64 {
	lv := float64(max(1, enemyLevel))
	st := float64(min(50, hiderStealth))
	if detectHidden {
		return 2048 - 1792*st/lv
	}
	return 1024 - 896*st/lv
}

func init() {
	const c = "rules-stealth"
	register(
		Rule{ID: "stealth.speed", Capability: c, Requirement: "Hidden movement is slower",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Stealth speed"}},
		Rule{ID: "stealth.detection", Capability: c, Requirement: "Others see a hidden player only within detection range",
			Status: Provisional, Sources: []string{"opendaoc", "disorder"}, Scenarios: []string{"Detect Hidden", "Creature detection", "Too close to hide"}},
	)
}
