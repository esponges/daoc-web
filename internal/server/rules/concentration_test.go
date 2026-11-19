package rules

import "testing"

const conc = "rules-concentration"

func TestConcentration(t *testing.T) {
	healer := Class{Name: "Healer", Casts: true, Acuity: PIE}
	warrior := Class{Name: "Warrior"}
	vector(t, conc, "A level 20 Healer", func(t *testing.T) {
		// 20 x 4 x 2.2 = 176; a piety of 60 adds 5 percent (88 / 2 ... (60-50)/2); halved.
		if got := MaxConcentration(healer, 20, 60); got != 92 {
			t.Errorf("level 20, piety 60: %d, want 92", got)
		}
		if got := MaxConcentration(healer, 20, 50); got != 88 {
			t.Errorf("level 20, piety 50: %d, want 88", got)
		}
		if got := MaxConcentration(healer, 20, 30); got != 70 {
			t.Errorf("level 20, piety 30 (20 percent less): %d, want 70", got)
		}
		if got := MaxConcentration(healer, 50, 100); got != 275 {
			t.Errorf("level 50, piety 100: %d, want 275", got)
		}
		if MaxConcentration(healer, 30, 80) <= MaxConcentration(healer, 29, 80) {
			t.Error("the pool should grow with level")
		}
	})
	vector(t, conc, "A Warrior", func(t *testing.T) {
		if got := MaxConcentration(warrior, 50, 100); got != 0 {
			t.Errorf("a Warrior has a pool of %d", got)
		}
	})
	vector(t, conc, "Room in the pool", func(t *testing.T) {
		free := ConcentrationFree(26, 6) // 20 free
		if !CanHold(free, 6) {
			t.Error("a 6-point buff fits in 20 free")
		}
		if got := ConcentrationFree(26, 6+6); got != 14 {
			t.Errorf("after holding 6 more, %d free, want 14", got)
		}
	})
	vector(t, conc, "No room", func(t *testing.T) {
		if CanHold(4, 6) {
			t.Error("a 6-point buff does not fit in 4 free")
		}
		if got := ConcentrationShort(6); got != "This spell requires 6 concentration points to cast!" {
			t.Errorf("the refusal reads %q", got)
		}
		if ConcentrationFree(10, 12) != 0 {
			t.Error("free concentration never goes below 0")
		}
	})
	vector(t, conc, "Recasting on the same target", func(t *testing.T) {
		// The old buff's cost comes back first: with 4 free and 6 held by it, the
		// recast sees 10 free and fits.
		if !CanHold(4+6, 6) {
			t.Error("a recast should fit once the old buff's cost is returned")
		}
	})
	vector(t, conc, "The target runs off", func(t *testing.T) {
		if OutOfRange(5000) || !OutOfRange(5001) {
			t.Error("range is 5,000 units")
		}
		if Lapsed(100, 119.9) || !Lapsed(100, 120) {
			t.Error("the buff ends 20 seconds after the target left range")
		}
	})
	vector(t, conc, "The caster dies", func(t *testing.T) {
		// Death ends every buff the caster holds and returns their cost; the rule
		// is that nothing is held by a dead caster.
		if got := ConcentrationFree(92, 0); got != 92 {
			t.Errorf("a caster holding nothing has %d free of 92", got)
		}
	})
}
