package importer

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

// minimalDump writes a dump in which every required table keeps one row, so
// an import of it reports nothing missing (dbimport then exits 0).
func minimalDump(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dumpTable(t, dir, "zones", "`ZoneID`, `RegionID`, `OffsetX`, `OffsetY`, `Width`, `Height`, `Name`, `WaterLevel`, `Realm`",
		"1, 1, 0, 0, 1, 1, 'Test', 0, 2")
	dumpTable(t, dir, "mob", "`Name`, `Region`, `X`, `Y`, `Z`, `Model`, `ClassType`, `ItemsListTemplateID`",
		"'Smith', 1, 100, 100, 0, 10, 'DOL.GS.GameMerchant', 'L1'")
	dumpTable(t, dir, "merchantitem", "`ItemListID`, `ItemTemplateID`, `PageNumber`, `SlotPosition`", "'L1', 'sword', 0, 1")
	dumpTable(t, dir, "npcequipment", "`TemplateID`, `Slot`, `Model`", "'E1', 10, 5")
	dumpTable(t, dir, "npctemplate", "`TemplateId`, `Name`", "1, 'wolf'")
	dumpTable(t, dir, "itemtemplate", "`Id_nb`, `Name`", "'sword', 'sword'")
	dumpTable(t, dir, "style", "`ID`, `ClassId`, `Name`", "1, 22, 'Slam'")
	dumpTable(t, dir, "startuplocation", "`RealmID`, `RaceID`, `ClassID`, `Region`, `XPos`, `YPos`, `ZPos`", "2, 6, 22, 1, 50, 50, 0")
	dumpTable(t, dir, "starterequipment", "`StarterEquipmentID`, `Class`, `TemplateID`", "1, '22;', 'sword'")
	dumpTable(t, dir, "path", "`PathID`, `PathType`", "'HS_X_Y', 1")
	dumpTable(t, dir, "pathpoints", "`PathID`, `Step`, `X`, `Y`, `Z`, `MaxSpeed`, `WaitTime`", "'HS_X_Y', 1, 100, 100, 0, 0, 0")
	for _, tb := range []string{"loottemplate", "mobxloottemplate", "specialization", "classxspecialization", "race", "zonepoint", "bindpoint"} {
		dumpTable(t, dir, tb, "`Unused`", "0")
	}
	return dir
}

