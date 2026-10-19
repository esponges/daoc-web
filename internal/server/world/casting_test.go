package world

import (
	"math"
	"math/rand"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

var (
	bolt  = &SpellDef{ID: 2901, Name: "Simple Runebolt", Type: rules.SpellDirectDamage, Target: "Enemy", Range: 1500, CastTime: 3, Recast: 30, Damage: 20, DamageType: 12}
	mend  = &SpellDef{ID: 10131, Name: "Mend", Type: rules.SpellHeal, Target: "Realm", Range: 2000, CastTime: 2, Value: 1000}
	shrud = &SpellDef{ID: 10082, Name: "Shroud of Doubt", Type: rules.SpellSpeedDecrease, Target: "Enemy", Range: 1500, Duration: 30, Value: 50}
	armor = &SpellDef{ID: 777, Name: "Bark Skin", Type: "BaseArmorFactorBuff", Target: "Self", Duration: 600}
)

// casterRig is a zone with one creature that knows the given spells, standing
// 1,000 units from a level 10 player, every hostile spell landing.
func casterRig(t *testing.T, spells ...*SpellDef) (*rig, *NPC, *Player) {
	r := newRig(t, wolf(1, 31000, 30000, 10))
	r.z.rng = rand.New(zeroSource{})
	n := r.z.npcs[0]
	n.Aggro, n.Roam = 0, 0
	n.Spawn.Template = 900
	r.w.Cat.Spells = map[int]*SpellDef{}
	r.w.Cat.TemplateSpells = map[int][]int{}
	for _, s := range spells {
		r.w.Cat.Spells[s.ID] = s
		r.w.Cat.TemplateSpells[900] = append(r.w.Cat.TemplateSpells[900], s.ID)
	}
	n.learnSpells(r.w.Cat)
	p := r.player("Mark", 10, 30000, 30000, true)
	p.hp = 400
	return r, n, p
}

// casts counts the cast, interrupt and spell events a player has been sent.
func castEvents(p *Player) map[string]int {
	out := map[string]int{}
	for _, s := range snaps(p) {
		for _, e := range s.Ev {
			out[e.K]++
		}
	}
	return out
}

func TestCreatureKnowsOnlyImplementedSpells(t *testing.T) {
	_, n, _ := casterRig(t, bolt, armor)
	if len(n.spells) != 1 || n.spells[0] != bolt {
		t.Errorf("spells %v, want only the bolt", n.spells)
	}
}

func TestCastThenMelee(t *testing.T) {
	r, n, p := casterRig(t, bolt)
	snaps(p)
	n.engage(p, r.z)
	r.ticks(1)
	if n.casting == nil || n.speedNow != 0 {
		t.Fatalf("it should stand and cast at 1,000 units: casting %v speed %v", n.casting, n.speedNow)
	}
	hp := p.hp
	r.ticks(3 * TickRate)
	if n.casting != nil || p.hp >= hp {
		t.Fatalf("after 3 s: casting %v, hp %v (was %v)", n.casting, p.hp, hp)
	}
	ev := castEvents(p)
	if ev["cast"] != 1 || ev["spell"] != 1 {
		t.Errorf("events %v, want one cast and one spell", ev)
	}
	// Now it closes in and swings, the bolt not ready for 30 seconds.
	r.ticks(10 * TickRate)
	if n.casting != nil || math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y) > n.Reach+40 {
		t.Errorf("it should be in melee: casting %v, %v units away", n.casting, math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y))
	}
	if ev := castEvents(p); ev["cast"] != 0 || ev["swing"] == 0 {
		t.Errorf("events %v, want swings and no cast", ev)
	}
}

func TestCastOutOfRangeClosesIn(t *testing.T) {
	r, n, p := casterRig(t, bolt)
	p.Body.X = 33000 // 2,000 units away, beyond the bolt's 1,500
	n.engage(p, r.z)
	d0 := math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y)
	r.ticks(TickRate)
	if n.casting != nil || math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y) >= d0 {
		t.Fatalf("out of range it should close in: casting %v", n.casting)
	}
	for i := 0; i < TickRate*5 && n.casting == nil; i++ {
		r.z.Tick()
	}
	if n.casting == nil {
		t.Fatal("it never cast once in range")
	}
	if d := math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y); d > bolt.Range {
		t.Errorf("began casting %v units away, beyond its range", d)
	}
}

