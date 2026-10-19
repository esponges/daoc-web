package rules

import "testing"

const se = "rules-spell-effects"

func TestSpellEffects(t *testing.T) {
	vector(t, se, "Even levels", func(t *testing.T) {
		near(t, "chance", SpellHitChance(10, 10), 0.85, 1e-9)
	})
	vector(t, se, "A higher caster", func(t *testing.T) {
		near(t, "chance", SpellHitChance(16, 10), 0.88, 1e-9)
	})
	vector(t, se, "Far below", func(t *testing.T) {
		near(t, "chance", SpellHitChance(1, 90), 0.50, 1e-9)
		near(t, "a far stronger caster is capped too", SpellHitChance(90, 1), 0.99, 1e-9)
	})
	vector(t, se, "A resisted fraction", func(t *testing.T) {
		if d := SpellDamage(100, 1, 10); d != 90 {
			t.Errorf("%d, want 90", d)
		}
		if d := SpellDamage(100, 1, 95); d != 30 {
			t.Errorf("a 95%% resist: %d, want 30 (capped at 70%%)", d)
		}
	})
	vector(t, se, "Middle of the range", func(t *testing.T) {
		v := SpellVariance(SpellDirectDamage, 0.5)
		if d := SpellDamage(40, v, 10); v != 1 || d != 36 {
			t.Errorf("variance %v damage %d, want 1 and 36", v, d)
		}
		if lo, hi := SpellVariance(SpellBolt, 0), SpellVariance(SpellBolt, 0.999999); lo != 0.75 || hi < 1.2499 {
			t.Errorf("range %v..%v", lo, hi)
		}
	})
	vector(t, se, "No variance", func(t *testing.T) {
		if d := SpellDamage(40, SpellVariance(SpellDDNoVariance, 0.03), 0); d != 40 {
			t.Errorf("%d, want 40", d)
		}
	})
	vector(t, se, "Blocked bolt", func(t *testing.T) {
		if !BoltBlocked(0.3, 0.1) || BoltBlocked(0.3, 0.5) || BoltBlocked(0, 0) {
			t.Error("a bolt is blocked when the roll is under the block chance, and never with none")
		}
	})
	vector(t, se, "Five ticks", func(t *testing.T) {
		if n, every := DoTTicks(20, 40); n != 5 || every != 4 {
			t.Errorf("%d ticks %v s apart, want 5 and 4", n, every)
		}
		if d := SpellDamage(10, SpellVariance(SpellDDNoVariance, 0.9), 0); d != 10 {
			t.Errorf("tick %d, want 10", d)
		}
	})
	vector(t, se, "A 40% slow", func(t *testing.T) {
		near(t, "slow", SlowFactor(40), 0.6, 1e-9)
		near(t, "strongest of two", SpeedFactor(false, SlowFactor(40), SlowFactor(25)), 0.6, 1e-9)
		near(t, "held", SpeedFactor(true, SlowFactor(40)), 0, 1e-9)
		near(t, "free", SpeedFactor(false), 1, 1e-9)
	})
	vector(t, se, "Drain", func(t *testing.T) {
		d := SpellDamage(40, SpellVariance(SpellLifedrain, 0.5), 0)
		if d != 40 || DrainHeal(d, 50) != 20 {
			t.Errorf("damage %d heal %d, want 40 and 20", d, DrainHeal(d, 50))
		}
	})
	vector(t, se, "A creature heals itself", func(t *testing.T) {
		if h := Healed(100, 300, 120); h != 220 {
			t.Errorf("%d, want 220", h)
		}
		if h := Healed(250, 300, 120); h != 300 {
			t.Errorf("over the top: %d, want 300", h)
		}
	})
	if !BrokenByDamage(SpellMesmerize) || BrokenByDamage(SpellStun) {
		t.Error("damage breaks a mesmerize and not a stun")
	}
}
