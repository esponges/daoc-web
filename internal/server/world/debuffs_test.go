package world

import (
	"testing"

	"daocweb/internal/server/rules"
)

func dbSpell(id int, name, typ string, value float64, dur int) *SpellDef {
	return &SpellDef{ID: id, Name: name, Type: typ, Target: "Enemy", Range: 1500, Value: value, Duration: dur}
}

func TestStatDebuffLowersACreaturesStrength(t *testing.T) {
	r, h, _, _, _ := concRig(t)
	n := r.z.npcs[0]
	before := n.cb.Stats[rules.STR]
	r.z.castSpell(h, n, dbSpell(9100, "Weaken", rules.SpellStrengthDebuff, 20, 60))
	if got := n.cb.Stats[rules.STR]; got != before-20 {
		t.Errorf("strength %d under a 20-point debuff, want %d", got, before-20)
	}
	r.z.castSpell(h, n, dbSpell(9101, "Weaken II", rules.SpellStrengthDebuff, 10, 60))
	if got := n.cb.Stats[rules.STR]; got != before-20 {
		t.Errorf("a weaker debuff deepened it to %d: the strongest counts", before-got)
	}
	r.z.castSpell(h, n, dbSpell(9102, "Clumsy", rules.SpellStrConDebuff, 8, 60))
	if n.cb.Stats[rules.CON] != n.baseCB.Stats[rules.CON]-8 {
		t.Errorf("constitution %d, want %d", n.cb.Stats[rules.CON], n.baseCB.Stats[rules.CON]-8)
	}
	r.ticks(70 * TickRate)
	if n.cb.Stats[rules.STR] != before {
		t.Errorf("strength %d after the debuffs ran out, want %d", n.cb.Stats[rules.STR], before)
	}
}

func TestResistDebuffAndDamageWithDebuff(t *testing.T) {
	r, h, _, _, _ := concRig(t)
	n := r.z.npcs[0]
	n.resists = map[int]float64{rules.DamageBody: 10}
	r.z.castSpell(h, n, dbSpell(9110, "Rot", rules.SpellBodyResistDebuff, 15, 30))
	if got := r.z.spellResist(n, rules.DamageBody); got != -5 {
		t.Errorf("body resist %v under a 15 percent debuff on 10, want -5", got)
	}
	n.effects = nil
	n.resists = map[int]float64{}
	hp := n.HP
	dd := &SpellDef{ID: 9111, Name: "Bonespike", Type: rules.SpellDDWithDebuff, Target: "Enemy", Range: 1500,
		Damage: 50, DamageType: rules.DamageBody, Value: 20, Duration: 30}
	r.z.castSpell(h, n, dd)
	if n.HP >= hp {
		t.Error("the damage-with-debuff did no damage")
	}
	if got := r.z.spellResist(n, rules.DamageBody); got != -20 {
		t.Errorf("the creature meets body resist %v after the spell, want -20", got)
	}
	ds := &SpellDef{ID: 9112, Name: "Chill", Type: rules.SpellDamageSpeedDecr, Target: "Enemy", Range: 1500, Damage: 20, DamageType: rules.DamageCold, Value: 40, Duration: 30}
	hp = n.HP
	r.z.castSpell(h, n, ds)
	if n.HP >= hp || speedFactor(n) > 0.61 {
		t.Errorf("damage and slow: hp %d -> %d, speed %v", hp, n.HP, speedFactor(n))
	}
}

func TestCombatSpeedDebuffSlowsSwings(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	n := r.z.npcs[0]
	base := n.Swing
	r.z.castSpell(h, n, dbSpell(9120, "Lethargy", rules.SpellCombatSpeedDebuff, 25, 30))
	if want := rules.SlowedSwing(base, 25); n.Swing < want-0.001 || n.Swing > want+0.001 || n.Swing <= base {
		t.Errorf("a creature's swing %v under 25 percent, want %v", n.Swing, want)
	}
	pb := w.swingTime()
	r.z.castSpell(n, w, dbSpell(9121, "Lethargy", rules.SpellCombatSpeedDebuff, 25, 30))
	if w.swingTime() <= pb {
		t.Errorf("a player's swing %v did not lengthen from %v", w.swingTime(), pb)
	}
}

