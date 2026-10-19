package rules

import "testing"

const ml = "rules-melee"

// seq is a fixed sequence of rolls; past its end every roll is 0.99, which
// fails any chance below 99%.
type seq []float64

func (s *seq) Float64() float64 {
	if len(*s) == 0 {
		return 0.99
	}
	v := (*s)[0]
	*s = (*s)[1:]
	return v
}

// hitRolls lands a blow at a variance roll v: no defence, torso, no miss, no crit.
func hitRolls(v float64) *seq { return &seq{0.99, 0.99, 0.99, 0, 0.99, v, 0.99, 0} }

func sword() Weapon {
	return Weapon{DPS: 4.5, Speed: 3.0, ObjectType: ObjSword, Quality: 1, Condition: 1, DamageType: DamageSlash}
}

// level10Troll is the rule book's level 10 Troll Warrior with sword 10.
func level10Troll() *Combatant {
	return &Combatant{Player: true, Level: 10, Stats: Stats{115, 83, 47, 35}, WeaponSkillBase: 460, Weapon: sword(), Spec: 10}
}

func plainCreature(level int) *Combatant {
	c := Creature(level, 0, 0, 0, false, DamageCrush, nil)
	return &c
}

func defender(dex, qui int) *Combatant {
	return &Combatant{Player: true, Level: 20, Stats: Stats{DEX: dex, QUI: qui}, Weapon: sword()}
}

