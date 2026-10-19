package rules

import "testing"

const cst = "rules-creature-casting"

func TestCreatureCasting(t *testing.T) {
	dd := CastOption{Ready: true, Range: 1500}
	vector(t, cst, "Cast, then melee", func(t *testing.T) {
		if i := ChooseCast([]CastOption{dd}, 1000, false); i != 0 {
			t.Errorf("a ready spell in range: %d, want it cast", i)
		}
		spent := dd
		spent.Ready = false
		if i := ChooseCast([]CastOption{spent}, 50, false); i != -1 {
			t.Errorf("within its recast: %d, want melee", i)
		}
	})
	vector(t, cst, "Out of range", func(t *testing.T) {
		if i := ChooseCast([]CastOption{dd}, 1600, false); i != -1 {
			t.Errorf("out of range: %d, want to close in", i)
		}
	})
	vector(t, cst, "A slow already on the target", func(t *testing.T) {
		slow := CastOption{Ready: true, Range: 1500, Already: true}
		if i := ChooseCast([]CastOption{slow}, 500, false); i != -1 {
			t.Errorf("a slow already on: %d, want a swing", i)
		}
	})
	vector(t, cst, "Healing a packmate", func(t *testing.T) {
		heal := CastOption{Heal: true, Ready: true, Range: 2000}
		if i := ChooseCast([]CastOption{dd, heal}, 500, true); i != 1 {
			t.Errorf("someone hurt: %d, want the heal first", i)
		}
		if i := ChooseCast([]CastOption{dd, heal}, 500, false); i != 0 {
			t.Errorf("nobody hurt: %d, want the hostile spell", i)
		}
		if !Hurt(74, 100) || Hurt(75, 100) {
			t.Error("hurt means below 75%")
		}
	})
	vector(t, cst, "A 3-second cast", func(t *testing.T) {
		if e := CastEnds(10, 3); e != 13 {
			t.Errorf("ends at %v, want 13", e)
		}
		if e := CastEnds(10, 0); e != 10 {
			t.Errorf("an instant spell ends at %v", e)
		}
		if !CastLands(true, 1000, 1500) {
			t.Error("a target in range should take it")
		}
	})
	vector(t, cst, "The target ran away", func(t *testing.T) {
		if CastLands(true, 1600, 1500) || CastLands(false, 10, 1500) {
			t.Error("a target out of range or dead should not take it")
		}
		if CastRetry != 2 {
			t.Error("a failed spell is ready again after 2 seconds")
		}
	})
	vector(t, cst, "Interrupt", func(t *testing.T) {
		if !Interrupts(false) {
			t.Error("an attack should interrupt")
		}
	})
	vector(t, cst, "Uninterruptible", func(t *testing.T) {
		if Interrupts(true) {
			t.Error("an uninterruptible spell should go on")
		}
	})
}
