package world

import (
	"database/sql"
	"sort"
)

// Recipe is something a tradeskill can make: the product, the skill that
// makes it, the level it asks of that skill and the materials it uses up.
type Recipe struct {
	Product string
	Skill   int
	Level   int
	Mats    []Material
}

// Material is one ingredient of a recipe.
type Material struct {
	Item  string
	Count int
}

// MatCount is how many materials a recipe uses in all.
func (r *Recipe) MatCount() int {
	n := 0
	for _, m := range r.Mats {
		n += m.Count
	}
	return n
}

// loadRecipes reads the recipes of Midgard (and of no realm): what its
// products are for. A store imported before crafting has none.
func (c *Catalog) loadRecipes(db *sql.DB, q func(string, func(*sql.Rows) error) error) error {
	c.Recipes = map[string]*Recipe{}
	c.RecipesBySkill = map[int][]*Recipe{}
	var has int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('recipe','recipe_material')`).Scan(&has)
	if has != 2 {
		return nil
	}
	if err := q(`SELECT r.product,r.skill,r.level FROM recipe r JOIN item i ON i.id=r.product WHERE i.realm IN (0,2)`, func(r *sql.Rows) error {
		rc := &Recipe{}
		if err := r.Scan(&rc.Product, &rc.Skill, &rc.Level); err != nil {
			return err
		}
		c.Recipes[rc.Product] = rc
		return nil
	}); err != nil {
		return err
	}
	if err := q(`SELECT product,material,count FROM recipe_material ORDER BY rowid`, func(r *sql.Rows) error {
		var p string
		var m Material
		if err := r.Scan(&p, &m.Item, &m.Count); err != nil {
			return err
		}
		if rc := c.Recipes[p]; rc != nil {
			rc.Mats = append(rc.Mats, m)
		}
		return nil
	}); err != nil {
		return err
	}
	for id, rc := range c.Recipes {
		if len(rc.Mats) == 0 {
			delete(c.Recipes, id)
			continue
		}
		c.RecipesBySkill[rc.Skill] = append(c.RecipesBySkill[rc.Skill], rc)
	}
	for _, l := range c.RecipesBySkill {
		sort.Slice(l, func(i, j int) bool {
			if l[i].Level != l[j].Level {
				return l[i].Level < l[j].Level
			}
			return l[i].Product < l[j].Product
		})
	}
	return nil
}
