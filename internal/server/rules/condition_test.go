package rules

import "testing"

const cnd = "rules-item-condition"

func TestItemCondition(t *testing.T) {
	vector(t, cnd, "A new item", func(t *testing.T) {
		if ConditionStart(0) != FullCondition || ConditionStart(1) != FullCondition || ConditionStart(0.6) != 6000 {
			t.Errorf("starts %d %d %d, want whole, whole and 6000", ConditionStart(0), ConditionStart(1), ConditionStart(0.6))
		}
	})
	vector(t, cnd, "Two hundred blows", func(t *testing.T) {
		c := FullCondition
		for i := 0; i < 200; i++ {
			c = Worn(c)
		}
		if c != 9000 || Worn(3) != 0 {
			t.Errorf("after 200 blows %d, want 9000; Worn(3) = %d, want 0", c, Worn(3))
		}
	})
	vector(t, cnd, "A block", func(t *testing.T) {
		a, s := BlowWears(Blocked)
		ha, hs := BlowWears(Hit)
		ma, ms := BlowWears(Missed)
		if a || !s || !ha || hs || ma || ms {
			t.Errorf("a block wears armour %v shield %v; a hit %v %v; a miss %v %v", a, s, ha, hs, ma, ms)
		}
	})
	vector(t, cnd, "A weapon at 90%", func(t *testing.T) {
		if got := 16.5 * Effectiveness(9000); got < 14.849 || got > 14.851 {
			t.Errorf("16.5 DPS at 90%% is %v, want 14.85", got)
		}
	})
	vector(t, cnd, "The 90% warning", func(t *testing.T) {
		if WarnCrossed(9002, 8997) != 9000 || WarnCrossed(8997, 8992) != 0 || WarnCrossed(8003, 7998) != 8000 || WarnCrossed(10000, 9995) != 0 {
			t.Error("the warning should come once as each of 90, 80 and 70 is crossed")
		}
	})
	vector(t, cnd, "Repair price", func(t *testing.T) {
		if got := RepairPrice(200, 8000); got != 20 {
			t.Errorf("repair price %d, want 20 copper", got)
		}
		if RepairPrice(200, FullCondition) != 0 || RepairPrice(1, 9999) != 1 {
			t.Error("a whole item costs nothing and any wear costs at least a copper")
		}
	})
	vector(t, cnd, "Not enough money", func(t *testing.T) {
		// Refusing for lack of money is the world's; the rule is the price it compares.
		if RepairPrice(5000, 0) != 2500 {
			t.Errorf("a ruined 50 silver item costs %d, want 2500", RepairPrice(5000, 0))
		}
	})
}
