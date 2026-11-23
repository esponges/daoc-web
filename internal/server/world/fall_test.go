package world

import (
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// pond is flat ground under two hundred units of water everywhere.
type pond struct{ flat }

func (pond) WaterAt(x, y float64) (float64, bool) { return 5200, true }

func idleStep(p *Player) {
	p.QueueInput(&protocol.Input{Steps: [][5]int{{1, 0, 0, 0, 0}}})
}

// A client that goes on sending steps while it falls: the fall costs health
// by the rule, from the height it started.
func TestAFallHurtsByTheRule(t *testing.T) {
	r := newRig(t)
	p := r.player("Faller", 20, 30000, 30000, false)
	start := p.hp
	p.Body.Z = 5000 + 960 // twice the safe height up
	for i := 0; i < 400 && !(p.Body.Z == 5000 && i > 2); i++ {
		idleStep(p)
		r.ticks(1)
	}
	want := rules.FallDamage(960, p.maxHP(), 0, false)
	if want == 0 {
		t.Fatal("the rule gives no damage for the test fall")
	}
	if got := int(start - p.hp); got != want {
		t.Errorf("the fall cost %d, want %d", got, want)
	}
	if p.Body.Air || p.Body.Z != 5000 {
		t.Errorf("body air %v z %v after the fall, want landed at 5000", p.Body.Air, p.Body.Z)
	}
}

func TestAShortDropAndAJumpCostNothing(t *testing.T) {
	r := newRig(t)
	p := r.player("Hopper", 20, 30000, 30000, false)
	start := p.hp
	p.QueueInput(&protocol.Input{Steps: [][5]int{{1, 0, 0, 0, protocol.FlagJump}}})
	r.ticks(1)
	if !p.Body.Air || !p.Body.Jumped {
		t.Fatalf("a jump did not leave the ground: air %v jumped %v", p.Body.Air, p.Body.Jumped)
	}
	for i := 0; i < 60 && p.Body.Air; i++ {
		idleStep(p)
		r.ticks(1)
	}
	if p.Body.Air || p.hp != start {
		t.Errorf("after a jump air %v, health %v -> %v", p.Body.Air, start, p.hp)
	}
}

func TestAFallCanKill(t *testing.T) {
	r := newRig(t)
	p := r.player("Unlucky", 20, 30000, 30000, false)
	p.hp = 10
	p.Body.Z = 5000 + rules.SafeFallHeight + rules.FallLethalSpan
	for i := 0; i < 600 && !p.dead; i++ {
		idleStep(p)
		r.ticks(1)
	}
	if !p.dead || p.hp != 0 {
		t.Errorf("dead %v hp %v after a lethal fall", p.dead, p.hp)
	}
	if p.lastLoss != 0 {
		t.Errorf("a fall cost %d experience; only a creature's kill does", p.lastLoss)
	}
}

// A client that stops sending while its body is in the air must not hang
// there: the server brings it down.
func TestABodyLeftInTheAirComesDown(t *testing.T) {
	r := newRig(t)
	p := r.player("Dropped", 20, 30000, 30000, false)
	p.Body.Z = 5000 + 200
	idleStep(p)
	r.ticks(1)
	if !p.Body.Air {
		t.Fatal("the body did not leave the ground")
	}
	r.ticks(int(60 / protocol.StepsPerSecond * 20)) // a second of ticks
	r.ticks(60)
	if p.Body.Air || p.Body.Z != 5000 {
		t.Errorf("the body is still up: air %v z %v", p.Body.Air, p.Body.Z)
	}
}

func TestAFallIntoWaterIsSoftAndASwimmerFloats(t *testing.T) {
	r := newRig(t)
	r.z.G = pond{}
	p := r.player("Diver", 20, 30000, 30000, false)
	start := p.hp
	p.Body.Z = 5000 + 3000
	for i := 0; i < 600 && !(i > 3 && !p.Body.Air); i++ {
		idleStep(p)
		r.ticks(1)
	}
	if p.hp != start {
		t.Errorf("a fall into water cost %v", start-p.hp)
	}
	if want := 5200 - rules.SwimDepth; p.Body.Z != want {
		t.Errorf("the swimmer floats at %v, want %v", p.Body.Z, want)
	}
}