func TestCastTargetRanAway(t *testing.T) {
	r, n, p := casterRig(t, bolt)
	n.engage(p, r.z)
	r.ticks(TickRate)
	p.Body.X = 28000 // 3,000 units off before the cast ends
	hp := p.hp
	r.ticks(2*TickRate + 1)
	if p.hp != hp || n.spellReady[bolt.ID] > r.z.now+rules.CastRetry+1e-6 {
		t.Errorf("hp %v (was %v), ready at %v (now %v): want no damage and a 2 s retry", p.hp, hp, n.spellReady[bolt.ID], r.z.now)
	}
}

func TestMeleeInterruptsCasting(t *testing.T) {
	for _, unint := range []bool{false, true} {
		r, n, p := casterRig(t, bolt)
		b := *bolt
		b.Uninterruptible = unint
		n.spells = []*SpellDef{&b}
		n.engage(p, r.z)
		r.ticks(TickRate) // a second into the cast
		snaps(p)
		r.z.playerSwing(p, n)
		r.ticks(1) // events go out with the snapshot
		ev := castEvents(p)
		if unint {
			if n.casting == nil || ev["interrupt"] != 0 {
				t.Errorf("an uninterruptible cast was stopped")
			}
			continue
		}
		if n.casting != nil || ev["interrupt"] != 1 {
			t.Errorf("casting %v, events %v: want it interrupted and seen", n.casting, ev)
		}
		hp := p.hp
		r.ticks(2*TickRate - 2)
		if p.hp < hp && castEvents(p)["spell"] > 0 {
			t.Error("the interrupted spell still landed")
		}
	}
}

func TestHealerHealsPackmate(t *testing.T) {
	r := newRig(t, wolf(1, 31000, 30000, 10), wolf(2, 31300, 30000, 10))
	r.z.rng = rand.New(zeroSource{})
	healer, mate := r.z.npcs[0], r.z.npcs[1]
	healer.Spawn.Template = 900
	r.w.Cat.Spells = map[int]*SpellDef{mend.ID: mend}
	r.w.Cat.TemplateSpells = map[int][]int{900: {mend.ID}}
	healer.learnSpells(r.w.Cat)
	p := r.player("Puller", 10, 30000, 30000, true)
	healer.engage(p, r.z)
	mate.HP = mate.MaxHP / 2
	r.ticks(1)
	if healer.casting == nil || healer.casting.target != combatant(mate) {
		t.Fatalf("the healer should heal its packmate: %+v", healer.casting)
	}
	r.ticks(2 * TickRate)
	if mate.HP != mate.MaxHP {
		t.Errorf("packmate at %d of %d", mate.HP, mate.MaxHP)
	}
}

func TestSlowNotCastTwice(t *testing.T) {
	r, n, p := casterRig(t, shrud)
	snaps(p)
	n.engage(p, r.z)
	r.ticks(1) // an instant spell: cast and landed in the one tick
	if !hasEffect(p, shrud.ID) {
		t.Fatal("the slow did not take hold")
	}
	r.ticks(TickRate * 3)
	if ev := castEvents(p); ev["cast"] != 1 {
		t.Errorf("%d casts, want one while the slow is on", ev["cast"])
	}
}

func TestCastEventCarriesItsTime(t *testing.T) {
	r, n, p := casterRig(t, bolt)
	snaps(p)
	n.engage(p, r.z)
	r.ticks(1)
	var cast *protocol.Event
	for _, s := range snaps(p) {
		for i, e := range s.Ev {
			if e.K == "cast" {
				cast = &s.Ev[i]
			}
		}
	}
	if cast == nil || cast.From != n.ID || cast.To != p.ID || cast.Text != bolt.Name || cast.Dur != 3 {
		t.Errorf("cast event %+v", cast)
	}
}

// A creature's offensive proc fires on its landed blows.
func TestCreatureProcOnHit(t *testing.T) {
	agony := &SpellDef{ID: 11243, Name: "Fungal Agony", Type: rules.SpellDamageOverTime, Target: "Enemy", Damage: 5, DamageType: 15, Duration: 12, Frequency: 30}
	blight := &SpellDef{ID: 11232, Name: "Blightstrike", Type: rules.SpellOffensiveProc, Target: "Self", Duration: 1200, Frequency: 100, Value: float64(agony.ID)}
	r, n, p := casterRig(t, blight)
	r.w.Cat.Spells[agony.ID] = agony
	if len(n.procs) != 1 || len(n.spells) != 0 {
		t.Fatalf("procs %v spells %v", n.procs, n.spells)
	}
	r.rngHit(n, p)
	if !hasEffect(p, agony.ID) {
		t.Error("a landed blow did not carry the proc")
	}
}

