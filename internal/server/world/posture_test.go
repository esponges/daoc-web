package world

import (
	"encoding/json"
	"math"
	"math/rand"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

func heard(p *Player) string {
	out, _, _ := inbox(p)
	return strings.Join(out, "|")
}

func TestSitAndStand(t *testing.T) {
	r := newRig(t)
	p := r.player("Rester", 20, 30000, 30000, false)
	inbox(p)
	r.z.Sit(p)
	if !p.seated || !strings.Contains(heard(p), "You sit down.") {
		t.Fatalf("seated %v after /sit", p.seated)
	}
	if got := r.z.playerEntity(p).Posture; got != "sit" {
		t.Errorf("the entity's posture is %q, want sit", got)
	}
	r.z.Sit(p)
	if p.seated || !strings.Contains(heard(p), "You stand up.") {
		t.Errorf("seated %v after /sit again", p.seated)
	}
	r.z.Stand(p)
	if !strings.Contains(heard(p), "already standing") {
		t.Error("/stand while standing was not answered")
	}
}

func TestSittingIsRefusedWhereItCannotBeDone(t *testing.T) {
	r := newRig(t)
	p := r.player("Rester", 20, 30000, 30000, false)
	p.Body.Air = true
	r.z.Sit(p)
	if p.seated || !strings.Contains(lastErr(t, p), "in the air") {
		t.Errorf("in the air: seated %v, error %q", p.seated, lastErr(t, p))
	}
	p.Body.Air = false
	r.z.G = pond{}
	r.z.Sit(p)
	if p.seated || !strings.Contains(lastErr(t, p), "swimming") {
		t.Errorf("swimming: seated %v, error %q", p.seated, lastErr(t, p))
	}
	r.z.G = flat{}
	p.dead = true
	r.z.Sit(p)
	if p.seated {
		t.Error("a dead player sat")
	}
}

// Everything that stands a seated player up.
func TestMovingJumpingAndTurningStandYouUp(t *testing.T) {
	for name, run := range map[string][5]int{
		"moving":  {1, 1, 0, 0, 0},
		"strafe":  {1, 0, 1, 0, 0},
		"turning": {1, 0, 0, 1, 0},
		"jumping": {1, 0, 0, 0, protocol.FlagJump},
	} {
		t.Run(name, func(t *testing.T) {
			r := newRig(t)
			p := r.player("Rester", 20, 30000, 30000, false)
			r.z.Sit(p)
			p.QueueInput(&protocol.Input{Steps: [][5]int{run}})
			r.ticks(1)
			if p.seated {
				t.Errorf("%s left them seated", name)
			}
		})
	}
}

func TestAnIdleStepKeepsYouSeated(t *testing.T) {
	r := newRig(t)
	p := r.player("Rester", 20, 30000, 30000, false)
	r.z.Sit(p)
	idleStep(p)
	r.ticks(1)
	if !p.seated {
		t.Error("a step with no keys stood them up")
	}
}

func TestAttackingStandsYouUp(t *testing.T) {
	r, p, _ := archerRig(t, 5)
	r.z.Sit(p)
	r.z.SetAttack(p, true)
	if p.seated {
		t.Error("starting to attack left them seated")
	}
}

func TestBeingAttackedStandsYouUp(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 10))
	n := r.z.npcs[0]
	p := r.player("Rester", 20, 30000, 30000, true)
	r.z.Sit(p)
	r.z.npcHitsPlayer(n, p)
	if p.seated {
		t.Error("a creature's blow left them seated")
	}
}

func TestShootingStandsYouUp(t *testing.T) {
	r, p, _ := archerRig(t, 5)
	r.z.Sit(p)
	r.z.Shoot(p, "")
	if p.draw == nil {
		t.Fatalf("the bow was not drawn: %q", lastErr(t, p))
	}
	if p.seated {
		t.Error("drawing a bow left them seated")
	}
}

func TestCastingStandsYouUp(t *testing.T) {
	r := areaWorld(t)
	p := caster(r, "Thane", rules.ClassThane, 5)
	r.z.Sit(p)
	r.z.Cast(p, 3521, 0)
	if p.seated {
		t.Error("casting left them seated")
	}
}

// What a seated character gets back, against a standing one, tick by tick.
func TestSeatedVitalsReturnTwiceAsFast(t *testing.T) {
	r := newRig(t)
	stand := r.player("Stander", 20, 30000, 30000, false)
	sit := r.player("Sitter", 20, 30100, 30000, false)
	r.z.Sit(sit)
	for _, p := range []*Player{stand, sit} {
		p.hp, p.end, p.power = 100, 40, 0
		p.sh.maxPower = 500
		p.regenAt, p.healAt, p.powerAt = 0, 0, 0
	}
	r.z.rng = rand.New(zeroSource{})
	r.ticks(1)
	if got, want := sit.hp-100, 2*(stand.hp-100); got != want || stand.hp == 100 {
		t.Errorf("a health tick: seated %v, standing %v", sit.hp-100, stand.hp-100)
	}
	if got, want := sit.end-40, 2*(stand.end-40); got != want || stand.end == 40 {
		t.Errorf("an endurance tick: seated %v, standing %v", sit.end-40, stand.end-40)
	}
	if got, want := sit.power, 2*stand.power; got < want-1e-9 || got > want+1e-9 || stand.power == 0 {
		t.Errorf("a power tick: seated %v, standing %v", sit.power, stand.power)
	}
}

