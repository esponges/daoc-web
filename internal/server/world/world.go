package world

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// Config says where the world's data is.
type Config struct {
	WebData string // converted zones, web/data
	Cat     *Catalog
	Store   *store.Store
	Log     func(format string, a ...any)
	// Mover builds the collision for an outdoor zone; nil means none.
	Collision func(webData string, zone int) Mover
	// DefaultZone is where a character with nowhere better starts.
	DefaultZone int
	// IdleUnload is how long an empty zone keeps running before it is
	// saved and unloaded, in seconds (20 if zero).
	IdleUnload float64
}

// World is the whole running game.
type World struct {
	Config
	mu    sync.Mutex
	zones map[int]*Zone

	omu    sync.Mutex
	online map[int64]*Player // by character id

	dropMu sync.Mutex
	drops  map[int][]string

	groups *groups
	rides  rides // riders in the middle of a zone change
}

// ErrOnline is returned when a character is already playing.
var ErrOnline = errors.New("that character is already playing")

// New builds a world.
func New(c Config) *World {
	if c.Log == nil {
		c.Log = func(string, ...any) {}
	}
	if c.DefaultZone == 0 {
		c.DefaultZone = 100
	}
	return &World{Config: c, zones: map[int]*Zone{}, online: map[int64]*Player{}, drops: map[int][]string{}, groups: newGroups()}
}

// Zone returns the running zone, starting it if needed.
func (w *World) zoneLocked(id int) (*Zone, error) {
	if z := w.zones[id]; z != nil {
		return z, nil
	}
	info := w.Cat.Zones[id]
	if info == nil || !info.Converted {
		return nil, fmt.Errorf("zone %d is not available", id)
	}
	g, name, err := LoadTerrain(w.WebData, id)
	var ground Ground = g
	if err != nil {
		ig, n2, ierr := LoadInterior(w.WebData, id)
		if ierr != nil {
			return nil, err
		}
		ground, name = ig, n2
	}
	spawns, err := w.Cat.SpawnsIn(id)
	if err != nil {
		return nil, err
	}
	var mv Mover
	if w.Collision != nil {
		mv = w.Collision(w.WebData, id)
	}
	if im, ok := ground.(Mover); ok && mv == nil {
		mv = im
	}
	z := NewZone(w, id, name, ground, mv, spawns)
	w.zones[id] = z
	z.Start()
	w.Log("zone %d %s started with %d spawns", id, name, len(spawns))
	return z, nil
}

