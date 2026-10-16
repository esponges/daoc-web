package world

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// Item types (the item table's slot numbers) mapped to the worn slot.
func wornSlot(it *Item) int {
	switch it.ItemType {
	case 10, 12:
		return SlotWeapon
	case 11:
		if it.IsShield() {
			return SlotShield
		}
		return SlotWeapon
	case 25:
		return SlotTorso
	case 27:
		return SlotLegs
	case 28:
		return SlotArms
	case 23:
		return SlotFeet
	case 22:
		return SlotHands
	case 21:
		return SlotHead
	}
	return -1
}

// A warrior's weapons and armour. The class tables name what each class may
// use, but the emulator keeps those in code, so this is ours.
func canUse(class int, it *Item, level int) error {
	if it.Level > level+2 && level < MaxLevel {
		return fmt.Errorf("you must be level %d to use %s", it.Level-2, it.Name)
	}
	if it.Realm != 0 && it.Realm != 2 {
		return errors.New("that is not Midgard's")
	}
	switch {
	case it.ObjectType == ObjShield:
		return nil
	case it.ObjectType >= 11 && it.ObjectType <= 13: // sword, hammer, axe
		if it.Hand == 2 {
			return errors.New("a warrior here fights one-handed")
		}
		return nil
	case it.ObjectType >= 32 && it.ObjectType <= 35: // cloth, leather, studded, chain
		return nil
	}
	return errors.New("you cannot use that")
}

func (w *World) loadInventory(p *Player) {
	p.inv = map[int]*invItem{}
	items, err := w.Store.Items(p.Char.ID)
	if err != nil {
		w.Log("loading items of %s: %v", p.Char.Name, err)
	}
	for _, it := range items {
		p.inv[it.Slot] = &invItem{DBID: it.ID, ID: it.ItemID, Slot: it.Slot, Count: it.Count}
	}
	w.recalcGear(p)
}

// recalcGear derives what a player holds and wears from their inventory.
func (w *World) recalcGear(p *Player) {
	p.weapon, p.shield, p.armour = nil, nil, 0
	if it := p.inv[SlotWeapon]; it != nil {
		p.weapon = w.Cat.Items[it.ID]
	}
	if it := p.inv[SlotShield]; it != nil {
		p.shield = w.Cat.Items[it.ID]
	}
	for s := SlotTorso; s < SlotCount; s++ {
		if it := p.inv[s]; it != nil {
			if x := w.Cat.Items[it.ID]; x != nil {
				p.armour += x.AF
			}
		}
	}
	if gear := gearOf(p.inv, w.Cat.Items); !sameGear(gear, p.gear) {
		p.gear = gear
		p.gearRev++
	}
}

// gearOf lists what a player wears as the pieces an NPC's equipment lists:
// [slot, model, colour, extension]. The slot is the item table's own, which is
// also what the community database uses for an NPC, so the client draws both
// the same way. Colour is not tracked for items, so it is 0 (untinted).
func gearOf(inv map[int]*invItem, items map[string]*Item) [][4]int {
	out := [][4]int{}
	for s := SlotWeapon; s < SlotCount; s++ {
		it := inv[s]
		if it == nil {
			continue
		}
		if x := items[it.ID]; x != nil && x.Model > 0 && wornSlot(x) == s {
			out = append(out, [4]int{x.ItemType, x.Model, 0, 0})
		}
	}
	return out
}

