package world

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"daocweb/internal/server/protocol"
)

// grouped is a rig whose players are in the game, so they can be found by
// name and grouped, on a clock the test moves.
type grouped struct {
	*rig
	now time.Time
}

func newGrouped(t *testing.T) *grouped {
	g := &grouped{rig: newRig(t), now: time.Unix(1_700_000_000, 0)}
	g.w.groups.clock = func() time.Time { return g.now }
	return g
}

func (g *grouped) join(name string, x float64) *Player {
	p := g.player(name, 10, x, 30000, false)
	g.w.omu.Lock()
	g.w.online[p.Char.ID] = p
	g.w.omu.Unlock()
	drain(p)
	return p
}

func (g *grouped) cmd(p *Player, line string) { g.w.command(g.z, p, line) }

// drain empties a player's queue.
func drain(p *Player) {
	for {
		select {
		case <-p.out:
		default:
			return
		}
	}
}

// inbox drains a player's queue: the chat lines, group windows and
// invitations they were sent.
func inbox(p *Player) (out []string, groups []protocol.Group, invites int) {
	for {
		select {
		case b := <-p.out:
			var env protocol.Envelope
			json.Unmarshal(b, &env)
			switch env.T {
			case protocol.TChatMsg:
				var m protocol.ChatMsg
				json.Unmarshal(b, &m)
				out = append(out, m.Text)
			case protocol.TGroup:
				var m protocol.Group
				json.Unmarshal(b, &m)
				groups = append(groups, m)
			case protocol.TGroupInv:
				invites++
			}
		default:
			return
		}
	}
}

func has(ls []string, part string) bool {
	for _, l := range ls {
		if strings.Contains(l, part) {
			return true
		}
	}
	return false
}

func (g *grouped) members(p *Player) []string {
	var out []string
	for _, c := range g.w.groupOf(p.Char.ID) {
		out = append(out, g.w.onlineChar(c).Char.Name)
	}
	return out
}

func (g *grouped) form(leader *Player, others ...*Player) {
	for _, o := range others {
		g.w.Invite(leader, o, o.Char.Name)
		g.w.AcceptInvite(o)
	}
	drain(leader)
	for _, p := range others {
		drain(p)
	}
}

// Forming a group: Ragnar invites Sigrid, she accepts, Ragnar leads, and both
// see the other.
func TestGroupForms(t *testing.T) {
	g := newGrouped(t)
	ragnar, sigrid := g.join("Ragnar", 30000), g.join("Sigrid", 30100)
	g.cmd(ragnar, "invite sigrid")
	if _, _, inv := inbox(sigrid); inv != 1 {
		t.Fatalf("Sigrid got %d invitations, want 1", inv)
	}
	if g.w.groupOf(sigrid.Char.ID) != nil {
		t.Fatal("Sigrid is grouped before accepting")
	}
	g.cmd(sigrid, "accept")
	if got := fmt.Sprint(g.members(sigrid)); got != "[Ragnar Sigrid]" {
		t.Fatalf("members %s, want Ragnar leading Sigrid", got)
	}
	g.ticks(TickRate)
	for _, p := range []*Player{ragnar, sigrid} {
		_, gs, _ := inbox(p)
		if len(gs) == 0 {
			t.Fatalf("%s was sent no group window", p.Char.Name)
		}
		last := gs[len(gs)-1]
		if len(last.Members) != 2 || last.Members[0].Name != "Ragnar" || last.Members[1].Name != "Sigrid" ||
			last.Members[1].Level != 10 || last.Members[1].HP != 100 || last.Members[1].Class != "Warrior" || !last.Members[1].InRange {
			t.Errorf("%s sees %+v", p.Char.Name, last.Members)
		}
	}
}

// Inviting by target.
func TestGroupInviteByTarget(t *testing.T) {
	g := newGrouped(t)
	ragnar, sigrid := g.join("Ragnar", 30000), g.join("Sigrid", 30100)
	g.z.SetTarget(ragnar, sigrid.ID)
	g.cmd(ragnar, "invite")
	if _, _, inv := inbox(sigrid); inv != 1 {
		t.Fatalf("Sigrid got %d invitations by target, want 1", inv)
	}
}

// Declining, or not answering within 60 seconds: no group, and the inviter
// is told.
func TestGroupInviteDeclinedOrLapsed(t *testing.T) {
	g := newGrouped(t)
	ragnar, sigrid := g.join("Ragnar", 30000), g.join("Sigrid", 30100)
	g.w.Invite(ragnar, sigrid, "Sigrid")
	g.cmd(sigrid, "decline")
	if ls, _, _ := inbox(ragnar); !has(ls, "Sigrid declines your invitation") {
		t.Errorf("Ragnar was told %q", ls)
	}
	g.w.AcceptInvite(sigrid)
	if g.w.groupOf(sigrid.Char.ID) != nil {
		t.Fatal("a declined invitation formed a group")
	}
	g.w.Invite(ragnar, sigrid, "Sigrid")
	drain(ragnar)
	g.now = g.now.Add(61 * time.Second)
	g.ticks(TickRate)
	if ls, _, _ := inbox(ragnar); !has(ls, "Sigrid declines your invitation") {
		t.Errorf("after the lapse Ragnar was told %q", ls)
	}
	g.w.AcceptInvite(sigrid)
	if g.w.groupOf(sigrid.Char.ID) != nil {
		t.Fatal("a lapsed invitation formed a group")
	}
}

