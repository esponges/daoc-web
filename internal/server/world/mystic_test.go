package world

import (
	"math/rand"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"daocweb/internal/server/rules"
)

// Every spell the Spiritmaster's and Bonedancer's lines give up to level 50 can
// be cast, on the imported data. A spell the server cannot cast would show in
// the book as not working, which is the thing this change is to end.
func TestMysticPetClassesCanCastEverySpellTheirLinesGive(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	for _, class := range []int{rules.ClassSpiritmaster, rules.ClassBonedancer} {
		name := rules.Classes[class].Name
		specs := cat.TrainableSpecs(class)
		if len(specs) == 0 {
			t.Fatalf("%s trains nothing", name)
		}
		total := 0
		var bad []string
		for _, spec := range specs {
			for _, l := range cat.SpellsFor(class, spec) {
				sp := cat.Spells[l.Spell]
				if sp == nil || l.Level > 50 {
					continue
				}
				total++
				if !spellSupported(sp) {
					bad = append(bad, spec+" "+strconv.Itoa(l.Level)+" "+sp.Name+" ("+sp.Type+", "+sp.Target+")")
				}
			}
		}
		sort.Strings(bad)
		t.Logf("%s: %d spells to level 50 across %v, %d not castable", name, total, specs, len(bad))
		for _, b := range bad {
			t.Errorf("%s cannot cast %s", name, b)
		}
		if total == 0 {
			t.Errorf("%s has no spells in the data", name)
		}
	}
}

// realCaster is a level 50 player of a class on the imported data, with every
// spec of theirs at 50 and a full pool, and a wolf close by to cast at.
func realCaster(t *testing.T, cat *Catalog, class int) (*rig, *Player, *NPC) {
	r := newRig(t, wolf(1, 30400, 30000, 8))
	r.w.Cat = cat
	r.z.rng = rand.New(rand.NewSource(5))
	n := r.z.npcs[0]
	n.Aggro, n.Roam = 0, 0
	n.setLevel(40)
	n.HP = n.MaxHP
	p := caster(r, "Mystic", class, 50)
	for _, spec := range cat.TrainableSpecs(class) {
		p.Char.Specs[spec] = 50
	}
	r.w.recalc(p)
	r.w.refreshBook(p)
	p.power = 1e6
	return r, p, n
}

// bestOfType is the highest spell of a type in the class's book, or nil.
func bestOfType(p *Player, typ string) *SpellDef {
	var best *SpellDef
	bestLevel := -1
	for _, b := range p.book {
		if b.Spell.Type == typ && b.Level > bestLevel {
			best, bestLevel = b.Spell, b.Level
		}
	}
	return best
}

