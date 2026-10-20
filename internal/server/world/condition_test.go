package world

import "testing"

// A bought or dropped item starts whole, or at its template's condition when
// the template gives one; the condition is saved with the row.
func TestNewItemsStartAtTheirCondition(t *testing.T) {
	r := newRig(t)
	r.w.Cat.Items["worn_axe"] = &Item{ID: "worn_axe", Name: "worn axe", Condition: 0.6, Droppable: true}
	p := r.player("Alice", 5, 30000, 30000, false)
	whole, _ := r.w.addItem(p, "hammer", 1)
	worn, _ := r.w.addItem(p, "worn_axe", 1)
	if p.inv[whole].Cond != 10000 || p.inv[worn].Cond != 6000 {
		t.Fatalf("conditions %d and %d, want 10000 and 6000", p.inv[whole].Cond, p.inv[worn].Cond)
	}
	again := newPlayer(p.Acct, p.Char)
	r.w.loadInventory(again)
	if again.inv[whole].Cond != 10000 || again.inv[worn].Cond != 6000 {
		t.Errorf("after a reload %d and %d, want them saved", again.inv[whole].Cond, again.inv[worn].Cond)
	}
}
