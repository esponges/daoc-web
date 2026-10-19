package world

import (
	"fmt"
	"math"
	"slices"

	"daocweb/internal/server/rules"
)

// UnsafeStart is an imported start the start rule replaced, with the creature
// that made it unsafe and where new characters go instead.
type UnsafeStart struct {
	Realm, Race, Class int
	At                 Start
	Worst              rules.Threat
	Dist               float64
	Instead            Start
}

func (u UnsafeStart) String() string {
	return fmt.Sprintf("start realm %d race %d class %d in zone %d is unsafe: %s, level %d, %.0f units away; new characters start in zone %d at %.0f,%.0f instead",
		u.Realm, u.Race, u.Class, u.At.Zone, u.Worst.Name, u.Worst.Level, u.Dist, u.Instead.Zone, u.Instead.X, u.Instead.Y)
}

// checkStarts tests every start against the creatures around it as they
// spawn, and maps each unsafe one to its realm's safe start. The store is not
// changed: only where new characters are placed. It returns the report lines.
func (c *Catalog) checkStarts() []string {
	threats := map[int][]rules.Threat{}
	for _, s := range c.Starts {
		if _, done := threats[s.Zone]; done || c.db == nil {
			continue
		}
		spawns, err := c.SpawnsIn(s.Zone)
		if err != nil {
			return []string{fmt.Sprintf("start check skipped: %v", err)}
		}
		list := []rules.Threat{}
		for _, sp := range spawns {
			level := sp.Level
			if len(sp.Levels) > 0 {
				level = slices.Max(sp.Levels)
			}
			list = append(list, rules.Threat{Name: sp.Name, Level: level, X: sp.X, Y: sp.Y,
				Aggressive: sp.Realm == 0 && sp.AggroLevel > 0 && sp.AggroRange > 0})
		}
		threats[s.Zone] = list
	}

	type found struct {
		worst rules.Threat
		dist  float64
	}
	unsafe := map[int]found{}
	for i, s := range c.Starts {
		if w, d, bad := rules.UnsafeStart(s.X, s.Y, threats[s.Zone]); bad {
			unsafe[i] = found{w, d}
		}
	}

	// Each realm's safe start: the safe place most of its rows use.
	safe := map[int]Start{}
	realms := map[int]bool{}
	for _, s := range c.Starts {
		realms[s.Realm] = true
	}
	for realm := range realms {
		var places []Start
		var uses []int
		for i, s := range c.Starts {
			if s.Realm != realm || !c.Zones[s.Zone].usable() {
				continue
			}
			if _, bad := unsafe[i]; bad {
				continue
			}
			j := slices.IndexFunc(places, func(p Start) bool { return p.Zone == s.Zone && p.X == s.X && p.Y == s.Y })
			if j < 0 {
				places, uses = append(places, s.Start), append(uses, 0)
				j = len(places) - 1
			}
			uses[j]++
		}
		if k := rules.SafeStart(uses); k >= 0 {
			safe[realm] = places[k]
		}
	}

	c.replaced = map[int]UnsafeStart{}
	c.Unsafe = nil
	var report []string
	for i, f := range unsafe {
		s := c.Starts[i]
		instead, ok := safe[s.Realm]
		if !ok {
			report = append(report, fmt.Sprintf("start realm %d race %d class %d is unsafe (%s, level %d) and the realm has no safe start; kept",
				s.Realm, s.Race, s.Class, f.worst.Name, f.worst.Level))
			continue
		}
		u := UnsafeStart{Realm: s.Realm, Race: s.Race, Class: s.Class, At: s.Start, Worst: f.worst, Dist: f.dist, Instead: instead}
		c.replaced[i] = u
		c.Unsafe = append(c.Unsafe, u)
		report = append(report, u.String())
	}
	slices.Sort(report)
	if len(report) == 0 {
		report = []string{fmt.Sprintf("starts: all %d imported starts are safe", len(c.Starts))}
	}
	return report
}

// UnsafeStartNear says whether a place lies within the start rule's range of a
// start it replaced, and where to go instead.
func (c *Catalog) UnsafeStartNear(zone int, x, y float64) (UnsafeStart, bool) {
	for _, u := range c.Unsafe {
		if u.At.Zone == zone && math.Hypot(u.At.X-x, u.At.Y-y) <= rules.UnsafeRange {
			return u, true
		}
	}
	return UnsafeStart{}, false
}
