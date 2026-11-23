package world

import (
	"math"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// Trade limits: how close two players must be, and how many items one side
// may offer.
const (
	TradeRange = 512.0
	TradeSlots = 10
)

// tradeSide is what one player has put up: pack slots, money, and whether
// they have accepted the offer as it stands.
type tradeSide struct {
	slots []int
	money int64
	ok    bool
}

// trade is two players bargaining face to face. It lives in their zone, and
// moves nothing until both have accepted the same offer.
type trade struct {
	p    [2]*Player
	side [2]tradeSide
}

func (t *trade) idx(p *Player) int {
	if t.p[0] == p {
		return 0
	}
	return 1
}

// TradeOp runs a player's request in a trade. It is called in their zone.
func (z *Zone) TradeOp(p *Player, m *protocol.TradeOp) {
	if m.Op == "open" {
		z.openTrade(p, m.ID)
		return
	}
	t := p.trade
	if t == nil {
		p.errMsg("trade", "You are not trading.")
		return
	}
	me := t.idx(p)
	switch m.Op {
	case "add":
		it := p.inv[m.Slot]
		switch {
		case m.Slot < 0 || m.Slot >= PackSize || it == nil:
			p.errMsg("trade", "You have nothing there to offer.")
			return
		case it.ID == taskItemID:
			p.errMsg("trade", "A task item cannot be traded.")
			return
		case len(t.side[me].slots) >= TradeSlots:
			p.errMsg("trade", "You can offer only 10 items.")
			return
		}
		for _, s := range t.side[me].slots {
			if s == m.Slot {
				return
			}
		}
		t.side[me].slots = append(t.side[me].slots, m.Slot)
	case "remove":
		kept := t.side[me].slots[:0]
		for _, s := range t.side[me].slots {
			if s != m.Slot {
				kept = append(kept, s)
			}
		}
		t.side[me].slots = kept
	case "money":
		if m.Copper < 0 || m.Copper > p.Char.Money {
			p.errMsg("trade", "You do not have that much.")
			return
		}
		t.side[me].money = m.Copper
	case "accept":
		t.side[me].ok = true
		if t.side[0].ok && t.side[1].ok {
			z.commitTrade(t)
			return
		}
		z.sendTrade(t)
		return
	case "cancel":
		z.endTrade(t, false, p.Char.Name+" cancels the trade.")
		return
	default:
		p.errMsg("trade", "Unknown trade request.")
		return
	}
	// Any change to either offer takes both accepts back.
	t.side[0].ok, t.side[1].ok = false, false
	z.sendTrade(t)
}

// openTrade starts a trade with the named player, or the one targeted.
func (z *Zone) openTrade(p *Player, id int64) {
	if id == 0 {
		id = p.target
	}
	o := z.players[id]
	switch {
	case p.dead:
		p.errMsg("trade", "You cannot trade while dead.")
		return
	case p.trade != nil:
		p.errMsg("trade", "You are already trading.")
		return
	case o == nil || o == p || o.dead || o.closed.Load():
		p.errMsg("trade", "Target a player to trade with.")
		return
	case o.Arch.Realm != p.Arch.Realm:
		p.errMsg("trade", "You cannot trade with that player.")
		return
	case o.trade != nil:
		p.errMsg("trade", o.Char.Name+" is already trading.")
		return
	case z.now < p.combatUntil || z.now < o.combatUntil:
		p.errMsg("trade", "You cannot trade while in combat.")
		return
	case math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) > TradeRange:
		p.errMsg("trade", "You are too far away to trade.")
		return
	}
	t := &trade{p: [2]*Player{p, o}}
	p.trade, o.trade = t, t
	sys(p, "You begin trading with "+o.Char.Name+".")
	sys(o, p.Char.Name+" begins trading with you.")
	z.sendTrade(t)
}

// sendTrade tells both players how the trade stands.
func (z *Zone) sendTrade(t *trade) {
	for i, p := range t.p {
		p.send(protocol.Trade{T: protocol.TTrade, With: t.p[1-i].Char.Name,
			Mine: z.sideView(t.p[i], t.side[i]), Theirs: z.sideView(t.p[1-i], t.side[1-i])})
	}
}

func (z *Zone) sideView(p *Player, s tradeSide) protocol.TradeSide {
	v := protocol.TradeSide{Items: []protocol.InvItem{}, Money: s.money, OK: s.ok}
	for _, slot := range s.slots {
		if it := p.inv[slot]; it != nil {
			if x := z.W.Cat.Items[it.ID]; x != nil {
				v.Items = append(v.Items, itemView(slot, x, it.Count))
			}
		}
	}
	return v
}

