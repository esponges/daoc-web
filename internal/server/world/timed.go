package world

import (
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// clockEvery is how often, in seconds, each client is sent the game clock.
const clockEvery = 30

// hide takes a timed creature out of the world: nothing sees it, targets it or
// fights it, and it does nothing until its time comes round.
func (n *NPC) hide() {
	n.hidden, n.gone, n.leaving = true, true, false
	n.state = npcIdle
	n.target, n.hate = nil, nil
}

// inFight says whether a creature is in a fight.
func (n *NPC) inFight() bool { return n.target != nil || n.state == npcChase || len(n.hate) > 0 }

// startTimed hides the timed creatures that are not in their time as a zone
// starts, and says how many it keeps.
func (z *Zone) startTimed() {
	gs := z.W.gameSeconds()
	night, day := 0, 0
	for _, n := range z.npcs {
		switch n.timed {
		case rules.NightOnly:
			night++
		case rules.DayOnly:
			day++
		default:
			continue
		}
		if !rules.Present(n.timed, n.appear, gs) {
			n.hide()
		}
	}
	if night+day > 0 {
		z.W.Log("zone %d %s: %d night-only and %d day-only spawns", z.ID, z.Name, night, day)
	}
}

// checkTimed runs each second: timed creatures arrive as their time begins,
// over its first game minute, and leave when it ends, a fighting one when the
// fight is over.
func (z *Zone) checkTimed() {
	gs := z.W.gameSeconds()
	for _, n := range z.npcs {
		if n.timed == rules.Always {
			continue
		}
		want := rules.Present(n.timed, n.appear, gs)
		switch {
		case want && n.hidden:
			n.hidden = false
			n.reset()
		case want:
			n.leaving = false
		case !n.hidden && n.inFight():
			n.leaving = true // stays until the fight ends
		case !n.hidden:
			n.hide()
		}
	}
}

// sendClock tells each player the game time of day.
func (z *Zone) sendClock() {
	m := protocol.Clock{T: protocol.TClock, GS: z.W.gameSeconds()}
	for _, p := range z.players {
		p.send(m)
	}
}
