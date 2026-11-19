package rules

import "testing"

const los = "rules-line-of-sight"

// wall is a vertical wall in the plane x = 500, 2000 wide and 400 high.
func wall() Scenery {
	a, b, c, d := [3]float64{500, -1000, 0}, [3]float64{500, 1000, 0}, [3]float64{500, 1000, 400}, [3]float64{500, -1000, 400}
	return Scenery{Tris: [][3][3]float64{{a, b, c}, {a, c, d}}}
}

func TestLineOfSight(t *testing.T) {
	flat := func(x, y float64) float64 { return 0 }
	me, them := Eye([3]float64{0, 0, 0}), Eye([3]float64{1000, 0, 0})
	vector(t, los, "Open ground", func(t *testing.T) {
		if !Visible(me, them, Scenery{Ground: flat}) || !Visible(them, me, Scenery{Ground: flat}) {
			t.Error("two bodies on flat ground with nothing between them should see each other")
		}
	})
	vector(t, los, "A hill between them", func(t *testing.T) {
		ridge := func(x, y float64) float64 { // a ridge across x = 500, higher than the eyes
			if x > 400 && x < 600 {
				return 300
			}
			return 0
		}
		if Visible(me, them, Scenery{Ground: ridge}) || Visible(them, me, Scenery{Ground: ridge}) {
			t.Error("a ridge above the segment should block it both ways")
		}
		low := func(x, y float64) float64 {
			if x > 400 && x < 600 {
				return 100
			}
			return 0
		}
		if !Visible(me, them, Scenery{Ground: low}) {
			t.Error("a rise below the eyes should not block")
		}
	})
	vector(t, los, "A dungeon wall between them", func(t *testing.T) {
		if Visible(me, them, wall()) || Visible(them, me, wall()) {
			t.Error("a wall between them should block both ways")
		}
		beside := Eye([3]float64{0, 3000, 0}) // the wall ends at y = 1000
		if !Visible(me, beside, wall()) {
			t.Error("a segment that misses the wall should pass")
		}
		if !Visible(me, Eye([3]float64{499, 0, 0}), wall()) {
			t.Error("a body this side of the wall should see up to it")
		}
	})
	vector(t, los, "A tree trunk", func(t *testing.T) {
		trunk := Scenery{Cyls: []Cylinder{{X: 500, Y: 0, R: 40, ZMin: 0, ZMax: 600}}}
		if Visible(me, them, trunk) {
			t.Error("a segment through a trunk should be blocked")
		}
		if !Visible(me, Eye([3]float64{1000, 200, 0}), trunk) {
			t.Error("a segment that passes beside the trunk should pass")
		}
		if !Visible(Eye([3]float64{0, 0, 700}), Eye([3]float64{1000, 0, 700}), trunk) {
			t.Error("a segment over the top of the trunk should pass")
		}
	})
	vector(t, los, "Casting at a creature behind a wall", func(t *testing.T) {
		if !NeedsSight(false, false, false) {
			t.Fatal("a bolt at a creature needs sight")
		}
		if Visible(me, them, wall()) {
			t.Error("the creature behind the wall should be unseen, so the cast is refused")
		}
		if SightMessage != "You can't see your target!" {
			t.Errorf("the message is %q", SightMessage)
		}
	})
	vector(t, los, "The target steps behind cover during the cast", func(t *testing.T) {
		if !Visible(me, them, Scenery{}) {
			t.Fatal("the cast should begin in the open")
		}
		if Visible(me, them, wall()) {
			t.Error("with the target behind a wall at landing, the spell fails")
		}
	})
	vector(t, los, "Area spell around the caster", func(t *testing.T) {
		if NeedsSight(false, false, true) || NeedsSight(true, false, false) || NeedsSight(false, true, false) {
			t.Error("a spell on the caster, the group or around the caster needs no sight to its centre")
		}
		open, behind := Eye([3]float64{-300, 0, 0}), Eye([3]float64{800, 0, 0})
		if !Visible(me, open, wall()) {
			t.Error("the creature in the open is hit")
		}
		if Visible(me, behind, wall()) {
			t.Error("the creature behind the wall is not")
		}
	})
	vector(t, los, "Caster creature behind a corner", func(t *testing.T) {
		bolt := CastOption{Ready: true, Range: 1500, Blind: true}
		if i := ChooseCast([]CastOption{bolt}, 800, false); i != -1 {
			t.Errorf("a creature that cannot see its target chose %d, want to close in", i)
		}
		bolt.Blind = false
		if i := ChooseCast([]CastOption{bolt}, 800, false); i != 0 {
			t.Errorf("once it sees its target it chose %d, want the spell", i)
		}
	})
	vector(t, los, "A zone without collision data", func(t *testing.T) {
		if !Visible(me, them, Scenery{}) {
			t.Error("with no scenery every target is visible")
		}
	})
}
