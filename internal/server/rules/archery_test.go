package rules

import (
	"math"
	"testing"
)

const ar = "rules-archery"

func hunter20() *Combatant {
	return &Combatant{Player: true, Level: 20, Stats: Stats{DEX: 80, QUI: 70}, WeaponSkillBase: 380, Spec: 20,
		Weapon: Weapon{DPS: 12.0, Speed: 4.5, ObjectType: ObjCompBow, TwoHanded: true, Quality: 1, Condition: 1, DamageType: DamageThrus}}
}

func TestArchery(t *testing.T) {
	vector(t, ar, "A full draw", func(t *testing.T) {
		// A 4.5 bow at quickness 50 draws for 4.5 s; each quickness point above 50 takes 0.2 percent off.
		near(t, "draw at 50", DrawTime(4.5, 50), 4.5, 1e-9)
		near(t, "draw at 80", DrawTime(4.5, 80), 4.5*(1-0.06), 1e-9)
		near(t, "never under 1.5", DrawTime(1.6, 250), 1.5, 1e-9)
	})
	vector(t, ar, "Range by bow", func(t *testing.T) {
		near(t, "longbow", BowRange(9, 100, 1), 1760, 1e-9)
		near(t, "recurve", BowRange(18, 100, 1), 1680, 1e-9)
		near(t, "composite", BowRange(ObjCompBow, 100, 1), 1600, 1e-9)
		near(t, "clout arrows", BowRange(ObjCompBow, 100, 0), 1600*0.85, 1e-9)
		near(t, "flight arrows", BowRange(ObjCompBow, 100, 3), 2000, 1e-9)
		near(t, "range modifier", BowRange(ObjCompBow, 110, 1), 1760, 1e-9)
		if ArrowRangeClass(60) != 3 || ArrowAccuracyClass(60) != 3 || ArrowRangeClass(16) != 0 || ArrowAccuracyClass(16) != 1 {
			t.Error("arrow classes read from the speed field")
		}
	})
	vector(t, ar, "A worked example", func(t *testing.T) {
		// A level 20 Hunter, composite bow 20, 12.0 DPS and 4.5 speed, against a level 20 creature.
		a, d := hunter20(), plainCreature(20)
		s := Swing{Attacker: a, Defender: d, InFront: true, Attackers: 1, Ranged: true}
		o := Resolve(s, hitRolls(0.5))
		if o.Result != Hit {
			t.Fatalf("result %v", o.Result)
		}
		// The same swing as melee with the same numbers: a shot is a swing from range.
		m := Resolve(Swing{Attacker: a, Defender: d, InFront: true, Attackers: 1}, hitRolls(0.5))
		if o.Damage != m.Damage || o.Damage != 52 {
			t.Errorf("a shot did %d, the same blow in melee %d; want 52 for both", o.Damage, m.Damage)
		}
		// Half the draw is half the damage.
		half := Resolve(Swing{Attacker: a, Defender: d, InFront: true, Attackers: 1, Ranged: true, Effectiveness: 0.5}, hitRolls(0.5))
		if half.Damage != o.Damage/2 {
			t.Errorf("half effectiveness did %d of %d", half.Damage, o.Damage)
		}
	})
	vector(t, ar, "Parry does not apply", func(t *testing.T) {
		d := plainCreature(20)
		d.ParryPct = 100
		melee := Resolve(Swing{Attacker: hunter20(), Defender: d, InFront: true, Attackers: 1}, &seq{0.99, 0, 0.99, 0, 0.99, 0.5, 0.99, 0})
		shot := Resolve(Swing{Attacker: hunter20(), Defender: d, InFront: true, Attackers: 1, Ranged: true}, &seq{0.99, 0, 0.99, 0, 0.99, 0.5, 0.99, 0})
		if melee.Result != Parried {
			t.Fatalf("the control blow was %v, want parried", melee.Result)
		}
		if shot.Result == Parried || shot.Chance.Parry != 0 {
			t.Errorf("a shot was %v with a parry chance of %v", shot.Result, shot.Chance.Parry)
		}
	})
	vector(t, ar, "Critical Shot", func(t *testing.T) {
		near(t, "rank 1 draw", CriticalDraw(4.0, 1), 8.0, 1e-9)
		near(t, "rank 5 draw", CriticalDraw(4.0, 5), 8.0-4*0.4, 1e-9)
		near(t, "even match", CriticalShotDamage(20, 20), 2.0, 1e-9)
		near(t, "higher target", CriticalShotDamage(20, 30), 2-0.75, 1e-9)
		near(t, "very high target", CriticalShotDamage(20, 60), 1.1, 1e-9)
		near(t, "lower target", CriticalShotDamage(30, 20), 2.0, 1e-9)
		if ShotEndurance(true, 0) != 10 || ShotEndurance(false, 0) != 5 {
			t.Error("shot endurance")
		}
	})
	vector(t, ar, "Rapid Fire", func(t *testing.T) {
		near(t, "half draw", RapidDraw(4.0), 2.0, 1e-9)
		near(t, "floor", RapidDraw(1.5), MinRapidDraw, 1e-9)
		near(t, "released at 75 percent", RapidFireDamage(3.0, 4.0), 0.75, 1e-9)
		near(t, "a full draw", RapidFireDamage(4.0, 4.0), 1, 1e-9)
		if ShotEndurance(false, 2) != 3 {
			t.Error("rank 2 costs half")
		}
	})
	vector(t, ar, "Penetrating Arrow", func(t *testing.T) {
		near(t, "rank 1", PenetratingEffect(1, false), 0.5, 1e-9)
		near(t, "rank 3", PenetratingEffect(3, false), 1.0, 1e-9)
		near(t, "no rank", PenetratingEffect(0, false), 0.25, 1e-9)
		near(t, "own bladeturn", PenetratingEffect(3, true), 0, 1e-9)
	})
	t.Run("arrow accuracy", func(t *testing.T) {
		if ArrowMissScale(0) <= 1 || ArrowMissScale(3) >= 1 || ArrowMissScale(1) != 1 || math.IsNaN(ArrowMissScale(2)) {
			t.Error("rough arrows miss more, footed fewer")
		}
	})
}

func TestLeftHand(t *testing.T) {
	vector(t, "rules-melee", "A Shadowblade with two axes", func(t *testing.T) {
		near(t, "left axe 20 damage", LeftAxeDamage(20), 0.625+0.068, 1e-9)
		near(t, "left axe 50 damage", LeftAxeDamage(50), 0.795, 1e-9)
		if LeftSwings(20) != 1 || LeftSwings(0) != 0 {
			t.Error("a Left Axe fighter swings the left hand every round, an untrained one not at all")
		}
	})
	vector(t, "rules-melee", "Hand to Hand blows", func(t *testing.T) {
		d, tr, q := HandToHandChances(50)
		near(t, "double", d, 0.25, 1e-9)
		near(t, "triple", tr, 0.125, 1e-9)
		near(t, "quad", q, 0.03125, 1e-9)
		for _, c := range []struct {
			roll float64
			want int
		}{{0.1, 1}, {0.3, 2}, {0.39, 3}, {0.5, 0}} {
			if got := HandToHandExtra(50, c.roll); got != c.want {
				t.Errorf("roll %v gave %d extra blows, want %d", c.roll, got, c.want)
			}
		}
		if HandToHandExtra(0, 0) != 0 {
			t.Error("an untrained fighter got extra blows")
		}
	})
}
