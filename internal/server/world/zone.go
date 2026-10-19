package world

import (
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// Zone is one running place. A single goroutine owns everything in it, so
// the rules need no locks; the rest of the server reaches in through Do.
type Zone struct {
	W    *World
	ID   int
	Name string
	G    Ground
	Mv   Mover

	cmds    chan func()
	players map[int64]*Player
	npcs    []*NPC
	byID    map[int64]*NPC
	tick    int64
	now     float64 // simulated seconds since the zone started
	ev      []evAt
	rng     *rand.Rand
	exits   []Exit
	emptyAt float64
	graves  map[int64]*grave // by owner character id

	stop    chan struct{}
	done    chan struct{}
	started sync.Once
}

type evAt struct {
	e    protocol.Event
	x, y float64
}

// NewZone builds a zone around a ground and its spawns. It is not running
// until Start; tests drive it with Tick.
func NewZone(w *World, id int, name string, g Ground, mv Mover, spawns []Spawn) *Zone {
	z := &Zone{
		W: w, ID: id, Name: name, G: g, Mv: mv,
		cmds: make(chan func(), 256), players: map[int64]*Player{}, byID: map[int64]*NPC{},
		rng: rand.New(rand.NewSource(int64(id)*7919 + 1)), stop: make(chan struct{}), done: make(chan struct{}),
	}
	for _, sp := range spawns {
		n := newNPC(sp, w.Cat, rand.New(rand.NewSource(sp.ID*31+int64(id))))
		n.Body.Z = g.GroundAt(n.Body.X, n.Body.Y, n.Body.Z)
		z.npcs = append(z.npcs, n)
		z.byID[n.ID] = n
	}
	z.loadGraves()
	if w.Cat != nil {
		for _, e := range w.Cat.Exits {
			if e.SrcZone == id {
				z.exits = append(z.exits, e)
			}
		}
	}
	return z
}

// Start runs the zone's tick loop until Stop.
func (z *Zone) Start() {
	z.started.Do(func() {
		go func() {
			defer close(z.done)
			t := time.NewTicker(time.Second / TickRate)
			defer t.Stop()
			for {
				select {
				case <-z.stop:
					z.drain()
					return
				case <-t.C:
					z.Tick()
					if z.idleTooLong() && z.W.tryUnload(z) {
						return
					}
				}
			}
		}()
	})
}

// Stop ends the loop and saves everyone in the zone.
func (z *Zone) Stop() {
	select {
	case <-z.stop:
	default:
		close(z.stop)
	}
	<-z.done
}

// Do runs fn in the zone's goroutine. It returns false if the zone has gone.
func (z *Zone) Do(fn func()) bool {
	select {
	case <-z.done:
		return false
	case z.cmds <- fn:
		return true
	}
}

func (z *Zone) drain() {
	for {
		select {
		case fn := <-z.cmds:
			fn()
		default:
			for _, p := range z.players {
				z.W.savePlayer(p)
			}
			return
		}
	}
}

func (z *Zone) idleTooLong() bool {
	if len(z.players) > 0 {
		z.emptyAt = z.now
		return false
	}
	limit := z.W.IdleUnload
	if limit <= 0 {
		limit = 20
	}
	return z.now-z.emptyAt > limit
}

// Tick advances the zone one step: inputs, creatures, fights, then what each
// player sees.
func (z *Zone) Tick() {
	for {
		select {
		case fn := <-z.cmds:
			fn()
			continue
		default:
		}
		break
	}
	z.tick++
	z.now += TickDT
	for _, p := range z.players {
		z.tickPlayer(p)
	}
	for _, n := range z.npcs {
		n.update(z)
	}
	z.tickEffects()
	z.snapshots()
	for id, p := range z.players {
		if p.closed.Load() {
			delete(z.players, id)
		}
	}
	z.ev = z.ev[:0]
}

// event records something that happened, for those near it.
func (z *Zone) event(e protocol.Event, x, y float64) {
	z.ev = append(z.ev, evAt{e, x, y})
}

// Players returns the number of players in the zone.
func (z *Zone) Players() int { return len(z.players) }

// NPCs returns a copy of the creature list, for tests and tools.
func (z *Zone) NPCs() []*NPC { return z.npcs }

// Now is the zone's clock.
func (z *Zone) Now() float64 { return z.now }

// add puts a player into the zone.
func (z *Zone) add(p *Player) {
	p.Zone = z
	z.players[p.ID] = p
	p.shop = 0 // a shop never follows a character into another zone
}

func (z *Zone) tickPlayer(p *Player) {
	if p.closed.Load() {
		return
	}
	if p.dead {
		if z.now-p.deadAt >= PlayerRespawnS {
			z.respawnPlayer(p)
		}
		return
	}
	p.applyInput(z)
	z.updatePrayer(p)
	z.checkShop(p)
	z.combatPlayer(p)
	// Vitals come back: endurance every second, health every few seconds and
	// only out of a fight.
	if z.now >= p.regenAt {
		p.regenAt = z.now + 1
		fighting := z.now < p.combatUntil
		gain := float64(rules.EnduranceRegenOut)
		if fighting {
			gain = rules.EnduranceRegenIn
		}
		p.end = math.Min(float64(rules.MaxEndurance), p.end+gain)
	}
	if z.now >= p.healAt {
		p.healAt = z.now + rules.HealthRegenPeriod
		if p.hp < float64(p.maxHP()) {
			p.hp = math.Min(float64(p.maxHP()), p.hp+float64(rules.HealthRegen(p.Char.Level, z.now < p.combatUntil)))
		}
	}
	// A character whose connection closed in a fight lingers, then goes.
	if p.graceEnd > 0 && z.now >= p.graceEnd && z.now >= p.combatUntil {
		z.W.savePlayer(p)
		p.closed.Store(true)
	}
	if z.now >= p.saveAt {
		p.saveAt = z.now + 30
		z.W.savePlayer(p)
	}
	// An exit takes you through only when you walk into it: someone who
	// arrives standing in one (coming through its twin) must step out first.
	e := z.exitAt(p)
	if e == nil {
		p.exitLatched = false
	}
	if z.now >= p.noExitUntil {
		if e != nil && !p.exitLatched {
			z.W.travel(z, p, e)
		} else if e == nil {
			z.borderCross(p)
		}
	}
}

// snapshots sends each player the world near them.
func (z *Zone) snapshots() {
	if len(z.players) == 0 {
		return
	}
	for _, p := range z.players {
		if p.closed.Load() {
			continue
		}
		s := protocol.Snap{T: protocol.TSnap, Tick: z.tick, Ack: p.stepsApplied, You: z.selfOf(p)}
		seen := map[int64]int{}
		for _, o := range z.players {
			if o == p || o.closed.Load() {
				continue
			}
			if math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) <= InterestRadius {
				e := z.playerEntity(o)
				// Like an NPC's, a player's gear goes out as they come into view,
				// and again whenever it changes (it may change to nothing).
				seen[o.ID] = o.gearRev + 1
				if p.gearSeen[o.ID] != o.gearRev+1 {
					e.Equip, e.GearSet = o.gear, true
				}
				s.Ents = append(s.Ents, e)
			}
		}
		p.gearSeen = seen
		inView := map[int64]bool{}
		for _, n := range z.npcs {
			if n.gone {
				continue
			}
			if math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y) <= InterestRadius {
				e := z.npcEntity(n, p)
				// Gear does not change in play, so it goes out once as the NPC comes into view.
				if len(n.Gear) > 0 {
					inView[n.ID] = true
					if !p.gearSent[n.ID] {
						e.Equip = n.Gear
					}
				}
				s.Ents = append(s.Ents, e)
			}
		}
		p.gearSent = inView
		for _, g := range z.graves {
			if math.Hypot(g.x-p.Body.X, g.y-p.Body.Y) <= InterestRadius {
				s.Ents = append(s.Ents, graveEntity(g, p))
			}
		}
		for _, e := range z.ev {
			if math.Hypot(e.x-p.Body.X, e.y-p.Body.Y) <= InterestRadius {
				s.Ev = append(s.Ev, e.e)
			}
		}
		p.correct = false
		p.send(s)
	}
}

