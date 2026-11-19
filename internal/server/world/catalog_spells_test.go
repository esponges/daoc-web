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
		`INSERT INTO spell(id,name,type,target,range,cast_time,recast,damage,damage_type,value,duration,frequency,life_drain,sub_spell,uninterruptible,icon,pulse,pulse_power)
		 VALUES(3613,'Simple Song of Rest','HealthRegenBuff','Group',1500,0,8,0,0,2,5,50,0,0,0,3,1,3)`,
		`INSERT INTO spell(id,name,type,target,range,cast_time,recast,damage,damage_type,value,duration,frequency,life_drain,sub_spell,uninterruptible,icon,resurrect_health,resurrect_mana,concentration,is_focus)
		 VALUES(1361,'Raise Fallen','Resurrect','Corpse',1500,4,0,0,0,0,0,0,0,0,0,4,30,10,0,0),
		       (2001,'Blessing of Health','ConstitutionBuff','Realm',1000,3,0,0,0,14,0,0,0,0,0,5,0,0,2,1)`,
	}
	if old {
		stmts = append([]string{`ALTER TABLE spell DROP COLUMN radius`, `ALTER TABLE spell DROP COLUMN pulse`, `ALTER TABLE spell DROP COLUMN pulse_power`,
			`ALTER TABLE spell DROP COLUMN resurrect_health`, `ALTER TABLE spell DROP COLUMN resurrect_mana`, `ALTER TABLE spell DROP COLUMN concentration`, `ALTER TABLE spell DROP COLUMN is_focus`}, stmts[0],
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

// A song's pulse and pulse power are loaded; a store without the columns
// loads with none, so its songs are single casts rather than toggles.
func TestCatalogLoadsSpellPulse(t *testing.T) {
	c := buildSpellStore(t, false)
	if sp := c.Spells[3613]; sp == nil || sp.Pulse != 1 || sp.PulsePower != 3 || sp.Frequency != 50 {
		t.Errorf("the song %+v, want pulse 1, pulse power 3, frequency 50", sp)
	}
	if sp := c.Spells[3501]; sp == nil || sp.Pulse != 0 {
		t.Errorf("the bolt %+v, want no pulse", sp)
	}
	old := buildSpellStore(t, true)
	if sp := old.Spells[3521]; sp == nil || sp.Pulse != 0 || sp.PulsePower != 0 {
		t.Errorf("an old store's spell %+v, want no pulse", sp)
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

// A resurrection's health and power, a buff's concentration cost and the focus
// flag are loaded; a store without the columns loads with none of them.
func TestCatalogLoadsRezAndConcentration(t *testing.T) {
	c := buildSpellStore(t, false)
	if sp := c.Spells[1361]; sp == nil || sp.ResurrectHealth != 30 || sp.ResurrectMana != 10 || sp.Concentration != 0 {
		t.Errorf("Raise Fallen %+v, want 30 percent health, 10 power", sp)
	}
	if sp := c.Spells[2001]; sp == nil || sp.Concentration != 2 || !sp.Focus || sp.ResurrectHealth != 0 {
		t.Errorf("Blessing of Health %+v, want concentration 2 and the focus flag", sp)
	}
	old := buildSpellStore(t, true)
	if sp := old.Spells[3521]; sp == nil || sp.Concentration != 0 || sp.ResurrectHealth != 0 || sp.Focus {
		t.Errorf("an old store's spell %+v, want none of the new fields", sp)
	}
}
