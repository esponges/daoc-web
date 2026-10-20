package world

import (
	"encoding/json"
	"testing"

	"daocweb/internal/server/protocol"
)

// vaultRig is a zone with two vault keepers far apart, as if in two towns.
func vaultRig(t *testing.T) *rig {
	t.Helper()
	keep := func(id int64, x float64) Spawn {
		return Spawn{ID: id, Name: "Keeper", Model: 56, X: x, Y: 30000, Z: 5000, Level: 10, Size: 50, Speed: 191, Realm: 2, ClassType: "DOL.GS.GameVaultKeeper"}
	}
	return newRig(t, keep(960, 30000), keep(961, 45000))
}

// lastVault is the latest vault message a player was sent, and whether any was.
func lastVault(p *Player) (protocol.Vault, bool) {
	types, raws := sent(p)
	var v protocol.Vault
	found := false
	for i, ty := range types {
		if ty == protocol.TVault {
			json.Unmarshal(raws[i], &v)
			found = true
		}
	}
	return v, found
}

func hasVaultItem(v protocol.Vault, id string) bool {
	for _, it := range v.Items {
		if it.ID == id {
			return true
		}
	}
	return false
}

func packHas(p *Player, id string) bool {
	for _, it := range p.inv {
		if it.ID == id {
			return true
		}
	}
	return false
}

func TestVaultOpensAndShowsOnlyYours(t *testing.T) {
	r := vaultRig(t)
	a := r.player("Alice", 5, 30080, 30000, false)
	b := r.player("Bobby", 5, 30090, 30000, false)
	r.w.addItem(a, "hammer", 1)
	r.w.addItem(b, "shield", 1)
	drain(a)
	drain(b)
	r.z.Interact(a, 960)
	r.z.Interact(b, 960)
	va, ok := lastVault(a)
	if !ok || len(va.Items) != 0 {
		t.Fatalf("Alice's vault %+v (%v), want it open and empty", va, ok)
	}
	r.z.MoveItem(a, packSlotOf(a, "hammer"), VaultBase+3)
	r.z.MoveItem(b, packSlotOf(b, "shield"), VaultBase)
	va, _ = lastVault(a)
	vb, _ := lastVault(b)
	if !hasVaultItem(va, "hammer") || hasVaultItem(va, "shield") || !hasVaultItem(vb, "shield") || hasVaultItem(vb, "hammer") {
		t.Errorf("Alice sees %+v, Bobby sees %+v: each should see only their own", va.Items, vb.Items)
	}
}

func packSlotOf(p *Player, id string) int {
	for s, it := range p.inv {
		if it.ID == id && s < PackSize {
			return s
		}
	}
	return -1
}

func TestVaultStoreTakeAndRestart(t *testing.T) {
	r := vaultRig(t)
	p := r.player("Alice", 5, 30080, 30000, false)
	r.w.addItem(p, "hammer", 1)
	r.z.Interact(p, 960)
	from := packSlotOf(p, "hammer")
	r.z.MoveItem(p, from, VaultBase+7)
	if packHas(p, "hammer") || p.vault[VaultBase+7] == nil {
		t.Fatal("the hammer is not in the vault after storing it")
	}
	// A restart: the character is loaded again from the store.
	again := newPlayer(p.Acct, p.Char)
	r.w.loadInventory(again)
	if again.vault[VaultBase+7] == nil || again.vault[VaultBase+7].ID != "hammer" || packHas(again, "hammer") {
		t.Fatalf("after a restart the vault holds %v and the pack %v, want the hammer in the vault only", again.vault, again.inv)
	}
	r.z.MoveItem(p, VaultBase+7, 5)
	if !packHas(p, "hammer") || len(p.vault) != 0 {
		t.Errorf("after taking it back the pack %v vault %v", p.inv, p.vault)
	}
	later := newPlayer(p.Acct, p.Char)
	r.w.loadInventory(later)
	if !packHas(later, "hammer") || len(later.vault) != 0 {
		t.Error("after another restart the hammer is not back in the pack")
	}
}

func TestVaultOccupiedSlotSwaps(t *testing.T) {
	r := vaultRig(t)
	p := r.player("Alice", 5, 30080, 30000, false)
	r.w.addItem(p, "hammer", 1)
	r.w.addItem(p, "shield", 1)
	r.z.Interact(p, 960)
	r.z.MoveItem(p, packSlotOf(p, "hammer"), VaultBase)
	slot := packSlotOf(p, "shield")
	r.z.MoveItem(p, VaultBase, slot) // the hammer onto the shield's slot
	if p.inv[slot] == nil || p.inv[slot].ID != "hammer" || p.vault[VaultBase] == nil || p.vault[VaultBase].ID != "shield" {
		t.Errorf("pack slot %d %v, vault %v; want the two swapped", slot, p.inv[slot], p.vault[VaultBase])
	}
	again := newPlayer(p.Acct, p.Char)
	r.w.loadInventory(again)
	if again.vault[VaultBase] == nil || again.vault[VaultBase].ID != "shield" || again.inv[slot].ID != "hammer" {
		t.Error("the swap was not saved")
	}
}

func TestVaultWalkingAwayClosesIt(t *testing.T) {
	r := vaultRig(t)
	p := r.player("Alice", 5, 30080, 30000, false)
	r.w.addItem(p, "hammer", 1)
	r.z.Interact(p, 960)
	drain(p)
	p.Body.X = 38000
	r.ticks(2)
	types, _ := sent(p)
	closed := false
	for _, ty := range types {
		closed = closed || ty == protocol.TVaultClose
	}
	if !closed || p.vaultNPC != 0 {
		t.Error("the vault stayed open after walking away")
	}
	r.z.MoveItem(p, packSlotOf(p, "hammer"), VaultBase)
	if !packHas(p, "hammer") || len(p.vault) != 0 {
		t.Error("a move into a closed vault went through")
	}
}

func TestVaultSameAtAnotherKeeperAndNoWornItems(t *testing.T) {
	r := vaultRig(t)
	p := r.player("Alice", 5, 30080, 30000, true) // equipped: a hammer and shield worn
	r.w.addItem(p, "hammer", 1)
	r.z.Interact(p, 960)
	r.z.MoveItem(p, SlotWeapon, VaultBase+1)
	if p.vault[VaultBase+1] != nil || p.inv[SlotWeapon] == nil {
		t.Error("a worn weapon went into the vault")
	}
	r.z.MoveItem(p, packSlotOf(p, "hammer"), VaultBase+2)
	p.Body.X = 44950
	r.ticks(2)
	drain(p)
	r.z.Interact(p, 961)
	v, ok := lastVault(p)
	if !ok || !hasVaultItem(v, "hammer") {
		t.Errorf("the vault at the other keeper %+v (%v), want the stored item", v, ok)
	}
}
