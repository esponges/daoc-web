package world

// The torch: a light a player carries at night, toggled with T (the stock
// keyboard's Torch action) or /torch. Everyone near sees it, so a player may
// put it out to be harder to spot. The server keeps whether it is lit and puts
// it on the player's entity; the client draws the light. A new player's torch
// is lit; the client sends its remembered choice when it connects.

// Torch is /torch: no argument toggles, "on" and "off" set it.
func (z *Zone) Torch(p *Player, arg string) {
	lit := !p.torchOut
	switch arg {
	case "on":
		lit = true
	case "off":
		lit = false
	case "":
		lit = !lit
	default:
		sys(p, "Use /torch, /torch on or /torch off.")
		return
	}
	p.torchOut = !lit
	if lit {
		sys(p, "You light your torch.")
	} else {
		sys(p, "You put out your torch.")
	}
}
