package rules

// Resurrection (rules-death). A dead character can be raised by another
// player's spell instead of releasing.

const (
	// RezOfferSeconds is how long a character has to answer an offer.
	RezOfferSeconds = 15.0
	// ReleaseAfter is how long a character has been dead before they may release.
	ReleaseAfter = 10.0
	// ReleaseForcedAfter is how long they may wait, dead, before release is forced.
	ReleaseForcedAfter = 900.0
	// SicknessSeconds and SicknessPercent are resurrection sickness: for this long
	// the character's damage and spell strength are lowered by this percent.
	SicknessSeconds = 60.0
	SicknessPercent = 30
)

// ReleaseStage is the second stage of a death: the experience lost again, the
// constitution lost and whether a gravestone holds the experience. A character
// who was raised skips it; what the death itself took stays lost.
func ReleaseStage(raised bool, lastLoss, xp, levelStart int64, deaths int) (xpLoss int64, conLoss int, grave bool) {
	if raised {
		return 0, 0, false
	}
	xpLoss = min(lastLoss, max(0, xp-levelStart))
	return xpLoss, ConLoss(deaths), xpLoss > 0
}

// Sickness is the resurrection sickness a resurrection leaves, by the percent
// of hit points it restored: a full resurrection halves the time.
func Sickness(restoredPercent int) (seconds float64, percent int) {
	seconds = SicknessSeconds
	if restoredPercent >= 100 {
		seconds /= 2
	}
	return seconds, SicknessPercent
}

// SicknessFactor is the factor sickness puts on damage and spell strength.
func SicknessFactor(sicknessPercent int) float64 { return 1 - float64(sicknessPercent)/100 }

// CanRelease says whether a character dead for this long may release.
func CanRelease(deadFor float64) bool { return deadFor >= ReleaseAfter }

// MustRelease says whether a character dead for this long is released.
func MustRelease(deadFor float64) bool { return deadFor >= ReleaseForcedAfter }

// OfferLapsed says whether an offer this old has gone unanswered.
func OfferLapsed(age float64) bool { return age >= RezOfferSeconds }

func init() {
	const d = "rules-death"
	register(
		Rule{ID: "rez.spare", Capability: d, Requirement: "Resurrection spares the release",
			Status: Documented, Sources: []string{"herald-patches"}, Scenarios: []string{"Raised at level 20"}},
		Rule{ID: "rez.sickness", Capability: d, Requirement: "Resurrection sickness",
			Status: Reference, Sources: []string{"data", "dolsharp"}, Scenarios: []string{"Sickness shows"}},
		Rule{ID: "rez.wait", Capability: d, Requirement: "Waiting for a resurrection",
			Status: Reference, Sources: []string{"herald-patches", "dolsharp"}, Scenarios: []string{"Release while an offer is pending"}},
	)
}
