package rules

// Crafting: skills, what a recipe asks of them, and how a craft turns out.
// A skill runs from 0 to 1100: the official tradeskill overview says the best
// gear needs 1050 to 1100. The numbers below are this project's own shape
// (no public source gives the 1.65 curves), so every rule here is
// provisional; they live in one file so play can tune them.

// The tradeskills, by the numbers the community database uses for a recipe's
// CraftingSkillType (read off the recipes: 6 makes brackets and hinges, 7
// bridles, 8 dolls, 14 breadboards). Gem cutting (9), herbal work (10) and
// spellcrafting (13) have no recipes in the data.
const (
	SkillWeapon  = 1
	SkillArmor   = 2
	SkillAlchemy = 4
	SkillMetal   = 6
	SkillLeather = 7
	SkillCloth   = 8
	SkillGem     = 9
	SkillTailor  = 11
	SkillFletch  = 12
	SkillSpell   = 13
	SkillWood    = 14
	MaxSkill     = 1100
	// A recipe this far above the skill cannot be tried.
	maxAttemptGap = 50
)

// Profession is a primary craft a character joins at its master. Its own
// skill is the primary; the others are secondary skills it also trains.
type Profession struct {
	Primary   int
	Name      string
	Secondary []int
}

// Professions are the six masters Midgard places, by primary skill.
var Professions = map[int]Profession{
	SkillWeapon:  {SkillWeapon, "Weaponcrafting", []int{SkillMetal, SkillWood}},
	SkillArmor:   {SkillArmor, "Armorcrafting", []int{SkillMetal, SkillLeather, SkillCloth}},
	SkillAlchemy: {SkillAlchemy, "Alchemy", nil},
	SkillFletch:  {SkillFletch, "Fletching", []int{SkillWood}},
	SkillSpell:   {SkillSpell, "Spellcrafting", []int{SkillGem}},
	SkillTailor:  {SkillTailor, "Tailoring", []int{SkillLeather, SkillCloth}},
}

// SkillName is a skill's name for the crafting window.
func SkillName(s int) string {
	switch s {
	case SkillWeapon:
		return "Weaponcrafting"
	case SkillArmor:
		return "Armorcrafting"
	case SkillAlchemy:
		return "Alchemy"
	case SkillMetal:
		return "Metalworking"
	case SkillLeather:
		return "Leatherworking"
	case SkillCloth:
		return "Clothworking"
	case SkillGem:
		return "Gem cutting"
	case SkillTailor:
		return "Tailoring"
	case SkillFletch:
		return "Fletching"
	case SkillSpell:
		return "Spellcrafting"
	case SkillWood:
		return "Woodworking"
	}
	return "Crafting"
}

// StartSkill is what joining a profession gives the primary skill, so the
// first recipes can be made at once.
const StartSkill = 1

// SecondaryCap is the most a secondary skill may reach: the primary's own
// skill, so a crafter's side skills never outrun their craft.
func SecondaryCap(primarySkill int) int { return primarySkill }

// MayRaise says whether a skill may rise by one: the cap of 1100 for a
// primary, and the primary's level for a secondary.
func MayRaise(skill, value int, primary bool, primaryValue int) bool {
	if value >= MaxSkill {
		return false
	}
	return primary || value < SecondaryCap(primaryValue)
}

// SuccessChance is the percent chance a craft succeeds, from the crafter's
// skill and the recipe's level. At or above the level it cannot fail; each
// point a recipe is above the skill costs a point of chance, and a recipe
// more than 50 above cannot be tried (0).
func SuccessChance(skill, level int) int {
	gap := level - skill
	switch {
	case gap <= 0:
		return 100
	case gap > maxAttemptGap:
		return 0
	}
	return 100 - gap
}

// GainChance is the percent chance a successful craft raises the skill by one:
// certain while the recipe is at or above the skill, falling a point for each
// point the skill is above it, and none from 100 above.
func GainChance(skill, level int) int {
	over := skill - level
	switch {
	case over <= 0:
		return 100
	case over >= 100:
		return 0
	}
	return 100 - over
}

// CraftSeconds is how long a craft takes: a base, more for a higher recipe
// and for each ten materials, never under 3 seconds or over a minute.
func CraftSeconds(level, materials int) int {
	t := 2 + level/50 + materials/10
	return min(max(t, 3), 60)
}

// FailureLoss is how many of a material stack a failed craft uses up: half,
// rounded down, but at least one of any stack of two or more.
func FailureLoss(count int) int {
	if count <= 1 {
		return 0
	}
	return max(count/2, 1)
}

// Quality is the percent quality of a finished item, 94 to 100. It starts at
// 94 and each further point is won with a probability that is 50% for a
// recipe at the crafter's skill, one point of percent more for each ten skill
// above it, within 20% to 80%. 100 is a masterpiece. rolls supplies the dice,
// one value in [0,1) per step, taken until one fails.
func Quality(skill, level int, rolls func() float64) int {
	p := 0.5 + float64(skill-level)/1000
	p = min(max(p, 0.2), 0.8)
	q := 94
	for q < 100 && rolls() < p {
		q++
	}
	return q
}

// Masterpiece says whether a quality is the perfect 100.
func Masterpiece(q int) bool { return q >= 100 }

func init() {
	const c = "rules-crafting"
	register(
		Rule{ID: "crafting.success", Capability: c, Requirement: "Success chance", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Orange recipe", "Out of reach"}},
		Rule{ID: "crafting.quality", Capability: c, Requirement: "Quality", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Distribution"}},
		Rule{ID: "crafting.time", Capability: c, Requirement: "Craft time", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Time"}},
		Rule{ID: "crafting.skill", Capability: c, Requirement: "Skill gain", Status: Provisional,
			Sources: []string{"project"}, Scenarios: []string{"Gain", "Secondary cap"}},
	)
}
