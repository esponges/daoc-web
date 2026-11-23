package world

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// FollowGap is how close /follow brings a player to who they follow.
const FollowGap = 120.0

// socialState is a player's friends and ignores and their anon, away and
// looking-for-group marks. Other zones read it when they route chat and tells,
// so it has a lock of its own.
type socialState struct {
	mu               sync.Mutex
	friends, ignores map[int64]bool // by character id
	anon, lfg        bool
	afk              string // the away message, "" when not away
	away             bool
}

func (p *Player) isFriend(char int64) bool {
	p.social.mu.Lock()
	defer p.social.mu.Unlock()
	return p.social.friends[char]
}

// ignoring says whether this player has marked the sender to be ignored.
func (p *Player) ignoring(from *Player) bool {
	p.social.mu.Lock()
	defer p.social.mu.Unlock()
	return p.social.ignores[from.Char.ID]
}

func (p *Player) awayNote() (string, bool) {
	p.social.mu.Lock()
	defer p.social.mu.Unlock()
	return p.social.afk, p.social.away
}

func (p *Player) clearAway() {
	p.social.mu.Lock()
	was := p.social.away
	p.social.away, p.social.afk = false, ""
	p.social.mu.Unlock()
	if was {
		sys(p, "You are no longer marked as away.")
	}
}

// loadSocial reads a character's friends and ignores from the store.
func (w *World) loadSocial(p *Player) {
	p.social.mu.Lock()
	defer p.social.mu.Unlock()
	p.social.friends, p.social.ignores = map[int64]bool{}, map[int64]bool{}
	for kind, m := range map[string]map[int64]bool{store.SocialFriend: p.social.friends, store.SocialIgnore: p.social.ignores} {
		ids, _ := w.Store.Social(p.Char.ID, kind)
		for _, id := range ids {
			m[id] = true
		}
	}
}

// sendFriends sends a player their friends list with who is in the game.
func (w *World) sendFriends(p *Player) {
	p.social.mu.Lock()
	ids := make([]int64, 0, len(p.social.friends))
	for id := range p.social.friends {
		ids = append(ids, id)
	}
	p.social.mu.Unlock()
	m := protocol.Friends{T: protocol.TFriends, List: []protocol.Friend{}}
	for _, id := range ids {
		m.List = append(m.List, protocol.Friend{Name: w.Store.CharName(id), On: w.onlineChar(id) != nil})
	}
	sort.Slice(m.List, func(i, j int) bool { return strings.ToLower(m.List[i].Name) < strings.ToLower(m.List[j].Name) })
	p.send(m)
}

// notifyFriends tells the online players who list p as a friend that p came
// or went, and refreshes their lists.
func (w *World) notifyFriends(p *Player, text string) {
	w.omu.Lock()
	var who []*Player
	for _, o := range w.online {
		if o != p && !o.closed.Load() && o.isFriend(p.Char.ID) {
			who = append(who, o)
		}
	}
	w.omu.Unlock()
	for _, o := range who {
		sys(o, p.Char.Name+" "+text)
		w.sendFriends(o)
	}
}

// friendCmd adds or removes a friend: /friend <name>, or /friend alone to list.
func (w *World) friendCmd(p *Player, name string) {
	if name == "" {
		w.sendFriends(p)
		p.social.mu.Lock()
		n := len(p.social.friends)
		p.social.mu.Unlock()
		sys(p, fmt.Sprintf("You have %d friends. Type /friend <name> to add or remove one.", n))
		return
	}
	id, stored := w.Store.CharID(name)
	if id == 0 {
		sys(p, "There is no character called "+name+".")
		return
	}
	if id == p.Char.ID {
		sys(p, "You are your own best friend already.")
		return
	}
	p.social.mu.Lock()
	if p.social.friends == nil {
		p.social.friends = map[int64]bool{}
	}
	on := !p.social.friends[id]
	if on {
		p.social.friends[id] = true
	} else {
		delete(p.social.friends, id)
	}
	p.social.mu.Unlock()
	w.Store.SetSocial(p.Char.ID, id, store.SocialFriend, on)
	if on {
		sys(p, stored+" is now your friend.")
	} else {
		sys(p, stored+" is no longer your friend.")
	}
	w.sendFriends(p)
}

// ignoreCmd adds or removes a character from the ignore list.
func (w *World) ignoreCmd(p *Player, name string) {
	if name == "" {
		sys(p, "Ignore whom? Type /ignore <name>; again to stop.")
		return
	}
	id, stored := w.Store.CharID(name)
	if id == 0 || id == p.Char.ID {
		sys(p, "You cannot ignore "+name+".")
		return
	}
	p.social.mu.Lock()
	if p.social.ignores == nil {
		p.social.ignores = map[int64]bool{}
	}
	on := !p.social.ignores[id]
	if on {
		p.social.ignores[id] = true
	} else {
		delete(p.social.ignores, id)
	}
	p.social.mu.Unlock()
	w.Store.SetSocial(p.Char.ID, id, store.SocialIgnore, on)
	if on {
		sys(p, "You now ignore "+stored+".")
	} else {
		sys(p, "You no longer ignore "+stored+".")
	}
}

// toggleAnon, toggleLFG and setAway flip the marks /who and tells use.
func (p *Player) toggleAnon() {
	p.social.mu.Lock()
	p.social.anon = !p.social.anon
	on := p.social.anon
	p.social.mu.Unlock()
	if on {
		sys(p, "You are now anonymous: /who shows only your name.")
	} else {
		sys(p, "You are no longer anonymous.")
	}
}

