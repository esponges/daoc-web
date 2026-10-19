package rules

import (
	"errors"
	"fmt"
)

// Stat names one of the eight character stats.
type Stat int

const (
	STR Stat = iota
	CON
	DEX
	QUI
	INT
	PIE
	EMP
	CHA
	NumStats
)

var statNames = [NumStats]string{"strength", "constitution", "dexterity", "quickness", "intelligence", "piety", "empathy", "charisma"}

func (s Stat) String() string { return statNames[s] }

// Stats is a value for each stat, in Stat order.
type Stats [NumStats]int

// Race is a playable race and its starting stats.
type Race struct {
	ID   int
	Name string
	Base Stats
}

// Races are the Midgard races of 1.65, by the race numbers the data and the
// protocol use.
var Races = map[int]Race{
	5:  {5, "Norseman", Stats{70, 70, 50, 50, 60, 60, 60, 60}},
	6:  {6, "Troll", Stats{100, 70, 35, 35, 60, 60, 60, 60}},
	7:  {7, "Dwarf", Stats{60, 80, 50, 50, 60, 60, 60, 60}},
	8:  {8, "Kobold", Stats{50, 50, 70, 70, 60, 60, 60, 60}},
	14: {14, "Valkyn", Stats{55, 45, 65, 75, 60, 60, 60, 60}},
}

// Class is what a class brings to the rules. The abilities and specs it
// learns are data, imported with the world; these numbers are not.
type Class struct {
	ID                           int
	Name                         string
	Primary, Secondary, Tertiary Stat
	BaseHP                       int // hit points per level, before constitution
	WeaponSkillBase              int
	SpecMultiplier               int    // spec points per level, in tenths
	Base                         string // the base class it grows from (a Warrior is a Viking until 5 in 1.65)
	Races                        []int
}

// ClassWarrior is the Warrior's class number.
const ClassWarrior = 22

// Classes are the classes the rules cover.
var Classes = map[int]Class{
	ClassWarrior: {ID: ClassWarrior, Name: "Warrior", Primary: STR, Secondary: CON, Tertiary: DEX,
		BaseHP: 880, WeaponSkillBase: 460, SpecMultiplier: 20, Base: "Viking", Races: []int{7, 8, 5, 6, 14}},
}

// TaughtBy says whether a trainer of the named class can teach this one: its
// own trainers, and those of its base class.
func (c Class) TaughtBy(trainer string) bool {
	return trainer != "" && (trainer == c.Name || trainer == c.Base)
}

// MayBe reports whether a race may take the class.
func (c Class) MayBe(race int) bool {
	for _, r := range c.Races {
		if r == race {
			return true
		}
	}
	return false
}

// --- creation ---

// CreationPoints is what a new character spends on their stats.
const CreationPoints = 30

// RaiseCost is the points it takes to raise a stat by n above its base:
// one each, one more each above 10, one more again each above 15.
func RaiseCost(n int) int {
	if n <= 0 {
		return 0
	}
	return n + max(0, n-10) + max(0, n-15)
}

// DefaultBuild is the build used when none is chosen: ten points each in the
// class's primary, secondary and tertiary stats.
func DefaultBuild(c Class) Stats {
	var b Stats
	b[c.Primary], b[c.Secondary], b[c.Tertiary] = 10, 10, 10
	return b
}

// Create checks a creation build (how much each stat is raised) and returns
// the character's stats at level 1.
func Create(race Race, class Class, build Stats) (Stats, error) {
	if !class.MayBe(race.ID) {
		return Stats{}, fmt.Errorf("a %s cannot be a %s", race.Name, class.Name)
	}
	spent := 0
	for s, n := range build {
		if n < 0 {
			return Stats{}, fmt.Errorf("%s cannot go below its starting value", Stat(s))
		}
		spent += RaiseCost(n)
	}
	if spent != CreationPoints {
		return Stats{}, fmt.Errorf("a character spends exactly %d points on stats; this build spends %d", CreationPoints, spent)
	}
	out := race.Base
	for s := range out {
		out[s] += build[s]
	}
	return out, nil
}

// --- levels ---

