package rules

import (
	"testing"
	"time"
)

const clk = "world-time"

func TestGameClock(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	vector(t, clk, "Two clients agree", func(t *testing.T) {
		// The clock is a function of real time alone, so any two readings in the
		// same real second agree, and a minute of real time is 24 game minutes.
		min := (GameSeconds(base.Add(time.Minute)) - GameSeconds(base) + DayGameSeconds) % DayGameSeconds
		if GameSeconds(base) != GameSeconds(base.Add(500*time.Millisecond)) || min != 24*60 {
			t.Errorf("two readings in one second differ, or a real minute moved the clock %d game seconds, want 1440", min)
		}
	})
	vector(t, clk, "A quarter of an hour later", func(t *testing.T) {
		d := GameSeconds(base.Add(15*time.Minute)) - GameSeconds(base)
		if d < 0 {
			d += DayGameSeconds
		}
		if d != 6*3600 {
			t.Errorf("15 real minutes moved the clock %d game seconds, want 21600 (6 hours)", d)
		}
	})
	vector(t, clk, "Dusk", func(t *testing.T) {
		if IsNight(NightStart-1) || !IsNight(NightStart) || !IsNight(0) || !IsNight(NightEnd-1) || IsNight(NightEnd) || IsNight(12*3600) {
			t.Error("night should be 18:00 up to 06:00")
		}
	})
	t.Run("timed spawns arrive over the first minute and leave together", func(t *testing.T) {
		if Present(NightOnly, 30, NightStart+10) || !Present(NightOnly, 30, NightStart+30) || !Present(NightOnly, 30, 3*3600) || Present(NightOnly, 30, NightEnd) {
			t.Error("a night-only spawn with offset 30 should appear 30 game seconds after dusk and leave at dawn")
		}
		if Present(DayOnly, 0, NightStart) || !Present(DayOnly, 0, 12*3600) || Present(DayOnly, 0, 0) || !Present(Always, 0, 0) {
			t.Error("a day-only spawn should be there by day only")
		}
	})
}