func TestSeatedRegenerationStopsInCombat(t *testing.T) {
	r := newRig(t)
	p := r.player("Sitter", 20, 30000, 30000, false)
	r.z.Sit(p)
	p.combatUntil = r.z.now + 30
	p.hp, p.end, p.healAt, p.regenAt = 100, 40, 0, 0
	r.ticks(1)
	if p.hp != 100 {
		t.Errorf("health rose to %v in combat", p.hp)
	}
	if p.end != 42 {
		t.Errorf("endurance %v after a combat second, want 42 (no seated bonus)", p.end)
	}
}

func runSecond(p *Player, r *rig, run [5]int) {
	run[0] = 60
	p.QueueInput(&protocol.Input{Steps: [][5]int{run}})
	r.ticks(21)
}

func TestSprintGoesFasterAndCostsEndurance(t *testing.T) {
	r := newRig(t)
	p := r.player("Sprinter", 20, 30000, 30000, false)
	p.end = 100
	r.z.Sprint(p)
	if !p.sprint {
		t.Fatalf("sprint is off after /sprint: %q", lastErr(t, p))
	}
	x0, y0 := p.Body.X, p.Body.Y
	runSecond(p, r, [5]int{0, 1, 0, 0, protocol.FlagRun})
	if d := hypot(p.Body.X-x0, p.Body.Y-y0); d < rules.SprintSpeed(Run)-2 || d > rules.SprintSpeed(Run)+2 {
		t.Errorf("a second of sprinting moved %.1f, want about %.1f", d, rules.SprintSpeed(Run))
	}
	if p.end > 100-rules.SprintCost+1 || p.end < 100-2*rules.SprintCost {
		t.Errorf("endurance after a sprinting second is %v", p.end)
	}
}

func TestStandingStillWithSprintOnCostsNothing(t *testing.T) {
	r := newRig(t)
	p := r.player("Sprinter", 20, 30000, 30000, false)
	p.end = 60
	r.z.Sprint(p)
	r.ticks(60) // three seconds
	if p.end < 60 || !p.sprint {
		t.Errorf("standing still: endurance %v, sprint %v", p.end, p.sprint)
	}
}

func TestSprintEndsWhenEnduranceRunsOut(t *testing.T) {
	r := newRig(t)
	p := r.player("Sprinter", 20, 30000, 30000, false)
	p.end = 8
	r.z.Sprint(p)
	inbox(p)
	for i := 0; i < 4 && p.sprint; i++ {
		runSecond(p, r, [5]int{0, 1, 0, 0, protocol.FlagRun})
	}
	if p.sprint || p.end != 0 {
		t.Errorf("sprint %v, endurance %v after running dry", p.sprint, p.end)
	}
	if !said(p, "too tired") {
		t.Error("they were not told why sprint ended")
	}
}

func TestSprintIsRefusedSeatedSwimmingOrDead(t *testing.T) {
	r := newRig(t)
	p := r.player("Sprinter", 20, 30000, 30000, false)
	p.end = 100
	r.z.Sit(p)
	r.z.Sprint(p)
	if p.sprint {
		t.Error("sprint turned on seated")
	}
	r.z.Stand(p)
	r.z.G = pond{}
	r.z.Sprint(p)
	if p.sprint {
		t.Error("sprint turned on while swimming")
	}
	r.z.G = flat{}
	p.end = 0
	r.z.Sprint(p)
	if p.sprint {
		t.Error("sprint turned on with no endurance")
	}
	p.end = 100
	r.z.Sprint(p)
	r.z.Sit(p)
	if p.sprint {
		t.Error("sitting left sprint on")
	}
}

func TestEffectsMessageCarriesSprint(t *testing.T) {
	r := newRig(t)
	p := r.player("Sprinter", 20, 30000, 30000, false)
	p.end = 100
	snaps(p)
	r.z.Sprint(p)
	found := false
	for {
		select {
		case b := <-p.out:
			var m protocol.Effects
			if json.Unmarshal(b, &m) == nil && m.T == protocol.TEffects && m.Spr {
				found = true
			}
			continue
		default:
		}
		break
	}
	if !found {
		t.Error("no effects message with spr after /sprint")
	}
}

func hypot(x, y float64) float64 { return math.Hypot(x, y) }
