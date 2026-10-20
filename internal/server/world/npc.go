package world

import (
	"math"
	"math/rand"
	"strings"

	"daocweb/internal/server/rules"
)

type npcState int

const (
	npcIdle npcState = iota
	npcWalk
	npcChase
	npcReturn
	npcDead
)

type lootItem struct {
	ID    string
	Count int
}

// NPC is a creature in a zone, run by the server.
type NPC struct {
	ID         int64
	attackedAt float64 // when an enemy last attacked it, for Engage
	Spawn      Spawn
	Name       string
	Model      int
	Gear       [][4]int // worn pieces, shared with every NPC of the same equipment template
	Level      int
	Size       int
	Scale      float64
	Body       Body
	HX, HY     float64 // home
	Roam       float64
	Walk       float64 // walking pace, units a second
	RunSpd     float64
	HP         int
	MaxHP      int
	Swing      float64
	cb         rules.Combatant // its side of a swing, under the creature rules
	Reach      float64
	Parry      float64
	Evade      float64
	Block      float64
	Aggro      float64 // 0 means it only fights back
	Friendly   bool    // guards, villagers, merchants: players cannot attack them, nor they players
	Guard      bool    // a town guard: fights the creatures that threaten its realm
	Kind       string  // what interacting with it does (KindMerchant), "" for nothing
	Teaches    string  // the class a trainer teaches
	Tether     float64

	twoHanded  bool
	damageType int
	resists    map[int]float64

	state      npcState
	timer      float64
	tx, ty     float64
	speedNow   float64
	target     combatant            // who it is fighting: the one it hates most, nil when calm
	hate       map[int64]*hateEntry // entity id -> how much it hates them
	hateOrder  int
	effects    []*effect       // lasting spells on it
	spells     []*SpellDef     // what it casts
	procs      []*SpellDef     // offensive procs it carries from spawn
	spellReady map[int]float64 // spell id -> when it may be cast again
	casting    *castState      // the spell it is casting, nil when none
	swingAt    float64
	returning  bool
	hurtBy     map[int64]int // damage dealt by each player, for experience and loot
	tagger     int64
	guardRange float64 // how far a guard watches
	watchAt    float64 // when a calm guard next looks around
	diedAt     float64
	respawnAt  float64
	loot       []lootItem
	money      int64
	lootOwner  int64
	gone       bool
	rng        *rand.Rand
}

// newNPC builds a creature from its spawn row.
func newNPC(sp Spawn, cat *Catalog, rng *rand.Rand) *NPC {
	level := spawnLevel(sp, rng)
	size := sp.Size
	if size <= 0 {
		size = 50
	}
	speed := float64(sp.Speed)
	if speed < 40 || speed > 400 {
		speed = Run
	}
	n := &NPC{
		ID: sp.ID, Spawn: sp, Name: sp.Name, Model: sp.Model, Level: level, Size: size,
		HX: sp.X, HY: sp.Y, Roam: float64(sp.Roam),
		RunSpd: speed, Walk: math.Max(40, speed*0.45),
		Reach:    90 + float64(size)*0.9,
		Friendly: sp.Realm != 0,
		hurtBy:   map[int64]int{},
		rng:      rng,
	}
	n.Scale = 1
	n.Kind = npcKind(sp.ClassType)
	if n.Guard = isGuard(sp.ClassType); n.Guard {
		n.guardRange = float64(sp.AggroRange)
		if n.guardRange <= 0 {
			n.guardRange = GuardRange
		}
	}
	n.Teaches = teaches(sp.ClassType)
	if sp.Equip != "" {
		n.Gear = cat.Gear[sp.Equip]
	}
	damageType := rules.DamageCrush
	if t := cat.Templates[sp.Template]; t != nil {
		n.Parry, n.Evade, n.Block = float64(t.Parry)/100, float64(t.Evade)/100, float64(t.Block)/100
		n.Tether = float64(t.Tether)
		if t.MeleeType >= rules.DamageCrush && t.MeleeType <= rules.DamageThrus {
			damageType = t.MeleeType
		}
	}
	for _, g := range n.Gear {
		if g[0] == 12 { // the two-handed slot
			n.twoHanded = true
		}
	}
	n.damageType, n.resists = damageType, cat.Resists[sp.Race]
	n.learnSpells(cat)
	n.setLevel(level)
	if n.Tether < 800 || n.Tether > 6000 {
		n.Tether = 2500
	}
	if !n.Friendly && sp.AggroLevel > 0 && sp.AggroRange > 0 {
		n.Aggro = math.Min(float64(sp.AggroRange), 700)
	}
	if n.Roam > 1500 {
		n.Roam = 1500
	}
	n.reset()
	return n
}

// setLevel gives the creature a level and the swing, defences and health
// that go with it.
func (n *NPC) setLevel(level int) {
	n.Level = level
	n.cb = rules.Creature(level, n.Evade*100, n.Parry*100, n.Block*100, n.twoHanded, n.damageType, n.resists)
	n.Swing = rules.CreatureSwing(&n.cb)
	n.MaxHP = rules.CreatureHP(level)
}

// spawnLevel is the level a creature comes back at: one of its template's
// levels, picked afresh each spawn, or its own row's level.
func spawnLevel(sp Spawn, rng *rand.Rand) int {
	level := sp.Level
	if len(sp.Levels) > 0 {
		level = sp.Levels[rng.Intn(len(sp.Levels))]
	}
	return max(level, 1)
}

// reset puts the creature at home, whole and calm.
func (n *NPC) reset() {
	n.Body = Body{X: n.HX, Y: n.HY, Yaw: float64(n.Spawn.Heading) / 4096 * 2 * math.Pi}
	n.Body.Z = n.Spawn.Z
	n.HP = n.MaxHP
	n.state = npcIdle
	n.timer = 1 + n.rng.Float64()*8
	n.target = nil
	n.hate = nil
	n.effects = nil
	n.casting = nil
	n.spellReady = map[int]float64{}
	n.returning = false
	n.hurtBy = map[int64]int{}
	n.tagger = 0
	n.loot, n.money, n.lootOwner = nil, 0, 0
	n.gone = false
	n.speedNow = 0
}

func (n *NPC) dead() bool { return n.state == npcDead }

// pickSpot is a dry spot near home to wander to, or home itself.
func (n *NPC) pickSpot(z *Zone) (float64, float64) {
	for i := 0; i < 16; i++ {
		a := n.rng.Float64() * 2 * math.Pi
		r := math.Sqrt(n.rng.Float64()) * n.Roam
		x, y := z.G.Clamp(n.HX+math.Cos(a)*r, n.HY+math.Sin(a)*r)
		if !z.G.IsWet(x, y) {
			return x, y
		}
	}
	return n.HX, n.HY
}

// stepToward moves toward (tx,ty) at a speed, turning first like the viewer's
// NPCs do. It returns the distance left, or -1 if water is in the way.
func (n *NPC) stepToward(z *Zone, tx, ty, speed float64) float64 {
	if len(n.effects) > 0 {
		speed *= speedFactor(n)
	}
	dx, dy := tx-n.Body.X, ty-n.Body.Y
	d := math.Hypot(dx, dy)
	n.speedNow = 0
	if d < 1e-3 {
		return 0
	}
	want := math.Atan2(dx, -dy)
	n.Body.Yaw = turnToward(n.Body.Yaw, want, 3.5*TickDT)
	if angleOff(want, n.Body.Yaw) >= 0.6 {
		return d
	}
	s := math.Min(d, speed*TickDT)
	nx, ny := n.Body.X+math.Sin(n.Body.Yaw)*s, n.Body.Y-math.Cos(n.Body.Yaw)*s
	nx, ny = z.G.Clamp(nx, ny)
	if z.G.IsWet(nx, ny) && !z.G.IsWet(n.Body.X, n.Body.Y) {
		return -1
	}
	if z.Mv != nil {
		nx, ny = z.Mv.Slide(n.Body.X, n.Body.Y, nx, ny, n.Body.Z)
	}
	moved := math.Hypot(nx-n.Body.X, ny-n.Body.Y)
	n.Body.X, n.Body.Y = nx, ny
	n.Body.Z = z.G.GroundAt(nx, ny, n.Body.Z)
	n.speedNow = moved / TickDT
	return d - s
}

func turnToward(cur, target, maxStep float64) float64 {
	d := math.Mod(target-cur, 2*math.Pi)
	if d > math.Pi {
		d -= 2 * math.Pi
	}
	if d < -math.Pi {
		d += 2 * math.Pi
	}
	return cur + math.Max(-maxStep, math.Min(maxStep, d))
}

// update runs one tick of a creature's life.
func (n *NPC) update(z *Zone) {
	switch n.state {
	case npcDead:
		n.speedNow = 0
		if !n.gone && z.now-n.diedAt > n.corpseTime() {
			n.gone = true
			n.loot, n.money = nil, 0
		}
		if z.now >= n.respawnAt {
			if len(n.Spawn.Levels) > 1 {
				n.setLevel(spawnLevel(n.Spawn, n.rng))
			}
			n.reset()
		}
		return
	}
	if held(n) {
		n.speedNow = 0
		return
	}
	if n.Friendly && !n.Guard {
		n.wander(z, false)
		return
	}
	// It fights whoever it hates most; with nobody left to fight, it goes home.
	if n.state == npcChase {
		if n.target = n.hated(z); n.target == nil {
			n.goHome()
		}
	}
	// A guard watches for creatures; an aggressive creature notices prey.
	if n.Guard && n.target == nil && n.state != npcReturn {
		n.watch(z)
	} else if n.target == nil && n.Aggro > 0 && n.state != npcReturn {
		if p := z.nearestPlayer(n.Body.X, n.Body.Y, n.Aggro, func(p *Player) bool {
			return !p.dead && p.ride == nil && p.Char.Level-n.Level < 8
		}); p != nil {
			n.engage(p, z)
		}
	}
	switch n.state {
	case npcChase:
		n.chase(z)
	case npcReturn:
		left := n.stepToward(z, n.HX, n.HY, n.RunSpd)
		n.HP = int(math.Min(float64(n.MaxHP), float64(n.HP)+math.Max(1, float64(n.MaxHP)*0.08*TickDT)))
		if left >= 0 && left < 20 || left < 0 {
			n.state, n.returning = npcIdle, false
			n.HP = n.MaxHP
			n.hurtBy = map[int64]int{}
			n.timer = 2
			n.speedNow = 0
		}
	default:
		n.wander(z, true)
	}
}

func (n *NPC) corpseTime() float64 {
	if len(n.loot) == 0 && n.money == 0 {
		return 12
	}
	return NPCCorpseSeconds
}

func (n *NPC) wander(z *Zone, regen bool) {
	if regen && n.HP < n.MaxHP {
		n.HP += int(math.Max(1, float64(n.MaxHP)*0.01))
		if n.HP > n.MaxHP {
			n.HP = n.MaxHP
		}
	}
	if n.Roam < 40 {
		n.speedNow = 0
		return
	}
	switch n.state {
	case npcIdle:
		n.speedNow = 0
		n.timer -= TickDT
		if n.timer <= 0 {
			n.tx, n.ty = n.pickSpot(z)
			n.state = npcWalk
			n.timer = 4*math.Hypot(n.tx-n.Body.X, n.ty-n.Body.Y)/n.Walk + 2
		}
	case npcWalk:
		n.timer -= TickDT
		left := math.Hypot(n.tx-n.Body.X, n.ty-n.Body.Y)
		if left >= math.Max(12, n.Walk*TickDT) && n.timer > 0 {
			if n.stepToward(z, n.tx, n.ty, n.Walk) < 0 {
				n.timer = 0
			}
		} else {
			n.state = npcIdle
			n.timer = 3 + n.rng.Float64()*9
			n.speedNow = 0
		}
	}
}

// goHome gives up the fight: it forgets everyone and walks back.
func (n *NPC) goHome() {
	n.state = npcReturn
	n.returning = true
	n.target = nil
	n.hate = nil
	n.casting = nil
}

// engage starts a fight with a combatant, or adds to one, and calls its
// packmates in: they take on the same anger.
func (n *NPC) engage(c combatant, z *Zone) {
	if n.dead() || n.Friendly && !n.Guard {
		return
	}
	first := n.target == nil
	if !n.hates(c) {
		n.addHate(c, 1)
	}
	n.target = n.hated(z)
	n.state = npcChase
	if first {
		n.swingAt = z.now + 0.6
		// Packmates: others of the same name close by join in.
		for _, o := range z.npcs {
			if o != n && !o.dead() && !o.Friendly && o.target == nil && o.state != npcReturn &&
				o.Name == n.Name && math.Hypot(o.Body.X-n.Body.X, o.Body.Y-n.Body.Y) < 900 {
				o.addHate(c, 1)
				o.target = c
				o.state = npcChase
				o.swingAt = z.now + 0.9
			}
		}
	}
}

// chase casts, closes on the target and strikes when it can.
func (n *NPC) chase(z *Zone) {
	tx, ty := n.target.position()
	if math.Hypot(n.Body.X-n.HX, n.Body.Y-n.HY) > n.Tether || math.Hypot(tx-n.HX, ty-n.HY) > n.Tether+1500 {
		n.goHome()
		return
	}
	if n.casting != nil {
		z.continueCasting(n)
		return
	}
	if z.startCasting(n) {
		return
	}
	dx, dy := tx-n.Body.X, ty-n.Body.Y
	d := math.Hypot(dx, dy)
	reach := n.Reach + 20
	n.Body.Yaw = turnToward(n.Body.Yaw, math.Atan2(dx, -dy), 6*TickDT)
	if d > reach {
		if n.stepToward(z, tx, ty, n.RunSpd) < 0 {
			n.goHome()
		}
		return
	}
	n.speedNow = 0
	if z.now >= n.swingAt {
		n.swingAt = z.now + n.Swing
		z.npcSwing(n, n.target)
	}
}

// nearestPlayer finds the closest player within range that passes ok.
func (z *Zone) nearestPlayer(x, y, rng float64, ok func(*Player) bool) *Player {
	var best *Player
	bd := rng
	for _, p := range z.players {
		if p.closed.Load() || (ok != nil && !ok(p)) {
			continue
		}
		if d := math.Hypot(p.Body.X-x, p.Body.Y-y); d <= bd {
			best, bd = p, d
		}
	}
	return best
}

func (n *NPC) title() string {
	if n.Spawn.Name == "" {
		return "creature"
	}
	return strings.TrimSpace(n.Spawn.Name)
}
