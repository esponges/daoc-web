package rules

import "testing"

func TestResurrection(t *testing.T) {
	vector(t, dt, "Raised at level 20", func(t *testing.T) {
		start := XPToReach(20)
		xp := start + 100000
		first := DeathLoss(20, xp, 0)
		if first == 0 {
			t.Fatal("the death should have cost something")
		}
		// Released: the same again, a gravestone and constitution.
		loss, con, grave := ReleaseStage(false, first, xp-first, start, 1)
		if loss != min(first, max(0, xp-first-start)) || con != 1 || (loss > 0) != grave {
			t.Errorf("released: lost %d, con %d, grave %v", loss, con, grave)
		}
		// Raised: nothing more.
		loss, con, grave = ReleaseStage(true, first, xp-first, start, 1)
		if loss != 0 || con != 0 || grave {
			t.Errorf("raised: lost %d, con %d, grave %v; want none", loss, con, grave)
		}
	})
	vector(t, dt, "Sickness shows", func(t *testing.T) {
		secs, pct := Sickness(30)
		if secs != 60 || pct != 30 {
			t.Errorf("sickness after a 30 percent raise: %v seconds, %d percent; want 60 and 30", secs, pct)
		}
		if secs, _ := Sickness(100); secs != 30 {
			t.Errorf("a full resurrection leaves %v seconds, want half of 60", secs)
		}
		if e := SicknessFactor(pct); e < 0.699 || e > 0.701 {
			t.Errorf("effectiveness %v, want 0.7", e)
		}
	})
	vector(t, dt, "Release while an offer is pending", func(t *testing.T) {
		if CanRelease(9.9) || !CanRelease(10) {
			t.Error("release opens after 10 seconds")
		}
		if MustRelease(899) || !MustRelease(900) {
			t.Error("release is forced at 900 seconds")
		}
		if OfferLapsed(14.9) || !OfferLapsed(15) {
			t.Error("an offer lapses after 15 seconds")
		}
		// A character who releases while an offer is pending pays the full release.
		start := XPToReach(10)
		loss, _, _ := ReleaseStage(false, 500, start+2000, start, 1)
		if loss != 500 {
			t.Errorf("a release with an offer pending cost %d, want 500", loss)
		}
	})
}
