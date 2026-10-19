package rules

// CreatureWeaponSkillBase is every ordinary creature's weapon skill base.
const CreatureWeaponSkillBase = 440

// CreatureStat is a creature's strength, constitution, dexterity and quickness.
func CreatureStat(level int) int { return 30 + max(0, level-1) }

// CreatureHP is a creature's maximum hit points.
func CreatureHP(level int) int {
	l := float64(level)
	con := float64(CreatureStat(level))
	return max(1, int((con-25)*1.8+11*(l+1)+0.6*l*l))
}

// CreatureArmour is a creature's armour factor and absorption on every location.
func CreatureArmour(level int) (af, absorb float64) {
	l := float64(level)
	return float64(int((1+l/200)*(l*12.5))) / 5, l * 0.0054
}

// CreatureWeaponSpeed is a creature's swing, in seconds, by what it holds.
func CreatureWeaponSpeed(twoHanded bool) float64 {
	if twoHanded {
		return 4.0
	}
	return 3.0
}

// Creature describes an ordinary creature for a swing. The template's defences
// are percents; resists are the race's, by damage type.
func Creature(level int, evade, parry, block float64, twoHanded bool, damageType int, resists map[int]float64) Combatant {
	level = max(1, level)
	st := CreatureStat(level)
	af, abs := CreatureArmour(level)
	c := Combatant{
		Level:           level,
		WeaponSkillBase: CreatureWeaponSkillBase,
		Weapon:          Weapon{Speed: CreatureWeaponSpeed(twoHanded), TwoHanded: twoHanded, Quality: 1, Condition: 1, DamageType: damageType},
		EvadePct:        evade, ParryPct: parry, BlockPct: block,
		CreatureAF: af, CreatureAbsorb: abs,
		Resists: resists,
	}
	for s := range c.Stats {
		c.Stats[s] = st
	}
	return c
}

// CreatureSwing is a creature's seconds between swings.
func CreatureSwing(c *Combatant) float64 { return CreatureSwingTime(c.Weapon.Speed, c.Stats[QUI]) }

func init() {
	register(
		Rule{ID: "creature.stats", Capability: "rules-creature-combat", Requirement: "Creature stats follow their level",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Level 10 creature"}},
		Rule{ID: "creature.hit-points", Capability: "rules-creature-combat", Requirement: "Creature hit points",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Hit points by level"}},
		Rule{ID: "creature.blows", Capability: "rules-creature-combat", Requirement: "Creature blows",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A level 10 creature hits a studded torso", "Bare skin hits the cap"}},
		Rule{ID: "creature.armour", Capability: "rules-creature-combat", Requirement: "Creature armour",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Level 50 creature"}},
		Rule{ID: "creature.swing", Capability: "rules-creature-combat", Requirement: "Creature swing speed",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Level 10 one-handed creature"}},
		Rule{ID: "creature.resists", Capability: "rules-creature-combat", Requirement: "Creature resists come from its race",
			Status: Documented, Sources: []string{"data"}, Scenarios: []string{"A thrust weakness"}},
		Rule{ID: "creature.defences", Capability: "rules-creature-combat", Requirement: "Creature defences come from the template",
			Status: Documented, Sources: []string{"data"}, Scenarios: []string{"A parrying bandit"}},
		Rule{ID: "creature.level-one", Capability: "rules-creature-combat", Requirement: "Fights at level 1 follow the same rules",
			Status: Reference, Sources: []string{"opendaoc", "mythic-con"}, Scenarios: []string{"A yellow creature at level 1", "An orange creature at level 1"}},
	)
}
