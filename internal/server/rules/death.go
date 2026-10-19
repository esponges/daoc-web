package rules

// PenaltyLevel is the level from which deaths cost experience and constitution.
const PenaltyLevel = 6

// LossPermille is the share of a level's span, in tenths of a percent, a death
// costs, given how many deaths the character has had since levelling.
func LossPermille(level, deathsBefore int) int {
	p := 10
	if level < 40 {
		p = 50 - level
	}
	switch deathsBefore {
	case 0:
		p /= 3
	case 1:
		p = p * 2 / 3
	}
	return p
}

// DeathLoss is the experience a death (or the release after it) costs a
// character of a level holding xp, never taking them below the level's start.
func DeathLoss(level int, xp int64, deathsBefore int) int64 {
	if level < PenaltyLevel || level >= MaxLevel {
		return 0
	}
	span := XPToReach(level+1) - XPToReach(level)
	loss := span * int64(LossPermille(level, deathsBefore)) / 1000
	return max(0, min(loss, xp-XPToReach(level)))
}

// ConLoss is the constitution a release costs, by the death count since
// levelling, this death included.
func ConLoss(deaths int) int { return max(1, min(3, deaths)) }

// restorePrice is the published price in copper, by level, of restoring one
// point of constitution at a healer.
var restorePrice = [MaxLevel + 1]int64{
	1, 3, 6, 10, 15, 21, 33, 53, 82, 125,
	188, 278, 352, 443, 553, 688, 851, 1048, 1288, 1578,
	1926, 2347, 2721, 3146, 3633, 4187, 4820, 5537, 6356, 7281,
	8337, 9532, 10886, 12421, 14161, 16131, 18360, 19965, 21857, 23821,
	25928, 28244, 30731, 33411, 36308, 39438, 42812, 46454, 50385, 54625,
	59195,
}

// RestorePrice is the price of restoring one point of constitution.
func RestorePrice(level int) int64 { return restorePrice[max(0, min(MaxLevel, level))] }

// Restore is how many of the lost points the money pays for, and the cost.
func Restore(level, lost int, money int64) (points int, cost int64) {
	price := RestorePrice(level)
	points = int(min(int64(lost), money/price))
	return points, int64(points) * price
}

// PrayRadius and PraySeconds govern praying at a gravestone.
const (
	PrayRadius  = 2000.0
	PraySeconds = 5.0
)

func init() {
	register(
		Rule{ID: "death.loss", Capability: "rules-death", Requirement: "Dying to a creature costs experience",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"First death at level 10", "Third death at level 10", "Below level 6", "At the start of a level"}},
		Rule{ID: "death.release", Capability: "rules-death", Requirement: "Releasing costs the same again, into a gravestone",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Release after the first death"}},
		Rule{ID: "death.pray", Capability: "rules-death", Requirement: "Praying at your gravestone returns its experience",
			Status: Reference, Sources: []string{"opendaoc"}},
		Rule{ID: "death.constitution", Capability: "rules-death", Requirement: "Releasing costs constitution",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Two deaths"}},
		Rule{ID: "death.healer", Capability: "rules-death", Requirement: "Healers restore lost constitution for money",
			Status: Documented, Sources: []string{"silicondragon"}, Scenarios: []string{"Restore at level 12", "Not enough money"}},
		Rule{ID: "death.return", Capability: "rules-death", Requirement: "A released character returns whole",
			Status: Reference, Sources: []string{"opendaoc"}},
	)
}
