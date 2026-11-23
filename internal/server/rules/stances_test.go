package rules

import "testing"

func TestStances(t *testing.T) {
	const c = "rules-abilities"
	vector(t, c, "Berserk raises damage", func(t *testing.T) {
		near(t, "level 1", BerserkDamage(1), 0.05, 1e-9)
		near(t, "level 4", BerserkDamage(4), 0.20, 1e-9)
		near(t, "past the top", BerserkDamage(9), 0.20, 1e-9)
		// Through Resolve: the same blow lands 20 percent harder under Berserk 4, and
		// the defender's chances fall by a quarter.
		a, d := level10Troll(), plainCreature(10)
		plain := Resolve(Swing{Attacker: a, Defender: d, InFront: true, Attackers: 1}, hitRolls(0.5))
		a.DamageBonus = BerserkDamage(4)
		strong := Resolve(Swing{Attacker: a, Defender: d, InFront: true, Attackers: 1}, hitRolls(0.5))
		near(t, "damage", float64(strong.Damage), float64(plain.Damage)*1.2, 1)
		p := defender(60, 50)
		p.EvadeLevel, p.ParrySpec = 3, 20
		base := Resolve(Swing{Attacker: level10Troll(), Defender: p, InFront: true, Attackers: 1}, hitRolls(0.5))
		p.DefencePenalty = BerserkDefenceLoss
		loss := Resolve(Swing{Attacker: level10Troll(), Defender: p, InFront: true, Attackers: 1}, hitRolls(0.5))
		near(t, "evade", loss.Chance.Evade, base.Chance.Evade*0.75, 1e-9)
		near(t, "parry", loss.Chance.Parry, base.Chance.Parry*0.75, 1e-9)
	})
	vector(t, c, "Reuse", func(t *testing.T) {
		near(t, "ten seconds in", StanceWait(10, BerserkReuse), 290, 1e-9)
		near(t, "ready", StanceWait(400, BerserkReuse), 0, 1e-9)
	})
	vector(t, c, "Charge in combat", func(t *testing.T) {
		// A slow of 0.5 would hold a body to half speed; Charge puts a run at 1.25.
		near(t, "factor", ChargeFactor, 1.25, 1e-9)
		near(t, "run speed", 191*ChargeFactor, 238.75, 1e-9)
	})
	vector(t, c, "Instant next spell", func(t *testing.T) {
		near(t, "cast ends now", CastEnds(100, 0), 100, 1e-9)
	})
	vector(t, c, "Creatures look away", func(t *testing.T) {
		if !DistractionTurns(300) || DistractionTurns(600) {
			t.Error("the radius is 512")
		}
	})
	vector(t, c, "Warning", func(t *testing.T) {
		if DangerSenseRange <= 0 || DangerSenseQuiet <= 0 {
			t.Error("Danger Sense has no range or no quiet time")
		}
	})
	vector(t, "ground-target", "Set in range", func(t *testing.T) {
		if !GroundTargetOK(1500) {
			t.Error("1500 units is in range")
		}
	})
	vector(t, "ground-target", "Out of range", func(t *testing.T) {
		if GroundTargetOK(2500) {
			t.Error("2500 units is out of range")
		}
	})
}
