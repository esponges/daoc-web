// Package store keeps the players: accounts, characters and what they carry.
// It is a SQLite file of its own, separate from the imported world, so
// re-importing the world never touches a player.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Account is a login: a Discord user, or the local administrator.
type Account struct {
	ID        int64
	DiscordID string // "admin:local" for the administrator
	Name      string
	Admin     bool
}

// Character is a player's character as saved.
type Character struct {
	ID        int64
	AccountID int64
	Name      string
	Race      int
	Gender    int // 0 male, 1 female; saved characters from before this column are male
	Class     int
	Level     int
	XP        int64
	Zone      int
	X, Y, Z   float64
	Heading   int
	HP        int
	Endurance int
	Money     int64

	// The rule book's state. Stats are the base chosen at creation (race plus
	// build); HasStats is false for a character saved before stats existed.
	Stats      [8]int
	HasStats   bool
	Specs      map[string]int // trained spec levels; a missing spec is level 1
	ConLost    int            // constitution lost to deaths and not yet restored
	DeathCount int            // deaths since the last level
	XPScheme   int            // 0: experience on the old curve, 1: on the rule book's table
	Grave      *Grave         // where the character's gravestone stands, if anywhere
	Bind       *Place         // where the character returns on release; nil for one saved before binding, bound on its first join
}

// Place is a spot in a zone, facing a heading.
type Place struct {
	Zone    int
	X, Y, Z float64
	Heading int
}

// Grave is a gravestone: where it is and the experience it holds.
type Grave struct {
	Zone    int
	X, Y, Z float64
	XP      int64
}

// XPSchemeRules marks experience counted on the rule book's table.
const XPSchemeRules = 1

// Item is one stack in a character's inventory or worn on it.
type Item struct {
	ID     int64
	CharID int64
	ItemID string // the world store's item.id
	Slot   int    // 0-39 backpack, 40+ equipment
	Count  int
	// Cond is the item's condition in hundredths of a percent, 10000 whole.
	// Rows from before condition was kept load at 10000.
	Cond    int
	Created time.Time
}

// FullCondition is a whole item's condition.
const FullCondition = 10000

var (
	ErrNameTaken = errors.New("that name is taken")
	ErrBadName   = errors.New("names are 3 to 15 letters")
	ErrNotYours  = errors.New("not your character")
)

const schema = `
CREATE TABLE IF NOT EXISTS account(
  id INTEGER PRIMARY KEY AUTOINCREMENT, discord_id TEXT UNIQUE NOT NULL, name TEXT NOT NULL,
  admin INTEGER NOT NULL DEFAULT 0, created INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS character(
  id INTEGER PRIMARY KEY AUTOINCREMENT, account_id INTEGER NOT NULL REFERENCES account(id),
  name TEXT NOT NULL UNIQUE COLLATE NOCASE, race INTEGER NOT NULL, gender INTEGER NOT NULL DEFAULT 0, class INTEGER NOT NULL,
  level INTEGER NOT NULL DEFAULT 1, xp INTEGER NOT NULL DEFAULT 0, zone INTEGER NOT NULL,
  x REAL NOT NULL, y REAL NOT NULL, z REAL NOT NULL, heading INTEGER NOT NULL DEFAULT 0,
  hp INTEGER NOT NULL DEFAULT 0, endurance INTEGER NOT NULL DEFAULT 0, money INTEGER NOT NULL DEFAULT 0,
  created INTEGER NOT NULL);
CREATE INDEX IF NOT EXISTS character_account ON character(account_id);
CREATE TABLE IF NOT EXISTS char_spec(
  char_id INTEGER NOT NULL REFERENCES character(id) ON DELETE CASCADE, spec TEXT NOT NULL, level INTEGER NOT NULL,
  PRIMARY KEY(char_id, spec));
CREATE TABLE IF NOT EXISTS item(
  id INTEGER PRIMARY KEY AUTOINCREMENT, char_id INTEGER NOT NULL REFERENCES character(id) ON DELETE CASCADE,
  item_id TEXT NOT NULL, slot INTEGER NOT NULL, count INTEGER NOT NULL DEFAULT 1, cond INTEGER NOT NULL DEFAULT 10000, created INTEGER NOT NULL,
  UNIQUE(char_id, slot));
`

// Store is the player database.
type Store struct{ db *sql.DB }

