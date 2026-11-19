package world

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// concRig is a Healer who knows two concentration buffs and a Warrior standing
// 500 units away to receive them. The first buff costs a third of the Healer's
// pool, the second most of it.
func concRig(t *testing.T) (r *rig, h, w *Player, small, big *SpellDef) {
	r = casterWorld(t)
	h = caster(r, "Healer", rules.ClassHealer, 10)
	w = r.player("Warrior", 10, 30500, 30000, true)
	pool := h.sh.maxConc
	if pool < 12 {
		t.Fatalf("a level 10 Healer's pool is %d", pool)
	}
	small = &SpellDef{ID: 8001, Name: "Blessing of Might", Type: rules.SpellStrConBuff, Target: "Realm", Range: 1000, CastTime: 1, Value: 12, Concentration: pool / 3}
	big = &SpellDef{ID: 8002, Name: "Blessing of Iron", Type: rules.SpellArmourBuff, Target: "Realm", Range: 1000, CastTime: 1, Value: 20, Concentration: pool - pool/3 + 1}
	for _, s := range []*SpellDef{small, big} {
		r.w.Cat.Spells[s.ID] = s
		r.w.Cat.LineSpells["Mending"] = append(r.w.Cat.LineSpells["Mending"], LineSpell{Spec: "Mending", Line: "Mending", Spell: s.ID, Level: 1})
	}
	r.w.refreshBook(h)
	return
}

// castAndLand casts a spell at a target and runs the clock until it lands.
func castAndLand(r *rig, p *Player, sp *SpellDef, target *Player) {
	r.z.Cast(p, sp.ID, target.ID)
	r.ticks(int(sp.CastTime*TickRate) + 5)
}

func concsOf(p *Player) []protocol.Conc { return sentOf[protocol.Conc](p, protocol.TConc) }

func hasConcEffect(p *Player) bool {
	for _, e := range p.effects {
		if e.cb != nil {
			return true
		}
	}
	return false
}

func TestConcentrationBuffHoldsItsCostWithNoTimer(t *testing.T) {
	r, h, w, small, _ := concRig(t)
	pool := h.sh.maxConc
	if m := h.statsMsg(r.w.Cat); m.MaxConc != pool || m.ConcUsed != 0 {
		t.Errorf("the sheet carries pool %d, in use %d; want %d and 0", m.MaxConc, m.ConcUsed, pool)
	}
	castAndLand(r, h, small, w)
	if !hasConcEffect(w) {
		t.Fatalf("the buff did not land: %v", errsOf(h))
	}
	if h.concUsed() != small.Concentration {
		t.Errorf("the Healer holds %d, want %d", h.concUsed(), small.Concentration)
	}
	// No timer: long after any ordinary buff would have ended it is still there.
	r.ticks(10 * 60 * TickRate / 10) // a minute
	if !hasConcEffect(w) {
		t.Error("the buff ended by itself")
	}
	if m := h.statsMsg(r.w.Cat); m.ConcUsed != small.Concentration {
		t.Errorf("the sheet says %d in use, want %d", m.ConcUsed, small.Concentration)
	}
	cs := concsOf(h)
	if len(cs) == 0 {
		t.Fatal("no concentration message was sent")
	}
	last := cs[len(cs)-1]
	if last.Used != small.Concentration || last.Max != pool || len(last.List) != 1 || last.List[0].Name != "Blessing of Might" || last.List[0].Target != "Warrior" {
		t.Errorf("the concentration list is %+v", last)
	}
	// The effect view has no clock to show.
	r.z.effectsChanged(w)
	var sick bool
	for _, e := range sentOf[protocol.Effects](w, protocol.TEffects) {
		for _, v := range e.List {
			if v.Name == "Blessing of Might" && (!v.Conc || v.Left != 0) {
				sick = true
			}
		}
	}
	if sick {
		t.Error("a concentration buff was sent with a clock")
	}
}

func TestConcentrationRefusesWhatDoesNotFit(t *testing.T) {
	r, h, w, small, big := concRig(t)
	castAndLand(r, h, small, w)
	power := h.power
	sentOf[protocol.Err](h, protocol.TErr)
	r.z.Cast(h, big.ID, w.ID)
	want := rules.ConcentrationShort(big.Concentration)
	if errs := errsOf(h); len(errs) == 0 || errs[0] != want || h.casting != nil {
		t.Errorf("errors %v, casting %v; want %q", errs, h.casting, want)
	}
	if h.power != power {
		t.Error("a refused buff cost power")
	}
}

func TestConcentrationRecastReplacesAndDoesNotDoubleCharge(t *testing.T) {
	r, h, w, small, _ := concRig(t)
	castAndLand(r, h, small, w)
	castAndLand(r, h, small, w)
	if h.concUsed() != small.Concentration || len(h.concs) != 1 {
		t.Errorf("after a recast the Healer holds %d in %d buffs, want %d in 1", h.concUsed(), len(h.concs), small.Concentration)
	}
	// A recast fits even when only the old buff's own cost is free.
	h.sh.maxConc = small.Concentration
	if msg := r.z.concRefusal(h, small, w); msg != "" {
		t.Errorf("a recast of the held buff was refused: %q", msg)
	}
}

