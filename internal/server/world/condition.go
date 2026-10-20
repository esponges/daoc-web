package world

import (
	"fmt"
	"math"
	"sort"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// KindSmith is what a blacksmith does when interacted with: repairs items.
const KindSmith = "smith"

// condOf is an item's condition as a multiplier; an item they do not hold in
// that slot counts as whole.
func (p *Player) condOf(slot int) float64 {
	if it := p.inv[slot]; it != nil {
		return rules.Effectiveness(it.Cond)
	}
	return 1
}

// wearItem lowers the condition of the item in a slot by one blow's wear,
// warning its owner as it falls below each threshold.
func (z *Zone) wearItem(p *Player, slot int) {
	it := p.inv[slot]
	if it == nil {
		return
	}
	before := it.Cond
	it.Cond = rules.Worn(before)
	if it.Cond == before {
		return
	}
	p.condDirty = true
	if t := rules.WarnCrossed(before, it.Cond); t > 0 {
		name := it.ID
		if x := z.W.Cat.Items[it.ID]; x != nil {
			name = x.Name
		}
		sys(p, fmt.Sprintf("Your %s is below %d%% condition and needs repair.", name, t/100))
	}
}

// wearBlow wears what a creature's blow struck: the armour at the place it
// landed, or the shield when it was blocked.
func (z *Zone) wearBlow(p *Player, o rules.Outcome) {
	armour, shield := rules.BlowWears(o.Result)
	switch {
	case armour && o.Location >= 0 && int(o.Location) < len(locationSlot):
		z.wearItem(p, locationSlot[o.Location])
	case shield:
		z.wearItem(p, SlotShield)
	}
}

// saveConditions writes the conditions that changed since the last save, in
// one transaction. Wear is kept in memory between saves.
func (w *World) saveConditions(p *Player) {
	if !p.condDirty {
		return
	}
	err := w.Store.Tx(func(tx *store.Tx) error {
		for slot, it := range p.inv {
			if err := tx.SetCondition(p.Char.ID, slot, it.Cond); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		w.Log("saving item condition of %s: %v", p.Char.Name, err)
		return
	}
	p.condDirty = false
}

// --- the blacksmith ---

// repairable says whether an item can wear: weapons, shields and armour.
func repairable(it *Item) bool { return it != nil && (it.IsWeapon() || it.IsShield() || it.IsArmour()) }

// repairPrice is what repairing the item in a slot costs, 0 if it needs none.
func (z *Zone) repairPrice(p *Player, slot int) int64 {
	it := p.inv[slot]
	if it == nil || slot >= SlotCount {
		return 0
	}
	x := z.W.Cat.Items[it.ID]
	if !repairable(x) {
		return 0
	}
	return rules.RepairPrice(x.Price, it.Cond)
}

// openSmith answers a blacksmith's interaction: what the player holds that
// needs repair, with the price of each.
func (z *Zone) openSmith(p *Player, n *NPC) {
	z.closeShop(p)
	z.closeVault(p)
	p.smithNPC = n.ID
	p.sendSmith(z, n)
}

func (p *Player) sendSmith(z *Zone, n *NPC) {
	m := protocol.Smith{T: protocol.TSmith, NPC: p.smithNPC, Items: []protocol.InvItem{}}
	if n != nil {
		m.Name = n.title()
	}
	slots := make([]int, 0, len(p.inv))
	for s := range p.inv {
		slots = append(slots, s)
	}
	sort.Ints(slots)
	for _, s := range slots {
		if price := z.repairPrice(p, s); price > 0 {
			v := itemView(s, z.W.Cat.Items[p.inv[s].ID], 1)
			v.Detail = conditionText(p.inv[s].Cond)
			v.Price = price
			m.Items = append(m.Items, v)
		}
	}
	p.send(m)
}

func (z *Zone) closeSmith(p *Player) {
	if p.smithNPC == 0 {
		return
	}
	p.smithNPC = 0
	p.send(protocol.SmithClose{T: protocol.TSmithClose})
}

// checkSmith closes the repair window when the blacksmith is out of reach.
func (z *Zone) checkSmith(p *Player) {
	if p.smithNPC == 0 {
		return
	}
	n := z.byID[p.smithNPC]
	if p.dead || n == nil || n.gone || n.dead() || !n.inReach(p) {
		z.closeSmith(p)
	}
}

func conditionText(cond int) string {
	return fmt.Sprintf("%.0f%% condition", math.Floor(float64(cond)/100))
}

// Repair handles a repair request at the open blacksmith. Without Accept it
// quotes the price; with it, takes the price the player was shown and repairs.
func (z *Zone) Repair(p *Player, m *protocol.Repair) {
	n := z.byID[p.smithNPC]
	if p.smithNPC == 0 || p.dead || n == nil || n.gone || n.dead() || !n.inReach(p) {
		z.closeSmith(p)
		p.errMsg("repair", "You are not talking to a blacksmith.")
		return
	}
	it := p.inv[m.Slot]
	price := z.repairPrice(p, m.Slot)
	if it == nil || price == 0 {
		p.errMsg("repair", "That does not need repair.")
		return
	}
	name := z.W.Cat.Items[it.ID].Name
	if !m.Accept || m.Price != price {
		// Quote the price; or, if the condition changed since they were
		// quoted, ask again.
		p.send(protocol.RepairQuote{T: protocol.TRepairQuote, Slot: m.Slot, Name: name, Price: price})
		return
	}
	if p.Char.Money < price {
		p.errMsg("repair", "You cannot afford to repair the "+name+".")
		return
	}
	before := p.Char.Money
	err := z.W.Store.Tx(func(tx *store.Tx) error {
		if err := tx.SetCondition(p.Char.ID, m.Slot, rules.FullCondition); err != nil {
			return err
		}
		p.Char.Money -= price
		return tx.SaveCharacter(p.Char)
	})
	if err != nil {
		p.Char.Money = before
		p.errMsg("repair", "Could not save that; try again.")
		return
	}
	it.Cond = rules.FullCondition
	z.W.recalc(p)
	sys(p, fmt.Sprintf("%s repairs your %s.", n.title(), name))
	p.sendInv(z.W.Cat, p.Char.Money)
	p.sendSmith(z, n)
}
