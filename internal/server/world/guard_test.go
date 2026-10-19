package world

import (
	"math"
	"testing"

	"daocweb/internal/server/rules"
)

// guardAt is a Midgard town guard with no aggro range of its own.
func guardAt(id int64, x, y float64) Spawn {
	return Spawn{ID: id, Name: "Guard Flintrock", Model: 137, X: x, Y: y, Z: 5000, Level: 50, Size: 50, Speed: 191,
		Realm: 2, ClassType: "DOL.GS.GameGuard"}
}

// A creature follows a player into town: the guard attacks it, and the
// creature turns on whoever it hates most.
func TestGuardAttacksACreatureChasingIntoTown(t *testing.T) {
	r := newRig(t, guardAt(1, 30000, 30000), wolf(2, 31600, 30000, 5))
	g, w := r.z.byID[1], r.z.byID[2]
	w.MaxHP, w.HP = 100000, 100000 // it must live long enough to choose
	if !g.Guard || g.guardRange != GuardRange {
		t.Fatalf("guard %v, range %v: want a guard watching %v", g.Guard, g.guardRange, GuardRange)
	}
	p := r.player("Runner", 5, 31300, 30000, true)
	r.ticks(TickRate * 2)
	if w.target != p {
		t.Fatal("the wolf did not attack the player")
	}
	if g.target != nil {
		t.Fatal("the guard attacked a fight 1,300 units away")
	}
	// The player runs into town, the wolf after them.
	p.Body.X = 30400
	for i := 0; i < TickRate*3 && g.target != combatant(w); i++ {
		r.z.Tick()
	}
	if g.target != combatant(w) {
		t.Fatalf("the guard did not take on the wolf chasing a Midgard player (guard target %v)", g.target)
	}
	for i := 0; i < TickRate*10 && !w.hates(g); i++ {
		r.z.Tick()
	}
	if !w.hates(g) {
		t.Fatal("the wolf does not hate the guard that fought it")
	}
	if want := w.hated(r.z); !w.dead() && w.target != want {
		t.Errorf("the wolf fights %v, want the one it hates most (%v)", w.target, want)
	}
	if p.dead {
		t.Error("the player died with a level 50 guard beside them")
	}
}

// An aggressive creature that wanders into a guard's range is attacked even
// with no player about.
func TestGuardAttacksAnAggressiveCreatureInRange(t *testing.T) {
	r := newRig(t, guardAt(1, 30000, 30000), wolf(2, 30500, 30000, 5))
	g, w := r.z.byID[1], r.z.byID[2]
	r.ticks(TickRate)
	if g.target != combatant(w) {
		t.Fatalf("guard target %v, want the aggressive wolf 500 units away", g.target)
	}
}

// Friendly creatures are left alone.
func TestGuardLeavesPassiveCreaturesAlone(t *testing.T) {
	hare := wolf(2, 30200, 30000, 5)
	hare.Name, hare.AggroLevel, hare.AggroRange = "hare", 0, 0
	r := newRig(t, guardAt(1, 30000, 30000), hare)
	r.player("Townsfolk", 5, 30100, 30000, true)
	r.ticks(TickRate * 10)
	if g := r.z.byID[1]; g.target != nil || g.state == npcChase {
		t.Errorf("the guard attacked a passive creature: target %v", g.target)
	}
}

// A guard chases to its leash and then goes back to its post.
func TestGuardLeashesAndReturns(t *testing.T) {
	r := newRig(t, guardAt(1, 30000, 30000), wolf(2, 30500, 30000, 5))
	g, w := r.z.byID[1], r.z.byID[2]
	w.MaxHP, w.HP = 1_000_000, 1_000_000 // it must outlast the chase
	r.ticks(TickRate)
	if g.target != combatant(w) {
		t.Fatal("the guard did not engage")
	}
	// The wolf is dragged away faster than the guard can follow.
	for i := 0; i < TickRate*20 && g.state == npcChase; i++ {
		w.Body.X += 60
		w.Aggro = 0
		r.z.Tick()
	}
	if g.state == npcChase {
		t.Fatal("the guard never gave up the chase")
	}
	r.ticks(TickRate * 30)
	if d := math.Hypot(g.Body.X-g.HX, g.Body.Y-g.HY); d > 50 {
		t.Errorf("the guard is %.0f units from its post, want back at it", d)
	}
}

// A guard does half: a level 10 Warrior who did half the damage to a level 10
// creature gains half its experience, and the guard's half goes to no one.
func TestGuardShareGoesToNoOne(t *testing.T) {
	r := newRig(t, guardAt(1, 30000, 30000), wolf(2, 30100, 30000, 10))
	g, w := r.z.byID[1], r.z.byID[2]
	w.Aggro = 0
	p := r.player("Helper", 10, 30200, 30000, true)
	before := p.Char.XP
	w.hurtBy[p.ID] = 100
	unsharedDamage(g, w, 100)
	r.z.killNPC(w)
	got := p.Char.XP - before
	if got != 2560 || got != rules.KillXP(10, 10, 0.5) {
		t.Errorf("the Warrior gained %d, want 2,560 (half of a level 10 kill)", got)
	}
}
