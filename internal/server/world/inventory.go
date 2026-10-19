package world

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
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

// canUse says whether a player may equip an item: their class's abilities at
// their level decide the weapon families, armour weight and shield size; the
// item must be Midgard's or no realm's and not ask for a higher level.
func canUse(p *Player, it *Item) error {
	if it.LevelReq > p.Char.Level {
		return fmt.Errorf("you must be level %d to use %s", it.LevelReq, it.Name)
	}
	if it.Realm != 0 && it.Realm != 2 {
		return errors.New("that is not Midgard's")
	}
	return rules.MayUse(p.sh.abilities, it.ObjectType, it.DamageType)
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
	w.recalc(p)
}

// recalcGear derives what a player holds and wears from their inventory.
func (w *World) recalcGear(p *Player) {
	p.weapon, p.shield = nil, nil
	if it := p.inv[SlotWeapon]; it != nil {
		p.weapon = w.Cat.Items[it.ID]
	}
	if it := p.inv[SlotShield]; it != nil {
		p.shield = w.Cat.Items[it.ID]
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
		return canUse(p, x)
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
	// Two hands hold one weapon: a two-handed weapon sends the shield to the
	// pack, and a shield cannot go on beside one.
	shieldOff := -1
	if to == SlotWeapon && p.inv[SlotShield] != nil {
		if x := w.Cat.Items[src.ID]; x != nil && x.TwoHanded() {
			shieldOff = SlotShield
		}
	}
	if to == SlotShield && p.weapon != nil && p.weapon.TwoHanded() {
		p.errMsg("equip", "You cannot use a shield with a two-handed weapon.")
		return
	}
	freeSlot := -1
	err := w.Store.Tx(func(tx *store.Tx) error {
		if shieldOff >= 0 {
			slot, err := tx.FreeSlot(p.Char.ID, PackSize)
			if err != nil {
				return err
			}
			if slot < 0 {
				return errPackFull
			}
			if err := tx.MoveItem(p.Char.ID, SlotShield, slot); err != nil {
				return err
			}
			freeSlot = slot
		}
		if dst != nil {
			return tx.SwapItems(p.Char.ID, from, to)
		}
		return tx.MoveItem(p.Char.ID, from, to)
	})
	if errors.Is(err, errPackFull) {
		p.errMsg("equip", "Your pack has no room for your shield.")
		return
	}
	if err != nil {
		p.errMsg("equip", "Could not move that.")
		return
	}
	if freeSlot >= 0 {
		sh := p.inv[SlotShield]
		delete(p.inv, SlotShield)
		p.inv[freeSlot] = sh
		sh.Slot = freeSlot
	}
	if dst != nil {
		p.inv[from], p.inv[to] = dst, src
		dst.Slot, src.Slot = from, to
	} else {
		delete(p.inv, from)
		p.inv[to] = src
		src.Slot = to
	}
	w.recalc(p)
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
	z.W.recalc(p)
	z.buildBar(p)
	p.sendInv(z.W.Cat, p.Char.Money)
}

// GiveStarterKit gives a new character what the imported starting-equipment
// table lists for their class, in the table's order: each item worn in its
// slot when the slot is still free and the class may use it, otherwise in the
// pack. An item the item table lacks is skipped and logged; nothing is chosen
// any other way.
func (w *World) GiveStarterKit(c *store.Character) error {
	ids := w.Cat.StartEquip[c.Class]
	if len(ids) == 0 {
		w.Log("starting equipment: the table lists nothing for class %d; %s starts with nothing", c.Class, c.Name)
		return nil
	}
	trained := map[string]int{}
	for _, k := range w.Cat.TrainableSpecs(c.Class) {
		trained[k] = 1
	}
	abilities := rules.Abilities(w.Cat.Grants[c.Class], w.Cat.Careers(c.Class), max(1, c.Level), trained)
	return w.Store.Tx(func(tx *store.Tx) error {
		taken := map[int]bool{}
		pack := 0
		for _, id := range ids {
			it := w.Cat.Items[id]
			if it == nil {
				w.Log("starting equipment: class %d lists %q, which is not in the item table", c.Class, id)
				continue
			}
			slot := wornSlot(it)
			if slot < 0 || taken[slot] || it.LevelReq > c.Level || rules.MayUse(abilities, it.ObjectType, it.DamageType) != nil {
				slot, pack = pack, pack+1
			}
			taken[slot] = true
			if err := tx.AddItem(c.ID, id, slot, 1); err != nil {
				return err
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