// endTrade closes a trade for both players, with the same line to each.
func (z *Zone) endTrade(t *trade, done bool, why string) {
	for _, p := range t.p {
		if p.trade != t {
			continue
		}
		p.trade = nil
		p.send(protocol.TradeClose{T: protocol.TTradeClose, Done: done, Why: why})
		if !done && why != "" {
			sys(p, why)
		}
	}
}

// checkTrade ends a trade whose players have drifted apart, died, gone into
// combat or left the game. Run each tick for the first player of a trade.
func (z *Zone) checkTrade(p *Player) {
	t := p.trade
	if t == nil || t.p[0] != p {
		return
	}
	a, b := t.p[0], t.p[1]
	switch {
	case a.closed.Load() || b.closed.Load() || z.players[a.ID] != a || z.players[b.ID] != b:
		z.endTrade(t, false, "The trade is cancelled.")
	case a.dead || b.dead:
		z.endTrade(t, false, "The trade is cancelled: someone died.")
	case z.now < a.combatUntil || z.now < b.combatUntil:
		z.endTrade(t, false, "The trade is cancelled: you are in combat.")
	case math.Hypot(a.Body.X-b.Body.X, a.Body.Y-b.Body.Y) > TradeRange:
		z.endTrade(t, false, "The trade is cancelled: you moved too far apart.")
	}
}

// packUsed is how many pack slots hold something.
func packUsed(p *Player) int {
	n := 0
	for s := range p.inv {
		if s >= 0 && s < PackSize {
			n++
		}
	}
	return n
}

// freePackSlots lists the first n pack slots that are empty, counting the ones
// in giving as empty (they leave in the same step).
func freePackSlots(p *Player, giving []int, n int) []int {
	gone := map[int]bool{}
	for _, s := range giving {
		gone[s] = true
	}
	var out []int
	for s := 0; s < PackSize && len(out) < n; s++ {
		if p.inv[s] == nil || gone[s] {
			out = append(out, s)
		}
	}
	return out
}

// commitTrade moves everything in one step, or nothing: it checks that each
// still holds what they offered and can pay, and that each has room for what
// they get, then writes both characters in one transaction.
func (z *Zone) commitTrade(t *trade) {
	fail := func(why string) {
		t.side[0].ok, t.side[1].ok = false, false
		z.endTrade(t, false, why)
	}
	for i, p := range t.p {
		for _, s := range t.side[i].slots {
			if p.inv[s] == nil {
				fail("The trade cannot be completed: an offered item is gone.")
				return
			}
		}
		if p.Char.Money < t.side[i].money {
			fail("The trade cannot be completed: not enough money.")
			return
		}
	}
	var into [2][]int // the pack slots each receives into
	for i, p := range t.p {
		other := 1 - i
		into[i] = freePackSlots(p, t.side[i].slots, len(t.side[other].slots))
		if len(into[i]) < len(t.side[other].slots) {
			fail("The trade cannot be completed: not enough room.")
			return
		}
	}
	moved := [2][]invItem{} // the items each receives, as they will sit
	for i := range t.p {
		other := 1 - i
		for k, s := range t.side[other].slots {
			it := *t.p[other].inv[s]
			it.Slot, it.DBID = into[i][k], 0
			moved[i] = append(moved[i], it)
		}
	}
	before := [2]int64{t.p[0].Char.Money, t.p[1].Char.Money}
	t.p[0].Char.Money += t.side[1].money - t.side[0].money
	t.p[1].Char.Money += t.side[0].money - t.side[1].money
	err := z.W.Store.Tx(func(tx *store.Tx) error {
		for i, p := range t.p {
			for _, s := range t.side[i].slots {
				if err := tx.RemoveItem(p.Char.ID, s); err != nil {
					return err
				}
			}
		}
		for i, p := range t.p {
			for _, it := range moved[i] {
				if err := tx.AddItemCond(p.Char.ID, it.ID, it.Slot, it.Count, it.Cond); err != nil {
					return err
				}
				if it.Poison > 0 {
					if err := tx.SetPoison(p.Char.ID, it.Slot, it.Poison, it.Charges); err != nil {
						return err
					}
				}
			}
			if err := tx.SaveCharacter(p.Char); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.p[0].Char.Money, t.p[1].Char.Money = before[0], before[1]
		fail("The trade could not be saved; nothing moved.")
		return
	}
	for i, p := range t.p {
		for _, s := range t.side[i].slots {
			delete(p.inv, s)
		}
	}
	for i, p := range t.p {
		for k := range moved[i] {
			it := moved[i][k]
			p.inv[it.Slot] = &it
		}
		z.W.recalcGear(p)
		z.buildBar(p)
		p.sendInv(z.W.Cat, p.Char.Money)
	}
	for _, p := range t.p {
		sys(p, "The trade is complete.")
	}
	z.endTrade(t, true, "")
}
