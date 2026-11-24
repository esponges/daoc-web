package world

import (
	"path/filepath"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// Every class a player may create has a start, a kit it can wear and trainers
// that teach it, on the imported data. Skipped without a world store.
func TestOfferedClassesOnRealData(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	for _, a := range Archetypes {
		if _, ok := cat.StartFor(a.Realm, a.Race, a.Class); !ok {
			t.Errorf("%s: the data has no start for race %d class %d", a.Label, a.Race, a.Class)
		}
		if len(cat.StartEquip[a.Class]) == 0 {
			t.Errorf("%s: no starting equipment", a.Label)
		}
		if len(cat.TrainableSpecs(a.Class)) == 0 {
			t.Errorf("%s: no spec lines to train", a.Label)
		}
	}
	for _, id := range []int{rules.ClassBerserker, rules.ClassSavage, rules.ClassHealer, rules.ClassShaman, rules.ClassRunemaster, rules.ClassThane, rules.ClassSkald, rules.ClassHunter, rules.ClassShadowblade} {
		st, _ := store.Open(":memory:")
		w := New(Config{Cat: cat, Store: st, Log: t.Logf})
		acct, _ := st.AccountFor("d:k", "k", false)
		c := &store.Character{AccountID: acct.ID, Name: "Kit", Race: rules.Classes[id].Races[0], Class: id, Level: 1, Zone: 100, XPScheme: store.XPSchemeRules}
		st.CreateCharacter(c)
		if err := w.GiveStarterKit(c); err != nil {
			t.Fatal(err)
		}
		worn := 0
		for s := range slotsOf(t, st, c.ID) {
			if s >= PackSize {
				worn++
			}
		}
		t.Logf("%s: %d worn of %d items; trainable %v", rules.Classes[id].Name, worn, len(slotsOf(t, st, c.ID)), cat.TrainableSpecs(id))
		if worn == 0 {
			t.Errorf("%s wears nothing from the kit", rules.Classes[id].Name)
		}
		st.Close()
	}
}

// A level 10 character of each new class has spec points by the class's
// multiplier, trains at the trainer of the class or its base class, and a
// melee class gains styles from what it trains. On the imported data.
func TestNewClassesTrainAndGainStyles(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	for _, c := range []struct {
		class       int
		trainerType string
		spec        string
		styles      bool
	}{
		{rules.ClassBerserker, "DOL.GS.Trainer.VikingTrainer", "Sword", true},
		{rules.ClassSavage, "DOL.GS.Trainer.SavageTrainer", "Hand to Hand", true},
		{rules.ClassHealer, "DOL.GS.Trainer.SeerTrainer", "Mending", false},
		{rules.ClassShaman, "DOL.GS.Trainer.ShamanTrainer", "Subterranean", false},
		{rules.ClassRunemaster, "DOL.GS.Trainer.MysticTrainer", "Runecarving", false},
		{rules.ClassThane, "DOL.GS.Trainer.ThaneTrainer", "Hammer", true},
		{rules.ClassThane, "DOL.GS.Trainer.VikingTrainer", "Stormcalling", false},
		{rules.ClassSkald, "DOL.GS.Trainer.SkaldTrainer", "Sword", true},
		{rules.ClassSkald, "DOL.GS.Trainer.VikingTrainer", "Battlesongs", false},
		{rules.ClassHunter, "DOL.GS.Trainer.HunterTrainer", "Spear", true},
		{rules.ClassShadowblade, "DOL.GS.Trainer.ShadowbladeTrainer", "Critical Strike", true},
		{rules.ClassShadowblade, "DOL.GS.Trainer.RogueTrainer", "Stealth", false},
		{rules.ClassHunter, "DOL.GS.Trainer.RogueTrainer", "Composite Bow", false},
	} {
		name := rules.Classes[c.class].Name
		r := newRig(t, npcAt(1, "Trainer", c.trainerType, 30050, 30000))
		r.w.Cat = cat
		p := r.player("Pupil", 10, 30000, 30000, false)
		p.Char.Class = c.class
		r.w.recalc(p)
		if p.sh.unspent <= 0 {
			t.Errorf("%s has %d spec points at level 10", name, p.sh.unspent)
		}
		before := len(p.bar)
		r.z.Train(p, c.spec, 2)
		if errs := errsOf(p); len(errs) > 0 {
			t.Errorf("%s: %v", name, errs)
		}
		if p.specLevel(c.spec) != 3 {
			t.Errorf("%s: %s is %d after training, want 3", name, c.spec, p.specLevel(c.spec))
		}
		if c.styles {
			r.z.computeBar(p)
			// Styles need the weapon of the line; count what the class could use.
			n := len(p.bar) + len(p.barAlso)
			if n <= before {
				t.Errorf("%s: training %s opened no styles (%d before, %d after)", name, c.spec, before, n)
			}
			t.Logf("%s: %d styles ready or held back after training %s", name, n, c.spec)
		}
	}
}

// A caster has a power pool that fills at first, drains and regenerates; a
// Warrior has none, and the snapshot carries power only for the caster.
func TestPlayerPower(t *testing.T) {
	r := newRig(t)
	w := r.player("Warrior", 5, 30000, 30000, true)
	h := r.player("Healer", 5, 30500, 30000, true)
	h.Char.Class = rules.ClassHealer
	h.powerSet = false // as for a character who has never had a pool
	r.w.recalc(h)
	if w.sh.maxPower != 0 || w.power != 0 {
		t.Errorf("a Warrior has power %v of %d", w.power, w.sh.maxPower)
	}
	if h.sh.maxPower <= 0 || h.power != float64(h.sh.maxPower) {
		t.Fatalf("a Healer has %v of %d power, want a full pool", h.power, h.sh.maxPower)
	}
	if s := r.z.selfOf(w); s.MaxPower != 0 || s.Power != 0 {
		t.Errorf("a Warrior's snapshot carries power %+v", s)
	}
	full := h.sh.maxPower
	h.power = 0
	r.ticks(int(rules.PowerRegenPeriod*float64(TickRate))*3 + 5)
	if h.power <= 0 || h.power > float64(full) {
		t.Errorf("after regeneration the Healer has %v of %d", h.power, full)
	}
	if s := r.z.selfOf(h); s.MaxPower != full || s.Power != int(h.power) {
		t.Errorf("the Healer's snapshot says %d of %d, want %d of %d", s.Power, s.MaxPower, int(h.power), full)
	}
}

// A Berserker and a Savage in the starting kit kill a level 1 creature, and a
// trained style is used on the way. On the imported data.
func TestMeleeClassesFightWithStyles(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	for _, c := range []struct {
		class int
		spec  string
	}{{rules.ClassBerserker, "Axe"}, {rules.ClassSavage, "Axe"}} {
		name := rules.Classes[c.class].Name
		r := newRig(t, wolf(1, 30100, 30000, 1))
		r.w.Cat = cat
		n := r.z.npcs[0]
		n.Aggro, n.Roam = 0, 0
		p := r.player("Fighter", 10, 30000, 30000, false)
		p.Char.Class = c.class
		p.Char.Specs = map[string]int{c.spec: 4}
		if err := r.w.GiveStarterKit(p.Char); err != nil {
			t.Fatal(err)
		}
		r.w.loadInventory(p)
		r.w.recalc(p)
		r.z.buildBar(p)
		if p.weapon == nil {
			t.Fatalf("%s wields nothing from the kit", name)
		}
		if len(p.bar) == 0 {
			t.Fatalf("%s has no styles on the bar after training %s to 4", name, c.spec)
		}
		p.Body.Yaw = angleTo(p.Body.X, p.Body.Y, n.Body.X, n.Body.Y)
		r.z.SetTarget(p, n.ID)
		r.z.SetAttack(p, true)
		styled := false
		for i := 0; i < 40*TickRate && !n.dead(); i++ {
			if ok, _ := r.z.styleReady(p, &p.bar[0]); ok && p.pending < 0 {
				r.z.RequestStyle(p, 0)
			}
			r.z.Tick()
			styled = styled || p.lastStyle != 0
		}
		if !n.dead() {
			t.Errorf("%s did not kill a level 1 creature in 40 seconds (it has %d of %d)", name, n.HP, n.MaxHP)
		}
		if !styled {
			t.Errorf("%s never landed a style", name)
		}
	}
}

// A svartalf chanter of the imported data casts Gleam Blast at a Healer who
// is fought.
func TestRealCasterCreatureHurts(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	sp := Spawn{ID: 1, Name: "svartalf chanter", Model: 622, X: 30800, Y: 30000, Z: 5000, Level: 14, Size: 50, Speed: 191, Template: 20008,
		RespawnS: 30, AggroLevel: 1, AggroRange: 800}
	r := newRig(t, sp)
	r.w.Cat = cat
	n := r.z.npcs[0]
	n.learnSpells(cat)
	if len(n.spells) == 0 {
		t.Fatal("the chanter knows no spell the server can cast")
	}
	p := r.player("Healer", 10, 30000, 30000, false)
	p.Char.Class = rules.ClassHealer
	p.powerSet = false
	r.w.recalc(p)
	hp := p.hp
	n.engage(p, r.z)
	var blast bool
	for i := 0; i < 20*TickRate && !blast; i++ {
		r.z.Tick()
		for _, s := range snaps(p) {
			for _, e := range s.Ev {
				blast = blast || (e.K == "spell" && e.Text == "Gleam Blast")
			}
		}
	}
	if !blast {
		t.Fatalf("no Gleam Blast in 20 seconds; the chanter has %v", n.spells)
	}
	if p.hp >= hp {
		t.Errorf("the Healer lost no health: %v of %v", p.hp, hp)
	}
	t.Logf("the chanter's spells: %d; the Healer is at %.0f of %.0f", len(n.spells), p.hp, hp)
}

// A Thane and a Skald are taught by their own trainers and the Viking's, not by
// a Warrior's or each other's, and a Thane at Stormcalling 5 holds the six
// spells of the data's first five levels. On the imported data.
func TestThaneAndSkaldTrainersAndBooks(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	r := newRig(t, npcAt(1, "Thane trainer", "DOL.GS.Trainer.ThaneTrainer", 30050, 30000),
		npcAt(2, "Skald trainer", "DOL.GS.Trainer.SkaldTrainer", 30000, 30050),
		npcAt(3, "Warrior trainer", "DOL.GS.Trainer.WarriorTrainer", 29950, 30000))
	r.w.Cat = cat
	for _, c := range []struct {
		class         int
		mine, notMine int64
		spec, notLine string
	}{{rules.ClassThane, 1, 2, "Stormcalling", "Battlesongs"}, {rules.ClassSkald, 2, 1, "Battlesongs", "Stormcalling"}} {
		name := rules.Classes[c.class].Name
		p := r.player(name, 10, 30000, 30000, false)
		p.Char.Class = c.class
		r.w.recalc(p)
		if !contains(cat.TrainableSpecs(c.class), c.spec) || contains(cat.TrainableSpecs(c.class), c.notLine) {
			t.Errorf("%s trains %v, want %s and not %s", name, cat.TrainableSpecs(c.class), c.spec, c.notLine)
		}
		r.z.openTrainer(p, r.z.byID[c.mine])
		if got := sentOf[protocol.Trainer](p, protocol.TTrainer); len(got) != 1 || len(got[0].Specs) == 0 {
			t.Errorf("%s: their own trainer sent %v", name, got)
		}
		for _, other := range []int64{c.notMine, 3} {
			r.z.openTrainer(p, r.z.byID[other])
			if errs := errsOf(p); len(errs) == 0 {
				t.Errorf("%s was taught by trainer %d", name, other)
			}
			snaps(p)
		}
	}
	// A Warrior is not taught by the Skald's trainer.
	w := r.player("Warrior", 10, 30000, 30000, false)
	r.z.openTrainer(w, r.z.byID[2])
	if len(errsOf(w)) == 0 {
		t.Error("a Warrior was taught by a Skald trainer")
	}

	thane := r.player("Bookish", 5, 30000, 30000, false)
	thane.Char.Class = rules.ClassThane
	thane.Char.Specs = map[string]int{"Stormcalling": 5}
	r.w.recalc(thane)
	r.w.refreshBook(thane)
	var names []string
	for _, b := range thane.book {
		names = append(names, b.Spell.Name)
	}
	want := []string{"Thor's Minor Bolt", "Thor's Vigor", "Thor's Lesser Bolt", "Thunder's Bash", "Thunder Shout", "Toothgrinder's Hoof"}
	if len(names) != len(want) {
		t.Fatalf("book %v, want %v", names, want)
	}
	for _, n := range want {
		if !contains(names, n) {
			t.Errorf("book %v lacks %s", names, n)
		}
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// A Savage's Hand to Hand styles need a claw in each hand: with the kit's axe
// and shield they are held back and say so; with two claws they are on the bar.
// On the imported data.
func TestSavageHandToHandStylesNeedTwoClaws(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	r := newRig(t)
	r.w.Cat = cat
	p := r.player("Savage", 10, 30000, 30000, false)
	p.Char.Class = rules.ClassSavage
	p.Char.Specs = map[string]int{"Hand to Hand": 4}
	if err := r.w.GiveStarterKit(p.Char); err != nil {
		t.Fatal(err)
	}
	r.w.loadInventory(p)
	r.w.recalc(p)
	r.z.computeBar(p)
	h2h := func(list []StyleDef) int {
		n := 0
		for _, s := range list {
			if s.Spec == "Hand to Hand" {
				n++
			}
		}
		return n
	}
	if h2h(p.bar) != 0 || h2h(p.barAlso) == 0 {
		t.Fatalf("with the kit: %d Hand to Hand styles usable, %d held back", h2h(p.bar), h2h(p.barAlso))
	}
	for _, s := range p.barAlso {
		if s.Spec == "Hand to Hand" {
			if nd := rules.StyleWeaponNeeds(s.WeaponReq, s.Spec); nd != "a hand to hand weapon in each hand" {
				t.Errorf("%s needs %q", s.Name, nd)
			}
		}
	}
	p.inv[SlotWeapon] = &invItem{ID: "bronze_moon_fang", Slot: SlotWeapon, Count: 1, Cond: 10000}
	p.inv[SlotShield] = &invItem{ID: "bronze_moon_fang", Slot: SlotShield, Count: 1, Cond: 10000}
	r.w.recalc(p)
	r.z.computeBar(p)
	if p.left == nil || h2h(p.bar) == 0 || h2h(p.barAlso) != 0 {
		t.Errorf("with two claws: left %v, %d Hand to Hand styles usable, %d held back", p.left != nil, h2h(p.bar), h2h(p.barAlso))
	}
}
