package world

import (
	"errors"
	"log"
	"math"

	"daocweb/internal/importer"
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// What interacting with an NPC does.
const (
	KindMerchant = "merchant"
	// KindNotClassic marks NPCs whose service 1.65 did not have; they answer with a line.
	KindNotClassic = "notclassic"
)

// A player can interact with an NPC this close to its edge. The reach is
// added to the NPC's own so a big creature can be talked to from further off.
const (
	interactBase = 128.0
	// A merchant pays this fraction of an item's price, rounded down, and
	// never less than a copper for something worth anything.
	sellNum, sellDen = 1, 2
)

// npcKind says what a spawn of this class type does when interacted with.
// Subclasses with their own behaviour (smiths, stable masters) are left out
// until something builds what they do.
func npcKind(classType string) string {
	switch {
	case importer.IsMerchantClass(classType):
		return KindMerchant
	case importer.IsStableClass(classType):
		return KindStable
	case importer.IsVaultClass(classType):
		return KindVault
	case importer.IsSmithClass(classType):
		return KindSmith
	case notClassic(classType):
		return KindNotClassic
	case importer.IsHealerClass(classType):
		return KindHealer
	case teaches(classType) != "":
		return KindTrainer
	}
	return ""
}

func notClassic(classType string) bool {
	_, ok := importer.NotClassicOf(classType)
	return ok
}

func interactRange(n *NPC) float64 { return interactBase + n.Reach*0.5 }

func (n *NPC) inReach(p *Player) bool {
	return math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y) <= interactRange(n)
}

// Interact opens what the NPC with this id offers. Every check is the
// server's: the client only names the NPC.
func (z *Zone) Interact(p *Player, id int64) {
	n := z.byID[id]
	if p.dead || n == nil || n.gone || n.dead() {
		p.errMsg("interact", "There is nothing there.")
		return
	}
	if p.ride != nil {
		p.errMsg("interact", "You cannot do that while riding.")
		return
	}
	if n.Kind == "" {
		p.errMsg("interact", "You cannot interact with that.")
		return
	}
	if !n.inReach(p) {
		p.errMsg("interact", "You are too far away to interact with the "+n.Name+".")
		return
	}
	switch n.Kind {
	case KindMerchant:
		z.openShop(p, n)
	case KindStable:
		z.openStable(p, n)
	case KindVault:
		z.openVault(p, n)
	case KindSmith:
		z.openSmith(p, n)
	case KindNotClassic:
		nc, _ := importer.NotClassicOf(n.Spawn.ClassType)
		sys(p, n.title()+" says: \"I cannot help you with that; it is not part of this world.\" ("+nc.Reason+")")
	case KindTrainer:
		z.openTrainer(p, n)
	case KindHealer:
		z.openHealer(p, n)
	}
}

func (z *Zone) openShop(p *Player, n *NPC) {
	cat := z.W.Cat
	list := cat.Stock[n.Spawn.ItemsList]
	msg := protocol.Shop{T: protocol.TShop, NPC: n.ID, Name: n.title(), Items: []protocol.InvItem{}}
	skipped := 0
	for _, id := range list {
		it := cat.Items[id]
		if it == nil {
			skipped++
			continue
		}
		v := itemView(len(msg.Items), it, 1)
		v.Price = it.Price
		msg.Items = append(msg.Items, v)
	}
	if skipped > 0 {
		log.Printf("shop %q (%s): %d of %d stock items are not in the item table", n.Name, n.Spawn.ItemsList, skipped, len(list))
	}
	z.closeVault(p)
	z.closeSmith(p)
	p.shop = n.ID
	p.send(msg)
}

// closeShop shuts the player's open shop, telling the client if there was one.
func (z *Zone) closeShop(p *Player) {
	if p.shop == 0 {
		return
	}
	p.shop = 0
	p.send(protocol.ShopClose{T: protocol.TShopClose})
}

// checkShop closes the open shop when it stops being usable. Run each tick.
func (z *Zone) checkShop(p *Player) {
	if p.shop == 0 {
		return
	}
	n := z.byID[p.shop]
	if p.dead || n == nil || n.gone || n.dead() || !n.inReach(p) {
		z.closeShop(p)
	}
}

// openMerchant returns the merchant whose shop the player has open and is
// still in reach of, closing the shop and answering the player if not.
func (z *Zone) openMerchant(p *Player) *NPC {
	if p.shop == 0 {
		p.errMsg("shop", "You are not talking to a merchant.")
		return nil
	}
	n := z.byID[p.shop]
	if p.dead || n == nil || n.gone || n.dead() || !n.inReach(p) {
		z.closeShop(p)
		p.errMsg("shop", "You are too far away.")
		return nil
	}
	return n
}

// Buy sells the player one of the open shop's items.
func (z *Zone) Buy(p *Player, itemID string) {
	n := z.openMerchant(p)
	if n == nil {
		return
	}
	cat := z.W.Cat
	it := cat.Items[itemID]
	inStock := false
	for _, id := range cat.Stock[n.Spawn.ItemsList] {
		if id == itemID {
			inStock = true
			break
		}
	}
	if it == nil || !inStock {
		p.errMsg("shop", "The "+n.Name+" does not sell that.")
		return
	}
	if p.Char.Money < it.Price {
		p.errMsg("shop", "You cannot afford the "+it.Name+".")
		return
	}
	slot, before := -1, p.Char.Money
	err := z.W.Store.Tx(func(tx *store.Tx) error {
		var err error
		if slot, err = tx.FreeSlot(p.Char.ID, PackSize); err != nil {
			return err
		}
		if slot < 0 {
			return errPackFull
		}
		if err := tx.AddItemCond(p.Char.ID, itemID, slot, 1, z.W.Cat.newCondition(itemID)); err != nil {
			return err
		}
		p.Char.Money -= it.Price
		return tx.SaveCharacter(p.Char)
	})
	if err != nil {
		p.Char.Money = before
		if err == errPackFull {
			p.errMsg("shop", "Your pack is full.")
		} else {
			p.errMsg("shop", "Could not save that; try again.")
		}
		return
	}
	p.inv[slot] = &invItem{ID: itemID, Slot: slot, Count: 1, Cond: z.W.Cat.newCondition(itemID)}
	p.sendInv(cat, p.Char.Money)
}

var errPackFull = errors.New("your pack is full")

// Sell buys the item in a pack slot from the player.
func (z *Zone) Sell(p *Player, slot int) {
	if z.openMerchant(p) == nil {
		return
	}
	if slot < 0 || slot >= PackSize || p.inv[slot] == nil {
		p.errMsg("shop", "There is nothing in that slot to sell.")
		return
	}
	held := p.inv[slot]
	it := z.W.Cat.Items[held.ID]
	if it == nil || it.Price <= 0 {
		p.errMsg("shop", "The merchant has no use for that.")
		return
	}
	count := held.Count
	if count < 1 {
		count = 1
	}
	pay := it.Price * int64(count) * sellNum / sellDen
	if pay < 1 {
		pay = 1
	}
	before := p.Char.Money
	err := z.W.Store.Tx(func(tx *store.Tx) error {
		if err := tx.RemoveItem(p.Char.ID, slot); err != nil {
			return err
		}
		p.Char.Money += pay
		return tx.SaveCharacter(p.Char)
	})
	if err != nil {
		p.Char.Money = before
		p.errMsg("shop", "Could not save that; try again.")
		return
	}
	delete(p.inv, slot)
	z.W.recalcGear(p)
	z.buildBar(p)
	p.sendInv(z.W.Cat, p.Char.Money)
}
