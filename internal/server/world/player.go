package world

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
	"sync/atomic"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// PlayerIDBase keeps player entity ids apart from NPC ids, which are spawn ids.
const PlayerIDBase = 1_000_000

// Archetype is what a player character is drawn as and starts with.
type Archetype struct {
	ID    string // what the creation page sends: "troll-warrior", "dwarf-healer"
	Key   string // the web/data/char directory drawn when no race base model fits
	Label string
	Race  int
	Class int
	Realm int
}

// Archetypes are the characters a player may create: every race a class the
// rules cover may be, in Midgard. The first two keep the ids they always had.
var Archetypes = buildArchetypes()

func buildArchetypes() []Archetype {
	var out []Archetype
	ids := make([]int, 0, len(rules.Classes))
	for id := range rules.Classes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return rules.Classes[ids[i]].Name < rules.Classes[ids[j]].Name })
	for _, id := range ids {
		c := rules.Classes[id]
		races := append([]int(nil), c.Races...)
		sort.Ints(races)
		for _, r := range races {
			race, ok := rules.Races[r]
			if !ok {
				continue
			}
			a := Archetype{ID: strings.ToLower(race.Name) + "-" + strings.ToLower(c.Name),
				Key: "norseman", Label: race.Name + " " + c.Name, Race: r, Class: id, Realm: 2}
			switch a.ID {
			case "troll-warrior":
				a.Key = "troll-warrior"
			case "norseman-warrior":
				a.ID = "norseman"
			}
			if r == 6 {
				a.Key = "troll-warrior"
			}
			out = append(out, a)
		}
	}
	return out
}

// ArchetypeFor finds the archetype of a saved character.
func ArchetypeFor(race, class int) Archetype {
	for _, a := range Archetypes {
		if a.Race == race && a.Class == class {
			return a
		}
	}
	for _, a := range Archetypes {
		if a.ID == "troll-warrior" {
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
	Cond  int // condition in hundredths of a percent, store.FullCondition whole
}

// Player is a connected character in a zone. All fields belong to the zone's
// goroutine; connections reach them through Zone.Do.
type Player struct {
	ID         int64
	power      float64         // spell power, up to sh.maxPower
	powerSet   bool            // the pool has been filled once
	powerAt    float64         // when the next power tick is due
	book       []bookSpell     // the spells they know
	bookSent   protocol.Book   // what the client was last told
	buffed     bool            // a spell buff was on at the last sheet
	casting    *playerCast     // the spell being cast, or nil
	spellReady map[int]float64 // spell id -> when its recast has passed
	ride       *ride           // the horse route they are on, nil when walking
	ab         abilityState    // Guard, Protect, Intercept and Engage in force
	gearSent   map[int64]bool  // NPCs whose equipment this player has been sent while they stayed in view
	gear       [][4]int        // what this player wears, as [slot, model, colour, extension]; slot is the item table's, so it reads like an NPC's
	gearRev    int             // changes whenever gear does, so other players are sent it again
	gearSeen   map[int64]int   // the gearRev of each player this one has been sent, while they stayed in view
	Acct       *store.Account
	Char       *store.Character
	Arch       Archetype
	Zone       *Zone
	Body       Body
	Admin      bool

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
	shop        int64            // the merchant whose shop is open, 0 for none
	vault       map[int]*invItem // their vault, by slot (VaultBase and up)
	vaultNPC    int64            // the vault keeper whose vault is open, 0 for none
	smithNPC    int64            // the blacksmith whose repair window is open, 0 for none
	condDirty   bool             // some item's condition changed since the last save
	weapon      *Item
	shield      *Item
	noExitUntil float64
	exitLatched bool
	bar         []StyleDef
	barAlso     []StyleDef // styles they have learned but cannot use with what they hold

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
		Arch: ArchetypeFor(ch.Race, ch.Class), out: make(chan []byte, 256), pending: -1, chatTokens: 4, spellReady: map[int]float64{},
	}
}
