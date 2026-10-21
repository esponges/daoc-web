package importer

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
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
  tether INTEGER, parry INTEGER, evade INTEGER, block INTEGER, melee_type INTEGER,
  levels TEXT, replace_values INTEGER);
CREATE TABLE spawn(
  id INTEGER PRIMARY KEY AUTOINCREMENT, zone INTEGER, name TEXT, guild TEXT, model INTEGER,
  x REAL, y REAL, z REAL, heading INTEGER, level INTEGER, size INTEGER, speed INTEGER,
  template INTEGER, roam INTEGER, respawn_s INTEGER, aggro_level INTEGER, aggro_range INTEGER,
  realm INTEGER, flags INTEGER, path TEXT, brain TEXT, equip TEXT, classtype TEXT, items_list TEXT, race INTEGER,
  levels TEXT, timed INTEGER);
CREATE INDEX spawn_zone ON spawn(zone);
CREATE TABLE merchant_stock(list TEXT, item TEXT, page INTEGER, slot INTEGER);
CREATE INDEX merchant_stock_list ON merchant_stock(list);
CREATE TABLE npc_equip(
  template TEXT, slot INTEGER, model INTEGER, color INTEGER, effect INTEGER, ext INTEGER, emblem INTEGER);
CREATE INDEX npc_equip_template ON npc_equip(template);
CREATE TABLE loot(template TEXT, item TEXT, chance INTEGER, count INTEGER);
CREATE INDEX loot_template ON loot(template);
CREATE TABLE mob_loot(mob_name TEXT, template TEXT, drop_count INTEGER);
CREATE INDEX mob_loot_name ON mob_loot(mob_name);
CREATE TABLE item(
  id TEXT PRIMARY KEY, name TEXT, level INTEGER, quality INTEGER, dps_af INTEGER, spd_abs INTEGER,
  hand INTEGER, damage_type INTEGER, object_type INTEGER, item_type INTEGER, model INTEGER,
  weight INTEGER, price INTEGER, realm INTEGER, allowed_classes TEXT, droppable INTEGER, level_req INTEGER,
  bonus INTEGER, condition INTEGER, max_condition INTEGER, durability INTEGER,
  proc_spell INTEGER, proc_chance INTEGER);
CREATE TABLE item_bonus(item TEXT, type INTEGER, amount INTEGER);
CREATE INDEX item_bonus_item ON item_bonus(item);
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
CREATE TABLE race(id INTEGER PRIMARY KEY, name TEXT,
  body INTEGER, cold INTEGER, crush INTEGER, energy INTEGER, heat INTEGER, matter INTEGER, natural INTEGER,
  slash INTEGER, spirit INTEGER, thrust INTEGER);
CREATE TABLE spec_ability(spec TEXT, spec_level INTEGER, ability TEXT, ability_level INTEGER, class_id INTEGER);
CREATE TABLE start_equip(class INTEGER, item TEXT, ord INTEGER);
CREATE TABLE route(path TEXT PRIMARY KEY, region INTEGER, start_zone INTEGER, points INTEGER);
CREATE TABLE route_point(path TEXT, step INTEGER, zone INTEGER, x REAL, y REAL, z REAL, speed INTEGER, wait INTEGER);
CREATE INDEX route_point_path ON route_point(path, step);
CREATE TABLE spell(
  id INTEGER PRIMARY KEY, name TEXT, type TEXT, target TEXT, range INTEGER, cast_time REAL, recast INTEGER,
  damage REAL, damage_type INTEGER, value REAL, duration INTEGER, frequency INTEGER, life_drain INTEGER,
  sub_spell INTEGER, uninterruptible INTEGER, icon INTEGER, radius INTEGER,
  message1 TEXT, message2 TEXT, message3 TEXT, message4 TEXT, power INTEGER);
