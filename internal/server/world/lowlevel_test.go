package world

import (
	"math"
	"math/rand"
	"testing"

	"daocweb/internal/server/store"
)

// classicKit is the Warrior's starting equipment as the data gives it: a
// 1.2 DPS, 2.5 second training axe of 90% quality, and bronze stelskodd armour
// of AF 4, 19% absorption, 85% quality.
func classicKit(c *Catalog) {
	c.Items["training_axe"] = &Item{ID: "training_axe", Name: "training axe", Quality: 90, DPS: 1.2, Speed: 2.5, ObjectType: 13, ItemType: 11, Hand: 2, Realm: 2, DamageType: 2}
	for slot, it := range map[int]int{SlotTorso: 25, SlotLegs: 27, SlotArms: 28, SlotFeet: 23, SlotHands: 22, SlotHead: 21} {
		id := "stelskodd" + string(rune('a'+slot-SlotTorso))
		c.Items[id] = &Item{ID: id, Name: "bronze stelskodd", Level: 2, Quality: 85, AF: 4, Absorb: 19, ObjectType: 34, ItemType: it, Realm: 2}
	}
	c.StartEquip = map[int][]string{22: {"training_axe", "stelskodda", "stelskoddb", "stelskoddc", "stelskoddd", "stelskodde", "stelskoddf"}}
}

// fight puts a new level 1 Warrior in the classic kit against one creature
// and reports whether the Warrior won.
func fight(t *testing.T, seed int64, creatureLevel int) bool {
	r := newRig(t, wolf(1, 30100, 30000, creatureLevel))
	r.z.rng = rand.New(rand.NewSource(seed))
	classicKit(r.w.Cat)
	n := r.z.byID[1]
	n.Roam, n.Aggro = 0, 0
	a, _ := r.st.AccountFor("d:lowlevel", "lowlevel", false)
	c := &store.Character{AccountID: a.ID, Name: "Newcomer", Race: 6, Class: 22, Level: 1, Zone: 100, X: 30000, Y: 30000, XPScheme: store.XPSchemeRules}
	stats, _ := CreationStats(6, 22, nil)
	c.Stats, c.HasStats = stats, true
	r.st.CreateCharacter(c)
	r.w.GiveStarterKit(c)
	p := newPlayer(a, c)
	r.z.join(p)
	p.Body.Yaw = math.Pi / 2
	r.z.SetTarget(p, 1)
	r.z.SetAttack(p, true)
	for i := 0; i < TickRate*180 && !n.dead() && !p.dead; i++ {
		r.z.Tick()
	}
	return n.dead()
}

// A new Warrior in the classic kit beats a yellow creature (their own level)
// and loses to an orange one (a level above, at level 1): a creature's
// armour and damage grow from near nothing, so one level counts for a lot.
func TestLowLevelFights(t *testing.T) {
	for _, c := range []struct {
		level   int
		atLeast int
		atMost  int
	}{{1, 15, 20}, {2, 0, 5}} {
		wins := 0
		for seed := int64(1); seed <= 20; seed++ {
			if fight(t, seed, c.level) {
				wins++
			}
		}
		t.Logf("level 1 Warrior against a level %d creature: %d of 20 won", c.level, wins)
		if wins < c.atLeast || wins > c.atMost {
			t.Errorf("level %d creature: %d of 20 won, want %d to %d", c.level, wins, c.atLeast, c.atMost)
		}
	}
}
