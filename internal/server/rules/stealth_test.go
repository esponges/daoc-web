package rules

import (
	"math"
	"testing"
)

const st = "rules-stealth"

func TestStealth(t *testing.T) {
	vector(t, st, "Stealth speed", func(t *testing.T) {
		near(t, "level 20, Stealth 20", StealthSpeed(20, 20), 0.6, 1e-9)
		near(t, "level 50, Stealth 50", StealthSpeed(50, 50), 0.6, 1e-9)
		near(t, "level 20, Stealth 10", StealthSpeed(10, 20), 0.3+20*0.3/30, 1e-9)
		near(t, "spec above level counts as level", StealthSpeed(40, 20), 0.6, 1e-9)
		if s := StealthSpeed(1, 1); s <= 0.3 || s >= 1 || math.IsNaN(s) {
			t.Errorf("level 1: %v", s)
		}
	})
	vector(t, st, "Detect Hidden", func(t *testing.T) {
		// A level 30 viewer against a hider with Stealth 20.
		near(t, "plain", PlayerDetectRange(30, 20, false, false), 10*20+125, 1e-9)
		near(t, "Detect Hidden", PlayerDetectRange(30, 20, true, false), 10*50+250, 1e-9)
		near(t, "the hider has Camouflage", PlayerDetectRange(30, 20, true, true), 10*20+125, 1e-9)
		near(t, "the hider outlevels the viewer", PlayerDetectRange(10, 30, false, false), 125, 1e-9)
		near(t, "spec counts to 50", PlayerDetectRange(50, 60, false, false), 125, 1e-9)
	})
	vector(t, st, "Creature detection", func(t *testing.T) {
		near(t, "level 10 creature, Stealth 5", CreatureDetectRadius(10, 5, false, false), 125+5*20, 1e-9)
		near(t, "a creature far below the hider", CreatureDetectRadius(1, 50, false, false), 126, 1e-9)
		near(t, "a creature that detects the hidden", CreatureDetectRadius(10, 5, true, false), 125+5*20+125, 1e-9)
		near(t, "countered by Camouflage", CreatureDetectRadius(10, 5, true, true), 125+5*20, 1e-9)
	})
	vector(t, st, "Too close to hide", func(t *testing.T) {
		near(t, "equal level and spec", StealthTooClose(20, 20, false), 1024-896, 1e-9)
		near(t, "a higher level enemy", StealthTooClose(40, 10, false), 1024-896*10/40.0, 1e-9)
		if StealthTooClose(10, 50, false) > 0 {
			t.Error("a hider far above the enemy has no limit")
		}
		near(t, "a detecting enemy", StealthTooClose(40, 10, true), 2048-1792*10/40.0, 1e-9)
	})
}