CREATE TABLE line_spell(spec TEXT, hint INTEGER, line TEXT, spell INTEGER, level INTEGER);
CREATE INDEX line_spell_spec ON line_spell(spec);
CREATE TABLE template_spell(template INTEGER, spell INTEGER, ord INTEGER);
CREATE INDEX template_spell_template ON template_spell(template);
CREATE TABLE style_spell(class INTEGER, style INTEGER, spell INTEGER, chance INTEGER);
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
	Tables            map[string][2]int // table -> rows read, rows kept
	DroppedModel      map[int]int       // model -> spawns dropped because the client cannot draw it
	DroppedGround     int               // spawns too far off the terrain
	DroppedOutside    int               // spawns in no zone
	UnknownEquip      int               // spawns whose equipment template has no rows
	MerchantsSeen     int               // mobs of the merchant class type in the dump
	MerchantsKept     int               // of those, the ones that became spawns
	Missing           []string          // required tables absent or empty
	Staged            string            // where the store was left when Out was in use, else ""
	Replaced          int               // spawns whose template replaces the row's values
	LevelReplaced     int               // of those, the ones whose own level is not in the template's list
	SpellLinksMissing int               // template and style links to spells the dump lacks
	Routes            int               // horse routes kept: a stable master sells a ticket named for a path whose points all lie in a zone
	RoutesDropped     int               // sold paths with a point outside every zone, or no points
}

// tmplValues are what a template marked ReplaceMobValues puts in place of a
// spawn row's own values.
type tmplValues struct {
	name, classType                      string
	levels                               []int
	aggroLevel, aggroRange, speed, flags int
}

