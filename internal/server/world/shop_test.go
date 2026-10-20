package world

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/importer"
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

const merchantID = 900

// shopRig is a zone with one merchant at (30000, 30000) selling a hammer, a
// shield and an item the item table does not have, and a guard that does
// nothing when interacted with.
func shopRig(t *testing.T) *rig {
	t.Helper()
	merchant := Spawn{ID: merchantID, Name: "Smith", Guild: "Weapons", Model: 56, X: 30000, Y: 30000, Z: 5000, Level: 10, Size: 50,
		Speed: 191, Realm: 2, ClassType: "DOL.GS.GameMerchant", ItemsList: "L1"}
	guard := Spawn{ID: merchantID + 1, Name: "Guard", Model: 56, X: 30050, Y: 30000, Z: 5000, Level: 10, Size: 50,
		Speed: 191, Realm: 2, ClassType: "DOL.GS.GameGuard"}
	r := newRig(t, merchant, guard)
	r.w.Cat.Stock = map[string][]string{"L1": {"hammer", "ghost", "shield"}}
	r.w.Cat.Items["free"] = &Item{ID: "free", Name: "trinket", Droppable: true}
	return r
}

// sent drains everything queued for a player as (type, raw) pairs.
func sent(p *Player) (types []string, raws [][]byte) {
	for {
		select {
		case b := <-p.out:
			var env protocol.Envelope
			json.Unmarshal(b, &env)
			types = append(types, env.T)
			raws = append(raws, b)
		default:
			return
		}
	}
}

func lastErr(t *testing.T, p *Player) string {
	t.Helper()
	types, raws := sent(p)
	for i := len(types) - 1; i >= 0; i-- {
		if types[i] == protocol.TErr {
			var e protocol.Err
			json.Unmarshal(raws[i], &e)
			return e.Text
		}
	}
	return ""
}

func lastShop(p *Player) (protocol.Shop, bool) {
	types, raws := sent(p)
	for i := len(types) - 1; i >= 0; i-- {
		if types[i] == protocol.TShop {
			var s protocol.Shop
			json.Unmarshal(raws[i], &s)
			return s, true
		}
	}
	return protocol.Shop{}, false
}

func savedMoney(t *testing.T, r *rig, p *Player) int64 {
	t.Helper()
	c, err := r.st.Character(p.Acct.ID, p.Char.ID)
	if err != nil {
		t.Fatal(err)
	}
	return c.Money
}

func savedItems(t *testing.T, r *rig, p *Player) map[int]string {
	t.Helper()
	items, err := r.st.Items(p.Char.ID)
	if err != nil {
		t.Fatal(err)
	}
	m := map[int]string{}
	for _, it := range items {
		m[it.Slot] = it.ItemID
	}
	return m
}

func TestNPCKindComesFromClassType(t *testing.T) {
	if npcKind("DOL.GS.GameMerchant") != KindMerchant {
		t.Error("GameMerchant is a merchant")
	}
	for _, c := range []string{"", "DOL.GS.GameGuard", "DOL.GS.GameNPC"} {
		if k := npcKind(c); k != "" {
			t.Errorf("npcKind(%q) = %q, want none", c, k)
		}
	}
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, false)
	r.ticks(2)
	s := snaps(p)
	byName := map[string]string{}
	for _, e := range s[len(s)-1].Ents {
		byName[e.Name] = e.Kind
	}
	if byName["Smith"] != "merchant" || byName["Guard"] != "" {
		t.Errorf("entity kinds = %v", byName)
	}
}

func TestInteractOpensShopWithStockInOrder(t *testing.T) {
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, false)
	r.z.Interact(p, merchantID)
	shop, ok := lastShop(p)
	if !ok {
		t.Fatal("no shop message")
	}
	if shop.NPC != merchantID || len(shop.Items) != 2 {
		t.Fatalf("shop = %+v", shop)
	}
	if shop.Items[0].ID != "hammer" || shop.Items[1].ID != "shield" || shop.Items[0].Price != 10 || shop.Items[0].Kind != "weapon" {
		t.Errorf("items = %+v", shop.Items)
	}
	if p.shop != merchantID {
		t.Error("the server must remember the open shop")
	}
}

func TestInteractIsRefusedWhenItShouldBe(t *testing.T) {
	r := shopRig(t)
	far := r.player("Faraway", 5, 30000+interactBase+400, 30000, false)
	r.z.Interact(far, merchantID)
	if msg := lastErr(t, far); !strings.Contains(msg, "too far") || far.shop != 0 {
		t.Errorf("far: err %q shop %d", msg, far.shop)
	}
	near := r.player("Nearby", 5, 30100, 30000, false)
	r.z.Interact(near, merchantID+1) // a guard
	if msg := lastErr(t, near); msg == "" || near.shop != 0 {
		t.Errorf("guard: err %q shop %d", msg, near.shop)
	}
	r.z.Interact(near, 424242) // no such NPC
	if msg := lastErr(t, near); msg == "" {
		t.Error("an unknown id must be refused")
	}
	near.dead = true
	r.z.Interact(near, merchantID)
	if msg := lastErr(t, near); msg == "" || near.shop != 0 {
		t.Errorf("dead player: err %q shop %d", msg, near.shop)
	}
}

