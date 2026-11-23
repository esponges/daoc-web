package rules

import "testing"

func TestSavageSpells(t *testing.T) {
	const c = "rules-spell-effects"
	vector(t, c, "Health cost", func(t *testing.T) {
		// Fangs of Kelgor costs -5 (percent of health); 880 hit points.
		if got := HealthCost(-5, 880); got != 44 {
			t.Errorf("a 5 percent cost of 880 is %d, want 44", got)
		}
		if got := HealthCost(-3, 10); got != 1 {
			t.Errorf("the cost is at least 1, got %d", got)
		}
		if got := HealthCost(0, 880); got != 0 {
			t.Errorf("a spell with no power costs %d", got)
		}
		if !PaysHealth(SpellSavageParry) || PaysHealth(SpellHeal) || PaysHealth(SpellTaunt) {
			t.Error("only the Savage's own buffs and the endurance heal pay in health")
		}
	})
	vector(t, c, "Not enough health", func(t *testing.T) {
		if CanPayHealth(44, 44) {
			t.Error("a cost that would take the caster to nothing was allowed")
		}
		if !CanPayHealth(45, 44) {
			t.Error("a cost the caster can pay was refused")
		}
	})
	vector(t, c, "Combat speed", func(t *testing.T) {
		// Passion of Kelgor is 16: a 2.0 s weapon swings in 1.68 s.
		near(t, "swing", HastenedSwing(2.0, 16), 1.68, 1e-9)
		near(t, "dps", SavagePercent(25), 0.25, 1e-9)
		if ty, ok := SavageResistType(SpellSavageSlash); !ok || ty != DamageSlash {
			t.Errorf("slash buff guards type %d", ty)
		}
		if _, ok := SavageResistType(SpellSavageDPS); ok {
			t.Error("the DPS buff is a resist buff")
		}
		// A Savage's evade and parry buffs add to the chances, and the defence
		// loss takes a share of all three.
		d := &Combatant{Player: true, EvadeLevel: 1, ParrySpec: 1, Stats: Stats{DEX: 60, QUI: 60}}
		base := EvadeChance(d, true, 1)
		d.EvadeBonus = 0.17
		near(t, "evade", EvadeChance(d, true, 1)-base, 0.17, 1e-9)
		pbase := ParryChance(d, true, 1, false)
		d.ParryBonus = 0.09
		near(t, "parry", ParryChance(d, true, 1, false)-pbase, 0.09, 1e-9)
	})
	vector(t, "rules-spell-effects", "Endurance restored", func(t *testing.T) {
		near(t, "Kelgor's Gift", EnduranceHealed(20, 100), 20, 1e-9)
		near(t, "Kelgor's Reward", EnduranceHealed(60, 100), 60, 1e-9)
	})
	vector(t, "rules-spell-effects", "Pulling hate", func(t *testing.T) {
		near(t, "Rude Gesture", TauntHate(38), 380, 1e-9)
	})
}
