package world

import (
	"fmt"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// Poisons, from the rules-poisons capability. A poison item goes on a worn
// weapon, which then fires the poison's spell on every hit it lands and uses
// one charge each time. The poison is saved with the weapon.

const objPoison = 46 // the object type of a poison item

// UseItem is the player's use of an item in a pack slot: a poison goes on a weapon.
func (z *Zone) UseItem(p *Player, slot int) {
	if p.dead || slot < 0 || slot >= PackSize {
		return
	}
	held := p.inv[slot]
	if held == nil {
		return
	}
	it := z.W.Cat.Items[held.ID]
	if it == nil || it.ObjectType != objPoison {
		p.errMsg("use", "You cannot use that.")
		return
	}
	z.applyPoison(p, slot, it)
}

// poisonTarget is the worn weapon slot a poison goes on: the right hand when it
// holds a weapon with no poison, else the left, else the right (replacing it).
func (p *Player) poisonTarget(cat *Catalog) int {
	weapon := func(slot int) bool {
		it := p.inv[slot]
		if it == nil {
			return false
		}
		x := cat.Items[it.ID]
		return x != nil && x.IsWeapon() && !x.IsShield()
	}
	switch {
	case weapon(SlotWeapon) && p.inv[SlotWeapon].Poison == 0:
		return SlotWeapon
	case weapon(SlotLeft) && p.inv[SlotLeft].Poison == 0:
		return SlotLeft
	case weapon(SlotWeapon):
		return SlotWeapon
	}
	return -1
}

func (z *Zone) applyPoison(p *Player, slot int, it *Item) {
	cat := z.W.Cat
	if !p.hasSpec(cat, "Envenom") || p.specLevel("Envenom") < 1 {
		p.errMsg("use", "You do not know how to use poisons.")
		return
	}
	target := p.poisonTarget(cat)
	if target < 0 {
		p.errMsg("use", "You have no weapon to poison.")
		return
	}
	if it.PoisonSpell == 0 || cat.Spells[it.PoisonSpell] == nil {
		p.errMsg("use", "That poison has no effect in this world.")
		return
	}
	if msg := rules.PoisonRefusal(p.specLevel("Envenom"), p.Char.Level, it.Level, z.now < p.combatUntil || held(p)); msg != "" {
		p.errMsg("use", msg)
		return
	}
	charges := max(1, it.PoisonCharges)
	err := z.W.Store.Tx(func(tx *store.Tx) error {
		if err := tx.SetPoison(p.Char.ID, target, it.PoisonSpell, charges); err != nil {
			return err
		}
		return z.takeOne(tx, p, slot)
	})
	if err != nil {
		p.errMsg("use", "Could not apply that; try again.")
		return
	}
	w := p.inv[target]
	w.Poison, w.Charges = it.PoisonSpell, charges
	z.dropOne(p, slot)
	wi := cat.Items[w.ID]
	sys(p, fmt.Sprintf("You apply %s to %s.", it.Name, wi.Name))
	p.sendInv(cat, p.Char.Money)
}

// takeOne removes one of a stack from a pack slot, in a transaction.
func (z *Zone) takeOne(tx *store.Tx, p *Player, slot int) error {
	held := p.inv[slot]
	if held.Count > 1 {
		return tx.SetCount(p.Char.ID, slot, held.Count-1)
	}
	return tx.RemoveItem(p.Char.ID, slot)
}

// dropOne mirrors takeOne in the player's memory.
func (z *Zone) dropOne(p *Player, slot int) {
	held := p.inv[slot]
	if held.Count > 1 {
		held.Count--
		return
	}
	delete(p.inv, slot)
}

// poisonSpell is the spell a poison spell does: a poison that is a proc carries
// the spell it fires in its value.
func (z *Zone) poisonSpell(id int) *SpellDef {
	sp := z.W.Cat.Spells[id]
	if sp != nil && sp.Type == rules.SpellOffensiveProc {
		return z.W.Cat.Spells[int(sp.Value)]
	}
	return sp
}

// firePoison is what a landed blow with the weapon in a slot does if it is
// poisoned: the poison's spell on the target, and one charge less.
func (z *Zone) firePoison(p *Player, n *NPC, slot int) {
	w := p.inv[slot]
	if w == nil || w.Poison == 0 || n.dead() {
		return
	}
	if sp := z.poisonSpell(w.Poison); sp != nil {
		z.castSpellAs(p, n, sp, p.Char.Level, false)
	}
	w.Charges--
	if w.Charges <= 0 {
		w.Poison, w.Charges = 0, 0
		sys(p, "Your weapon's poison is spent.")
		z.event(protocol.Event{K: "poisonspent", From: p.ID}, p.Body.X, p.Body.Y)
	}
	z.W.Store.Tx(func(tx *store.Tx) error { return tx.SetPoison(p.Char.ID, slot, w.Poison, w.Charges) })
	p.sendInv(z.W.Cat, p.Char.Money)
}
