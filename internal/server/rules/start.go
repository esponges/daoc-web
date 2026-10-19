package rules

import "math"

// A start is unsafe when an aggressive creature at least UnsafeLevel is within
// UnsafeRange of it: five levels above a new character, close enough to be
// pulled on the first steps.
const (
	UnsafeLevel = 6
	UnsafeRange = 1500.0
)

// Threat is a creature near a start, as it spawns.
type Threat struct {
	Name       string
	Level      int
	Aggressive bool
	X, Y       float64
}

// UnsafeStart reports whether a start at x, y is unsafe, with the hardest
// creature that makes it so (the highest level, then the nearest) and its
// distance.
func UnsafeStart(x, y float64, near []Threat) (worst Threat, dist float64, unsafe bool) {
	for _, t := range near {
		if !t.Aggressive || t.Level < UnsafeLevel {
			continue
		}
		d := math.Hypot(t.X-x, t.Y-y)
		if d > UnsafeRange {
			continue
		}
		if !unsafe || t.Level > worst.Level || (t.Level == worst.Level && d < dist) {
			worst, dist, unsafe = t, d, true
		}
	}
	return
}

// SafeStart picks the realm's safe start among its safe ones: the one the
// most race and class rows use, the first on a tie. It returns -1 when none
// is safe.
func SafeStart(uses []int) int {
	best := -1
	for i, n := range uses {
		if best < 0 || n > uses[best] {
			best = i
		}
	}
	return best
}

func init() {
	register(
		Rule{ID: "start.town", Capability: "starting-zone-safety", Requirement: "New characters start where 1.65 started them",
			Status: Documented, Sources: []string{"zam-starts", "data"}},
		Rule{ID: "start.unsafe", Capability: "starting-zone-safety", Requirement: "An unsafe start is replaced by the realm's safe start",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"An unsafe start", "A safe start is kept"}},
	)
}