func TestDiseaseWeakensSlowsAndHalvesHealing(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	n := r.z.npcs[0]
	str := w.sh.stats[rules.STR]
	r.z.castSpell(n, w, &SpellDef{ID: 9130, Name: "Plague", Type: rules.SpellDisease, Target: "Enemy", Range: 1500, Duration: 60})
	if got, want := w.sh.stats[rules.STR], rules.DiseasedStrength(str); got != want {
		t.Errorf("strength %d under disease, want %d", got, want)
	}
	if f := speedFactor(w); f < 0.849 || f > 0.851 {
		t.Errorf("speed %v under disease, want 0.85", f)
	}
	w.hp = 100
	r.z.heal(w, 200)
	if w.hp != 200 {
		t.Errorf("a diseased player healed 200 now has %v, want 200 (100 + half)", w.hp)
	}
	cure := buffSpell(r, h, &SpellDef{ID: 9131, Name: "Purge", Type: rules.SpellCureDisease, Target: "Realm", Range: 1500, CastTime: 1})
	castAndLand(r, h, cure, w)
	if hasKind(w, effDisease) || w.sh.stats[rules.STR] != str {
		t.Errorf("the cure left the disease or the weakness: strength %d of %d", w.sh.stats[rules.STR], str)
	}
}

func TestNearsightShortensRangeAndLeavesAnImmunity(t *testing.T) {
	r, h, _, _, _ := concRig(t)
	n := r.z.npcs[0]
	bolt := &SpellDef{Range: 1000}
	r.z.castSpell(h, n, dbSpell(9140, "Blind", rules.SpellNearsight, 40, 5))
	if got := spellRange(n, bolt); got != 600 {
		t.Errorf("a creature's range %v under a 40 percent nearsight, want 600", got)
	}
	r.z.castSpell(h, n, dbSpell(9141, "Blind II", rules.SpellNearsight, 60, 5))
	if got := spellRange(n, bolt); got != 600 {
		t.Errorf("a second nearsight took hold on the first: %v", got)
	}
	r.ticks(6 * TickRate)
	if spellRange(n, bolt) != 1000 {
		t.Error("nearsight did not end")
	}
	r.z.castSpell(h, n, dbSpell(9142, "Blind III", rules.SpellNearsight, 60, 5))
	if spellRange(n, bolt) != 1000 {
		t.Error("a nearsight took hold in the moment of immunity after one ended")
	}
	r.ticks(int(rules.NearsightImmunity+1) * TickRate)
	r.z.castSpell(h, n, dbSpell(9143, "Blind IV", rules.SpellNearsight, 60, 5))
	if spellRange(n, bolt) != 400 {
		t.Errorf("after the immunity a nearsight takes hold again: range %v", spellRange(n, bolt))
	}
}

func TestConfusionAndAmnesiaOnACreature(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	n := r.z.npcs[0]
	n.engage(w, r.z)
	if n.target != combatant(w) {
		t.Fatal("the creature does not fight the Warrior")
	}
	h.Body.X, h.Body.Y = n.Body.X+100, n.Body.Y
	r.z.castSpell(h, n, &SpellDef{ID: 9150, Name: "Confuse", Type: rules.SpellConfusion, Target: "Enemy", Range: 1500, Value: 100, Duration: 60})
	if !hasKind(n, effConfusion) {
		t.Fatal("the confusion did not take hold")
	}
	n.target = nil
	r.ticks(int(rules.ConfusionPulse+1) * TickRate)
	if n.target == nil {
		t.Error("a confused creature picked no new enemy at its pulse")
	}
	r.z.castSpell(h, n, &SpellDef{ID: 9151, Name: "Forget", Type: rules.SpellAmnesia, Target: "Enemy", Range: 1500})
	if n.target != nil || len(n.hate) != 0 {
		t.Errorf("amnesia left the creature fighting %v with hate %v", n.target, n.hate)
	}
	// Against a player a confusion only interrupts a cast.
	r.z.castSpell(n, h, &SpellDef{ID: 9152, Name: "Confuse", Type: rules.SpellConfusion, Target: "Enemy", Range: 1500, Value: 100, Duration: 60})
	if hasKind(h, effConfusion) {
		t.Error("a player was left confused")
	}
}

