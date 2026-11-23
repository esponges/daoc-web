package world

import (
	"fmt"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// airGrace is how long a body may hang in the air with no input from its
// client before the server steps it down itself, in seconds. Longer than any
// gap between a client's batches.
const airGrace = 0.5

// landing settles a landing a step reported: what the fall costs, and a death
// if it costs everything. A landing in water costs nothing.
func (z *Zone) landing(p *Player, m Motion) {
	if !m.Landed || p.dead {
		return
	}
	dmg := rules.FallDamage(m.FallHeight, p.maxHP(), rules.SafeFallLevel(p.stealthSpec(z.W.Cat)), m.Water)
	if dmg <= 0 {
		return
	}
	p.hp -= float64(dmg)
	z.event(protocol.Event{K: "fall", To: p.ID, Amount: dmg, Text: fmt.Sprintf("You take %d damage from the fall!", dmg)}, p.Body.X, p.Body.Y)
	if p.hp <= 0 {
		z.killPlayer(p, nil)
	}
}

// settleAir steps a body in the air whose client has stopped sending input,
// so that it comes down: these steps are the server's own and are not counted
// among the client's.
func (z *Zone) settleAir(p *Player) {
	if !p.Body.Air || len(p.queue) > 0 || z.now-p.lastStepAt <= airGrace {
		return
	}
	for i := 0; i < int(float64(protocol.StepsPerSecond)*TickDT); i++ {
		k := Keys{Slow: 1 - speedFactor(p)}
		z.landing(p, Step(z.G, z.Mv, &p.Body, k))
	}
}
