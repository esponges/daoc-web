package rules

import "testing"

const cc = "rules-creature-combat"

// level10Warrior is the defender of the creature scenarios: level 10, a
// studded torso of AF 20 and nothing else.
func level10Warrior() *Combatant {
	d := &Combatant{Player: true, Level: 10, Stats: Stats{115, 83, 47, 35}, Weapon: sword()}
	d.Armour[LocTorso] = &Piece{AF: 20, Absorb: 0.19, Quality: 1, Condition: 1}
	return d
}

func TestCreatureCombat(t *testing.T) {
	vector(t, cc, "Level 10 creature", func(t *testing.T) {
		c := plainCreature(10)
		for _, s := range []Stat{STR, CON, DEX, QUI} {
			if c.Stats[s] != 39 {
				t.Errorf("%v %d, want 39", s, c.Stats[s])
			}
		}
	})
	vector(t, cc, "Hit points by level", func(t *testing.T) {
		for l, want := range map[int]int{1: 31, 10: 206, 50: 2158} {
			if hp := CreatureHP(l); hp != want {
				t.Errorf("level %d: %d, want %d", l, hp, want)
			}
		}
	})
	vector(t, cc, "A level 10 creature hits a studded torso", func(t *testing.T) {
		o := Resolve(Swing{Attacker: plainCreature(10), Defender: level10Warrior(), InFront: false, Attackers: 1}, hitRolls(0.5))
		if o.Result != Hit || o.Location != LocTorso || o.Damage != 15 {
			t.Errorf("%v on %v for %d, want a hit on the torso for 15", o.Result, o.Location, o.Damage)
		}
	})
	vector(t, cc, "Bare skin hits the cap", func(t *testing.T) {
		// A location roll of 0.5 lands on the legs, which are bare.
		o := Resolve(Swing{Attacker: plainCreature(10), Defender: level10Warrior(), Attackers: 1},
			&seq{0.99, 0.99, 0.99, 0.5, 0.99, 0.5, 0.99, 0})
		if o.Location != LocLegs || o.Damage != 52 {
			t.Errorf("%v for %d, want the legs for 52", o.Location, o.Damage)
		}
	})
	vector(t, cc, "Level 50 creature", func(t *testing.T) {
		af, abs := CreatureArmour(50)
		near(t, "af", af, 156.2, 1e-9)
		near(t, "absorb", abs, 0.27, 1e-9)
	})
	vector(t, cc, "Level 10 one-handed creature", func(t *testing.T) {
		near(t, "swing", CreatureSwing(plainCreature(10)), 3.066, 1e-9)
	})
	vector(t, cc, "A thrust weakness", func(t *testing.T) {
		a := level10Troll()
		a.Weapon.DamageType = DamageThrus
		plain := plainCreature(10)
		weak := Creature(10, 0, 0, 0, false, DamageCrush, map[int]float64{DamageThrus: -15})
		r0, _, _ := HitDamage(Swing{Attacker: a, Defender: plain}, LocNone, 0.5)
		r1, _, _ := HitDamage(Swing{Attacker: a, Defender: &weak}, LocNone, 0.5)
		near(t, "weak / plain", r1/r0, 1.15, 1e-9)
	})
	vector(t, cc, "A parrying bandit", func(t *testing.T) {
		c := Creature(10, 0, 10, 0, false, DamageSlash, nil)
		near(t, "parry", ParryChance(&c, true, 3, true), 0.10, 1e-9)
		if ParryChance(&c, false, 1, false) != 0 {
			t.Error("parried from behind")
		}
	})
}
