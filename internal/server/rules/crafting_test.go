package rules

import "testing"

func TestCraftingRules(t *testing.T) {
	const c = "rules-crafting"
	vector(t, c, "Orange recipe", func(t *testing.T) {
		for _, w := range []struct{ skill, level, want int }{
			{500, 400, 100}, {500, 500, 100}, {490, 500, 90}, {450, 500, 50}, {440, 500, 0},
		} {
			if got := SuccessChance(w.skill, w.level); got != w.want {
				t.Errorf("skill %d on level %d: %d%%, want %d%%", w.skill, w.level, got, w.want)
			}
		}
		if SuccessChance(490, 500) >= SuccessChance(500, 500) {
			t.Error("a recipe above the skill must be harder than an equal one")
		}
	})
	vector(t, c, "Out of reach", func(t *testing.T) {
		if SuccessChance(100, 151) != 0 || SuccessChance(100, 150) != 50 {
			t.Error("fifty above is the limit")
		}
	})
	vector(t, c, "Distribution", func(t *testing.T) {
		// Dice that always pass give a masterpiece; always fail give 94.
		if q := Quality(500, 500, func() float64 { return 0 }); q != 100 || !Masterpiece(q) {
			t.Errorf("all passing: %d", q)
		}
		if q := Quality(500, 500, func() float64 { return 0.99 }); q != 94 || Masterpiece(q) {
			t.Errorf("all failing: %d", q)
		}
		// Three passes then a failure is 97.
		n := 0
		seq := func() float64 {
			n++
			if n <= 3 {
				return 0.1
			}
			return 0.9
		}
		if q := Quality(500, 500, seq); q != 97 {
			t.Errorf("three passes: %d", q)
		}
		// A deterministic sweep of the dice: the share reaching each step is p^k.
		count := map[int]int{}
		const runs = 10000
		for i := 0; i < runs; i++ {
			// Each run draws from a fixed low-discrepancy sequence.
			a := float64(i) / runs
			step := 0
			seq := func() float64 {
				step++
				return frac(a * float64(step*step+3))
			}
			count[Quality(500, 500, seq)]++
		}
		for q := 94; q <= 100; q++ {
			if q < 100 && count[q] == 0 {
				t.Errorf("no item of quality %d in %d crafts", q, runs)
			}
		}
		if count[100] > runs/10 {
			t.Errorf("masterpieces are not rare: %d of %d", count[100], runs)
		}
		// More skill moves the odds up, within 20% and 80%.
		hi, lo := 0, 0
		for i := 0; i < runs; i++ {
			a := float64(i) / runs
			hi += Quality(900, 500, func() float64 { return a }) - 94
			lo += Quality(100, 500, func() float64 { return a }) - 94
		}
		if hi <= lo {
			t.Errorf("a skilled crafter should do better: %d vs %d", hi, lo)
		}
	})
	vector(t, c, "Time", func(t *testing.T) {
		if got := CraftSeconds(500, 20); got != 14 {
			t.Errorf("level 500 with 20 materials: %ds, want 14", got)
		}
		if CraftSeconds(0, 1) != 3 {
			t.Errorf("limits: %d %d", CraftSeconds(0, 1), CraftSeconds(1100, 200))
		}
		if CraftSeconds(5000, 5000) != 60 {
			t.Error("a minute is the most")
		}
	})
	vector(t, c, "Gain", func(t *testing.T) {
		for _, w := range []struct{ skill, level, want int }{
			{100, 200, 100}, {100, 100, 100}, {150, 100, 50}, {200, 100, 0}, {300, 100, 0},
		} {
			if got := GainChance(w.skill, w.level); got != w.want {
				t.Errorf("skill %d on level %d: %d%%, want %d%%", w.skill, w.level, got, w.want)
			}
		}
		if FailureLoss(1) != 0 || FailureLoss(2) != 1 || FailureLoss(7) != 3 || FailureLoss(10) != 5 {
			t.Error("a failure loses half of a stack, rounded down")
		}
	})
	vector(t, c, "Secondary cap", func(t *testing.T) {
		if !MayRaise(SkillWood, 99, false, 100) || MayRaise(SkillWood, 100, false, 100) {
			t.Error("a secondary stops at the primary's skill")
		}
		if !MayRaise(SkillWeapon, 1099, true, 1099) || MayRaise(SkillWeapon, MaxSkill, true, MaxSkill) {
			t.Error("a primary stops at 1100")
		}
		if len(Professions) != 6 || Professions[SkillWeapon].Name != "Weaponcrafting" {
			t.Error("six professions, one per master")
		}
	})
}

func frac(x float64) float64 { return x - float64(int64(x)) }