func TestBuyTakesMoneyAndGivesTheItem(t *testing.T) {
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, false)
	p.Char.Money = 25
	r.z.Interact(p, merchantID)
	sent(p)
	r.z.Buy(p, "hammer")
	if p.Char.Money != 15 || p.inv[0] == nil || p.inv[0].ID != "hammer" {
		t.Fatalf("money %d inv %+v", p.Char.Money, p.inv[0])
	}
	if savedMoney(t, r, p) != 15 || savedItems(t, r, p)[0] != "hammer" {
		t.Errorf("the purchase was not saved: money %d items %v", savedMoney(t, r, p), savedItems(t, r, p))
	}
	types, _ := sent(p)
	if len(types) == 0 || types[len(types)-1] != protocol.TInv {
		t.Errorf("inv was not resent: %v", types)
	}
}

func TestBuyRefusals(t *testing.T) {
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, false)

	r.z.Buy(p, "hammer") // no shop open
	if lastErr(t, p) == "" {
		t.Error("buying with no shop open must be refused")
	}

	r.z.Interact(p, merchantID)
	sent(p)
	p.Char.Money = 9
	r.z.Buy(p, "hammer")
	if msg := lastErr(t, p); !strings.Contains(msg, "afford") || p.Char.Money != 9 || len(p.inv) != 0 {
		t.Errorf("too poor: %q money %d inv %v", msg, p.Char.Money, p.inv)
	}

	p.Char.Money = 100
	r.z.Buy(p, "pelt") // exists, but this merchant does not sell it
	if msg := lastErr(t, p); msg == "" || p.Char.Money != 100 || len(p.inv) != 0 {
		t.Errorf("not in stock: %q money %d", msg, p.Char.Money)
	}
	r.z.Buy(p, "ghost") // in the list, not in the item table
	if msg := lastErr(t, p); msg == "" || p.Char.Money != 100 {
		t.Errorf("unknown item: %q money %d", msg, p.Char.Money)
	}

	// A full pack takes no money.
	r.st.Tx(func(tx *store.Tx) error {
		for i := 0; i < PackSize; i++ {
			tx.AddItem(p.Char.ID, "pelt", i, 1)
		}
		return nil
	})
	r.z.Buy(p, "hammer")
	if msg := lastErr(t, p); !strings.Contains(msg, "pack is full") || p.Char.Money != 100 || savedMoney(t, r, p) == 90 {
		t.Errorf("full pack: %q money %d saved %d", msg, p.Char.Money, savedMoney(t, r, p))
	}
}

func TestBuyAfterWalkingAwayIsRefusedAndClosesTheShop(t *testing.T) {
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, false)
	p.Char.Money = 100
	r.z.Interact(p, merchantID)
	sent(p)
	p.Body.X = 30000 + interactBase + 600
	r.z.Buy(p, "hammer")
	types, _ := sent(p)
	closed := false
	for _, ty := range types {
		closed = closed || ty == protocol.TShopClose
	}
	if !closed || p.shop != 0 || p.Char.Money != 100 || len(p.inv) != 0 {
		t.Errorf("types %v shop %d money %d inv %v", types, p.shop, p.Char.Money, p.inv)
	}
}

func TestSellPaysHalfAndRemovesTheItem(t *testing.T) {
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, false)
	r.w.addItem(p, "hammer", 1) // slot 0, price 10
	r.w.addItem(p, "free", 1)   // slot 1, worth nothing
	r.z.Interact(p, merchantID)
	sent(p)

	r.z.Sell(p, 0)
	if p.Char.Money != 5 || p.inv[0] != nil {
		t.Fatalf("money %d slot0 %+v", p.Char.Money, p.inv[0])
	}
	if _, ok := savedItems(t, r, p)[0]; ok || savedMoney(t, r, p) != 5 {
		t.Errorf("the sale was not saved: money %d items %v", savedMoney(t, r, p), savedItems(t, r, p))
	}

	r.z.Sell(p, 1) // priced zero
	if msg := lastErr(t, p); msg == "" || p.inv[1] == nil || p.Char.Money != 5 {
		t.Errorf("worthless: %q money %d", msg, p.Char.Money)
	}
	r.z.Sell(p, 5) // empty
	if msg := lastErr(t, p); msg == "" {
		t.Error("an empty slot must be refused")
	}
	r.z.Sell(p, SlotWeapon) // a worn slot is not a pack slot
	if msg := lastErr(t, p); msg == "" {
		t.Error("a worn slot must be refused")
	}
}

