package rules

import (
	"math"
	"testing"
)

const xp = "rules-experience"
const dt = "rules-death"

func TestExperience(t *testing.T) {
	vector(t, xp, "The thresholds", func(t *testing.T) {
		for l := 2; l <= MaxLevel; l++ {
			if LevelFor(XPToReach(l)) != l || LevelFor(XPToReach(l)-1) != l-1 {
				t.Errorf("level %d threshold %d does not reach it exactly", l, XPToReach(l))
			}
		}
		if XPToReach(50) != 169999999950 || XPToReach(2) != 50 {
			t.Error("table ends wrong")
		}
	})
	vector(t, xp, "Level 10 to 11", func(t *testing.T) {
		if d := XPToReach(11) - XPToReach(10); d != 256000 {
			t.Errorf("span %d, want 256000", d)
		}
	})
	vector(t, xp, "Values", func(t *testing.T) {
		for l, want := range map[int]int64{10: 5120, 20: 163840, 35: 5242880, 50: 41943040} {
			if v := CreatureXP(l); v != want {
				t.Errorf("level %d worth %d, want %d", l, v, want)
			}
		}
	})
	vector(t, xp, "The level 22 quirk", func(t *testing.T) {
		curve := int64(163840 * math.Pow(2, 2.0/3))
		if CreatureXP(22) != 231705 || curve != 260079 {
			t.Errorf("level 22 worth %d (curve %d)", CreatureXP(22), curve)
		}
	})
	vector(t, xp, "An even solo kill", func(t *testing.T) {
		if v := KillXP(10, 10, 1); v != 5120 {
			t.Errorf("%d, want 5120", v)
		}
	})
	vector(t, xp, "Capped", func(t *testing.T) {
		if v := KillXP(12, 10, 1); v != 6400 {
			t.Errorf("%d, want 6400", v)
		}
	})
	vector(t, xp, "Shared", func(t *testing.T) {
		if a, b := KillXP(10, 10, 0.75), KillXP(10, 10, 0.25); a != 3840 || b != 1280 {
			t.Errorf("%d and %d, want 3840 and 1280", a, b)
		}
	})
	vector(t, xp, "A grey kill", func(t *testing.T) {
		if !Grey(20, 9) || Grey(20, 12) {
			t.Error("a level 9 creature should be grey to level 20, and level 12 not")
		}
	})
	vector(t, xp, "The bands", func(t *testing.T) {
		// Level 20: band 3.
		for lvl, want := range map[int]int{11: ConGrey, 14: ConGreen, 17: ConBlue, 20: ConYellow, 23: ConOrange, 26: ConRed, 27: ConPurple} {
			if c := Con(20, lvl); c != want {
				t.Errorf("level %d to 20: %d, want %d", lvl, c, want)
			}
		}
	})
	vector(t, xp, "Level 40", func(t *testing.T) {
		for lvl, want := range map[int]int{25: ConGrey, 30: ConGreen, 35: ConBlue, 40: ConYellow, 45: ConOrange} {
			if c := Con(40, lvl); c != want {
				t.Errorf("level %d to 40: %d, want %d", lvl, c, want)
			}
		}
	})
	vector(t, xp, "Level 50", func(t *testing.T) {
		if Con(50, 61) != ConPurple || Con(50, 60) != ConRed {
			t.Error("level 61 should be purple and 60 red to level 50")
		}
	})
}

func TestDeath(t *testing.T) {
	vector(t, dt, "First death at level 10", func(t *testing.T) {
		if p := LossPermille(10, 0); p != 13 {
			t.Errorf("p %d, want 13", p)
		}
		if l := DeathLoss(10, 300000, 0); l != 3328 {
			t.Errorf("loss %d, want 3328", l)
		}
	})
	vector(t, dt, "Third death at level 10", func(t *testing.T) {
		if l := DeathLoss(10, 300000, 2); l != 10240 {
			t.Errorf("loss %d, want 10240", l)
		}
	})
	vector(t, dt, "Below level 6", func(t *testing.T) {
		if DeathLoss(5, 3000, 0) != 0 {
			t.Error("a level 5 death cost experience")
		}
	})
	vector(t, dt, "At the start of a level", func(t *testing.T) {
		if l := DeathLoss(10, XPToReach(10)+100, 2); l != 100 {
			t.Errorf("loss %d, want 100 (down to the level's start)", l)
		}
	})
	vector(t, dt, "Release after the first death", func(t *testing.T) {
		first := DeathLoss(10, 300000, 0)
		if again := DeathLoss(10, 300000-first, 0); again != 3328 {
			t.Errorf("release loss %d, want 3328", again)
		}
	})
	vector(t, dt, "Two deaths", func(t *testing.T) {
		if ConLoss(1)+ConLoss(2) != 3 || ConLoss(7) != 3 {
			t.Error("constitution loss by death count is wrong")
		}
	})
	vector(t, dt, "Restore at level 12", func(t *testing.T) {
		if n, c := Restore(12, 3, 2000); n != 3 || c != 1056 {
			t.Errorf("restored %d for %d, want 3 for 1056", n, c)
		}
	})
	vector(t, dt, "Not enough money", func(t *testing.T) {
		if n, c := Restore(12, 3, 500); n != 1 || c != 352 {
			t.Errorf("restored %d for %d, want 1 for 352", n, c)
		}
	})
}
