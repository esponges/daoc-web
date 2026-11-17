package world

import (
	"math/rand"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

var (
	chantBrawl = &SpellDef{ID: 3601, Name: "Chant of the Brawl", Type: rules.SpellDamageAdd, Target: "Group", Range: 1500,
		Duration: 5, Frequency: 50, Pulse: 1, Recast: 8, Damage: 1.3, DamageType: 10}
	chantFight = &SpellDef{ID: 3602, Name: "Chant of the Fight", Type: rules.SpellDamageAdd, Target: "Group", Range: 1500,
		Duration: 5, Frequency: 50, Pulse: 1, Recast: 8, Damage: 1.9, DamageType: 10}
	songCostly = &SpellDef{ID: 3603, Name: "Costly Song", Type: rules.SpellDamageAdd, Target: "Group", Range: 1500,
		Duration: 5, Frequency: 20, Pulse: 1, PulsePower: 4, Damage: 1.3, DamageType: 10}
)

// songWorld is a zone, a Skald's Battlesongs line of two chants and a costly
// song, and a Skald with two group mates in range.
func songWorld(t *testing.T) (r *rig, skald, a, b *Player) {
	r = newRig(t)
	r.z.rng = rand.New(zeroSource{})
	c := r.w.Cat
	c.Spells = map[int]*SpellDef{}
	for _, s := range []*SpellDef{chantBrawl, chantFight, songCostly} {
		cp := *s
		c.Spells[s.ID] = &cp
	}
	c.LineSpells = map[string][]LineSpell{"Battlesongs": {
		{Spec: "Battlesongs", Line: "Battlesongs", Spell: 3601, Level: 1},
		{Spec: "Battlesongs", Line: "Battlesongs", Spell: 3602, Level: 1},
		{Spec: "Battlesongs", Line: "Battlesongs", Spell: 3603, Level: 1}}}
	c.ClassSpecs[rules.ClassSkald] = []string{"SkaldCareer", "Battlesongs"}
	skald = caster(r, "Skald", rules.ClassSkald, 10)
	a = r.player("Alma", 10, 30300, 30000, true)
	b = r.player("Bjorn", 10, 30400, 30000, true)
	r.w.omu.Lock()
	for _, o := range []*Player{skald, a, b} {
		r.w.online[o.Char.ID] = o
	}
	r.w.omu.Unlock()
	for _, o := range []*Player{a, b} {
		r.w.command(r.z, skald, "invite "+o.Char.Name)
		r.w.command(r.z, o, "accept")
	}
	return r, skald, a, b
}

func seconds(n float64) int { return int(n * float64(TickRate)) }

func TestSongPulsesOnTheGroup(t *testing.T) {
	t.Run("A chant on a group of three", func(t *testing.T) {
		r, s, a, b := songWorld(t)
		r.z.Cast(s, 3601, 0)
		for _, p := range []*Player{s, a, b} {
			if !hasEffect(p, 3601) {
				t.Errorf("%s does not have the chant at once", p.Char.Name)
			}
		}
		if !s.songPlaying(r.w.Cat.Spells[3601]) {
			t.Fatal("the Skald's song is not on")
		}
		r.ticks(seconds(13)) // past two pulses; one application lasts only six seconds
		for _, p := range []*Player{s, a, b} {
			if !hasEffect(p, 3601) {
				t.Errorf("%s lost the chant while it is on", p.Char.Name)
			}
		}
	})
	t.Run("Leaving range", func(t *testing.T) {
		r, s, a, b := songWorld(t)
		r.z.Cast(s, 3601, 0)
		b.Body.X = 33000 // 3000 units away
		r.ticks(seconds(7))
		if hasEffect(b, 3601) {
			t.Error("a member out of range still has the chant")
		}
		if !hasEffect(a, 3601) {
			t.Error("a member in range lost it")
		}
		b.Body.X = 30400
		r.ticks(seconds(6))
		if !hasEffect(b, 3601) {
			t.Error("a member who came back in range did not get it again")
		}
	})
	t.Run("Toggling off", func(t *testing.T) {
		r, s, a, _ := songWorld(t)
		r.z.Cast(s, 3601, 0)
		r.z.Cast(s, 3601, 0) // the same chant again
		if s.song != nil {
			t.Fatal("casting the chant again did not turn it off")
		}
		if errs := errsOf(s); len(errs) > 0 {
			t.Errorf("turning it off said %v", errs)
		}
		r.ticks(seconds(8))
		if hasEffect(a, 3601) || hasEffect(s, 3601) {
			t.Error("the chant is still on the group after it was turned off")
		}
	})
	t.Run("A second song", func(t *testing.T) {
		r, s, a, _ := songWorld(t)
		r.z.Cast(s, 3601, 0)
		r.z.Cast(s, 3602, 0)
		if s.song == nil || s.song.spell.ID != 3602 {
			t.Fatalf("the song on is %v, want Chant of the Fight", s.song)
		}
		r.ticks(seconds(8))
		if hasEffect(a, 3601) || !hasEffect(a, 3602) {
			t.Error("the first chant did not give way to the second")
		}
	})
	t.Run("The caster dies", func(t *testing.T) {
		r, s, a, _ := songWorld(t)
		r.z.Cast(s, 3601, 0)
		r.z.killPlayer(s, nil)
		if s.song != nil {
			t.Error("a dead Skald is still playing")
		}
		r.ticks(seconds(8))
		if hasEffect(a, 3601) {
			t.Error("the chant outlived its caster")
		}
	})
	t.Run("A held caster", func(t *testing.T) {
		r, s, a, _ := songWorld(t)
		r.z.Cast(s, 3601, 0)
		mez := &SpellDef{ID: 9001, Name: "Mez", Type: rules.SpellMesmerize, Target: "Enemy", Duration: 12}
		r.z.addEffect(s, a, mez, effMez)
		r.ticks(seconds(8))
		if s.song == nil {
			t.Error("a held Skald lost the song")
		}
		if hasEffect(a, 3601) {
			t.Error("a held Skald pulsed")
		}
		r.z.clearEffects(s)
		r.ticks(seconds(6))
		if !hasEffect(a, 3601) {
			t.Error("the song did not resume when the Skald was free")
		}
	})
	t.Run("Power per pulse", func(t *testing.T) {
		r, s, _, _ := songWorld(t)
		s.power, s.powerAt = 10, 1e9 // no regeneration, so the cost is exact
		r.z.Cast(s, 3603, 0)
		r.ticks(seconds(2) + 1) // first pulse
		if s.power != 6 {
			t.Errorf("after one pulse the Skald has %v power, want 6", s.power)
		}
		r.ticks(seconds(2)) // second pulse: 2 left
		if s.power != 2 || s.song == nil {
			t.Errorf("after two pulses: power %v, song on %v, want 2 and on", s.power, s.song != nil)
		}
		r.ticks(seconds(2)) // third: not enough
		if s.song != nil {
			t.Error("the song kept going without power")
		}
	})
}

// The book marks a song, and the effects message names the one that is on.
func TestSongIsOnTheWire(t *testing.T) {
	r, s, _, _ := songWorld(t)
	m := s.bookMsg(r.w.Cat)
	var marked int
	for _, l := range m.Lines {
		for _, sp := range l.Spells {
			if sp.Pulse {
				marked++
			}
		}
	}
	if marked != 3 {
		t.Errorf("%d songs are marked in the book, want 3", marked)
	}
	sentOf[protocol.Effects](s, protocol.TEffects) // drop what came before
	r.z.Cast(s, 3601, 0)
	got := sentOf[protocol.Effects](s, protocol.TEffects)
	if len(got) == 0 || got[len(got)-1].Song != 3601 {
		t.Errorf("effects after turning the chant on: %+v", got)
	}
	r.z.Cast(s, 3601, 0)
	got = sentOf[protocol.Effects](s, protocol.TEffects)
	if len(got) == 0 || got[len(got)-1].Song != 0 {
		t.Errorf("effects after turning it off: %+v", got)
	}
}

var (
	songTravel = &SpellDef{ID: 3608, Name: "Simple Song of Travel", Type: rules.SpellSpeedEnhancement, Target: "Group", Range: 2000,
		Duration: 5, Frequency: 50, Pulse: 1, Recast: 8, Value: 144}
	songTravel2 = &SpellDef{ID: 3609, Name: "Song of Travel", Type: rules.SpellSpeedEnhancement, Target: "Group", Range: 2000,
		Duration: 5, Frequency: 50, Pulse: 1, Recast: 8, Value: 159}
	songRest = &SpellDef{ID: 3613, Name: "Simple Song of Rest", Type: rules.SpellHealthRegenBuff, Target: "Group", Range: 1500,
		Duration: 5, Frequency: 50, Pulse: 1, Recast: 8, Value: 10}
	chantHeat = &SpellDef{ID: 3657, Name: "Heat Diminishing Chant", Type: rules.SpellHeatResistBuff, Target: "Group", Range: 1500,
		Duration: 5, Frequency: 50, Pulse: 1, Recast: 8, Value: 10}
	chantBody = &SpellDef{ID: 3651, Name: "Body Bolstering Chant", Type: rules.SpellBodyResistBuff, Target: "Group", Range: 1500,
		Duration: 5, Frequency: 50, Pulse: 1, Recast: 8, Value: 30}
	chantBody2 = &SpellDef{ID: 3652, Name: "Body Bolstering Song", Type: rules.SpellBodyResistBuff, Target: "Group", Range: 1500,
		Duration: 5, Frequency: 50, Pulse: 1, Recast: 8, Value: 10}
	slow40 = &SpellDef{ID: 9002, Name: "Slow", Type: rules.SpellSpeedDecrease, Target: "Enemy", Duration: 30, Value: 40}
)

// songBook adds the speed, regeneration and resist songs to the Skald's line,
// and to the table.
func songBook(r *rig, s *Player) {
	for _, sp := range []*SpellDef{songTravel, songTravel2, songRest, chantHeat, chantBody, chantBody2, slow40} {
		cp := *sp
		r.w.Cat.Spells[sp.ID] = &cp
	}
	for _, id := range []int{3608, 3609, 3613, 3657, 3651, 3652} {
		r.w.Cat.LineSpells["Battlesongs"] = append(r.w.Cat.LineSpells["Battlesongs"], LineSpell{Spec: "Battlesongs", Line: "Battlesongs", Spell: id, Level: 1})
	}
	r.w.refreshBook(s)
}

func TestSpeedSong(t *testing.T) {
	near := func(t *testing.T, what string, got, want float64) {
		t.Helper()
		if d := got - want; d > 1e-9 || d < -1e-9 {
			t.Errorf("%s: %v, want %v", what, got, want)
		}
	}
	t.Run("Simple Song of Travel", func(t *testing.T) {
		r, s, a, b := songWorld(t)
		songBook(r, s)
		r.z.Cast(s, 3608, 0)
		for _, p := range []*Player{s, a, b} {
			near(t, p.Char.Name+"'s speed", speedFactor(p), 1.44)
		}
		got := sentOf[protocol.Effects](a, protocol.TEffects)
		if len(got) == 0 || got[len(got)-1].Speed != 1.44 || got[len(got)-1].List[0].Kind != "speed" {
			t.Errorf("the group mate was sent %+v", got)
		}
	})
	t.Run("The strongest counts", func(t *testing.T) {
		r, s, a, _ := songWorld(t)
		songBook(r, s)
		r.z.castSpell(s, a, r.w.Cat.Spells[3608])
		r.z.castSpell(s, a, r.w.Cat.Spells[3609])
		near(t, "a 144 and a 159", speedFactor(a), 1.59)
	})
	t.Run("With a slow", func(t *testing.T) {
		r, s, a, _ := songWorld(t)
		songBook(r, s)
		r.z.castSpell(s, a, r.w.Cat.Spells[3608])
		r.z.addEffect(a, s, r.w.Cat.Spells[9002], effSlow)
		near(t, "a 144 song and a 40% slow", speedFactor(a), 1.44*0.6)
	})
	t.Run("Combat", func(t *testing.T) {
		r, s, a, _ := songWorld(t)
		songBook(r, s)
		r.z.Cast(s, 3608, 0)
		a.combatUntil = r.z.now + 100
		r.ticks(2)
		if hasEffect(a, 3608) || speedFactor(a) != 1 {
			t.Error("the song stayed on someone who entered combat")
		}
		r.ticks(seconds(6)) // a pulse falls due while they are still fighting
		if hasEffect(a, 3608) {
			t.Error("the song took hold in combat")
		}
		a.combatUntil = 0
		r.ticks(seconds(6))
		if !hasEffect(a, 3608) {
			t.Error("the song did not return when the fight was over")
		}
	})
}

func TestRegenSongAddsToHealthTicks(t *testing.T) {
	r, s, a, b := songWorld(t)
	songBook(r, s)
	r.z.castSpell(s, a, r.w.Cat.Spells[3613]) // Song of Rest on Alma only
	tick := func(p *Player) float64 {
		p.hp, p.healAt = 10, r.z.now
		before := p.hp
		r.ticks(1)
		return p.hp - before
	}
	plain := tick(b)
	if plain <= 0 {
		t.Fatalf("a resting player regains %v", plain)
	}
	if got := tick(a); got != plain+10 {
		t.Errorf("with the song a tick returns %v, want %v", got, plain+10)
	}
	a.combatUntil = r.z.now + 100
	if got := tick(a); got != 0 {
		t.Errorf("a tick in combat returns %v", got)
	}
}

func TestResistChantsRaiseTheResist(t *testing.T) {
	r, s, a, _ := songWorld(t)
	songBook(r, s)
	heat, cold, body := r.z.spellResist(a, rules.DamageHeat), r.z.spellResist(a, rules.DamageCold), r.z.spellResist(a, rules.DamageBody)
	r.z.castSpell(s, a, r.w.Cat.Spells[3657])
	if got := r.z.spellResist(a, rules.DamageHeat); got != heat+10 {
		t.Errorf("heat resist %v, want %v", got, heat+10)
	}
	if got := r.z.spellResist(a, rules.DamageCold); got != cold {
		t.Errorf("a heat chant moved the cold resist to %v", got)
	}
	r.z.castSpell(s, a, r.w.Cat.Spells[3651]) // a 30-point body chant
	r.z.castSpell(s, a, r.w.Cat.Spells[3652]) // and a 10-point one
	if got := r.z.spellResist(a, rules.DamageBody); got != body+24 {
		t.Errorf("body resist %v, want %v (the buff cap)", got, body+24)
	}
	// The spell meets it: 100 damage of heat takes off the chant's share.
	if want, got := rules.SpellDamage(100, 1, heat+10), rules.SpellDamage(100, 1, r.z.spellResist(a, rules.DamageHeat)); got != want {
		t.Errorf("damage under the chant %d, want %d", got, want)
	}
}
