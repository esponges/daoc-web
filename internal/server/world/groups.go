package world

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// Group limits.
const (
	GroupMax    = 8                // members in a group
	InviteLapse = 60 * time.Second // an unanswered invitation lapses
	GroupRange  = 6000.0           // a member this close, in the same zone, is in range
)

// group is players of one realm fighting together. Members are kept by
// character, so a member stays in it across zones; the first leads, the rest
// are in the order they joined.
type group struct {
	members []groupSeat
}

type groupSeat struct {
	char int64
	name string
}

func (g *group) leader() int64 { return g.members[0].char }

func (g *group) has(char int64) bool { return g.index(char) >= 0 }

func (g *group) index(char int64) int {
	for i, m := range g.members {
		if m.char == char {
			return i
		}
	}
	return -1
}

// invite is a standing invitation, by the one who sent it.
type invite struct {
	from    int64
	name    string // the inviter's name
	to      string // the invited player's name, for the lapse line
	expires time.Time
}

// memberStat is a member's state as their zone last published it.
type memberStat struct {
	name    string
	level   int
	class   string
	hp, end int
	zone    int
	x, y    float64
	dead    bool
}

// groups is the world's groups, under their own lock: zones call in from
// their own goroutines, and the lock is never held while taking another.
type groups struct {
	mu      sync.Mutex
	of      map[int64]*group  // by member character id
	invites map[int64]*invite // by the invited character id
	stats   map[int64]memberStat
	sent    map[int64]protocol.Group // the last group message each member was sent
	clock   func() time.Time
}

func newGroups() *groups {
	return &groups{of: map[int64]*group{}, invites: map[int64]*invite{}, stats: map[int64]memberStat{},
		sent: map[int64]protocol.Group{}, clock: time.Now}
}

// onlineByName finds a character in the game by name, without regard to case.
func (w *World) onlineByName(name string) *Player {
	w.omu.Lock()
	defer w.omu.Unlock()
	for _, p := range w.online {
		if strings.EqualFold(p.Char.Name, name) && !p.closed.Load() {
			return p
		}
	}
	return nil
}

// onlineChar is the character's live player, nil when they are not in the game.
func (w *World) onlineChar(char int64) *Player {
	w.omu.Lock()
	defer w.omu.Unlock()
	if p := w.online[char]; p != nil && !p.closed.Load() {
		return p
	}
	return nil
}

func sys(p *Player, text string) {
	if p != nil {
		p.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "sys", Text: text})
	}
}

// GroupOp runs a client's group request; it is called in the player's zone.
func (w *World) GroupOp(p *Player, m *protocol.GroupOp) {
	switch m.Op {
	case protocol.GroupInvite:
		var to *Player
		if m.ID != 0 && p.Zone != nil {
			to = p.Zone.players[m.ID]
		} else if m.Name != "" {
			to = w.onlineByName(m.Name)
		}
		w.Invite(p, to, m.Name)
	case protocol.GroupAccept:
		w.AcceptInvite(p)
	case protocol.GroupDecline:
		w.DeclineInvite(p)
	case protocol.GroupLeave:
		w.LeaveGroup(p.Char.ID)
	case protocol.GroupRemove:
		w.RemoveMember(p, m.Name)
	}
}

// Invite asks another player to join the inviter's group, or to start one.
// name is what the inviter typed, for the line when nobody answers to it.
func (w *World) Invite(p, to *Player, name string) {
	if to == nil {
		if name == "" {
			sys(p, "Invite whom? Target a player or type /invite <name>.")
		} else {
			sys(p, name+" is not in the game.")
		}
		return
	}
	if to == p || to.Char.ID == p.Char.ID {
		sys(p, "You cannot invite yourself.")
		return
	}
	if to.Arch.Realm != p.Arch.Realm {
		sys(p, "You cannot group with a player of another realm.")
		return
	}
	gs := w.groups
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if g := gs.of[p.Char.ID]; g != nil {
		if g.leader() != p.Char.ID {
			sys(p, "Only the group's leader can invite.")
			return
		}
		if len(g.members) >= GroupMax {
			sys(p, "Your group is full.")
			return
		}
	}
	if gs.of[to.Char.ID] != nil {
		sys(p, to.Char.Name+" is already in a group.")
		return
	}
	if inv := gs.invites[to.Char.ID]; inv != nil && inv.from != p.Char.ID && gs.clock().Before(inv.expires) {
		sys(p, to.Char.Name+" is considering another invitation.")
		return
	}
	gs.invites[to.Char.ID] = &invite{from: p.Char.ID, name: p.Char.Name, to: to.Char.Name, expires: gs.clock().Add(InviteLapse)}
	to.send(protocol.GroupInv{T: protocol.TGroupInv, From: p.Char.Name, Secs: InviteLapse.Seconds()})
	sys(to, p.Char.Name+" has invited you to join a group. Accept or decline.")
	sys(p, "You have invited "+to.Char.Name+" to join your group.")
}