// LevelGains is what levels 6 up to level add to a class's stats.
func LevelGains(c Class, level int) Stats {
	var g Stats
	for l := 6; l <= level; l++ {
		g[c.Primary]++
		if (l-6)%2 == 0 {
			g[c.Secondary]++
		}
		if (l-6)%3 == 0 {
			g[c.Tertiary]++
		}
	}
	return g
}

// StatCap is the most any stat can be.
const StatCap = 350

// ItemStatCap is the most items may add to one stat.
func ItemStatCap(level int) int { return level * 3 / 2 }

// ItemHPCap is the most items may add to hit points.
func ItemHPCap(level int) int { return level * 4 }

// Modified is a character's stats with level gains and capped item bonuses,
// less constitution lost to deaths, kept within 1..StatCap.
func Modified(base Stats, c Class, level int, items Stats, conLost int) Stats {
	out := base
	g := LevelGains(c, level)
	cp := ItemStatCap(level)
	for s := range out {
		out[s] += g[s] + min(items[s], cp)
	}
	out[CON] -= conLost
	for s := range out {
		out[s] = max(1, min(StatCap, out[s]))
	}
	return out
}

// MaxHP is a character's hit points from class, level and constitution, plus
// capped item hit points.
func MaxHP(c Class, level, con, itemHP int) int {
	k := con - 50
	if k < 0 {
		k *= 2
	}
	h := c.BaseHP * level
	hp := 20 + h/50 + h*k/10000 // Go truncates toward zero, as the rule says
	return max(1, hp+min(itemHP, ItemHPCap(level)))
}

// --- endurance and regeneration ---

// MaxEndurance is everyone's endurance pool.
const MaxEndurance = 100

// Endurance returned a second, out of and in combat.
const (
	EnduranceRegenOut = 6
	EnduranceRegenIn  = 2
)

// HealthRegenPeriod is the seconds between health ticks out of combat.
const HealthRegenPeriod = 3.0

// HealthRegen is the hit points a tick returns out of combat; in combat none.
func HealthRegen(level int, inCombat bool) int {
	if inCombat {
		return 0
	}
	return max(1, int(2.5+0.25*float64(level)))
}

// ErrUnknown is returned for a race or class the rules do not cover.
var ErrUnknown = errors.New("the rules do not cover that race or class")

func init() {
	register(
		Rule{ID: "character.race-stats", Capability: "rules-character-stats", Requirement: "Races give the starting stats",
			Status: Documented, Sources: []string{"disorder", "opendaoc"}, Scenarios: []string{"The Midgard races"}},
		Rule{ID: "character.class", Capability: "rules-character-stats", Requirement: "Classes name a primary, secondary and tertiary stat and base numbers",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Warrior definition"}},
		Rule{ID: "character.creation", Capability: "rules-character-stats", Requirement: "Creation spends exactly 30 points",
			Status: Documented, Sources: []string{"disorder", "opendaoc"}, Scenarios: []string{"Cost of a stat", "Default build", "Overspent build"}},
		Rule{ID: "character.level-stats", Capability: "rules-character-stats", Requirement: "Levels from 6 raise the class stats",
			Status: Documented, Sources: []string{"disorder", "opendaoc"}, Scenarios: []string{"Troll Warrior at 50", "Level 7"}},
		Rule{ID: "character.item-stats", Capability: "rules-character-stats", Requirement: "Items add to stats up to a cap",
			Status: Documented, Sources: []string{"disorder"}, Scenarios: []string{"Over the cap"}},
		Rule{ID: "character.hit-points", Capability: "rules-character-stats", Requirement: "Hit points come from class, level and constitution",
			Status: Reference, Sources: []string{"disorder", "opendaoc"}, Scenarios: []string{"Level 1 Troll Warrior", "Level 50 Troll Warrior", "Low constitution", "Item hit points"}},
		Rule{ID: "character.endurance", Capability: "rules-character-stats", Requirement: "Endurance is a pool of 100",
			Status: Provisional, Sources: []string{"project"}, Scenarios: []string{"After a style chain"}},
		Rule{ID: "character.health-regen", Capability: "rules-character-stats", Requirement: "Health returns slowly out of combat",
			Status: Provisional, Sources: []string{"opendaoc", "project"}, Scenarios: []string{"Level 20 resting"}},
	)
}
