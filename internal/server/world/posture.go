package world

import (
	"daocweb/internal/server/rules"
)

// Posture: a player sits to rest and stands to do anything else, and may
// sprint. See the player-postures capability. The server keeps the posture;
// other clients see it on the entity.

// Sit toggles sitting: /sit.
func (z *Zone) Sit(p *Player) {
	if p.seated {
		z.standUp(p)
		return
	}
	switch {
	case p.dead:
		return
	case p.ride != nil:
		p.errMsg("posture", "You cannot sit while riding.")
	case p.Body.Air:
		p.errMsg("posture", "You cannot sit in the air.")
	case swimming(z.G, p.Body.X, p.Body.Y):
		p.errMsg("posture", "You cannot sit here: you are swimming.")
	default:
		p.seated, p.sprint = true, false
		sys(p, "You sit down.")
	}
}

// Stand is /stand: up if seated.
func (z *Zone) Stand(p *Player) {
	if !p.seated {
		sys(p, "You are already standing.")
		return
	}
	z.standUp(p)
}

// standUp stands a seated player up, saying so. It is the one way out of
// sitting: moving, jumping, attacking, casting, shooting, using a style,
// emoting and being attacked all come here.
func (z *Zone) standUp(p *Player) {
	if p == nil || !p.seated {
		return
	}
	p.seated = false
	sys(p, "You stand up.")
}

// Sprint toggles sprint: /sprint.
func (z *Zone) Sprint(p *Player) {
	if p.sprint {
		p.sprint = false
		sys(p, "You stop sprinting.")
		z.effectsChanged(p)
		return
	}
	switch {
	case p.dead:
		return
	case p.seated:
		p.errMsg("posture", "You cannot sprint while sitting.")
	case swimming(z.G, p.Body.X, p.Body.Y):
		p.errMsg("posture", "You cannot sprint while swimming.")
	case p.end <= 0:
		p.errMsg("posture", "You are too tired to sprint.")
	default:
		p.sprint = true
		sys(p, "You start sprinting.")
		z.effectsChanged(p)
	}
}

// tickSprint spends a second of sprint's endurance, and turns it off when
// there is none left. moved says whether the body moved this second.
func (z *Zone) tickSprint(p *Player, moved bool) {
	if !p.sprint {
		return
	}
	if p.dead || p.seated || swimming(z.G, p.Body.X, p.Body.Y) {
		p.sprint = false
		z.effectsChanged(p)
		return
	}
	p.end = max(0, p.end-rules.SprintDrain(1, moved))
	if !rules.SprintContinues(p.end) {
		p.sprint = false
		sys(p, "You are too tired to go on sprinting.")
		z.effectsChanged(p)
	}
}

// posture is what other players see of a body: "sit", "air" or "swim", or ""
// standing. A drop under a few steps' height is not shown as a fall.
func (z *Zone) posture(p *Player) string {
	switch {
	case p.dead:
		return ""
	case p.seated:
		return "sit"
	case p.Body.Air && (p.Body.Jumped || p.Body.PeakZ-p.Body.Z > airShown):
		return "air"
	case !p.Body.Air && swimming(z.G, p.Body.X, p.Body.Y):
		return "swim"
	}
	return ""
}

// healthTick is the hit points a player's health tick returns: the tick's
// own amount, doubled when seated out of combat, and the regeneration buff.
func (z *Zone) healthTick(p *Player) float64 {
	fighting := z.now < p.combatUntil
	base := float64(rules.HealthRegen(p.Char.Level, fighting))
	return rules.SeatedRegen(base, p.seated, fighting) + float64(p.regenBuff()*b2i(!fighting))
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ifAir is a body's vertical speed while it is in the air, for the entity.
func ifAir(p *Player) float64 {
	if p.Body.Air {
		return p.Body.VZ
	}
	return 0
}

// airShown is how far a body must have dropped for the fall to be drawn.
const airShown = 60.0
