package world

import (
	"math/rand"
	"testing"

	"daocweb/internal/server/rules"
)

// abilityRig is a creature beside a Warrior (the tank) and a member of their
// group (the Skald), with the tank holding the abilities and facing the creature.
func abilityRig(t *testing.T, tankX float64) (*grouped, *Player, *Player, *NPC) {
	g := &grouped{rig: newRig(t, wolf(1, 30100, 30000, 10))}
	g.w.groups.clock = newGrouped(t).w.groups.clock
	n := g.z.npcs[0]
	n.Aggro, n.Roam = 0, 0
	n.cb.EvadePct, n.cb.ParryPct, n.cb.BlockPct = 0, 0, 0
	skald := g.join("Skald", 30000)
	tank := g.join("Tank", tankX)
	g.form(skald, tank)
	tank.sh.abilities = map[string]int{rules.AbShield: 1, rules.AbGuard: 1, rules.AbProtect: 2, rules.AbIntercept: 1, rules.AbEngage: 1}
	tank.shield = g.w.Cat.Items["shield"]
	tank.sh.stats[rules.DEX] = 100
	tank.Body.Yaw = angleTo(tank.Body.X, tank.Body.Y, n.Body.X, n.Body.Y)
	g.z.SetTarget(tank, skald.ID)
	return g, tank, skald, n
}

func TestUseAbilityNeedsAGroupMember(t *testing.T) {
	g, tank, _, _ := abilityRig(t, 30000)
	stranger := g.join("Stranger", 30050)
	g.z.SetTarget(tank, stranger.ID)
	g.z.UseAbility(tank, "guard")
	if out, _, _ := inbox(tank); !has(out, "must be in your group") {
		t.Errorf("tank was told %v, want a refusal about the group", out)
	}
	if tank.ab.guard != 0 {
		t.Error("guard was set on a stranger")
	}
}

func TestUseAbilityRefusedWithoutIt(t *testing.T) {
	_, tank, _, _ := abilityRig(t, 30000)
	delete(tank.sh.abilities, rules.AbProtect)
	tank.Zone.UseAbility(tank, "protect")
	if out, _, _ := inbox(tank); !has(out, "do not have Protect") {
		t.Errorf("tank was told %v, want they lack Protect", out)
	}
}

func TestUseAgainEndsIt(t *testing.T) {
	g, tank, skald, _ := abilityRig(t, 30000)
	g.z.UseAbility(tank, "guard")
	if tank.ab.guard != skald.ID {
		t.Fatal("guard not set")
	}
	g.z.UseAbility(tank, "guard")
	if tank.ab.guard != 0 {
		t.Error("using guard again did not end it")
	}
}

func TestUseAbilityByName(t *testing.T) {
	g, tank, skald, _ := abilityRig(t, 30000)
	g.z.SetTarget(tank, 0)
	g.z.UseAbility(tank, "guard", "skald")
	if tank.ab.guard != skald.ID {
		t.Error("guard by name did not set")
	}
}

// blocks counts the blows out of 300 a member is told a guard blocked.
func guardedBlocks(g *grouped, n *NPC, skald *Player) int {
	drain(skald)
	g.z.rng = rand.New(rand.NewSource(1))
	blocks := 0
	for i := 0; i < 300; i++ {
		g.z.npcHitsPlayer(n, skald)
		skald.hp = float64(skald.maxHP())
		out, _, _ := inbox(skald)
		if has(out, "blocks the blow for you") {
			blocks++
		}
	}
	return blocks
}

func TestGuardBlocksForAMemberInReachAndFront(t *testing.T) {
	for _, c := range []struct {
		name    string
		x       float64
		turn    bool // the tank turns his back on the creature
		blocked bool
	}{{"near and facing", 30150, false, true}, {"too far", 30500, false, false}, {"back turned", 30150, true, false}} {
		g, tank, skald, n := abilityRig(t, c.x)
		if c.turn {
			tank.Body.Yaw += 3.14
		}
		g.z.UseAbility(tank, "guard")
		if got := guardedBlocks(g, n, skald); (got > 0) != c.blocked {
			t.Errorf("%s: guard blocked %d of 300 blows, want blocks=%v", c.name, got, c.blocked)
		}
	}
}

