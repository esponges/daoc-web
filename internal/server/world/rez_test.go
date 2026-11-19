package world

import (
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// release lets a dead player's ten seconds pass and releases them.
func (r *rig) release(p *Player) {
	p.deadAt = r.z.now - rules.ReleaseAfter
	r.z.ReleasePlayer(p)
}

var raiseFallen = &SpellDef{ID: 7001, Name: "Raise Fallen", Type: rules.SpellResurrect, Target: "Corpse", Range: 1500, CastTime: 4, ResurrectHealth: 30, ResurrectMana: 10, Power: 10}

// rezWorld is a Healer standing 300 units from a level 10 Warrior who has just
// been killed by a wolf.
func rezWorld(t *testing.T) (*rig, *Player, *Player) {
	r := casterWorld(t)
	r.w.Cat.Spells[raiseFallen.ID] = raiseFallen
	r.w.Cat.LineSpells["Mending"] = append(r.w.Cat.LineSpells["Mending"], LineSpell{Spec: "Mending", Line: "Mending", Spell: 7001, Level: 1})
	h := caster(r, "Healer", rules.ClassHealer, 10)
	w := r.player("Fallen", 10, 30300, 30000, true)
	w.Char.XP = 300000
	r.w.recalc(w)
	r.z.killPlayer(w, r.z.npcs[0])
	return r, h, w
}

func offered(p *Player) []protocol.RezOffer { return sentOf[protocol.RezOffer](p, protocol.TRezOffer) }

// tickCollect runs n ticks, taking the player's rez offers as they come: a
// real connection drains its queue all the time, a test's would fill.
func (r *rig) tickCollect(p *Player, n int) (out []protocol.RezOffer) {
	for i := 0; i < n; i += TickRate / 2 {
		r.ticks(min(TickRate/2, n-i))
		out = append(out, offered(p)...)
	}
	return out
}

func TestResurrectionRaisesTheDead(t *testing.T) {
	r, h, w := rezWorld(t)
	xp, con := w.Char.XP, w.Char.ConLost
	power := h.power
	r.z.Cast(h, 7001, w.ID)
	if h.casting == nil {
		t.Fatalf("the cast did not begin: %v", errsOf(h))
	}
	r.ticks(4*TickRate + 5)
	if w.rez == nil {
		t.Fatal("no offer reached the dead player")
	}
	if offs := offered(w); len(offs) == 0 || offs[0].Health != 30 || offs[0].From != "Healer" || offs[0].Secs != rules.RezOfferSeconds {
		t.Errorf("the offer sent was %+v", offs)
	}
	if h.power >= power {
		t.Error("the cast cost the Healer no power")
	}
	r.z.AnswerResurrection(w, true)
	if w.dead || w.rez != nil {
		t.Fatalf("not raised: dead %v offer %v", w.dead, w.rez)
	}
	if want := float64(w.maxHP()) * 0.3; w.hp < want-1 || w.hp > want+1 {
		t.Errorf("raised with %v hit points, want 30 percent of %d", w.hp, w.maxHP())
	}
	if w.Body.X != h.Body.X || w.Body.Y != h.Body.Y {
		t.Errorf("raised at %v,%v, want where the Healer stands", w.Body.X, w.Body.Y)
	}
	if w.Char.XP != xp || w.Char.ConLost != con || r.z.graves[w.Char.ID] != nil {
		t.Errorf("raised with xp %d (was %d), con lost %d, grave %v: the release stage should have been skipped", w.Char.XP, xp, w.Char.ConLost, r.z.graves[w.Char.ID])
	}
	sick := false
	for _, e := range w.effects {
		if e.kind == effSick && e.spell.Name == "Resurrection Sickness" {
			sick = true
		}
	}
	if !sick {
		t.Error("the raised player has no resurrection sickness on their list")
	}
}

func TestResurrectionDeclineLapseAndRelease(t *testing.T) {
	t.Run("A decline gives the power back", func(t *testing.T) {
		r, h, w := rezWorld(t)
		power := h.power
		r.z.Cast(h, 7001, w.ID)
		r.ticks(4*TickRate + 5)
		r.z.AnswerResurrection(w, false)
		if !w.dead || w.rez != nil {
			t.Errorf("after a decline: dead %v offer %v", w.dead, w.rez)
		}
		if h.power < power-0.01 {
			t.Errorf("power %v, want %v given back", h.power, power)
		}
	})
	t.Run("An unanswered offer lapses", func(t *testing.T) {
		r, h, w := rezWorld(t)
		power := h.power
		r.z.Cast(h, 7001, w.ID)
		offs := r.tickCollect(w, 4*TickRate+5)
		offs = append(offs, r.tickCollect(w, int(rules.RezOfferSeconds*TickRate)+5)...)
		if w.rez != nil || !w.dead {
			t.Errorf("after the time: offer %v dead %v", w.rez, w.dead)
		}
		if len(offs) < 2 || offs[len(offs)-1].Secs != 0 {
			t.Errorf("the prompt was not withdrawn: %+v", offs)
		}
		if h.power < power-0.01 {
			t.Error("the lapsed offer did not give the power back")
		}
	})
	t.Run("Release while an offer is pending", func(t *testing.T) {
		r, h, w := rezWorld(t)
		power := h.power
		r.z.Cast(h, 7001, w.ID)
		r.ticks(4*TickRate + 5)
		xp := w.Char.XP
		r.release(w)
		if w.dead || w.rez != nil {
			t.Fatalf("not released: dead %v offer %v", w.dead, w.rez)
		}
		if w.Char.XP >= xp || w.Char.ConLost != 1 || r.z.graves[w.Char.ID] == nil {
			t.Errorf("the release cost xp %d -> %d, con lost %d, grave %v: it should cost what it always does", xp, w.Char.XP, w.Char.ConLost, r.z.graves[w.Char.ID])
		}
		if h.power < power-0.01 {
			t.Error("the withdrawn offer did not give the power back")
		}
		r.z.AnswerResurrection(w, true)
		if !said(w, "No one is currently trying") {
			t.Error("answering a withdrawn offer was not told so")
		}
	})
}

func TestResurrectionNeedsADeadPlayerOfTheRealm(t *testing.T) {
	r, h, w := rezWorld(t)
	living := r.player("Alive", 10, 30200, 30000, true)
	r.z.Cast(h, 7001, living.ID)
	if errs := errsOf(h); len(errs) == 0 || !strings.Contains(errs[0], "dead player") {
		t.Errorf("a resurrection aimed at a living player: %v", errs)
	}
	_ = w
}

func TestReleaseWaitsAndIsForcedLate(t *testing.T) {
	r, _, w := rezWorld(t)
	r.z.ReleasePlayer(w)
	if !w.dead || !said(w, "must wait") {
		t.Errorf("released at once: dead %v", w.dead)
	}
	r.ticks(int(rules.ReleaseAfter*TickRate) + 5)
	r.z.ReleasePlayer(w)
	if w.dead {
		t.Error("ten seconds on, release should work")
	}
	r2, _, w2 := rezWorld(t)
	r2.ticks(int((rules.ReleaseForcedAfter-1)*TickRate) - 5)
	if !w2.dead {
		t.Error("released before the forced time")
	}
	r2.ticks(2*TickRate + 10)
	if w2.dead {
		t.Error("still dead after the forced release time")
	}
}

func TestSicknessLowersDamageAndEnds(t *testing.T) {
	r, _, w := rezWorld(t)
	r.z.raise(w, &rezOffer{from: 0, fromName: "Someone", spell: raiseFallen, at: r.z.now})
	if f := sickFactor(w); f < 0.699 || f > 0.701 {
		t.Errorf("sickness leaves %v of the damage, want 0.7", f)
	}
	if got := sickened(100, w); got != 70 {
		t.Errorf("100 damage became %d, want 70", got)
	}
	if got := sickened(1, w); got != 1 {
		t.Errorf("1 damage became %d, want it kept at 1", got)
	}
	r.ticks(int(rules.SicknessSeconds*TickRate) + 10)
	if f := sickFactor(w); f != 1 {
		t.Errorf("after the time sickness still leaves %v", f)
	}
}