func (z *Zone) selfOf(p *Player) protocol.Self {
	c := p.Char
	next, cur := rules.XPToReach(c.Level+1), rules.XPToReach(c.Level)
	s := protocol.Self{
		X: p.Body.X, Y: p.Body.Y, Z: p.Body.Z, Yaw: p.Body.Yaw,
		HP: int(math.Ceil(p.hp)), MaxHP: p.maxHP(), End: int(p.end), MaxEnd: rules.MaxEndurance,
		Level: c.Level, XP: c.XP - cur, XPNext: next - cur, Money: c.Money,
		Target: p.target, Attacking: p.attacking, Dead: p.dead, Correct: p.correct,
		Back: p.motion.Back, Side: p.motion.Side, Speed: p.motion.Speed,
		Fighting: z.now < p.combatUntil,
	}
	if c.Level >= MaxLevel {
		s.XP, s.XPNext = 1, 1
	}
	for i, def := range p.bar {
		if ok, _ := z.styleReady(p, &def); ok {
			s.Avail = append(s.Avail, i)
		}
	}
	return s
}

func (z *Zone) playerEntity(o *Player) protocol.Entity {
	st := ""
	if o.dead {
		st = "dead"
	} else if z.now < o.combatUntil {
		st = "fight"
	}
	return protocol.Entity{
		ID: o.ID, K: "p", Name: o.Char.Name, Char: o.Arch.Key, Race: o.Char.Race, Female: o.Char.Gender == 1, Level: o.Char.Level,
		X: o.Body.X, Y: o.Body.Y, Z: o.Body.Z, Yaw: o.Body.Yaw, Speed: o.motion.Speed,
		HP: math.Max(0, o.hp/float64(o.maxHP())), State: st, Back: o.motion.Back, Side: o.motion.Side,
	}
}

