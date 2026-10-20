package rules

import "testing"

const abl = "rules-abilities"

func TestAbilities(t *testing.T) {
	vector(t, abl, "Guard at Shields 5", func(t *testing.T) {
		g := []Grant{{Source: "Shields", Level: 5, Ability: AbGuard, AbilityLevel: 1}, {Source: "Shields", Level: 10, Ability: AbGuard, AbilityLevel: 2}}
		if got := Abilities(g, nil, 10, map[string]int{"Shields": 5})[AbGuard]; got != 1 {
			t.Errorf("Guard level %d at Shields 5, want 1", got)
		}
	})
	vector(t, abl, "Too early", func(t *testing.T) {
		g := []Grant{{Source: "WarriorCareer", Level: 7, Ability: AbProtect, AbilityLevel: 1}}
		if _, ok := Abilities(g, map[string]bool{"WarriorCareer": true}, 6, nil)[AbProtect]; ok {
			t.Error("a level 6 Warrior has Protect")
		}
	})
	vector(t, abl, "Not grouped", func(t *testing.T) {
		// Grouping is the world's to check; the range a use needs is the rule's.
		if AbilityRange != 2000 {
			t.Errorf("use range %v, want 2000", AbilityRange)
		}
	})
	vector(t, abl, "Guarded blow", func(t *testing.T) {
		if !GuardBlocks(150, 0.30, 0.20) {
			t.Error("the guard at 150 units with 30% block and a 0.20 roll should block")
		}
	})
	vector(t, abl, "Too far", func(t *testing.T) {
		if GuardBlocks(400, 0.30, 0.20) {
			t.Error("a guard 400 units away should not act")
		}
	})
	vector(t, abl, "Intercepted", func(t *testing.T) {
		if !InterceptTakes(100) || InterceptTakes(200) {
			t.Error("intercept should take a blow at 100 units and not at 200")
		}
	})
	vector(t, abl, "Protect II", func(t *testing.T) {
		if got := ProtectHate(2, 100); got != 20 {
			t.Errorf("hate %v, want 20", got)
		}
	})
	vector(t, abl, "Engaged", func(t *testing.T) {
		if !EngageBlocks(true, 50, 0.9) || EngageBlocks(true, 50, 0.96) || EngageBlocks(false, 50, 0.1) || EngageBlocks(true, 4, 0.1) {
			t.Error("engage should block 95% of frontal blows while there is endurance")
		}
	})
	vector(t, abl, "Already in combat", func(t *testing.T) {
		if EngageAllowed(2) || !EngageAllowed(5) {
			t.Error("engage should be refused 2s after a blow and allowed after 5s")
		}
	})
}