// ZoneRunning reports the running zones, for tests and the status page.
func (w *World) ZoneRunning() []int {
	w.mu.Lock()
	defer w.mu.Unlock()
	var ids []int
	for id := range w.zones {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

// tryUnload stops an empty zone. A join that has been queued for it vetoes.
func (w *World) tryUnload(z *Zone) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(z.cmds) > 0 || len(z.players) > 0 {
		return false
	}
	delete(w.zones, z.ID)
	w.Log("zone %d unloaded", z.ID)
	return true
}

// Shutdown stops every zone, saving the players in them.
func (w *World) Shutdown() {
	w.mu.Lock()
	zs := make([]*Zone, 0, len(w.zones))
	for _, z := range w.zones {
		zs = append(zs, z)
	}
	w.zones = map[int]*Zone{}
	w.mu.Unlock()
	for _, z := range zs {
		z.Stop()
	}
}

// --- joining and leaving ---

// Join puts a character into the world. The returned player's Out channel
// carries the messages for its connection.
func (w *World) Join(acct *store.Account, charID int64) (*Player, error) {
	ch, err := w.Store.Character(acct.ID, charID)
	if err != nil {
		return nil, err
	}
	p := newPlayer(acct, ch)
	w.leaveUnsafeStart(p)
	w.omu.Lock()
	if old := w.online[ch.ID]; old != nil && !old.closed.Load() {
		if !old.disconnected.Load() {
			w.omu.Unlock()
			return nil, ErrOnline
		}
		// Their old session is gone but the character lingers in a fight:
		// the player coming back takes it over.
		w.omu.Unlock()
		w.evict(old)
		w.omu.Lock()
	}
	w.online[ch.ID] = p
	w.omu.Unlock()

	w.mu.Lock()
	z, err := w.zoneLocked(ch.Zone)
	if err != nil {
		// The saved zone is gone (or was never converted): start again at the beginning.
		st, _ := w.Cat.StartFor(p.Arch.Realm, p.Arch.Race, p.Arch.Class)
		ch.Zone, ch.X, ch.Y, ch.Z = st.Zone, st.X, st.Y, st.Z
		if ch.Zone == 0 {
			ch.Zone = w.DefaultZone
		}
		z, err = w.zoneLocked(ch.Zone)
	}
	if err != nil {
		w.mu.Unlock()
		w.forget(p)
		return nil, err
	}
	ready := make(chan struct{})
	ok := z.Do(func() { z.join(p); close(ready) })
	w.mu.Unlock()
	if ok {
		select {
		case <-ready:
		case <-time.After(5 * time.Second):
			ok = false
		}
	}
	if !ok {
		w.forget(p)
		return nil, errors.New("zone is shutting down; try again")
	}
	return p, nil
}

func (w *World) forget(p *Player) {
	w.omu.Lock()
	if w.online[p.Char.ID] == p {
		delete(w.online, p.Char.ID)
	}
	w.omu.Unlock()
}

// Leave is called when a connection closes.
func (w *World) Leave(p *Player) {
	p.disconnected.Store(true)
	if !p.zoning.Load() {
		w.logoutGroup(p.Char.ID)
		w.rides.take(p.Char.ID)
	}
	z := p.Zone
	if z == nil {
		w.forget(p)
		return
	}
	z.Do(func() {
		if p.closed.Load() {
			return
		}
		if z.now < p.combatUntil && !p.dead {
			p.graceEnd = z.now + DisconnectGrace
			p.out = make(chan []byte, 1) // nobody is listening
			return
		}
		w.savePlayer(p)
		p.closed.Store(true)
	})
	// The character may play again as soon as the connection is gone, even
	// while a fighter lingers in the world.
	w.forgetLater(p)
}

func (w *World) forgetLater(p *Player) {
	go func() {
		for i := 0; i < 400; i++ {
			if p.closed.Load() {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		w.forget(p)
	}()
}

// join runs in the zone goroutine.
func (z *Zone) join(p *Player) {
	w := z.W
	c := p.Char
	if c.X == 0 && c.Y == 0 {
		st, _ := w.Cat.StartFor(p.Arch.Realm, p.Arch.Race, p.Arch.Class)
		if st.Zone == z.ID {
			c.X, c.Y, c.Z, c.Heading = st.X, st.Y, st.Z, st.Heading
		} else {
			c.X, c.Y = z.G.Extent()/2, z.G.Extent()/2
		}
	}
	x, y := z.G.Clamp(c.X, c.Y)
	p.Body = Body{X: x, Y: y, Z: z.G.GroundAt(x, y, c.Z), Yaw: float64(c.Heading) / 1000}
	z.settleBind(p)
	if prepareCharacter(c) {
		if err := w.Store.Save(c); err != nil {
			w.Log("saving %s after bringing it up to the rule book: %v", c.Name, err)
		}
	}
	p.end = float64(rules.MaxEndurance)
	w.loadInventory(p) // computes the sheet, and so the hit points
	if c.HP <= 0 {
		c.HP = p.maxHP()
	}
	p.hp = math.Min(float64(c.HP), float64(p.maxHP()))
	z.computeBar(p)
	z.resumeRide(p)
	z.add(p)
	p.exitLatched = z.exitAt(p) != nil
	p.send(protocol.Welcome{T: protocol.TWelcome, V: protocol.Version, You: p.ID, Name: c.Name, Char: p.Arch.Key, Race: c.Race, Female: c.Gender == 1,
		Zone: z.ID, X: p.Body.X, Y: p.Body.Y, Z: p.Body.Z, Yaw: p.Body.Yaw, Level: c.Level, Admin: p.Admin})
	p.welcomed = true
	p.statsSent = p.statsMsg(w.Cat)
	p.send(p.statsSent)
	p.sendInv(w.Cat, c.Money)
	p.sendBar(w.Cat)
	z.event(protocol.Event{K: "join", To: p.ID, Text: c.Name}, p.Body.X, p.Body.Y)
	if p.notice != "" {
		p.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "sys", Text: p.notice})
		p.notice = ""
	}
}

// savePlayer writes a character to the database.
func (w *World) savePlayer(p *Player) {
	c := p.Char
	if p.Zone != nil {
		c.Zone = p.Zone.ID
	}
	c.X, c.Y, c.Z = p.Body.X, p.Body.Y, p.Body.Z
	c.Heading = int(p.Body.Yaw * 1000)
	c.HP = int(math.Ceil(p.hp))
	c.Endurance = int(p.end)
	if err := w.Store.Save(c); err != nil {
		w.Log("saving %s: %v", c.Name, err)
	}
}

// --- places ---

// bindFor is where a dead player returns to: their bind point, or for a
// character not yet bound, the nearest bind point in the zone or their start.
func (w *World) bindFor(p *Player) (zone int, x, y, z, yaw float64) {
	if b := p.Char.Bind; b != nil {
		return b.Zone, b.X, b.Y, b.Z, float64(b.Heading) / 1000
	}
	best := math.Inf(1)
	for _, b := range w.Cat.Binds {
		if b.Zone != p.Zone.ID || (b.Realm != 0 && b.Realm != p.Arch.Realm) {
			continue
		}
		if d := math.Hypot(b.X-p.Body.X, b.Y-p.Body.Y); d < best {
			best, zone, x, y, z = d, b.Zone, b.X, b.Y, b.Z
		}
	}
	if best < math.Inf(1) {
		return
	}
	if st, ok := w.Cat.StartFor(p.Arch.Realm, p.Arch.Race, p.Arch.Class); ok {
		return st.Zone, st.X, st.Y, st.Z, float64(st.Heading) / 1000
	}
	e := p.Zone.G.Extent() / 2
	return p.Zone.ID, e, e, 0, 0
}

// travel takes a player through an exit if its far side exists.
func (w *World) travel(z *Zone, p *Player, e *Exit) {
	info := w.Cat.Zones[e.DstZone]
	if info == nil || !info.Converted {
		p.noExitUntil = z.now + 8
		p.errMsg("zone", "The way is shut: that place is not part of this world yet.")
		return
	}
	w.moveToZone(z, p, e.DstZone, e.TX, e.TY, e.TZ)
}

// moveToZone saves a player at a place in another zone and tells the client
// to load it; the client reconnects there.
func (w *World) moveToZone(z *Zone, p *Player, dst int, x, y, zz float64) {
	c := p.Char
	c.Zone, c.X, c.Y, c.Z = dst, x, y, zz
	c.HP, c.Endurance = int(math.Ceil(p.hp)), int(p.end)
	p.Body.X, p.Body.Y = x, y
	if err := w.Store.Save(c); err != nil {
		w.Log("saving %s: %v", c.Name, err)
	}
	p.send(protocol.ZoneChange{T: protocol.TZone, Zone: dst, X: x, Y: y})
	p.zoning.Store(true)
	p.closed.Store(true)
	w.forget(p)
}

// --- chat ---

// Chat delivers a line from a player.
func (w *World) Chat(p *Player, ch, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if len(text) > 200 {
		text = text[:200]
	}
	z := p.Zone
	if z == nil {
		return
	}
	now := z.now
	// A token bucket: a line a second, a burst of four.
	p.chatTokens = math.Min(4, p.chatTokens+(now-p.chatAt))
	p.chatAt = now
	if p.chatTokens < 1 {
		p.errMsg("chat", "You are talking too fast.")
		return
	}
	p.chatTokens--
	if strings.HasPrefix(text, "/") {
		w.command(z, p, text[1:])
		return
	}
	m := protocol.ChatMsg{T: protocol.TChatMsg, Ch: ch, From: p.Char.Name, Text: text}
	if ch == "all" {
		w.mu.Lock()
		zs := make([]*Zone, 0, len(w.zones))
		for _, zz := range w.zones {
			zs = append(zs, zz)
		}
		w.mu.Unlock()
		for _, zz := range zs {
			if zz == z {
				for _, o := range z.players {
					o.send(m)
				}
				continue
			}
			zz.Do(func() {
				for _, o := range zz.players {
					o.send(m)
				}
			})
		}
		return
	}
	m.Ch = "say"
	for _, o := range z.players {
		if math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) <= SayRange {
			o.send(m)
		}
	}
}

