package world

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// lines drains a player's queue and returns the chat and error lines sent.
func lines(p *Player) (out []string) {
	for {
		select {
		case m := <-p.out:
			var e struct {
				T    string `json:"t"`
				Text string `json:"text"`
			}
			json.Unmarshal(m, &e)
			if e.T == protocol.TChatMsg || e.T == protocol.TErr {
				out = append(out, e.Text)
			}
		default:
			return
		}
	}
}

func said(p *Player, part string) bool {
	for _, l := range lines(p) {
		if strings.Contains(l, part) {
			return true
		}
	}
	return false
}

// bindRig is a zone with a bindstone at 30000,30000 and another at 50000,50000.
func bindRig(t *testing.T) *rig {
	r := newRig(t)
	r.w.Cat.Binds = []Bind{
		{Zone: 100, X: 30000, Y: 30000, Z: 5000, Radius: 750, Realm: 2},
		{Zone: 100, X: 50000, Y: 50000, Z: 5000, Radius: 750, Realm: 2},
		{Zone: 100, X: 40000, Y: 40000, Z: 5000, Radius: 750, Realm: 1}, // another realm's
	}
	return r
}

func TestBindInTownAndReturnOnDeath(t *testing.T) {
	r := bindRig(t)
	p := r.player("Binder", 3, 50300, 50000, false)
	r.w.command(r.z, p, "bind")
	if !said(p, "You are bound here") {
		t.Fatal("no word that the player is bound")
	}
	if b := p.Char.Bind; b == nil || b.X != 50000 || b.Y != 50000 {
		t.Fatalf("bind %+v, want the 50000 bindstone", p.Char.Bind)
	}
	saved, err := r.st.Character(p.Char.AccountID, p.Char.ID)
	if err != nil || saved.Bind == nil || saved.Bind.X != 50000 {
		t.Errorf("the bind was not saved: %+v %v", saved.Bind, err)
	}
	// Dying far away returns them to the bind, not to the nearest bindstone.
	p.Body.X, p.Body.Y = 30500, 30000
	p.dead = true
	r.z.ReleasePlayer(p)
	if d := math.Hypot(p.Body.X-50000, p.Body.Y-50000); d > 1 {
		t.Errorf("released at %v,%v, want the bind at 50000,50000", p.Body.X, p.Body.Y)
	}
}

func TestBindRefusedInOpenCountry(t *testing.T) {
	r := bindRig(t)
	p := r.player("Wanderer", 3, 30000, 30000, false)
	before := *p.Char.Bind
	p.Body.X, p.Body.Y = 36000, 30000
	r.w.command(r.z, p, "bind")
	if !said(p, "nothing to bind to here") {
		t.Error("no refusal")
	}
	if *p.Char.Bind != before {
		t.Errorf("bind moved to %+v", *p.Char.Bind)
	}
	// Another realm's bindstone does not count either.
	p.Body.X, p.Body.Y = 40100, 40000
	r.w.command(r.z, p, "bind")
	if *p.Char.Bind != before {
		t.Errorf("bound at another realm's bindstone: %+v", *p.Char.Bind)
	}
}

// A character saved before binding is bound, on joining, at the bindstone
// nearest to where they stand, and that is saved.
func TestOldCharacterBoundOnLogin(t *testing.T) {
	r := bindRig(t)
	p := r.player("Veteran", 12, 47000, 47000, false)
	if b := p.Char.Bind; b == nil || b.X != 50000 {
		t.Fatalf("bind %+v, want the nearest bindstone of their realm", p.Char.Bind)
	}
	saved, _ := r.st.Character(p.Char.AccountID, p.Char.ID)
	if saved.Bind == nil || saved.Bind.X != 50000 {
		t.Errorf("first-login bind not saved: %+v", saved.Bind)
	}
}

// A low-level character at a start the rule replaced is moved and bound to
// the replacement once, and told; a higher one stays.
func TestLowLevelMovedOffUnsafeStart(t *testing.T) {
	for _, c := range []struct {
		level int
		moved bool
	}{{1, true}, {12, false}} {
		r := bindRig(t)
		r.w.Cat.Unsafe = []UnsafeStart{{Realm: 2, Race: 6, Class: 22, At: Start{Zone: 100, X: 20000, Y: 20000},
			Worst: rules.Threat{Name: "bog bear", Level: 26}, Instead: Start{Zone: 100, X: 50000, Y: 50000, Z: 5000}}}
		a, _ := r.st.AccountFor("d:mover", "mover", false)
		ch := &store.Character{AccountID: a.ID, Name: "Mover", Race: 6, Class: 22, Level: c.level, XP: rules.XPToReach(c.level),
			Zone: 100, X: 20300, Y: 20000, XPScheme: store.XPSchemeRules, Bind: &store.Place{Zone: 100, X: 20000, Y: 20000}}
		r.st.CreateCharacter(ch)
		p := newPlayer(a, ch)
		r.w.leaveUnsafeStart(p)
		r.z.join(p)
		moved := math.Hypot(p.Body.X-50000, p.Body.Y-50000) < 1
		if moved != c.moved {
			t.Errorf("level %d: at %v,%v, moved %v, want %v", c.level, p.Body.X, p.Body.Y, moved, c.moved)
		}
		if c.moved {
			if p.Char.Bind.X != 50000 {
				t.Errorf("level %d: bound at %+v, want the replacement", c.level, p.Char.Bind)
			}
			if !said(p, "too dangerous for a new character") {
				t.Errorf("level %d: no word of the move", c.level)
			}
			// Once moved they are at a safe place, so the next login leaves them be.
			r.w.leaveUnsafeStart(p)
			if p.notice != "" {
				t.Error("moved a second time")
			}
		} else if p.Char.Bind.X != 20000 {
			t.Errorf("level %d: bind changed to %+v", c.level, p.Char.Bind)
		}
	}
}
