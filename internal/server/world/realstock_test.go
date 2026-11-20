package world

import (
	"database/sql"
	"path/filepath"
	"testing"

	"daocweb/internal/server/rules"
)

// The project's added stock, in the real world store: each row is in its
// merchant's list, is a piece a Midgard player may wear, and a level 45
// Warrior can buy a chain piece of level 41 to 45 and wear it.
func TestRealAddedStockCanBeBoughtAndWorn(t *testing.T) {
	path := filepath.Join("..", "..", "..", "server", "data", "world.db")
	cat, err := LoadCatalog(path)
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Skip(err)
	}
	defer db.Close()
	var cols int
	db.QueryRow(`SELECT count(*) FROM pragma_table_info('merchant_stock') WHERE name='added'`).Scan(&cols)
	if cols == 0 {
		t.Skip("the world store was imported before the added stock")
	}
	rows, err := db.Query(`SELECT list,item FROM merchant_stock WHERE added!=''`)
	if err != nil {
		t.Fatal(err)
	}
	type row struct{ list, item string }
	var added []row
	for rows.Next() {
		var x row
		rows.Scan(&x.list, &x.item)
		added = append(added, x)
	}
	rows.Close()
	if len(added) == 0 {
		t.Fatal("the store has no added stock")
	}
	inList := map[row]bool{}
	for l, ids := range cat.Stock {
		for _, id := range ids {
			inList[row{l, id}] = true
		}
	}
	var chain string
	for _, a := range added {
		it := cat.Items[a.item]
		if it == nil {
			t.Fatalf("added item %s is not in the item table", a.item)
		}
		if !inList[a] {
			t.Errorf("%s is not in the catalog's list %s", a.item, a.list)
		}
		if it.Level < 31 || it.Level > 50 || it.LevelReq > it.Level || (it.Realm != 0 && it.Realm != 2) || it.Price <= 0 {
			t.Errorf("%s: level %d req %d realm %d price %d", it.ID, it.Level, it.LevelReq, it.Realm, it.Price)
		}
		if it.ObjectType == 35 && it.Level >= 41 && it.Level <= 45 && it.ItemType == 25 && chain == "" {
			chain = it.ID
		}
	}
	if chain == "" {
		t.Fatal("no level 41-45 chain hauberk was added")
	}

	r := shopRig(t)
	r.w.Cat = cat
	list := ""
	for _, a := range added {
		if a.item == chain {
			list = a.list
		}
	}
	for _, n := range r.z.npcs {
		if n.ID == merchantID {
			n.Spawn.ItemsList = list
		}
	}
	p := r.player("Thane", 45, 30100, 30000, false)
	p.Char.Class = rules.ClassWarrior
	p.Char.Money = 1 << 40
	r.w.recalc(p)
	r.z.Interact(p, merchantID)
	sent(p)
	r.z.Buy(p, chain)
	if msg := lastErr(t, p); msg != "" {
		t.Fatalf("buying %s: %s", chain, msg)
	}
	if p.inv[0] == nil || p.inv[0].ID != chain {
		t.Fatalf("the hauberk is not in the pack: %+v", p.inv[0])
	}
	if err := canUse(p, cat.Items[chain]); err != nil {
		t.Errorf("a level 45 Warrior cannot wear %s: %v", chain, err)
	}
}