// With the real spell data, each class casts a spell of every kind its lines
// are made of and the server does what the spell says: a Spiritmaster summons a
// spirit, debuffs, transfers life and turns a blade; a Bonedancer summons a
// commander and its minions, damages and slows, and wards.
func TestMysticClassesCastTheirKindsOfSpellOnRealData(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	cast := func(r *rig, p *Player, sp *SpellDef, target combatant) {
		r.z.Cast(p, sp.ID, target.entityID())
		r.ticks(int(sp.CastTime*TickRate) + 10)
	}
	type check func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC)
	pet := func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC) {
		if p.pet == nil || !p.pet.isPet() || p.pet.Level < 1 {
			t.Fatalf("%s left no pet", sp.Name)
		}
		t.Logf("%s summoned %q, level %d", sp.Name, p.pet.Name, p.pet.Level)
	}
	debuffed := func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC) {
		if len(wolf.effects) == 0 {
			t.Errorf("%s left nothing on the creature", sp.Name)
		}
	}
	hurt := func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC) {
		if wolf.HP >= wolf.MaxHP {
			t.Errorf("%s did no damage", sp.Name)
		}
	}
	for _, c := range []struct {
		class int
		kinds []struct {
			typ string
			chk check
		}
	}{
		{rules.ClassSpiritmaster, []struct {
			typ string
			chk check
		}{
			{rules.SpellSummonSpirit, pet},
			{rules.SpellDexQuiDebuff, debuffed},
			{rules.SpellBodyResistDebuff, debuffed},
			{rules.SpellMesmerize, debuffed},
			{rules.SpellLifedrain, hurt},
			{rules.SpellBladeturn, func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC) {
				if !hasKind(p, effBlade) {
					t.Errorf("%s left no bladeturn on its target", sp.Name)
				}
			}},
			{rules.SpellLifeTransfer, func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC) {
				friend := r.player("Friend", 50, 30200, 30000, true)
				friend.hp = 10
				hp := p.hp
				cast(r, p, sp, friend)
				if friend.hp <= 10 || p.hp >= hp {
					t.Errorf("%s: the friend has %v hp (was 10), the caster %v (was %v)", sp.Name, friend.hp, p.hp, hp)
				}
			}},
		}},
		{rules.ClassBonedancer, []struct {
			typ string
			chk check
		}{
			{rules.SpellSummonCommander, pet},
			{rules.SpellDamageSpeedDecr, func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC) {
				hurt(t, r, p, sp, wolf)
				debuffed(t, r, p, sp, wolf)
			}},
			{rules.SpellLifedrain, hurt},
			{rules.SpellBladeturn, func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC) {
				if !hasKind(p, effBlade) {
					t.Errorf("%s left no bladeturn on its target", sp.Name)
				}
			}},
			{rules.SpellAblativeArmor, func(t *testing.T, r *rig, p *Player, sp *SpellDef, wolf *NPC) {
				// A Bonedancer's ward is for the pet: with none, the cast was refused.
				if p.pet != nil {
					t.Fatalf("%s was cast with no pet and made one", sp.Name)
				}
				cast(r, p, bestOfType(p, rules.SpellSummonCommander), p)
				cast(r, p, sp, p)
				if p.pet == nil || !hasKind(p.pet, effAblative) {
					t.Errorf("%s left no ablative on the commander", sp.Name)
				}
			}},
		}},
	} {
		name := rules.Classes[c.class].Name
		for _, k := range c.kinds {
			t.Run(name+" "+k.typ, func(t *testing.T) {
				r, p, wolf := realCaster(t, cat, c.class)
				sp := bestOfType(p, k.typ)
				if sp == nil {
					t.Fatalf("%s has no %s spell in the data", name, k.typ)
				}
				target := combatant(p)
				if sp.Target == "Enemy" {
					target = wolf
				}
				if k.typ != rules.SpellLifeTransfer {
					cast(r, p, sp, target)
				}
				k.chk(t, r, p, sp, wolf)
			})
		}
	}
	t.Run("a commander takes minions", func(t *testing.T) {
		r, p, _ := realCaster(t, cat, rules.ClassBonedancer)
		cmd, minion := bestOfType(p, rules.SpellSummonCommander), bestOfType(p, rules.SpellSummonMinion)
		if cmd == nil || minion == nil {
			t.Fatalf("a Bonedancer's lines give a commander %v and a minion %v", cmd != nil, minion != nil)
		}
		cast(r, p, cmd, p)
		if p.pet == nil {
			t.Fatal("no commander")
		}
		cast(r, p, minion, p)
		if len(p.pet.pet.minions) == 0 {
			t.Errorf("%s gave the commander no minion", minion.Name)
		}
	})
}

// Every spell on the pet that either class's lines give lands on a summoned
// pet: a lasting effect on it, or a heal when it was hurt.
func TestEveryPetSpellInTheBooksLandsOnThePet(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	for _, c := range []struct {
		class  int
		summon string
	}{{rules.ClassSpiritmaster, rules.SpellSummonSpirit}, {rules.ClassBonedancer, rules.SpellSummonCommander}} {
		name := rules.Classes[c.class].Name
		r, p, _ := realCaster(t, cat, c.class)
		summon := bestOfType(p, c.summon)
		r.z.Cast(p, summon.ID, p.ID)
		r.ticks(int(summon.CastTime*TickRate) + 10)
		if p.pet == nil {
			t.Fatalf("%s has no pet", name)
		}
		landed := 0
		for _, b := range p.book {
			sp := b.Spell
			if sp.Target != "Pet" {
				continue
			}
			p.power = 1e6
			p.spellReady = map[int]float64{}
			p.pet.HP = p.pet.MaxHP / 2
			p.pet.effects = nil
			hp, fx := p.pet.HP, 0
			r.z.Cast(p, sp.ID, 0)
			if p.casting == nil {
				t.Errorf("%s: %s (%s) was refused: %q", name, sp.Name, sp.Type, lastErr(t, p))
				continue
			}
			r.ticks(int(sp.CastTime*TickRate) + 10)
			fx = len(p.pet.effects)
			if fx == 0 && p.pet.HP <= hp {
				t.Errorf("%s: %s (%s) did nothing to the pet", name, sp.Name, sp.Type)
				continue
			}
			landed++
		}
		t.Logf("%s: %d pet spells landed", name, landed)
		if landed == 0 {
			t.Errorf("%s has no spell on the pet in the data", name)
		}
	}
}
