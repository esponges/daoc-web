package rules

import "testing"

const sp = "rules-specialization"

// warriorGrants are rows shaped like the class and spec ability tables, written
// for the test rather than copied from the data.
var warriorGrants = []Grant{
	{"VikingCareer", 1, AbMidArmor, 3},
	{"VikingCareer", 1, AbShield, 1},
	{"VikingCareer", 1, "Weaponry: Swords", 0},
	{"VikingCareer", 1, "Weaponry: Hammers", 0},
	{"VikingCareer", 1, "Weaponry: Axes", 0},
	{"WarriorCareer", 5, AbShield, 2},
	{"WarriorCareer", 7, "Protect", 1},
	{"WarriorCareer", 10, AbMidArmor, 4},
	{"WarriorCareer", 10, AbShield, 3},
	{"WarriorCareer", 15, AbEvade, 1},
	{"Shields", 5, "Guard", 1},
}

var warriorCareers = map[string]bool{"VikingCareer": true, "WarriorCareer": true}

func TestSpecialization(t *testing.T) {
	vector(t, sp, "Level 6 Warrior", func(t *testing.T) {
		if d := SpecPoints(warrior, 6, false) - SpecPoints(warrior, 5, false); d != 12 {
			t.Errorf("level 6 gives %d, want 12", d)
		}
	})
	vector(t, sp, "Half-level", func(t *testing.T) {
		if d := SpecPoints(warrior, 44, true) - SpecPoints(warrior, 44, false); d != 44 {
			t.Errorf("44.5 gives %d, want 44", d)
		}
	})
	vector(t, sp, "Total at 50", func(t *testing.T) {
		if p := SpecPoints(warrior, 50, false); p != 2979 {
			t.Errorf("total %d, want 2979", p)
		}
		if p := SpecPoints(warrior, 5, false); p != 14 {
			t.Errorf("levels 2-5 give %d, want 14", p)
		}
	})
	vector(t, sp, "Sword to 10", func(t *testing.T) {
		if c, err := Train(10, 1, 9, 1000); err != nil || c != 54 {
			t.Errorf("cost %d, %v; want 54", c, err)
		}
	})
	vector(t, sp, "Above level", func(t *testing.T) {
		if _, err := Train(10, 10, 1, 1000); err == nil {
			t.Error("trained above level")
		}
	})
	vector(t, sp, "Not enough points", func(t *testing.T) {
		if _, err := Train(10, 5, 1, 5); err == nil {
			t.Error("trained without the points")
		}
	})
	vector(t, sp, "A +3 sword at level 10", func(t *testing.T) {
		if s := CombatSpec(8, 5, 10); s != 11 {
			t.Errorf("combat sword %d, want 11", s)
		}
	})
	vector(t, sp, "Chain at level 10", func(t *testing.T) {
		at9 := Abilities(warriorGrants, warriorCareers, 9, nil)
		at10 := Abilities(warriorGrants, warriorCareers, 10, nil)
		if MayUse(at9, ObjChain, 0) == nil {
			t.Error("chain at level 9")
		}
		if err := MayUse(at10, ObjChain, 0); err != nil {
			t.Error(err)
		}
		if at10[AbShield] != 3 {
			t.Errorf("shield %d at 10, want 3", at10[AbShield])
		}
	})
	vector(t, sp, "Evade at level 15", func(t *testing.T) {
		if Abilities(warriorGrants, warriorCareers, 14, nil)[AbEvade] != 0 || Abilities(warriorGrants, warriorCareers, 15, nil)[AbEvade] != 1 {
			t.Error("evade is not first known at 15")
		}
	})
	vector(t, sp, "An ability not yet implemented", func(t *testing.T) {
		ab := Abilities(warriorGrants, warriorCareers, 7, map[string]int{"Shields": 1})
		if ab["Protect"] != 1 || ActiveAbility("Protect") {
			t.Errorf("protect %d, active %v", ab["Protect"], ActiveAbility("Protect"))
		}
		if ab["Guard"] != 0 {
			t.Error("guard granted below shields 5")
		}
	})
	vector(t, sp, "An existing level 20 Warrior", func(t *testing.T) {
		if p := SpecPoints(warrior, 20, false); p != 404 {
			t.Errorf("points %d, want 404", p)
		}
	})
}

func TestMayUse(t *testing.T) {
	ab := Abilities(warriorGrants, warriorCareers, 3, nil)
	if MayUse(ab, ObjShield, 2) == nil {
		t.Error("medium shield at level 3")
	}
	if err := MayUse(ab, ObjShield, 1); err != nil {
		t.Error(err)
	}
	if err := MayUse(ab, ObjAxe, 0); err != nil {
		t.Error(err)
	}
	if MayUse(ab, ObjSpear, 0) == nil {
		t.Error("spear without training")
	}
}
