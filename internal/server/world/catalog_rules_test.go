package world

import (
	"database/sql"
	"path/filepath"
	"testing"

	"daocweb/internal/importer"
	"daocweb/internal/server/rules"
)

// A store imported before the rule book lacks item bonuses, races and
// abilities: it still loads, with none of them.
func TestCatalogLoadsRuleData(t *testing.T) {
	build := func(old bool) *Catalog {
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
			`INSERT INTO item(id,name,level,quality,dps_af,spd_abs,hand,damage_type,object_type,item_type,model,weight,price,realm,droppable,level_req,bonus,condition,max_condition,durability)
			 VALUES('helm','helm',10,90,12,19,0,0,34,21,93,10,100,2,1,0,5,40000,50000,50000)`,
			`INSERT INTO item_bonus VALUES('helm',1,4),('helm',52,3)`,
			`INSERT INTO race(id,name,crush,slash,thrust) VALUES(2000,'Animal',0,5,-15)`,
			`INSERT INTO class_spec VALUES(22,'WarriorCareer',-100),(22,'Shields',-7),(31,'BerserkerCareer',-100)`,
			`INSERT INTO specialization VALUES('Shields','Shield')`,
			`INSERT INTO spec_ability VALUES('WarriorCareer',10,'MidArmor',4,0),('Shields',5,'Guard',1,0),('BerserkerCareer',1,'Evade',1,0)`,
			`INSERT INTO spawn(zone,name,model,x,y,z,heading,level,size,speed,template,roam,respawn_s,aggro_level,aggro_range,realm,flags,race) VALUES(100,'wolf',56,10,10,0,0,1,50,191,0,0,30,0,0,0,0,2000)`,
		}
		if old {
			stmts = []string{
				`DROP TABLE item_bonus`, `DROP TABLE spec_ability`,
				`ALTER TABLE spawn DROP COLUMN race`,
				`ALTER TABLE item DROP COLUMN bonus`, `ALTER TABLE item DROP COLUMN condition`,
				`ALTER TABLE item DROP COLUMN max_condition`, `ALTER TABLE item DROP COLUMN durability`,
				stmts[0],
				`INSERT INTO item(id,name,level,quality,dps_af,spd_abs,hand,damage_type,object_type,item_type,model,weight,price,realm,droppable,level_req) VALUES('helm','helm',10,90,12,19,0,0,34,21,93,10,100,2,1,0)`,
				`INSERT INTO class_spec VALUES(22,'WarriorCareer',-100)`,
				`INSERT INTO spawn(zone,name,model,x,y,z,heading,level,size,speed,template,roam,respawn_s,aggro_level,aggro_range,realm,flags) VALUES(100,'wolf',56,10,10,0,0,1,50,191,0,0,30,0,0,0,0)`,
			}
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

	c := build(false)
	if !c.HasRules() {
		t.Fatal("a current store carries the rule data")
	}
	h := c.Items["helm"]
	if h.Bonus != 5 || h.Condition != 0.8 || h.Absorb != 19 || len(h.Bonuses) != 2 {
		t.Errorf("helm %+v", h)
	}
	if c.Resists[2000][rules.DamageThrus] != -15 {
		t.Errorf("animal resists %v", c.Resists[2000])
	}
	if g := c.Grants[22]; len(g) != 2 {
		t.Errorf("warrior grants %v, want MidArmor and Guard", g)
	}
	if !c.Careers(22)["WarriorCareer"] || len(c.TrainableSpecs(22)) != 1 || c.SpecName("Shields") != "Shield" {
		t.Errorf("warrior careers %v, specs %v", c.Careers(22), c.TrainableSpecs(22))
	}
	sp, err := c.SpawnsIn(100)
	if err != nil || len(sp) != 1 || sp[0].Race != 2000 {
		t.Errorf("spawns %+v, %v", sp, err)
	}

	old := build(true)
	if old.HasRules() {
		t.Error("an old store has no rule data")
	}
	if h := old.Items["helm"]; h == nil || h.Condition != 1 || len(h.Bonuses) != 0 {
		t.Errorf("old helm %+v", h)
	}
	if sp, err := old.SpawnsIn(100); err != nil || len(sp) != 1 || sp[0].Race != 0 {
		t.Errorf("old spawns %+v, %v", sp, err)
	}
}
