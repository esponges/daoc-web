package world

import (
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// poisonRig is a hider-class player with the Envenom spec, a hammer in hand, a
// poison in the pack, and a wolf beside them.
func poisonRig(t *testing.T, envenom int) (*rig, *Player, *NPC) {
	r := newRig(t, wolf(1, 30060, 30000, 8))
	r.z.rng = rand.New(rand.NewSource(8))
	c := r.w.Cat
	c.ClassSpecs[rules.ClassHunter] = []string{"Envenom"}
	c.Items["poison5"] = &Item{ID: "poison5", Name: "lesser lethal poison", Level: 5, ObjectType: objPoison, PoisonSpell: 9950, PoisonCharges: 2}
	c.Items["poison20"] = &Item{ID: "poison20", Name: "major lethal poison", Level: 20, ObjectType: objPoison, PoisonSpell: 9950, PoisonCharges: 1}
	c.Spells = map[int]*SpellDef{}
	c.Spells[9950] = &SpellDef{ID: 9950, Name: "Lesser Lethal Poison", Type: rules.SpellDamageOverTime, Target: "Enemy", Range: 500,
		Damage: 9, Duration: 20, Frequency: 40, DamageType: 10}
	w := r.z.npcs[0]
	w.Aggro, w.Roam = 0, 0
	w.setLevel(20)
	p := r.player("Poisoner", 20, 30000, 30000, true)
	p.Char.Class = rules.ClassHunter
	p.Char.Specs = map[string]int{"Envenom": envenom}
	r.w.recalc(p)
	r.ticks(TickRate)
	r.give(p, 0, "poison5", 3)
	r.give(p, 1, "poison20", 1)
	return r, p, w
}

func savedPoison(t *testing.T, r *rig, p *Player, slot int) (int, int) {
	t.Helper()
	items, _ := r.st.Items(p.Char.ID)
	for _, it := range items {
		if it.Slot == slot {
			return it.Poison, it.Charges
		}
	}
	return -1, -1
}

func TestApplyingAPoisonToTheWeapon(t *testing.T) {
	r, p, _ := poisonRig(t, 15)
	r.z.UseItem(p, 0)
	w := p.inv[SlotWeapon]
	if w.Poison != 9950 || w.Charges != 2 {
		t.Fatalf("the weapon holds poison %d with %d charges (%v)", w.Poison, w.Charges, errsOf(p))
	}
	if p.inv[0].Count != 2 {
		t.Errorf("%d poisons left in the stack, want 2", p.inv[0].Count)
	}
	if sp, ch := savedPoison(t, r, p, SlotWeapon); sp != 9950 || ch != 2 {
		t.Errorf("saved poison %d with %d charges", sp, ch)
	}
	if left := savedCount(t, r, p, 0); left != 2 {
		t.Errorf("the saved stack is %d, want 2", left)
	}
	sentOf[protocol.Inv](p, protocol.TInv) // drain
	p.sendInv(r.w.Cat, 0)
	found := false
	for _, m := range sentOf[protocol.Inv](p, protocol.TInv) {
		for _, it := range m.Items {
			if it.Slot == SlotWeapon && strings.Contains(it.Detail, "poisoned with Lesser Lethal Poison (2)") && it.Poison == 2 {
				found = true
			}
		}
	}
	if !found {
		t.Error("the weapon does not show its poison and charges")
	}
	// A weapon already poisoned is poisoned again (replaced) when it is the only weapon.
	r.z.UseItem(p, 0)
	if p.inv[SlotWeapon].Charges != 2 || p.inv[0].Count != 1 {
		t.Errorf("a second poison: %d charges, %d left in the stack", p.inv[SlotWeapon].Charges, p.inv[0].Count)
	}
}

func TestNotEnoughEnvenom(t *testing.T) {
	r, p, _ := poisonRig(t, 5)
	r.z.UseItem(p, 1) // a level 20 poison
	if p.inv[SlotWeapon].Poison != 0 || p.inv[1] == nil || !strings.Contains(lastErr(t, p), "too low") {
		t.Errorf("a level 20 poison with Envenom 5: poison %d, error %q", p.inv[SlotWeapon].Poison, lastErr(t, p))
	}
	r2, p2, _ := poisonRig(t, 0)
	p2.Char.Class = rules.ClassWarrior
	r2.z.UseItem(p2, 0)
	if p2.inv[SlotWeapon].Poison != 0 {
		t.Error("a class without Envenom applied a poison")
	}
}

func TestAPoisonedWeaponFiresOnHitsAndRunsOut(t *testing.T) {
	r, p, w := poisonRig(t, 15)
	r.z.UseItem(p, 0)
	p.Char.Specs["Hammer"] = 30
	w.HP, w.MaxHP = 100000, 100000
	hits := 0
	for i := 0; i < 80 && p.inv[SlotWeapon].Poison != 0; i++ {
		before := p.inv[SlotWeapon].Charges
		r.z.playerSwing(p, w)
		if p.inv[SlotWeapon].Charges < before {
			hits++
			if !hasKind(w, effDoT) {
				t.Fatal("a hit used a charge but left no poison on the creature")
			}
		}
	}
	if hits != 2 || p.inv[SlotWeapon].Poison != 0 {
		t.Errorf("the poison fired %d times (want 2) and left poison %d", hits, p.inv[SlotWeapon].Poison)
	}
	if sp, ch := savedPoison(t, r, p, SlotWeapon); sp != 0 || ch != 0 {
		t.Errorf("a spent poison is still saved: %d with %d charges", sp, ch)
	}
	if !said(p, "spent") {
		t.Error("no message that the poison is spent")
	}
}

// Every poison a Midgard merchant sells can be put on a weapon, or is named
// here as one that cannot (its spell is not in the data).
func TestEveryPoisonOnSaleHasASpell(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	sold := map[string]bool{}
	for _, ids := range cat.Stock {
		for _, id := range ids {
			if it := cat.Items[id]; it != nil && it.ObjectType == objPoison {
				sold[id] = true
			}
		}
	}
	if len(sold) == 0 {
		t.Fatal("no merchant sells a poison")
	}
	bad, good := 0, 0
	for id := range sold {
		it := cat.Items[id]
		sp := cat.Spells[it.PoisonSpell]
		if sp != nil && sp.Type == rules.SpellOffensiveProc {
			sp = cat.Spells[int(sp.Value)]
		}
		if it.PoisonSpell == 0 || sp == nil {
			bad++
			t.Logf("%s (level %d) sells with no spell: %d", id, it.Level, it.PoisonSpell)
			continue
		}
		if !rules.ImplementedPlayerSpell(sp.Type) {
			t.Errorf("poison %s fires %s, a type the server cannot cast", id, sp.Type)
		}
		good++
	}
	t.Logf("%d poisons on sale work, %d lack a spell", good, bad)
	if good == 0 {
		t.Error("no poison on sale works")
	}
}