func TestSellOneCopperMinimum(t *testing.T) {
	r := shopRig(t)
	r.w.Cat.Items["bead"] = &Item{ID: "bead", Name: "bead", Price: 1, Droppable: true}
	p := r.player("Alice", 5, 30100, 30000, false)
	r.w.addItem(p, "bead", 1)
	r.z.Interact(p, merchantID)
	r.z.Sell(p, 0)
	if p.Char.Money != 1 {
		t.Errorf("money = %d, a one-copper item must still pay a copper", p.Char.Money)
	}
}

func TestShopClosesWhenThePlayerWalksOffOrDies(t *testing.T) {
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, false)
	r.z.Interact(p, merchantID)
	sent(p)
	r.ticks(2)
	if p.shop == 0 {
		t.Fatal("standing still must keep the shop open")
	}
	p.Body.X = 30000 + interactBase + 600
	r.ticks(2)
	types, _ := sent(p)
	closed := false
	for _, ty := range types {
		closed = closed || ty == protocol.TShopClose
	}
	if !closed || p.shop != 0 {
		t.Errorf("walking off: types %v shop %d", types, p.shop)
	}

	p.Body.X = 30100
	r.z.Interact(p, merchantID)
	sent(p)
	r.z.killPlayer(p, nil)
	if p.shop != 0 {
		t.Error("dying must close the shop")
	}
}

func TestSecondShopReplacesTheFirst(t *testing.T) {
	r := shopRig(t)
	r.w.Cat.Stock["L2"] = []string{"axe"}
	r.z.byID[merchantID+1].Kind = KindMerchant
	r.z.byID[merchantID+1].Spawn.ItemsList = "L2"
	p := r.player("Alice", 5, 30100, 30000, false)
	r.z.Interact(p, merchantID)
	r.z.Interact(p, merchantID+1)
	shop, _ := lastShop(p)
	if p.shop != merchantID+1 || len(shop.Items) != 1 || shop.Items[0].ID != "axe" {
		t.Errorf("shop %d items %+v", p.shop, shop.Items)
	}
}

// A world.db imported before merchants has neither the stock table nor the
// class type column: the server must still load it, with interaction off.
func TestCatalogLoadsStoresWithAndWithoutMerchantData(t *testing.T) {
	build := func(old bool) *Catalog {
		path := filepath.Join(t.TempDir(), "world.db")
		db, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(importer.Schema); err != nil {
			t.Fatal(err)
		}
		stmts := []string{
			`INSERT INTO zone VALUES(100,100,'Test',0,0,1,1,0,2,1)`,
			`INSERT INTO spawn(zone,name,model,x,y,z,heading,level,size,speed,template,roam,respawn_s,aggro_level,aggro_range,realm,flags,classtype,items_list) VALUES(100,'Smith',56,10,10,0,0,1,50,191,0,0,30,0,0,2,0,'DOL.GS.GameMerchant','L1')`,
			`INSERT INTO merchant_stock VALUES('L1','b',1,0),('L1','a',0,2),('L1','c',0,1)`,
		}
		if old {
			stmts = append(stmts[:1], `INSERT INTO spawn(zone,name,model,x,y,z,heading,level,size,speed,template,roam,respawn_s,aggro_level,aggro_range,realm,flags) VALUES(100,'Smith',56,10,10,0,0,1,50,191,0,0,30,0,0,2,0)`)
			stmts = append([]string{`DROP TABLE merchant_stock`, `ALTER TABLE spawn DROP COLUMN classtype`, `ALTER TABLE spawn DROP COLUMN items_list`}, stmts...)
		}
		for _, s := range stmts {
			if _, err := db.Exec(s); err != nil {
				t.Fatalf("%s: %v", s, err)
			}
		}
		db.Close()
		c, err := LoadCatalog(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(c.CloseDB)
		return c
	}

	c := build(false)
	if !c.Shops() {
		t.Error("a current store has merchant data")
	}
	if got := strings.Join(c.Stock["L1"], ","); got != "c,a,b" {
		t.Errorf("stock order = %s, want page then slot: c,a,b", got)
	}
	sp, err := c.SpawnsIn(100)
	if err != nil || len(sp) != 1 || sp[0].ClassType != "DOL.GS.GameMerchant" || sp[0].ItemsList != "L1" || newNPC(sp[0], c, rand.New(rand.NewSource(1))).Kind != KindMerchant {
		t.Errorf("spawns %+v, %v", sp, err)
	}

	old := build(true)
	if old.Shops() {
		t.Error("an old store has no merchant data")
	}
	sp, err = old.SpawnsIn(100)
	if err != nil || len(sp) != 1 || newNPC(sp[0], old, rand.New(rand.NewSource(1))).Kind != "" {
		t.Errorf("old store spawns %+v, %v", sp, err)
	}
}
