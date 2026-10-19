package rules

import (
	"fmt"
	"testing"
)

func TestGroupExperience(t *testing.T) {
	vector(t, xp, "A group and a solo player", func(t *testing.T) {
		award := GroupAward(10, 0.75)
		parts := GroupKillXP(10, 0.75, []int{10, 10})
		if award != 3840 || parts[0]+parts[1] != 3840 || KillXP(10, 10, 0.25) != 1280 {
			t.Errorf("group award %v split %v, solo %d; want 3,840 split and 1,280", award, parts, KillXP(10, 10, 0.25))
		}
	})
	vector(t, xp, "Two even members", func(t *testing.T) {
		if got := fmt.Sprint(GroupKillXP(10, 1, []int{10, 10})); got != "[2560 2560]" {
			t.Errorf("%s, want 2,560 each", got)
		}
	})
	vector(t, xp, "Split by level, then capped", func(t *testing.T) {
		if got := fmt.Sprint(GroupKillXP(10, 1, []int{10, 5})); got != "[3413 200]" {
			t.Errorf("%s, want 3,413 and the level 5's 1,706 capped at 200", got)
		}
	})
	vector(t, xp, "A member far away", func(t *testing.T) {
		// The far member is left out by the caller: the two in range split as two.
		if got := fmt.Sprint(GroupKillXP(10, 1, []int{10, 10})); got != "[2560 2560]" {
			t.Errorf("%s, want the two in range to split as a group of two", got)
		}
	})
	vector(t, xp, "A high-level helper", func(t *testing.T) {
		if got := fmt.Sprint(GroupKillXP(10, 1, []int{20, 10})); got != "[0 0]" {
			t.Errorf("%s, want nothing: the creature is grey to the level 20", got)
		}
	})
}
