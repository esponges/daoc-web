package world

import (
	"testing"
)

// A creature fighting two players turns to the one who has hurt it most.
func TestCreatureTurnsToWhoHurtItMost(t *testing.T) {
	r := newRig(t, wolf(1, 30000, 30000, 30))
	n := r.z.npcs[0]
	n.Aggro = 0
	a := r.player("Opener", 20, 29900, 30000, true)
	b := r.player("Hitter", 20, 30100, 30000, true)
	r.z.provoke(n, a, 10)
	r.ticks(1)
	if n.target != a {
		t.Fatalf("target %v, want the first to strike", n.target)
	}
	r.z.provoke(n, b, 30)
	r.ticks(1)
	if n.target != b {
		t.Errorf("target %v, want the one who did more damage", n.target)
	}
	// An even total keeps the one it fought first.
	r.z.provoke(n, a, 20)
	r.ticks(1)
	if n.target != a {
		t.Errorf("target %v on a tie, want the first", n.target)
	}
}

// When the player it fights dies, a creature turns on the next it hates
// instead of going home; with nobody left, it goes home.
func TestCreatureTurnsToTheNextWhenOneDies(t *testing.T) {
	r := newRig(t, wolf(1, 30000, 30000, 30))
	n := r.z.npcs[0]
	n.Aggro = 0
	a := r.player("Tank", 20, 29900, 30000, true)
	b := r.player("Helper", 20, 30100, 30000, true)
	r.z.provoke(n, a, 50)
	r.z.provoke(n, b, 5)
	r.ticks(1)
	r.z.killPlayer(a, n)
	r.ticks(1)
	if n.target != b || n.state != npcChase {
		t.Fatalf("after the tank died: target %v state %v, want the helper", n.target, n.state)
	}
	r.z.killPlayer(b, n)
	r.ticks(1)
	if n.target != nil || n.state == npcChase || len(n.hate) > 0 {
		t.Errorf("with nobody left: target %v state %v, want it calm and going home", n.target, n.state)
	}
}

// Creatures can fight each other: the one struck hates the striker, fights
// back, and dies to it.
func TestCreaturesFightEachOther(t *testing.T) {
	r := newRig(t, wolf(1, 30000, 30000, 40), wolf(2, 30100, 30000, 2))
	strong, weak := r.z.npcs[0], r.z.npcs[1]
	if strong.Level != 40 {
		strong, weak = weak, strong
	}
	strong.Name, weak.Name = "bear", "hare" // not packmates
	strong.engage(weak, r.z)
	for i := 0; i < TickRate*60 && !weak.dead(); i++ {
		r.z.Tick()
	}
	if !weak.dead() {
		t.Fatal("a level 40 creature did not kill a level 2 one in a minute")
	}
	r.ticks(1)
	if strong.target == weak || strong.hates(weak) {
		t.Error("the dead creature is still fought")
	}
}