// command runs a slash command typed in chat; the GM ones need an administrator.
func (w *World) command(z *Zone, p *Player, line string) {
	f := strings.Fields(line)
	if len(f) == 0 {
		return
	}
	switch f[0] {
	case "who":
		var names []string
		for _, o := range z.players {
			names = append(names, o.Char.Name)
		}
		sort.Strings(names)
		p.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "sys", From: "", Text: fmt.Sprintf("In %s: %s", z.Name, strings.Join(names, ", "))})
	case "gm":
		w.GM(p, strings.Join(f[1:], " "))
	case "pray":
		z.Pray(p)
	case "bind":
		z.BindHere(p)
	case "send", "tell":
		if len(f) < 3 {
			sys(p, "Send to whom? Type /"+f[0]+" <name> <text>.")
			return
		}
		w.tell(p, f[1], restAfter(line, 2))
	case "g":
		if len(f) < 2 {
			return
		}
		w.groupSay(p, restAfter(line, 1))
	case "invite":
		if len(f) > 1 {
			w.Invite(p, w.onlineByName(f[1]), f[1])
		} else {
			w.GroupOp(p, &protocol.GroupOp{Op: protocol.GroupInvite, ID: p.target})
		}
	case "guard", "protect", "intercept", "engage":
		z.UseAbility(p, f[0], f[1:]...)
	case "accept":
		w.AcceptInvite(p)
	case "decline":
		w.DeclineInvite(p)
	case "disband", "leave":
		w.LeaveGroup(p.Char.ID)
	case "remove":
		if len(f) < 2 {
			sys(p, "Remove whom? Type /remove <name>.")
			return
		}
		w.RemoveMember(p, f[1])
	default:
		p.errMsg("chat", "Unknown command /"+f[0]+".")
	}
}

