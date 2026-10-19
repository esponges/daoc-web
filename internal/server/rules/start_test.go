package rules

import "testing"

const sz = "starting-zone-safety"

func TestStartSafety(t *testing.T) {
	vector(t, sz, "An unsafe start", func(t *testing.T) {
		near := []Threat{
			{Name: "frog", Level: 1, Aggressive: true, X: 100},
			{Name: "slither", Level: 26, Aggressive: true, X: 800},
			{Name: "bear", Level: 26, Aggressive: true, X: 600},
		}
		w, d, unsafe := UnsafeStart(0, 0, near)
		if !unsafe || w.Name != "bear" || d != 600 {
			t.Errorf("got %v %v %v, want the bear at 600", w.Name, d, unsafe)
		}
		if i := SafeStart([]int{2, 10, 6}); i != 1 {
			t.Errorf("safe start %d, want the one most rows use", i)
		}
	})
	vector(t, sz, "A safe start is kept", func(t *testing.T) {
		near := []Threat{
			{Name: "frog", Level: 5, Aggressive: true, X: 100},
			{Name: "guardian", Level: 40, Aggressive: false, X: 200},
			{Name: "far bear", Level: 26, Aggressive: true, X: 1501},
		}
		if w, _, unsafe := UnsafeStart(0, 0, near); unsafe {
			t.Errorf("unsafe because of %v", w.Name)
		}
		if i := SafeStart(nil); i != -1 {
			t.Errorf("no safe start: %d, want -1", i)
		}
	})
}