// Open opens (creating if needed) the player database at path; ":memory:"
// gives a throwaway one for tests.
func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	if path == ":memory:" {
		dsn = "file::memory:?_pragma=foreign_keys(1)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite writers are serial; one connection keeps :memory: coherent too
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	// A database made before a column gets it here. Every saved character is
	// male until a creation flow lets a player choose; one saved before the
	// rule book has no stats (NULL) and experience on the old curve (scheme 0),
	// which the world converts on its first join.
	for _, col := range []struct{ name, def string }{
		{"gender", "INTEGER NOT NULL DEFAULT 0"},
		{"str", "INTEGER"}, {"con", "INTEGER"}, {"dex", "INTEGER"}, {"qui", "INTEGER"},
		{"int", "INTEGER"}, {"pie", "INTEGER"}, {"emp", "INTEGER"}, {"cha", "INTEGER"},
		{"con_lost", "INTEGER NOT NULL DEFAULT 0"}, {"death_count", "INTEGER NOT NULL DEFAULT 0"},
		{"xp_scheme", "INTEGER NOT NULL DEFAULT 0"},
		{"grave_zone", "INTEGER"}, {"grave_x", "REAL"}, {"grave_y", "REAL"}, {"grave_z", "REAL"}, {"grave_xp", "INTEGER"},
		{"bind_zone", "INTEGER"}, {"bind_x", "REAL"}, {"bind_y", "REAL"}, {"bind_z", "REAL"}, {"bind_heading", "INTEGER"},
	} {
		var has int
		if err := db.QueryRow(`SELECT count(*) FROM pragma_table_info('character') WHERE name=?`, col.name).Scan(&has); err != nil {
			db.Close()
			return nil, err
		}
		if has == 0 {
			if _, err := db.Exec(`ALTER TABLE character ADD COLUMN ` + col.name + ` ` + col.def); err != nil {
				db.Close()
				return nil, err
			}
		}
	}
	// An item table from before condition was kept gets its column here, every
	// existing item at full condition.
	var hasCond int
	if err := db.QueryRow(`SELECT count(*) FROM pragma_table_info('item') WHERE name='cond'`).Scan(&hasCond); err != nil {
		db.Close()
		return nil, err
	}
	if hasCond == 0 {
		if _, err := db.Exec(`ALTER TABLE item ADD COLUMN cond INTEGER NOT NULL DEFAULT 10000`); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// AccountFor finds or creates the account for a Discord id (or "admin:local").
func (s *Store) AccountFor(discordID, name string, admin bool) (*Account, error) {
	a := &Account{DiscordID: discordID, Name: name, Admin: admin}
	var oldName string
	var oldAdmin bool
	err := s.db.QueryRow(`SELECT id, name, admin FROM account WHERE discord_id=?`, discordID).
		Scan(&a.ID, &oldName, &oldAdmin)
	if err == nil {
		if name != oldName || admin != oldAdmin {
			_, err = s.db.Exec(`UPDATE account SET name=?, admin=? WHERE id=?`, name, admin, a.ID)
		}
		return a, err
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	res, err := s.db.Exec(`INSERT INTO account(discord_id,name,admin,created) VALUES(?,?,?,?)`,
		discordID, name, admin, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	a.ID, _ = res.LastInsertId()
	return a, nil
}

// Account loads an account by id.
func (s *Store) Account(id int64) (*Account, error) {
	a := &Account{ID: id}
	err := s.db.QueryRow(`SELECT discord_id, name, admin FROM account WHERE id=?`, id).
		Scan(&a.DiscordID, &a.Name, &a.Admin)
	return a, err
}

// ValidName reports whether a character name is acceptable: 3 to 15 ASCII
// letters, so it can sit over a head in the world and in a chat line.
func ValidName(n string) bool {
	if len(n) < 3 || len(n) > 15 {
		return false
	}
	for _, c := range n {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return false
		}
	}
	return true
}

// CreateCharacter adds a character at its starting place. The name is
// stored as typed but is unique ignoring case.
func (s *Store) CreateCharacter(c *Character) error {
	c.Name = strings.TrimSpace(c.Name)
	if !ValidName(c.Name) {
		return ErrBadName
	}
	if c.Level < 1 {
		c.Level = 1
	}
	res, err := s.db.Exec(`INSERT INTO character(account_id,name,race,gender,class,level,xp,zone,x,y,z,heading,hp,endurance,money,created,xp_scheme)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		c.AccountID, c.Name, c.Race, c.Gender, c.Class, c.Level, c.XP, c.Zone, c.X, c.Y, c.Z, c.Heading, c.HP, c.Endurance, c.Money, time.Now().Unix(), c.XPScheme)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return ErrNameTaken
		}
		return err
	}
	c.ID, _ = res.LastInsertId()
	if c.HasStats || len(c.Specs) > 0 || c.Bind != nil {
		return s.Save(c)
	}
	return nil
}

const charCols = `id,account_id,name,race,gender,class,level,xp,zone,x,y,z,heading,hp,endurance,money,
  str,con,dex,qui,int,pie,emp,cha,con_lost,death_count,xp_scheme,grave_zone,grave_x,grave_y,grave_z,grave_xp,
  bind_zone,bind_x,bind_y,bind_z,bind_heading`

func scanChar(r interface{ Scan(...any) error }) (*Character, error) {
	c := &Character{}
	var st [8]sql.NullInt64
	var gz, gxp sql.NullInt64
	var gx, gy, gzz sql.NullFloat64
	var bz, bh sql.NullInt64
	var bx, by, bzz sql.NullFloat64
	err := r.Scan(&c.ID, &c.AccountID, &c.Name, &c.Race, &c.Gender, &c.Class, &c.Level, &c.XP, &c.Zone,
		&c.X, &c.Y, &c.Z, &c.Heading, &c.HP, &c.Endurance, &c.Money,
		&st[0], &st[1], &st[2], &st[3], &st[4], &st[5], &st[6], &st[7], &c.ConLost, &c.DeathCount, &c.XPScheme,
		&gz, &gx, &gy, &gzz, &gxp,
		&bz, &bx, &by, &bzz, &bh)
	if err != nil {
		return c, err
	}
	c.HasStats = st[0].Valid
	for i := range st {
		c.Stats[i] = int(st[i].Int64)
	}
	if gz.Valid {
		c.Grave = &Grave{Zone: int(gz.Int64), X: gx.Float64, Y: gy.Float64, Z: gzz.Float64, XP: gxp.Int64}
	}
	if bz.Valid {
		c.Bind = &Place{Zone: int(bz.Int64), X: bx.Float64, Y: by.Float64, Z: bzz.Float64, Heading: int(bh.Int64)}
	}
	return c, nil
}

// loadSpecs fills in a character's trained specs.
func (s *Store) loadSpecs(c *Character) error {
	rows, err := s.db.Query(`SELECT spec,level FROM char_spec WHERE char_id=?`, c.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	c.Specs = map[string]int{}
	for rows.Next() {
		var sp string
		var lv int
		if err := rows.Scan(&sp, &lv); err != nil {
			return err
		}
		c.Specs[sp] = lv
	}
	return rows.Err()
}

// Characters lists an account's characters, oldest first.
func (s *Store) Characters(accountID int64) ([]*Character, error) {
	rows, err := s.db.Query(`SELECT `+charCols+` FROM character WHERE account_id=? ORDER BY id`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Character
	for rows.Next() {
		c, err := scanChar(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Character loads one character, but only if the account owns it.
func (s *Store) Character(accountID, id int64) (*Character, error) {
	c, err := scanChar(s.db.QueryRow(`SELECT `+charCols+` FROM character WHERE id=?`, id))
	if err == sql.ErrNoRows || err == nil && c.AccountID != accountID {
		return nil, ErrNotYours
	}
	if err != nil {
		return nil, err
	}
	return c, s.loadSpecs(c)
}

// Save writes a character's state back.
func (s *Store) Save(c *Character) error {
	return s.Tx(func(t *Tx) error { return t.SaveCharacter(c) })
}

// DeleteCharacter removes a character the account owns, and its items.
func (s *Store) DeleteCharacter(accountID, id int64) error {
	res, err := s.db.Exec(`DELETE FROM character WHERE id=? AND account_id=?`, id, accountID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotYours
	}
	return nil
}

// Items lists what a character carries and wears.
func (s *Store) Items(charID int64) ([]Item, error) {
	rows, err := s.db.Query(`SELECT id,char_id,item_id,slot,count,cond,created FROM item WHERE char_id=? ORDER BY slot`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		var it Item
		var created int64
		if err := rows.Scan(&it.ID, &it.CharID, &it.ItemID, &it.Slot, &it.Count, &it.Cond, &created); err != nil {
			return nil, err
		}
		it.Created = time.Unix(created, 0)
		out = append(out, it)
	}
	return out, rows.Err()
}

// Tx runs fn in a transaction. Loot, equipping and saving use it so an item
// move and the character's save land together or not at all.
func (s *Store) Tx(fn func(*Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(&Tx{tx: tx}); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Tx is the item and character operations inside one transaction.
type Tx struct{ tx *sql.Tx }

// FreeSlot is the lowest empty backpack slot, or -1 if the pack is full.
func (t *Tx) FreeSlot(charID int64, packSize int) (int, error) {
	rows, err := t.tx.Query(`SELECT slot FROM item WHERE char_id=? AND slot<?`, charID, packSize)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	used := map[int]bool{}
	for rows.Next() {
		var s int
		rows.Scan(&s)
		used[s] = true
	}
	for i := 0; i < packSize; i++ {
		if !used[i] {
			return i, nil
		}
	}
	return -1, nil
}

// AddItem puts an item in a slot (which must be empty).
func (t *Tx) AddItem(charID int64, itemID string, slot, count int) error {
	_, err := t.tx.Exec(`INSERT INTO item(char_id,item_id,slot,count,created) VALUES(?,?,?,?,?)`,
		charID, itemID, slot, count, time.Now().Unix())
	return err
}

// AddItemCond puts an item in a slot (which must be empty) at a condition, in
// hundredths of a percent.
func (t *Tx) AddItemCond(charID int64, itemID string, slot, count, cond int) error {
	_, err := t.tx.Exec(`INSERT INTO item(char_id,item_id,slot,count,cond,created) VALUES(?,?,?,?,?,?)`,
		charID, itemID, slot, count, cond, time.Now().Unix())
	return err
}

// SetCondition saves the condition of the item in a slot.
func (t *Tx) SetCondition(charID int64, slot, cond int) error {
	_, err := t.tx.Exec(`UPDATE item SET cond=? WHERE char_id=? AND slot=?`, cond, charID, slot)
	return err
}

// MoveItem moves the item in slot `from` to empty slot `to`.
func (t *Tx) MoveItem(charID int64, from, to int) error {
	res, err := t.tx.Exec(`UPDATE item SET slot=? WHERE char_id=? AND slot=?`, to, charID, from)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("no item in slot %d", from)
	}
	return nil
}

// RemoveItem deletes the item in a slot.
func (t *Tx) RemoveItem(charID int64, slot int) error {
	_, err := t.tx.Exec(`DELETE FROM item WHERE char_id=? AND slot=?`, charID, slot)
	return err
}

// SaveCharacter writes a character inside the transaction.
func (t *Tx) SaveCharacter(c *Character) error {
	var st [8]any
	for i := range st {
		if c.HasStats {
			st[i] = c.Stats[i]
		}
	}
	var gz, gx, gy, gzz, gxp any
	if g := c.Grave; g != nil {
		gz, gx, gy, gzz, gxp = g.Zone, g.X, g.Y, g.Z, g.XP
	}
	var bz, bx, by, bzz, bh any
	if b := c.Bind; b != nil {
		bz, bx, by, bzz, bh = b.Zone, b.X, b.Y, b.Z, b.Heading
	}
	_, err := t.tx.Exec(`UPDATE character SET level=?,xp=?,zone=?,x=?,y=?,z=?,heading=?,hp=?,endurance=?,money=?,
		str=?,con=?,dex=?,qui=?,int=?,pie=?,emp=?,cha=?,con_lost=?,death_count=?,xp_scheme=?,
		grave_zone=?,grave_x=?,grave_y=?,grave_z=?,grave_xp=?,
		bind_zone=?,bind_x=?,bind_y=?,bind_z=?,bind_heading=? WHERE id=?`,
		c.Level, c.XP, c.Zone, c.X, c.Y, c.Z, c.Heading, c.HP, c.Endurance, c.Money,
		st[0], st[1], st[2], st[3], st[4], st[5], st[6], st[7], c.ConLost, c.DeathCount, c.XPScheme,
		gz, gx, gy, gzz, gxp, bz, bx, by, bzz, bh, c.ID)
	if err != nil {
		return err
	}
	if _, err := t.tx.Exec(`DELETE FROM char_spec WHERE char_id=?`, c.ID); err != nil {
		return err
	}
	for sp, lv := range c.Specs {
		if lv > 1 {
			if _, err := t.tx.Exec(`INSERT INTO char_spec VALUES(?,?,?)`, c.ID, sp, lv); err != nil {
				return err
			}
		}
	}
	return nil
}

// SwapItems exchanges the items in two slots.
func (t *Tx) SwapItems(charID int64, a, b int) error {
	const tmp = 1 << 20
	for _, m := range [][2]int{{a, tmp}, {b, a}, {tmp, b}} {
		if _, err := t.tx.Exec(`UPDATE item SET slot=? WHERE char_id=? AND slot=?`, m[1], charID, m[0]); err != nil {
			return err
		}
	}
	return nil
}

// GraveRow is a character's gravestone, as a zone loads it.
type GraveRow struct {
	CharID int64
	Name   string
	Grave
}

// GravesIn lists the gravestones standing in a zone.
func (s *Store) GravesIn(zone int) ([]GraveRow, error) {
	rows, err := s.db.Query(`SELECT id,name,grave_x,grave_y,grave_z,grave_xp FROM character WHERE grave_zone=?`, zone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GraveRow
	for rows.Next() {
		g := GraveRow{Grave: Grave{Zone: zone}}
		if err := rows.Scan(&g.CharID, &g.Name, &g.X, &g.Y, &g.Z, &g.XP); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}
