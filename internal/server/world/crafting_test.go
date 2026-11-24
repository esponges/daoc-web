package world

import (
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// craftRig is a zone with a weaponcrafting master and a catalogue with two
// recipes: a dagger from three bars (level 0) and a hard sword (level 100).
func craftRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t, Spawn{ID: 970, Name: "Aase", Model: 56, X: 30000, Y: 30000, Z: 5000, Level: 30, Size: 50, Speed: 191, Realm: 2,
		ClassType: "DOL.GS.WeaponCraftingMaster"})
	c := r.w.Cat
	c.Items["bronze_bar"] = &Item{ID: "bronze_bar", Name: "bronze bar", MaxCount: 200, Price: 3}
	c.Items["bronze_dagger"] = &Item{ID: "bronze_dagger", Name: "bronze dagger", Level: 1, Realm: 2}
	c.Items["steel_sword"] = &Item{ID: "steel_sword", Name: "steel sword", Level: 30, Realm: 2}
	dagger := &Recipe{Product: "bronze_dagger", Skill: rules.SkillWeapon, Level: 0, Mats: []Material{{"bronze_bar", 3}}}
	sword := &Recipe{Product: "steel_sword", Skill: rules.SkillWeapon, Level: 100, Mats: []Material{{"bronze_bar", 10}}}
	c.Recipes = map[string]*Recipe{dagger.Product: dagger, sword.Product: sword}
	c.RecipesBySkill = map[int][]*Recipe{rules.SkillWeapon: {dagger, sword}}
	return r
}

func crafter(t *testing.T, r *rig, name string) *Player {
	p := r.player(name, 5, 30050, 30000, false)
	r.w.loadCraft(p)
	drain(p)
	return p
}

func lastCraft(p *Player) (protocol.Craft, bool) {
	all := sentOf[protocol.Craft](p, protocol.TCraft)
	if len(all) == 0 {
		return protocol.Craft{}, false
	}
	return all[len(all)-1], true
}

func TestJoinAtAMasterAndKeepIt(t *testing.T) {
	r := craftRig(t)
	p := crafter(t, r, "Smithy")
	r.z.Craft(p, &protocol.CraftOp{Op: "make", ID: "bronze_dagger"})
	if p.craft.making != nil {
		t.Fatal("crafted before joining")
	}
	r.z.Interact(p, 970)
	w, ok := lastCraft(p)
	if !ok || w.Offer != rules.SkillWeapon || w.Prof != 0 {
		t.Fatalf("a master offers a craft to one who has none: %+v", w)
	}
	r.z.Craft(p, &protocol.CraftOp{Op: "join"})
	if p.craft.primary != rules.SkillWeapon || p.craft.skills[rules.SkillWeapon] != rules.StartSkill ||
		p.craft.skills[rules.SkillMetal] != rules.StartSkill || p.craft.skills[rules.SkillWood] != rules.StartSkill {
		t.Fatalf("joined with %+v", p.craft)
	}
	// It was saved: a fresh load sees it.
	again := &Player{Char: p.Char}
	r.w.loadCraft(again)
	if again.craft.primary != rules.SkillWeapon || len(again.craft.skills) != 3 {
		t.Errorf("not saved: %+v", again.craft)
	}
	w, _ = lastCraft(p)
	if w.Prof != rules.SkillWeapon || len(w.Recipes) != 1 || w.Recipes[0].ID != "bronze_dagger" {
		t.Errorf("only the dagger is within reach at skill 1: %+v", w.Recipes)
	}
	if w.Recipes[0].Chance != 100 || w.Recipes[0].Secs != 3 || w.Recipes[0].Mats[0].Have != 0 {
		t.Errorf("recipe line: %+v", w.Recipes[0])
	}
	// Away from a master, joining does nothing, and a second craft is refused.
	p.Body.X = 50000
	r.z.Craft(p, &protocol.CraftOp{Op: "join"})
	if p.craft.primary != rules.SkillWeapon {
		t.Error("the craft changed")
	}
}

func TestMakeAnItem(t *testing.T) {
	r := craftRig(t)
	p := crafter(t, r, "Smithy")
	r.z.Interact(p, 970)
	r.z.Craft(p, &protocol.CraftOp{Op: "join"})
	drain(p)
	r.z.Craft(p, &protocol.CraftOp{Op: "make", ID: "bronze_dagger"})
	if p.craft.making != nil {
		t.Fatal("crafted without materials")
	}
	if _, err := r.w.addItem(p, "bronze_bar", 7); err != nil {
		t.Fatal(err)
	}
	drain(p)
	r.z.Craft(p, &protocol.CraftOp{Op: "make", ID: "steel_sword"})
	if p.craft.making != nil {
		t.Fatal("made something far beyond the skill")
	}
	r.z.Craft(p, &protocol.CraftOp{Op: "make", ID: "bronze_dagger"})
	if p.craft.making == nil {
		t.Fatal("did not start")
	}
	r.ticks(TickRate * 2)
	if p.craft.making == nil {
		t.Fatal("finished before the time was up")
	}
	r.ticks(TickRate * 2)
	if p.craft.making != nil {
		t.Fatal("not finished")
	}
	var made *invItem
	bars := 0
	for _, it := range p.inv {
		if it.ID == "bronze_dagger" {
			made = it
		}
		if it.ID == "bronze_bar" {
			bars += it.Count
		}
	}
	if made == nil || made.Crafter != "Smithy" || made.Quality < 94 || made.Quality > 100 {
		t.Fatalf("made %+v", made)
	}
	if bars != 4 {
		t.Errorf("%d bars left, want 4", bars)
	}
	if p.craft.skills[rules.SkillWeapon] != 2 {
		t.Errorf("a level 0 recipe at skill 1 always teaches: %d", p.craft.skills[rules.SkillWeapon])
	}
	// The item and the skill are in the store.
	items, _ := r.st.Items(p.Char.ID)
	found := false
	for _, it := range items {
		if it.ItemID == "bronze_dagger" && it.Crafter == "Smithy" && it.Quality == made.Quality {
			found = true
		}
	}
	if !found {
		t.Error("the crafted item was not saved with its maker")
	}
	c, _ := r.st.LoadCraft(p.Char.ID)
	if c.Skills[rules.SkillWeapon] != 2 {
		t.Errorf("saved skill %d", c.Skills[rules.SkillWeapon])
	}
	// What was made carries its quality into the swing.
	p.inv[SlotWeapon] = &invItem{ID: "bronze_dagger", Quality: 97}
	if q := p.qualOf(SlotWeapon, 100); q != 0.97 {
		t.Errorf("quality in combat %v", q)
	}
	if q := p.qualOf(SlotTorso, 80); q != 0.8 {
		t.Errorf("a found item keeps its template quality: %v", q)
	}
}

