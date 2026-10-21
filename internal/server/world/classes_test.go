package world

import (
	"path/filepath"
	"testing"

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
	for _, id := range []int{rules.ClassBerserker, rules.ClassSavage, rules.ClassHealer, rules.ClassShaman, rules.ClassRunemaster} {
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
