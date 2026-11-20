package world

import (
	"math/rand"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// stealthRig is a zone with a wolf and a hider of a class that has the Stealth
// spec, a Hunter, at 30000,30000.
func stealthRig(t *testing.T, wolfAt float64) (*rig, *Player, *NPC) {
	r := newRig(t, wolf(1, wolfAt, 30000, 8))
	r.z.rng = rand.New(rand.NewSource(4))
	r.w.Cat.ClassSpecs[rules.ClassHunter] = []string{"Stealth"}
	w := r.z.npcs[0]
	w.Aggro, w.Roam = 0, 0
	p := r.player("Hider", 20, 30000, 30000, true)
	p.Char.Class = rules.ClassHunter
	p.Char.Specs = map[string]int{"Stealth": 10}
	r.w.recalc(p)
	r.ticks(TickRate) // the clock is past zero
	sentOf[protocol.Effects](p, protocol.TEffects)
	return r, p, w
}

func TestHideOutOfCombat(t *testing.T) {
	r, p, _ := stealthRig(t, 36000)
	r.z.Stealth(p)
	if !p.hidden() {
		t.Fatalf("not hidden: %v", errsOf(p))
	}
	m := sentOf[protocol.Effects](p, protocol.TEffects)
	if len(m) == 0 || m[len(m)-1].Speed >= 1 || m[len(m)-1].Speed != rules.StealthSpeed(10, 20) {
		t.Errorf("the effects message did not carry the stealth speed %v: %+v", rules.StealthSpeed(10, 20), m)
	}
	r.z.Stealth(p) // within the 2 second gap
	if !p.hidden() {
		t.Error("stealth ended inside the change gap")
	}
	r.ticks(TickRate * 3)
	r.z.Stealth(p)
	if p.hidden() {
		t.Error("stealth did not end when asked")
	}
}

func TestHideInCombatIsRefused(t *testing.T) {
	r, p, _ := stealthRig(t, 36000)
	p.combatUntil = r.z.now + 5
	r.z.Stealth(p)
	if p.hidden() || !strings.Contains(lastErr(t, p), "can't hide while in combat") {
		t.Errorf("hidden %v, error %q", p.hidden(), lastErr(t, p))
	}
}

func TestACreatureTooCloseStopsYouHiding(t *testing.T) {
	r, p, w := stealthRig(t, 30100)
	w.setLevel(30)
	r.z.Stealth(p)
	if p.hidden() || !strings.Contains(lastErr(t, p), "too close") {
		t.Errorf("hidden %v, error %q", p.hidden(), lastErr(t, p))
	}
}

func TestStealthEndsOnSwingDamageAndCast(t *testing.T) {
	for name, act := range map[string]func(r *rig, p *Player, w *NPC){
		"a swing": func(r *rig, p *Player, w *NPC) { w.Body.X = 30060; r.z.playerSwing(p, w) },
		"damage":  func(r *rig, p *Player, w *NPC) { r.z.hurt(p, w, 3) },
	} {
		r, p, w := stealthRig(t, 36000)
		r.z.Stealth(p)
		if !p.hidden() {
			t.Fatalf("%s: not hidden", name)
		}
		act(r, p, w)
		if p.hidden() {
			t.Errorf("%s did not end stealth", name)
		}
	}
}

func hasEntity(p *Player, id int64) bool {
	for _, s := range snaps(p) {
		for _, e := range s.Ents {
			if e.ID == id {
				return true
			}
		}
	}
	return false
}

func TestOthersSeeAHiderOnlyInsideTheirDetectionRange(t *testing.T) {
	r, h, _ := stealthRig(t, 36000)
	// A level 30 viewer, hider's Stealth 10: range (30-10)*20+125 = 525.
	far := r.player("Far", 30, 30900, 30000, true)
	near := r.player("Near", 30, 30400, 30000, true)
	r.z.Stealth(h)
	r.ticks(3)
	if hasEntity(far, h.ID) {
		t.Error("a viewer beyond the range was sent the hider")
	}
	if !hasEntity(near, h.ID) {
		t.Error("a viewer inside the range was not sent the hider")
	}
	// Detect Hidden: (30-10)*50+250 = 1250 reaches the far viewer.
	far.sh.abilities["Detect Hidden"] = 1
	r.ticks(3)
	if !hasEntity(far, h.ID) {
		t.Error("Detect Hidden did not widen the range")
	}
	r.z.Stealth(h) // too soon
	r.ticks(TickRate * 3)
	r.z.Stealth(h)
	far.sh.abilities["Detect Hidden"] = 0
	r.ticks(3)
	if !hasEntity(far, h.ID) {
		t.Error("an unhidden player was not sent")
	}
}

func TestAnAggressiveCreatureNoticesAHiderOnlyInsideItsRadius(t *testing.T) {
	// A level 18 wolf and Stealth 10: radius 125 + (18-10)*20 = 285.
	r, p, w := stealthRig(t, 31200)
	w.setLevel(18)
	w.Aggro = 700
	r.z.Stealth(p)
	if !p.hidden() {
		t.Fatalf("not hidden: %v", errsOf(p))
	}
	w.Body.X = 30500
	r.ticks(TickRate * 3)
	if w.target != nil || !p.hidden() {
		t.Fatalf("the wolf noticed a hider 500 away (inside its aggro range, outside its detection)")
	}
	w.Body.X = 30250 // inside 285
	r.ticks(TickRate * 2)
	if w.target == nil {
		t.Error("the wolf did not notice a hider 250 away")
	}
	if p.hidden() {
		t.Error("the hider stayed hidden after being noticed")
	}
}

func TestStealthIsNotKeptByTheSpeedSong(t *testing.T) {
	r, p, _ := stealthRig(t, 36000)
	p.effects = append(p.effects, &effect{spell: &SpellDef{Name: "Speed", Value: 144}, kind: effSpeed, ends: 1e9})
	r.z.Stealth(p)
	if hasKind(p, effSpeed) {
		t.Error("a speed effect survived hiding")
	}
}

func TestAStealthOpenerNeedsStealthAndEndsIt(t *testing.T) {
	r, p, w := stealthRig(t, 30060)
	w.setLevel(1) // close, but a level 1 creature allows hiding
	w.Body.X = 36000
	backstab := StyleDef{ID: 777, Class: rules.ClassHunter, Name: "Backstab", Spec: "Stealth", SpecLevel: 1, Endurance: 5, Stealth: true, GrowthRate: 0.5}
	p.end = 50
	if ok, why := r.z.styleReady(p, &backstab); ok || !strings.Contains(why, "hidden") {
		t.Fatalf("an opener was ready out of stealth: %v %q", ok, why)
	}
	r.z.Stealth(p)
	if ok, why := r.z.styleReady(p, &backstab); !ok {
		t.Fatalf("an opener was not ready in stealth: %q", why)
	}
	w.Body.X = 30060
	p.bar = []StyleDef{backstab}
	p.pending = 0
	p.target = w.ID
	p.Char.Specs["Stealth"] = 10
	for i := 0; i < 30 && p.hidden(); i++ { // the first swing may miss, but it ends stealth either way
		p.swingAt = 0
		r.z.playerSwing(p, w)
		p.pending = 0
	}
	if p.hidden() {
		t.Error("a swing from stealth did not end it")
	}
}
