package world

import (
	"encoding/json"
	"math"
	"sync/atomic"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// PlayerIDBase keeps player entity ids apart from NPC ids, which are spawn ids.
const PlayerIDBase = 1_000_000

// Archetype is what a player character is drawn as and starts with.
type Archetype struct {
	Key   string // the web/data/char directory
	Label string
	Race  int
	Class int
	Realm int
}

// Archetypes are the characters a player may create. Both are Midgard
// warriors: the class whose styles the data and the combat rules cover.
var Archetypes = []Archetype{
	{Key: "troll-warrior", Label: "Troll Warrior", Race: 6, Class: 22, Realm: 2},
	{Key: "norseman", Label: "Norse Warrior", Race: 5, Class: 22, Realm: 2},
}

// ArchetypeFor finds the archetype of a saved character.
func ArchetypeFor(race, class int) Archetype {
	for _, a := range Archetypes {
		if a.Race == race && a.Class == class {
			return a
		}
	}
	return Archetypes[0]
}

type stepRun struct {
	n int
	k Keys
}

type defence struct {
	kind string // "parry", "block", "evade"
	at   float64
}

// invItem is one stack a player carries.
type invItem struct {
	DBID  int64
	ID    string
	Slot  int
	Count int
}

// Player is a connected character in a zone. All fields belong to the zone's
// goroutine; connections reach them through Zone.Do.
type Player struct {
	ID       int64
	gearSent map[int64]bool // NPCs whose equipment this player has been sent while they stayed in view
	gear     [][4]int       // what this player wears, as [slot, model, colour, extension]; slot is the item table's, so it reads like an NPC's
	gearRev  int            // changes whenever gear does, so other players are sent it again
	gearSeen map[int64]int  // the gearRev of each player this one has been sent, while they stayed in view
	Acct     *store.Account
	Char     *store.Character
	Arch     Archetype
	Zone     *Zone
	Body     Body
	Admin    bool

	out          chan []byte
	closed       atomic.Bool
	disconnected atomic.Bool
	zoning       atomic.Bool // closed to load another zone, not logging out
	// graceEnd is when a disconnected character in a fight leaves the world.
	graceEnd float64

	// Movement: inputs queued by the connection and consumed against a
	// budget of real time, so a client cannot move faster than the clock.
	queue        []stepRun
	queuedSteps  int
	budget       float64
	stepsApplied int64
	motion       Motion
	correct      bool
	lastPX       float64
	lastPY       float64
	havePredict  bool

	// Vitals and fight.
	hp, end     float64
	dead        bool
	deadAt      float64
	target      int64
	attacking   bool
	swingAt     float64 // when the next swing is ready
	combatUntil float64
	pending     int // style bar slot queued for the next swing, -1 for none
	lastDef     defence
	lastStyle   int
	lastStyleAt float64
	lastHit     bool
	lastHitAt   float64
	regenAt     float64
	healAt      float64

	inv         map[int]*invItem
	shop        int64 // the merchant whose shop is open, 0 for none
	weapon      *Item
	shield      *Item
	noExitUntil float64
	exitLatched bool
	bar         []StyleDef

	// The rules: the sheet derived from class, level, specs and gear, the last
	// one sent, and the defence bonus of the player's last styled hit, which
	// holds until their next swing.
	sh           sheet
	statsSent    protocol.Stats
	styleDefence int
	welcomed     bool      // the welcome has gone out, so the sheet may follow
	notice       string    // a system line to send once welcomed, such as a moved start
	effects      []*effect // lasting spells on them; not saved, so a logout ends them

	// Death and prayer: where they fell, what the death cost (the release costs
	// as much again), and a prayer in progress at their gravestone.
	deathAt       [3]float64
	lastLoss      int64
	creatureDeath bool
	prayUntil     float64
	prayFrom      [2]float64
	prayCombat    float64
	saveAt        float64
	invSent       bool
	chatTokens    float64
	chatAt        float64
}

func (p *Player) maxHP() int { return max(1, p.sh.maxHP) }

// send queues a message for the client; a client too slow to keep up loses
// frames rather than stalling the zone.
func (p *Player) send(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	p.sendRaw(b)
}

func (p *Player) sendRaw(b []byte) {
	if p.closed.Load() {
		return
	}
	select {
	case p.out <- b:
	default:
	}
}

// Out is the channel the connection drains to the client.
func (p *Player) Out() <-chan []byte { return p.out }

func (p *Player) errMsg(code, text string) {
	p.send(protocol.Err{T: protocol.TErr, Code: code, Text: text})
}

// angleTo is the heading from (x,y) to (tx,ty) in the viewer's convention.
func angleTo(x, y, tx, ty float64) float64 { return math.Atan2(tx-x, -(ty - y)) }

// angleOff is how far apart two headings are, 0..pi.
func angleOff(a, b float64) float64 {
	d := math.Mod(a-b, 2*math.Pi)
	if d > math.Pi {
		d -= 2 * math.Pi
	}
	if d < -math.Pi {
		d += 2 * math.Pi
	}
	return math.Abs(d)
}

// QueueInput adds a batch of movement steps. At most a second and a half of
// backlog is kept so a stalled connection does not turn into a burst, and
// each tick consumes only what real time allows.
func (p *Player) QueueInput(in *protocol.Input) {
	if p.dead || p.closed.Load() {
		return
	}
	for _, r := range in.Steps {
		n, k := KeysFromRun(r)
		if n == 0 {
			continue
		}
		p.queue = append(p.queue, stepRun{n, k})
		p.queuedSteps += n
	}
	for p.queuedSteps > protocol.StepsPerSecond/2*3 && len(p.queue) > 0 {
		// Drop the oldest: the newest keys are the current wish.
		p.queuedSteps -= p.queue[0].n
		p.queue = p.queue[1:]
	}
	p.lastPX, p.lastPY, p.havePredict = in.PX, in.PY, true
}

// applyInput consumes this tick's share of queued steps.
func (p *Player) applyInput(z *Zone) {
	p.budget += float64(protocol.StepsPerSecond) * TickDT
	if p.budget > 20 {
		p.budget = 20
	}
	moved := false
	var last Motion
	p.motion = Motion{}
	for p.budget >= 1 && len(p.queue) > 0 {
		r := &p.queue[0]
		k := r.k
		k.Slow = 1 - speedFactor(p)
		m := Step(z.G, z.Mv, &p.Body, k)
		if m.Speed > 0 {
			moved = true
			last = m
		}
		p.motion = m
		r.n--
		p.queuedSteps--
		p.budget--
		p.stepsApplied++
		if r.n == 0 {
			p.queue = p.queue[1:]
		}
	}
	// A tick that moved at all is a moving one, even if its last step ran
	// into a wall or the edge of the world.
	if moved {
		p.motion = last
	} else {
		p.motion = Motion{}
	}
	// The client says where its own simulation ended up after the steps it
	// has sent. Once we have caught up with them, a position far from ours is
	// a prediction gone wrong -- or a lie -- and the client is told to snap.
	if p.havePredict && len(p.queue) == 0 {
		if math.Hypot(p.lastPX-p.Body.X, p.lastPY-p.Body.Y) > 60 {
			p.correct = true
		}
		p.havePredict = false
	}
}

// newPlayer builds the in-world form of a saved character.
func newPlayer(acct *store.Account, ch *store.Character) *Player {
	return &Player{
		ID: PlayerIDBase + ch.ID, Acct: acct, Char: ch, Admin: acct.Admin,
		Arch: ArchetypeFor(ch.Race, ch.Class), out: make(chan []byte, 256), pending: -1, chatTokens: 4,
	}
}