// rngHit lands one blow from a creature on a player.
func (r *rig) rngHit(n *NPC, p *Player) {
	r.z.rng = rand.New(rand.NewSource(3))
	for i := 0; i < 50 && !hasEffect(p, 11243); i++ {
		r.z.rng = rand.New(rand.NewSource(int64(i)))
		r.z.npcHitsPlayer(n, p)
	}
}

// procRig is a level 10 Warrior with a hammer beside a creature that never
// defends unless told to, a stun on Slam (style 164) at chance 100, and the
// stun and a damage proc in the catalog.
func procRig(t *testing.T) (*rig, *Player, *NPC) {
	r := newRig(t, wolf(1, 30100, 30000, 10))
	n := r.z.npcs[0]
	n.Aggro, n.Roam = 0, 0
	n.cb.EvadePct, n.cb.ParryPct, n.cb.BlockPct = 0, 0, 0
	stun := &SpellDef{ID: 20204, Name: "Slam Stun", Type: rules.SpellStun, Target: "Enemy", Duration: 9}
	zap := &SpellDef{ID: 61020, Name: "Energy DD Proc", Type: rules.SpellDDNoVariance, Target: "Enemy", Range: 1500, Damage: 70, DamageType: 12}
	r.w.Cat.Spells = map[int]*SpellDef{stun.ID: stun, zap.ID: zap}
	r.w.Cat.StyleSpells = []StyleSpell{{Class: 0, Style: 164, Spell: stun.ID, Chance: 100}}
	p := r.player("Slammer", 10, 30000, 30000, true)
	p.Body.Yaw = math.Pi / 2
	p.attacking = true
	return r, p, n
}

// swingStyle swings with Slam until a blow lands (or, with parry, is parried).
func swingStyle(r *rig, p *Player, n *NPC, want rules.Result) {
	slam := StyleDef{ID: 164, Class: 22, Name: "Slam", Spec: "Hammer", WeaponReq: ObjHammer}
	for seed := int64(1); seed < 200; seed++ {
		r.z.rng = rand.New(rand.NewSource(seed))
		p.bar, p.pending = []StyleDef{slam}, 0
		hp := n.HP
		r.z.playerSwing(p, n)
		landed := n.HP < hp
		if (want == rules.Hit) == landed {
			return
		}
	}
}

func TestSlamStuns(t *testing.T) {
	r, p, n := procRig(t)
	swingStyle(r, p, n, rules.Hit)
	if !held(n) || !hasEffect(n, 20204) {
		t.Fatal("a landed Slam did not stun")
	}
	n.HP = n.MaxHP
	r.ticks(TickRate * 5)
	if n.speedNow != 0 {
		t.Errorf("a stunned creature moved (speed %v)", n.speedNow)
	}
}

func TestParriedStyleCastsNothing(t *testing.T) {
	r, p, n := procRig(t)
	n.cb.ParryPct = 100
	n.Body.Yaw = -math.Pi / 2 // facing the player, so it can parry
	swingStyle(r, p, n, rules.Parried)
	if held(n) || len(n.effects) > 0 {
		t.Error("a parried Slam stunned")
	}
}

func TestWeaponProcFires(t *testing.T) {
	r, p, n := procRig(t)
	r.w.Cat.StyleSpells = nil
	hammer := *p.weapon
	hammer.ProcSpell, hammer.ProcChance = 61020, 100
	p.weapon = &hammer
	for seed := int64(1); seed < 200; seed++ {
		r.z.rng = rand.New(rand.NewSource(seed))
		before := n.HP
		snaps(p)
		r.z.playerSwing(p, n)
		r.ticks(1)
		var procs int
		for _, s := range snaps(p) {
			for _, e := range s.Ev {
				if e.K == "spell" && e.Text == "Energy DD Proc" {
					procs++
				}
			}
		}
		if n.HP < before && procs > 0 {
			return
		}
		n.HP = n.MaxHP
	}
	t.Error("a 100% weapon proc never fired on a landed blow")
}
