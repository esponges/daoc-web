package rules

import "time"

// The game clock is a function of real time: a game day (86,400 game seconds)
// lasts 60 real minutes, so game seconds run 24 times real seconds. It needs
// no state, so it survives restarts and every zone computes the same value.
const (
	DayGameSeconds   = 86400
	ClockRate        = 24
	NightStart       = 18 * 3600 // 18:00
	NightEnd         = 6 * 3600  // 06:00
	AppearGameSecond = 60        // timed spawns arrive over the first game minute
)

// GameSeconds is the game time of day, in seconds since midnight, at a real time.
func GameSeconds(t time.Time) int {
	return int((t.Unix() * ClockRate) % DayGameSeconds)
}

// IsNight says whether a game time of day is night: 18:00 up to 06:00.
func IsNight(gs int) bool { return gs >= NightStart || gs < NightEnd }

// When a spawn is present: always, only at night, or only by day.
const (
	Always    = 0
	NightOnly = 1
	DayOnly   = 2
)

// Present says whether a timed spawn is in the world at a game time. Each has
// its own offset into the first game minute of its time, so a zone's creatures
// arrive over that minute and not all at once; they leave together when it ends.
func Present(timing, offset, gs int) bool {
	switch timing {
	case NightOnly:
		return gs >= NightStart+offset || gs < NightEnd
	case DayOnly:
		return gs >= NightEnd+offset && gs < NightStart
	}
	return true
}

func init() {
	const c = "world-time"
	register(
		Rule{ID: "time.day-length", Capability: c, Requirement: "The server keeps a game clock",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Two clients agree", "A quarter of an hour later"}},
		Rule{ID: "time.night", Capability: c, Requirement: "Night is 18:00 to 06:00",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Dusk"}},
	)
}
