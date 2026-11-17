package world

import (
	"database/sql"
	"path/filepath"
	"testing"

	"daocweb/internal/importer"
)

// buildSpellStore writes a small world store holding two spells: Thunder Shout
// (range 0, radius 300) and a bolt (a single target). old drops the radius
// column, as a store imported before radii were loaded lacks it.
func buildSpellStore(t *testing.T, old bool) *Catalog {
	t.Helper()
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
		`INSERT INTO spell(id,name,type,target,range,cast_time,recast,damage,damage_type,value,duration,frequency,life_drain,sub_spell,uninterruptible,icon,radius,power)
		 VALUES(3521,'Thunder Shout','DirectDamage','Enemy',0,0,20,15,12,0,0,0,0,0,0,1,300,4),
		       (3501,'Thor''s Minor Bolt','DirectDamage','Enemy',1500,3,0,4,12,0,0,0,0,0,0,2,0,2)`,
	}
	if old {
		stmts = append([]string{`ALTER TABLE spell DROP COLUMN radius`}, stmts[0],
			`INSERT INTO spell(id,name,type,target,range,cast_time,recast,damage,damage_type,value,duration,frequency,life_drain,sub_spell,uninterruptible,icon,power)
			 VALUES(3521,'Thunder Shout','DirectDamage','Enemy',0,0,20,15,12,0,0,0,0,0,0,1,4)`)
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

// A spell's radius is loaded; a store without the column loads with none.
func TestCatalogLoadsSpellRadius(t *testing.T) {
	c := buildSpellStore(t, false)
	if sp := c.Spells[3521]; sp == nil || sp.Radius != 300 || sp.Range != 0 {
		t.Errorf("Thunder Shout %+v, want range 0 and radius 300", sp)
	}
	if sp := c.Spells[3501]; sp == nil || sp.Radius != 0 {
		t.Errorf("the bolt %+v, want no radius", sp)
	}
	old := buildSpellStore(t, true)
	if sp := old.Spells[3521]; sp == nil || sp.Radius != 0 || sp.Power != 4 {
		t.Errorf("an old store's Thunder Shout %+v, want no radius and its power", sp)
	}
}

// The Thane's area spells carry their radii in the imported data.
func TestThaneAreaSpellsOnRealData(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	for id, want := range map[int][2]int{3521: {0, 300}, 3528: {0, 300}, 3538: {1500, 350}, 3541: {1500, 350}, 3503: {1500, 0}, 3542: {1500, 0}} {
		sp := cat.Spells[id]
		if sp == nil {
			t.Errorf("spell %d is not in the store", id)
			continue
		}
		if int(sp.Range) != want[0] || sp.Radius != want[1] {
			t.Errorf("%s: range %v radius %d, want %d and %d", sp.Name, sp.Range, sp.Radius, want[0], want[1])
		}
	}
}
