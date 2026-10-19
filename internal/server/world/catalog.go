package world

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"daocweb/internal/importer"
	"daocweb/internal/server/rules"

	_ "modernc.org/sqlite"
)

// Item is an item template as the game rules use it.
type Item struct {
	ID         string
	Name       string
	Level      int
	Quality    int
	DPS        float64 // damage per second, from the table's tenths
	Speed      float64 // seconds a swing, from tenths
	AF         int     // armour factor, for armour
	Hand       int
	DamageType int
	ObjectType int
	ItemType   int
	Model      int
	Price      int64
	Realm      int
	Droppable  bool
	LevelReq   int
	Absorb     int         // armour absorption percent
	Bonus      int         // enchantment percent
	Condition  float64     // 0..1 of its maximum
	Bonuses    []ItemBonus // stat and skill bonuses
}

// ItemBonus is one bonus an item gives, by the item table's bonus type numbers.
type ItemBonus struct {
	Type, Amount int
}

// Object types the rules care about, as the item table numbers them.
const (
	ObjHammer = 12
	ObjShield = 42
)

// IsWeapon reports whether the item is something held to fight with.
func (i *Item) IsWeapon() bool { return i.ItemType == 10 || i.ItemType == 11 || i.ItemType == 12 }

// TwoHanded is a weapon held in both hands.
func (i *Item) TwoHanded() bool { return i.IsWeapon() && (i.Hand == 1 || i.ItemType == 12) }

// IsShield is a shield.
func (i *Item) IsShield() bool { return i.ObjectType == ObjShield }

// IsArmour is something worn on the body.
func (i *Item) IsArmour() bool { return i.ObjectType >= 32 && i.ObjectType <= 38 }

// StyleDef is one style from the style table.
type StyleDef struct {
	ID         int
	Class      int
	Name       string
	Spec       string
	SpecLevel  int
	Icon       int
	Endurance  int
	OpenType   int // 0 offensive/anytime, 1 defensive, 2 positional
	OpenValue  int
	ResultReq  int // for offensive and defensive: what had to happen
	WeaponReq  int
	GrowthRate float64
	GrowthOff  float64
	BonusToHit int
	BonusToDef int
	Anim       int
}

// Spawn is an NPC placement from the world store.
type Spawn struct {
	ID         int64
	Zone       int
	Name       string
	Guild      string
	Equip      string // equipment template id, "" for none
	Model      int
	X, Y, Z    float64
	Heading    int
	Level      int
	Size       int
	Speed      int
	Template   int
	Roam       int
	RespawnS   int
	AggroLevel int
	AggroRange int
	Realm      int
	Flags      int
	ClassType  string // the community class type, "" when the store predates it
	ItemsList  string // merchant stock list id, "" for none
	Race       int    // the race table row its resists come from, 0 for none
	Levels     []int  // the levels it may spawn at, from a template that replaces the row; nil uses Level
}

// Template is the optional fixed stats of an NPC kind.
type Template struct {
	ID        int
	Name      string
	Parry     int
	Evade     int
	Block     int
	Tether    int
	LevelLo   int
	LevelHi   int
	MeleeType int // damage type of its blows: 1 crush, 2 slash, 3 thrust
}

// LootEntry is one row of a loot table.
type LootEntry struct {
	Item   string
	Chance int
	Count  int
}

// Start is where a new character begins.
type Start struct {
	Zone    int
	X, Y, Z float64
	Heading int
}

// Exit is a zone exit.
type Exit struct {
	SrcZone    int
	X, Y, Z    float64
	DstZone    int
	TX, TY, TZ float64
	Heading    int
	Realm      int
}

// Bind is a place to come back to after death.
type Bind struct {
	Zone    int
	X, Y, Z float64
	Radius  int
	Realm   int
}

// ZoneInfo is a row of the zone table.
type ZoneInfo struct {
	ID        int
	Region    int
	Name      string
	Realm     int
	Converted bool
	// Where the zone sits in its region, in 8192-unit blocks, and how many
	// blocks it spans, which is how neighbouring zones are found.
	OffX, OffY, W, H int
}