func spawnCount(t *testing.T, path string) int {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM spawn`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// A store nobody holds is replaced in place, and no staged copy is left.
func TestImportReplacesFreeStore(t *testing.T) {
	dump := minimalDump(t)
	out := filepath.Join(t.TempDir(), "world.db")
	os.WriteFile(out+".next", []byte("stale"), 0o644)
	rep, err := Run(Options{Dump: dump, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Missing) > 0 || rep.Staged != "" {
		t.Fatalf("missing %v, staged %q; want neither", rep.Missing, rep.Staged)
	}
	if spawnCount(t, out) != 1 {
		t.Error("the new store is not in place")
	}
	for _, p := range []string{out + ".next", out + ".tmp"} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("%s left behind", filepath.Base(p))
		}
	}
}

// When the store cannot be replaced, the import is staged beside it and the
// old store is untouched. A directory stands in for the held file, since only
// some systems refuse to replace an open one.
func TestImportStagesWhenStoreCannotBeReplaced(t *testing.T) {
	dump := minimalDump(t)
	out := filepath.Join(t.TempDir(), "world.db")
	os.MkdirAll(filepath.Join(out, "held"), 0o755)
	rep, err := Run(Options{Dump: dump, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Missing) > 0 {
		t.Fatalf("missing %v", rep.Missing)
	}
	if rep.Staged != out+".next" {
		t.Fatalf("staged %q, want %q", rep.Staged, out+".next")
	}
	if spawnCount(t, rep.Staged) != 1 {
		t.Error("the staged store is not a complete import")
	}
	if _, err := os.Stat(filepath.Join(out, "held")); err != nil {
		t.Error("the store in use was disturbed")
	}
}

// On Windows, a store a running server has open cannot be replaced: the import
// is staged and still succeeds.
func TestImportStagesWhileStoreIsOpen(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("other systems replace an open file without complaint")
	}
	dump := minimalDump(t)
	out := filepath.Join(t.TempDir(), "world.db")
	if _, err := Run(Options{Dump: dump, Out: out}); err != nil {
		t.Fatal(err)
	}
	held, err := sql.Open("sqlite", "file:"+out+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	var n int
	if err := held.QueryRow(`SELECT count(*) FROM spawn`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	rep, err := Run(Options{Dump: dump, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Staged != out+".next" || len(rep.Missing) > 0 {
		t.Fatalf("staged %q, missing %v", rep.Staged, rep.Missing)
	}
	if spawnCount(t, rep.Staged) != 1 {
		t.Error("the staged store is not a complete import")
	}
}

// A template marked ReplaceMobValues overrides the spawn row: its name,
// levels, aggression and class type win, as the emulator loads them. One
// without the mark leaves the row alone.
func TestTemplateReplacesMobValues(t *testing.T) {
	dir := minimalDump(t)
	dumpTable(t, dir, "npctemplate", "`TemplateId`, `Name`, `Level`, `AggroLevel`, `AggroRange`, `MaxSpeed`, `ClassType`, `ReplaceMobValues`",
		"10, 'giant bull frog', '1', 0, 500, 175, 'DOL.GS.GameNPC', 1",
		"11, 'hatchling', '2;4;3', 30, 400, 0, 'DOL.GS.GameMerchant', 1",
		"12, 'keeper', '40', 0, 0, 0, 'DOL.GS.GameNPC', 0")
	dumpTable(t, dir, "mob", "`Name`, `Region`, `X`, `Y`, `Z`, `Model`, `Level`, `AggroLevel`, `AggroRange`, `Speed`, `NPCTemplateID`, `ClassType`",
		"'giant bull frog', 1, 100, 100, 0, 10, 26, 200, 500, 191, 10, 'DOL.GS.GameNPC'",
		"'old name', 1, 110, 100, 0, 10, 3, 0, 0, 191, 11, 'DOL.GS.GameNPC'",
		"'Keeper', 1, 120, 100, 0, 10, 50, 0, 0, 191, 12, 'DOL.GS.GameNPC'")
	out := filepath.Join(t.TempDir(), "world.db")
	rep, err := Run(Options{Dump: dir, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Replaced != 2 || rep.LevelReplaced != 1 {
		t.Errorf("replaced %d, level replaced %d; want 2 and 1", rep.Replaced, rep.LevelReplaced)
	}
	db, _ := sql.Open("sqlite", out)
	defer db.Close()
	type row struct {
		name, levels, class string
		level, aggro, rng   int
		speed               int
	}
	get := func(x float64) row {
		var r row
		if err := db.QueryRow(`SELECT name,level,COALESCE(levels,''),aggro_level,aggro_range,speed,classtype FROM spawn WHERE x=?`, x).
			Scan(&r.name, &r.level, &r.levels, &r.aggro, &r.rng, &r.speed, &r.class); err != nil {
			t.Fatal(err)
		}
		return r
	}
	if r := get(100); r.level != 1 || r.levels != "1" || r.aggro != 0 || r.speed != 175 {
		t.Errorf("frog = %+v, want level 1, passive, speed 175", r)
	}
	if r := get(110); r.name != "hatchling" || r.level != 2 || r.levels != "2;4;3" || r.aggro != 30 || r.rng != 400 || r.speed != 191 || r.class != "DOL.GS.GameMerchant" {
		t.Errorf("hatchling = %+v", r)
	}
	if r := get(120); r.name != "Keeper" || r.level != 50 || r.levels != "" {
		t.Errorf("unmarked template changed the row: %+v", r)
	}
}

func TestNumberList(t *testing.T) {
	for in, want := range map[string]string{"1": "1", "13;18;16": "13;18;16", "1-3": "1;2;3", "": "", "x;5": "5", " 7 , 8": "7;8"} {
		if got := joinLevels(NumberList(in)); got != want {
			t.Errorf("NumberList(%q) = %q, want %q", in, got, want)
		}
	}
}

// Start equipment is stored once per class its row lists, in the dump's order;
// a dump without the table reports it missing.
func TestImportStartEquipment(t *testing.T) {
	dir := minimalDump(t)
	dumpTable(t, dir, "starterequipment", "`StarterEquipmentID`, `Class`, `TemplateID`",
		"66, '35;31;22;', 'small_training_shield'", "67, '22;', 'training_axe'", "68, '', 'nobody'")
	out := filepath.Join(t.TempDir(), "world.db")
	rep, err := Run(Options{Dump: dir, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if c := rep.Tables["starterequipment"]; c != [2]int{3, 2} || len(rep.Missing) > 0 {
		t.Errorf("read/kept %v, missing %v", c, rep.Missing)
	}
	db, _ := sql.Open("sqlite", out)
	defer db.Close()
	var got []string
	rows, _ := db.Query(`SELECT class, item FROM start_equip ORDER BY class, ord`)
	for rows.Next() {
		var c int
		var it string
		rows.Scan(&c, &it)
		got = append(got, fmt.Sprintf("%d %s", c, it))
	}
	rows.Close()
	want := "[22 small_training_shield 22 training_axe 31 small_training_shield 35 small_training_shield]"
	if fmt.Sprint(got) != want {
		t.Errorf("start_equip = %v, want %v", got, want)
	}

	os.Remove(filepath.Join(dir, "starterequipment.sql"))
	rep, err = Run(Options{Dump: dir, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(rep.Missing, "starterequipment") {
		t.Errorf("missing %v, want starterequipment", rep.Missing)
	}
}

// Spells come in with what names them: a template's list, a style's spell,
// an item's proc, and the spells those fire in turn. A link to a spell the
// dump lacks is dropped and counted.
func TestImportSpells(t *testing.T) {
	dir := minimalDump(t)
	dumpTable(t, dir, "npctemplate", "`TemplateId`, `Name`, `Spells`", "1, 'wolf', ''", "2, 'shaman', '1001;1002;9999'")
	dumpTable(t, dir, "stylexspell", "`SpellID`, `ClassID`, `StyleID`, `Chance`", "1003, 0, 164, 100")
	dumpTable(t, dir, "itemtemplate", "`Id_nb`, `Name`, `ProcSpellID`, `ProcChance`", "'sword', 'sword', 1004, 0")
	dumpTable(t, dir, "spell", "`SpellID`, `Name`, `Type`, `Target`, `Range`, `CastTime`, `Damage`, `DamageType`, `Duration`, `Frequency`, `RecastDelay`, `Value`, `LifeDrainReturn`, `SubSpellID`, `Uninterruptible`, `Message1`",
		"1001, 'Spirit Bolt', 'Bolt', 'Enemy', 1875, 2.5, 24, 12, 0, 0, 20, 0, 0, 0, 0, 'A bolt hits you!'",
		"1002, 'Mend', 'Heal', 'Self', 0, 3, 0, 0, 0, 0, 0, 36, 0, 0, 1, ''",
		"1003, 'Slam Stun', 'Stun', 'Enemy', 0, 0, 0, 10, 3, 0, 0, 0, 0, 0, 0, 'You are stunned!'",
		"1004, 'Blightstrike', 'OffensiveProc', 'Self', 0, 0, 0, 0, 1200, 15, 0, 1005, 0, 0, 0, ''",
		"1005, 'Fungus', 'DirectDamage', 'Enemy', 1500, 0, 20, 15, 0, 0, 0, 0, 0, 1006, 0, ''",
		"1006, 'Fungus Drain', 'Lifedrain', 'Enemy', 1500, 0, 10, 15, 0, 0, 0, 0, 50, 0, 0, ''",
		"2000, 'Unused', 'Heal', 'Self', 0, 0, 0, 0, 0, 0, 0, 10, 0, 0, 0, ''")
	out := filepath.Join(t.TempDir(), "world.db")
	rep, err := Run(Options{Dump: dir, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if rep.SpellLinksMissing != 1 {
		t.Errorf("missing links %d, want the one to 9999", rep.SpellLinksMissing)
	}
	db, _ := sql.Open("sqlite", out)
	defer db.Close()
	ids := func(q string) string {
		rows, err := db.Query(q)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var l []string
		for rows.Next() {
			var s string
			rows.Scan(&s)
			l = append(l, s)
		}
		return fmt.Sprint(l)
	}
	if got := ids(`SELECT id FROM spell ORDER BY id`); got != "[1001 1002 1003 1004 1005 1006]" {
		t.Errorf("spells %s, want the named ones and what they fire, not 2000", got)
	}
	if got := ids(`SELECT spell FROM template_spell WHERE template=2 ORDER BY ord`); got != "[1001 1002]" {
		t.Errorf("template 2 casts %s", got)
	}
	if got := ids(`SELECT class||'/'||style||'/'||spell||'/'||chance FROM style_spell`); got != "[0/164/1003/100]" {
		t.Errorf("style spells %s", got)
	}
	if got := ids(`SELECT proc_spell||'/'||proc_chance FROM item WHERE id='sword'`); got != "[1004/0]" {
		t.Errorf("sword proc %s", got)
	}
	var cast, dmg float64
	var freq, unint int
	var msg string
	db.QueryRow(`SELECT cast_time, damage, frequency, uninterruptible, message1 FROM spell WHERE id=1001`).Scan(&cast, &dmg, &freq, &unint, &msg)
	if cast != 2.5 || dmg != 24 || msg != "A bolt hits you!" {
		t.Errorf("bolt cast %v damage %v message %q", cast, dmg, msg)
	}
	db.QueryRow(`SELECT uninterruptible FROM spell WHERE id=1002`).Scan(&unint)
	if unint != 1 {
		t.Error("the heal lost its uninterruptible mark")
	}
}
