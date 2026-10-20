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
		// A 30% block and Guard II make a 40% guard chance: a 0.35 roll blocks.
		if c := GuardChance(0.30, 2); c < 0.399 || c > 0.401 || !GuardBlocks(150, c, 0.35) {
			t.Errorf("guard chance %v at 150 units with a 0.35 roll, want a block at 40%%", c)
		}
	})
	vector(t, abl, "Too far", func(t *testing.T) {
		if GuardBlocks(400, 0.40, 0.20) {
			t.Error("a guard 400 units away should not act")
		}
	})
	vector(t, abl, "Intercepted", func(t *testing.T) {
		if !InterceptTakes(100, 0.49) || InterceptTakes(200, 0.1) {
			t.Error("intercept should take a blow at 100 units on a 0.49 roll and not at 200")
		}
	})
	vector(t, abl, "Intercept fails its roll", func(t *testing.T) {
		if InterceptTakes(100, 0.5) {
			t.Error("intercept took a blow on a 0.5 roll")
		}
	})
	vector(t, abl, "Protect II", func(t *testing.T) {
		if got := ProtectShift(2, 100); got != 20 {
			t.Errorf("shift %v, want 20 moved to the protector", got)
		}
	})
	vector(t, abl, "Engaged", func(t *testing.T) {
		d := &Combatant{Player: true, Stats: Stats{DEX: 60}, ShieldSpec: 10, Shield: &Shield{Size: 2, Quality: 1, Condition: 1}}
		base := BlockChance(d, true)
		d.EngageBlock = EngageBonus
		if got := BlockChance(d, true) - base; got < 0.649 || got > 0.651 {
			t.Errorf("engage raised block by %v, want 0.65", got)
		}
	})
	vector(t, abl, "Already in combat", func(t *testing.T) {
		if EngageAllowed(2) || !EngageAllowed(5) {
			t.Error("engage should be refused 2s after an enemy attacked the creature and allowed after 5s")
		}
	})
}
