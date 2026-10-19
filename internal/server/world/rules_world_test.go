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

// oldCharacter saves a character as the server did before the rule book: no
// stats, experience on the old curve.
func (r *rig) oldCharacter(name string, level int, xp int64) *Player {
	r.t.Helper()
	a, _ := r.st.AccountFor("d:"+name, name, false)
	ch := &store.Character{AccountID: a.ID, Name: name, Race: 6, Class: 22, Level: level, XP: xp, Zone: 100, X: 30000, Y: 30000}
	if err := r.st.CreateCharacter(ch); err != nil {
		r.t.Fatal(err)
	}
	loaded, err := r.st.Character(a.ID, ch.ID)
	if err != nil {
		r.t.Fatal(err)
	}
	p := newPlayer(a, loaded)
	r.z.join(p)
	return p
}

// sent decodes the messages of a type waiting for a player.
func sentOf[T any](p *Player, typ string) (out []T) {
	for {
		select {
		case b := <-p.out:
			var env protocol.Envelope
			json.Unmarshal(b, &env)
			if env.T == typ {
				var v T
				json.Unmarshal(b, &v)
				out = append(out, v)
			}
		default:
			return
		}
	}
}

func TestOldCharactersAreBroughtUpToTheRuleBook(t *testing.T) {
	r := newRig(t)
	// Halfway through 10 on the old curve: 3420 at 10, 4200 at 11.
	p := r.oldCharacter("Halfway", 10, 3810)
	if p.Char.Level != 10 || p.Char.XP != 331950 {
		t.Errorf("level %d xp %d, want 10 and 331950", p.Char.Level, p.Char.XP)
	}
	if !p.Char.HasStats || p.Char.Stats != [8]int{110, 80, 45, 35, 60, 60, 60, 60} {
		t.Errorf("stats %v", p.Char.Stats)
	}
	saved, _ := r.st.Character(p.Acct.ID, p.Char.ID)
	if saved.XPScheme != store.XPSchemeRules || saved.XP != 331950 || !saved.HasStats {
		t.Errorf("not saved converted: %+v", saved)
	}
	old20 := r.oldCharacter("Veteran", 20, 15000)
	if old20.sh.unspent != 404 {
		t.Errorf("a level 20 Warrior has %d unspent points, want 404", old20.sh.unspent)
	}
	for k := range old20.Char.Specs {
		t.Errorf("spec %s trained on a migrated character", k)
	}
}

func TestSheetIsSentToItsOwnerAfterTheWelcome(t *testing.T) {
	r := newRig(t)
	a := r.player("Alice", 5, 30000, 30000, true)
	order := []string{}
	for {
		select {
		case b := <-a.out:
			var env protocol.Envelope
			json.Unmarshal(b, &env)
			order = append(order, env.T)
			continue
		default:
		}
		break
	}
	if len(order) < 2 || order[0] != protocol.TWelcome || order[1] != protocol.TStats {
		t.Errorf("join sent %v; want welcome then stats", order)
	}
	b := r.player("Bobby", 5, 30200, 30000, true)
	sentOf[protocol.Stats](b, protocol.TStats)
	r.st.Tx(func(tx *store.Tx) error { return tx.AddItem(a.Char.ID, "strhelm", 0, 1) })
	r.w.loadInventory(a)
	r.z.MoveItem(a, 0, SlotHead)
	if got := sentOf[protocol.Stats](b, protocol.TStats); len(got) != 0 {
		t.Error("another player was sent Alice's sheet")
	}
	got := sentOf[protocol.Stats](a, protocol.TStats)
	if len(got) == 0 || got[len(got)-1].Stats[rules.STR] != a.sh.stats[rules.STR] || got[len(got)-1].Class != "Warrior" {
		t.Errorf("Alice's sheets %+v", got)
	}
}

