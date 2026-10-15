package world

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"daocweb/internal/server/protocol"
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

func testCatalog() *Catalog {
	c := &Catalog{
		Items: map[string]*Item{
			"hammer":  {ID: "hammer", Name: "bronze hammer", Level: 2, DPS: 1.8, Speed: 3.1, ObjectType: ObjHammer, ItemType: 10, Realm: 2, Droppable: true, Price: 10},
			"shield":  {ID: "shield", Name: "round shield", Level: 1, ObjectType: ObjShield, ItemType: 11, Realm: 2, Droppable: true, Price: 10},
			"pelt":    {ID: "pelt", Name: "wolf pelt", Level: 1, ObjectType: 0, ItemType: 40, Droppable: true, Price: 2},
			"helm":    {ID: "helm", Name: "studded helm", Level: 3, AF: 6, ObjectType: 34, ItemType: 21, Realm: 2, Droppable: true, Price: 5},
			"gsword":  {ID: "gsword", Name: "great sword", Level: 1, DPS: 3, Speed: 4, ObjectType: 11, ItemType: 12, Hand: 2, Realm: 2},
			"axe":     {ID: "axe", Name: "bronze axe", Level: 1, DPS: 2, Speed: 3, ObjectType: 13, ItemType: 10, Realm: 2, Droppable: true, Price: 5},
			"hiblade": {ID: "hiblade", Name: "hibernian blade", Level: 1, DPS: 2, Speed: 3, ObjectType: 11, ItemType: 10, Realm: 3},
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
		Templates: map[int]*Template{},
		Zones:     map[int]*ZoneInfo{100: {ID: 100, Name: "Test", Converted: true, Realm: 2}},
		Loot:      map[string][]LootEntry{"wolf": {{Item: "pelt", Chance: 100, Count: 1}}},
		MobLoot:   map[string][]string{"wolf": {"wolf"}},
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
	ch := &store.Character{AccountID: a.ID, Name: name, Race: 6, Class: 22, Level: level, XP: xpToReach(level), Zone: 100, X: x, Y: y, HP: playerMaxHP(level)}
	if err := r.st.CreateCharacter(ch); err != nil {
		r.t.Fatal(err)
	}
	if equip {
		r.st.Tx(func(tx *store.Tx) error {
			tx.AddItem(ch.ID, "hammer", SlotWeapon, 1)
			tx.AddItem(ch.ID, "shield", SlotShield, 1)
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
		t.Errorf("endurance %v -> %v; a 15-cost style should have been paid", end0, p.end)
	}
}

func TestLevelUpUnlocksStyles(t *testing.T) {
	r := newRig(t)
	p := r.player("Climber", 1, 30000, 30000, true)
	before := len(p.bar)
	r.z.award(p, xpToReach(6)-p.Char.XP)
	if p.Char.Level != 6 {
		t.Fatalf("level = %d", p.Char.Level)
	}
	if len(p.bar) <= before {
		t.Errorf("bar went from %d to %d styles at level 6", before, len(p.bar))
	}
	if p.hp != float64(p.maxHP()) {
		t.Error("a level up should heal")
	}
	// Saved.
	got, _ := r.st.Character(p.Acct.ID, p.Char.ID)
	if got.Level != 6 {
		t.Errorf("saved level %d", got.Level)
	}
}

func TestExperienceCurve(t *testing.T) {
	if killXP(5, 5) <= killXP(5, 10) {
		t.Error("a grey kill should pay less")
	}
	if killXP(1, 30) != 0 {
		t.Error("a very grey kill should pay nothing")
	}
	if killXP(8, 5) <= killXP(5, 5) {
		t.Error("a red kill should pay more")
	}
	// About eight even-level kills to the next level.
	perLevel := float64(xpToReach(11)-xpToReach(10)) / float64(killXP(10, 10))
	if perLevel < 5 || perLevel > 12 {
		t.Errorf("%.1f kills a level at 10", perLevel)
	}
	if npcMaxHP(10) <= npcMaxHP(5) {
		t.Error("tougher mobs should have more hit points")
	}
}

func TestEquipRules(t *testing.T) {
	r := newRig(t)
	p := r.player("Dresser", 1, 30000, 30000, false)
	r.st.Tx(func(tx *store.Tx) error {
		for i, id := range []string{"hammer", "gsword", "hiblade", "helm", "pelt"} {
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
	try(1, SlotWeapon) // a two-hander: a warrior here refuses
	try(2, SlotWeapon) // Hibernian: no
	if p.weapon.ID != "hammer" {
		t.Errorf("weapon is %s after illegal swaps", p.weapon.ID)
	}
	try(3, SlotHead) // a level-3 helm at level 1: allowed within two levels
	if p.armour != 6 {
		t.Errorf("armour = %d", p.armour)
	}
	try(4, SlotHead) // a pelt on the head: no, and it must not displace the helm
	if p.inv[SlotHead] == nil || p.inv[SlotHead].ID != "helm" {
		t.Error("the helm was displaced by a pelt")
	}
	// Persisted.
	r2 := newRigOn(t, r)
	items, _ := r2.st.Items(p.Char.ID)
	slots := map[int]string{}
	for _, it := range items {
		slots[it.Slot] = it.ItemID
	}
	if slots[SlotWeapon] != "hammer" || slots[SlotHead] != "helm" {
		t.Errorf("saved slots: %v", slots)
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
	// Defence rates over many blows land near the configured chances.
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
	block := float64(counts["block"]) / 4000
	if want := playerBlock(20, true); math.Abs(block-want*npcHitChance(5, 20)) > 0.05 {
		t.Errorf("block rate %.3f, want about %.3f", block, want*npcHitChance(5, 20))
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
