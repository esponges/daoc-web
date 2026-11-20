package rules

import "math"

// Archery at 1.65, from the rule book's rules-archery capability. A bow shot is
// a swing from range: it is rolled like one, with no parry, and damaged by the
// bow's numbers. The abilities change one shot each.

const ObjArrow = 43 // the object type of an arrow in the item table

// Ammunition bits live in an arrow's speed/absolute field: bits 2-3 are its
// range class, bits 4-5 its accuracy class.
const (
	MaxDraw      = 15.0 // seconds a bow may be held drawn before the archer tires
	MinRapidDraw = 0.9  // a rapid fire draw is never shorter than this, seconds
)

// BowRange is how far a bow shoots, in units: by the bow's object type, the
// archer's range modifier in percent (100 for none) and the arrow's range
// class (0 clout, 1 plain, 3 flight). Provisional: one source.
func BowRange(objectType, rangePct, rangeClass int) float64 {
	base := 1200.0
	switch objectType {
	case 9: // longbow
		base = 1760
	case 18: // recurved bow
		base = 1680
	case ObjCompBow:
		base = 1600
	}
	r := math.Max(32, base*float64(rangePct)/100)
	switch rangeClass {
	case 0:
		r *= 0.85
	case 3:
		r *= 1.25
	}
	return math.Max(32, r)
}

// ArrowRangeClass and ArrowAccuracyClass read an arrow's two quality classes.
func ArrowRangeClass(spdAbs int) int    { return (spdAbs >> 2) & 3 }
func ArrowAccuracyClass(spdAbs int) int { return (spdAbs >> 4) & 3 }

// ArrowMissScale is what an arrow's accuracy class does to the miss chance: rough
// arrows miss 15 percent more, footed ones 25 percent less.
func ArrowMissScale(accuracyClass int) float64 {
	switch accuracyClass {
	case 0:
		return 1.15
	case 3:
		return 0.75
	}
	return 1
}

// DrawTime is the seconds a normal draw takes: the swing rule on the bow's
// speed and the archer's quickness (never under 1.5). Provisional.
func DrawTime(speed float64, quickness int) float64 { return SwingTime(speed, quickness) }

// CriticalDraw is a Critical Shot draw of rank 1 to 9: twice the normal draw,
// less a tenth of the normal draw for each rank above the first. Provisional.
func CriticalDraw(base float64, rank int) float64 {
	return base*2 - float64(max(1, rank)-1)*base/10
}

// CriticalShotDamage is what a Critical Shot's damage is multiplied by: 2 for
// an even match, less 0.075 for each level the target is above the archer,
// kept between 1.1 and 2. Provisional.
func CriticalShotDamage(archerLevel, targetLevel int) float64 {
	return math.Max(1.1, math.Min(2.0, 2+float64(archerLevel-targetLevel)*0.075))
}

// RapidDraw is a rapid fire draw: half the normal one, never under 0.9 seconds.
func RapidDraw(base float64) float64 { return math.Max(MinRapidDraw, base/2) }

// RapidFireDamage is the share of its damage a rapid fire shot does: the share
// of the normal draw it used (Camelot Herald, Grab Bag: release at 75 percent
// of the time and it does 75 percent damage). At most 1.
func RapidFireDamage(drawn, normal float64) float64 {
	if normal <= 0 || drawn >= normal {
		return 1
	}
	return math.Max(0, drawn/normal)
}

// ShotEndurance is a shot's endurance cost: 5, or half that at Rapid Fire
// rank 2; 10 for a Critical Shot. Provisional.
func ShotEndurance(critical bool, rapidRank int) int {
	switch {
	case critical:
		return 10
	case rapidRank == 2:
		return 3
	}
	return 5
}

// PenetratingEffect is the share of its damage an arrow does through a
// bladeturn: 25 percent plus 25 for each rank of Penetrating Arrow; nothing
// when the archer cast the bladeturn themselves, rank 0 is the
// base quarter.
func PenetratingEffect(rank int, ownBladeturn bool) float64 {
	if ownBladeturn {
		return 0
	}
	return 0.25 + float64(max(0, rank))*0.25
}

func init() {
	const c = "rules-archery"
	register(
		Rule{ID: "archery.draw", Capability: c, Requirement: "Drawing takes the bow's speed and can be interrupted",
			Status: Provisional, Sources: []string{"opendaoc", "herald"}, Scenarios: []string{"A full draw"}},
		Rule{ID: "archery.range", Capability: c, Requirement: "A bow shot needs a bow, an arrow and a visible target in range",
			Status: Provisional, Sources: []string{"opendaoc"}, Scenarios: []string{"Range by bow"}},
		Rule{ID: "archery.damage", Capability: c, Requirement: "An arrow hits and does damage like a swing from range",
			Status: Provisional, Sources: []string{"opendaoc", "disorder"}, Scenarios: []string{"A worked example", "Parry does not apply"}},
		Rule{ID: "archery.abilities", Capability: c, Requirement: "Critical Shot, Rapid Fire and Penetrating Arrow change a shot",
			Status: Provisional, Sources: []string{"opendaoc", "herald"}, Scenarios: []string{"Critical Shot", "Rapid Fire", "Penetrating Arrow"}},
	)
}
