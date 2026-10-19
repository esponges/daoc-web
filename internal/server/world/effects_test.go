package world

import (
	"encoding/json"
	"math"
	"math/rand"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// zeroSource makes every roll 0, so every hostile spell lands.
type zeroSource struct{}

func (zeroSource) Int63() int64 { return 0 }
func (zeroSource) Seed(int64)   {}

func spellRig(t *testing.T) (*rig, *Player, *NPC) {
	r := newRig(t, wolf(1, 30300, 30000, 10))
	r.z.rng = rand.New(zeroSource{})
	n := r.z.npcs[0]
	n.Aggro, n.Roam = 0, 0
	p := r.player("Target", 10, 30000, 30000, true)
	p.hp = 500
	lines(p)
	return r, p, n
}

func lastEffects(p *Player) (m protocol.Effects, ok bool) {
	for {
		select {
		case b := <-p.out:
			var e protocol.Envelope
			json.Unmarshal(b, &e)
			if e.T == protocol.TEffects {
				json.Unmarshal(b, &m)
				ok = true
			}
		default:
			return
		}
	}
}

var dot = &SpellDef{ID: 15001, Name: "Mind Agony", Type: rules.SpellDamageOverTime, Target: "Enemy", Damage: 10, DamageType: 15, Duration: 20, Frequency: 40}

func TestDamageOverTimeTicksFiveTimes(t *testing.T) {
	r, p, n := spellRig(t)
	hp := p.hp
	if !r.z.castSpell(n, p, dot) {
		t.Fatal("the spell did not land")
	}
	if m, ok := lastEffects(p); !ok || len(m.List) != 1 || m.List[0].Kind != effDoT || m.List[0].Left != 20 {
		t.Fatalf("effects sent %+v", m)
	}
	ticks := 0
	var last protocol.Effects
	for i := 0; i < TickRate*25; i++ {
		before := p.hp
		r.z.Tick()
		if m, ok := lastEffects(p); ok {
			last = m
		}
		if p.hp < before {
			ticks++
			if d := before - p.hp; d != 10 {
				t.Errorf("a tick did %v, want 10", d)
			}
		}
	}
	if ticks != 5 || hp-p.hp != 50 {
		t.Errorf("%d ticks, %v damage; want 5 and 50", ticks, hp-p.hp)
	}
	if len(p.effects) != 0 {
		t.Error("the effect outlived its duration")
	}
	if last.T != protocol.TEffects || len(last.List) != 0 || last.Speed != 1 {
		t.Errorf("no word that it ended: %+v", last)
	}
}

func TestDamageOverTimeRecastRestarts(t *testing.T) {
	r, p, n := spellRig(t)
	r.z.castSpell(n, p, dot)
	r.ticks(TickRate*8 + 1) // two ticks
	r.z.castSpell(n, p, dot)
	if len(p.effects) != 1 {
		t.Fatalf("%d effects, want one", len(p.effects))
	}
	if left := p.effects[0].ends - r.z.now; math.Abs(left-20) > 1e-6 {
		t.Errorf("%v seconds to run, want 20", left)
	}
}

func TestMesmerizeBrokenByDamage(t *testing.T) {
	r, p, n := spellRig(t)
	mez := &SpellDef{ID: 1171, Name: "Compelling Cadence", Type: rules.SpellMesmerize, Target: "Enemy", Duration: 12}
	stun := &SpellDef{ID: 1161, Name: "Stunning Yell", Type: rules.SpellStun, Target: "Enemy", Duration: 3}
	r.z.castSpell(n, p, mez)
	if !held(p) || speedFactor(p) != 0 {
		t.Fatal("a mesmerized player can move")
	}
	r.z.hurt(p, n, 5)
	if held(p) {
		t.Error("damage did not end the mesmerize")
	}
	r.z.castSpell(n, p, stun)
	r.z.hurt(p, n, 5)
	if !held(p) {
		t.Error("damage ended a stun")
	}
	r.ticks(TickRate*3 + 1)
	if held(p) {
		t.Error("the stun outlived its duration")
	}
}

func TestEffectsEndAtDeath(t *testing.T) {
	r, p, n := spellRig(t)
	r.z.castSpell(n, p, dot)
	r.z.killPlayer(p, n)
	if len(p.effects) != 0 {
		t.Error("a dead player keeps their effects")
	}
	r.z.castSpell(p, n, dot)
	r.z.killNPC(n)
	if len(n.effects) != 0 {
		t.Error("a dead creature keeps its effects")
	}
}

// A heal and a drain: the heal is capped at the maximum, the drain returns
// its share to the caster.
func TestHealAndDrain(t *testing.T) {
	r, p, n := spellRig(t)
	n.HP = n.MaxHP - 50
	r.z.castSpell(n, n, &SpellDef{ID: 1, Name: "Mend", Type: rules.SpellHeal, Target: "Self", Value: 120})
	if n.HP != n.MaxHP {
		t.Errorf("healed to %d of %d", n.HP, n.MaxHP)
	}
	n.HP = 10
	hp := p.hp
	r.z.castSpell(n, p, &SpellDef{ID: 2, Name: "Drain", Type: rules.SpellLifedrain, Target: "Enemy", Damage: 40, DamageType: 10, LifeDrain: 50})
	took := hp - p.hp
	if took <= 0 || n.HP != 10+int(took)/2 {
		t.Errorf("player took %v, creature at %d; want it healed by half", took, n.HP)
	}
}
