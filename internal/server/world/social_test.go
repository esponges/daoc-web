package world

import (
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
)

func socialPair(t *testing.T) (*grouped, *Player, *Player) {
	g := newGrouped(t)
	a := g.join("Thora", 30000)
	b := g.join("Bjorn", 30100)
	g.w.loadSocial(a)
	g.w.loadSocial(b)
	return g, a, b
}

func TestFriendsListIsKeptAndComesOnline(t *testing.T) {
	g, a, b := socialPair(t)
	g.cmd(a, "friend Bjorn")
	if !a.isFriend(b.Char.ID) {
		t.Fatal("Bjorn is not Thora's friend")
	}
	// It was written: a fresh load sees it.
	a.social.friends = nil
	g.w.loadSocial(a)
	if !a.isFriend(b.Char.ID) {
		t.Error("the friend was not saved")
	}
	drain(a)
	g.w.notifyFriends(b, "has entered the game.")
	out, _, _ := inbox(a)
	if !has(out, "Bjorn has entered the game.") {
		t.Errorf("Thora was not told her friend came: %v", out)
	}
	g.cmd(a, "friend bjorn") // again removes
	if a.isFriend(b.Char.ID) {
		t.Error("friend again did not remove")
	}
	_ = protocol.TFriends
}

func TestIgnoredSpeechAndTellsAreHidden(t *testing.T) {
	g, a, b := socialPair(t)
	g.cmd(a, "ignore Bjorn")
	drain(a)
	drain(b)
	g.w.Chat(b, "say", "hello Thora")
	if out, _, _ := inbox(a); has(out, "hello Thora") {
		t.Error("an ignored player's say reached Thora")
	}
	g.cmd(b, "tell Thora hi there")
	if out, _, _ := inbox(a); has(out, "hi there") {
		t.Error("an ignored player's tell reached Thora")
	}
	if out, _, _ := inbox(b); !has(out, "hi there") {
		t.Error("the ignored sender did not see their own tell go")
	}
	g.cmd(a, "ignore Bjorn") // off again
	drain(a)
	g.w.Chat(b, "say", "hello again")
	if out, _, _ := inbox(a); !has(out, "hello again") {
		t.Error("speech did not return after un-ignoring")
	}
}

func TestAfkAnswersTellsAndEndsWhenYouTalk(t *testing.T) {
	g, a, b := socialPair(t)
	g.cmd(a, "afk back soon")
	drain(a)
	drain(b)
	g.cmd(b, "tell Thora are you there")
	if out, _, _ := inbox(b); !has(out, "Thora is AFK: back soon") {
		t.Errorf("Bjorn got no away reply: %v", out)
	}
	g.w.Chat(a, "say", "I am back")
	if _, away := a.awayNote(); away {
		t.Error("speaking did not end being away")
	}
}

func TestWhoHidesAnonymousPlayersAndMarksLFG(t *testing.T) {
	g, a, b := socialPair(t)
	a.sh.class.Name = "Warrior"
	b.sh.class.Name = "Healer"
	g.cmd(a, "anon")
	g.cmd(b, "lfg")
	drain(b)
	g.cmd(b, "who")
	out, _, _ := inbox(b)
	if len(out) == 0 {
		t.Fatal("no /who answer")
	}
	line := out[len(out)-1]
	if !strings.Contains(line, "Thora") || strings.Contains(line, "Thora (L") {
		t.Errorf("an anonymous player shows details: %q", line)
	}
	if !strings.Contains(line, "Bjorn (L10 Healer) [LFG]") {
		t.Errorf("Bjorn's own line: %q", line)
	}
}

func TestAssistFaceAndFollow(t *testing.T) {
	g, a, b := socialPair(t)
	c := g.join("Sigrid", 30200)
	a.target = c.ID
	b.target = a.ID
	g.z.Assist(b)
	if b.target != c.ID {
		t.Errorf("assist left Bjorn targeting %d, want Sigrid %d", b.target, c.ID)
	}
	// Face: Bjorn turns to Thora.
	b.Body.Yaw = 0
	b.target = a.ID
	g.z.Face(b)
	if want := angleTo(b.Body.X, b.Body.Y, a.Body.X, a.Body.Y); b.Body.Yaw != want {
		t.Errorf("yaw %v, want %v", b.Body.Yaw, want)
	}
	// Follow: Thora walks off, Bjorn comes after her.
	b.target = a.ID
	g.z.Follow(b)
	if b.follow != a.ID {
		t.Fatal("Bjorn does not follow")
	}
	a.Body.X += 3000
	d0 := a.Body.X - b.Body.X
	for i := 0; i < TickRate*5; i++ {
		g.z.Tick()
		drain(a)
		drain(b)
	}
	if a.Body.X-b.Body.X >= d0 {
		t.Errorf("Bjorn did not close in: gap %v from %v", a.Body.X-b.Body.X, d0)
	}
	// A movement key ends it.
	b.queue = []stepRun{{n: 1, k: Keys{Fwd: 1}}}
	g.z.Tick()
	if b.follow != 0 {
		t.Error("pressing a movement key did not stop the follow")
	}
}

func TestInspectListsWornItems(t *testing.T) {
	g, a, b := socialPair(t)
	g.give(b, SlotWeapon, "hammer", 1)
	a.target = b.ID
	drain(a)
	g.z.Inspect(a)
	if out, _, _ := inbox(a); !has(out, "bronze hammer") {
		t.Errorf("inspect showed %v", out)
	}
}
