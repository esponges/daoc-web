package world

import (
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
)

// tradeRig is a zone with two players of one realm, close together, who each
// hold something to trade.
func tradeRig(t *testing.T) (*rig, *Player, *Player) {
	r := newRig(t)
	a := r.player("Thora", 10, 30000, 30000, false)
	b := r.player("Bjorn", 10, 30200, 30000, false)
	r.give(a, 0, "axe", 1)
	r.give(b, 0, "pelt", 3)
	a.Char.Money, b.Char.Money = 500, 100
	return r, a, b
}

func tradeOp(r *rig, p *Player, op string, slot int, copper int64) {
	r.z.TradeOp(p, &protocol.TradeOp{Op: op, Slot: slot, Copper: copper})
}

func openTrade(t *testing.T, r *rig, a, b *Player) {
	t.Helper()
	a.target = b.ID
	r.z.TradeOp(a, &protocol.TradeOp{Op: "open"})
	if a.trade == nil || a.trade != b.trade {
		t.Fatal("the trade did not open")
	}
}

func TestTradeOpensWithEmptyOffers(t *testing.T) {
	r, a, b := tradeRig(t)
	lines(a)
	openTrade(t, r, a, b)
	var seen int
	for _, p := range []*Player{a, b} {
		for _, m := range sentOf[protocol.Trade](p, protocol.TTrade) {
			seen++
			if len(m.Mine.Items) != 0 || len(m.Theirs.Items) != 0 {
				t.Errorf("a new trade shows offers: %+v", m)
			}
		}
	}
	if seen != 2 {
		t.Errorf("both players should see the trade window, %d did", seen)
	}
}

func TestTradeRefusedWhenTooFar(t *testing.T) {
	r, a, b := tradeRig(t)
	b.Body.X = a.Body.X + 1000
	a.target = b.ID
	lines(a)
	r.z.TradeOp(a, &protocol.TradeOp{Op: "open"})
	if a.trade != nil || !containsLine(lines(a), "too far away to trade") {
		t.Error("a trade 1000 units away was not refused")
	}
}

func containsLine(ls []string, sub string) bool {
	for _, l := range ls {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}

func TestChangedOfferClearsBothAccepts(t *testing.T) {
	r, a, b := tradeRig(t)
	openTrade(t, r, a, b)
	tradeOp(r, a, "add", 0, 0)
	tradeOp(r, a, "accept", 0, 0)
	tradeOp(r, b, "accept", 0, 0) // would complete: the axe for nothing
	if a.trade != nil {
		t.Fatal("two accepts of one offer did not complete it")
	}
	// Again: Thora accepts, then Bjorn changes his offer.
	r, a, b = tradeRig(t)
	openTrade(t, r, a, b)
	tradeOp(r, a, "add", 0, 0)
	tradeOp(r, a, "accept", 0, 0)
	if !a.trade.side[0].ok {
		t.Fatal("the accept was not kept")
	}
	tradeOp(r, b, "add", 0, 0)
	if a.trade.side[0].ok || a.trade.side[1].ok {
		t.Error("a change to Bjorn's offer left an accept standing")
	}
}

func TestCompletedTradeMovesItemsAndMoney(t *testing.T) {
	r, a, b := tradeRig(t)
	openTrade(t, r, a, b)
	tradeOp(r, a, "add", 0, 0)
	tradeOp(r, b, "money", 0, 60)
	tradeOp(r, a, "accept", 0, 0)
	tradeOp(r, b, "accept", 0, 0)
	if a.trade != nil || b.trade != nil {
		t.Fatal("the trade did not close")
	}
	if a.inv[0] != nil && a.inv[0].ID == "axe" {
		t.Error("Thora still has the axe")
	}
	got := false
	for _, it := range b.inv {
		if it.ID == "axe" {
			got = true
		}
	}
	if !got {
		t.Error("Bjorn did not receive the axe")
	}
	if a.Char.Money != 560 || b.Char.Money != 40 {
		t.Errorf("money after the trade: Thora %d, Bjorn %d, want 560 and 40", a.Char.Money, b.Char.Money)
	}
	// It was written, not only held in memory.
	st, err := r.st.Items(b.Char.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range st {
		if row.ItemID == "axe" {
			found = true
		}
	}
	if !found {
		t.Error("the axe is not in Bjorn's saved items")
	}
}

func TestTradeWithNoRoomMovesNothing(t *testing.T) {
	r, a, b := tradeRig(t)
	for s := 1; s < PackSize; s++ { // Bjorn's pack is full
		r.give(b, s, "pelt", 1)
	}
	openTrade(t, r, a, b)
	tradeOp(r, a, "add", 0, 0)
	tradeOp(r, a, "accept", 0, 0)
	lines(b)
	tradeOp(r, b, "accept", 0, 0)
	if a.inv[0] == nil || a.inv[0].ID != "axe" {
		t.Error("the axe left Thora though the trade could not complete")
	}
	if !containsLine(lines(b), "not enough room") {
		t.Error("Bjorn was not told there was no room")
	}
}

func TestWalkingAwayCancelsATrade(t *testing.T) {
	r, a, b := tradeRig(t)
	openTrade(t, r, a, b)
	tradeOp(r, a, "add", 0, 0)
	b.Body.X = a.Body.X + 2000
	r.ticks(2)
	if a.trade != nil || b.trade != nil {
		t.Error("the trade survived a walk out of range")
	}
	if a.inv[0] == nil || a.inv[0].ID != "axe" {
		t.Error("something moved in a cancelled trade")
	}
}
