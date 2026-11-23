package rules

import "testing"

func TestSeatedRegeneration(t *testing.T) {
	const c = "rules-resting"
	vector(t, c, "Level 20 seated", func(t *testing.T) {
		tick := float64(HealthRegen(20, false)) // 7 every 3 s: 10 ticks in 30 s
		near(t, "standing", 10*SeatedRegen(tick, false, false), 70, 1e-9)
		near(t, "seated", 10*SeatedRegen(tick, true, false), 140, 1e-9)
		near(t, "seated in combat", SeatedRegen(tick, true, true), tick, 1e-9)
	})
	vector(t, c, "Seated caster", func(t *testing.T) {
		near(t, "power tick", SeatedRegen(PowerRegen(50), true, false), 2*(5+50/2.75), 1e-9)
	})
	vector(t, c, "Seated endurance", func(t *testing.T) {
		near(t, "a second", 40+SeatedRegen(EnduranceRegenOut, true, false), 52, 1e-9)
		if 40+5*SeatedRegen(EnduranceRegenOut, true, false) < MaxEndurance {
			t.Error("a seated character is not back to full in five seconds")
		}
	})
}

func TestFalling(t *testing.T) {
	const c = "rules-falling"
	vector(t, c, "A short drop", func(t *testing.T) {
		if d := FallDamage(SafeFallHeight, 400, 0, false); d != 0 {
			t.Errorf("a safe drop cost %d", d)
		}
	})
	vector(t, c, "A long drop", func(t *testing.T) {
		if d := FallDamage(2*SafeFallHeight, 400, 0, false); d != 100 {
			t.Errorf("twice the safe height cost %d of 400, want 100 (25 percent)", d)
		}
	})
	vector(t, c, "A fall can kill", func(t *testing.T) {
		if d := FallDamage(SafeFallHeight+FallLethalSpan*2, 400, 0, false); d != 400 {
			t.Errorf("a lethal fall cost %d, want all 400", d)
		}
	})
	vector(t, c, "A Shadowblade with Stealth 30", func(t *testing.T) {
		if got := SafeFallLevel(30); got != 3 {
			t.Fatalf("Stealth 30 gives Safe Fall %d, want 3", got)
		}
		if SafeFallLevel(99) != MaxSafeFall || SafeFallLevel(-5) != 0 {
			t.Error("Safe Fall is not clamped to 0..5")
		}
		warrior, blade := FallDamage(2000, 400, 0, false), FallDamage(2000, 400, 3, false)
		if blade >= warrior {
			t.Errorf("Safe Fall 3 took %d, a Warrior %d", blade, warrior)
		}
	})
	vector(t, c, "Into a lake", func(t *testing.T) {
		if d := FallDamage(5000, 400, 0, true); d != 0 {
			t.Errorf("a fall into water cost %d", d)
		}
	})
}

func TestMovementNumbers(t *testing.T) {
	const c = "rules-movement"
	vector(t, c, "Sprint speed", func(t *testing.T) {
		near(t, "sprint", SprintSpeed(191), 238.75, 1e-9)
	})
	vector(t, c, "Standing sprint costs nothing", func(t *testing.T) {
		near(t, "standing ten seconds", SprintDrain(10, false), 0, 1e-9)
		near(t, "moving ten seconds", SprintDrain(10, true), 50, 1e-9)
	})
	vector(t, c, "Out of endurance", func(t *testing.T) {
		if !SprintContinues(1) || SprintContinues(0) {
			t.Error("sprint should go on with endurance and stop at zero")
		}
	})
	vector(t, c, "Swimming", func(t *testing.T) {
		near(t, "swim speed", SwimSpeed, 120, 1e-9)
	})
	vector(t, c, "Airtime", func(t *testing.T) {
		if n := JumpAirSteps(); n != 50 {
			t.Errorf("a jump lasts %d steps, want 50", n)
		}
	})
}
