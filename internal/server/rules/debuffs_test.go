package rules

import "testing"

func TestDebuffsAndWards(t *testing.T) {
	vector(t, se, "Strength debuff on a creature", func(t *testing.T) {
		got := DebuffedStats(SpellStrengthDebuff, 20)
		if len(got) != 1 || got[STR] != 20 {
			t.Errorf("a 20-point strength debuff lowers %v", got)
		}
		if got := DebuffedStats(SpellStrConDebuff, 16); got[STR] != 16 || got[CON] != 16 {
			t.Errorf("strength/constitution debuff lowers %v", got)
		}
		if got := DebuffedStats(SpellDexQuiDebuff, 21); got[DEX] != 21 || got[QUI] != 21 {
			t.Errorf("dexterity/quickness debuff lowers %v", got)
		}
		if DebuffedStats(SpellDisease, 0) != nil {
			t.Error("a disease is no stat debuff")
		}
		if d := DebuffDuration(60, 0, 0); d != 60 {
			t.Errorf("duration %v", d)
		}
		if d := DebuffDuration(60, 0, 50); d != 30 {
			t.Errorf("a 50 percent resist halves the debuff: %v", d)
		}
		if d := DebuffDuration(60, 500, 0); d != 240 {
			t.Errorf("four times the spell's at most: %v", d)
		}
		if d := DebuffDuration(2, 0, 100); d != 1 {
			t.Errorf("a second at least: %v", d)
		}
		if got := SlowedSwing(3, 25); got < 3.99 || got > 4.01 {
			t.Errorf("a 3.0 swing slowed 25 percent is %v, want 4.0", got)
		}
		if StrongestOf(10, 30, 20) != 30 || StrongestOf() != 0 {
			t.Error("the strongest counts")
		}
	})
	vector(t, se, "Body resist debuff", func(t *testing.T) {
		ty, ok := ResistDebuffType(SpellBodyResistDebuff)
		if !ok || ty != DamageBody {
			t.Fatalf("body resist debuff names %d", ty)
		}
		if _, ok := ResistDebuffType(SpellStrengthDebuff); ok {
			t.Error("a stat debuff is no resist debuff")
		}
		// 10 percent resist under a 15 percent debuff meets -5 percent.
		if f := ResistFactor(10 - 15); f < 1.0499 || f > 1.0501 {
			t.Errorf("-5%% resist leaves %v of the damage, want 1.05", f)
		}
	})
	vector(t, se, "Disease", func(t *testing.T) {
		if DiseasedHeal(200) != 100 {
			t.Errorf("a diseased target is healed %d of 200", DiseasedHeal(200))
		}
		if DiseasedStrength(80) != 74 {
			t.Errorf("80 strength under disease is %d, want 74", DiseasedStrength(80))
		}
		if got := NearsightRange(1500, 40); got != 900 {
			t.Errorf("nearsight of 40 percent on 1500 range: %v", got)
		}
	})
	vector(t, se, "Amnesia", func(t *testing.T) {
		if !ConfusionTakesHold(-30, 0.99) || !ConfusionTakesHold(100, 0.5) || ConfusionTakesHold(40, 0.5) || !ConfusionTakesHold(40, 0.3) {
			t.Error("confusion takes hold by its percent; a negative value always does")
		}
		if !ConfusionAttacksAlly(-30, 0.1) || ConfusionAttacksAlly(-30, 0.5) || ConfusionAttacksAlly(30, 0.0) {
			t.Error("only a negative confusion turns a creature on its own side, by that percent")
		}
		if AmnesiaChance != 100 {
			t.Error("the data gives amnesia no chance, so it always lands")
		}
	})
	vector(t, se, "Bladeturn", func(t *testing.T) {
		// A bladeturn takes one hit whole: the rule is in the world; here the
		// ablative's part of the same requirement is checked at the edges.
		if got := AblativePercent(0); got != 25 {
			t.Errorf("an ablative with no damage column absorbs %v percent", got)
		}
		if got := AblativePercent(300); got != 100 {
			t.Errorf("at most 100 percent: %v", got)
		}
	})
	vector(t, se, "Ablative pool", func(t *testing.T) {
		abs, left := AblativeAbsorb(120, 25, 100)
		if abs != 30 || left != 70 {
			t.Errorf("a 25%% 100-point ablative on a 120 hit absorbed %d, left %d, want 30 and 70", abs, left)
		}
		abs, left = AblativeAbsorb(1000, 100, 100)
		if abs != 100 || left != 0 {
			t.Errorf("the pool is all there is: absorbed %d, left %d", abs, left)
		}
		if FatigueFactor(-30) != 1.3 || MezDuration(10, 50) != 5 || MezDuration(10, 99) != 1 || MezDuration(10, 0) != 10 {
			t.Error("fatigue and mesmerize duration factors")
		}
	})
	vector(t, se, "Shorter mesmerize", func(t *testing.T) {
		if MezDuration(10, 50) != 5 || MezDuration(10, 0) != 10 || MezDuration(10, 99) != 1 || MezDuration(10, 100) != 1 || MezDuration(10, -20) != 10 {
			t.Errorf("a 10 second mesmerize under 50, 0, 99, 100 and -20 percent lasts %v %v %v %v %v, want 5 10 1 1 10",
				MezDuration(10, 50), MezDuration(10, 0), MezDuration(10, 99), MezDuration(10, 100), MezDuration(10, -20))
		}
	})
	vector(t, se, "Life transfer", func(t *testing.T) {
		heal, cost := LifeTransfer(1000, 1000, 10, false)
		if heal != 200 || cost != 100 {
			t.Errorf("a 10 percent transfer from 1000 healed %d and cost %d, want 200 and 100", heal, cost)
		}
		heal, cost = LifeTransfer(1000, 1000, 10, true)
		if heal != 100 || cost != 100 {
			t.Errorf("a diseased target gets half: %d, cost %d", heal, cost)
		}
		heal, cost = LifeTransfer(1000, 41, 10, false)
		if cost != 40 || heal != 80 {
			t.Errorf("the caster is left with 1 hit point: healed %d cost %d", heal, cost)
		}
	})
}
