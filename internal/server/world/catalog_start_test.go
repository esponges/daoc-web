package world

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/importer"
)

// startStore is a store with Midgard starts in zone 100: a Troll Warrior start
// beside an aggressive level 26 creature, a passive level 40 one beside a
// Dwarf start, and two Norse rows at a third place.
func startStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "world.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, s := range []string{
		importer.Schema,
		`INSERT INTO zone VALUES(100,100,'Vale',0,0,8,8,0,2,1)`,
		`INSERT INTO startup VALUES(2,6,22,100,100,10000,10000,0,0),(2,7,22,100,100,30000,30000,0,0),
			(2,5,22,100,100,50000,50000,0,0),(2,5,24,100,100,50000,50000,0,0)`,
		`INSERT INTO spawn(zone,name,model,x,y,z,heading,level,size,speed,template,roam,respawn_s,aggro_level,aggro_range,realm,flags,race,levels)
			VALUES(100,'bog bear',56,10800,10000,0,0,26,50,191,0,0,30,60,500,0,0,0,''),
			      (100,'old oak',56,30500,30000,0,0,40,50,191,0,0,30,0,0,0,0,0,''),
			      (100,'frog',56,50500,50000,0,0,1,50,191,0,0,30,60,500,0,0,0,'1;2')`,
	} {
		if _, err := db.Exec(s); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestUnsafeStartReplaced(t *testing.T) {
	path := startStore(t)
	before, _ := os.ReadFile(path)
	cat, err := LoadCatalog(path)
	if err != nil {
		t.Fatal(err)
	}
	report := cat.checkStarts()
	cat.CloseDB()

	if len(cat.Unsafe) != 1 {
		t.Fatalf("unsafe starts %+v, want only the Troll's", cat.Unsafe)
	}
	u := cat.Unsafe[0]
	if u.Race != 6 || u.Worst.Name != "bog bear" || u.Worst.Level != 26 || u.Dist != 800 {
		t.Errorf("unsafe start %+v", u)
	}
	if s, _ := cat.StartFor(2, 6, 22); s.X != 50000 {
		t.Errorf("the Troll starts at %v, want the Norse place most rows use", s.X)
	}
	if s, _ := cat.StartFor(2, 7, 22); s.X != 30000 {
		t.Errorf("the Dwarf start beside a passive creature moved to %v", s.X)
	}
	if len(report) != 1 || !strings.Contains(report[0], "bog bear, level 26, 800 units away") {
		t.Errorf("report %q", report)
	}
	if _, near := cat.UnsafeStartNear(100, 10500, 10500); !near {
		t.Error("a place 700 units from the unsafe start is not near it")
	}
	if _, near := cat.UnsafeStartNear(100, 30000, 30000); near {
		t.Error("the Dwarf start counts as near an unsafe one")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Error("loading the catalog changed the store")
	}
}

// The imported Troll Warrior start, in Galplen, is safe once the store honours
// the templates that replace mob values.
func TestGalplenStartIsSafe(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	defer cat.CloseDB()
	if !cat.hasLevels {
		t.Skip("the world store predates honoured templates: run go run ./cmd/dbimport")
	}
	for _, u := range cat.Unsafe {
		if u.Realm == 2 && u.Race == 6 && u.Class == 22 {
			t.Errorf("the Troll Warrior start is unsafe: %v", u)
		}
	}
	if s, ok := cat.StartFor(2, 6, 22); !ok || s.Zone != 105 {
		t.Errorf("the Troll Warrior starts in zone %d, want Myrkwood Forest (105)", s.Zone)
	}
}
