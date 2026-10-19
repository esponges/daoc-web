package world

// combatant is anything a creature can fight: a player, or another creature
// (a guard and its quarry).
type combatant interface {
	entityID() int64
	position() (x, y float64)
	// fightable says whether it can still be fought in a zone: alive, there,
	// and for a player, still connected.
	fightable(z *Zone) bool
}

func (p *Player) entityID() int64             { return p.ID }
func (p *Player) position() (float64, float64) { return p.Body.X, p.Body.Y }
func (p *Player) fightable(z *Zone) bool       { return !p.dead && !p.closed.Load() && p.Zone == z }

func (n *NPC) entityID() int64             { return n.ID }
func (n *NPC) position() (float64, float64) { return n.Body.X, n.Body.Y }
func (n *NPC) fightable(z *Zone) bool       { return !n.dead() && z.byID[n.ID] == n }

// hateEntry is how much a creature hates one combatant, and when it began,
// which settles ties in favour of the first.
type hateEntry struct {
	who   combatant
	hate  float64
	order int
}

// addHate makes a creature hate a combatant more: damage adds its amount, a
// blow that missed or an aggressive creature's notice a token amount.
func (n *NPC) addHate(c combatant, amount float64) {
	if n.hate == nil {
		n.hate = map[int64]*hateEntry{}
	}
	e := n.hate[c.entityID()]
	if e == nil {
		n.hateOrder++
		e = &hateEntry{who: c, order: n.hateOrder}
		n.hate[c.entityID()] = e
	}
	e.hate += amount
}

// forget drops a combatant from a creature's hate.
func (n *NPC) forget(c combatant) {
	delete(n.hate, c.entityID())
	if n.target == c {
		n.target = nil
	}
}

// hated is the combatant a creature hates most among those it can still
// fight in the zone, dropping the rest; nil when there is none.
func (n *NPC) hated(z *Zone) combatant {
	var best *hateEntry
	for id, e := range n.hate {
		if !e.who.fightable(z) {
			delete(n.hate, id)
			continue
		}
		if best == nil || e.hate > best.hate || e.hate == best.hate && e.order < best.order {
			best = e
		}
	}
	if best == nil {
		return nil
	}
	return best.who
}

// hates says whether a creature is angry with a combatant.
func (n *NPC) hates(c combatant) bool { return n.hate[c.entityID()] != nil }
