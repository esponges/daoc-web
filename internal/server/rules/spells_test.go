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
	vector(t, se, "Centre and edge", func(t *testing.T) {
		// A radius-300 spell hits at the centre and at the edge, not beyond, and
		// the damage does not depend on where inside the circle the creature stands.
		if !AreaHits(0, 300) || !AreaHits(300, 300) || AreaHits(300.5, 300) {
			t.Error("the radius does not bound the area")
		}
		if AreaHits(0, 0) {
			t.Error("a spell with no radius is an area")
		}
		centre := SpellDamage(40, SpellVariance(SpellDirectDamage, 0.5), 10)
		edge := SpellDamage(40, SpellVariance(SpellDirectDamage, 0.5), 10)
		if centre != edge || centre != 36 {
			t.Errorf("centre %d edge %d, want 36 each", centre, edge)
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
	vector(t, se, "A fully specced caster", func(t *testing.T) {
		lo, hi := PlayerVarianceRange(20, 20, 20)
		near(t, "low end", lo, 1.2, 1e-9)
		near(t, "high end", hi, 1.25, 1e-9)
		if d := PlayerSpellDamage(40, 100, 20, 20, 20, 0.5, 0); d != 53 {
			t.Errorf("a 40-damage spell does %d, want 53", d)
		}
	})
	vector(t, se, "A low spec", func(t *testing.T) {
		if d := PlayerSpellDamage(40, 100, 5, 20, 20, 0, 0); d != 19 {
			t.Errorf("the worst roll does %d, want 19", d)
		}
	})
	vector(t, se, "The cap", func(t *testing.T) {
		if d := PlayerSpellDamage(10, 300, 50, 50, 1, 0.5, 0); d != 30 {
			t.Errorf("a 10-damage spell does %d, want the cap 30", d)
		}
	})
	vector(t, se, "A player heal at spec", func(t *testing.T) {
		lo, hi := PlayerHealRange(120, 20, 20)
		if lo != 144 || hi != 150 {
			t.Errorf("range %d to %d, want 144 to 150", lo, hi)
		}
		if h := PlayerHeal(120, 20, 20, 0.999); h != 150 {
			t.Errorf("the best roll heals %d, want 150", h)
		}
	})
	vector(t, se, "A player heal below spec", func(t *testing.T) {
		lo, hi := PlayerHealRange(120, 5, 20)
		if lo != 54 || hi != 150 {
			t.Errorf("range %d to %d, want 54 to 150", lo, hi)
		}
	})
	vector(t, se, "A percent heal", func(t *testing.T) {
		if h := PercentHeal(-40, 720); h != 288 {
			t.Errorf("a -40 heal from a caster of 720 restores %d, want 288", h)
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

func TestSpellScalingOverTime(t *testing.T) {
	const se = "rules-spell-effects"
	vector(t, se, "A player's tick at spec", func(t *testing.T) {
		lo, hi := PlayerDoTRange(20, 20)
		near(t, "low end", lo, 1.08, 1e-9)
		near(t, "high end", hi, 1.13, 1e-9)
		if d := PlayerDoTDamage(10, 100, 20, 20, 0.5, 0); d != 12 {
			t.Errorf("tick %d, want 12", d)
		}
	})
	vector(t, se, "A low spec tick", func(t *testing.T) {
		lo, _ := PlayerDoTRange(1, 20)
		near(t, "low end", lo, 0.13, 1e-9)
		if d := PlayerDoTDamage(10, 100, 1, 20, 0, 0); d != 1 {
			t.Errorf("tick %d, want 1", d)
		}
	})
	vector(t, se, "The tick cap", func(t *testing.T) {
		if d := PlayerDoTDamage(5, 600, 50, 50, 1, 0); d != 15 {
			t.Errorf("tick %d, want 15", d)
		}
	})
	vector(t, se, "A slow weapon", func(t *testing.T) {
		if d := DamageAdd(4.5, 20, 3.0, 40, 0.5, 0); d != 10 {
			t.Errorf("add %d, want 10", d)
		}
	})
	vector(t, se, "The damage add cap", func(t *testing.T) {
		near(t, "cap at level 3", DamageAddCap(3), 1.47, 1e-9)
		if d := DamageAdd(9, 3, 3.0, 40, 1, 0); d != 4 {
			t.Errorf("add %d, want 4", d)
		}
	})
	vector(t, se, "A percent add", func(t *testing.T) {
		if d := DamageAdd(-20, 20, 3.0, 50, 0.5, 0); d != 10 {
			t.Errorf("add %d, want 10", d)
		}
	})
	vector(t, "player-spellcasting", "A buff is capped by level", func(t *testing.T) {
		if c := BuffStatCap(10); c != 12 {
			t.Errorf("stat cap %d, want 12", c)
		}
	})
	vector(t, "player-spellcasting", "A buff does not scale with the caster", func(t *testing.T) {
		if b := BuffedStats(SpellStrengthBuff, 20); b[STR] != 20 || len(b) != 1 {
			t.Errorf("buff %v, want strength 20", b)
		}
	})
	vector(t, "player-spellcasting", "A big armour buff", func(t *testing.T) {
		if c := BuffAFCap(10); c != 20 {
			t.Errorf("armour cap %d, want 20", c)
		}
	})
}