func (p *Player) toggleLFG() {
	p.social.mu.Lock()
	p.social.lfg = !p.social.lfg
	on := p.social.lfg
	p.social.mu.Unlock()
	if on {
		sys(p, "You are now looking for a group.")
	} else {
		sys(p, "You are no longer looking for a group.")
	}
}

func (p *Player) setAway(msg string) {
	p.social.mu.Lock()
	if p.social.away && msg == "" {
		p.social.away, p.social.afk = false, ""
		p.social.mu.Unlock()
		sys(p, "You are no longer marked as away.")
		return
	}
	if msg == "" {
		msg = "away from keyboard"
	}
	p.social.away, p.social.afk = true, msg
	p.social.mu.Unlock()
	sys(p, "You are now away: "+msg)
}

// who answers /who: the players in the zone, each with their level and class
// unless anonymous (to everyone but their group) and marked when looking for
// a group.
func (w *World) who(z *Zone, p *Player) {
	var names []string
	for _, o := range z.players {
		o.social.mu.Lock()
		anon, lfg := o.social.anon, o.social.lfg
		o.social.mu.Unlock()
		s := o.Char.Name
		if tag := guildLine(o); tag != "" {
			s += " " + tag
		}
		if !anon || o == p || w.sameGroup(p.Char.ID, o.Char.ID) {
			s += fmt.Sprintf(" (L%d %s)", o.Char.Level, o.sh.class.Name)
		}
		if lfg {
			s += " [LFG]"
		}
		names = append(names, s)
	}
	sort.Strings(names)
	sys(p, fmt.Sprintf("In %s: %s", z.Name, strings.Join(names, ", ")))
}

// Assist takes the target of the player you have targeted.
func (z *Zone) Assist(p *Player) {
	o := z.players[p.target]
	if o == nil || o == p {
		sys(p, "Target a player to assist.")
		return
	}
	if o.target == 0 {
		sys(p, o.Char.Name+" has no target.")
		return
	}
	z.SetTarget(p, o.target)
}

// Face turns the player toward their target.
func (z *Zone) Face(p *Player) {
	var x, y float64
	if o := z.players[p.target]; o != nil {
		x, y = o.Body.X, o.Body.Y
	} else if n := z.byID[p.target]; n != nil {
		x, y = n.Body.X, n.Body.Y
	} else {
		sys(p, "You have no target to face.")
		return
	}
	p.Body.Yaw = angleTo(p.Body.X, p.Body.Y, x, y)
	p.correct = true
}

// Follow walks the player behind the targeted player until they press a
// movement key, or give the command again.
func (z *Zone) Follow(p *Player) {
	if p.follow != 0 {
		p.follow = 0
		sys(p, "You stop following.")
		return
	}
	o := z.players[p.target]
	if o == nil || o == p || o.dead {
		sys(p, "Target a player to follow.")
		return
	}
	p.follow = o.ID
	sys(p, "You follow "+o.Char.Name+".")
}

// followStep moves a following player toward who they follow, with the same
// step maths as their own input. It returns whether it took the tick over.
func (z *Zone) followStep(p *Player) bool {
	if p.follow == 0 {
		return false
	}
	o := z.players[p.follow]
	moving := false
	for _, r := range p.queue {
		if r.k.Fwd != 0 || r.k.Strafe != 0 || r.k.Jump {
			moving = true
		}
	}
	if o == nil || o.dead || o.closed.Load() || moving || p.dead || p.ride != nil {
		if moving {
			sys(p, "You stop following.")
		}
		p.follow = 0
		return false
	}
	z.standUp(p)
	p.budget += float64(protocol.StepsPerSecond) * TickDT
	if p.budget > 20 {
		p.budget = 20
	}
	var last Motion
	moved := false
	for p.budget >= 1 {
		p.budget--
		if math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) <= FollowGap {
			break
		}
		p.Body.Yaw = angleTo(p.Body.X, p.Body.Y, o.Body.X, o.Body.Y)
		k := Keys{Fwd: 1, Run: true, Slow: 1 - speedFactor(p), Sprint: p.sprint}
		m := Step(z.G, z.Mv, &p.Body, k)
		z.landing(p, m)
		if p.dead {
			break
		}
		p.lastStepAt = z.now
		if m.Speed > 0 {
			moved, last = true, m
			p.movedSecond = true
		}
	}
	p.motion = Motion{}
	if moved {
		p.motion = last
	}
	z.settleAir(p)
	p.correct = true // the client does not predict a follow: it takes the server's place
	p.queue, p.queuedSteps = nil, 0
	return true
}

// Inspect shows what another player wears.
func (z *Zone) Inspect(p *Player) {
	o := z.players[p.target]
	if o == nil {
		sys(p, "Target a player to inspect.")
		return
	}
	if math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) > 2000 {
		sys(p, o.Char.Name+" is too far away to inspect.")
		return
	}
	sys(p, fmt.Sprintf("%s: level %d %s.", o.Char.Name, o.Char.Level, o.sh.class.Name))
	any := false
	for s := SlotWeapon; s < SlotCount; s++ {
		if it := o.inv[s]; it != nil {
			if x := z.W.Cat.Items[it.ID]; x != nil {
				sys(p, "  "+x.Name)
				any = true
			}
		}
	}
	if !any {
		sys(p, "  nothing worn.")
	}
}
