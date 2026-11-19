package world

import (
	"math"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// sightScene is a collision grid around the test rig's 30,000 area with a
// tall wall across x = 30500 (y from 29000 to 31000), and a trunk if asked.
func sightScene(trunk bool) *Collision {
	b := &builder{}
	for _, t := range vwall(30500, 29000, 31000, -50000, 50000) {
		b.add(t, false)
	}
	if trunk {
		b.cyls = append(b.cyls, cyl{x: 30200, y: 33000, r: 40, zmin: -50000, zmax: 50000})
	}
	return b.build(25000, 25000, 36000, 36000)
}

func TestCollisionSees(t *testing.T) {
	c := sightScene(true)
	eye := func(x, y float64) [3]float64 { return [3]float64{x, y, 5150} }
	if c.Sees(eye(30000, 30000), eye(31000, 30000)) {
		t.Error("the wall between them should block")
	}
	if !c.Sees(eye(30000, 30000), eye(30400, 30000)) {
		t.Error("this side of the wall they should see each other")
	}
	if !c.Sees(eye(30000, 28000), eye(31000, 28000)) {
		t.Error("where the wall does not stand they should see each other")
	}
	if c.Sees(eye(30000, 33000), eye(30500, 33000)) {
		t.Error("a trunk on the line should block")
	}
	if !c.Sees(eye(30000, 33200), eye(30500, 33200)) {
		t.Error("a line beside the trunk should pass")
	}
	var none *Collision
	if !none.Sees(eye(30000, 30000), eye(31000, 30000)) {
		t.Error("a zone with no collision data sees everything")
	}
	if !c.Sees(eye(100, 100), eye(200, 100)) {
		t.Error("outside the grid there is nothing to block")
	}
}

func TestPlayerSpellsNeedLineOfSight(t *testing.T) {
	t.Run("A bolt behind a wall is refused", func(t *testing.T) {
		r := casterWorld(t)
		r.z.Mv = sightScene(false)
		m := caster(r, "Rune", rules.ClassRunemaster, 6)
		n := r.z.npcs[0] // 1,000 units east, behind the wall at x = 30500
		m.target = n.ID
		power := m.power
		r.z.Cast(m, 5001, 0)
		if errs := errsOf(m); len(errs) == 0 || !strings.Contains(errs[0], "can't see your target") || m.casting != nil {
			t.Errorf("errors %v, casting %v", errs, m.casting)
		}
		if m.power != power {
			t.Error("a refused cast spent power")
		}
	})
	t.Run("The same bolt in the open", func(t *testing.T) {
		r := casterWorld(t)
		r.z.Mv = sightScene(false)
		m := caster(r, "Rune", rules.ClassRunemaster, 6)
		n := r.z.npcs[0]
		n.Body.X, n.Body.Y = 30300, 30000 // this side of the wall
		m.target = n.ID
		r.z.Cast(m, 5001, 0)
		if m.casting == nil {
			t.Errorf("the target is in plain sight and the cast was refused: %v", errsOf(m))
		}
	})
	t.Run("The target steps behind cover during the cast", func(t *testing.T) {
		r := casterWorld(t)
		r.z.Mv = sightScene(false)
		m := caster(r, "Rune", rules.ClassRunemaster, 6)
		n := r.z.npcs[0]
		n.Body.X, n.Body.Y = 30300, 30000
		m.target = n.ID
		hp, power := n.HP, m.power
		r.z.Cast(m, 5001, 0)
		if m.casting == nil {
			t.Fatal("the cast did not begin")
		}
		n.Body.X = 31000 // behind the wall
		r.ticks(2*TickRate + 5)
		if n.HP != hp {
			t.Errorf("the bolt landed through the wall: %d of %d", n.HP, hp)
		}
		if m.power != power {
			t.Error("a failed cast spent power")
		}
		found := false
		for _, s := range snaps(m) {
			for _, e := range s.Ev {
				if e.K == "castfail" {
					found = true
				}
			}
		}
		if !found {
			t.Error("the player was not shown the cast fail")
		}
	})
	t.Run("A spell on yourself or the group needs no sight", func(t *testing.T) {
		r := casterWorld(t)
		r.z.Mv = sightScene(false)
		h := caster(r, "Healer", rules.ClassHealer, 6)
		r.z.Cast(h, 6001, 0) // Might, on the caster
		if h.casting == nil && len(errsOf(h)) > 0 {
			t.Errorf("a self buff was refused: %v", errsOf(h))
		}
	})
	t.Run("Thunder Shout skips a creature behind the wall", func(t *testing.T) {
		r := areaWorld(t) // wolves at 200 east, 200 north and 150 west of the caster, and one at 400
		b := &builder{}   // a short wall at x = 29950 hides the wolf to the west
		for _, tr := range vwall(29950, 29900, 30100, -50000, 50000) {
			b.add(tr, false)
		}
		r.z.Mv = b.build(25000, 25000, 36000, 36000)
		p := caster(r, "Thane", rules.ClassThane, 5)
		full := [4]int{}
		for i, n := range r.z.npcs {
			full[i] = n.HP
		}
		r.z.Cast(p, 3521, 0)
		for i := 0; i < 2; i++ {
			if r.z.npcs[i].HP >= full[i] {
				t.Errorf("wolf %d in the open inside the radius was not hit", r.z.npcs[i].ID)
			}
		}
		if hidden := r.z.npcs[2]; hidden.HP != full[2] {
			t.Errorf("the wolf behind the wall was hit: %d of %d", hidden.HP, full[2])
		}
	})
	t.Run("A zone with no collision data casts as before", func(t *testing.T) {
		r := casterWorld(t)
		m := caster(r, "Rune", rules.ClassRunemaster, 6)
		m.target = r.z.npcs[0].ID
		r.z.Cast(m, 5001, 0)
		if m.casting == nil {
			t.Errorf("no collision data and the cast was refused: %v", errsOf(m))
		}
	})
}

func TestCreatureWithoutSightClosesIn(t *testing.T) {
	r, n, p := casterRig(t, bolt)
	r.z.Mv = sightScene(false) // the wall at x = 30500 is between them
	n.engage(p, r.z)
	d0 := math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y)
	for i := 0; i < TickRate*2; i++ {
		r.z.Tick()
	}
	if n.casting != nil {
		t.Fatal("a creature that cannot see its target began a cast")
	}
	for _, e := range []string{"cast"} {
		if castEvents(p)[e] > 0 {
			t.Errorf("a %s event was sent through the wall", e)
		}
	}
	if d := math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y); d >= d0 {
		t.Errorf("it stayed %v away instead of closing in (was %v)", d, d0)
	}
	_ = protocol.Event{}
}
