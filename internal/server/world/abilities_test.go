package world

import (
	"math"
	"math/rand"
	"testing"

	"daocweb/internal/server/rules"
)

// abilityRig is a creature beside a Warrior (the tank) and a member of their
// group (the Skald), with the tank holding the abilities.
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
	g.z.SetTarget(tank, skald.ID)
	return g, tank, skald, n
}

func TestUseAbilityNeedsAGroupMember(t *testing.T) {
	g, tank, _, _ := abilityRig(t, 30100)
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
	_, tank, _, _ := abilityRig(t, 30100)
	delete(tank.sh.abilities, rules.AbProtect)
	tank.Zone.UseAbility(tank, "protect")
	if out, _, _ := inbox(tank); !has(out, "do not have Protect") {
		t.Errorf("tank was told %v, want they lack Protect", out)
	}
}

func TestUseAgainEndsIt(t *testing.T) {
	g, tank, skald, _ := abilityRig(t, 30100)
	g.z.UseAbility(tank, "guard")
	if tank.ab.guard != skald.ID {
		t.Fatal("guard not set")
	}
	g.z.UseAbility(tank, "guard")
	if tank.ab.guard != 0 {
		t.Error("using guard again did not end it")
	}
}

func TestGuardBlocksForAMemberInReach(t *testing.T) {
	for _, c := range []struct {
		x       float64
		blocked bool
	}{{30150, true}, {30500, false}} {
		g, tank, skald, n := abilityRig(t, c.x)
		g.z.UseAbility(tank, "guard")
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
		if (blocks > 0) != c.blocked {
			t.Errorf("guard at x=%v blocked %d of 300 blows, want blocks=%v", c.x, blocks, c.blocked)
		}
	}
}

func TestInterceptTakesTheNextBlowOnce(t *testing.T) {
	g, tank, skald, n := abilityRig(t, 30080)
	g.z.UseAbility(tank, "intercept")
	if tank.ab.intercept != skald.ID {
		t.Fatal("intercept not set")
	}
	full, tankFull := skald.hp, tank.hp
	for i := 0; i < 40 && tank.hp == tankFull; i++ {
		g.z.rng = rand.New(rand.NewSource(int64(i)))
		g.z.npcHitsPlayer(n, skald)
		if tank.ab.intercept == 0 {
			break
		}
	}
	if tank.ab.intercept != 0 {
		t.Fatal("intercept did not end after a blow")
	}
	if skald.hp != full {
		t.Errorf("the Skald lost %v hit points to an intercepted blow", full-skald.hp)
	}
	// Used again at once it is refused for the cooldown.
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

func TestProtectDrawsHate(t *testing.T) {
	g, tank, skald, n := abilityRig(t, 30100)
	g.z.UseAbility(tank, "protect")
	n.target = skald
	for i := 0; i < 40 && n.hate[skald.entityID()] == nil; i++ {
		g.z.rng = rand.New(rand.NewSource(int64(i)))
		g.z.npcHitsPlayer(n, skald)
	}
	h := n.hate[tank.entityID()]
	if h == nil || h.hate <= 0 {
		t.Fatalf("hate for the protecting tank is %v, want some", h)
	}
}

func TestEngageBlocksFrontBlowsAndEnds(t *testing.T) {
	g, tank, _, n := abilityRig(t, 30100)
	tank.Body.X, tank.Body.Yaw = 30000, math.Pi/2 // facing the creature at +x
	g.z.SetTarget(tank, n.ID)
	g.z.UseAbility(tank, "engage")
	if tank.ab.engage != n.ID {
		t.Fatal("engage not set")
	}
	tank.end = 100
	g.z.rng = rand.New(rand.NewSource(5))
	blocked := 0
	for i := 0; i < 20; i++ {
		before := tank.end
		g.z.npcHitsPlayer(n, tank)
		tank.hp = float64(tank.maxHP())
		if before-tank.end == rules.EngageCost {
			blocked++
		}
		tank.end = 100
	}
	if blocked < 15 {
		t.Errorf("engage blocked %d of 20 frontal blows, want about 95%%", blocked)
	}
	g.z.SetTarget(tank, 0)
	if tank.ab.engage != 0 {
		t.Error("engage did not end when the target changed")
	}
}

func TestEngageRefusedAfterBeingHit(t *testing.T) {
	g, tank, _, n := abilityRig(t, 30100)
	g.z.SetTarget(tank, n.ID)
	g.z.now = 100
	tank.ab.attackedAt = g.z.now - 2
	g.z.UseAbility(tank, "engage")
	if tank.ab.engage != 0 {
		t.Error("engage was allowed 2 seconds after a blow")
	}
}