// AcceptInvite puts a player in the group they were invited to, starting it
// with the inviter as leader if there was none.
func (w *World) AcceptInvite(p *Player) {
	gs := w.groups
	gs.mu.Lock()
	defer gs.mu.Unlock()
	inv := gs.invites[p.Char.ID]
	delete(gs.invites, p.Char.ID)
	if inv == nil || !gs.clock().Before(inv.expires) {
		sys(p, "You have no invitation to accept.")
		return
	}
	from := w.onlineChar(inv.from)
	if from == nil {
		sys(p, inv.name+" is no longer in the game.")
		return
	}
	if gs.of[p.Char.ID] != nil {
		sys(p, "You are already in a group.")
		return
	}
	g := gs.of[inv.from]
	switch {
	case g == nil:
		g = &group{members: []groupSeat{{inv.from, inv.name}}}
		gs.of[inv.from] = g
	case g.leader() != inv.from:
		sys(p, inv.name+" no longer leads a group.")
		return
	case len(g.members) >= GroupMax:
		sys(p, "That group is full.")
		return
	}
	g.members = append(g.members, groupSeat{p.Char.ID, p.Char.Name})
	gs.of[p.Char.ID] = g
	for _, m := range g.members {
		if m.char == p.Char.ID {
			sys(p, "You have joined "+inv.name+"'s group.")
		} else {
			sys(w.onlineChar(m.char), p.Char.Name+" has joined the group.")
		}
	}
	w.sendGroupLocked(g)
}

// DeclineInvite turns an invitation down and tells the inviter.
func (w *World) DeclineInvite(p *Player) {
	gs := w.groups
	gs.mu.Lock()
	defer gs.mu.Unlock()
	inv := gs.invites[p.Char.ID]
	if inv == nil {
		return
	}
	delete(gs.invites, p.Char.ID)
	sys(w.onlineChar(inv.from), p.Char.Name+" declines your invitation.")
}

// LeaveGroup takes a character out of their group (/disband, or logging out).
func (w *World) LeaveGroup(char int64) {
	gs := w.groups
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if gs.of[char] == nil {
		sys(w.onlineChar(char), "You are not in a group.")
		return
	}
	w.dropLocked(char, "%s has left the group.", "You leave the group.")
}

// logoutGroup takes a character who has left the game out of their group,
// and drops any invitation they sent or were sent.
func (w *World) logoutGroup(char int64) {
	gs := w.groups
	gs.mu.Lock()
	defer gs.mu.Unlock()
	delete(gs.invites, char)
	for to, inv := range gs.invites {
		if inv.from == char {
			delete(gs.invites, to)
		}
	}
	if gs.of[char] != nil {
		w.dropLocked(char, "%s has left the group.", "You leave the group.")
	}
}

// RemoveMember is the leader taking a member out.
func (w *World) RemoveMember(p *Player, name string) {
	gs := w.groups
	gs.mu.Lock()
	defer gs.mu.Unlock()
	g := gs.of[p.Char.ID]
	if g == nil || g.leader() != p.Char.ID {
		sys(p, "Only the group's leader can remove a member.")
		return
	}
	for _, m := range g.members {
		if strings.EqualFold(m.name, name) && m.char != p.Char.ID {
			w.dropLocked(m.char, "%s has been removed from the group.", "You have been removed from the group.")
			return
		}
	}
	sys(p, name+" is not in your group.")
}

