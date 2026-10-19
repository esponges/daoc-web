package importer

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
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