func TestMelee(t *testing.T) {
	vector(t, ml, "A parried swing", func(t *testing.T) {
		d := defender(60, 50)
		d.ParrySpec = 20
		o := Resolve(Swing{Attacker: plainCreature(20), Defender: d, InFront: true, Attackers: 1}, &seq{0.99, 0})
		if o.Result != Parried {
			t.Errorf("result %v, want parry", o.Result)
		}
	})
	vector(t, ml, "Evade I", func(t *testing.T) {
		d := defender(60, 35)
		d.EvadeLevel = 1
		near(t, "evade", EvadeChance(d, true, 1), 0.049, 1e-9)
	})
	vector(t, ml, "From behind", func(t *testing.T) {
		d := defender(60, 35)
		d.EvadeLevel = 1
		if EvadeChance(d, false, 1) != 0 {
			t.Error("evaded from behind")
		}
	})
	vector(t, ml, "Parry 20", func(t *testing.T) {
		d := defender(60, 50)
		d.ParrySpec = 20
		near(t, "parry", ParryChance(d, true, 1, false), 0.15, 1e-9)
	})
	vector(t, ml, "Two attackers", func(t *testing.T) {
		d := defender(60, 50)
		d.ParrySpec = 20
		near(t, "parry", ParryChance(d, true, 2, false), 0.075, 1e-9)
		near(t, "parry against two-handed", ParryChance(d, true, 1, true), 0.075, 1e-9)
	})
	vector(t, ml, "Untrained", func(t *testing.T) {
		d := defender(45, 50)
		d.ParrySpec = 1
		near(t, "parry", ParryChance(d, true, 1, false), 0.048, 1e-9)
	})
	vector(t, ml, "Shields 1", func(t *testing.T) {
		d := defender(60, 50)
		d.ShieldSpec, d.Shield = 1, &Shield{Size: 1, Quality: 1, Condition: 1}
		near(t, "block", BlockChance(d, true), 0.055, 1e-9)
	})
	vector(t, ml, "Worn shield", func(t *testing.T) {
		d := defender(60, 50)
		d.ShieldSpec, d.Shield = 1, &Shield{Size: 1, Quality: 0.9, Condition: 0.8}
		near(t, "block", BlockChance(d, true), 0.0396, 1e-9)
	})
	vector(t, ml, "Even fight", func(t *testing.T) {
		near(t, "miss", MissChance(Swing{Attacker: level10Troll(), Defender: plainCreature(10)}, LocNone), 0.18, 1e-9)
	})
	vector(t, ml, "Higher creature", func(t *testing.T) {
		near(t, "miss", MissChance(Swing{Attacker: level10Troll(), Defender: plainCreature(13)}, LocNone), 0.22, 1e-9)
	})
	vector(t, ml, "A style with a to-hit bonus", func(t *testing.T) {
		s := Swing{Attacker: level10Troll(), Defender: plainCreature(10), Style: &Style{BonusToHit: 5, Location: LocNone}}
		near(t, "miss", MissChance(s, LocNone), 0.13, 1e-9)
	})
	vector(t, ml, "An enchanted weapon", func(t *testing.T) {
		a := level10Troll()
		a.Level, a.Weapon.Bonus = 20, 10
		near(t, "miss", MissChance(Swing{Attacker: a, Defender: plainCreature(20)}, LocNone), 0.162, 1e-9)
	})
	vector(t, ml, "Bare legs", func(t *testing.T) {
		d := defender(60, 50)
		d.Armour[LocTorso] = &Piece{AF: 20, Absorb: 0.19, Quality: 1, Condition: 1}
		af, abs := LocationArmour(d, LocLegs)
		near(t, "af", af, 8, 1e-9)
		near(t, "absorb", abs, 0, 0)
	})
	vector(t, ml, "Over-level weapon", func(t *testing.T) {
		base, cap := BaseDamage(level10Troll())
		near(t, "base", base, 12.978, 1e-3)
		near(t, "cap", cap, 38.934, 1e-3)
	})
	hit := func(t *testing.T, a, d *Combatant, v float64) Outcome {
		t.Helper()
		o := Resolve(Swing{Attacker: a, Defender: d, InFront: true, Attackers: 1}, hitRolls(v))
		if o.Result != Hit {
			t.Fatalf("result %v, want a hit", o.Result)
		}
		return o
	}
	vector(t, ml, "Mid roll on an even creature", func(t *testing.T) {
		a, d := level10Troll(), plainCreature(10)
		near(t, "variance", Variance(a, 10, 0.5), 1.409, 1e-3)
		af, abs := LocationArmour(d, LocNone)
		near(t, "armour term", ArmourTerm(af, abs), 27.696, 1e-3)
		if o := hit(t, a, d, 0.5); o.Damage != 23 {
			t.Errorf("damage %d, want 23", o.Damage)
		}
	})
	vector(t, ml, "The spread", func(t *testing.T) {
		if o := hit(t, level10Troll(), plainCreature(10), 0); o.Damage != 19 {
			t.Errorf("low roll %d, want 19", o.Damage)
		}
		if o := hit(t, level10Troll(), plainCreature(10), 1); o.Damage != 28 {
			t.Errorf("high roll %d, want 28", o.Damage)
		}
	})
	vector(t, ml, "Untrained weapon", func(t *testing.T) {
		a := level10Troll()
		a.Spec = 1
		near(t, "variance", Variance(a, 10, 0.5), 1.0, 1e-9)
		if o := hit(t, a, plainCreature(10), 0.5); o.Damage != 16 {
			t.Errorf("damage %d, want 16", o.Damage)
		}
	})
	vector(t, ml, "Two-handed", func(t *testing.T) {
		a := &Combatant{Player: true, Level: 20, Stats: Stats{STR: 125}, WeaponSkillBase: 460, Spec: 20,
			Weapon: Weapon{DPS: 7.2, Speed: 5.0, ObjectType: ObjAxe, TwoHanded: true, Quality: 1, Condition: 1, DamageType: DamageSlash}}
		if o := hit(t, a, plainCreature(20), 0.5); o.Damage != 82 {
			t.Errorf("damage %d, want 82", o.Damage)
		}
	})
	vector(t, ml, "A big crit", func(t *testing.T) {
		o := Resolve(Swing{Attacker: level10Troll(), Defender: plainCreature(10), InFront: true, Attackers: 1},
			&seq{0.99, 0.99, 0.99, 0, 0.99, 0.5, 0, 1.0})
		if o.Critical != 23 || o.Damage != 46 {
			t.Errorf("critical %d, damage %d; want 23 and 46", o.Critical, o.Damage)
		}
	})
	vector(t, ml, "A styled hit", func(t *testing.T) {
		s := Swing{Attacker: level10Troll(), Defender: plainCreature(10), InFront: true, Attackers: 1,
			Style: &Style{GrowthRate: 0.5, Spec: 10, Location: LocNone}}
		raw, styled, fired := HitDamage(s, LocNone, 0.5)
		near(t, "unstyled", raw, 23.919, 1e-3)
		near(t, "style", styled, 9.215, 1e-3)
		if o := Resolve(s, hitRolls(0.5)); !fired || !o.StyleFired || o.Damage != 33 {
			t.Errorf("styled hit %d (fired %v), want 33", o.Damage, o.StyleFired)
		}
	})
	vector(t, ml, "Quick sword", func(t *testing.T) {
		if c := StyleCost(3.0, 10); c != 8 {
			t.Errorf("cost %d, want 8", c)
		}
	})
	vector(t, ml, "Slow axe", func(t *testing.T) {
		if c := StyleCost(5.0, 15); c != 18 {
			t.Errorf("cost %d, want 18", c)
		}
		if c := FailedStyleCost(5.0, 15); c != 9 {
			t.Errorf("failed cost %d, want 9", c)
		}
	})
	vector(t, ml, "A slow Troll", func(t *testing.T) {
		near(t, "swing", SwingTime(4.0, 35), 4.12, 1e-9)
	})
	vector(t, ml, "Floor", func(t *testing.T) {
		near(t, "swing", SwingTime(1.5, 100), 1.5, 1e-9)
	})
	vector(t, ml, "Studded torso", func(t *testing.T) {
		d := &Combatant{Player: true, Level: 10}
		d.Armour[LocTorso] = &Piece{AF: 20, Absorb: 0.19, Quality: 1, Condition: 1}
		af, abs := LocationArmour(d, LocTorso)
		near(t, "af", af, 24, 1e-9)
		near(t, "absorb", abs, 0.19, 1e-9)
	})
}

func TestLocationWeights(t *testing.T) {
	count := map[Location]int{}
	for i := 0; i < 100; i++ {
		count[RollLocation(float64(i)/100)]++
	}
	want := map[Location]int{LocTorso: 40, LocLegs: 25, LocArms: 15, LocHead: 10, LocHands: 5, LocFeet: 5}
	for l, n := range want {
		if count[l] != n {
			t.Errorf("%v: %d of 100, want %d", l, count[l], n)
		}
	}
}
