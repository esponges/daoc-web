package store

import (
	"database/sql"
	"testing"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestAccountAndCharacter(t *testing.T) {
	s := open(t)
	a, err := s.AccountFor("123", "grimm#0", false)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := s.AccountFor("123", "grimm#0", false)
	if a.ID != b.ID {
		t.Error("same Discord id must be the same account")
	}
	c := &Character{AccountID: a.ID, Name: "Grimm", Race: 7, Class: 22, Zone: 100, X: 1000, Y: 2000, Z: 5000, HP: 220}
	if err := s.CreateCharacter(c); err != nil {
		t.Fatal(err)
	}
	// name taken, ignoring case, by anyone
	other, _ := s.AccountFor("456", "x", false)
	if err := s.CreateCharacter(&Character{AccountID: other.ID, Name: "gRIMM", Zone: 100}); err != ErrNameTaken {
		t.Errorf("duplicate name: %v", err)
	}
	for _, n := range []string{"", "ab", "waytoolongname1234", "gr1mm", "gr mm"} {
		if err := s.CreateCharacter(&Character{AccountID: a.ID, Name: n}); err != ErrBadName {
			t.Errorf("name %q: %v", n, err)
		}
	}
	// another account's character is not yours
	if _, err := s.Character(other.ID, c.ID); err != ErrNotYours {
		t.Errorf("other's character: %v", err)
	}
	if err := s.DeleteCharacter(other.ID, c.ID); err != ErrNotYours {
		t.Errorf("deleting other's: %v", err)
	}
	// a save round-trips
	c.X, c.Level, c.XP = 4242.5, 3, 900
	if err := s.Save(c); err != nil {
		t.Fatal(err)
	}
	got, err := s.Character(a.ID, c.ID)
	if err != nil || got.X != 4242.5 || got.Level != 3 || got.XP != 900 || got.Name != "Grimm" {
		t.Errorf("reload = %+v, %v", got, err)
	}
}

func TestPersistsAcrossReopen(t *testing.T) {
	path := t.TempDir() + "/p.db"
	s, _ := Open(path)
	a, _ := s.AccountFor("1", "a", true)
	c := &Character{AccountID: a.ID, Name: "Keeper", Zone: 100, X: 77}
	s.CreateCharacter(c)
	s.Close()
	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	list, _ := s2.Characters(a.ID)
	a2, _ := s2.Account(a.ID)
	if len(list) != 1 || list[0].X != 77 || !a2.Admin {
		t.Errorf("after reopen: %+v admin=%v", list, a2.Admin)
	}
}

func TestItemsAtomic(t *testing.T) {
	s := open(t)
	a, _ := s.AccountFor("1", "a", false)
	c := &Character{AccountID: a.ID, Name: "Looter", Zone: 100}
	s.CreateCharacter(c)
	err := s.Tx(func(tx *Tx) error {
		slot, _ := tx.FreeSlot(c.ID, 40)
		if err := tx.AddItem(c.ID, "wolf_pelt", slot, 1); err != nil {
			return err
		}
		c.Money += 50
		return tx.SaveCharacter(c)
	})
	if err != nil {
		t.Fatal(err)
	}
	// A failing transaction leaves neither the item nor the money.
	err = s.Tx(func(tx *Tx) error {
		tx.AddItem(c.ID, "second", 1, 1)
		c.Money += 999
		tx.SaveCharacter(c)
		return ErrNotYours
	})
	if err == nil {
		t.Fatal("expected the rollback error")
	}
	items, _ := s.Items(c.ID)
	if len(items) != 1 || items[0].ItemID != "wolf_pelt" {
		t.Errorf("items = %+v", items)
	}
	got, _ := s.Character(a.ID, c.ID)
	if got.Money != 50 {
		t.Errorf("money = %d, want 50", got.Money)
	}
}

func TestGenderDefaultsToMaleAndIsKept(t *testing.T) {
	s := open(t)
	a, _ := s.AccountFor("g1", "g#0", false)
	m := &Character{AccountID: a.ID, Name: "Malefic", Race: 6, Class: 22, Zone: 100}
	f := &Character{AccountID: a.ID, Name: "Femalia", Race: 6, Gender: 1, Class: 22, Zone: 100}
	for _, c := range []*Character{m, f} {
		if err := s.CreateCharacter(c); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []*Character{m, f} {
		got, err := s.Character(a.ID, c.ID)
		if err != nil || got.Gender != c.Gender {
			t.Errorf("%s: gender %d, want %d (%v)", c.Name, got.Gender, c.Gender, err)
		}
	}
}

func TestOldDatabaseGainsGender(t *testing.T) {
	path := t.TempDir() + "/old.db"
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	// A character table as it was before the gender column.
	_, err = db.Exec(`CREATE TABLE account(id INTEGER PRIMARY KEY AUTOINCREMENT, discord_id TEXT UNIQUE NOT NULL, name TEXT NOT NULL, admin INTEGER NOT NULL DEFAULT 0, created INTEGER NOT NULL);
CREATE TABLE character(id INTEGER PRIMARY KEY AUTOINCREMENT, account_id INTEGER NOT NULL, name TEXT NOT NULL UNIQUE COLLATE NOCASE, race INTEGER NOT NULL, class INTEGER NOT NULL,
 level INTEGER NOT NULL DEFAULT 1, xp INTEGER NOT NULL DEFAULT 0, zone INTEGER NOT NULL, x REAL NOT NULL, y REAL NOT NULL, z REAL NOT NULL, heading INTEGER NOT NULL DEFAULT 0,
 hp INTEGER NOT NULL DEFAULT 0, endurance INTEGER NOT NULL DEFAULT 0, money INTEGER NOT NULL DEFAULT 0, created INTEGER NOT NULL);
INSERT INTO account VALUES(1,'old','old',0,1);
INSERT INTO character(account_id,name,race,class,zone,x,y,z,created) VALUES(1,'Elder',5,22,100,1,2,3,1);`)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cs, err := s.Characters(1)
	if err != nil || len(cs) != 1 || cs[0].Gender != 0 || cs[0].Name != "Elder" {
		t.Errorf("migrated characters: %+v, %v", cs, err)
	}
}