func (z *Zone) npcEntity(n *NPC, to *Player) protocol.Entity {
	st := ""
	switch {
	case n.dead():
		st = "dead"
	case n.target != nil:
		st = "fight"
	}
	e := protocol.Entity{
		ID: n.ID, K: "n", Name: n.title(), Guild: strings.TrimSpace(n.Spawn.Guild), Model: n.Model, Level: n.Level,
		X: n.Body.X, Y: n.Body.Y, Z: n.Body.Z, Yaw: n.Body.Yaw, Speed: n.speedNow,
		HP: math.Max(0, float64(n.HP)/float64(n.MaxHP)), State: st, Aggro: n.Aggro > 0, Friendly: n.Friendly, Kind: n.Kind,
	}
	if n.dead() && (len(n.loot) > 0 || n.money > 0) && n.lootOwner == to.ID {
		e.Loot = true
	}
	if !n.Friendly {
		con := rules.Con(to.Char.Level, n.Level)
		e.Con = &con
	}
	return e
}

// exitAt finds the exit a player stands in, if any.
func (z *Zone) exitAt(p *Player) *Exit {
	for i := range z.exits {
		e := &z.exits[i]
		if math.Hypot(e.X-p.Body.X, e.Y-p.Body.Y) < 400 && math.Abs(e.Z-p.Body.Z) < 800 {
			return e
		}
	}
	return nil
}

// sortedIDs is for deterministic iteration in tests and tools.
func sortedIDs[V any](m map[int64]V) []int64 {
	ids := make([]int64, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// borderCross carries a player who walks off an outdoor zone's edge into the
// neighbouring zone of the same region, if there is one and it is converted.
// The zones are one seamless land in the game; here each is its own place,
// so crossing loads the next and sets the player down just inside it.
func (z *Zone) borderCross(p *Player) {
	if _, outdoor := z.G.(*Terrain); !outdoor || p.motion.Speed == 0 || z.W.Cat == nil {
		return
	}
	info := z.W.Cat.Zones[z.ID]
	if info == nil || info.W == 0 {
		return
	}
	ext := z.G.Extent()
	const edge, push = 40.0, 600.0
	var dx, dy float64
	switch {
	case p.Body.X <= edge:
		dx = -1
	case p.Body.X >= ext-edge:
		dx = 1
	case p.Body.Y <= edge:
		dy = -1
	case p.Body.Y >= ext-edge:
		dy = 1
	default:
		return
	}
	rx := float64(info.OffX*8192) + p.Body.X + dx*push
	ry := float64(info.OffY*8192) + p.Body.Y + dy*push
	for _, o := range z.W.Cat.Zones {
		if o.ID == z.ID || o.Region != info.Region || !o.Converted || o.W == 0 {
			continue
		}
		lx, ly := rx-float64(o.OffX*8192), ry-float64(o.OffY*8192)
		if lx >= 0 && ly >= 0 && lx < float64(o.W*8192) && ly < float64(o.H*8192) && lx < 65280 && ly < 65280 {
			z.W.moveToZone(z, p, o.ID, lx, ly, 0)
			return
		}
	}
	p.noExitUntil = z.now + 5 // a world edge: no point asking again at once
}
