package world

import (
	"math"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// archerRig is an archer with a composite bow and a quiver of arrows, a wolf
// 1000 units ahead of them, and the archery abilities.
func archerRig(t *testing.T, arrows int) (*rig, *Player, *NPC) {
	r := newRig(t, wolf(1, 31000, 30000, 8))
	r.z.rng = rand.New(rand.NewSource(3))
	c := r.w.Cat
	c.Items["cbow"] = &Item{ID: "cbow", Name: "composite bow", Level: 5, DPS: 12, Speed: 4.5, ObjectType: rules.ObjCompBow, ItemType: 13, Realm: 2, Hand: 1}
	c.Items["arrows"] = &Item{ID: "arrows", Name: "standard arrows", ObjectType: rules.ObjArrow, ItemType: 40, Speed: 6.0, PackSize: 20, MaxCount: 200, Price: 480}
	w := r.z.npcs[0]
	w.Aggro, w.Roam = 0, 0
	w.setLevel(8)
	w.HP = w.MaxHP
	p := r.player("Archer", 20, 30000, 30000, true)
	p.Body.Yaw = angleTo(30000, 30000, 31000, 30000)
	r.give(p, SlotRanged, "cbow", 1)
	if arrows > 0 {
		r.give(p, SlotQuiver, "arrows", arrows)
	}
	p.sh.abilities[rules.AbCriticalShot], p.sh.abilities[rules.AbRapidFire], p.sh.abilities[rules.AbPenetrating] = 3, 2, 1
	p.sh.stats[rules.QUI] = 60
	p.end = 100
	p.target = w.ID
	return r, p, w
}

// give puts an item in a slot of a player's inventory and saves it.
func (r *rig) give(p *Player, slot int, id string, count int) {
	r.t.Helper()
	if err := r.st.Tx(func(tx *store.Tx) error { return tx.AddItem(p.Char.ID, id, slot, count) }); err != nil {
		r.t.Fatal(err)
	}
	p.inv[slot] = &invItem{ID: id, Slot: slot, Count: count}
	r.w.recalcGear(p)
}

func arrowsLeft(p *Player) int {
	n := 0
	for s := SlotQuiver; s < SlotQuiver+4; s++ {
		if it := p.inv[s]; it != nil {
			n += it.Count
		}
	}
	return n
}

func TestAShotNeedsABowAnArrowAndATargetInRange(t *testing.T) {
	r, p, w := archerRig(t, 0)
	r.z.Shoot(p, "")
	if p.draw != nil || !strings.Contains(lastErr(t, p), "no ammunition") {
		t.Errorf("an empty quiver: draw %v, error %q", p.draw != nil, lastErr(t, p))
	}
	r, p, w = archerRig(t, 5)
	w.Body.X = 33000 // beyond a composite bow's 1600 (the arrows' range class is 3, so 2000)
	r.z.Shoot(p, "")
	if p.draw != nil || !strings.Contains(lastErr(t, p), "too far away") {
		t.Errorf("out of range: draw %v, error %q", p.draw != nil, lastErr(t, p))
	}
	r, p, _ = archerRig(t, 5)
	r.st.Tx(func(tx *store.Tx) error { return tx.RemoveItem(p.Char.ID, SlotRanged) })
	delete(p.inv, SlotRanged)
	r.w.recalcGear(p)
	r.z.Shoot(p, "")
	if p.draw != nil || !strings.Contains(lastErr(t, p), "no bow") {
		t.Errorf("no bow: draw %v, error %q", p.draw != nil, lastErr(t, p))
	}
}

func TestAFullDrawReleasesOneArrow(t *testing.T) {
	r, p, w := archerRig(t, 20)
	r.z.Shoot(p, "")
	if p.draw == nil {
		t.Fatalf("no draw: %q", lastErr(t, p))
	}
	want := rules.DrawTime(4.5, 60)
	if math.Abs(p.draw.ends-r.z.now-want) > 1e-6 {
		t.Errorf("the draw takes %v, want %v", p.draw.ends-r.z.now, want)
	}
	r.ticks(int(want*TickRate) - 2)
	if arrowsLeft(p) != 20 || w.HP != w.MaxHP {
		t.Fatal("an arrow flew before the draw ended")
	}
	r.ticks(4)
	if arrowsLeft(p) != 19 {
		t.Errorf("%d arrows after one shot, want 19", arrowsLeft(p))
	}
	if p.end > 99 {
		t.Errorf("endurance %v: a shot costs 5", p.end)
	}
	if left := savedCount(t, r, p, SlotQuiver); left != 19 {
		t.Errorf("the saved stack is %d, want 19", left)
	}
	// It goes on shooting at the target until told to stop.
	r.ticks(int(want*TickRate) + 4)
	if arrowsLeft(p) != 18 {
		t.Errorf("%d arrows after two shots, want 18", arrowsLeft(p))
	}
	r.z.Shoot(p, "stop")
	if p.draw != nil || p.shooting {
		t.Error("stop left the bow drawn")
	}
}

func savedCount(t *testing.T, r *rig, p *Player, slot int) int {
	t.Helper()
	items, _ := r.st.Items(p.Char.ID)
	for _, it := range items {
		if it.Slot == slot {
			return it.Count
		}
	}
	return 0
}

func TestMovingInterruptsTheDrawAndUsesNoArrow(t *testing.T) {
	r, p, _ := archerRig(t, 20)
	r.z.Shoot(p, "")
	r.ticks(10)
	p.Body.X += 100
	r.ticks(2)
	if p.draw != nil || p.shooting {
		t.Error("the draw went on after moving")
	}
	if arrowsLeft(p) != 20 {
		t.Errorf("%d arrows, want 20", arrowsLeft(p))
	}
}

func TestBeingHitInterruptsTheDraw(t *testing.T) {
	r, p, _ := archerRig(t, 20)
	r.z.Shoot(p, "")
	r.ticks(5)
	r.z.damaged(p)
	if p.draw != nil {
		t.Error("a blow left the draw going")
	}
}

func TestCriticalAndRapidDraws(t *testing.T) {
	r, p, w := archerRig(t, 20)
	base := rules.DrawTime(4.5, 60)
	r.z.Shoot(p, "critical")
	if p.draw == nil || p.draw.kind != "critical" || math.Abs(p.draw.ends-r.z.now-rules.CriticalDraw(base, 3)) > 1e-6 {
		t.Fatalf("critical draw %+v", p.draw)
	}
	r.ticks(int(rules.CriticalDraw(base, 3)*TickRate) + 4)
	if arrowsLeft(p) != 19 || p.draw != nil || p.shooting {
		t.Errorf("a critical shot: %d arrows, draw %v, shooting %v (one shot, then stop)", arrowsLeft(p), p.draw != nil, p.shooting)
	}
	if w.hurtBy[p.ID] == 0 && w.HP == w.MaxHP {
		t.Log("the critical shot missed")
	}
	r.z.Shoot(p, "rapid")
	if !p.rapid {
		t.Fatal("rapid fire is not on")
	}
	r.z.Shoot(p, "")
	if p.draw == nil || p.draw.kind != "rapid" || math.Abs(p.draw.ends-r.z.now-rules.RapidDraw(base)) > 1e-6 {
		t.Fatalf("rapid draw %+v", p.draw)
	}
	r.z.Shoot(p, "stop")
	r.z.Shoot(p, "critical")
	if p.rapid {
		t.Error("Critical Shot left rapid fire on")
	}
}

func TestShotsAreNeverParried(t *testing.T) {
	r, p, w := archerRig(t, 200)
	w.cb.ParryPct, w.baseCB.ParryPct = 100, 100
	r.z.Shoot(p, "")
	r.ticks(int(rules.DrawTime(4.5, 60)*TickRate*12) + 20)
	r.z.Shoot(p, "stop")
	if w.hurtBy[p.ID] == 0 {
		t.Errorf("twelve shots at a creature that parries everything did no damage (%d arrows used)", 200-arrowsLeft(p))
	}
}

func TestAHunterKillsWithArrowsAndGetsTheExperience(t *testing.T) {
	r, p, w := archerRig(t, 200)
	w.setLevel(19)                                                                                         // a grey creature gives nothing
	r.z.addEffect(w, p, &SpellDef{ID: 9900, Name: "Stun", Type: rules.SpellStun, Duration: 1000}, effStun) // it cannot fight back
	w.HP = 5
	p.Char.Specs["Composite Bow"] = 20
	xp := p.Char.XP
	r.z.Shoot(p, "")
	for i := 0; i < 20*60 && !w.dead(); i++ {
		w.HP = min(w.HP, 5) // the rig's creature regenerates fast; keep it low
		r.ticks(1)
	}
	if !w.dead() {
		t.Fatalf("the wolf lived: %d of %d hp, %d arrows used", w.HP, w.MaxHP, 200-arrowsLeft(p))
	}
	if p.Char.XP <= xp {
		t.Error("no experience for the kill")
	}
	if p.shooting {
		t.Error("the archer kept shooting at a dead target")
	}
}

func TestArrowsAreSoldInPacks(t *testing.T) {
	r := shopRig(t)
	r.w.Cat.Items["arrows"] = &Item{ID: "arrows", Name: "standard arrows", ObjectType: rules.ObjArrow, ItemType: 40, Speed: 6.0, PackSize: 20, Price: 480}
	r.w.Cat.Stock["L1"] = append(r.w.Cat.Stock["L1"], "arrows")
	p := r.player("Buyer", 10, 30010, 30000, true)
	p.Char.Money = 1000
	r.z.Interact(p, merchantID)
	r.z.Buy(p, "arrows")
	found := 0
	for s, it := range p.inv {
		if it.ID == "arrows" && s < PackSize {
			found = it.Count
		}
	}
	if found != 20 || p.Char.Money != 520 {
		t.Errorf("bought %d arrows leaving %d, want a pack of 20 and 520", found, p.Char.Money)
	}
	// Selling the pack back pays once for the pack, not for each arrow.
	for s, it := range p.inv {
		if it.ID == "arrows" {
			before := p.Char.Money
			r.z.Sell(p, s)
			if got := p.Char.Money - before; got > 480 || got < 1 {
				t.Errorf("a pack of 20 sold for %d, more than its price of 480", got)
			}
		}
	}
}

func TestBowAndArrowsGoInTheirSlots(t *testing.T) {
	r, p, _ := archerRig(t, 0)
	r.st.Tx(func(tx *store.Tx) error { return tx.RemoveItem(p.Char.ID, SlotRanged) })
	delete(p.inv, SlotRanged)
	p.sh.abilities["Weaponry: Composite Bows"] = 1
	p.Char.Level = 20
	r.give(p, 0, "cbow", 1)
	r.give(p, 1, "arrows", 20)
	r.z.MoveItem(p, 0, SlotRanged)
	r.z.MoveItem(p, 1, SlotQuiver+2)
	if p.inv[SlotRanged] == nil || p.inv[SlotQuiver+2] == nil {
		t.Fatalf("bow %v, arrows %v", p.inv[SlotRanged] != nil, p.inv[SlotQuiver+2] != nil)
	}
	if p.bow == nil {
		t.Error("the bow is not in hand after equipping")
	}
	r.give(p, 2, "arrows", 5)
	r.z.MoveItem(p, 2, SlotWeapon)
	if p.inv[SlotWeapon] != nil && p.inv[SlotWeapon].ID == "arrows" {
		t.Error("arrows went in the weapon hand")
	}
}

// A Hunter on the imported data: every spell their lines give to level 50 can
// be cast, the Beastcraft summon makes a pet, the bow is theirs to use, and
// the three bow abilities come with the Composite Bow spec at the data's levels.
func TestAHunterOnTheImportedData(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	if _, ok := cat.StartFor(2, 5, rules.ClassHunter); !ok {
		t.Error("the data has no start for a Norseman Hunter")
	}
	total := 0
	for _, spec := range cat.TrainableSpecs(rules.ClassHunter) {
		for _, l := range cat.SpellsFor(rules.ClassHunter, spec) {
			sp := cat.Spells[l.Spell]
			if sp == nil || l.Level > 50 {
				continue
			}
			total++
			if !spellSupported(sp) {
				t.Errorf("a Hunter cannot cast %s %d %s (%s, %s)", spec, l.Level, sp.Name, sp.Type, sp.Target)
			}
		}
	}
	if total == 0 {
		t.Error("a Hunter has no spells in the data")
	}
	trained := map[string]int{"Composite Bow": 50}
	ab := rules.Abilities(cat.Grants[rules.ClassHunter], cat.Careers(rules.ClassHunter), 50, trained)
	if ab[rules.AbCriticalShot] != 9 || ab[rules.AbPenetrating] != 3 || ab[rules.AbRapidFire] != 2 || ab["Weaponry: Composite Bows"] == 0 {
		t.Errorf("a level 50 Hunter with Composite Bow 50: Critical Shot %d, Penetrating Arrow %d, Rapid Fire %d", ab[rules.AbCriticalShot], ab[rules.AbPenetrating], ab[rules.AbRapidFire])
	}
	r := newRig(t, wolf(1, 30400, 30000, 8))
	r.w.Cat = cat
	r.z.rng = rand.New(rand.NewSource(5))
	p := r.player("Hunter", 20, 30000, 30000, false)
	p.Char.Class = rules.ClassHunter
	p.Char.Specs = map[string]int{"Beastcraft": 20}
	r.w.recalc(p)
	r.w.refreshBook(p)
	p.power = 1e6
	sp := bestOfType(p, rules.SpellSummonHunterPet)
	if sp == nil {
		t.Fatal("a Hunter has no pet summon")
	}
	r.z.Cast(p, sp.ID, p.ID)
	r.ticks(int(sp.CastTime*TickRate) + 10)
	if p.pet == nil || !p.pet.isPet() {
		t.Errorf("%s made no pet", sp.Name)
	}
}
