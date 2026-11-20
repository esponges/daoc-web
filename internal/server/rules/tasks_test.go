package rules

import "testing"

const rt = "rules-tasks"

func TestTasks(t *testing.T) {
	vector(t, rt, "Task limits", func(t *testing.T) {
		if MaxTasks(1) != 1 || MaxTasks(8) != 8 || MaxTasks(20) != 20 || MaxTasks(21) != 0 || MaxTasks(0) != 0 {
			t.Error("a player may finish as many tasks as their level, none above 20")
		}
		lo, hi := KillTargetLevels(8)
		if lo != 6 || hi != 8 {
			t.Errorf("a level 8 player's targets are %d to %d, want 6 to 8", lo, hi)
		}
		if lo, _ := KillTargetLevels(1); lo != 1 {
			t.Errorf("a level 1 player's lowest target is %d", lo)
		}
		if TaskSeconds != 7200 || TaskAskPause != 300 {
			t.Error("a task lasts 2 hours and an ask is retried after 5 minutes")
		}
	})
	vector(t, rt, "Task givers", func(t *testing.T) {
		for name, want := range map[string]bool{
			"Sir Gudrun": true, "Captain Frode": true, "Jarl Brand": true, "Soldier Ulf": true, "Sentinel Hakon": true,
			"Viking Dreng": false, "Viking Huscarl": false, "Viking Aldor": true, "Huntress Sif": true, "Bork Erlend": true,
			"Guard": false, "Guardian": false, "City Guard": false, "Guard Sigrun": true, "Guardian of the Gate": false,
			"Sentinel of Runes": false, "Flayer Kegnar": true, "Flayer Torvald": false, "Hild": false,
		} {
			if got := KillGiver(name, ""); got != want {
				t.Errorf("%q gives kill tasks: %v, want %v", name, got, want)
			}
		}
		if KillGiver("Sir Gudrun", "House") || !KillGiver("Jarl Brand", "House") {
			t.Error("a guilded Sir gives no task; a Jarl does whatever their guild")
		}
		if !DeliveryGiver(true, false, "Ragna") || !DeliveryGiver(false, true, "Baldus") || !DeliveryGiver(false, false, "Town Crier") || DeliveryGiver(false, false, "Hild") {
			t.Error("trainers, merchants and criers give delivery tasks")
		}
	})
	vector(t, rt, "Reward at level 10", func(t *testing.T) {
		// Money for a kill task at level 10 is 632 copper, 3 percent either way (614 to 650).
		if got := TaskMoney(10, false, 0); got != 632-18 {
			t.Errorf("level 10 kill money at the low roll %d, want 614", got)
		}
		if got := TaskMoney(10, false, 0.999); got != 632+18 {
			t.Errorf("level 10 kill money at the high roll %d, want 650", got)
		}
		if got := TaskMoney(10, true, 0.5); got < 379-12 || got > 379+12 {
			t.Errorf("level 10 delivery money %d, want about 379", got)
		}
		want := int64(float64(XPToReach(11)-XPToReach(10)) * 0.5 / 9)
		if got := TaskXP(10); got != want || got <= 0 {
			t.Errorf("level 10 task experience %d, want %d", got, want)
		}
		t.Logf("level 10: %d experience, 614 to 650 copper", TaskXP(10))
	})
	vector(t, rt, "Rewards by level band", func(t *testing.T) {
		prevMoney, prevXP := 0, int64(-1)
		for _, lv := range []int{1, 5, 10, 15, 20} {
			m, x := TaskMoney(lv, false, 0.5), TaskXP(lv)
			if m <= prevMoney || x <= prevXP {
				t.Errorf("level %d pays %d copper and %d experience, not more than the band before (%d, %d)", lv, m, x, prevMoney, prevXP)
			}
			prevMoney, prevXP = m, x
			t.Logf("level %2d: %5d copper (kill), %d experience", lv, m, x)
		}
		if TaskMoney(21, false, 0.5) != 0 || TaskXP(21) != 0 {
			t.Error("a task above level 20 paid")
		}
	})
}