func TestInterceptTakesOneBlowThenEndsAndRecharges(t *testing.T) {
	g, tank, skald, n := abilityRig(t, 30080)
	g.z.UseAbility(tank, "intercept")
	if tank.ab.intercept != skald.ID {
		t.Fatal("intercept not set")
	}
	taken := false
	tankFull := tank.hp
	for i := 0; i < 80 && tank.ab.intercept != 0; i++ {
		g.z.rng = rand.New(rand.NewSource(int64(i)))
		skald.hp = float64(skald.maxHP())
		before := skald.hp
		g.z.npcHitsPlayer(n, skald)
		if tank.ab.intercept == 0 {
			taken = true
			if skald.hp != before {
				t.Errorf("the Skald lost hit points to the intercepted blow")
			}
		}
	}
	if !taken || tank.hp >= tankFull {
		t.Fatalf("intercept taken %v, tank hp %v of %v: want the blow to land on the Warrior", taken, tank.hp, tankFull)
	}
	g.z.UseAbility(tank, "intercept")
	if tank.ab.intercept != 0 {
		t.Error("intercept was usable again before 60 seconds")
	}
	g.z.now += rules.InterceptCooldown
	g.z.UseAbility(tank, "intercept")
	if tank.ab.intercept != skald.ID {
		t.Error("intercept was not usable after 60 seconds")
	}
}

func TestInterceptStaysArmedWhenOutOfRange(t *testing.T) {
	g, tank, skald, n := abilityRig(t, 30300)
	g.z.UseAbility(tank, "intercept")
	for i := 0; i < 60; i++ {
		g.z.npcHitsPlayer(n, skald)
	}
	if tank.ab.intercept != skald.ID {
		t.Error("intercept ended without a blow being taken in range")
	}
}

func TestProtectMovesHate(t *testing.T) {
	g, tank, skald, n := abilityRig(t, 30000)
	g.z.UseAbility(tank, "protect")
	g.z.provoke(n, skald, 100)
	if h := n.hate[skald.entityID()]; h == nil || h.hate != 80 {
		t.Errorf("hate for the protected member %v, want 80", h)
	}
	if h := n.hate[tank.entityID()]; h == nil || h.hate != 20 {
		t.Errorf("hate for the protecting Warrior %v, want 20 moved to them", h)
	}
}

// engagedBlocks counts blows blocked (told by endurance spent) of 40 from a creature.
func engagedBlocks(g *grouped, n *NPC, tank *Player) (blocks int) {
	g.z.rng = rand.New(rand.NewSource(5))
	for i := 0; i < 40; i++ {
		tank.end, tank.hp = 100, float64(tank.maxHP())
		tank.lastDef = defence{}
		g.z.npcHitsPlayer(n, tank)
		if tank.lastDef.kind == "block" {
			blocks++
		}
	}
	return
}

func TestEngageRaisesBlockAgainstOneCreature(t *testing.T) {
	g, tank, _, n := abilityRig(t, 30000)
	plain := engagedBlocks(g, n, tank)
	g.z.SetTarget(tank, n.ID)
	g.z.UseAbility(tank, "engage")
	if tank.ab.engage != n.ID {
		t.Fatal("engage not set")
	}
	engaged := engagedBlocks(g, n, tank)
	if engaged <= plain+10 {
		t.Errorf("blocked %d of 40 engaged against %d plain, want far more", engaged, plain)
	}
	tank.end = 100
	g.z.npcHitsPlayer(n, tank)
	if tank.end != 100-rules.EngageCost {
		t.Errorf("endurance %v after a blow, want %d spent", tank.end, rules.EngageCost)
	}
	tank.end = 3
	g.z.npcHitsPlayer(n, tank)
	if tank.ab.engage != 0 {
		t.Error("engage did not end when the Warrior ran out of endurance")
	}
}

func TestEngageRefusedWhileOthersAttackTheCreature(t *testing.T) {
	g, tank, skald, n := abilityRig(t, 30000)
	g.z.now = 100
	g.z.provoke(n, skald, 5)
	g.z.SetTarget(tank, n.ID)
	g.z.UseAbility(tank, "engage")
	if tank.ab.engage != 0 {
		t.Error("engage was allowed on a creature attacked a moment ago")
	}
	g.z.now += rules.EngageQuiet
	g.z.UseAbility(tank, "engage")
	if tank.ab.engage != n.ID {
		t.Error("engage was refused 5 seconds after the last attack")
	}
	g.z.SetAttack(tank, true)
	if tank.ab.engage != 0 {
		t.Error("attacking did not end engage")
	}
}
