package world

import (
	"encoding/json"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// A bought or dropped item starts whole, or at its template's condition when
// the template gives one; the condition is saved with the row.
func TestNewItemsStartAtTheirCondition(t *testing.T) {
	r := newRig(t)
	r.w.Cat.Items["worn_axe"] = &Item{ID: "worn_axe", Name: "worn axe", Condition: 0.6, Droppable: true}
	p := r.player("Alice", 5, 30000, 30000, false)
	whole, _ := r.w.addItem(p, "hammer", 1)
	worn, _ := r.w.addItem(p, "worn_axe", 1)
	if p.inv[whole].Cond != 10000 || p.inv[worn].Cond != 6000 {
		t.Fatalf("conditions %d and %d, want 10000 and 6000", p.inv[whole].Cond, p.inv[worn].Cond)
	}
	again := newPlayer(p.Acct, p.Char)
	r.w.loadInventory(again)
	if again.inv[whole].Cond != 10000 || again.inv[worn].Cond != 6000 {
		t.Errorf("after a reload %d and %d, want them saved", again.inv[whole].Cond, again.inv[worn].Cond)
	}
}

func conditionRig(t *testing.T) (*rig, *Player) {
	t.Helper()
	smith := Spawn{ID: 970, Name: "Smith", Model: 56, X: 30000, Y: 30000, Z: 5000, Level: 10, Size: 50, Speed: 191, Realm: 2, ClassType: "DOL.GS.Blacksmith"}
	r := newRig(t, smith)
	r.w.Cat.Items["hammer"].Price = 200
	p := r.player("Alice", 5, 30080, 30000, true) // a hammer, shield and armour worn
	return r, p
}

func TestWearTwoHundredBlows(t *testing.T) {
	r, p := conditionRig(t)
	for i := 0; i < 200; i++ {
		r.z.wearItem(p, SlotWeapon)
	}
	if p.inv[SlotWeapon].Cond != 9000 {
		t.Errorf("weapon at %d after 200 blows, want 9000", p.inv[SlotWeapon].Cond)
	}
}

func TestBlockWearsTheShieldNotTheArmour(t *testing.T) {
	r, p := conditionRig(t)
	r.z.wearBlow(p, rules.Outcome{Result: rules.Blocked, Location: rules.LocTorso})
	if p.inv[SlotShield].Cond != 9995 || p.inv[SlotTorso].Cond != 10000 {
		t.Errorf("shield %d torso %d, want 9995 and 10000", p.inv[SlotShield].Cond, p.inv[SlotTorso].Cond)
	}
	r.z.wearBlow(p, rules.Outcome{Result: rules.Hit, Location: rules.LocTorso})
	if p.inv[SlotTorso].Cond != 9995 || p.inv[SlotShield].Cond != 9995 {
		t.Errorf("after a hit torso %d shield %d, want 9995 and 9995", p.inv[SlotTorso].Cond, p.inv[SlotShield].Cond)
	}
}

func TestWearLowersEffectiveness(t *testing.T) {
	r, p := conditionRig(t)
	p.inv[SlotWeapon].Cond = 9000
	p.inv[SlotTorso].Cond = 5000
	cb := p.combatant(r.w.Cat)
	if cb.Weapon.Condition != 0.9 {
		t.Errorf("weapon condition %v in combat, want 0.9", cb.Weapon.Condition)
	}
	if got := cb.Armour[rules.LocTorso]; got == nil || got.Condition != 0.5 {
		t.Errorf("torso piece %+v, want condition 0.5", got)
	}
}

func TestWearWarnsOncePerThreshold(t *testing.T) {
	r, p := conditionRig(t)
	p.inv[SlotWeapon].Cond = 9002
	drain(p)
	r.z.wearItem(p, SlotWeapon) // 9002 -> 8997
	out, _, _ := inbox(p)
	if !has(out, "below 90%") {
		t.Fatalf("told %v, want the 90%% warning", out)
	}
	r.z.wearItem(p, SlotWeapon)
	if out, _, _ := inbox(p); len(out) != 0 {
		t.Errorf("told %v on the next blow, want silence", out)
	}
}

func TestWearIsSavedWithTheCharacter(t *testing.T) {
	r, p := conditionRig(t)
	for i := 0; i < 20; i++ {
		r.z.wearItem(p, SlotWeapon)
	}
	r.w.savePlayer(p)
	again := newPlayer(p.Acct, p.Char)
	r.w.loadInventory(again)
	if again.inv[SlotWeapon].Cond != 9900 {
		t.Errorf("reloaded weapon at %d, want 9900", again.inv[SlotWeapon].Cond)
	}
}

func TestNewItemShowsItsConditionInTheTooltip(t *testing.T) {
	_, p := conditionRig(t)
	drain(p)
	p.sendInv(p.Zone.W.Cat, p.Char.Money)
	_, raws := sent(p)
	var inv protocol.Inv
	for _, b := range raws {
		json.Unmarshal(b, &inv)
	}
	found := false
	for _, it := range inv.Items {
		if it.Slot == SlotWeapon {
			found = strings.Contains(it.Detail, "100% condition")
		}
	}
	if !found {
		t.Errorf("the worn weapon's detail has no condition: %+v", inv.Items)
	}
}

func TestBlacksmithRepairsForMoney(t *testing.T) {
	r, p := conditionRig(t)
	p.inv[SlotWeapon].Cond = 8000
	p.Char.Money = 100
	r.z.Interact(p, 970)
	types, raws := sent(p)
	var smith protocol.Smith
	for i, ty := range types {
		if ty == protocol.TSmith {
			json.Unmarshal(raws[i], &smith)
		}
	}
	if len(smith.Items) != 1 || smith.Items[0].Slot != SlotWeapon || smith.Items[0].Price != 20 {
		t.Fatalf("smith offers %+v, want the weapon at 20 copper", smith.Items)
	}
	r.z.Repair(p, &protocol.Repair{Slot: SlotWeapon})
	types, raws = sent(p)
	var quote protocol.RepairQuote
	for i, ty := range types {
		if ty == protocol.TRepairQuote {
			json.Unmarshal(raws[i], &quote)
		}
	}
	if quote.Price != 20 || p.inv[SlotWeapon].Cond != 8000 || p.Char.Money != 100 {
		t.Fatalf("quote %+v, weapon %d, money %d: want a quote and nothing taken yet", quote, p.inv[SlotWeapon].Cond, p.Char.Money)
	}
	r.z.Repair(p, &protocol.Repair{Slot: SlotWeapon, Accept: true, Price: quote.Price})
	if p.inv[SlotWeapon].Cond != 10000 || p.Char.Money != 80 {
		t.Errorf("after accepting: weapon %d money %d, want 10000 and 80", p.inv[SlotWeapon].Cond, p.Char.Money)
	}
	again := newPlayer(p.Acct, p.Char)
	r.w.loadInventory(again)
	if again.inv[SlotWeapon].Cond != 10000 {
		t.Error("the repair was not saved")
	}
}

func TestBlacksmithRefusesWithoutMoney(t *testing.T) {
	r, p := conditionRig(t)
	p.inv[SlotWeapon].Cond = 8000
	p.Char.Money = 5
	r.z.Interact(p, 970)
	drain(p)
	r.z.Repair(p, &protocol.Repair{Slot: SlotWeapon, Accept: true, Price: 20})
	if lastErr(t, p) == "" || p.inv[SlotWeapon].Cond != 8000 || p.Char.Money != 5 {
		t.Errorf("weapon %d money %d, want nothing changed and a refusal", p.inv[SlotWeapon].Cond, p.Char.Money)
	}
}
