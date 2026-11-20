package world

import (
	"math/rand"
	"testing"

	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// leftRig is a fighter with an axe in the right hand and a left axe and a pair of
// claws in the pack, against a wolf close by.
func leftRig(t *testing.T) (*rig, *Player, *NPC) {
	r := newRig(t, wolf(1, 30060, 30000, 8))
	r.z.rng = rand.New(rand.NewSource(12))
	c := r.w.Cat
	c.Items["laxe"] = &Item{ID: "laxe", Name: "left axe", Level: 1, DPS: 4, Speed: 3.0, ObjectType: rules.ObjLeftAxe, ItemType: 11, Hand: 2, Realm: 2}
	c.Items["claw_r"] = &Item{ID: "claw_r", Name: "right claw", Level: 1, DPS: 4, Speed: 3.0, ObjectType: rules.ObjHandToHand, ItemType: 10, Realm: 2}
	c.Items["claw_l"] = &Item{ID: "claw_l", Name: "left claw", Level: 1, DPS: 4, Speed: 3.0, ObjectType: rules.ObjHandToHand, ItemType: 11, Hand: 2, Realm: 2}
	w := r.z.npcs[0]
	w.Aggro, w.Roam = 0, 0
	w.setLevel(10)
	w.HP, w.MaxHP = 1e6, 1e6
	p := r.player("Twin", 20, 30000, 30000, true) // a hammer in the right hand
	r.ticks(TickRate)
	return r, p, w
}

func swingsInRound(r *rig, p *Player, w *NPC) int {
	before := len(r.z.ev)
	p.swingAt = 0
	r.z.playerSwing(p, w)
	n := 0
	for _, e := range r.z.ev[before:] {
		if e.e.K == "swing" {
			n++
		}
	}
	return n
}

func TestAWarriorCannotWearAWeaponInTheLeftHand(t *testing.T) {
	r, p, _ := leftRig(t)
	r.give(p, 0, "laxe", 1)
	r.z.MoveItem(p, 0, SlotLeft)
	if p.left != nil || p.inv[SlotLeft] == nil || p.inv[SlotLeft].ID != "shield" {
		t.Error("a Warrior wore a left axe in the left hand")
	}
	if errs := errsOf(p); len(errs) == 0 {
		t.Error("no reason was given")
	}
}

func TestALeftAxeSwingsEveryRoundAtItsShare(t *testing.T) {
	r, p, w := leftRig(t)
	p.sh.abilities["Weaponry: Left Axes"] = 1
	p.sh.abilities["Weaponry: Hammers"] = 1
	p.Char.Specs = map[string]int{"Left Axe": 20}
	r.give(p, 0, "laxe", 1)
	r.z.MoveItem(p, 0, SlotLeft)
	if p.left == nil || p.inv[SlotLeft] == nil {
		t.Fatalf("the left axe did not equip: %v", errsOf(p))
	}
	p.sh.abilities["Weaponry: Left Axes"] = 1 // the move recalculated the sheet
	for i := 0; i < 5; i++ {
		if got := swingsInRound(r, p, w); got != 2 {
			t.Fatalf("round %d had %d swings, want the main hand and the left", i, got)
		}
	}
	if w.hurtBy[p.ID] == 0 {
		t.Error("no blow landed in five rounds")
	}
	// Without the spec's weaponry the left hand does nothing.
	p.sh.abilities["Weaponry: Left Axes"] = 0
	if got := swingsInRound(r, p, w); got != 1 {
		t.Errorf("an untrained left hand swung: %d", got)
	}
}

func TestHandToHandExtraBlowsComeByChance(t *testing.T) {
	r, p, w := leftRig(t)
	p.sh.abilities["Weaponry: Hand to Hand"] = 1
	p.Char.Specs = map[string]int{"Hand to Hand": 50}
	r.inventoryClaws(p)
	r.w.recalcGear(p)
	if p.left == nil || p.weapon == nil {
		t.Fatalf("the claws did not equip: %v", errsOf(p))
	}
	total := 0
	const rounds = 3000
	for i := 0; i < rounds; i++ {
		total += swingsInRound(r, p, w) - 1
		r.z.ev = r.z.ev[:0]
		w.HP = w.MaxHP
	}
	// Chances at spec 50: 25 percent for one extra, 12.5 for two, 3.1 for three.
	want := 0.25 + 2*0.125 + 3*0.03125
	if got := float64(total) / rounds; got < want-0.06 || got > want+0.06 {
		t.Errorf("%.3f extra blows a round, want about %.3f", got, want)
	}
}

func (r *rig) inventoryClaws(p *Player) {
	r.st.Tx(func(tx *store.Tx) error {
		tx.RemoveItem(p.Char.ID, SlotShield)
		return tx.RemoveItem(p.Char.ID, SlotWeapon)
	})
	delete(p.inv, SlotWeapon)
	delete(p.inv, SlotShield)
	r.give(p, SlotWeapon, "claw_r", 1)
	r.give(p, SlotLeft, "claw_l", 1)
}
