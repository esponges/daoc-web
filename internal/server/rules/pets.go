package rules

import "math"

// Pets and charm (the pets spec, midgard-to-fifty part C). Numbers from
// findings.md, "7.1".
const (
	SpellSummonMinion    = "SummonMinion"
	SpellSummonCommander = "SummonCommander"
	SpellSummonSpirit    = "SummonSpiritFighter"
	SpellSummonHunterPet = "SummonHunterPet"
	SpellCharm           = "Charm"
)

// Pet distances, in game units, and times in seconds.
const (
	PetFollowDist  = 150.0   // how far behind its owner a following pet stands
	PetAggroRange  = 512.0   // how far an aggressive pet notices enemies
	PetLeash       = 10000.0 // beyond this from its owner a pet starts to be lost
	PetLeashGrace  = 10.0    // seconds beyond the leash before it is released
	CharmedHostile = 10.0    // seconds a creature whose charm ended is hostile to its old owner (provisional)
)

// Pet stances.
const (
	StancePassive    = "passive"
	StanceDefensive  = "defensive"
	StanceAggressive = "aggressive"
)

// ValidStance says whether a name is a stance.
func ValidStance(s string) bool {
	return s == StancePassive || s == StanceDefensive || s == StanceAggressive
}

// PetLevel is the level of a pet a summon spell makes: a spell damage below
// zero is a percent of the caster's level and one above is a level itself; the
// spell's value caps it, and it is never under 1.
func PetLevel(damage, value float64, casterLevel int) int {
	lv := int(damage)
	if damage < 0 {
		lv = int(math.Floor(float64(casterLevel) * -damage / 100))
	}
	if value > 0 {
		lv = min(lv, int(value))
	}
	return max(1, lv)
}

// Body types of the creature templates.
const (
	BodyAnimal   = 1
	BodyElement  = 4
	BodyHumanoid = 6
	BodyInsect   = 7
	BodyMagical  = 8
	BodyReptile  = 9
	BodyPlant    = 10
	BodyUndead   = 11
)

// CharmKindOK says whether a charm spell of a charm kind takes a creature of a
// body type. The kinds are the spell data's: 0 all, 1 humanoid, 2 animal, 3
// insect, 4 humanoid and animal, 5 those and insects and reptiles, 6 those and
// magical, plant and elemental, 7 those and undead, 8 reptile.
func CharmKindOK(kind, body int) bool {
	switch kind {
	case 0:
		return true
	case 1:
		return body == BodyHumanoid
	case 2:
		return body == BodyAnimal
	case 3:
		return body == BodyInsect
	case 4:
		return body == BodyHumanoid || body == BodyAnimal
	case 5:
		return CharmKindOK(4, body) || body == BodyInsect || body == BodyReptile
	case 6:
		return CharmKindOK(5, body) || body == BodyMagical || body == BodyPlant || body == BodyElement
	case 7:
		return CharmKindOK(6, body) || body == BodyUndead
	case 8:
		return body == BodyReptile
	}
	return false
}

// CharmLevelOK says whether a permanent charm takes a creature of a level: the
// spell's value is the highest, and the caster's own level is a limit too.
func CharmLevelOK(spellValue float64, casterLevel, mobLevel int) bool {
	return float64(mobLevel) <= spellValue && mobLevel <= casterLevel
}

// CharmResistChance is the chance a permanent charm is resisted: what the hit
// rule leaves.
func CharmResistChance(hitChance float64) float64 { return 1 - hitChance }

// MinionLimit is how many minions a commander summon controls: the number in the
// spell's resurrect-health column, which its text states ("Controls 3 additional
// pets").
func MinionLimit(n int) int { return max(0, n) }

func init() {
	const c = "pets"
	register(
		Rule{ID: "pet.level", Capability: c, Requirement: "A summon spell creates a pet from its template",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"A Spiritmaster summons"}},
		Rule{ID: "pet.charm", Capability: c, Requirement: "Charm takes an existing creature as a pet",
			Status: Provisional, Sources: []string{"opendaoc", "data", "project"}, Scenarios: []string{"Charm an animal", "Too high", "Wrong kind"}},
		Rule{ID: "pet.minions", Capability: c, Requirement: "A Bonedancer commands minions through a commander",
			Status: Reference, Sources: []string{"opendaoc", "data"}, Scenarios: []string{"Full army"}},
	)
}