func TestBladeturnAndAblativeStopBlows(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	n := r.z.npcs[0]
	blade := buffSpell(r, h, &SpellDef{ID: 9160, Name: "Turn Blade", Type: rules.SpellBladeturn, Target: "Realm", Range: 1000, CastTime: 1, Duration: 60})
	castAndLand(r, h, blade, w)
	if !hasKind(w, effBlade) {
		t.Fatal("no bladeturn on the Warrior")
	}
	if got, turned := r.z.wardBlow(w, n, 50); got != 0 || !turned || hasKind(w, effBlade) {
		t.Errorf("a bladeturn let %d through (turned %v) or stayed on", got, turned)
	}
	abl := buffSpell(r, h, &SpellDef{ID: 9161, Name: "Bone Armor", Type: rules.SpellAblativeArmor, Target: "Realm", Range: 1000, CastTime: 1, Duration: 60, Damage: 25, Value: 100})
	castAndLand(r, h, abl, w)
	if got, turned := r.z.wardBlow(w, n, 120); got != 90 || turned {
		t.Errorf("a 25%% ablative let %d of 120 through, want 90", got)
	}
	var e *effect
	for _, x := range w.effects {
		if x.kind == effAblative {
			e = x
		}
	}
	if e == nil || e.pool != 70 {
		t.Fatalf("the pool is not 70: %+v", e)
	}
	r.z.wardBlow(w, n, 1000)
	r.z.wardBlow(w, n, 1000)
	if hasKind(w, effAblative) {
		t.Error("a spent ablative stayed on")
	}
}

func TestArmourAbsorptionAndFatigue(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	n := r.z.npcs[0]
	if w.combatant(r.w.Cat).BuffAbsorb != 0 {
		t.Fatal("absorption before any buff")
	}
	abs := buffSpell(r, h, &SpellDef{ID: 9170, Name: "Absorb", Type: rules.SpellArmorAbsorbBuff, Target: "Realm", Range: 1000, CastTime: 1, Duration: 60, Value: 10})
	castAndLand(r, h, abs, w)
	if got := w.combatant(r.w.Cat).BuffAbsorb; got != 0.10 {
		t.Errorf("buffed absorption %v, want 0.10", got)
	}
	st := &StyleDef{Endurance: 10}
	before := w.styleCost(st)
	r.z.castSpell(n, w, &SpellDef{ID: 9171, Name: "Tire", Type: rules.SpellFatigueDebuff, Target: "Enemy", Range: 1500, Value: -30, Duration: 30})
	if after := w.styleCost(st); after <= before {
		t.Errorf("a style cost %d under fatigue, %d without", after, before)
	}
}

func TestLifeTransferMovesHealth(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	lt := buffSpell(r, h, &SpellDef{ID: 9180, Name: "Transfer", Type: rules.SpellLifeTransfer, Target: "Realm", Range: 2000, CastTime: 1, Value: 10})
	h.hp = float64(h.maxHP())
	w.hp = 1
	hh, wh := h.hp, w.hp
	wantHeal, wantCost := rules.LifeTransfer(h.maxHP(), int(h.hp), 10, false)
	r.z.castSpell(h, w, lt)
	if gotCost := hh - h.hp; int(gotCost) != wantCost || int(w.hp-wh) != min(wantHeal, w.maxHP()-1) {
		t.Errorf("caster lost %v want %d; target gained %v want %d", gotCost, wantCost, w.hp-wh, wantHeal)
	}
	if wantHeal-wantCost*2 < 0 || wantHeal-wantCost*2 > 1 {
		t.Errorf("heal %d is not twice the cost %d", wantHeal, wantCost)
	}
}

func TestCreaturesBuffThemselvesBeforeFighting(t *testing.T) {
	buff := &SpellDef{ID: 9190, Name: "Might of the Pack", Type: rules.SpellStrConBuff, Target: "Self", Range: 0, Value: 12, Duration: 600}
	r, n, p := casterRig(t, buff, bolt)
	n.engage(p, r.z)
	str := n.cb.Stats[rules.STR]
	r.ticks(2 * TickRate)
	if !hasEffect(n, buff.ID) {
		t.Fatal("the creature did not cast its buff")
	}
	if got := n.cb.Stats[rules.STR]; got != str+12 {
		t.Errorf("strength %d with the buff, want %d", got, str+12)
	}
	armour := &SpellDef{ID: 9191, Name: "Bark Skin", Type: rules.SpellArmourBuff, Target: "Self", Value: 15, Duration: 600}
	r.z.castSpell(n, n, armour)
	if n.cb.BuffAF != 15 {
		t.Errorf("armour factor buff %v, want 15", n.cb.BuffAF)
	}
}
