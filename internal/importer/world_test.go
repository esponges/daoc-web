package importer

import (
	"database/sql"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// repo returns the paths of the checkout and converted zones, skipping the
// test when this machine has neither (they are never in the repository).
func repo(t *testing.T) (dump, web string) {
	t.Helper()
	dump = filepath.Join("..", "..", "server", "data", "community", "opendaoc-db-core")
	web = filepath.Join("..", "..", "web", "data")
	if _, err := os.Stat(filepath.Join(dump, "mob.sql")); err != nil {
		t.Skip("no community database checked out in server/data/community")
	}
	if _, err := os.Stat(filepath.Join(web, "zone100", "heights.u16")); err != nil {
		t.Skip("zone 100 is not converted")
	}
	return
}

func TestImportZone100(t *testing.T) {
	if testing.Short() {
		t.Skip("imports the whole dump")
	}
	dump, web := repo(t)
	out := filepath.Join(t.TempDir(), "world.db")
	rep, err := Run(Options{Dump: dump, Out: out, WebDir: web})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Missing) > 0 {
		t.Fatalf("missing tables: %v", rep.Missing)
	}
	db, err := sql.Open("sqlite", out)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var n int
	db.QueryRow(`SELECT count(*) FROM spawn WHERE zone=100`).Scan(&n)
	if n < 1500 {
		t.Errorf("zone 100 kept %d spawns, want over 1500", n)
	}
	// Every kept spawn is within MaxDz of the ground, and in the zone.
	hb, _ := os.ReadFile(filepath.Join(web, "zone100", "heights.u16"))
	rows, err := db.Query(`SELECT x,y,z FROM spawn WHERE zone=100`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	bad := 0
	for rows.Next() {
		var x, y, z float64
		rows.Scan(&x, &y, &z)
		if x < 0 || y < 0 || x >= 65536 || y >= 65536 {
			t.Fatalf("spawn outside the zone: %v,%v", x, y)
		}
		g := float64(binary.LittleEndian.Uint16(hb[(int(y/256)*256+int(x/256))*2:]))
		if d := z - g; d > 500 || d < -500 {
			bad++
		}
	}
	if bad > 0 {
		t.Errorf("%d spawns more than 500 units from the ground", bad)
	}
	// A wolf is model 56, which the client draws as its Large Grey Wolf.
	db.QueryRow(`SELECT count(*) FROM spawn WHERE zone=100 AND model=56`).Scan(&n)
	if n == 0 {
		t.Error("no model-56 spawns in Mularn")
	}
}

func TestMissingTableReported(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(Options{Dump: dir, Out: filepath.Join(t.TempDir(), "w.db")}); err == nil {
		t.Error("an empty dump directory must be an error")
	}
}

// dumpTable writes one table in the community dump's shape: a REPLACE with a
// column list, then tuples.
func dumpTable(t *testing.T, dir, table, cols string, rows ...string) {
	t.Helper()
	body := "REPLACE INTO `" + table + "` (" + cols + ") VALUES\n"
	for i, r := range rows {
		sep := ","
		if i == len(rows)-1 {
			sep = ";"
		}
		body += "\t(" + r + ")" + sep + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, table+".sql"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestImportLinksNPCEquipment(t *testing.T) {
	dir := t.TempDir()
	dumpTable(t, dir, "zones", "`ZoneID`, `RegionID`, `OffsetX`, `OffsetY`, `Width`, `Height`, `Name`, `WaterLevel`, `Realm`",
		"1, 1, 0, 0, 1, 1, 'Test', 0, 1")
	dumpTable(t, dir, "npcequipment", "`TemplateID`, `Slot`, `Model`, `Color`, `Effect`, `Extension`, `Emblem`",
		"'set-a', 25, 662, 17, 0, 0, 0", "'set-a', 10, 311, 0, 0, 0, 0", "'set-b', 27, 47, 0, 0, 0, 0")
	dumpTable(t, dir, "npctemplate", "`TemplateId`, `Name`, `Model`, `Size`, `Level`, `EquipmentTemplateID`",
		"7, 'tmpl', '10', '50', '5', 'set-b'")
	dumpTable(t, dir, "mob", "`Name`, `Region`, `X`, `Y`, `Z`, `Model`, `NPCTemplateID`, `EquipmentTemplateID`",
		"'own', 1, 100, 100, 0, 10, 0, 'set-a'",
		"'inherits', 1, 200, 100, 0, 10, 7, ''",
		"'bare', 1, 300, 100, 0, 10, 0, ''",
		"'dangling', 1, 400, 100, 0, 10, 0, 'nope'")
	for _, tb := range []string{"loottemplate", "mobxloottemplate", "itemtemplate", "style", "specialization",
		"classxspecialization", "race", "zonepoint", "startuplocation", "bindpoint", "merchantitem"} {
		dumpTable(t, dir, tb, "`Unused`", "0")
	}
	out := filepath.Join(t.TempDir(), "world.db")
	rep, err := Run(Options{Dump: dir, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if c := rep.Tables["npcequipment"]; c != [2]int{3, 3} {
		t.Errorf("npcequipment read/kept = %v", c)
	}
	if rep.UnknownEquip != 1 {
		t.Errorf("UnknownEquip = %d, want 1", rep.UnknownEquip)
	}
	db, err := sql.Open("sqlite", out)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	want := map[string]string{"own": "set-a", "inherits": "set-b", "bare": "", "dangling": ""}
	for name, eq := range want {
		var got sql.NullString
		if err := db.QueryRow(`SELECT equip FROM spawn WHERE name=?`, name).Scan(&got); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got.String != eq {
			t.Errorf("%s equip = %q, want %q", name, got.String, eq)
		}
	}
	var n int
	db.QueryRow(`SELECT count(*) FROM npc_equip WHERE template='set-a'`).Scan(&n)
	if n != 2 {
		t.Errorf("set-a has %d rows, want 2", n)
	}
}

func TestImportKeepsMerchantsAndStock(t *testing.T) {
	dir := t.TempDir()
	dumpTable(t, dir, "zones", "`ZoneID`, `RegionID`, `OffsetX`, `OffsetY`, `Width`, `Height`, `Name`, `WaterLevel`, `Realm`",
		"1, 1, 0, 0, 1, 1, 'Test', 0, 1")
	dumpTable(t, dir, "mob", "`Name`, `Region`, `X`, `Y`, `Z`, `Model`, `ClassType`, `ItemsListTemplateID`",
		"'Smith', 1, 100, 100, 0, 10, 'DOL.GS.GameMerchant', 'L1'",
		"'Guard', 1, 200, 100, 0, 10, 'DOL.GS.GameGuard', ''",
		"'Lost', 99, 100, 100, 0, 10, 'DOL.GS.GameMerchant', 'L1'")
	dumpTable(t, dir, "merchantitem", "`ItemListID`, `ItemTemplateID`, `PageNumber`, `SlotPosition`",
		"'L1', 'sword', 0, 1", "'L1', 'shield', 1, 0", "'', 'orphan', 0, 0")
	for _, tb := range []string{"npcequipment", "npctemplate", "loottemplate", "mobxloottemplate", "itemtemplate", "style",
		"specialization", "classxspecialization", "race", "zonepoint", "startuplocation", "bindpoint"} {
		dumpTable(t, dir, tb, "`Unused`", "0")
	}
	out := filepath.Join(t.TempDir(), "world.db")
	rep, err := Run(Options{Dump: dir, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if rep.MerchantsSeen != 2 || rep.MerchantsKept != 1 {
		t.Errorf("merchants seen/kept = %d/%d, want 2/1", rep.MerchantsSeen, rep.MerchantsKept)
	}
	if c := rep.Tables["merchantitem"]; c != [2]int{3, 2} {
		t.Errorf("merchantitem read/kept = %v, want [3 2]", c)
	}
	db, err := sql.Open("sqlite", out)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var ct, list string
	if err := db.QueryRow(`SELECT classtype, items_list FROM spawn WHERE name='Smith'`).Scan(&ct, &list); err != nil {
		t.Fatal(err)
	}
	if ct != "DOL.GS.GameMerchant" || list != "L1" {
		t.Errorf("Smith = %q %q", ct, list)
	}
	var n int
	db.QueryRow(`SELECT count(*) FROM merchant_stock WHERE list='L1'`).Scan(&n)
	if n != 2 {
		t.Errorf("L1 stock rows = %d, want 2", n)
	}
}

func TestMissingMerchantStockIsReported(t *testing.T) {
	dir := t.TempDir()
	dumpTable(t, dir, "zones", "`ZoneID`, `RegionID`, `OffsetX`, `OffsetY`, `Width`, `Height`, `Name`, `WaterLevel`, `Realm`",
		"1, 1, 0, 0, 1, 1, 'Test', 0, 1")
	dumpTable(t, dir, "merchantitem", "`ItemListID`, `ItemTemplateID`, `PageNumber`, `SlotPosition`", "'', '', 0, 0")
	for _, tb := range []string{"mob", "npcequipment", "npctemplate", "loottemplate", "mobxloottemplate", "itemtemplate", "style",
		"specialization", "classxspecialization", "race", "zonepoint", "startuplocation", "bindpoint"} {
		dumpTable(t, dir, tb, "`Unused`", "0")
	}
	rep, err := Run(Options{Dump: dir, Out: filepath.Join(t.TempDir(), "w.db")})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range rep.Missing {
		if m == "merchantitem" {
			found = true
		}
	}
	if !found {
		t.Errorf("Missing = %v, want merchantitem in it", rep.Missing)
	}
}
