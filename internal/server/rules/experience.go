package rules

import "math"

// xpToReach is the published experience table: the total a character holds on
// reaching each level, by level.
var xpToReach = [MaxLevel + 1]int64{
	0, 0, 50, 250, 850, 2300, 6350, 15950, 37950, 88950,
	203950, 459950, 839950, 1399950, 2199950, 3399950, 5199950, 7899950, 11799950, 17499950,
	25899950, 38199950, 54699950, 76999950, 106999950, 146999950, 199999950, 269999950, 359999950, 479999950,
	639999950, 849999950, 1119999950, 1469999950, 1929999950, 2529999950, 3319999950, 4299999950, 5499999950, 6899999950,
	8599999950, 12899999950, 20699999950, 29999999950, 40799999950, 53999999950, 69599999950, 88499999950, 110999999950, 137999999950,
	169999999950,
}

// XPToReach is the experience at which a character reaches a level.
func XPToReach(level int) int64 {
	return xpToReach[max(1, min(MaxLevel, level))]
}

// LevelFor is the level a total of experience reaches.
func LevelFor(xp int64) int {
	l := 1
	for l < MaxLevel && xp >= xpToReach[l+1] {
		l++
	}
	return l
}

// HalfLevel is the experience at which a level's second half begins.
func HalfLevel(level int) int64 {
	if level >= MaxLevel {
		return math.MaxInt64
	}
	cur, next := XPToReach(level), XPToReach(level+1)
	return 1 + cur + (next-cur)/2
}

// creatureXP is the published table of what a creature is worth, by level.
// It follows 5 x 2^level to 10, then doubles every 2 levels to 20, every 3 to
// 35 and every 5 after, except at 15, 22 and 48 where the table has its own
// values.
var creatureXP = [MaxLevel + 1]int64{
	5, 10, 20, 40, 80, 160, 320, 640, 1280, 2560,
	5120, 7240, 10240, 14480, 20480, 28980, 40960, 57930, 81920, 115850,
	163840, 206435, 231705, 327680, 412850, 520160, 655360, 825702, 1040319, 1310720,
	1651404, 2080638, 2621440, 3302807, 4161277, 5242880, 6022488, 6918022, 7946720, 9128384,
	10485760, 12044975, 13836043, 15893440, 18258769, 20971520, 24089951, 27672087, 31625241, 36513537,
	41943040,
}

// CreatureXP is what killing a creature of a level is worth.
func CreatureXP(level int) int64 {
	if level < 0 {
		level = 0
	}
	if level <= MaxLevel {
		return creatureXP[level]
	}
	return int64(float64(creatureXP[MaxLevel]) * math.Pow(2, float64(level-MaxLevel)/5))
}

// XPCap is the most one kill gives a character, as a multiple of a same-level kill.
const XPCap = 1.25

// KillXP is one player's experience for a kill: their share of the damage
// times the creature's worth, capped by their own level. Grey kills are the
// caller's to refuse (see Grey).
func KillXP(creatureLevel, playerLevel int, share float64) int64 {
	if playerLevel >= MaxLevel {
		return 0
	}
	xp := float64(CreatureXP(creatureLevel)) * share
	return int64(math.Min(xp, XPCap*float64(CreatureXP(playerLevel))))
}

// Con colours, as numbers from grey to purple.
const (
	ConGrey = -3 + iota
	ConGreen
	ConBlue
	ConYellow
	ConOrange
	ConRed
	ConPurple
)

// conBand is the width of a con colour band for a player level.
func conBand(level int) int { return min(5, level/10+1) }

// Con is a creature's con colour to a player.
func Con(playerLevel, creatureLevel int) int {
	w, d := conBand(playerLevel), creatureLevel-playerLevel
	switch {
	case d <= -3*w:
		return ConGrey
	case d <= -2*w:
		return ConGreen
	case d <= -w:
		return ConBlue
	case d <= 0:
		return ConYellow
	case d <= w:
		return ConOrange
	case d <= 2*w:
		return ConRed
	}
	return ConPurple
}

// Grey says whether a creature is worth nothing to a player.
func Grey(playerLevel, creatureLevel int) bool { return Con(playerLevel, creatureLevel) == ConGrey }

func init() {
	register(
		Rule{ID: "experience.table", Capability: "rules-experience", Requirement: "The experience table",
			Status: Documented, Sources: []string{"silicondragon", "opendaoc"}, Scenarios: []string{"The thresholds", "Level 10 to 11"}},
		Rule{ID: "experience.creature-value", Capability: "rules-experience", Requirement: "Creatures are worth a value by level",
			Status: Documented, Sources: []string{"silicondragon", "opendaoc"}, Scenarios: []string{"Values", "The level 22 quirk"}},
		Rule{ID: "experience.kill", Capability: "rules-experience", Requirement: "Kills award a damage share, capped",
			Status: Documented, Sources: []string{"kb-xpclamp", "opendaoc"}, Scenarios: []string{"An even solo kill", "Capped", "Shared", "A grey kill"}},
		Rule{ID: "experience.con", Capability: "rules-experience", Requirement: "Con colours",
			Status: Documented, Sources: []string{"mythic-con"}, Scenarios: []string{"The bands", "Level 40", "Level 50"}},
	)
}