func TestMovingOrFightingStopsACraft(t *testing.T) {
	r := craftRig(t)
	p := crafter(t, r, "Smithy")
	r.z.Interact(p, 970)
	r.z.Craft(p, &protocol.CraftOp{Op: "join"})
	r.w.addItem(p, "bronze_bar", 3)
	r.z.Craft(p, &protocol.CraftOp{Op: "make", ID: "bronze_dagger"})
	r.ticks(TickRate)
	p.Body.X += 40
	drain(p)
	r.ticks(2)
	done := sentOf[protocol.CraftDone](p, protocol.TCraftDone)
	if p.craft.making != nil || len(done) != 1 || done[0].OK {
		t.Fatalf("moving: %+v", done)
	}
	if p.have("bronze_bar") != 3 {
		t.Errorf("a cancelled craft used materials: %d left", p.have("bronze_bar"))
	}
	r.z.Craft(p, &protocol.CraftOp{Op: "make", ID: "bronze_dagger"})
	p.combatUntil = r.z.now + 10
	r.ticks(2)
	if p.craft.making != nil {
		t.Error("a fight did not stop the craft")
	}
}

func TestAFailureLosesHalf(t *testing.T) {
	r := craftRig(t)
	p := crafter(t, r, "Smithy")
	r.z.Interact(p, 970)
	r.z.Craft(p, &protocol.CraftOp{Op: "join"})
	p.craft.skills[rules.SkillWeapon] = 50 // the sword (level 100) works half the time
	r.w.addItem(p, "bronze_bar", 200)
	failed, made := 0, 0
	for i := 0; i < 40 && (failed == 0 || made == 0); i++ {
		before := p.have("bronze_bar")
		drain(p)
		r.z.Craft(p, &protocol.CraftOp{Op: "make", ID: "steel_sword"})
		r.ticks(TickRate * 12)
		used := before - p.have("bronze_bar")
		done := sentOf[protocol.CraftDone](p, protocol.TCraftDone)
		if len(done) == 0 {
			t.Fatal("no result")
		}
		if done[len(done)-1].OK {
			made++
			if used != 10 {
				t.Fatalf("a success used %d", used)
			}
		} else {
			failed++
			if used != 5 {
				t.Fatalf("a failure used %d, want half of 10", used)
			}
		}
	}
	if failed == 0 || made == 0 {
		t.Errorf("at 50%% in 40 tries: %d made, %d failed", made, failed)
	}
}

func TestShopStacksMaterials(t *testing.T) {
	r := craftRig(t)
	p := crafter(t, r, "Buyer")
	r.w.addItem(p, "bronze_bar", 5)
	slot := stackFor(p, r.w.Cat.Items["bronze_bar"])
	if slot < 0 {
		t.Fatal("no stack found to top up")
	}
	p.inv[slot].Count = 200
	if stackFor(p, r.w.Cat.Items["bronze_bar"]) >= 0 {
		t.Error("a full stack was topped up")
	}
	p.inv[slot].Count, p.inv[slot].Crafter = 5, "Someone"
	if stackFor(p, r.w.Cat.Items["bronze_bar"]) >= 0 {
		t.Error("a crafted item was stacked")
	}
}

func TestJoinMeansTheMasterYouTalkedTo(t *testing.T) {
	// A weaponcrafter and a tailor stand side by side, both in reach.
	r := newRig(t,
		Spawn{ID: 970, Name: "Aase", Model: 56, X: 30000, Y: 30000, Z: 5000, Level: 30, Size: 50, Speed: 191, Realm: 2, ClassType: "DOL.GS.WeaponCraftingMaster"},
		Spawn{ID: 971, Name: "Eskil", Model: 56, X: 30040, Y: 30000, Z: 5000, Level: 30, Size: 50, Speed: 191, Realm: 2, ClassType: "DOL.GS.TailoringMaster"})
	p := crafter(t, r, "Tailor")
	r.z.Interact(p, 971)
	r.z.Craft(p, &protocol.CraftOp{Op: "join"})
	if p.craft.primary != rules.SkillTailor {
		t.Errorf("joined %d, want the tailor's craft", p.craft.primary)
	}
}
