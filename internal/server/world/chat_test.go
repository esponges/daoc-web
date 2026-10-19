package world

import (
	"encoding/json"
	"testing"

	"daocweb/internal/server/protocol"
)

// chats drains a player's queue and returns the chat lines sent.
func chats(p *Player) (out []protocol.ChatMsg) {
	for {
		select {
		case b := <-p.out:
			var m protocol.ChatMsg
			json.Unmarshal(b, &m)
			if m.T == protocol.TChatMsg {
				out = append(out, m)
			}
		default:
			return
		}
	}
}

// A tell across zones reaches only the named player, matched without regard
// to case, and the sender sees it as sent.
func TestTellAcrossZones(t *testing.T) {
	g := newGrouped(t)
	ragnar, sigrid, other := g.join("Ragnar", 30000), g.join("Sigrid", 30100), g.join("Bystander", 30050)
	// Sigrid is in Myrkwood Forest.
	z2 := NewZone(g.w, 105, "Myrkwood Forest", flat{}, nil, nil)
	delete(g.z.players, sigrid.ID)
	z2.add(sigrid)
	g.w.Chat(ragnar, "say", "/tell sIGRID meet at  the bridge")
	got := chats(sigrid)
	if len(got) != 1 || got[0].Ch != "tell" || got[0].From != "Ragnar" || got[0].Text != "meet at  the bridge" {
		t.Fatalf("Sigrid got %+v", got)
	}
	mine := chats(ragnar)
	if len(mine) != 1 || mine[0].Ch != "tell" || mine[0].To != "Sigrid" || mine[0].From != "" {
		t.Errorf("Ragnar saw %+v, want his own line as sent to Sigrid", mine)
	}
	if c := chats(other); len(c) != 0 {
		t.Errorf("a bystander got %+v", c)
	}
	g.w.Chat(ragnar, "say", "/send Sigrid again")
	if c := chats(sigrid); len(c) != 1 || c[0].Text != "again" {
		t.Errorf("/send: %+v", c)
	}
}

// Not in the game: the sender is told, and nothing is sent.
func TestTellNotInTheGame(t *testing.T) {
	g := newGrouped(t)
	ragnar, hib := g.join("Ragnar", 30000), g.join("Fionn", 30100)
	hib.Arch.Realm = 3
	for _, name := range []string{"Nobody", "Fionn"} {
		g.w.Chat(ragnar, "say", "/tell "+name+" hello")
		c := chats(ragnar)
		if len(c) != 1 || c[0].Ch != "sys" || c[0].Text != name+" is not in the game." {
			t.Errorf("telling %s: %+v", name, c)
		}
	}
	if c := chats(hib); len(c) != 0 {
		t.Errorf("a player of another realm got %+v", c)
	}
}

// A group line reaches the three members and not a player beside them; a
// player in no group is told so.
func TestGroupChat(t *testing.T) {
	g := newGrouped(t)
	a, b, c := g.join("Ragnar", 30000), g.join("Sigrid", 30100), g.join("Bjorn", 30200)
	stranger := g.join("Stranger", 30010)
	g.form(a, b, c)
	g.w.Chat(b, "say", "/g pull now")
	for _, p := range []*Player{a, b, c} {
		got := chats(p)
		if len(got) != 1 || got[0].Ch != "group" || got[0].From != "Sigrid" || got[0].Text != "pull now" {
			t.Errorf("%s got %+v", p.Char.Name, got)
		}
	}
	if got := chats(stranger); len(got) != 0 {
		t.Errorf("a player outside the group got %+v", got)
	}
	g.w.Chat(stranger, "say", "/g hello")
	if got := chats(stranger); len(got) != 1 || got[0].Text != "You are not in a group." {
		t.Errorf("no group: %+v", got)
	}
}

// Tells are rate limited like say.
func TestTellsAreRateLimited(t *testing.T) {
	g := newGrouped(t)
	ragnar, sigrid := g.join("Ragnar", 30000), g.join("Sigrid", 30100)
	ragnar.chatTokens, ragnar.chatAt = 4, g.z.now
	for i := 0; i < 6; i++ {
		g.w.Chat(ragnar, "say", "/tell Sigrid spam")
	}
	if got := len(chats(sigrid)); got != 4 {
		t.Errorf("Sigrid got %d tells, want the burst of 4", got)
	}
	var tooFast int
	for {
		select {
		case b := <-ragnar.out:
			var e protocol.Err
			json.Unmarshal(b, &e)
			if e.T == protocol.TErr && e.Text == "You are talking too fast." {
				tooFast++
			}
			continue
		default:
		}
		break
	}
	if tooFast != 2 {
		t.Errorf("told %d times about talking too fast, want 2", tooFast)
	}
}
