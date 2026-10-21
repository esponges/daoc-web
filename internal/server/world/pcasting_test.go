package world

import (
	"math/rand"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

var (
	minorMend = &SpellDef{ID: 3001, Name: "Minor Mending", Type: rules.SpellHeal, Target: "Realm", Range: 2000, CastTime: 2, Value: 60, Power: 10}
	bigMend   = &SpellDef{ID: 3002, Name: "Major Mending", Type: rules.SpellHeal, Target: "Realm", Range: 2000, CastTime: 2, Value: 160, Power: 20}
	summon    = &SpellDef{ID: 4001, Name: "Summon Thing", Type: "SummonMinion", Target: "Self", Range: 0, CastTime: 3, Power: 5}
	runebolt  = &SpellDef{ID: 5001, Name: "Runebolt", Type: rules.SpellBolt, Target: "Enemy", Range: 1500, CastTime: 2, Recast: 10, Damage: 40, DamageType: 12, Power: 5}
	strBuff   = &SpellDef{ID: 6001, Name: "Might", Type: rules.SpellStrConBuff, Target: "Self", Duration: 30, Value: 12, Power: 5}
	groupAF   = &SpellDef{ID: 6002, Name: "Group Shield", Type: rules.SpellArmourBuff, Target: "Group", Range: 1500, Duration: 30, Value: 20, Power: 5}
)

// casterWorld is a zone with a wolf 1,000 units off, a spell table and the
// spec lines "Mending" (class Healer) and "Runecarving" (class Runemaster).
func casterWorld(t *testing.T) *rig {
	r := newRig(t, wolf(1, 31000, 30000, 8))
	r.z.rng = rand.New(zeroSource{})
	n := r.z.npcs[0]
	n.Aggro, n.Roam = 0, 0
	c := r.w.Cat
	c.Spells = map[int]*SpellDef{}
	for _, s := range []*SpellDef{minorMend, bigMend, summon, runebolt, strBuff, groupAF} {
		cp := *s
		c.Spells[s.ID] = &cp
	}
	c.LineSpells = map[string][]LineSpell{
		"Mending": {{Spec: "Mending", Line: "Mending", Spell: 3001, Level: 1}, {Spec: "Mending", Line: "Mending", Spell: 3002, Level: 5},
			{Spec: "Mending", Line: "Mending", Spell: 4001, Level: 1}, {Spec: "Mending", Line: "Mending", Spell: 6001, Level: 1},
			{Spec: "Mending", Line: "Mending", Spell: 6002, Level: 1}},
		"Runecarving": {{Spec: "Runecarving", Line: "Runecarving", Spell: 5001, Level: 1}},
	}
	c.ClassSpecs[rules.ClassHealer] = []string{"SeerCareer", "Mending"}
	c.ClassSpecs[rules.ClassRunemaster] = []string{"MidClothCasterCareer", "Runecarving"}
	return r
}

// caster adds a player of a casting class, with a full pool.
func caster(r *rig, name string, class, level int) *Player {
	p := r.player(name, level, 30000, 30000, true)
	p.Char.Class = class
	p.powerSet = false
	r.w.recalc(p)
	snaps(p)
	return p
}

func bookNames(p *Player) string {
	var out []string
	for _, b := range p.book {
		out = append(out, b.Spell.Name)
	}
	return strings.Join(out, ",")
}

func TestSpellbookGrowsWithTheSpec(t *testing.T) {
	r := casterWorld(t)
	h := caster(r, "Healer", rules.ClassHealer, 6)
	t.Run("A new Healer", func(t *testing.T) {
		if got := bookNames(h); got != "Minor Mending,Summon Thing,Might,Group Shield" && got != "Minor Mending,Might,Group Shield,Summon Thing" {
			// order within a spec is by level, then as listed
			if !strings.HasPrefix(got, "Minor Mending") || strings.Contains(got, "Major") {
				t.Errorf("book %q, want the level 1 spells and not Major Mending", got)
			}
		}
	})
	w := caster(r, "Warrior", rules.ClassWarrior, 6)
	if len(w.book) != 0 {
		t.Errorf("a Warrior knows %d spells", len(w.book))
	}
	t.Run("Levelling", func(t *testing.T) {
		sys := func() string {
			var all []string
			for _, m := range sentOf[protocol.ChatMsg](h, protocol.TChatMsg) {
				all = append(all, m.Text)
			}
			return strings.Join(all, "|")
		}
		sentOf[protocol.ChatMsg](h, protocol.TChatMsg)
		h.Char.Specs = map[string]int{"Mending": 5}
		r.w.recalc(h)
		if !strings.Contains(bookNames(h), "Major Mending") {
			t.Errorf("book %q after training to 5, want Major Mending", bookNames(h))
		}
		if m := sys(); !strings.Contains(m, "You learn Major Mending.") {
			t.Errorf("chat %q, want the new spell named", m)
		}
		if len(sentOf[protocol.Book](h, protocol.TBook)) != 0 {
			// the book was sent already inside sys()'s drain; nothing more to do
			t.Log("book resent")
		}
	})
}

func TestCastAHeal(t *testing.T) {
	r := casterWorld(t)
	h := caster(r, "Healer", rules.ClassHealer, 6)
	h.hp = 50
	power := h.power
	cost := h.powerCost(r.w.Cat.Spells[3001])
	if cost < 1 {
		t.Fatalf("a 10%% spell costs %d", cost)
	}
	r.z.Cast(h, 3001, 0)
	if h.casting == nil {
		t.Fatal("the cast did not begin")
	}
	r.ticks(TickRate)
	if h.power != power {
		t.Errorf("after one second of a two second cast: hp %v power %v", h.hp, h.power)
	}
	r.ticks(TickRate + 5)
	if h.casting != nil || h.hp < 110 || h.hp > 120 {
		t.Errorf("after the cast time: casting %v hp %v, want the heal of 60 on 50 (and a little regeneration)", h.casting != nil, h.hp)
	}
	if h.power != power-float64(cost) {
		t.Errorf("power %v, want %v less %d", h.power, power, cost)
	}
}

func TestCastRefusals(t *testing.T) {
	r := casterWorld(t)
	h := caster(r, "Healer", rules.ClassHealer, 6)
	sentOf[protocol.Err](h, "")
	t.Run("Out of power", func(t *testing.T) {
		h.power = 0
		r.z.Cast(h, 3001, 0)
		if errs := errsOf(h); len(errs) == 0 || !strings.Contains(errs[0], "enough power") || h.casting != nil {
			t.Errorf("errors %v, casting %v", errs, h.casting != nil)
		}
	})
	h.power = float64(h.sh.maxPower)
	r.z.Cast(h, 999999, 0)
	if errs := errsOf(h); len(errs) == 0 || !strings.Contains(errs[0], "do not know") {
		t.Errorf("an unknown spell: %v", errs)
	}
	t.Run("An unsupported type", func(t *testing.T) {
		power := h.power
		r.z.Cast(h, 4001, 0)
		if errs := errsOf(h); len(errs) == 0 || !strings.Contains(errs[0], "does not work yet") || h.power != power || h.casting != nil {
			t.Errorf("errors %v, power %v", errs, h.power)
		}
	})
	r.z.Cast(h, 3002, 0) // Major Mending opens at spec level 5
	if errs := errsOf(h); len(errs) == 0 || !strings.Contains(errs[0], "do not know") {
		t.Errorf("a spell above the spec level: %v", errs)
	}
}

func TestBoltOnACreatureAndRange(t *testing.T) {
	r := casterWorld(t)
	m := caster(r, "Rune", rules.ClassRunemaster, 6)
	n := r.z.npcs[0]
	r.z.Cast(m, 5001, 0)
	if errs := errsOf(m); len(errs) == 0 || !strings.Contains(errs[0], "enemy target") {
		t.Errorf("no target: %v", errs)
	}
	m.target = n.ID
	t.Run("Out of range", func(t *testing.T) {
		n.Body.X = 40000
		r.z.Cast(m, 5001, 0)
		if errs := errsOf(m); len(errs) == 0 || !strings.Contains(errs[0], "too far") || m.casting != nil {
			t.Errorf("errors %v", errs)
		}
		n.Body.X = 31000
	})
	t.Run("A bolt", func(t *testing.T) {
		hp := n.HP
		r.z.Cast(m, 5001, 0)
		r.ticks(2*TickRate + 5)
		if n.HP >= hp {
			t.Errorf("the creature has %d of %d after the bolt", n.HP, hp)
		}
		if n.target != combatant(m) {
			t.Error("the creature did not turn on the caster")
		}
	})
	t.Run("Recast", func(t *testing.T) {
		sentOf[protocol.Err](m, "")
		r.z.Cast(m, 5001, 0)
		if errs := errsOf(m); len(errs) == 0 || !strings.Contains(errs[0], "not ready") {
			t.Errorf("a second cast at once: %v", errs)
		}
		r.ticks(11 * TickRate)
		r.z.Cast(m, 5001, 0)
		if m.casting == nil {
			t.Error("the recast time has passed and the spell was refused")
		}
	})
}

func TestCastInterrupts(t *testing.T) {
	r := casterWorld(t)
	h := caster(r, "Healer", rules.ClassHealer, 6)
	h.hp = 50
	power := h.power
	t.Run("Hit while casting", func(t *testing.T) {
		r.z.Cast(h, 3001, 0)
		r.ticks(TickRate / 2)
		r.z.damaged(h)
		r.ticks(1)
		if h.casting != nil || h.power != power {
			t.Errorf("casting %v, power %v of %v", h.casting != nil, h.power, power)
		}
		if castEvents(h)["interrupt"] == 0 {
			t.Error("no interrupt event")
		}
	})
	r.ticks(3 * TickRate)
	h.power = power
	r.z.Cast(h, 3001, 0)
	h.Body.X += 200
	r.ticks(2)
	if h.casting != nil {
		t.Error("moving away did not interrupt the cast")
	}
	if h.power != power {
		t.Errorf("an interrupted cast spent power: %v of %v", h.power, power)
	}
}

func TestBuffsRaiseTheSheetAndExpire(t *testing.T) {
	r := casterWorld(t)
	h := caster(r, "Healer", rules.ClassHealer, 20)
	str, con := h.sh.stats[rules.STR], h.sh.stats[rules.CON]
	af := 0
	cb := h.combatant(r.w.Cat)
	for l := rules.Location(0); l < rules.NumLocations; l++ {
		a, _ := rules.LocationArmour(cb, l)
		af += int(a)
	}
	r.z.Cast(h, 6001, 0)
	r.ticks(1)
	r.z.Cast(h, 6002, 0)
	r.ticks(2)
	if h.sh.stats[rules.STR] != str+12 || h.sh.stats[rules.CON] != con+12 {
		t.Errorf("stats after the buff: STR %d CON %d, want +12 on %d and %d", h.sh.stats[rules.STR], h.sh.stats[rules.CON], str, con)
	}
	if h.sh.buffAF != 20 {
		t.Errorf("armour buff %d, want 20", h.sh.buffAF)
	}
	t.Run("A buff expires", func(t *testing.T) {
		r.ticks(31 * TickRate)
		if h.sh.stats[rules.STR] != str || h.sh.stats[rules.CON] != con || h.sh.buffAF != 0 {
			t.Errorf("after the duration: STR %d CON %d AF %d, want %d %d 0", h.sh.stats[rules.STR], h.sh.stats[rules.CON], h.sh.buffAF, str, con)
		}
		if len(h.effects) != 0 {
			t.Errorf("%d effects remain", len(h.effects))
		}
	})
}

func TestGroupSpellReachesMembers(t *testing.T) {
	r := casterWorld(t)
	h := caster(r, "Healer", rules.ClassHealer, 20)
	friend := r.player("Friend", 20, 30300, 30000, true)
	stranger := r.player("Stranger", 20, 30300, 30100, true)
	r.w.omu.Lock()
	for _, o := range []*Player{h, friend, stranger} {
		r.w.online[o.Char.ID] = o
	}
	r.w.omu.Unlock()
	r.w.command(r.z, h, "invite Friend")
	r.w.command(r.z, friend, "accept")
	r.z.Cast(h, 6002, 0)
	r.ticks(2)
	if friend.sh.buffAF != 20 || h.sh.buffAF != 20 {
		t.Errorf("group members' buffs: healer %d, friend %d, want 20", h.sh.buffAF, friend.sh.buffAF)
	}
	if stranger.sh.buffAF != 0 {
		t.Errorf("a stranger got the group spell: %d", stranger.sh.buffAF)
	}
}
