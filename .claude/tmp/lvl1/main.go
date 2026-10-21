// Scratch: level 1 Warrior against level 1 and 2 creatures, at mid variance.
package main

import (
	"fmt"

	"daocweb/internal/server/rules"
)

func main() {
	w := rules.Combatant{Player: true, Level: 1, WeaponSkillBase: 460, Spec: 1,
		Weapon: rules.Weapon{DPS: 1.2, Speed: 2.5, ObjectType: 13, Quality: 0.9, Condition: 1, DamageType: rules.DamageSlash}}
	w.Stats[rules.STR] = 110
	w.Stats[rules.CON] = 80
	w.Armour[rules.LocTorso] = &rules.Piece{AF: 4, Absorb: 0.19, Quality: 0.85, Condition: 1}
	for _, l := range []int{1, 2} {
		c := rules.Creature(l, 0, 0, 0, false, rules.DamageCrush, nil)
		out, _, _ := rules.HitDamage(rules.Swing{Attacker: &w, Defender: &c, Attackers: 1}, rules.LocTorso, 0.5)
		in, _, _ := rules.HitDamage(rules.Swing{Attacker: &c, Defender: &w, Attackers: 1}, rules.LocTorso, 0.5)
		fmt.Printf("level %d creature: hp %d, con %d; warrior hits %.2f, creature hits %.2f\n", l, rules.CreatureHP(l), rules.Con(1, l), out, in)
	}
	fmt.Println("warrior hp", rules.MaxHP(rules.Classes[22], 1, 80, 0))
}
