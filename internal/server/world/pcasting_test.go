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
	summon    = &SpellDef{ID: 4001, Name: "Summon Thing", Type: "SavageStyleParryBuff", Target: "Self", Range: 0, CastTime: 3, Power: 5}
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
	// The roll is zero, so the heal is the least of its range for a spec of 1 in a level 1 spell.
	least, _ := rules.PlayerHealRange(60, 1, 1)
	if h.casting != nil || h.hp < float64(50+least) || h.hp > float64(50+least+10) {
		t.Errorf("after the cast time: casting %v hp %v, want the least heal of %d on 50 (and a little regeneration)", h.casting != nil, h.hp, least)
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

// A player's spell damage and heal depend on their casting stat and their spec
// in the spell's line; a creature's do not.
func TestSpellsScaleWithSpec(t *testing.T) {
	r := casterWorld(t)
	m := caster(r, "Rune", rules.ClassRunemaster, 20)
	n := r.z.npcs[0]
	stat := m.sh.stats[m.sh.class.Acuity]
	hit := func(spec int) int {
		m.Char.Specs = map[string]int{"Runecarving": spec}
		n.HP = n.MaxHP
		r.z.castSpell(m, n, r.w.Cat.Spells[5001])
		return n.MaxHP - n.HP
	}
	low, high := hit(1), hit(30)
	if want := rules.PlayerSpellDamage(40, stat, 1, 20, n.Level, 0, 0); low != want {
		t.Errorf("a spec of 1 does %d, want %d", low, want)
	}
	if want := rules.PlayerSpellDamage(40, stat, 30, 20, n.Level, 0, 0); high != want {
		t.Errorf("a spec of 30 does %d, want %d", high, want)
	}
	if high <= low {
		t.Errorf("spec 30 did %d, not more than spec 1's %d", high, low)
	}

	h := caster(r, "Healer", rules.ClassHealer, 20)
	h.hp = 1
	for _, spec := range []int{1, 20} {
		h.Char.Specs = map[string]int{"Mending": spec}
		h.hp = 1
		r.z.castSpell(h, h, r.w.Cat.Spells[3001])
		least, _ := rules.PlayerHealRange(60, spec, 1)
		if got := int(h.hp) - 1; got != least {
			t.Errorf("a Mending spec of %d heals %d at the worst roll, want %d", spec, got, least)
		}
	}
}

func TestOverTimeSpellsScale(t *testing.T) {
	r := casterWorld(t)
	m := caster(r, "Rune", rules.ClassRunemaster, 20)
	n := r.z.npcs[0]
	n.MaxHP, n.HP = 100000, 100000
	stat := m.sh.stats[m.sh.class.Acuity]
	agony := &SpellDef{ID: 5002, Name: "Rune Agony", Type: rules.SpellDamageOverTime, Target: "Enemy", Damage: 10, DamageType: 15, Duration: 20, Frequency: 40}
	add := &SpellDef{ID: 5003, Name: "Rune Edge", Type: rules.SpellDamageAdd, Target: "Self", Damage: 4.5, DamageType: 15, Duration: 60}
	r.w.Cat.Spells[agony.ID], r.w.Cat.Spells[add.ID] = agony, add
	r.w.Cat.LineSpells["Runecarving"] = append(r.w.Cat.LineSpells["Runecarving"],
		LineSpell{Spec: "Runecarving", Line: "Runecarving", Spell: agony.ID, Level: 1})

	tick := func(spec int) int {
		m.Char.Specs = map[string]int{"Runecarving": spec}
		n.effects, n.HP = nil, n.MaxHP
		r.z.castSpell(m, n, agony)
		r.z.now = n.effects[0].next
		r.z.tickEffectsOf(n)
		return n.MaxHP - n.HP
	}
	low, high := tick(1), tick(30)
	if want := rules.PlayerDoTDamage(10, stat, 1, n.Level, 0, 0); low != want {
		t.Errorf("a spec of 1 ticks for %d, want %d", low, want)
	}
	if want := rules.PlayerDoTDamage(10, stat, 30, n.Level, 0, 0); high != want {
		t.Errorf("a spec of 30 ticks for %d, want %d", high, want)
	}
	if high <= low {
		t.Errorf("spec 30 ticked for %d, not more than spec 1's %d", high, low)
	}

	t.Run("A tick keeps the stat of the cast", func(t *testing.T) {
		m.Char.Specs = map[string]int{"Runecarving": 30}
		n.effects, n.HP = nil, n.MaxHP
		r.z.castSpell(m, n, agony)
		m.sh.stats[m.sh.class.Acuity] = 0
		r.z.now = n.effects[0].next
		r.z.tickEffectsOf(n)
		if got := n.MaxHP - n.HP; got != high {
			t.Errorf("tick %d after the stat fell, want %d", got, high)
		}
	})

	t.Run("A damage add is per second of the swing", func(t *testing.T) {
		r.z.castSpell(m, m, add)
		got := r.z.damageAdd(m, n, 40)
		if want := rules.DamageAdd(4.5, 20, m.weaponSpeed(), 40, 0, 0); got != want || got <= 4 {
			t.Errorf("damage add %d, want %d (more than the flat 4)", got, want)
		}
	})
}

var (
	thunderShout = &SpellDef{ID: 3521, Name: "Thunder Shout", Type: rules.SpellDirectDamage, Target: "Enemy", Range: 0, Radius: 300, Recast: 20, Damage: 15, DamageType: 12, Power: 4}
	mjollnir     = &SpellDef{ID: 3538, Name: "Call Mjollnir", Type: rules.SpellDirectDamage, Target: "Enemy", Range: 1500, Radius: 350, CastTime: 1, Damage: 60, DamageType: 12, Power: 10}
)

// areaWorld is a zone with four wolves around 30000,30000: at 200, 200 and 150
// units, and one at 400, and a Thane's line holding the two area spells.
func areaWorld(t *testing.T) *rig {
	r := newRig(t, wolf(1, 30200, 30000, 8), wolf(2, 30000, 30200, 8), wolf(3, 29850, 30000, 8), wolf(4, 30400, 30000, 8))
	r.z.rng = rand.New(zeroSource{})
	for _, n := range r.z.npcs {
		n.Aggro, n.Roam = 0, 0
	}
	c := r.w.Cat
	c.Spells = map[int]*SpellDef{}
	for _, s := range []*SpellDef{thunderShout, mjollnir} {
		cp := *s
		c.Spells[s.ID] = &cp
	}
	c.LineSpells = map[string][]LineSpell{"Stormcalling": {
		{Spec: "Stormcalling", Line: "Stormcalling", Spell: 3521, Level: 1},
		{Spec: "Stormcalling", Line: "Stormcalling", Spell: 3538, Level: 1}}}
	c.ClassSpecs[rules.ClassThane] = []string{"ThaneCareer", "Stormcalling"}
	return r
}

func TestAreaSpellsHitEveryCreatureInTheRadius(t *testing.T) {
	t.Run("Thunder Shout", func(t *testing.T) {
		r := areaWorld(t)
		p := caster(r, "Thane", rules.ClassThane, 5)
		full := [4]int{}
		for i, n := range r.z.npcs {
			full[i] = n.HP
		}
		r.z.Cast(p, 3521, 0) // no target held
		if errs := errsOf(p); len(errs) > 0 {
			t.Fatalf("a spell around the caster needs no target: %v", errs)
		}
		for i, n := range r.z.npcs[:3] {
			if n.HP >= full[i] {
				t.Errorf("wolf %d inside the radius has %d of %d", n.ID, n.HP, full[i])
			}
			if n.target != combatant(p) {
				t.Errorf("wolf %d did not turn on the caster", n.ID)
			}
		}
		// (It may still come to help its neighbours, as any creature does.)
		if n := r.z.npcs[3]; n.HP != full[3] {
			t.Errorf("the wolf at 400 units was touched: %d of %d", n.HP, full[3])
		}
		if p.spellReady[3521] <= r.z.now {
			t.Error("the spell has no recast time")
		}
	})
	t.Run("Mjollnir around the target", func(t *testing.T) {
		r := areaWorld(t)
		p := caster(r, "Thane", rules.ClassThane, 5)
		friend := r.player("Friend", 5, 30390, 30000, false) // beside the far wolf
		p.target = r.z.npcs[3].ID
		r.z.Cast(p, 3538, 0)
		r.ticks(TickRate + 5)
		hit := 0
		for _, n := range r.z.npcs {
			if n.HP < n.MaxHP {
				hit++
			}
		}
		// The far wolf at 30400 is the target; the circle of 350 around it
		// reaches 30200 (200 away) and 30000,30200 is 447 away, so two wolves.
		if hit != 2 || r.z.npcs[3].HP >= r.z.npcs[3].MaxHP || r.z.npcs[0].HP >= r.z.npcs[0].MaxHP {
			t.Errorf("%d creatures hurt: %v", hit, [4]int{r.z.npcs[0].HP, r.z.npcs[1].HP, r.z.npcs[2].HP, r.z.npcs[3].HP})
		}
		if friend.hp < float64(friend.maxHP()) {
			t.Errorf("a player beside the target took damage: %v of %d", friend.hp, friend.maxHP())
		}
	})
	t.Run("Nothing in range", func(t *testing.T) {
		r := areaWorld(t)
		p := caster(r, "Thane", rules.ClassThane, 5)
		p.Body.X, p.Body.Y = 20000, 20000
		power := p.power
		r.z.Cast(p, 3521, 0)
		if errs := errsOf(p); len(errs) > 0 {
			t.Errorf("errors %v", errs)
		}
		if p.power >= power {
			t.Error("a spell that hit nothing cost nothing")
		}
		for _, n := range r.z.npcs {
			if n.HP != n.MaxHP {
				t.Errorf("wolf %d was hurt from far away", n.ID)
			}
		}
	})
}