// Catalog is the imported world, loaded into memory. It is read-only after
// Load, so zones share it without locks.
type Catalog struct {
	Items     map[string]*Item
	Styles    []StyleDef
	Templates map[int]*Template
	Zones     map[int]*ZoneInfo
	Exits     []Exit
	Binds     []Bind
	Starts    []struct {
		Realm, Race, Class int
		Start
	}
	Loot       map[string][]LootEntry  // loot template -> entries
	Gear       map[string][][4]int     // equipment template -> [slot, model, colour, extension] per worn piece
	hasGear    bool                    // spawn.equip exists (a world.db imported before equipment was kept lacks it)
	MobLoot    map[string][]string     // lower-case mob name -> loot template names
	Stock      map[string][]string     // merchant list id -> item ids, in page then slot order
	hasShops   bool                    // spawn.classtype and merchant_stock exist (a world.db imported before merchants lacks them)
	hasRules   bool                    // item bonuses, races and abilities exist (a world.db imported before the rule book lacks them)
	hasLevels  bool                    // spawn.levels exists (a world.db imported before templates were honoured lacks it)
	Resists    map[int]map[int]float64 // race -> damage type -> percent
	ClassSpecs map[int][]string        // class -> its spec and career keys
	SpecNames  map[string]string       // spec key -> name
	Grants     map[int][]rules.Grant   // class -> the ability grants of its careers and specs
	Unsafe     []UnsafeStart           // starts the start rule replaced, from checkStarts
	replaced   map[int]UnsafeStart     // index in Starts -> its replacement
	dropOnce   sync.Once
	dropPool   map[int][]string
	db         *sql.DB
}

