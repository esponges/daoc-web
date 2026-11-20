package rules

import "testing"

func TestPetRules(t *testing.T) {
	const pc = "pets"
	vector(t, pc, "A Spiritmaster summons", func(t *testing.T) {
		if got := PetLevel(-88, 50, 10); got != 8 {
			t.Errorf("a level 10 caster's -88%% spirit fighter is level %d, want 8", got)
		}
		if got := PetLevel(-88, 50, 50); got != 44 {
			t.Errorf("at 50: %d, want 44", got)
		}
		if got := PetLevel(-75, 15, 40); got != 15 {
			t.Errorf("the spell's value caps the level: %d, want 15", got)
		}
		if got := PetLevel(20, 0, 5); got != 20 {
			t.Errorf("a positive damage is a level: %d", got)
		}
		if got := PetLevel(-75, 15, 1); got != 1 {
			t.Errorf("never under 1: %d", got)
		}
	})
	vector(t, pc, "Charm an animal", func(t *testing.T) {
		if !CharmKindOK(2, BodyAnimal) || CharmKindOK(2, BodyHumanoid) {
			t.Error("kind 2 takes animals only")
		}
		if !CharmLevelOK(16, 20, 15) || !CharmLevelOK(16, 20, 16) {
			t.Error("a creature at or under the spell's limit is taken")
		}
		if got := CharmResistChance(0.85); got < 0.149 || got > 0.151 {
			t.Errorf("resist chance %v, want 0.15", got)
		}
	})
	vector(t, pc, "Too high", func(t *testing.T) {
		if CharmLevelOK(16, 20, 17) {
			t.Error("above the spell's limit")
		}
		if CharmLevelOK(30, 20, 21) {
			t.Error("above the caster's own level")
		}
	})
	vector(t, pc, "Wrong kind", func(t *testing.T) {
		if CharmKindOK(3, BodyAnimal) || !CharmKindOK(3, BodyInsect) {
			t.Error("kind 3 takes insects only")
		}
		if !CharmKindOK(4, BodyAnimal) || !CharmKindOK(4, BodyHumanoid) || CharmKindOK(4, BodyInsect) {
			t.Error("kind 4 takes humanoids and animals")
		}
		if !CharmKindOK(7, BodyUndead) || CharmKindOK(6, BodyUndead) || !CharmKindOK(0, BodyUndead) {
			t.Error("undead only from kind 7, and anything from kind 0")
		}
		if CharmKindOK(99, BodyAnimal) {
			t.Error("an unknown kind takes nothing")
		}
	})
	vector(t, pc, "Full army", func(t *testing.T) {
		if MinionLimit(3) != 3 || MinionLimit(-1) != 0 {
			t.Error("the commander's limit is its pulse column")
		}
		if !ValidStance(StanceDefensive) || ValidStance("angry") {
			t.Error("stances")
		}
	})
}
