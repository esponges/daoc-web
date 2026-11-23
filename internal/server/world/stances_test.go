package world

import (
	"math"
	"strings"
	"testing"

	"daocweb/internal/server/rules"
)

// stanceRig is a zone with a player who holds the timed abilities at level 3.
func stanceRig(t *testing.T) (*rig, *Player, *NPC) {
	r := newRig(t, wolf(1, 30600, 30000, 8))
	n := r.z.npcs[0]
	n.Aggro, n.Roam = 0, 0
	p := r.player("Bjorn", 20, 30000, 30000, true)
	for _, k := range []string{rules.AbBerserk, rules.AbCharge, rules.AbQuickCast, rules.AbDistraction} {
		p.sh.abilities[k] = 3
	}
	return r, p, n
}

func TestBerserkRaisesDamageAndLowersDefence(t *testing.T) {
	r, p, _ := stanceRig(t)
	before := p.combatant(r.w.Cat)
	r.z.UseAbility(p, "berserk")
	if !hasKind(p, effBerserk) {
		t.Fatal("Berserk left no effect")
	}
	cb := p.combatant(r.w.Cat)
	near := func(got, want float64) bool { return math.Abs(got-want) < 1e-9 }
	if !near(cb.DamageBonus-before.DamageBonus, rules.BerserkDamage(3)) || cb.DefencePenalty != rules.BerserkDefenceLoss {
		t.Errorf("damage bonus %v, defence penalty %v after Berserk 3", cb.DamageBonus, cb.DefencePenalty)
	}
	// Used again at once, it is refused with the time left.
	lines(p) // drain
	r.z.UseAbility(p, "berserk")
	if !sawLine(p, "not ready for another") {
		t.Error("a second Berserk was not refused")
	}
	// After it ends the reuse timer still holds.
	r.ticks(int((rules.BerserkSeconds + 1) * TickRate))
	lines(p)
	if hasKind(p, effBerserk) {
		t.Error("Berserk did not end")
	}
	lines(p) // the snapshots of 31 seconds fill the queue; clear it
	r.z.UseAbility(p, "berserk")
	if hasKind(p, effBerserk) || !sawLine(p, "not ready for another") {
		t.Error("Berserk was used again before its reuse time")
	}
}

func TestChargeIgnoresSlows(t *testing.T) {
	r, p, _ := stanceRig(t)
	slow := &SpellDef{ID: 9001, Name: "Slow", Type: rules.SpellSpeedDecrease, Target: "Enemy", Value: 50, Duration: 60}
	r.z.addEffect(p, p, slow, effSlow)
	if f := speedFactor(p); f >= 1 {
		t.Fatalf("a slow left speed at %v", f)
	}
	r.z.UseAbility(p, "charge")
	if f := speedFactor(p); math.Abs(f-rules.ChargeFactor) > 1e-9 {
		t.Errorf("speed during a Charge %v, want %v", f, rules.ChargeFactor)
	}
	r.ticks(int((rules.ChargeSeconds + 1) * TickRate))
	if f := speedFactor(p); f >= 1 {
		t.Errorf("after the Charge the slow should hold again, speed %v", f)
	}
}

func TestQuickCastMakesTheNextSpellInstant(t *testing.T) {
	r, p, _ := stanceRig(t)
	sp := &SpellDef{ID: 9100, Name: "Slow Bolt", Type: rules.SpellHeal, Target: "Self", CastTime: 3, Value: 10}
	if got := r.z.castSeconds(p, sp); got != 3 {
		t.Fatalf("cast time without QuickCast %v", got)
	}
	r.z.UseAbility(p, "quickcast")
	if got := r.z.castSeconds(p, sp); got != 0 {
		t.Errorf("cast time under QuickCast %v", got)
	}
	r.z.spendQuickCast(p)
	if got := r.z.castSeconds(p, sp); got != 3 {
		t.Errorf("the spell after it takes %v, want 3", got)
	}
}

func TestGroundTargetAndDistraction(t *testing.T) {
	r, p, n := stanceRig(t)
	r.z.UseAbility(p, "distraction")
	if !sawLine(p, "You must have a ground target") {
		t.Error("Distraction with no ground target was not refused")
	}
	r.z.SetGroundTarget(p, p.Body.X+2500, p.Body.Y, 5000, false)
	if p.groundPoint() != nil {
		t.Error("a ground target 2500 units away was kept")
	}
	r.z.SetGroundTarget(p, p.Body.X+1500, p.Body.Y, 5000, false)
	if p.groundPoint() == nil {
		t.Fatal("a ground target 1500 units away was refused")
	}
	if gt := p.groundMsg(); len(gt) != 3 || gt[0] != math.Round(p.Body.X+1500) {
		t.Errorf("the ground target as sent: %v", gt)
	}
	n.Body.X, n.Body.Y, n.Body.Yaw = p.Body.X+1400, p.Body.Y+100, 3
	before := n.Body.Yaw
	r.z.UseAbility(p, "distraction")
	if n.Body.Yaw == before {
		t.Error("an idle creature near the ground target did not turn")
	}
	// Another zone forgets it.
	p.Zone = nil
	if p.groundPoint() != nil {
		t.Error("the ground target survived a zone change")
	}
}

func sawLine(p *Player, sub string) bool {
	for _, m := range lines(p) {
		if strings.Contains(m, sub) {
			return true
		}
	}
	return false
}

func TestDangerSenseWarnsOfHiddenEnemies(t *testing.T) {
	r, p, _ := stanceRig(t)
	p.sh.abilities[rules.AbDangerSense] = 1
	o := r.player("Lurker", 20, p.Body.X+800, p.Body.Y, true)
	o.Arch.Realm = p.Arch.Realm + 1
	o.effects = append(o.effects, &effect{spell: &SpellDef{Name: "Stealth", Type: "Stealth"}, kind: effStealth, ends: math.Inf(1)})
	lines(p)
	r.ticks(2)
	if !sawLine(p, "You sense danger nearby!") {
		t.Fatal("a hidden enemy 800 units away was not sensed")
	}
	lines(p)
	r.ticks(5)
	if sawLine(p, "You sense danger nearby!") {
		t.Error("the same enemy was announced again at once")
	}
	o.Body.X = p.Body.X + 5000
	lines(p)
	r.ticks(2)
	if sawLine(p, "You sense danger nearby!") {
		t.Error("an enemy 5000 units away was sensed")
	}
}