// GM runs an administrator's command.
func (w *World) GM(p *Player, line string) {
	z := p.Zone
	if z == nil {
		return
	}
	say := func(s string) {
		p.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "sys", Text: s})
	}
	if !p.Admin {
		p.errMsg("gm", "Only an administrator can do that.")
		return
	}
	f := strings.Fields(line)
	if len(f) == 0 {
		say("gm: goto <zone> | tp <x> <y> | level <n> | heal | give <item> [n] | kill | where")
		return
	}
	num := func(i int) (float64, bool) {
		if i >= len(f) {
			return 0, false
		}
		var v float64
		_, err := fmt.Sscanf(f[i], "%g", &v)
		return v, err == nil
	}
	switch f[0] {
	case "goto":
		v, ok := num(1)
		info := w.Cat.Zones[int(v)]
		if !ok || info == nil || !info.Converted {
			say("no such zone")
			return
		}
		st, _ := w.Cat.StartFor(2, 0, 0)
		e := 32768.0
		if int(v) == st.Zone {
			w.moveToZone(z, p, int(v), st.X, st.Y, st.Z)
			return
		}
		w.moveToZone(z, p, int(v), e, e, 0)
	case "tp":
		x, ok1 := num(1)
		y, ok2 := num(2)
		if !ok1 || !ok2 {
			say("tp <x> <y> in zone units")
			return
		}
		x, y = z.G.Clamp(x, y)
		p.Body = Body{X: x, Y: y, Z: z.G.GroundAt(x, y, 0), Yaw: p.Body.Yaw}
		p.correct = true
	case "level":
		v, ok := num(1)
		if !ok || v < 1 || v > MaxLevel {
			say("level 1.." + fmt.Sprint(MaxLevel))
			return
		}
		p.Char.Level = int(v)
		p.Char.XP = rules.XPToReach(int(v))
		w.recalc(p)
		p.hp = float64(p.maxHP())
		z.buildBar(p)
		w.savePlayer(p)
	case "heal":
		p.hp, p.end = float64(p.maxHP()), float64(rules.MaxEndurance)
	case "give":
		if len(f) < 2 {
			say("give <item id> [count]")
			return
		}
		n := 1
		if v, ok := num(2); ok {
			n = int(v)
		}
		if _, err := w.addItem(p, f[1], n); err != nil {
			say(err.Error())
		}
	case "kill":
		if n := z.byID[p.target]; n != nil && !n.dead() {
			n.hurtBy[p.ID] += n.HP
			n.HP = 0
			z.killNPC(n)
		}
	case "where":
		say(fmt.Sprintf("zone %d (%s) at %.0f, %.0f, %.0f", z.ID, z.Name, p.Body.X, p.Body.Y, p.Body.Z))
	default:
		say("unknown gm command")
	}
}

// CreationStats checks a creation build (how far each stat is raised, in the
// order strength, constitution, dexterity, quickness, intelligence, piety,
// empathy, charisma) and returns the character's base stats. A nil build is
// the class's default.
func CreationStats(race, class int, build *[8]int) ([8]int, error) {
	cl, ok := rules.Classes[class]
	r, rok := rules.Races[race]
	if !ok || !rok {
		return [8]int{}, rules.ErrUnknown
	}
	b := rules.DefaultBuild(cl)
	if build != nil {
		b = rules.Stats(*build)
	}
	st, err := rules.Create(r, cl, b)
	return [8]int(st), err
}

// StartHP is a new level-1 character's hit points, from their class and
// constitution.
func StartHP(class int, stats [8]int) int {
	return rules.MaxHP(classOf(class), 1, stats[rules.CON], 0)
}

// RequestStyle, TakeLoot and Pong are the entry points the connection uses.
func (z *Zone) RequestStyle(p *Player, slot int)      { z.requestStyle(p, slot) }
func (z *Zone) TakeLoot(p *Player, id int64, idx int) { z.takeLoot(p, id, idx) }

// Pong answers a client's ping.
func (p *Player) Pong() { p.send(protocol.Envelope{T: protocol.TPong}) }

// evict saves and removes a lingering character, waiting briefly for the
// zone to do it.
func (w *World) evict(p *Player) {
	if z := p.Zone; z != nil {
		z.Do(func() {
			w.savePlayer(p)
			p.closed.Store(true)
		})
	}
	for i := 0; i < 100 && !p.closed.Load(); i++ {
		time.Sleep(20 * time.Millisecond)
	}
}
