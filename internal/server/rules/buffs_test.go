package rules

import (
	"slices"
	"testing"
)

func TestHealerAndShamanBuffs(t *testing.T) {
	healer := Class{Name: "Healer", Casts: true, Acuity: PIE}
	warrior := Class{Name: "Warrior"}
	vector(t, se, "Acuity on a Healer", func(t *testing.T) {
		got := BuffedStatsFor(healer, SpellAcuityBuff, 21)
		if len(got) != 1 || got[PIE] != 21 {
			t.Errorf("a Healer's acuity buff raises %v, want piety by 21", got)
		}
		if min(got[PIE], BuffStatCap(20)) != 21 || min(got[PIE], BuffStatCap(10)) != 12 {
			t.Error("the buff cap should allow 21 at level 20 and 12 at level 10")
		}
		if BuffedStatsFor(warrior, SpellAcuityBuff, 21) != nil {
			t.Error("a class that does not cast has no acuity to raise")
		}
		if got := BuffedStatsFor(warrior, SpellConstitutionBuff, 14); got[CON] != 14 || len(got) != 1 {
			t.Errorf("a constitution buff raises %v", got)
		}
		if got := BuffedStatsFor(warrior, SpellDexterityBuff, 16); got[DEX] != 16 || len(got) != 1 {
			t.Errorf("a dexterity buff raises %v", got)
		}
		if got := BuffedStatsFor(warrior, SpellStrConBuff, 10); got[STR] != 10 || got[CON] != 10 {
			t.Errorf("the older buffs still work: %v", got)
		}
	})
	vector(t, se, "Haste", func(t *testing.T) {
		if got := HastenedSwing(3.0, 10); got < 2.699 || got > 2.701 {
			t.Errorf("a 3.0 second swing under 10 percent haste is %v, want 2.7", got)
		}
		if HastenedSwing(3.0, 0) != 3.0 {
			t.Error("no haste leaves the swing alone")
		}
		if got := HastenedSwing(1.1, 50); got != 1.0 {
			t.Errorf("the floor is 1.0, got %v", got)
		}
	})
	vector(t, se, "Damage shield", func(t *testing.T) {
		if got := ShieldDamage(5, 0); got != 5 {
			t.Errorf("a 5-point shield on a creature with no resist does %d", got)
		}
		if got, want := ShieldDamage(5, 0.4), int(5*ResistFactor(0.4)); got != want || got >= 5 {
			t.Errorf("a 5-point shield against 40 percent resist does %d, want %d (less than 5)", got, want)
		}
		if got := ShieldDamage(0.4, 0.5); got < 1 {
			t.Errorf("a shield that hits does at least 1, got %d", got)
		}
	})
	vector(t, se, "Power regeneration", func(t *testing.T) {
		plain, buffed := PowerRegen(20), PowerRegenGain(20, 3)
		if d := buffed - plain; d < 2.999 || d > 3.001 {
			t.Errorf("a power regeneration buff of 3 returns %v more, want 3", d)
		}
		if got := EnduranceRegenGain(2, 4); got != 6 {
			t.Errorf("an endurance regeneration buff of 4 on a 2 tick returns %v", got)
		}
	})
	vector(t, se, "Cure of a mesmerize", func(t *testing.T) {
		if !Cures(SpellCureMezz, "mez") || Cures(SpellCureMezz, "stun") || Cures(SpellCurePoison, "mez") {
			t.Error("a mezz cure ends a mesmerize and only that")
		}
		if !Cures(SpellCureDisease, "disease") || !Cures(SpellCurePoison, "poison") || !Cures(SpellCureNearsight, "nearsight") {
			t.Error("each cure ends its own kind")
		}
	})
	vector(t, se, "Spread heal", func(t *testing.T) {
		got := SpreadShares([]int{300, 100}, 187)
		if !slices.Equal(got, []int{141, 46}) {
			t.Errorf("a 187 spread heal over 300 and 100 missing gave %v, want [141 46]", got)
		}
		if got := SpreadShares([]int{5, 5}, 100); !slices.Equal(got, []int{5, 5}) {
			t.Errorf("no one is healed past full: %v", got)
		}
		if got := SpreadShares([]int{0, 0}, 100); !slices.Equal(got, []int{0, 0}) {
			t.Errorf("a group at full health gets nothing: %v", got)
		}
		total := 0
		for _, s := range SpreadShares([]int{90, 40, 30}, 100) {
			total += s
		}
		if total != 100 {
			t.Errorf("the shares add to %d of 100 healed", total)
		}
	})
}