func sameGear(a, b [][4]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func itemView(slot int, it *Item, count int) protocol.InvItem {
	v := protocol.InvItem{Slot: slot, ID: it.ID, Name: it.Name, Count: count, Level: it.Level, Model: it.Model}
	switch {
	case it.IsShield():
		v.Kind, v.Detail = "shield", fmt.Sprintf("shield, level %d", it.Level)
	case it.IsWeapon():
		v.Kind, v.Detail = "weapon", fmt.Sprintf("%.1f dps, %.1fs swing, level %d", it.DPS, it.Speed, it.Level)
		v.Line = map[int]string{11: "Sword", 12: "Hammer", 13: "Axe"}[it.ObjectType]
	case it.IsArmour():
		v.Kind, v.Detail = "armour", fmt.Sprintf("AF %d, level %d", it.AF, it.Level)
		if s := wornSlot(it); s >= 0 {
			v.Worn = s
		}
	default:
		v.Kind = "misc"
		if it.Level > 0 {
			v.Detail = fmt.Sprintf("level %d", it.Level)
		}
	}
	return v
}

func (p *Player) sendInv(c *Catalog, money int64) {
	m := protocol.Inv{T: protocol.TInv, Money: money, Eq: p.gear}
	if m.Eq == nil {
		m.Eq = [][4]int{}
	}
	for _, slot := range sortedSlots(p.inv) {
		if it := c.Items[p.inv[slot].ID]; it != nil {
			m.Items = append(m.Items, itemView(slot, it, p.inv[slot].Count))
		}
	}
	p.send(m)
}

func sortedSlots(m map[int]*invItem) []int {
	s := make([]int, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	sort.Ints(s)
	return s
}

// addItem puts an item in the first free backpack slot, saving it with the
// character in one transaction.
func (w *World) addItem(p *Player, id string, count int) (int, error) {
	it := w.Cat.Items[id]
	if it == nil {
		return -1, fmt.Errorf("there is no item %q", id)
	}
	if count < 1 {
		count = 1
	}
	slot := -1
	err := w.Store.Tx(func(tx *store.Tx) error {
		var err error
		slot, err = tx.FreeSlot(p.Char.ID, PackSize)
		if err != nil {
			return err
		}
		if slot < 0 {
			return errors.New("your pack is full")
		}
		if err := tx.AddItem(p.Char.ID, id, slot, count); err != nil {
			return err
		}
		return tx.SaveCharacter(p.Char)
	})
	if err != nil {
		return -1, err
	}
	p.inv[slot] = &invItem{ID: id, Slot: slot, Count: count}
	p.sendInv(w.Cat, p.Char.Money)
	return slot, nil
}

// MoveItem moves (or swaps) an item between slots. Moving into a worn slot
// equips it, if the item and the character allow.
func (z *Zone) MoveItem(p *Player, from, to int) {
	w := z.W
	if from == to || from < 0 || to < 0 || from >= SlotCount || to >= SlotCount {
		return
	}
	src := p.inv[from]
	if src == nil {
		return
	}
	check := func(it *invItem, slot int) error {
		if slot < PackSize || it == nil {
			return nil
		}
		x := w.Cat.Items[it.ID]
		if x == nil {
			return errors.New("unknown item")
		}
		if wornSlot(x) != slot {
			return errors.New("that does not go there")
		}
		return canUse(p.Char.Class, x, p.Char.Level)
	}
	dst := p.inv[to]
	if err := check(src, to); err != nil {
		p.errMsg("equip", err.Error())
		return
	}
	if err := check(dst, from); err != nil {
		p.errMsg("equip", err.Error())
		return
	}
	err := w.Store.Tx(func(tx *store.Tx) error {
		if dst != nil {
			return tx.SwapItems(p.Char.ID, from, to)
		}
		return tx.MoveItem(p.Char.ID, from, to)
	})
	if err != nil {
		p.errMsg("equip", "Could not move that.")
		return
	}
	if dst != nil {
		p.inv[from], p.inv[to] = dst, src
		dst.Slot, src.Slot = from, to
	} else {
		delete(p.inv, from)
		p.inv[to] = src
		src.Slot = to
	}
	w.recalcGear(p)
	z.buildBar(p)
	p.sendInv(w.Cat, p.Char.Money)
}

// DropItem destroys the item in a slot.
func (z *Zone) DropItem(p *Player, slot int) {
	if p.inv[slot] == nil {
		return
	}
	if err := z.W.Store.Tx(func(tx *store.Tx) error { return tx.RemoveItem(p.Char.ID, slot) }); err != nil {
		return
	}
	delete(p.inv, slot)
	z.W.recalcGear(p)
	z.buildBar(p)
	p.sendInv(z.W.Cat, p.Char.Money)
}

// GiveStarterKit equips a new Midgard warrior from the item table: the
// lowest-level hammer, shield and studded armour it has.
func (w *World) GiveStarterKit(charID int64) error {
	pick := func(match func(*Item) bool) *Item {
		var best *Item
		for _, it := range w.Cat.Items {
			if !it.Droppable || !match(it) {
				continue
			}
			if best == nil || it.Level < best.Level || it.Level == best.Level && it.ID < best.ID {
				best = it
			}
		}
		return best
	}
	kit := []struct {
		slot  int
		match func(*Item) bool
	}{
		{SlotWeapon, func(i *Item) bool {
			return i.ObjectType == ObjHammer && i.ItemType == 10 && i.DPS > 0 && i.Realm == 2 && i.Hand == 0
		}},
		{SlotShield, func(i *Item) bool { return i.ObjectType == ObjShield && i.Realm == 2 && i.Level >= 1 }},
		{SlotTorso, func(i *Item) bool { return i.ObjectType == 34 && i.ItemType == 25 && i.Realm == 2 && i.Level >= 1 }},
		{SlotLegs, func(i *Item) bool { return i.ObjectType == 34 && i.ItemType == 27 && i.Realm == 2 && i.Level >= 1 }},
		{SlotArms, func(i *Item) bool { return i.ObjectType == 34 && i.ItemType == 28 && i.Realm == 2 && i.Level >= 1 }},
		{SlotFeet, func(i *Item) bool { return i.ObjectType == 34 && i.ItemType == 23 && i.Realm == 2 && i.Level >= 1 }},
		{SlotHands, func(i *Item) bool { return i.ObjectType == 34 && i.ItemType == 22 && i.Realm == 2 && i.Level >= 1 }},
	}
	return w.Store.Tx(func(tx *store.Tx) error {
		for _, k := range kit {
			if it := pick(k.match); it != nil {
				if err := tx.AddItem(charID, it.ID, k.slot, 1); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// randomDrop picks a plausible item for a creature of a level to carry.
func (c *Catalog) randomDrop(rng *rand.Rand, level int) string {
	c.dropOnce.Do(func() {
		c.dropPool = map[int][]string{}
		for _, it := range c.Items {
			if !it.Droppable || it.Price <= 0 || it.Level < 1 || (it.Realm != 0 && it.Realm != 2) {
				continue
			}
			switch {
			case it.ObjectType >= 11 && it.ObjectType <= 13 && it.Hand != 2, it.ObjectType == ObjShield,
				it.ObjectType >= 32 && it.ObjectType <= 35 && wornSlot(it) >= 0:
				c.dropPool[it.Level] = append(c.dropPool[it.Level], it.ID)
			}
		}
		for _, l := range c.dropPool {
			sort.Strings(l)
		}
	})
	var pool []string
	for l := level - 2; l <= level; l++ {
		pool = append(pool, c.dropPool[l]...)
	}
	if len(pool) == 0 {
		return ""
	}
	return pool[rng.Intn(len(pool))]
}

// AutoEquip wears an item in the slot it belongs in.
func (z *Zone) AutoEquip(p *Player, from int) {
	it := p.inv[from]
	if it == nil || from >= PackSize {
		return
	}
	x := z.W.Cat.Items[it.ID]
	if x == nil || wornSlot(x) < 0 {
		p.errMsg("equip", "That cannot be worn.")
		return
	}
	z.MoveItem(p, from, wornSlot(x))
}