// Another realm: refused, and nothing reaches the other player.
func TestGroupRefusesAnotherRealm(t *testing.T) {
	g := newGrouped(t)
	ragnar, hib := g.join("Ragnar", 30000), g.join("Fionn", 30100)
	hib.Arch.Realm = 3
	g.w.Invite(ragnar, hib, "Fionn")
	if ls, _, _ := inbox(ragnar); !has(ls, "another realm") {
		t.Errorf("Ragnar was told %q", ls)
	}
	if ls, gs, inv := inbox(hib); len(ls)+len(gs)+inv != 0 {
		t.Errorf("the other realm's player was sent %q %v %d", ls, gs, inv)
	}
}

// A full group, an already-grouped player, and only the leader inviting.
func TestGroupLimits(t *testing.T) {
	g := newGrouped(t)
	var ps []*Player
	for i := 0; i < 9; i++ {
		ps = append(ps, g.join(fmt.Sprintf("Member%c", 'a'+i), 30000+float64(i)*50))
	}
	g.form(ps[0], ps[1:8]...)
	if n := len(g.w.groupOf(ps[0].Char.ID)); n != 8 {
		t.Fatalf("group of %d, want 8", n)
	}
	g.w.Invite(ps[0], ps[8], "")
	if ls, _, _ := inbox(ps[0]); !has(ls, "full") {
		t.Errorf("inviting a ninth: %q", ls)
	}
	if _, _, inv := inbox(ps[8]); inv != 0 {
		t.Error("the ninth was invited into a full group")
	}
	g.w.Invite(ps[1], ps[8], "")
	if ls, _, _ := inbox(ps[1]); !has(ls, "Only the group's leader") {
		t.Errorf("a member inviting: %q", ls)
	}
	g.w.Invite(ps[8], ps[2], "")
	if ls, _, _ := inbox(ps[8]); !has(ls, "already in a group") {
		t.Errorf("inviting a grouped player: %q", ls)
	}
}

// The leader leaves a three-member group: the first to join leads, and both
// are told. Then one of two leaves and the group ends.
func TestGroupLeaderLeavesThenEnds(t *testing.T) {
	g := newGrouped(t)
	a, b, c := g.join("Ragnar", 30000), g.join("Sigrid", 30100), g.join("Bjorn", 30200)
	g.form(a, b, c)
	g.cmd(a, "disband")
	if got := fmt.Sprint(g.members(b)); got != "[Sigrid Bjorn]" {
		t.Fatalf("after the leader left: %s, want Sigrid leading", got)
	}
	for _, p := range []*Player{b, c} {
		if ls, _, _ := inbox(p); !has(ls, "Ragnar has left the group") || !has(ls, "Sigrid is now the group's leader") {
			t.Errorf("%s was told %q", p.Char.Name, ls)
		}
	}
	g.w.LeaveGroup(c.Char.ID)
	if g.w.groupOf(b.Char.ID) != nil {
		t.Fatal("a group of one still stands")
	}
	if ls, gs, _ := inbox(b); !has(ls, "disbanded") || len(gs) == 0 || len(gs[len(gs)-1].Members) != 0 {
		t.Errorf("the last member was told %q and sent %v", ls, gs)
	}
}

// The leader removes a member; a member who logs out leaves; one moving zone
// stays.
func TestGroupRemoveAndLogout(t *testing.T) {
	g := newGrouped(t)
	a, b, c := g.join("Ragnar", 30000), g.join("Sigrid", 30100), g.join("Bjorn", 30200)
	d := g.join("Astrid", 30300)
	g.form(a, b, c, d)
	g.w.RemoveMember(b, "Bjorn")
	if ls, _, _ := inbox(b); !has(ls, "Only the group's leader") {
		t.Errorf("a member removing: %q", ls)
	}
	g.cmd(a, "remove bjorn")
	if g.w.groupOf(c.Char.ID) != nil {
		t.Fatal("Bjorn is still grouped after the leader removed him")
	}
	if ls, _, _ := inbox(c); !has(ls, "removed from the group") {
		t.Errorf("Bjorn was told %q", ls)
	}
	g.w.Leave(d)
	if got := fmt.Sprint(g.members(a)); got != "[Ragnar Sigrid]" {
		t.Errorf("after Astrid logged out: %s", got)
	}
	b.zoning.Store(true)
	g.w.Leave(b)
	if len(g.w.groupOf(a.Char.ID)) != 2 {
		t.Error("a member changing zone left the group")
	}
}

// A member in another zone stays, and shows out of range with the zone named;
// a player outside the group is sent none of it.
func TestGroupStateGoesToMembersOnly(t *testing.T) {
	g := newGrouped(t)
	a, b := g.join("Ragnar", 30000), g.join("Sigrid", 30100)
	other := g.join("Stranger", 30050)
	g.form(a, b)
	b.hp = float64(b.maxHP()) / 2
	g.ticks(TickRate)
	_, gs, _ := inbox(a)
	if len(gs) == 0 || gs[len(gs)-1].Members[1].HP > 60 {
		t.Fatalf("Ragnar does not see Sigrid hurt: %v", gs)
	}
	if _, gs, _ := inbox(other); len(gs) != 0 {
		t.Errorf("a player outside the group was sent %v", gs)
	}
	// Sigrid walks into another zone, which publishes her there.
	delete(g.z.players, b.ID)
	g.w.groups.mu.Lock()
	st := g.w.groups.stats[b.Char.ID]
	st.zone = 101
	g.w.groups.stats[b.Char.ID] = st
	g.w.groups.mu.Unlock()
	g.w.Cat.Zones[101] = &ZoneInfo{ID: 101, Name: "Elsewhere"}
	g.ticks(TickRate)
	_, gs, _ = inbox(a)
	if len(gs) == 0 {
		t.Fatal("no update after Sigrid changed zone")
	}
	m := gs[len(gs)-1].Members[1]
	if m.InRange || m.Zone != 101 || m.ZoneN != "Elsewhere" || m.Gone {
		t.Errorf("Sigrid in another zone shows as %+v", m)
	}
}