func TestAStyleThatIsParried(t *testing.T) {
	sp := wolf(1, 30100, 30000, 5)
	r := newRig(t, sp)
	r.w.Cat.Templates[7] = &Template{ID: 7, Parry: 100}
	sp.Template = 7
	r.z = NewZone(r.w, 100, "Test", flat{}, nil, []Spawn{sp})
	r.w.zones[100] = r.z
	p := r.player("Parried", 5, 30000, 30000, true)
	r.train(p, map[string]int{"Hammer": 5})
	n := r.z.byID[1]
	n.Roam, n.Aggro = 0, 0
	n.Body.Yaw = -math.Pi / 2 // facing the player, so it can parry
	p.Body.Yaw = math.Pi / 2
	r.z.SetTarget(p, 1)
	r.z.SetAttack(p, true)
	anvil := -1
	for i, d := range p.bar {
		if d.Name == "Thor's Anvil" {
			anvil = i
		}
	}
	r.z.requestStyle(p, anvil)
	p.end = 60
	for i := 0; i < TickRate*8 && p.pending >= 0; i++ {
		r.z.Tick()
	}
	half := float64(rules.FailedStyleCost(3.1, 15))
	if p.lastStyle != 0 || math.Abs(p.end-(60-half)) > float64(rules.EnduranceRegenIn)*8 {
		t.Errorf("after a parried style: last style %d, endurance %v (60 less half the cost, %v, plus regeneration)", p.lastStyle, p.end, half)
	}
	if p.end > 60-half+float64(rules.EnduranceRegenIn)*8 || p.end < 60-float64(rules.StyleCost(3.1, 15)) {
		t.Errorf("endurance %v: a parried style costs half", p.end)
	}
}

func TestStylesFollowTrainedSpec(t *testing.T) {
	r := newRig(t)
	p := r.player("Untrained", 12, 30000, 30000, true)
	has := func(name string) bool {
		for _, d := range p.bar {
			if d.Name == name {
				return true
			}
		}
		return false
	}
	if has("Thor's Anvil") || has("Crushing Blow") || has("Numb") {
		t.Errorf("a level 12 Warrior with hammer and shields at 1 has %v", p.bar)
	}
	r.train(p, map[string]int{"Hammer": 4})
	if !has("Thor's Anvil") || !has("Crushing Blow") || has("High Level") {
		t.Errorf("hammer 4 bar: %v", p.bar)
	}
}

func TestKillExperienceFollowsTheRules(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 10), wolf(2, 30200, 30000, 10), wolf(3, 30300, 30000, 9))
	a := r.player("Solo", 10, 30000, 30000, true)
	b := r.player("Helper", 10, 30000, 30100, true)
	c := r.player("Elder", 20, 30000, 30200, true)
	kill := func(id int64, dmg map[*Player]int) {
		n := r.z.byID[id]
		for p, d := range dmg {
			n.hurtBy[p.ID] = d
		}
		r.z.killNPC(n)
	}
	xa := a.Char.XP
	kill(1, map[*Player]int{a: 200})
	if d := a.Char.XP - xa; d != 5120 {
		t.Errorf("an even solo kill gave %d, want 5120", d)
	}
	xa, xb := a.Char.XP, b.Char.XP
	kill(2, map[*Player]int{a: 150, b: 50})
	if da, db := a.Char.XP-xa, b.Char.XP-xb; da != 3840 || db != 1280 {
		t.Errorf("shared kill gave %d and %d, want 3840 and 1280", da, db)
	}
	xa, xc := a.Char.XP, c.Char.XP
	kill(3, map[*Player]int{a: 100, c: 100})
	if a.Char.XP != xa || c.Char.XP != xc {
		t.Error("a creature grey to one of its killers gave experience")
	}
}

func TestLevelUpAnnouncesWhatItBrought(t *testing.T) {
	r := newRig(t)
	p := r.player("Climber", 6, 30000, 30000, true)
	r.z.ev = r.z.ev[:0]
	r.z.award(p, rules.XPToReach(7)-p.Char.XP)
	var text string
	for _, e := range r.z.ev {
		if e.e.K == "level" {
			text = e.e.Text
		}
	}
	if !strings.Contains(text, "level 7") || !strings.Contains(text, "strength rises by 1") || !strings.Contains(text, "14 specialization points") {
		t.Errorf("level 7 announcement %q", text)
	}
}

func TestCreaturesCarryTheirConColour(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 12))
	p := r.player("Viewer", 10, 30000, 30000, true)
	r.ticks(2)
	ss := snaps(p)
	ents := ss[len(ss)-1].Ents
	if len(ents) != 1 || ents[0].Con == nil || *ents[0].Con != rules.ConOrange {
		t.Fatalf("entities %+v", ents)
	}
	b, _ := json.Marshal(ents[0])
	if !strings.Contains(string(b), `"cc":1`) {
		t.Errorf("the wire form lacks cc: %s", b)
	}
}
