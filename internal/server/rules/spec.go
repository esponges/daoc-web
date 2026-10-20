package rules

import "fmt"

// SpecPoints is the specialisation points a class has earned in all by a
// level, counting the half-level when half is true (from level 40 on, the
// midpoint of a level's experience).
func SpecPoints(c Class, level int, half bool) int {
	pts := 0
	for l := 2; l <= level; l++ {
		pts += levelSpecPoints(c, l)
		if l > 40 {
			pts += c.SpecMultiplier * (l - 1) / 20 // the half-level of the level before
		}
	}
	if half && level >= 40 && level < MaxLevel {
		pts += c.SpecMultiplier * level / 20
	}
	return pts
}

func levelSpecPoints(c Class, l int) int {
	if l <= 5 {
		return l
	}
	return c.SpecMultiplier * l / 10
}

// MaxLevel is the highest character level.
const MaxLevel = 50

// TrainCost is the points it takes to raise a spec from level from to level to.
func TrainCost(from, to int) int {
	cost := 0
	for n := from; n < to; n++ {
		cost += n + 1
	}
	return cost
}

// Train checks raising a spec from cur by n levels for a character of level
// with unspent points, and returns its cost.
func Train(level, cur, n, unspent int) (int, error) {
	if n < 1 {
		return 0, fmt.Errorf("nothing to train")
	}
	if cur+n > level {
		return 0, fmt.Errorf("you cannot train a skill above your level (%d)", level)
	}
	cost := TrainCost(cur, cur+n)
	if cost > unspent {
		return 0, fmt.Errorf("that costs %d points and you have %d", cost, unspent)
	}
	return cost, nil
}

// ItemSkillCap is the most items may add to one spec.
func ItemSkillCap(level int) int { return 1 + level/5 }

// CombatSpec is a spec as combat uses it: trained level plus capped item bonus.
func CombatSpec(trained, itemBonus, level int) int {
	return trained + min(max(0, itemBonus), ItemSkillCap(level))
}

// --- abilities ---

// Grant is one row of the class and spec ability tables: Source (a career
// such as "WarriorCareer" or a spec such as "Shields") at Level grants
// Ability at AbilityLevel.
type Grant struct {
	Source       string
	Level        int
	Ability      string
	AbilityLevel int
}

// Abilities resolves what a character knows: career grants by character
// level, spec grants by the spec's trained level. Where an ability is
// granted more than once the highest level wins; a level of 0 counts as 1.
func Abilities(grants []Grant, careers map[string]bool, level int, specs map[string]int) map[string]int {
	out := map[string]int{}
	for _, g := range grants {
		var reached bool
		if careers[g.Source] {
			reached = level >= g.Level
		} else if s, ok := specs[g.Source]; ok {
			reached = s >= g.Level
		}
		if !reached {
			continue
		}
		lv := max(1, g.AbilityLevel)
		if lv > out[g.Ability] {
			out[g.Ability] = lv
		}
	}
	return out
}

// Ability keys the rules act on.
const (
	AbEvade    = "Evade"
	AbShield   = "Shield"
	AbMidArmor = "MidArmor"
)

// ActiveAbility reports whether the server implements an ability; the others
// are known but have no effect yet.
func ActiveAbility(key string) bool {
	switch key {
	case AbEvade, AbShield, AbMidArmor, AbGuard, AbProtect, AbIntercept, AbEngage:
		return true
	}
	_, weapon := weaponryAbility[key]
	return weapon
}

// Item object types, as the item table numbers them.
const (
	ObjStaff    = 8
	ObjSword    = 11
	ObjHammer   = 12
	ObjAxe      = 13
	ObjSpear    = 14
	ObjCompBow  = 15
	ObjThrown   = 16
	ObjLeftAxe  = 17
	ObjCloth    = 32
	ObjLeather  = 33
	ObjStudded  = 34
	ObjChain    = 35
	ObjPlate    = 36
	ObjShield   = 42
	HandTwo     = 1 // the item table's Hand for two-handed weapons
	HandLeft    = 2 // and for left-hand ones
	DamageCrush = 1
	DamageSlash = 2
	DamageThrus = 3
)

// The Midgard weapon families and the ability each needs.
var weaponNeeds = map[int]string{
	ObjStaff:   "Weaponry: Staves",
	ObjSword:   "Weaponry: Swords",
	ObjHammer:  "Weaponry: Hammers",
	ObjAxe:     "Weaponry: Axes",
	ObjSpear:   "Weaponry: Spears",
	ObjCompBow: "Weaponry: Composite Bows",
	ObjThrown:  "Weaponry: Thrown",
	ObjLeftAxe: "Weaponry: Left Axes",
}

var weaponryAbility = func() map[string]bool {
	m := map[string]bool{}
	for _, a := range weaponNeeds {
		m[a] = true
	}
	return m
}()

// armourLevel is the MidArmor level each armour type needs.
var armourLevel = map[int]int{ObjCloth: 1, ObjLeather: 2, ObjStudded: 3, ObjChain: 4, ObjPlate: 5}

var armourName = map[int]string{ObjCloth: "cloth", ObjLeather: "leather", ObjStudded: "studded", ObjChain: "chain", ObjPlate: "plate"}

// MayUse says whether a character with these abilities may use an item of
// an object type; size is a shield's size (its damage type field, 1 to 3).
func MayUse(abilities map[string]int, objectType, size int) error {
	if a, ok := weaponNeeds[objectType]; ok {
		if abilities[a] == 0 {
			return fmt.Errorf("you have no training in that weapon")
		}
		return nil
	}
	if lv, ok := armourLevel[objectType]; ok {
		if abilities[AbMidArmor] < lv {
			return fmt.Errorf("you cannot wear %s armour yet", armourName[objectType])
		}
		return nil
	}
	if objectType == ObjShield {
		if have := abilities[AbShield]; have < max(1, size) {
			if have == 0 {
				return fmt.Errorf("you cannot use a shield")
			}
			return fmt.Errorf("that shield is too large for you yet")
		}
		return nil
	}
	return fmt.Errorf("you cannot use that")
}

func init() {
	register(
		Rule{ID: "spec.points", Capability: "rules-specialization", Requirement: "Levels grant specialisation points",
			Status: Reference, Sources: []string{"opendaoc", "disorder"}, Scenarios: []string{"Level 6 Warrior", "Half-level", "Total at 50"}},
		Rule{ID: "spec.training", Capability: "rules-specialization", Requirement: "Training a spec costs the level being bought",
			Status: Documented, Sources: []string{"disorder", "opendaoc"}, Scenarios: []string{"Sword to 10", "Above level", "Not enough points"}},
		Rule{ID: "spec.items", Capability: "rules-specialization", Requirement: "Items add to specs in combat",
			Status: Documented, Sources: []string{"disorder"}, Scenarios: []string{"A +3 sword at level 10"}},
		Rule{ID: "spec.abilities", Capability: "rules-specialization", Requirement: "Abilities come from the class's careers and specs",
			Status: Documented, Sources: []string{"data"}, Scenarios: []string{"Chain at level 10", "Evade at level 15", "An ability not yet implemented"}},
		Rule{ID: "spec.migration", Capability: "rules-specialization", Requirement: "Spec state is kept and migrated",
			Status: Documented, Sources: []string{"project"}, Scenarios: []string{"An existing level 20 Warrior"}},
	)
}