func TestConcentrationEnds(t *testing.T) {
	t.Run("The target runs off", func(t *testing.T) {
		r, h, w, small, _ := concRig(t)
		castAndLand(r, h, small, w)
		w.Body.X = 30000 + rules.ConcentrationRange + 500
		r.ticks(int((rules.ConcentrationGrace-3)*TickRate) + 2)
		if !hasConcEffect(w) || h.concUsed() == 0 {
			t.Fatal("the buff ended inside the grace time")
		}
		w.Body.X = 30500 // back in range: the grace time starts again
		r.ticks(2 * TickRate)
		w.Body.X = 30000 + rules.ConcentrationRange + 500
		r.ticks(int((rules.ConcentrationGrace-3)*TickRate) + 2)
		if !hasConcEffect(w) {
			t.Fatal("coming back into range did not reset the grace time")
		}
		r.ticks(5 * TickRate)
		if hasConcEffect(w) || h.concUsed() != 0 || len(h.concs) != 0 {
			t.Errorf("past the grace time the buff held: effect %v, held %d", hasConcEffect(w), h.concUsed())
		}
	})
	t.Run("The caster dies", func(t *testing.T) {
		r, h, w, small, _ := concRig(t)
		castAndLand(r, h, small, w)
		castAndLand(r, h, small, h) // and one on themselves
		if len(h.concs) != 2 {
			t.Fatalf("the Healer holds %d buffs", len(h.concs))
		}
		r.z.killPlayer(h, r.z.npcs[0])
		if hasConcEffect(w) || hasConcEffect(h) || h.concUsed() != 0 {
			t.Errorf("after death: on the Warrior %v, on the Healer %v, held %d", hasConcEffect(w), hasConcEffect(h), h.concUsed())
		}
	})
	t.Run("The target dies", func(t *testing.T) {
		r, h, w, small, _ := concRig(t)
		castAndLand(r, h, small, w)
		r.z.killPlayer(w, r.z.npcs[0])
		if h.concUsed() != 0 || len(h.concs) != 0 {
			t.Errorf("the target died and the Healer still holds %d", h.concUsed())
		}
	})
	t.Run("The target leaves", func(t *testing.T) {
		r, h, w, small, _ := concRig(t)
		castAndLand(r, h, small, w)
		w.closed.Store(true)
		r.ticks(2 * TickRate)
		if h.concUsed() != 0 {
			t.Errorf("the target left and the Healer still holds %d", h.concUsed())
		}
	})
	t.Run("The caster leaves", func(t *testing.T) {
		r, h, w, small, _ := concRig(t)
		castAndLand(r, h, small, w)
		h.closed.Store(true)
		r.ticks(2 * TickRate)
		if hasConcEffect(w) {
			t.Error("the caster left and the buff is still on the Warrior")
		}
	})
	t.Run("Cancelling from the list", func(t *testing.T) {
		r, h, w, small, _ := concRig(t)
		castAndLand(r, h, small, w)
		id := h.concs[0].id
		r.z.CancelConc(h, id+99) // an id that is not theirs does nothing
		if h.concUsed() == 0 {
			t.Fatal("cancelling an unknown id ended the buff")
		}
		r.z.CancelConc(h, id)
		if hasConcEffect(w) || h.concUsed() != 0 {
			t.Errorf("after cancelling: on the Warrior %v, held %d", hasConcEffect(w), h.concUsed())
		}
		raw := drainRaw(h)
		var last protocol.Conc
		seen, told := false, false
		for _, m := range raw {
			if strings.Contains(m, `"t":"conc"`) {
				json.Unmarshal([]byte(m), &last)
				seen = true
			}
			if strings.Contains(m, "You cancel your Blessing of Might") {
				told = true
			}
		}
		if !seen || last.Used != 0 || len(last.List) != 0 {
			t.Errorf("the freed concentration was not sent at once: %+v", last)
		}
		if !told {
			t.Error("the Healer was not told")
		}
	})
}

// Whatever ends a buff, nothing is left behind: not on the target, not held by
// the caster, and every end path agrees.
func TestConcentrationNeverLeaks(t *testing.T) {
	for _, how := range []string{"cancel", "caster dies", "target dies", "target leaves", "caster leaves", "range"} {
		t.Run(how, func(t *testing.T) {
			r, h, w, small, _ := concRig(t)
			castAndLand(r, h, small, w)
			castAndLand(r, h, small, h)
			switch how {
			case "cancel":
				for len(h.concs) > 0 {
					r.z.CancelConc(h, h.concs[0].id)
				}
			case "caster dies":
				r.z.killPlayer(h, r.z.npcs[0])
			case "target dies":
				r.z.killPlayer(w, r.z.npcs[0])
				r.z.killPlayer(h, r.z.npcs[0])
			case "target leaves":
				w.closed.Store(true)
				h.closed.Store(true)
				r.ticks(2 * TickRate)
			case "caster leaves":
				h.closed.Store(true)
				r.ticks(2 * TickRate)
			case "range":
				w.Body.X = 90000
				r.ticks(int((rules.ConcentrationGrace+2)*TickRate) + 5)
				r.z.CancelConc(h, h.concs[0].id)
			}
			for _, p := range []*Player{h, w} {
				if hasConcEffect(p) {
					t.Errorf("%s still carries a concentration effect", p.Char.Name)
				}
			}
			if n := h.concUsed(); n != 0 {
				t.Errorf("the Healer still holds %d", n)
			}
			_ = fmt.Sprint
			_ = strings.Contains
		})
	}
}

// drainRaw takes everything waiting for a player, as text.
func drainRaw(p *Player) (out []string) {
	for {
		select {
		case b := <-p.out:
			out = append(out, string(b))
		default:
			return
		}
	}
}
