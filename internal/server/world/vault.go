package world

import (
	"errors"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// KindVault is what a vault keeper does when interacted with: opens the
// character's vault.
const KindVault = "vault"

// A vault has 40 slots, kept as inventory rows numbered from VaultBase so a
// move between pack and vault is one row update in the existing transaction
// code. The pack, the paper doll and loot never see this range: the vault is
// a separate map on the player.
const (
	VaultBase = 1000
	VaultSize = 40
)

func inVault(slot int) bool { return slot >= VaultBase && slot < VaultBase+VaultSize }

// openVault answers a vault keeper's interaction: the character's vault, which
// is the same at every keeper of their realm.
func (z *Zone) openVault(p *Player, n *NPC) {
	if n.Spawn.Realm != 0 && n.Spawn.Realm != p.Arch.Realm {
		sys(p, n.title()+" cannot open a vault for you.")
		return
	}
	z.closeShop(p)
	z.closeSmith(p)
	p.vaultNPC = n.ID
	p.sendVault(z.W.Cat, n)
	p.sendInv(z.W.Cat, p.Char.Money)
}

// sendVault sends a player their own vault: only to them, never to others.
func (p *Player) sendVault(c *Catalog, n *NPC) {
	m := protocol.Vault{T: protocol.TVault, NPC: p.vaultNPC, Items: []protocol.InvItem{}}
	if n != nil {
		m.Name = n.title()
	}
	for slot := VaultBase; slot < VaultBase+VaultSize; slot++ {
		if it := p.vault[slot]; it != nil {
			if x := c.Items[it.ID]; x != nil {
				m.Items = append(m.Items, itemView(slot, x, it.Count))
			}
		}
	}
	p.send(m)
}

// closeVault shuts the player's open vault, telling the client if there was one.
func (z *Zone) closeVault(p *Player) {
	if p.vaultNPC == 0 {
		return
	}
	p.vaultNPC = 0
	p.send(protocol.VaultClose{T: protocol.TVaultClose})
}

// checkVault closes the open vault when its keeper is out of reach or gone.
// Run each tick.
func (z *Zone) checkVault(p *Player) {
	if p.vaultNPC == 0 {
		return
	}
	n := z.byID[p.vaultNPC]
	if p.dead || n == nil || n.gone || n.dead() || !n.inReach(p) {
		z.closeVault(p)
	}
}

// vaultMove moves or swaps an item between pack and vault, or within the vault,
// in one transaction. Worn items must be taken off first.
func (z *Zone) vaultMove(p *Player, from, to int) {
	if from == to {
		return
	}
	n := z.byID[p.vaultNPC]
	if p.vaultNPC == 0 || p.dead || n == nil || n.gone || n.dead() || !n.inReach(p) {
		z.closeVault(p)
		p.errMsg("vault", "Your vault is not open.")
		return
	}
	slotOf := func(s int) (map[int]*invItem, bool) {
		switch {
		case inVault(s):
			return p.vault, true
		case s >= 0 && s < PackSize:
			return p.inv, true
		}
		return nil, false
	}
	srcM, ok1 := slotOf(from)
	dstM, ok2 := slotOf(to)
	if !ok1 || !ok2 {
		p.errMsg("vault", "Take that off before storing it.")
		return
	}
	src, dst := srcM[from], dstM[to]
	if src == nil {
		return
	}
	err := z.W.Store.Tx(func(tx *store.Tx) error {
		if dst != nil {
			return tx.SwapItems(p.Char.ID, from, to)
		}
		return tx.MoveItem(p.Char.ID, from, to)
	})
	if err != nil {
		p.errMsg("vault", errors.New("Could not move that.").Error())
		return
	}
	delete(srcM, from)
	if dst != nil {
		srcM[from] = dst
		dst.Slot = from
	}
	dstM[to] = src
	src.Slot = to
	p.sendVault(z.W.Cat, n)
	p.sendInv(z.W.Cat, p.Char.Money)
}