// LoadCatalog reads world.db.
func LoadCatalog(path string) (*Catalog, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, err
	}
	c := &Catalog{db: db, Items: map[string]*Item{}, Templates: map[int]*Template{}, Zones: map[int]*ZoneInfo{},
		Loot: map[string][]LootEntry{}, MobLoot: map[string][]string{}, Stock: map[string][]string{}}
	q := func(query string, scan func(*sql.Rows) error) error {
		rows, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("world store: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			if err := scan(rows); err != nil {
				return err
			}
		}
		return rows.Err()
	}
	if err := q(`SELECT id,region,name,realm,converted,offset_x,offset_y,width,height FROM zone`, func(r *sql.Rows) error {
		z := &ZoneInfo{}
		if err := r.Scan(&z.ID, &z.Region, &z.Name, &z.Realm, &z.Converted, &z.OffX, &z.OffY, &z.W, &z.H); err != nil {
			return err
		}
		c.Zones[z.ID] = z
		return nil
	}); err != nil {
		return nil, err
	}
	c.Gear = map[string][][4]int{}
	var gearTables int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE (type='table' AND name='npc_equip') OR (type='table' AND name='spawn' AND sql LIKE '%equip TEXT%')`).Scan(&gearTables)
	if c.hasGear = gearTables == 2; c.hasGear {
		if err := q(`SELECT template,slot,model,color,ext FROM npc_equip ORDER BY template,slot`, func(r *sql.Rows) error {
			var id string
			var g [4]int
			if err := r.Scan(&id, &g[0], &g[1], &g[2], &g[3]); err != nil {
				return err
			}
			c.Gear[id] = append(c.Gear[id], g)
			return nil
		}); err != nil {
			return nil, err
		}
	}
	var ruleParts int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND (name IN ('item_bonus','spec_ability') OR (name='item' AND sql LIKE '%max_condition%') OR (name='spawn' AND sql LIKE '%race INTEGER%'))`).Scan(&ruleParts)
	c.hasRules = ruleParts == 4
	var levelCols int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='spawn' AND sql LIKE '%levels TEXT%'`).Scan(&levelCols)
	c.hasLevels = levelCols == 1
	itemCols := "0,100,1"
	if c.hasRules {
		itemCols = "COALESCE(bonus,0),COALESCE(condition,0),COALESCE(max_condition,0)"
	}
	if err := q(`SELECT id,name,level,quality,dps_af,spd_abs,hand,damage_type,object_type,item_type,model,price,realm,droppable,COALESCE(level_req,0),`+itemCols+` FROM item`, func(r *sql.Rows) error {
		it := &Item{}
		var dps, spd int
		var drop int
		var cond, maxCond int
		if err := r.Scan(&it.ID, &it.Name, &it.Level, &it.Quality, &dps, &spd, &it.Hand, &it.DamageType, &it.ObjectType, &it.ItemType, &it.Model, &it.Price, &it.Realm, &drop, &it.LevelReq, &it.Bonus, &cond, &maxCond); err != nil {
			return err
		}
		it.DPS, it.Speed, it.AF, it.Absorb = float64(dps)/10, float64(spd)/10, dps, spd
		it.Condition = 1
		if maxCond > 0 && cond > 0 && cond < maxCond {
			it.Condition = float64(cond) / float64(maxCond)
		}
		it.Droppable = drop != 0
		c.Items[it.ID] = it
		return nil
	}); err != nil {
		return nil, err
	}
	if err := q(`SELECT id,class_id,name,spec,spec_level,icon,endurance,opening_type,opening_value,result_req,weapon_req,growth_rate,growth_offset,bonus_hit,bonus_def,anim FROM style`, func(r *sql.Rows) error {
		s := StyleDef{}
		if err := r.Scan(&s.ID, &s.Class, &s.Name, &s.Spec, &s.SpecLevel, &s.Icon, &s.Endurance, &s.OpenType, &s.OpenValue, &s.ResultReq, &s.WeaponReq, &s.GrowthRate, &s.GrowthOff, &s.BonusToHit, &s.BonusToDef, &s.Anim); err != nil {
			return err
		}
		c.Styles = append(c.Styles, s)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := q(`SELECT id,name,parry,evade,block,tether,level_lo,level_hi,COALESCE(melee_type,0) FROM npc_template`, func(r *sql.Rows) error {
		t := &Template{}
		if err := r.Scan(&t.ID, &t.Name, &t.Parry, &t.Evade, &t.Block, &t.Tether, &t.LevelLo, &t.LevelHi, &t.MeleeType); err != nil {
			return err
		}
		c.Templates[t.ID] = t
		return nil
	}); err != nil {
		return nil, err
	}
	if err := q(`SELECT src_zone,sx,sy,sz,dst_zone,tx,ty,tz,heading,realm FROM zonepoint`, func(r *sql.Rows) error {
		e := Exit{}
		if err := r.Scan(&e.SrcZone, &e.X, &e.Y, &e.Z, &e.DstZone, &e.TX, &e.TY, &e.TZ, &e.Heading, &e.Realm); err != nil {
			return err
		}
		c.Exits = append(c.Exits, e)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := q(`SELECT zone,x,y,z,radius,realm FROM bindpoint`, func(r *sql.Rows) error {
		b := Bind{}
		if err := r.Scan(&b.Zone, &b.X, &b.Y, &b.Z, &b.Radius, &b.Realm); err != nil {
			return err
		}
		c.Binds = append(c.Binds, b)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := q(`SELECT realm,race,class,zone,x,y,z,heading FROM startup`, func(r *sql.Rows) error {
		var s struct {
			Realm, Race, Class int
			Start
		}
		if err := r.Scan(&s.Realm, &s.Race, &s.Class, &s.Zone, &s.X, &s.Y, &s.Z, &s.Heading); err != nil {
			return err
		}
		c.Starts = append(c.Starts, s)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := q(`SELECT template,item,chance,count FROM loot`, func(r *sql.Rows) error {
		var t string
		e := LootEntry{}
		if err := r.Scan(&t, &e.Item, &e.Chance, &e.Count); err != nil {
			return err
		}
		c.Loot[t] = append(c.Loot[t], e)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := q(`SELECT mob_name,template FROM mob_loot`, func(r *sql.Rows) error {
		var m, t string
		if err := r.Scan(&m, &t); err != nil {
			return err
		}
		c.MobLoot[m] = append(c.MobLoot[m], t)
		return nil
	}); err != nil {
		return nil, err
	}
	var shopParts int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE (type='table' AND name='merchant_stock') OR (type='table' AND name='spawn' AND sql LIKE '%classtype TEXT%')`).Scan(&shopParts)
	if c.hasShops = shopParts == 2; c.hasShops {
		if err := q(`SELECT list,item FROM merchant_stock ORDER BY list,page,slot`, func(r *sql.Rows) error {
			var l, it string
			if err := r.Scan(&l, &it); err != nil {
				return err
			}
			c.Stock[l] = append(c.Stock[l], it)
			return nil
		}); err != nil {
			return nil, err
		}
	} else {
		log.Printf("world store has no merchant data; NPC interaction is off (run cmd/dbimport again)")
	}
	if err := c.loadRules(q); err != nil {
		return nil, err
	}
	for _, l := range c.checkStarts() {
		log.Print(l)
	}
	return c, nil
}

