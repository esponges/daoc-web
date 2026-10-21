package world

import (
	"database/sql"

	"daocweb/internal/server/rules"
)

// SpellDef is a spell as the server casts it: the imported values, in the
// units the data gives (cast time and recast in seconds, duration in seconds,
// frequency in tenths of a second).
type SpellDef struct {
	ID              int
	Name            string
	Type            string
	Target          string // "Enemy", "Self", "Realm", "Pet", ...
	Range           float64
	CastTime        float64
	Recast          float64
	Damage          float64
	DamageType      int
	Value           float64
	Duration        int
	Frequency       int
	LifeDrain       int
	SubSpell        int
	Uninterruptible bool
	Icon            int
	Power           int       // a percentage of the caster's pool; 0 costs nothing
	Messages        [4]string // to the target, to others, when it ends to the target, to others
}

// Hostile says whether a spell is cast at an enemy.
func (s *SpellDef) Hostile() bool {
	return s.Target == "Enemy" || s.Target == "Area" || s.Target == "Cone"
}

// StyleSpell is a spell a style fires when it lands: for one class, or for
// every class when Class is 0.
type StyleSpell struct {
	Class, Style, Spell, Chance int
}

// LineSpell is a spell of a spell line a spec trains: it opens when the spec
// reaches Level. A line with a Hint is meant for that one class.
type LineSpell struct {
	Spec, Line string
	Hint       int
	Spell      int
	Level      int
}

// SpellsFor lists the spells a class may know from a spec line: those of the
// lines every class with the spec shares, and those of the lines made for this
// class, which carry on from the shared ones. Each comes with the spec level
// it opens at.
func (c *Catalog) SpellsFor(class int, spec string) []LineSpell {
	var out []LineSpell
	for _, l := range c.LineSpells[spec] {
		if l.Hint == 0 || l.Hint == class {
			out = append(out, l)
		}
	}
	return out
}

// loadSpells reads the spells, the template spell lists, the style spells
// and the item procs. A store imported before spells has none of them.
func (c *Catalog) loadSpells(db *sql.DB, q func(string, func(*sql.Rows) error) error) error {
	c.Spells = map[int]*SpellDef{}
	c.TemplateSpells = map[int][]int{}
	c.LineSpells = map[string][]LineSpell{}
	var has int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('spell','template_spell','style_spell')`).Scan(&has)
	if has != 3 {
		return nil
	}
	powerCol := "0"
	var hasPower int
	db.QueryRow(`SELECT count(*) FROM pragma_table_info('spell') WHERE name='power'`).Scan(&hasPower)
	if hasPower > 0 {
		powerCol = "COALESCE(power,0)"
	}
	if err := q(`SELECT id,name,type,target,range,cast_time,recast,damage,damage_type,value,duration,frequency,life_drain,
		sub_spell,uninterruptible,icon,COALESCE(message1,''),COALESCE(message2,''),COALESCE(message3,''),COALESCE(message4,''),`+powerCol+` FROM spell`, func(r *sql.Rows) error {
		s := &SpellDef{}
		var unint int
		if err := r.Scan(&s.ID, &s.Name, &s.Type, &s.Target, &s.Range, &s.CastTime, &s.Recast, &s.Damage, &s.DamageType, &s.Value,
			&s.Duration, &s.Frequency, &s.LifeDrain, &s.SubSpell, &unint, &s.Icon, &s.Messages[0], &s.Messages[1], &s.Messages[2], &s.Messages[3], &s.Power); err != nil {
			return err
		}
		s.Uninterruptible = unint != 0
		c.Spells[s.ID] = s
		return nil
	}); err != nil {
		return err
	}
	if err := q(`SELECT template,spell FROM template_spell ORDER BY template,ord`, func(r *sql.Rows) error {
		var t, s int
		if err := r.Scan(&t, &s); err != nil {
			return err
		}
		c.TemplateSpells[t] = append(c.TemplateSpells[t], s)
		return nil
	}); err != nil {
		return err
	}
	var hasLines int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='line_spell'`).Scan(&hasLines)
	if hasLines > 0 {
		if err := q(`SELECT spec,hint,line,spell,level FROM line_spell ORDER BY spec,level,spell`, func(r *sql.Rows) error {
			var l LineSpell
			if err := r.Scan(&l.Spec, &l.Hint, &l.Line, &l.Spell, &l.Level); err != nil {
				return err
			}
			c.LineSpells[l.Spec] = append(c.LineSpells[l.Spec], l)
			return nil
		}); err != nil {
			return err
		}
	}
	if err := q(`SELECT class,style,spell,chance FROM style_spell`, func(r *sql.Rows) error {
		var s StyleSpell
		if err := r.Scan(&s.Class, &s.Style, &s.Spell, &s.Chance); err != nil {
			return err
		}
		c.StyleSpells = append(c.StyleSpells, s)
		return nil
	}); err != nil {
		return err
	}
	return q(`SELECT id,proc_spell,COALESCE(proc_chance,0) FROM item WHERE proc_spell<>0`, func(r *sql.Rows) error {
		var id string
		var sp, ch int
		if err := r.Scan(&id, &sp, &ch); err != nil {
			return err
		}
		if it := c.Items[id]; it != nil {
			it.ProcSpell, it.ProcChance = sp, ch
		}
		return nil
	})
}

// CastableSpells is a template's spells the server can cast, in order.
func (c *Catalog) CastableSpells(template int) []*SpellDef {
	var out []*SpellDef
	for _, id := range c.TemplateSpells[template] {
		if s := c.Spells[id]; s != nil && rules.ImplementedSpell(s.Type) {
			out = append(out, s)
		}
	}
	return out
}
