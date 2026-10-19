package world

import (
	"database/sql"
	"log"
	"strings"

	"daocweb/internal/server/rules"
)

// loadRules reads what the rule book needs from the store: item bonuses, race
// resists, each class's specs and the abilities they grant. A store imported
// before these were kept still loads, with none of them and a warning.
func (c *Catalog) loadRules(q func(string, func(*sql.Rows) error) error) error {
	c.Resists = map[int]map[int]float64{}
	c.ClassSpecs = map[int][]string{}
	c.SpecNames = map[string]string{}
	c.Grants = map[int][]rules.Grant{}
	if err := q(`SELECT class_id,spec FROM class_spec ORDER BY class_id,level,spec`, func(r *sql.Rows) error {
		var cl int
		var sp string
		if err := r.Scan(&cl, &sp); err != nil {
			return err
		}
		c.ClassSpecs[cl] = append(c.ClassSpecs[cl], sp)
		return nil
	}); err != nil {
		return err
	}
	if err := q(`SELECT key,name FROM specialization`, func(r *sql.Rows) error {
		var k, n string
		if err := r.Scan(&k, &n); err != nil {
			return err
		}
		c.SpecNames[k] = n
		return nil
	}); err != nil {
		return err
	}
	if !c.hasRules {
		log.Printf("world store has no item bonuses, races or abilities; characters know no abilities (run cmd/dbimport again)")
		return nil
	}
	if err := q(`SELECT item,type,amount FROM item_bonus`, func(r *sql.Rows) error {
		var id string
		var b ItemBonus
		if err := r.Scan(&id, &b.Type, &b.Amount); err != nil {
			return err
		}
		if it := c.Items[id]; it != nil {
			it.Bonuses = append(it.Bonuses, b)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := q(`SELECT id,COALESCE(crush,0),COALESCE(slash,0),COALESCE(thrust,0) FROM race`, func(r *sql.Rows) error {
		var id int
		var cr, sl, th float64
		if err := r.Scan(&id, &cr, &sl, &th); err != nil {
			return err
		}
		if cr != 0 || sl != 0 || th != 0 {
			c.Resists[id] = map[int]float64{rules.DamageCrush: cr, rules.DamageSlash: sl, rules.DamageThrus: th}
		}
		return nil
	}); err != nil {
		return err
	}
	type grantRow struct {
		g     rules.Grant
		class int
	}
	var grants []grantRow
	if err := q(`SELECT spec,spec_level,ability,ability_level,class_id FROM spec_ability`, func(r *sql.Rows) error {
		var g grantRow
		if err := r.Scan(&g.g.Source, &g.g.Level, &g.g.Ability, &g.g.AbilityLevel, &g.class); err != nil {
			return err
		}
		grants = append(grants, g)
		return nil
	}); err != nil {
		return err
	}
	for cl, specs := range c.ClassSpecs {
		has := map[string]bool{}
		for _, s := range specs {
			has[s] = true
		}
		for _, g := range grants {
			if has[g.g.Source] && (g.class == 0 || g.class == cl) {
				c.Grants[cl] = append(c.Grants[cl], g.g)
			}
		}
	}
	return nil
}

// IsCareer says whether a class_spec key is a career (abilities by character
// level) rather than a spec that is trained.
func IsCareer(key string) bool { return strings.HasSuffix(key, "Career") }

// Careers and TrainableSpecs split a class's keys.
func (c *Catalog) Careers(class int) map[string]bool {
	out := map[string]bool{}
	for _, k := range c.ClassSpecs[class] {
		if IsCareer(k) {
			out[k] = true
		}
	}
	return out
}

// TrainableSpecs are the specs a class trains, in the store's order.
func (c *Catalog) TrainableSpecs(class int) []string {
	var out []string
	for _, k := range c.ClassSpecs[class] {
		if !IsCareer(k) {
			out = append(out, k)
		}
	}
	return out
}

// SpecName is a spec's display name.
func (c *Catalog) SpecName(key string) string {
	if n := c.SpecNames[key]; n != "" {
		return n
	}
	return key
}

// HasRules reports whether the store carries what the rule book reads.
func (c *Catalog) HasRules() bool { return c.hasRules }