// Shops reports whether the store carries merchant data.
func (c *Catalog) Shops() bool { return c.hasShops }

// SpawnsIn reads a zone's spawns.
func (c *Catalog) SpawnsIn(zone int) ([]Spawn, error) {
	equip := "''"
	if c.hasGear {
		equip = "COALESCE(equip,'')"
	}
	shop := "'',''"
	if c.hasShops {
		shop = "COALESCE(classtype,''),COALESCE(items_list,'')"
	}
	race := "0"
	if c.hasRules {
		race = "COALESCE(race,0)"
	}
	levels := "''"
	if c.hasLevels {
		levels = "COALESCE(levels,'')"
	}
	rows, err := c.db.Query(`SELECT id,zone,name,COALESCE(guild,''),model,x,y,z,heading,level,size,speed,template,roam,respawn_s,aggro_level,aggro_range,realm,flags,`+equip+`,`+shop+`,`+race+`,`+levels+` FROM spawn WHERE zone=?`, zone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Spawn
	for rows.Next() {
		s := Spawn{}
		var levels string
		if err := rows.Scan(&s.ID, &s.Zone, &s.Name, &s.Guild, &s.Model, &s.X, &s.Y, &s.Z, &s.Heading, &s.Level, &s.Size, &s.Speed, &s.Template, &s.Roam, &s.RespawnS, &s.AggroLevel, &s.AggroRange, &s.Realm, &s.Flags, &s.Equip, &s.ClassType, &s.ItemsList, &s.Race, &levels); err != nil {
			return nil, err
		}
		s.Levels = importer.LevelList(levels)
		out = append(out, s)
	}
	return out, rows.Err()
}

// StylesFor lists a class's styles in the order a bar shows them: by line,
// then by the level they come at.
func (c *Catalog) StylesFor(class int) []StyleDef {
	var out []StyleDef
	for _, s := range c.Styles {
		if s.Class == class {
			out = append(out, s)
		}
	}
	return out
}

// StyleByID finds a class's style by its ID column.
func (c *Catalog) StyleByID(class, id int) *StyleDef {
	for i := range c.Styles {
		if c.Styles[i].Class == class && c.Styles[i].ID == id {
			return &c.Styles[i]
		}
	}
	return nil
}

// StartFor picks where a new character of a realm begins, falling back to the
// first start in the realm.
func (c *Catalog) StartFor(realm, race, class int) (Start, bool) {
	var fallback *Start
	for i := range c.Starts {
		s := &c.Starts[i]
		if s.Realm != realm || !c.Zones[s.Zone].usable() {
			continue
		}
		if s.Race == race && s.Class == class {
			if u, ok := c.replaced[i]; ok {
				return u.Instead, true
			}
			return s.Start, true
		}
		if _, bad := c.replaced[i]; fallback == nil && !bad {
			fallback = &s.Start
		}
	}
	if fallback != nil {
		return *fallback, true
	}
	return Start{}, false
}

func (z *ZoneInfo) usable() bool { return z != nil && z.Converted }

// LootFor lists the loot tables a mob name draws from.
func (c *Catalog) LootFor(name string) []LootEntry {
	var out []LootEntry
	for _, t := range c.MobLoot[strings.ToLower(name)] {
		out = append(out, c.Loot[t]...)
	}
	return out
}

// EmptyCatalog is the world with no imported data: every converted zone,
// nothing in it. The server falls back to it when world.db is missing, so a
// fresh checkout can still be walked around.
func EmptyCatalog(webData string) *Catalog {
	c := &Catalog{Items: map[string]*Item{}, Templates: map[int]*Template{}, Zones: map[int]*ZoneInfo{},
		Loot: map[string][]LootEntry{}, MobLoot: map[string][]string{}, Stock: map[string][]string{}}
	b, err := os.ReadFile(filepath.Join(webData, "zones.json"))
	if err != nil {
		return c
	}
	var list []struct {
		Num  int    `json:"num"`
		Name string `json:"name"`
	}
	if json.Unmarshal(b, &list) == nil {
		for _, z := range list {
			c.Zones[z.Num] = &ZoneInfo{ID: z.Num, Region: 100, Name: z.Name, Converted: true}
		}
	}
	return c
}

// Close releases the store.
func (c *Catalog) CloseDB() {
	if c.db != nil {
		c.db.Close()
	}
}