// NumberList reads a list column, such as a template's levels or a class
// list: one number, a list ("13;18;16") or a range ("1-3"), in any mix.
// Anything unreadable is skipped.
func NumberList(s string) []int {
	var out []int
	for _, p := range strings.FieldsFunc(s, func(r rune) bool { return r == ';' || r == ',' }) {
		p = strings.TrimSpace(p)
		if a, b, ok := strings.Cut(p, "-"); ok && a != "" {
			lo, e1 := strconv.Atoi(strings.TrimSpace(a))
			hi, e2 := strconv.Atoi(strings.TrimSpace(b))
			if e1 == nil && e2 == nil && lo <= hi && hi-lo < 100 {
				for l := lo; l <= hi; l++ {
					out = append(out, l)
				}
			}
			continue
		}
		if v, err := strconv.Atoi(p); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func joinLevels(l []int) string {
	s := make([]string, len(l))
	for i, v := range l {
		s[i] = strconv.Itoa(v)
	}
	return strings.Join(s, ";")
}

type zoneRow struct {
	id, region, ox, oy, w, h int
	name                     string
}

// Run imports the dump into a fresh world store. It builds the store beside
// Out and then moves it into place; when Out is held open (a running server
// on Windows), the new store is left at Out+".next" instead, for the server
// to swap in when it next starts. A failed import leaves Out as it was.
func Run(o Options) (*Report, error) {
	final := o.Out
	o.Out = final + ".tmp"
	rep, err := build(o)
	if err != nil {
		os.Remove(o.Out)
		return nil, err
	}
	if rep.Staged, err = place(o.Out, final); err != nil {
		os.Remove(o.Out)
		return nil, err
	}
	return rep, nil
}

// place moves a finished store to path, or to path+".next" when path cannot
// be replaced, and returns the staged path in that case. A stale staged store
// is removed once path itself is current, so a restart cannot swap it back in.
func place(tmp, path string) (string, error) {
	next := path + ".next"
	if err := os.Rename(tmp, path); err == nil {
		os.Remove(next)
		return "", nil
	}
	os.Remove(next)
	if err := os.Rename(tmp, next); err != nil {
		return "", err
	}
	return next, nil
}

func build(o Options) (*Report, error) {
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

	// --- equipment ---
	equipOf := map[string]bool{}
	if err := read("npcequipment", func(r *Row) (bool, error) {
		id := r.Str("TemplateID")
		if id == "" {
			return false, nil
		}
		equipOf[id] = true
		return true, ins(`INSERT INTO npc_equip VALUES(?,?,?,?,?,?,?)`, id, r.Int("Slot"), r.Int("Model"), r.Int("Color"), r.Int("Effect"), r.Int("Extension"), r.Int("Emblem"))
	}); err != nil {
		return nil, err
	}

	// --- templates ---
	tmplEquip := map[int]string{} // npc template id -> equipment template id
	replaces := map[int]tmplValues{}
	tmplSpells := map[int][]int{} // npc template id -> the spells it casts, in order
	wantSpell := map[int]bool{}   // spells something kept names
	if err := read("npctemplate", func(r *Row) (bool, error) {
		id := r.Int("TemplateId")
		sl, sh := r.Range("Size")
		levels := NumberList(r.Str("Level"))
		ll, lh := 0, 0
		if len(levels) > 0 {
			ll, lh = slices.Min(levels), slices.Max(levels)
		}
		ml, _ := r.Range("Model")
		if e := r.Str("EquipmentTemplateID"); e != "" {
			tmplEquip[id] = e
		}
		if s := NumberList(r.Str("Spells")); len(s) > 0 {
			tmplSpells[id] = s
			for _, sp := range s {
				wantSpell[sp] = true
			}
		}
		replace := r.Int("ReplaceMobValues") != 0
		if replace {
			replaces[id] = tmplValues{name: r.Str("Name"), levels: levels, aggroLevel: r.Int("AggroLevel"), aggroRange: r.Int("AggroRange"),
				speed: r.Int("MaxSpeed"), flags: r.Int("Flags"), classType: r.Str("ClassType")}
		} else {
			delete(replaces, id)
		}
		return true, ins(`INSERT OR REPLACE INTO npc_template VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			id, r.Str("Name"), ml, sl, sh, ll, lh, r.Int("MaxSpeed"),
			r.Int("AggroLevel"), r.Int("AggroRange"), r.Int("TetherRange"),
			r.Int("ParryChance"), r.Int("EvadeChance"), r.Int("BlockChance"), r.Int("MeleeDamageType"),
			joinLevels(levels), replace)
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
		classType := r.Str("ClassType")
		name, level, levels := r.Str("Name"), r.Int("Level"), ""
		aggroLevel, aggroRange, speed, flags := r.Int("AggroLevel"), r.Int("AggroRange"), r.Int("Speed"), r.Int("Flags")
		replaced, outside := false, false
		// A template marked to replace mob values overrides the row, as the
		// emulator this data is kept for loads it: the level is then picked
		// from the template's list each time the creature spawns.
		if t, ok := replaces[r.Int("NPCTemplateID")]; ok {
			replaced = true
			if t.name != "" {
				name = t.name
			}
			if len(t.levels) > 0 {
				if !slices.Contains(t.levels, level) {
					outside = true
				}
				level, levels = slices.Min(t.levels), joinLevels(t.levels)
			}
			aggroLevel, aggroRange, flags = t.aggroLevel, t.aggroRange, t.flags
			if t.speed > 0 {
				speed = t.speed
			}
			if t.classType != "" && t.classType != "DOL.GS.GameNPC" {
				classType = t.classType
			}
		}
		merchant := IsMerchantClass(classType)
		if merchant {
			rep.MerchantsSeen++
		}
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
		// A mob's own equipment wins; otherwise its NPC template's.
		equip := r.Str("EquipmentTemplateID")
		if equip == "" {
			equip = tmplEquip[r.Int("NPCTemplateID")]
		}
		if equip != "" && !equipOf[equip] {
			rep.UnknownEquip++
			equip = ""
		}
		if merchant {
			rep.MerchantsKept++
		}
		if replaced {
			rep.Replaced++
		}
		if outside {
			rep.LevelReplaced++
		}
		return true, ins(`INSERT INTO spawn(zone,name,guild,model,x,y,z,heading,level,size,speed,template,roam,respawn_s,aggro_level,aggro_range,realm,flags,path,brain,equip,classtype,items_list,race,levels,timed)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			z.id, name, r.Str("Guild"), model, lx, ly, zz, r.Int("Heading"), level, r.Int("Size"),
			speed, r.Int("NPCTemplateID"), r.Int("RoamingRange"), r.Int("RespawnInterval"),
			aggroLevel, aggroRange, r.Int("Realm"), flags, r.Str("PathID"), r.Str("Brain"), equip,
			classType, r.Str("ItemsListTemplateID"), r.Int("Race"), levels, SpawnTiming(classType))
	}); err != nil {
		return nil, err
	}
	if err := read("merchantitem", func(r *Row) (bool, error) {
		list, item := r.Str("ItemListID"), r.Str("ItemTemplateID")
		if list == "" || item == "" {
			return false, nil
		}
		return true, ins(`INSERT INTO merchant_stock VALUES(?,?,?,?)`, list, item, r.Int("PageNumber"), r.Int("SlotPosition"))
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
		id := r.Str("Id_nb")
		if p := r.Int("ProcSpellID"); p != 0 {
			wantSpell[p] = true
		}
		for i := 1; i <= 10; i++ {
			ty, amt := r.Int(fmt.Sprintf("Bonus%dType", i)), r.Int(fmt.Sprintf("Bonus%d", i))
			if ty != 0 && amt != 0 {
				if err := ins(`INSERT INTO item_bonus VALUES(?,?,?)`, id, ty, amt); err != nil {
					return false, err
				}
			}
		}
		return true, ins(`INSERT OR REPLACE INTO item VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			r.Str("Id_nb"), r.Str("Name"), r.Int("Level"), r.Int("Quality"), r.Int("DPS_AF"), r.Int("SPD_ABS"),
			r.Int("Hand"), r.Int("Type_Damage"), r.Int("Object_Type"), r.Int("Item_Type"), r.Int("Model"),
			r.Int("Weight"), r.Int("Price"), r.Int("Realm"), r.Str("AllowedClasses"), r.Int("CanDropAsLoot"), r.Int("LevelRequirement"),
			r.Int("Bonus"), r.Int("Condition"), r.Int("MaxCondition"), r.Int("Durability"), r.Int("ProcSpellID"), r.Int("ProcChance"))
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
	midgardSpecs := map[string]bool{} // spec lines the Midgard classes (21 to 38) train
	if err := read("classxspecialization", func(r *Row) (bool, error) {
		if c := r.Int("ClassID"); c >= 21 && c <= 38 {
			midgardSpecs[r.Str("SpecKeyName")] = true
		}
		return true, ins(`INSERT INTO class_spec VALUES(?,?,?)`, r.Int("ClassID"), r.Str("SpecKeyName"), r.Int("LevelAcquired"))
	}); err != nil {
		return nil, err
	}
	if err := read("race", func(r *Row) (bool, error) {
		return true, ins(`INSERT OR REPLACE INTO race VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, r.Int("ID"), r.Str("Name"),
			r.Int("ResistBody"), r.Int("ResistCold"), r.Int("ResistCrush"), r.Int("ResistEnergy"), r.Int("ResistHeat"),
			r.Int("ResistMatter"), r.Int("ResistNatural"), r.Int("ResistSlash"), r.Int("ResistSpirit"), r.Int("ResistThrust"))
	}); err != nil {
		return nil, err
	}

	// Which abilities each career and spec grants, by level. Older dumps may
	// lack the table; the server then knows no abilities and says so.
	if _, err := os.Stat(filepath.Join(o.Dump, "specxability.sql")); err == nil {
		if err := read("specxability", func(r *Row) (bool, error) {
			if r.Str("Spec") == "" || r.Str("AbilityKey") == "" {
				return false, nil
			}
			return true, ins(`INSERT INTO spec_ability VALUES(?,?,?,?,?)`, r.Str("Spec"), r.Int("SpecLevel"), r.Str("AbilityKey"),
				r.Int("AbilityLevel"), r.Int("ClassId"))
		}); err != nil {
			return nil, err
		}
	} else {
		o.Log("  specxability           absent: no abilities imported")
	}

	// What each class is given at creation: one row per class the dump's
	// class list names. A missing table is reported with the required ones.
	if _, err := os.Stat(filepath.Join(o.Dump, "starterequipment.sql")); err == nil {
		if err := read("starterequipment", func(r *Row) (bool, error) {
			item := r.Str("TemplateID")
			classes := NumberList(r.Str("Class"))
			if item == "" || len(classes) == 0 {
				return false, nil
			}
			for _, c := range classes {
				if err := ins(`INSERT INTO start_equip VALUES(?,?,?)`, c, item, r.Int("StarterEquipmentID")); err != nil {
					return false, err
				}
			}
			return true, nil
		}); err != nil {
			return nil, err
		}
	}

	// --- spells: those templates, styles and items name, and their sub-spells ---
	type styleSpell struct{ class, style, spell, chance int }
	var styleSpells []styleSpell
	if _, err := os.Stat(filepath.Join(o.Dump, "stylexspell.sql")); err == nil {
		if err := read("stylexspell", func(r *Row) (bool, error) {
			s := styleSpell{r.Int("ClassID"), r.Int("StyleID"), r.Int("SpellID"), r.Int("Chance")}
			if s.spell == 0 || s.style == 0 {
				return false, nil
			}
			styleSpells = append(styleSpells, s)
			wantSpell[s.spell] = true
			return true, nil
		}); err != nil {
			return nil, err
		}
	}
	// The spells of the lines those classes train: a line belongs to a spec, and
	// may be meant for one class (its hint) or for every class with the spec.
	type lineSpell struct {
		spec, line         string
		hint, spell, level int
	}
	var lineSpells []lineSpell
	if _, e1 := os.Stat(filepath.Join(o.Dump, "linexspell.sql")); e1 == nil && fileExists(filepath.Join(o.Dump, "spellline.sql")) {
		lines := map[string]lineSpell{}
		if err := read("spellline", func(r *Row) (bool, error) {
			if sp := r.Str("Spec"); midgardSpecs[sp] {
				lines[r.Str("KeyName")] = lineSpell{spec: sp, line: r.Str("KeyName"), hint: r.Int("ClassIDHint")}
			}
			return false, nil
		}); err != nil {
			return nil, err
		}
		if err := read("linexspell", func(r *Row) (bool, error) {
			l, ok := lines[r.Str("LineName")]
			if !ok || r.Int("SpellID") == 0 || r.Int("Level") < 1 || r.Int("Level") > 50 {
				return false, nil
			}
			l.spell, l.level = r.Int("SpellID"), r.Int("Level")
			lineSpells = append(lineSpells, l)
			wantSpell[l.spell] = true
			return true, nil
		}); err != nil {
			return nil, err
		}
	}
	spells := map[int]*Row{}
	if _, err := os.Stat(filepath.Join(o.Dump, "spell.sql")); err == nil {
		if _, err := ReadTable(filepath.Join(o.Dump, "spell.sql"), func(r *Row) error {
			spells[r.Int("SpellID")] = r
			return nil
		}); err != nil {
			return nil, fmt.Errorf("spell: %w", err)
		}
	}
	// A spell's sub-spell, and the spell an offensive proc fires (its value),
	// come with it.
	for queue := slices.Sorted(maps.Keys(wantSpell)); len(queue) > 0; queue = queue[1:] {
		r := spells[queue[0]]
		if r == nil {
			continue
		}
		next := []int{r.Int("SubSpellID")}
		if t := r.Str("Type"); t == "OffensiveProc" || t == "DefensiveProc" {
			next = append(next, r.Int("Value"))
		}
		for _, id := range next {
			if id != 0 && !wantSpell[id] {
				wantSpell[id] = true
				queue = append(queue, id)
			}
		}
	}
	kept := 0
	for _, id := range slices.Sorted(maps.Keys(wantSpell)) {
		r := spells[id]
		if r == nil {
			continue
		}
		kept++
		if err := ins(`INSERT INTO spell VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, id, r.Str("Name"), r.Str("Type"), r.Str("Target"),
			r.Int("Range"), r.Float("CastTime"), r.Int("RecastDelay"), r.Float("Damage"), r.Int("DamageType"), r.Float("Value"),
			r.Int("Duration"), r.Int("Frequency"), r.Int("LifeDrainReturn"), r.Int("SubSpellID"), r.Int("Uninterruptible"),
			r.Int("Icon"), r.Int("Radius"), r.Str("Message1"), r.Str("Message2"), r.Str("Message3"), r.Str("Message4"), r.Int("Power")); err != nil {
			return nil, err
		}
	}
	rep.Tables["spell"] = [2]int{len(spells), kept}
	o.Log("  %-22s %7d read %7d kept", "spell", len(spells), kept)
	links := 0
	for _, t := range slices.Sorted(maps.Keys(tmplSpells)) {
		for i, sp := range tmplSpells[t] {
			if spells[sp] == nil {
				rep.SpellLinksMissing++
				continue
			}
			links++
			if err := ins(`INSERT INTO template_spell VALUES(?,?,?)`, t, sp, i); err != nil {
				return nil, err
			}
		}
	}
	for _, l := range lineSpells {
		if spells[l.spell] == nil {
			rep.SpellLinksMissing++
			continue
		}
		if err := ins(`INSERT INTO line_spell VALUES(?,?,?,?,?)`, l.spec, l.hint, l.line, l.spell, l.level); err != nil {
			return nil, err
		}
	}
	for _, s := range styleSpells {
		if spells[s.spell] == nil {
			rep.SpellLinksMissing++
			continue
		}
		if err := ins(`INSERT INTO style_spell VALUES(?,?,?,?)`, s.class, s.style, s.spell, s.chance); err != nil {
			return nil, err
		}
	}
	o.Log("  %-22s %7d template spell links, %d style spells, %d links to spells the dump lacks", "spell links", links, len(styleSpells), rep.SpellLinksMissing)

	// --- horse routes ---
	// A ticket is an item whose id is also the id of a path (checked over the
	// dump: 216 of its 223 tickets); a stable master sells the tickets of its
	// stock list. A path's points are in region coordinates and have no region
	// of their own, so a route takes the region of the stable that sells it.
	pathName := map[string]string{}
	if err := read("path", func(r *Row) (bool, error) {
		n := r.Str("PathID")
		if n == "" {
			return false, nil
		}
		pathName[strings.ToLower(n)] = n
		return true, nil
	}); err != nil {
		return nil, err
	}
	type pathPoint struct {
		step, speed, wait int
		x, y, z           float64
	}
	points := map[string][]pathPoint{}
	if err := read("pathpoints", func(r *Row) (bool, error) {
		n := strings.ToLower(r.Str("PathID"))
		if _, ok := pathName[n]; !ok {
			return false, nil
		}
		points[n] = append(points[n], pathPoint{r.Int("Step"), r.Int("MaxSpeed"), r.Int("WaitTime"), r.Float("X"), r.Float("Y"), r.Float("Z")})
		return true, nil
	}); err != nil {
		return nil, err
	}
	regionOf := map[int]int{}
	for _, z := range zones {
		regionOf[z.id] = z.region
	}
	type sold struct {
		zone int
		item string
	}
	var tickets []sold
	seen := map[string]bool{}
	rows, err := tx.Query(`SELECT DISTINCT s.zone, m.item FROM spawn s JOIN merchant_stock m ON m.list = s.items_list
		WHERE s.classtype = 'DOL.GS.GameStableMaster' ORDER BY s.zone, m.item`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var t sold
		if err := rows.Scan(&t.zone, &t.item); err != nil {
			rows.Close()
			return nil, err
		}
		tickets = append(tickets, t)
	}
	rows.Close()
	for _, t := range tickets {
		key := strings.ToLower(t.item)
		if _, ok := pathName[key]; !ok {
			continue // not a route: something else a stable master sells
		}
		pts := points[key]
		sort.Slice(pts, func(i, j int) bool { return pts[i].step < pts[j].step })
		if seen[key] {
			continue
		}
		seen[key] = true
		type placed struct {
			zone    int
			x, y, z float64
			p       pathPoint
		}
		var out []placed
		for _, p := range pts {
			z, lx, ly, ok := zoneAt(regionOf[t.zone], p.x, p.y)
			if !ok {
				out = nil
				break
			}
			out = append(out, placed{z.id, lx, ly, p.z, p})
		}
		if len(out) == 0 {
			rep.RoutesDropped++
			continue
		}
		if err := ins(`INSERT INTO route VALUES(?,?,?,?)`, t.item, regionOf[t.zone], t.zone, len(out)); err != nil {
			return nil, err
		}
		for _, p := range out {
			if err := ins(`INSERT INTO route_point VALUES(?,?,?,?,?,?,?,?)`, t.item, p.p.step, p.zone, p.x, p.y, p.z, p.p.speed, p.p.wait); err != nil {
				return nil, err
			}
		}
		rep.Routes++
	}
	o.Log("  %-22s %7d horse routes kept, %d dropped (a point outside every zone)", "routes", rep.Routes, rep.RoutesDropped)

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
	for _, t := range []string{"zones", "mob", "npctemplate", "npcequipment", "style", "itemtemplate", "startuplocation", "merchantitem", "starterequipment", "path", "pathpoints"} {
		if c := rep.Tables[t]; c[0] == 0 || c[1] == 0 {
			rep.Missing = append(rep.Missing, t)
		}
	}
	sort.Strings(rep.Missing)
	return rep, nil
}

// When a spawn is present. A creature of class type NightSpawn is there only at
// night, and one of class type DaySpawn only by day; everything else is
// always there.
const (
	Always    = 0
	NightOnly = 1
	DayOnly   = 2
)

// SpawnTiming says when a spawn of this class type is present.
func SpawnTiming(classType string) int {
	switch classType {
	case "DOL.GS.NightSpawn":
		return NightOnly
	case "DOL.GS.DaySpawn":
		return DayOnly
	}
	return Always
}

// IsMerchantClass reports whether a mob's class type is the plain merchant.
// Subclasses with their own behaviour (smiths, healers, trainers) are left
// alone until something builds what they do.
func IsMerchantClass(classType string) bool {
	return classType == "DOL.GS.GameMerchant"
}

// IsStableClass reports whether a mob's class type is a stable master, who
// sells horse tickets and rides them.
func IsStableClass(classType string) bool {
	return classType == "DOL.GS.GameStableMaster"
}

// IsVaultClass reports whether a mob's class type is a vault keeper. The
// account vault keeper is another class type and is not classic.
func IsVaultClass(classType string) bool {
	return classType == "DOL.GS.GameVaultKeeper"
}

// IsSmithClass reports whether a mob's class type is a blacksmith, who
// repairs worn items for money.
func IsSmithClass(classType string) bool {
	return classType == "DOL.GS.Blacksmith"
}

// IsHealerClass reports whether a mob's class type is a healer, who restores
// constitution lost to deaths.
func IsHealerClass(classType string) bool {
	return classType == "DOL.GS.GameHealer"
}

// TrainerClass is the class name a trainer's class type teaches, such as
// "Warrior" for DOL.GS.Trainer.WarriorTrainer, or "" for anything else.
func TrainerClass(classType string) string {
	const pre, suf = "DOL.GS.Trainer.", "Trainer"
	if !strings.HasPrefix(classType, pre) || !strings.HasSuffix(classType, suf) {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(classType, pre), suf)
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

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }
