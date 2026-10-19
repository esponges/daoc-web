package world

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"time"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// flat is level ground with a lake in one corner.
type flat struct{}

func (flat) GroundAt(x, y, _ float64) float64 { return 5000 }
func (flat) IsWet(x, y float64) bool          { return x < 2000 && y < 2000 }
func (flat) Clamp(x, y float64) (float64, float64) {
	return math.Max(0, math.Min(65280, x)), math.Max(0, math.Min(65280, y))
}
func (flat) Extent() float64 { return 65280 }

// grant is one row of the ability tables.
func grant(source string, level int, ability string, abilityLevel int) rules.Grant {
	return rules.Grant{Source: source, Level: level, Ability: ability, AbilityLevel: abilityLevel}
}

func testCatalog() *Catalog {
	c := &Catalog{
		Items: map[string]*Item{
			"hammer":   {ID: "hammer", Name: "bronze hammer", Model: 311, Level: 2, DPS: 1.8, Speed: 3.1, ObjectType: ObjHammer, ItemType: 10, Realm: 2, Droppable: true, Price: 10},
			"shield":   {ID: "shield", Name: "round shield", Model: 473, Level: 1, ObjectType: ObjShield, ItemType: 11, Realm: 2, Droppable: true, Price: 10},
			"pelt":     {ID: "pelt", Name: "wolf pelt", Level: 1, ObjectType: 0, ItemType: 40, Droppable: true, Price: 2},
			"helm":     {ID: "helm", Name: "studded helm", Model: 93, Level: 3, AF: 6, ObjectType: 34, ItemType: 21, Realm: 2, Droppable: true, Price: 5},
			"gsword":   {ID: "gsword", Name: "great sword", Level: 1, DPS: 3, Speed: 4, ObjectType: 11, ItemType: 12, Hand: 1, Realm: 2},
			"hauberk":  {ID: "hauberk", Name: "chain hauberk", Level: 8, AF: 16, Absorb: 27, ObjectType: 35, ItemType: 25, Realm: 2},
			"bighelm":  {ID: "bighelm", Name: "great helm", Level: 20, AF: 30, ObjectType: 34, ItemType: 21, Realm: 2, LevelReq: 20},
			"hsword":   {ID: "hsword", Name: "heavy sword", Level: 20, DPS: 7, Speed: 3, ObjectType: 11, ItemType: 10, Realm: 2},
			"strhelm":  {ID: "strhelm", Name: "helm of might", Level: 3, AF: 6, ObjectType: 34, ItemType: 21, Realm: 2, Bonuses: []ItemBonus{{1, 4}}},
			"axe":      {ID: "axe", Name: "bronze axe", Level: 1, DPS: 2, Speed: 3, ObjectType: 13, ItemType: 10, Realm: 2, Droppable: true, Price: 5},
			"vest":     {ID: "vest", Name: "studded vest", Level: 1, AF: 6, Absorb: 19, ObjectType: 34, ItemType: 25, Realm: 2},
			"leggings": {ID: "leggings", Name: "studded leggings", Level: 1, AF: 6, Absorb: 19, ObjectType: 34, ItemType: 27, Realm: 2},
			"sleeves":  {ID: "sleeves", Name: "studded sleeves", Level: 1, AF: 6, Absorb: 19, ObjectType: 34, ItemType: 28, Realm: 2},
			"boots":    {ID: "boots", Name: "studded boots", Level: 1, AF: 6, Absorb: 19, ObjectType: 34, ItemType: 23, Realm: 2},
			"gloves":   {ID: "gloves", Name: "studded gloves", Level: 1, AF: 6, Absorb: 19, ObjectType: 34, ItemType: 22, Realm: 2},
			"cap":      {ID: "cap", Name: "studded cap", Level: 1, AF: 6, Absorb: 19, ObjectType: 34, ItemType: 21, Realm: 2},
			"hiblade":  {ID: "hiblade", Name: "hibernian blade", Level: 1, DPS: 2, Speed: 3, ObjectType: 11, ItemType: 10, Realm: 3},
		},
		Styles: []StyleDef{
			{ID: 161, Class: 22, Name: "Thor's Anvil", Spec: "Hammer", SpecLevel: 2, Endurance: 15, WeaponReq: ObjHammer, GrowthRate: 0.037},
			{ID: 162, Class: 22, Name: "Crushing Blow", Spec: "Hammer", SpecLevel: 4, Endurance: 5, OpenType: openDef, ResultReq: resParry, WeaponReq: ObjHammer, GrowthRate: 0.857},
			{ID: 164, Class: 22, Name: "Slam", Spec: "Hammer", SpecLevel: 4, Endurance: 10, OpenType: 0, OpenValue: 162, ResultReq: resStyle, WeaponReq: ObjHammer, GrowthRate: 0.857},
			{ID: 169, Class: 22, Name: "Revenge", Spec: "Hammer", SpecLevel: 4, Endurance: 5, OpenType: openDef, ResultReq: resBlock, WeaponReq: ObjHammer, GrowthRate: 0.857},
			{ID: 166, Class: 22, Name: "Frost Hammer", Spec: "Hammer", SpecLevel: 4, Endurance: 10, OpenType: openPos, OpenValue: posSide, WeaponReq: ObjHammer, GrowthRate: 0.671},
			{ID: 171, Class: 22, Name: "Conquer", Spec: "Hammer", SpecLevel: 4, Endurance: 5, OpenType: openPos, OpenValue: posBack, WeaponReq: ObjHammer, GrowthRate: 0.719},
			{ID: 221, Class: 22, Name: "Numb", Spec: "Shields", SpecLevel: 3, Endurance: 15, WeaponReq: ObjShield},
			{ID: 900, Class: 22, Name: "High Level", Spec: "Hammer", SpecLevel: 40, Endurance: 5, WeaponReq: ObjHammer},
		},
		Templates:  map[int]*Template{},
		ClassSpecs: map[int][]string{22: {"VikingCareer", "WarriorCareer", "Axe", "Hammer", "Sword", "Parry", "Shields"}},
		Grants: map[int][]rules.Grant{22: {
			grant("VikingCareer", 1, rules.AbMidArmor, 3), grant("VikingCareer", 1, rules.AbShield, 1),
			grant("VikingCareer", 1, "Weaponry: Swords", 0), grant("VikingCareer", 1, "Weaponry: Hammers", 0), grant("VikingCareer", 1, "Weaponry: Axes", 0),
			grant("WarriorCareer", 5, rules.AbShield, 2), grant("WarriorCareer", 10, rules.AbMidArmor, 4), grant("WarriorCareer", 10, rules.AbShield, 3),
			grant("WarriorCareer", 15, rules.AbEvade, 1), grant("WarriorCareer", 7, "Protect", 1),
		}},
		Resists: map[int]map[int]float64{},
		Zones:   map[int]*ZoneInfo{100: {ID: 100, Name: "Test", Converted: true, Realm: 2}},
		Loot:    map[string][]LootEntry{"wolf": {{Item: "pelt", Chance: 100, Count: 1}}},
		MobLoot: map[string][]string{"wolf": {"wolf"}},
	}
	return c
}

type rig struct {
	t  *testing.T
	w  *World
	z  *Zone
	st *store.Store
}

func newRig(t *testing.T, spawns ...Spawn) *rig {
	t.Helper()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	w := New(Config{Cat: testCatalog(), Store: st, Log: t.Logf})
	z := NewZone(w, 100, "Test", flat{}, nil, spawns)
	w.zones[100] = z
	return &rig{t: t, w: w, z: z, st: st}
}

func (r *rig) player(name string, level int, x, y float64, equip bool) *Player {
	r.t.Helper()
	a, err := r.st.AccountFor("d:"+name, name, false)
	if err != nil {
		r.t.Fatal(err)
	}
	ch := &store.Character{AccountID: a.ID, Name: name, Race: 6, Class: 22, Level: level, XP: rules.XPToReach(level), Zone: 100, X: x, Y: y, XPScheme: store.XPSchemeRules}
	if err := r.st.CreateCharacter(ch); err != nil {
		r.t.Fatal(err)
	}
	if equip {
		r.st.Tx(func(tx *store.Tx) error {
			tx.AddItem(ch.ID, "hammer", SlotWeapon, 1)
			tx.AddItem(ch.ID, "shield", SlotShield, 1)
			for slot, id := range map[int]string{SlotTorso: "vest", SlotLegs: "leggings", SlotArms: "sleeves", SlotFeet: "boots", SlotHands: "gloves", SlotHead: "cap"} {
				tx.AddItem(ch.ID, id, slot, 1)
			}
			return nil
		})
	}
	p := newPlayer(a, ch)
	r.z.join(p)
	return p
}

// ticks runs n ticks.
func (r *rig) ticks(n int) {
	for i := 0; i < n; i++ {
		r.z.Tick()
	}
}

// snaps drains and decodes everything sent to a player.
func snaps(p *Player) (out []protocol.Snap) {
	for {
		select {
		case b := <-p.out:
			var env protocol.Envelope
			json.Unmarshal(b, &env)
			if env.T == protocol.TSnap {
				var s protocol.Snap
				json.Unmarshal(b, &s)
				out = append(out, s)
			}
		default:
			return
		}
	}
}

func wolf(id int64, x, y float64, level int) Spawn {
	return Spawn{ID: id, Name: "wolf", Model: 56, X: x, Y: y, Z: 5000, Level: level, Size: 50, Speed: 191, Roam: 400,
		RespawnS: 30, AggroLevel: 1, AggroRange: 400}
}

func TestPlayersSeeEachOther(t *testing.T) {
	r := newRig(t)
	a := r.player("Alice", 5, 30000, 30000, true)
	b := r.player("Bobby", 5, 30500, 30000, true)
	far := r.player("Faraway", 5, 60000, 60000, true)
	r.ticks(2)
	sa := snaps(a)
	last := sa[len(sa)-1]
	if len(last.Ents) != 1 || last.Ents[0].ID != b.ID || last.Ents[0].Name != "Bobby" || last.Ents[0].Char != "troll-warrior" {
		t.Errorf("Alice sees %+v", last.Ents)
	}
	if sf := snaps(far); len(sf[len(sf)-1].Ents) != 0 {
		t.Errorf("a far player sees %+v", sf[len(sf)-1].Ents)
	}
}

func TestMovementFollowsInputAndClock(t *testing.T) {
	r := newRig(t)
	p := r.player("Runner", 5, 30000, 30000, true)
	in := &protocol.Input{T: "input", Seq: 60, Steps: [][5]int{{60, 1, 0, 0, protocol.FlagRun}}}
	p.QueueInput(in)
	r.ticks(22) // 60 steps at 3 a tick
	if p.stepsApplied != 60 {
		t.Fatalf("applied %d steps, want 60", p.stepsApplied)
	}
	if d := math.Hypot(p.Body.X-30000, p.Body.Y-30000); math.Abs(d-Run) > 1 {
		t.Errorf("a second of running moved %.1f, want %v", d, Run)
	}
	if s := snaps(p); s[len(s)-1].Ack != 60 {
		t.Errorf("ack = %d", s[len(s)-1].Ack)
	}
}

func TestSpeedHackIsLimitedByTheClock(t *testing.T) {
	r := newRig(t)
	p := r.player("Cheat", 5, 30000, 30000, true)
	// Ten seconds of running, sent all at once, then one real second passes.
	for i := 0; i < 5; i++ {
		p.QueueInput(&protocol.Input{T: "input", Steps: [][5]int{{120, 1, 0, 0, protocol.FlagRun}}, PX: 30000, PY: 20000})
	}
	r.ticks(TickRate)
	moved := math.Hypot(p.Body.X-30000, p.Body.Y-30000)
	// A second of clock buys a second of steps (plus a little burst), never ten.
	if moved > Run*1.5 {
		t.Errorf("moved %.0f units in one second; the limit is about %v", moved, Run)
	}
	// And the position the client claimed is not believed.
	r.ticks(TickRate * 4)
	s := snaps(p)
	if got := s[len(s)-1].You; math.Hypot(got.X-30000, got.Y-20000) < 5000 {
		t.Errorf("server accepted the claimed position: %+v", got)
	}
}

func TestBelowGroundIsPutOnGround(t *testing.T) {
	r := newRig(t)
	p := r.player("Low", 5, 30000, 30000, true)
	p.Body.Z = -9999
	r.ticks(1)
	p.QueueInput(&protocol.Input{T: "input", Steps: [][5]int{{3, 1, 0, 0, 1}}})
	r.ticks(1)
	if p.Body.Z != 5000 {
		t.Errorf("z = %v, want the ground", p.Body.Z)
	}
}

func TestFightKillsWolfAndPaysOut(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 3))
	p := r.player("Fighter", 3, 30000, 30000, true)
	// Face the wolf (east).
	p.Body.Yaw = math.Pi / 2
	n := r.z.byID[1]
	n.Roam = 0
	r.z.SetTarget(p, 1)
	r.z.SetAttack(p, true)
	xp0 := p.Char.XP
	for i := 0; i < 20*120 && !n.dead(); i++ {
		r.z.Tick()
		if p.dead {
			t.Fatalf("a level 3 warrior lost to a level 3 wolf at tick %d", i)
		}
	}
	if !n.dead() {
		t.Fatal("the wolf did not die in two minutes")
	}
	if p.Char.XP <= xp0 {
		t.Error("no experience for the kill")
	}
	if n.lootOwner != p.ID || len(n.loot) == 0 || n.money == 0 {
		t.Errorf("loot: owner %d items %v money %d", n.lootOwner, n.loot, n.money)
	}
	// Take everything.
	r.z.takeLoot(p, 1, 0)
	r.z.takeLoot(p, 1, -1)
	if p.inv[0] == nil || p.inv[0].ID != "pelt" || p.Char.Money == 0 {
		t.Errorf("inventory %+v money %d", p.inv[0], p.Char.Money)
	}
	// It is in the database, not just in memory.
	items, _ := r.st.Items(p.Char.ID)
	found := false
	for _, it := range items {
		found = found || it.ItemID == "pelt"
	}
	if !found {
		t.Error("the pelt was not saved")
	}
	// The wolf comes back.
	r.ticks(TickRate * 130)
	if n.dead() || n.HP != n.MaxHP {
		t.Errorf("wolf did not respawn whole: dead=%v hp=%d/%d", n.dead(), n.HP, n.MaxHP)
	}
}

func TestOtherPlayersCannotLootExclusively(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 1))
	a := r.player("Killer", 3, 30000, 30000, true)
	b := r.player("Thiefy", 3, 30050, 30000, true)
	n := r.z.byID[1]
	n.Roam = 0
	a.Body.Yaw = math.Pi / 2
	r.z.SetTarget(a, 1)
	r.z.SetAttack(a, true)
	for i := 0; i < 20*60 && !n.dead(); i++ {
		r.z.Tick()
	}
	if !n.dead() {
		t.Fatal("no kill")
	}
	r.z.takeLoot(b, 1, 0)
	if b.inv[0] != nil {
		t.Error("another player looted a fresh kill")
	}
	r.ticks(int(LootExclusive*TickRate) + 20)
	if n.gone {
		return // the corpse is gone, which is also fine
	}
	r.z.takeLoot(b, 1, 0)
}

func TestAggressiveWolfAttacksAndLeashes(t *testing.T) {
	r := newRig(t, wolf(1, 30300, 30000, 5))
	p := r.player("Prey", 5, 30000, 30000, true)
	n := r.z.byID[1]
	r.ticks(TickRate * 3)
	if n.target != p {
		t.Fatal("an aggressive wolf 300 units away did not attack")
	}
	// Run far away: it gives up and goes home.
	p.Body.X, p.Body.Y = 30000+9000, 30000
	r.ticks(TickRate * 30)
	if n.target != nil {
		t.Error("the wolf never gave up")
	}
	r.ticks(TickRate * 40)
	if math.Hypot(n.Body.X-n.HX, n.Body.Y-n.HY) > n.Roam+50 && n.state != npcIdle && n.state != npcWalk {
		t.Errorf("wolf is not home: state %v at %.0f,%.0f", n.state, n.Body.X, n.Body.Y)
	}
}

func TestPlayerDiesAndComesBack(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 25), wolf(2, 30150, 30060, 25))
	p := r.player("Doomed", 2, 30000, 30000, false)
	died := false
	for i := 0; i < TickRate*60 && !died; i++ {
		r.z.Tick()
		died = p.dead
	}
	if !died {
		t.Fatal("two level 25 wolves did not kill a level 2 fighter")
	}
	for _, n := range r.z.npcs {
		if n.target == p {
			t.Error("a wolf still has the dead player as its target")
		}
	}
	r.ticks(int(PlayerRespawnS*TickRate) + 5)
	if p.dead || p.hp <= 0 {
		t.Errorf("player did not come back: dead=%v hp=%v", p.dead, p.hp)
	}
}

func TestPlayerCanReleaseEarly(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 25), wolf(2, 30150, 30060, 25))
	p := r.player("Doomed", 2, 30000, 30000, false)
	died := false
	for i := 0; i < TickRate*60 && !died; i++ {
		r.z.Tick()
		died = p.dead
	}
	if !died {
		t.Fatal("two level 25 wolves did not kill a level 2 fighter")
	}
	r.z.ReleasePlayer(p)
	if p.dead || p.hp <= 0 {
		t.Errorf("release did not stand the player up: dead=%v hp=%v", p.dead, p.hp)
	}
	hp := p.hp
	r.z.ReleasePlayer(p) // a living player is left alone
	if p.hp != hp || p.dead {
		t.Errorf("release changed a living player: hp %v -> %v", hp, p.hp)
	}
}

func TestWanderingStaysDryAndLocal(t *testing.T) {
	// A wolf at the edge of the lake, roaming over it.
	sp := wolf(1, 2600, 2600, 3)
	sp.AggroRange = 0
	sp.Roam = 1200
	r := newRig(t, sp)
	r.player("Watcher", 5, 30000, 30000, true) // keeps the zone busy
	n := r.z.byID[1]
	moved := 0.0
	px, py := n.Body.X, n.Body.Y
	for i := 0; i < TickRate*600; i++ {
		r.z.Tick()
		moved += math.Hypot(n.Body.X-px, n.Body.Y-py)
		if d := math.Hypot(n.Body.X-px, n.Body.Y-py); d > n.Walk*TickDT*1.05 {
			t.Fatalf("tick %d: moved %.2f in a tick, walk allows %.2f", i, d, n.Walk*TickDT)
		}
		px, py = n.Body.X, n.Body.Y
		if (flat{}).IsWet(n.Body.X, n.Body.Y) {
			t.Fatalf("tick %d: wolf stands in the lake at %.0f,%.0f", i, n.Body.X, n.Body.Y)
		}
		if math.Hypot(n.Body.X-n.HX, n.Body.Y-n.HY) > n.Roam+100 {
			t.Fatalf("tick %d: wolf strayed to %.0f,%.0f", i, n.Body.X, n.Body.Y)
		}
	}
	if moved < 500 {
		t.Errorf("the wolf barely moved in ten minutes (%.0f units)", moved)
	}
}

func TestStyleOpenings(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 3))
	p := r.player("Stylist", 5, 30000, 30000, true)
	r.train(p, map[string]int{"Hammer": 5, "Shields": 3})
	n := r.z.byID[1]
	n.Roam, n.Aggro = 0, 0
	p.Body.Yaw = math.Pi / 2
	r.z.SetTarget(p, 1)
	byName := func(s string) int {
		for i, d := range p.bar {
			if d.Name == s {
				return i
			}
		}
		return -1
	}
	if byName("High Level") >= 0 {
		t.Error("a level 40 style is on a level 5 bar")
	}
	if byName("Thor's Anvil") < 0 || byName("Numb") < 0 {
		t.Fatalf("bar = %+v", p.bar)
	}
	crush, slam, anvil := byName("Crushing Blow"), byName("Slam"), byName("Thor's Anvil")
	ready := func(i int) bool { ok, _ := r.z.styleReady(p, &p.bar[i]); return ok }
	if !ready(anvil) || ready(crush) || ready(slam) {
		t.Errorf("fresh: anvil=%v crush=%v slam=%v", ready(anvil), ready(crush), ready(slam))
	}
	// A parry opens Crushing Blow for a few seconds, then it closes.
	p.lastDef = defence{kind: "parry", at: r.z.now}
	if !ready(crush) || ready(byName("Revenge")) {
		t.Error("a parry should open Crushing Blow and not Revenge")
	}
	r.ticks(TickRate * 7)
	if ready(crush) {
		t.Error("the opening should have closed")
	}
	// A block opens Revenge.
	p.lastDef = defence{kind: "block", at: r.z.now}
	if !ready(byName("Revenge")) {
		t.Error("a block should open Revenge")
	}
	// Chain: Slam follows Crushing Blow.
	p.lastStyle, p.lastStyleAt = 162, r.z.now
	if !ready(slam) {
		t.Error("Slam should follow Crushing Blow")
	}
	p.lastStyle = 161
	if ready(slam) {
		t.Error("Slam should not follow Thor's Anvil")
	}
	// Positional: the wolf is east of us. Facing west it faces us (front);
	// facing east it has its back to us; facing north we are on its flank.
	side, back := byName("Frost Hammer"), byName("Conquer")
	n.Body.Yaw = -math.Pi / 2
	if ready(side) || ready(back) {
		t.Error("positional styles should not be ready from the front")
	}
	n.Body.Yaw = math.Pi / 2
	if !ready(back) || ready(side) {
		t.Error("Conquer needs the back")
	}
	n.Body.Yaw = 0
	if !ready(side) || ready(back) {
		t.Error("Frost Hammer needs the side")
	}
	// Not enough endurance.
	p.end = 1
	if ready(anvil) {
		t.Error("a style was ready with 1 endurance")
	}
}

func TestStyleIsSpentOnSwing(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 1))
	p := r.player("Swinger", 5, 30000, 30000, true)
	r.train(p, map[string]int{"Hammer": 5})
	n := r.z.byID[1]
	n.Roam, n.Aggro, n.MaxHP, n.HP = 0, 0, 100000, 100000
	p.Body.Yaw = math.Pi / 2
	r.z.SetTarget(p, 1)
	r.z.SetAttack(p, true)
	anvil := -1
	for i, d := range p.bar {
		if d.Name == "Thor's Anvil" {
			anvil = i
		}
	}
	r.z.requestStyle(p, anvil)
	if p.pending != anvil {
		t.Fatal("style not queued")
	}
	p.end = 60
	end0 := p.end
	for i := 0; i < TickRate*8 && p.pending >= 0; i++ {
		r.z.Tick()
	}
	if p.pending >= 0 {
		t.Fatal("style never used")
	}
	if p.end > end0-5 {
		t.Errorf("endurance %v -> %v; a 15-cost style (12 with a 3.1 s hammer) should have been paid", end0, p.end)
	}
}

func TestLevelUpGrantsStatsAndPoints(t *testing.T) {
	r := newRig(t)
	p := r.player("Climber", 5, 30000, 30000, true)
	str, pts := p.sh.stats[rules.STR], p.sh.unspent
	r.z.award(p, rules.XPToReach(6)-p.Char.XP)
	if p.Char.Level != 6 {
		t.Fatalf("level = %d", p.Char.Level)
	}
	if p.sh.stats[rules.STR] != str+1 || p.sh.unspent != pts+12 {
		t.Errorf("level 6: strength %d -> %d, points %d -> %d; want +1 and +12", str, p.sh.stats[rules.STR], pts, p.sh.unspent)
	}
	if p.hp != float64(p.maxHP()) {
		t.Error("a level up should heal")
	}
	got, _ := r.st.Character(p.Acct.ID, p.Char.ID)
	if got.Level != 6 {
		t.Errorf("saved level %d", got.Level)
	}
}

// train sets a player's specs as if trained, and rebuilds what depends on them.
func (r *rig) train(p *Player, specs map[string]int) {
	for k, v := range specs {
		p.Char.Specs[k] = v
	}
	r.w.recalc(p)
	r.z.buildBar(p)
}

func TestEquipRules(t *testing.T) {
	r := newRig(t)
	p := r.player("Dresser", 1, 30000, 30000, false)
	r.st.Tx(func(tx *store.Tx) error {
		for i, id := range []string{"hammer", "gsword", "hiblade", "helm", "pelt", "shield", "hauberk", "bighelm", "hsword"} {
			tx.AddItem(p.Char.ID, id, i, 1)
		}
		return nil
	})
	r.w.loadInventory(p)
	try := func(from, to int) { r.z.MoveItem(p, from, to) }
	try(0, SlotWeapon) // a bronze hammer: fine
	if p.weapon == nil || p.weapon.ID != "hammer" {
		t.Error("the hammer did not equip")
	}
	try(1, SlotShield) // a sword into the shield slot: no
	try(2, SlotWeapon) // Hibernian: no
	if p.weapon.ID != "hammer" {
		t.Errorf("weapon is %s after illegal swaps", p.weapon.ID)
	}
	try(3, SlotHead) // a level-3 helm with no level requirement: allowed
	try(4, SlotHead) // a pelt on the head: no, and it must not displace the helm
	if p.inv[SlotHead] == nil || p.inv[SlotHead].ID != "helm" {
		t.Error("the helm was displaced by a pelt")
	}
	try(7, SlotHead) // level requirement 20: no
	if p.inv[SlotHead].ID != "helm" {
		t.Error("a level 20 helm went on a level 1 head")
	}
	try(6, SlotTorso) // chain at level 1: no
	if p.inv[SlotTorso] != nil {
		t.Error("chain went on a level 1 Warrior")
	}
	// A high-level weapon with no requirement goes on, but hits only as hard as the level allows.
	try(8, SlotWeapon)
	if p.weapon == nil || p.weapon.ID != "hsword" {
		t.Fatal("the heavy sword did not equip")
	}
	if base, _ := rules.BaseDamage(p.combatant(r.w.Cat)); base > rules.DPSCap(1)*3*1.04 {
		t.Errorf("base damage %.2f is over the level 1 DPS cap", base)
	}
	// A great sword with a shield worn: the shield goes to the pack.
	try(5, SlotShield)
	if p.shield == nil {
		t.Fatal("the shield did not equip")
	}
	try(1, SlotWeapon)
	if p.weapon == nil || p.weapon.ID != "gsword" || p.shield != nil || p.inv[SlotShield] != nil {
		t.Errorf("two-hander: weapon %v, shield %v", p.weapon, p.shield)
	}
	inPack := false
	for s := 0; s < PackSize; s++ {
		inPack = inPack || p.inv[s] != nil && p.inv[s].ID == "shield"
	}
	if !inPack {
		t.Error("the shield is not in the pack")
	}
	// And no shield beside it.
	for s := 0; s < PackSize; s++ {
		if p.inv[s] != nil && p.inv[s].ID == "shield" {
			try(s, SlotShield)
		}
	}
	if p.shield != nil {
		t.Error("a shield went on beside a two-handed sword")
	}
	// Persisted.
	items, _ := r.st.Items(p.Char.ID)
	slots := map[int]string{}
	for _, it := range items {
		slots[it.Slot] = it.ItemID
	}
	if slots[SlotWeapon] != "gsword" || slots[SlotHead] != "helm" || slots[SlotShield] != "" {
		t.Errorf("saved slots: %v", slots)
	}
}

func TestChainAtTen(t *testing.T) {
	r := newRig(t)
	p := r.player("Hauberk", 10, 30000, 30000, false)
	r.st.Tx(func(tx *store.Tx) error { return tx.AddItem(p.Char.ID, "hauberk", 0, 1) })
	r.w.loadInventory(p)
	r.z.MoveItem(p, 0, SlotTorso)
	if p.inv[SlotTorso] == nil {
		t.Error("a level 10 Warrior could not wear chain")
	}
}

func TestStrengthItemRaisesTheSheet(t *testing.T) {
	r := newRig(t)
	p := r.player("Mighty", 5, 30000, 30000, true)
	r.st.Tx(func(tx *store.Tx) error { return tx.AddItem(p.Char.ID, "strhelm", 0, 1) })
	r.w.loadInventory(p)
	str, ws := p.sh.stats[rules.STR], rules.WeaponSkill(p.combatant(r.w.Cat))
	snaps(p)
	r.z.MoveItem(p, 0, SlotHead)
	if p.sh.stats[rules.STR] != str+4 || rules.WeaponSkill(p.combatant(r.w.Cat)) <= ws {
		t.Errorf("strength %d -> %d", str, p.sh.stats[rules.STR])
	}
	if p.statsSent.Stats[rules.STR] != str+4 {
		t.Error("the sheet sent does not show the new strength")
	}
}

func newRigOn(t *testing.T, r *rig) *rig { return r }

func TestChatRangeAndRate(t *testing.T) {
	r := newRig(t)
	a := r.player("Talker", 5, 30000, 30000, true)
	b := r.player("Nearby", 5, 31000, 30000, true)
	c := r.player("Distant", 5, 50000, 30000, true)
	drain := func(p *Player) (n int) {
		for {
			select {
			case m := <-p.out:
				var e protocol.Envelope
				json.Unmarshal(m, &e)
				if e.T == protocol.TChatMsg {
					n++
				}
			default:
				return
			}
		}
	}
	drain(a)
	drain(b)
	drain(c)
	r.w.Chat(a, "say", "hello")
	if drain(a) != 1 || drain(b) != 1 || drain(c) != 0 {
		t.Error("say should reach those within range only")
	}
	// A flood: only the burst gets through.
	delivered := 0
	for i := 0; i < 20; i++ {
		r.w.Chat(a, "say", "spam")
		delivered += drain(b)
	}
	if delivered > 5 {
		t.Errorf("%d of 20 flooded lines delivered", delivered)
	}
}

func TestAdminCommandsNeedAdmin(t *testing.T) {
	r := newRig(t)
	p := r.player("Ordinary", 5, 30000, 30000, true)
	r.w.GM(p, "level 50")
	if p.Char.Level != 5 {
		t.Error("an ordinary player used a GM command")
	}
	p.Admin = true
	r.w.GM(p, "level 20")
	if p.Char.Level != 20 {
		t.Errorf("level = %d", p.Char.Level)
	}
	r.w.GM(p, "tp 40000 41000")
	if p.Body.X != 40000 || p.Body.Y != 41000 {
		t.Errorf("tp: %v,%v", p.Body.X, p.Body.Y)
	}
}

func TestPackmatesJoin(t *testing.T) {
	r := newRig(t, wolf(1, 30300, 30000, 3), wolf(2, 30400, 30100, 3), wolf(3, 30500, 30000, 3))
	for _, n := range r.z.npcs {
		n.Aggro = 0 // only fights back
	}
	p := r.player("Puller", 5, 30000, 30000, true)
	p.Body.Yaw = math.Pi / 2
	r.z.npcs[0].engage(p, r.z)
	if r.z.npcs[1].target != p || r.z.npcs[2].target != p {
		t.Error("packmates within 900 units should join the fight")
	}
}

func TestExitsTravelAndLatch(t *testing.T) {
	r := newRig(t)
	r.w.Cat.Zones[101] = &ZoneInfo{ID: 101, Name: "Other", Converted: true}
	r.w.Cat.Zones[102] = &ZoneInfo{ID: 102, Name: "Unconverted"}
	r.z.exits = []Exit{
		{SrcZone: 100, X: 31000, Y: 30000, Z: 5000, DstZone: 101, TX: 111, TY: 222, TZ: 333},
		{SrcZone: 100, X: 35000, Y: 30000, Z: 5000, DstZone: 102, TX: 1, TY: 2},
	}
	// Someone who arrives standing in an exit is not sent straight back.
	arrived := r.player("Arrival", 5, 31010, 30000, true)
	r.ticks(10)
	if arrived.closed.Load() {
		t.Error("a player arriving inside an exit was bounced through it")
	}
	// Stepping out and back in does take them through.
	arrived.Body.X = 30000
	r.ticks(2)
	arrived.Body.X = 31005
	r.ticks(2)
	if !arrived.closed.Load() || arrived.Char.Zone != 101 || arrived.Char.X != 111 || arrived.Char.Z != 333 {
		t.Errorf("exit: closed=%v char=%+v", arrived.closed.Load(), arrived.Char)
	}
	saved, _ := r.st.Character(arrived.Acct.ID, arrived.Char.ID)
	if saved.Zone != 101 || saved.Y != 222 {
		t.Errorf("saved %+v", saved)
	}
	// An exit to a place that is not converted refuses, and says so.
	p := r.player("Blocked", 5, 30000, 30000, true)
	r.ticks(2)
	p.Body.X, p.Body.Y = 35010, 30000
	r.ticks(2)
	if p.closed.Load() || p.Char.Zone != 100 {
		t.Error("walked through an exit to nowhere")
	}
}

// playerIn joins a player into a given zone of the rig.
func (r *rig) playerIn(z *Zone, name string, x, y float64) *Player {
	r.t.Helper()
	a, _ := r.st.AccountFor("d:"+name, name, false)
	ch := &store.Character{AccountID: a.ID, Name: name, Race: 6, Class: 22, Level: 5, Zone: z.ID, X: x, Y: y, HP: 500}
	if err := r.st.CreateCharacter(ch); err != nil {
		r.t.Fatal(err)
	}
	p := newPlayer(a, ch)
	z.join(p)
	return p
}

func TestEmptyZoneUnloadsAndResets(t *testing.T) {
	st, _ := store.Open(":memory:")
	t.Cleanup(func() { st.Close() })
	w := New(Config{Cat: testCatalog(), Store: st, Log: t.Logf, IdleUnload: 0.4})
	z := NewZone(w, 100, "Test", flat{}, nil, []Spawn{wolf(1, 30100, 30000, 3)})
	w.zones[100] = z
	z.Start()
	a, _ := st.AccountFor("d:u", "u", false)
	ch := &store.Character{AccountID: a.ID, Name: "Visitor", Race: 6, Class: 22, Level: 5, Zone: 100, X: 30000, Y: 30000, HP: 500}
	st.CreateCharacter(ch)
	p := newPlayer(a, ch)
	ready := make(chan struct{})
	z.Do(func() { z.join(p); close(ready) })
	<-ready
	// Hurt the wolf, then leave: the zone should save the player and stop,
	// and a fresh zone should have the wolf whole again.
	z.Do(func() { z.byID[1].HP = 1 })
	w.Leave(p)
	deadline := time.Now().Add(5 * time.Second)
	for len(w.ZoneRunning()) > 0 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if len(w.ZoneRunning()) != 0 {
		t.Fatal("the empty zone never unloaded")
	}
	saved, _ := st.Character(a.ID, ch.ID)
	if saved.Zone != 100 || saved.X != 30000 {
		t.Errorf("player not saved: %+v", saved)
	}
	z2 := NewZone(w, 100, "Test", flat{}, nil, []Spawn{wolf(1, 30100, 30000, 3)})
	if n := z2.byID[1]; n.HP != n.MaxHP {
		t.Errorf("a reloaded zone's wolf has %d/%d", n.HP, n.MaxHP)
	}
}

func TestReachAndDefences(t *testing.T) {
	// A blow out of reach does nothing: the client only asks, the server decides.
	r := newRig(t, wolf(1, 31000, 30000, 1))
	p := r.player("Distant", 5, 30000, 30000, true)
	n := r.z.byID[1]
	n.Roam, n.Aggro = 0, 0
	p.Body.Yaw = math.Pi / 2
	r.z.SetTarget(p, 1)
	r.z.SetAttack(p, true)
	r.ticks(TickRate * 10)
	if n.HP != n.MaxHP {
		t.Errorf("a wolf 1000 units away took damage (%d/%d)", n.HP, n.MaxHP)
	}
	// Defence rates over many blows land near the rules' chances.
	r2 := newRig(t, wolf(2, 30100, 30000, 5))
	p2 := r2.player("Shielded", 20, 30000, 30000, true)
	w2 := r2.z.byID[2]
	w2.Aggro = 0
	p2.Body.Yaw = math.Pi / 2
	w2.Body.Yaw = -math.Pi / 2 // facing the player
	counts := map[string]int{}
	for i := 0; i < 4000; i++ {
		r2.z.ev = r2.z.ev[:0]
		p2.hp = float64(p2.maxHP())
		w2.swingAt = 0
		r2.z.npcSwing(w2, p2)
		for _, e := range r2.z.ev {
			counts[e.e.K]++
		}
	}
	cb := p2.combatant(r2.w.Cat)
	e, pa, b := rules.EvadeChance(cb, true, 1), rules.ParryChance(cb, true, 1, false), rules.BlockChance(cb, true)
	want := (1 - e) * (1 - pa) * b
	if block := float64(counts["block"]) / 4000; math.Abs(block-want) > 0.02 {
		t.Errorf("block rate %.3f, want about %.3f", block, want)
	}
	if counts["parry"] == 0 || counts["evade"] == 0 || counts["hit"] == 0 || counts["miss"] == 0 {
		t.Errorf("outcomes: %v", counts)
	}
	// From behind a shield cannot block.
	counts = map[string]int{}
	p2.Body.Yaw = -math.Pi / 2 // now facing away from the wolf
	for i := 0; i < 2000; i++ {
		r2.z.ev = r2.z.ev[:0]
		p2.hp = float64(p2.maxHP())
		r2.z.npcSwing(w2, p2)
		for _, e := range r2.z.ev {
			counts[e.e.K]++
		}
	}
	if counts["block"] != 0 || counts["parry"] != 0 {
		t.Errorf("blocked or parried a blow from behind: %v", counts)
	}
}

func TestNPCGuildIsSentWithTheSnapshot(t *testing.T) {
	sp := wolf(1, 30100, 30000, 3)
	sp.Guild = " Smith "
	r := newRig(t, sp)
	a := r.player("Alice", 5, 30000, 30000, true)
	r.ticks(2)
	sa := snaps(a)
	last := sa[len(sa)-1]
	if len(last.Ents) != 1 || last.Ents[0].Guild != "Smith" {
		t.Errorf("Alice sees %+v", last.Ents)
	}
}

func TestNPCGearIsSentOnceWhileInView(t *testing.T) {
	geared := wolf(1, 30100, 30000, 3)
	geared.Equip = "set"
	unknown := wolf(2, 30200, 30000, 3)
	unknown.Equip = "nope" // no rows in the catalog
	bare := wolf(3, 30300, 30000, 3)
	r := newRig(t, geared, unknown, bare)
	r.w.Cat.Gear = map[string][][4]int{"set": {{25, 662, 17, 0}, {10, 311, 0, 0}}}
	// Respawn the zone's NPCs from the new catalog data.
	r.z = NewZone(r.w, 100, "Test", flat{}, nil, []Spawn{geared, unknown, bare})
	r.w.zones[100] = r.z
	a := r.player("Alice", 5, 30000, 30000, true)
	eqOf := func(ss []protocol.Snap) (first, later int, _ int) {
		for i, s := range ss {
			for _, e := range s.Ents {
				switch {
				case e.Name == "wolf" && len(e.Equip) > 0 && i == 0:
					first++
				case e.Name == "wolf" && len(e.Equip) > 0:
					later++
				}
			}
		}
		return
	}
	r.ticks(4)
	first, later, _ := eqOf(snaps(a))
	if first != 1 || later != 0 {
		t.Errorf("gear sent %d times in the first snapshot and %d after; want 1 and 0", first, later)
	}
	// Out of view and back: it is sent again.
	a.Body.X, a.Body.Y = 90000, 90000
	r.ticks(2)
	snaps(a)
	a.Body.X, a.Body.Y = 30000, 30000
	r.ticks(2)
	first, later, _ = eqOf(snaps(a))
	if first+later != 1 {
		t.Errorf("gear sent %d times after coming back into view, want 1", first+later)
	}
}

func TestPlayerGearIsSentWhenItChanges(t *testing.T) {
	r := newRig(t)
	a := r.player("Alice", 5, 30000, 30000, true) // hammer and shield
	b := r.player("Bobby", 5, 30500, 30000, false)
	r.ticks(2)
	want := [][4]int{{10, 311, 0, 0}, {11, 473, 0, 0}}
	if !reflect.DeepEqual(a.gear, want) {
		t.Fatalf("Alice wears %v, want %v", a.gear, want)
	}
	count := func(p *Player, id int64) (sent int, last [][4]int) {
		for _, s := range snaps(p) {
			for _, e := range s.Ents {
				if e.ID == id && e.GearSet {
					sent++
					last = e.Equip
				}
			}
		}
		return
	}
	// Bobby is sent Alice's gear once as she comes into view, with her race.
	r.ticks(1)
	n, got := count(b, a.ID)
	if n != 1 || !reflect.DeepEqual(got, want) {
		t.Errorf("Bobby was sent Alice's gear %d times: %v", n, got)
	}
	if n, _ := count(b, a.ID); n != 0 {
		t.Errorf("gear was repeated while nothing changed: %d", n)
	}
	// She takes the shield off; Bobby is sent the new set, and Alice's own inventory carries it too.
	r.z.MoveItem(a, SlotShield, 0)
	r.ticks(2)
	n, got = count(b, a.ID)
	if n != 1 || !reflect.DeepEqual(got, [][4]int{{10, 311, 0, 0}}) {
		t.Errorf("after unequipping, Bobby was sent %d times: %v", n, got)
	}
	// She strips everything: an empty set is still sent, so Bobby does not keep the old look.
	r.z.MoveItem(a, SlotWeapon, 1)
	r.ticks(2)
	n, got = count(b, a.ID)
	if n != 1 || len(got) != 0 {
		t.Errorf("after stripping, Bobby was sent %d times: %v", n, got)
	}
}

func TestPlayerRaceAndGenderAreSent(t *testing.T) {
	r := newRig(t)
	r.player("Alice", 5, 30000, 30000, false)
	b := r.player("Bobby", 5, 30500, 30000, false)
	r.player("Cathy", 5, 30700, 30000, false).Char.Gender = 1
	r.ticks(2)
	last := snaps(b)
	for _, e := range last[len(last)-1].Ents {
		if e.Race != 6 || e.Female != (e.Name == "Cathy") {
			t.Errorf("%s: race %d female %v", e.Name, e.Race, e.Female)
		}
	}
}
