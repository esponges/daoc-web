package importer

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// Schema is this project's own world store. It is deliberately not the
// community database's shape: only what the server reads, in zone-local
// coordinates, with nothing copied from the emulator's code.
const Schema = `
CREATE TABLE zone(
  id INTEGER PRIMARY KEY, region INTEGER, name TEXT,
  offset_x INTEGER, offset_y INTEGER, width INTEGER, height INTEGER,
  water_level INTEGER, realm INTEGER, converted INTEGER);
CREATE TABLE npc_template(
  id INTEGER PRIMARY KEY, name TEXT, model INTEGER, size_lo INTEGER, size_hi INTEGER,
  level_lo INTEGER, level_hi INTEGER, speed INTEGER, aggro_level INTEGER, aggro_range INTEGER,
  tether INTEGER, parry INTEGER, evade INTEGER, block INTEGER, melee_type INTEGER);
CREATE TABLE spawn(
  id INTEGER PRIMARY KEY AUTOINCREMENT, zone INTEGER, name TEXT, guild TEXT, model INTEGER,
  x REAL, y REAL, z REAL, heading INTEGER, level INTEGER, size INTEGER, speed INTEGER,
  template INTEGER, roam INTEGER, respawn_s INTEGER, aggro_level INTEGER, aggro_range INTEGER,
  realm INTEGER, flags INTEGER, path TEXT, brain TEXT);
CREATE INDEX spawn_zone ON spawn(zone);
CREATE TABLE loot(template TEXT, item TEXT, chance INTEGER, count INTEGER);
CREATE INDEX loot_template ON loot(template);
CREATE TABLE mob_loot(mob_name TEXT, template TEXT, drop_count INTEGER);
CREATE INDEX mob_loot_name ON mob_loot(mob_name);
CREATE TABLE item(
  id TEXT PRIMARY KEY, name TEXT, level INTEGER, quality INTEGER, dps_af INTEGER, spd_abs INTEGER,
  hand INTEGER, damage_type INTEGER, object_type INTEGER, item_type INTEGER, model INTEGER,
  weight INTEGER, price INTEGER, realm INTEGER, allowed_classes TEXT, droppable INTEGER, level_req INTEGER);
CREATE TABLE style(
  id INTEGER, class_id INTEGER, name TEXT, spec TEXT, spec_level INTEGER, icon INTEGER,
  endurance INTEGER, opening_type INTEGER, opening_value INTEGER, result_req INTEGER,
  weapon_req INTEGER, growth_rate REAL, growth_offset REAL, bonus_hit INTEGER, bonus_def INTEGER,
  anim INTEGER, stealth INTEGER, armor_loc INTEGER, style_id INTEGER);
CREATE INDEX style_class ON style(class_id);
CREATE TABLE specialization(key TEXT PRIMARY KEY, name TEXT);
CREATE TABLE class_spec(class_id INTEGER, spec TEXT, level INTEGER);
CREATE TABLE zonepoint(
  id INTEGER, src_zone INTEGER, src_region INTEGER, sx REAL, sy REAL, sz REAL,
  dst_zone INTEGER, dst_region INTEGER, tx REAL, ty REAL, tz REAL, heading INTEGER, realm INTEGER);
CREATE TABLE startup(realm INTEGER, race INTEGER, class INTEGER, zone INTEGER, region INTEGER,
  x REAL, y REAL, z REAL, heading INTEGER);
CREATE TABLE bindpoint(zone INTEGER, region INTEGER, x REAL, y REAL, z REAL, radius INTEGER, realm INTEGER);
CREATE TABLE race(id INTEGER PRIMARY KEY, name TEXT);
`

// Options say where things are and which spawns to keep.
type Options struct {
	Dump   string       // the community database's table directory
	Out    string       // world.db to write (replaced)
	WebDir string       // converted zones, for the ground-height check; "" skips it
	Models map[int]bool // monster model numbers the client can draw; nil keeps all
	MaxDz  float64      // furthest a spawn may be from the ground
	Log    func(format string, a ...any)
}