// dropLocked removes a member, hands on the lead, and ends a group left with
// one. others is the line for the rest (with the name), self the leaver's.
func (w *World) dropLocked(char int64, others, self string) {
	gs := w.groups
	g := gs.of[char]
	i := g.index(char)
	name := g.members[i].name
	g.members = append(g.members[:i], g.members[i+1:]...)
	delete(gs.of, char)
	delete(gs.sent, char)
	if p := w.onlineChar(char); p != nil {
		sys(p, self)
		p.send(protocol.Group{T: protocol.TGroup, Members: []protocol.GroupMember{}})
	}
	if len(g.members) == 1 {
		last := g.members[0].char
		delete(gs.of, last)
		delete(gs.sent, last)
		if p := w.onlineChar(last); p != nil {
			sys(p, fmt.Sprintf(others, name)+" Your group has disbanded.")
			p.send(protocol.Group{T: protocol.TGroup, Members: []protocol.GroupMember{}})
		}
		return
	}
	for _, m := range g.members {
		o := w.onlineChar(m.char)
		sys(o, fmt.Sprintf(others, name))
		if i == 0 {
			sys(o, g.members[0].name+" is now the group's leader.")
		}
	}
	w.sendGroupLocked(g)
}

// groupOf is the members of a character's group by character id, leader
// first, or nil when they are not in one.
func (w *World) groupOf(char int64) []int64 {
	gs := w.groups
	gs.mu.Lock()
	defer gs.mu.Unlock()
	g := gs.of[char]
	if g == nil {
		return nil
	}
	out := make([]int64, len(g.members))
	for i, m := range g.members {
		out[i] = m.char
	}
	return out
}

// groupTick runs once a second in each zone: it publishes the zone's
// players' state for their groups, sends members what changed, and lets
// lapsed invitations go.
func (z *Zone) groupTick() {
	w := z.W
	gs := w.groups
	gs.mu.Lock()
	defer gs.mu.Unlock()
	for _, p := range z.players {
		if p.closed.Load() {
			continue
		}
		cls := ""
		if c, ok := rules.Classes[p.Char.Class]; ok {
			cls = c.Name
		}
		gs.stats[p.Char.ID] = memberStat{name: p.Char.Name, level: p.Char.Level, class: cls,
			hp: pct(p.hp, float64(p.maxHP())), end: pct(p.end, rules.MaxEndurance),
			zone: z.ID, x: p.Body.X, y: p.Body.Y, dead: p.dead}
	}
	now := gs.clock()
	for char, inv := range gs.invites {
		if !now.Before(inv.expires) {
			delete(gs.invites, char)
			sys(w.onlineChar(inv.from), inv.to+" declines your invitation.")
		}
	}
	for _, p := range z.players {
		if g := gs.of[p.Char.ID]; g != nil && !p.closed.Load() {
			w.sendMemberLocked(g, p)
		}
	}
}

func pct(v, max float64) int {
	if max <= 0 {
		return 0
	}
	return int(math.Round(100 * math.Max(0, math.Min(v, max)) / max))
}

// sendGroupLocked sends every member in the game their view of the group.
func (w *World) sendGroupLocked(g *group) {
	for _, m := range g.members {
		if p := w.onlineChar(m.char); p != nil {
			w.sendMemberLocked(g, p)
		}
	}
}

// sendMemberLocked sends one member the group as they see it, when it has
// changed since the last.
func (w *World) sendMemberLocked(g *group, to *Player) {
	gs := w.groups
	me := gs.stats[to.Char.ID]
	msg := protocol.Group{T: protocol.TGroup, Members: make([]protocol.GroupMember, 0, len(g.members))}
	for _, m := range g.members {
		v := protocol.GroupMember{Name: m.name, You: m.char == to.Char.ID}
		st, ok := gs.stats[m.char]
		if w.onlineChar(m.char) == nil {
			v.Gone = true
		} else if ok {
			v.Level, v.Class, v.HP, v.End, v.Zone, v.Dead = st.level, st.class, st.hp, st.end, st.zone, st.dead
			if info := w.Cat.Zones[st.zone]; info != nil {
				v.ZoneN = info.Name
			}
			v.InRange = st.zone == me.zone && math.Hypot(st.x-me.x, st.y-me.y) <= GroupRange
		}
		msg.Members = append(msg.Members, v)
	}
	if last, ok := gs.sent[to.Char.ID]; ok && reflect.DeepEqual(last, msg) {
		return
	}
	gs.sent[to.Char.ID] = msg
	to.send(msg)
}