// Report is what an import kept and dropped.
type Report struct {
	Tables         map[string][2]int // table -> rows read, rows kept
	DroppedModel   map[int]int       // model -> spawns dropped because the client cannot draw it
	DroppedGround  int               // spawns too far off the terrain
	DroppedOutside int               // spawns in no zone
	Missing        []string          // required tables absent or empty
}

type zoneRow struct {
	id, region, ox, oy, w, h int
	name                     string
}

// Run imports the dump into a fresh world store.
func Run(o Options) (*Report, error) {
	if o.Log == nil {
		o.Log = func(string, ...any) {}
	}
	if o.MaxDz == 0 {
		o.MaxDz = 500
	}
	rep := &Report{Tables: map[string][2]int{}, DroppedModel: map[int]int{}}
	if err := os.MkdirAll(filepath.Dir(o.Out), 0o755); err != nil {
		return nil, err
	}
	os.Remove(o.Out)
	db, err := sql.Open("sqlite", o.Out)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode=OFF; PRAGMA synchronous=OFF;`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(Schema); err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	read := func(table string, fn func(*Row) (bool, error)) error {
		path := filepath.Join(o.Dump, table+".sql")
		var kept int
		n, err := ReadTable(path, func(r *Row) error {
			ok, err := fn(r)
			if ok {
				kept++
			}
			return err
		})
		if err != nil {
			return fmt.Errorf("%s: %w", table, err)
		}
		rep.Tables[table] = [2]int{n, kept}
		o.Log("  %-22s %7d read %7d kept", table, n, kept)
		return nil
	}
	ins := func(q string, a ...any) error { _, err := tx.Exec(q, a...); return err }

	// --- zones ---
	var zones []zoneRow
	converted := convertedZones(o.WebDir)
	if err := read("zones", func(r *Row) (bool, error) {
		z := zoneRow{id: r.Int("ZoneID"), region: r.Int("RegionID"), ox: r.Int("OffsetX"), oy: r.Int("OffsetY"),
			w: r.Int("Width"), h: r.Int("Height"), name: r.Str("Name")}
		zones = append(zones, z)
		c := 0
		if converted[z.id] {
			c = 1
		}
		return true, ins(`INSERT OR REPLACE INTO zone VALUES(?,?,?,?,?,?,?,?,?,?)`, z.id, z.region, z.name,
			z.ox, z.oy, z.w, z.h, r.Int("WaterLevel"), r.Int("Realm"), c)
	}); err != nil {
		return nil, err
	}
	zoneAt := func(region int, x, y float64) (zoneRow, float64, float64, bool) {
		for _, z := range zones {
			if z.region != region || z.w == 0 {
				continue
			}
			lx, ly := x-float64(z.ox*8192), y-float64(z.oy*8192)
			if lx >= 0 && ly >= 0 && lx < float64(z.w*8192) && ly < float64(z.h*8192) {
				return z, lx, ly, true
			}
		}
		return zoneRow{}, 0, 0, false
	}

	// --- templates ---
	if err := read("npctemplate", func(r *Row) (bool, error) {
		sl, sh := r.Range("Size")
		ll, lh := r.Range("Level")
		ml, _ := r.Range("Model")
		return true, ins(`INSERT OR REPLACE INTO npc_template VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			r.Int("TemplateId"), r.Str("Name"), ml, sl, sh, ll, lh, r.Int("MaxSpeed"),
			r.Int("AggroLevel"), r.Int("AggroRange"), r.Int("TetherRange"),
			r.Int("ParryChance"), r.Int("EvadeChance"), r.Int("BlockChance"), r.Int("MeleeDamageType"))
	}); err != nil {
		return nil, err
	}

	// --- spawns ---
	heights := map[int][]uint16{}
	groundAt := func(zone int, lx, ly float64) (float64, bool) {
		if o.WebDir == "" {
			return 0, false
		}
		h, ok := heights[zone]
		if !ok {
			b, err := os.ReadFile(filepath.Join(o.WebDir, fmt.Sprintf("zone%03d", zone), "heights.u16"))
			if err == nil && len(b) == 256*256*2 {
				h = make([]uint16, 256*256)
				for i := range h {
					h[i] = binary.LittleEndian.Uint16(b[i*2:])
				}
			}
			heights[zone] = h
		}
		if h == nil {
			return 0, false
		}
		cx, cy := int(lx/256), int(ly/256)
		if cx < 0 || cy < 0 || cx > 255 || cy > 255 {
			return 0, false
		}
		return float64(h[cy*256+cx]), true
	}
	if err := read("mob", func(r *Row) (bool, error) {
		z, lx, ly, ok := zoneAt(r.Int("Region"), r.Float("X"), r.Float("Y"))
		if !ok {
			rep.DroppedOutside++
			return false, nil
		}
		model := r.Int("Model")
		if o.Models != nil && !o.Models[model] {
			rep.DroppedModel[model]++
			return false, nil
		}
		zz := r.Float("Z")
		if g, ok := groundAt(z.id, lx, ly); ok {
			if d := zz - g; d > o.MaxDz || d < -o.MaxDz {
				rep.DroppedGround++
				return false, nil
			}
		} else if o.WebDir != "" && converted != nil && !converted[z.id] {
			return false, nil // a zone we have not converted has nowhere to put it
		}
		return true, ins(`INSERT INTO spawn(zone,name,guild,model,x,y,z,heading,level,size,speed,template,roam,respawn_s,aggro_level,aggro_range,realm,flags,path,brain)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			z.id, r.Str("Name"), r.Str("Guild"), model, lx, ly, zz, r.Int("Heading"), r.Int("Level"), r.Int("Size"),
			r.Int("Speed"), r.Int("NPCTemplateID"), r.Int("RoamingRange"), r.Int("RespawnInterval"),
			r.Int("AggroLevel"), r.Int("AggroRange"), r.Int("Realm"), r.Int("Flags"), r.Str("PathID"), r.Str("Brain"))
	}); err != nil {
		return nil, err
	}

	// --- loot, items ---
	if err := read("loottemplate", func(r *Row) (bool, error) {
		return true, ins(`INSERT INTO loot VALUES(?,?,?,?)`, r.Str("TemplateName"), r.Str("ItemTemplateID"), r.Int("Chance"), r.Int("Count"))
	}); err != nil {
		return nil, err
	}
	if err := read("mobxloottemplate", func(r *Row) (bool, error) {
		return true, ins(`INSERT INTO mob_loot VALUES(?,?,?)`, strings.ToLower(r.Str("MobName")), r.Str("LootTemplateName"), r.Int("DropCount"))
	}); err != nil {
		return nil, err
	}
	if err := read("itemtemplate", func(r *Row) (bool, error) {
		return true, ins(`INSERT OR REPLACE INTO item VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			r.Str("Id_nb"), r.Str("Name"), r.Int("Level"), r.Int("Quality"), r.Int("DPS_AF"), r.Int("SPD_ABS"),
			r.Int("Hand"), r.Int("Type_Damage"), r.Int("Object_Type"), r.Int("Item_Type"), r.Int("Model"),
			r.Int("Weight"), r.Int("Price"), r.Int("Realm"), r.Str("AllowedClasses"), r.Int("CanDropAsLoot"), r.Int("LevelRequirement"))
	}); err != nil {
		return nil, err
	}

	// --- styles, specialisations, classes ---
	if err := read("style", func(r *Row) (bool, error) {
		return true, ins(`INSERT INTO style VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			r.Int("ID"), r.Int("ClassId"), r.Str("Name"), r.Str("SpecKeyName"), r.Int("SpecLevelRequirement"), r.Int("Icon"),
			r.Int("EnduranceCost"), r.Int("OpeningRequirementType"), r.Int("OpeningRequirementValue"), r.Int("AttackResultRequirement"),
			r.Int("WeaponTypeRequirement"), r.Float("GrowthRate"), r.Float("GrowthOffset"), r.Int("BonusToHit"), r.Int("BonusToDefense"),
			r.Int("TwoHandAnimation"), r.Int("StealthRequirement"), r.Int("ArmorHitLocation"), r.Int("StyleID"))
	}); err != nil {
		return nil, err
	}
	if err := read("specialization", func(r *Row) (bool, error) {
		return true, ins(`INSERT OR REPLACE INTO specialization VALUES(?,?)`, r.Str("KeyName"), r.Str("Name"))
	}); err != nil {
		return nil, err
	}
	if err := read("classxspecialization", func(r *Row) (bool, error) {
		return true, ins(`INSERT INTO class_spec VALUES(?,?,?)`, r.Int("ClassID"), r.Str("SpecKeyName"), r.Int("LevelAcquired"))
	}); err != nil {
		return nil, err
	}
	if err := read("race", func(r *Row) (bool, error) {
		return true, ins(`INSERT OR REPLACE INTO race VALUES(?,?)`, r.Int("ID"), r.Str("Name"))
	}); err != nil {
		return nil, err
	}

	// --- travel: exits, starts, bind points ---
	if err := read("zonepoint", func(r *Row) (bool, error) {
		sz, slx, sly, sok := zoneAt(r.Int("SourceRegion"), r.Float("SourceX"), r.Float("SourceY"))
		dz, dlx, dly, dok := zoneAt(r.Int("TargetRegion"), r.Float("TargetX"), r.Float("TargetY"))
		if !sok || !dok {
			return false, nil
		}
		return true, ins(`INSERT INTO zonepoint VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, r.Int("Id"), sz.id, sz.region, slx, sly, r.Float("SourceZ"),
			dz.id, dz.region, dlx, dly, r.Float("TargetZ"), r.Int("TargetHeading"), r.Int("Realm"))
	}); err != nil {
		return nil, err
	}
	if err := read("startuplocation", func(r *Row) (bool, error) {
		z, lx, ly, ok := zoneAt(r.Int("Region"), r.Float("XPos"), r.Float("YPos"))
		if !ok {
			return false, nil
		}
		return true, ins(`INSERT INTO startup VALUES(?,?,?,?,?,?,?,?,?)`, r.Int("RealmID"), r.Int("RaceID"), r.Int("ClassID"),
			z.id, z.region, lx, ly, r.Float("ZPos"), r.Int("Heading"))
	}); err != nil {
		return nil, err
	}
	if err := read("bindpoint", func(r *Row) (bool, error) {
		z, lx, ly, ok := zoneAt(r.Int("Region"), r.Float("X"), r.Float("Y"))
		if !ok {
			return false, nil
		}
		return true, ins(`INSERT INTO bindpoint VALUES(?,?,?,?,?,?,?)`, z.id, z.region, lx, ly, r.Float("Z"), r.Int("Radius"), r.Int("Realm"))
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	for _, t := range []string{"zones", "mob", "npctemplate", "style", "itemtemplate", "startuplocation"} {
		if c := rep.Tables[t]; c[0] == 0 || c[1] == 0 {
			rep.Missing = append(rep.Missing, t)
		}
	}
	sort.Strings(rep.Missing)
	return rep, nil
}

// convertedZones reads which zone folders exist under web/data. nil means unknown.
func convertedZones(web string) map[int]bool {
	if web == "" {
		return nil
	}
	ents, err := os.ReadDir(web)
	if err != nil {
		return nil
	}
	m := map[int]bool{}
	for _, e := range ents {
		var n int
		if e.IsDir() && strings.HasPrefix(e.Name(), "zone") {
			if _, err := fmt.Sscanf(e.Name(), "zone%d", &n); err == nil {
				if _, err := os.Stat(filepath.Join(web, e.Name(), "zone.json")); err == nil {
					m[n] = true
				}
			}
		}
	}
	return m
}
