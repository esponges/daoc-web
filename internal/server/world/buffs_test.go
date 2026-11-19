package world

import (
	"math/rand"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// buffSpell registers a spell with the rig's catalog and the Healer's line.
func buffSpell(r *rig, h *Player, sp *SpellDef) *SpellDef {
	r.w.Cat.Spells[sp.ID] = sp
	r.w.Cat.LineSpells["Mending"] = append(r.w.Cat.LineSpells["Mending"], LineSpell{Spec: "Mending", Line: "Mending", Spell: sp.ID, Level: 1})
	r.w.refreshBook(h)
	return sp
}

func TestStatBuffsConstitutionDexterityAcuity(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	friend := caster(r, "Friend", rules.ClassHealer, 10)
	friend.Body.X, friend.Body.Y = 30300, 30000
	con := buffSpell(r, h, &SpellDef{ID: 9001, Name: "Blessing of Health", Type: rules.SpellConstitutionBuff, Target: "Realm", Range: 1000, CastTime: 1, Value: 14, Concentration: 2})
	dex := buffSpell(r, h, &SpellDef{ID: 9002, Name: "Blessed Dexterity", Type: rules.SpellDexterityBuff, Target: "Realm", Range: 1000, CastTime: 1, Value: 16, Concentration: 3})
	acu := buffSpell(r, h, &SpellDef{ID: 9003, Name: "Acuity of the Seer", Type: rules.SpellAcuityBuff, Target: "Realm", Range: 1000, CastTime: 1, Value: 21, Concentration: 4})
	c0, d0, p0, wp0 := w.sh.stats[rules.CON], w.sh.stats[rules.DEX], friend.sh.stats[rules.PIE], w.sh.stats[rules.PIE]
	castAndLand(r, h, con, w)
	castAndLand(r, h, dex, w)
	castAndLand(r, h, acu, friend)
	capv := rules.BuffStatCap(10)
	if got := w.sh.stats[rules.CON] - c0; got != min(14, capv) {
		t.Errorf("a constitution buff of 14 raised constitution by %d, want %d (the level 10 cap is %d)", got, min(14, capv), capv)
	}
	if got := w.sh.stats[rules.DEX] - d0; got != min(16, capv) {
		t.Errorf("a dexterity buff raised dexterity by %d, want %d", got, min(16, capv))
	}
	if got := friend.sh.stats[rules.PIE] - p0; got != min(21, capv) {
		t.Errorf("an acuity buff raised a Healer's piety by %d, want %d", got, min(21, capv))
	}
	castAndLand(r, h, acu, w)
	if w.sh.stats[rules.PIE] != wp0 {
		t.Errorf("an acuity buff raised a Warrior's piety to %d from %d: a class that does not cast has no acuity", w.sh.stats[rules.PIE], wp0)
	}
	if h.concUsed() != con.Concentration+dex.Concentration+acu.Concentration+acu.Concentration {
		t.Errorf("the Healer holds %d, want the four buffs' costs", h.concUsed())
	}
}

func TestHasteShortensSwings(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	haste := buffSpell(r, h, &SpellDef{ID: 9010, Name: "Purity of Attack", Type: rules.SpellCombatSpeedBuff, Target: "Realm", Range: 1000, CastTime: 1, Value: 10, Concentration: 1})
	before := w.swingTime()
	castAndLand(r, h, haste, w)
	after := w.swingTime()
	if want := rules.HastenedSwing(before, 10); after < want-0.001 || after > want+0.001 || after >= before {
		t.Errorf("a swing of %v under 10 percent haste is %v, want %v", before, after, want)
	}
	// The strongest haste counts, not their sum.
	haste2 := buffSpell(r, h, &SpellDef{ID: 9011, Name: "Purity of Offense", Type: rules.SpellCombatSpeedBuff, Target: "Realm", Range: 1000, CastTime: 1, Value: 15, Concentration: 1})
	castAndLand(r, h, haste2, w)
	if got, want := w.swingTime(), rules.HastenedSwing(before, 15); got < want-0.001 || got > want+0.001 {
		t.Errorf("two hastes gave %v, want the stronger alone: %v", got, want)
	}
}

func TestDamageShieldHurtsAttackers(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	shield := buffSpell(r, h, &SpellDef{ID: 9020, Name: "Shards of Bone", Type: rules.SpellDamageShield, Target: "Realm", Range: 1000, CastTime: 1, Damage: 5, Concentration: 2})
	castAndLand(r, h, shield, w)
	n := r.z.npcs[0]
	n.Aggro, n.Roam = 0, 0
	hp := n.HP
	r.z.shieldHits(w, n)
	if want := rules.ShieldDamage(5, r.z.spellResist(n, 0)); hp-n.HP != want {
		t.Errorf("the shield did %d, want %d", hp-n.HP, want)
	}
	// And through a creature's blows (the rig's rolls always miss: use real ones).
	r.z.rng = rand.New(rand.NewSource(7))
	w.hp = float64(w.maxHP())
	n.HP = n.MaxHP
	hits := 0
	for i := 0; i < 60 && n.HP == n.MaxHP; i++ {
		r.z.npcHitsPlayer(n, w)
		hits++
	}
	if n.HP == n.MaxHP {
		t.Errorf("a creature struck %d times and the shield never answered", hits)
	}
}

func TestRegenerationBuffsAddToEachTick(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	endBuff := buffSpell(r, h, &SpellDef{ID: 9030, Name: "Earth Invigoration", Type: rules.SpellEnduranceRegen, Target: "Realm", Range: 1000, CastTime: 1, Value: 5, Concentration: 3})
	powBuff := buffSpell(r, h, &SpellDef{ID: 9031, Name: "Purity of Mind", Type: rules.SpellPowerRegen, Target: "Realm", Range: 1000, CastTime: 1, Value: 3, Concentration: 3})
	gain := func(p *Player) (end, pow float64) {
		p.end, p.power = 10, 10
		p.combatUntil = 0
		p.regenAt, p.powerAt = 0, 0
		e0, w0 := p.end, p.power
		r.ticks(5) // one tick of each: they come once a second or three
		return p.end - e0, p.power - w0
	}
	e1, _ := gain(w)
	castAndLand(r, h, endBuff, w)
	w.hp = float64(w.maxHP())
	e2, _ := gain(w)
	if e2-e1 < 4.9 || e2-e1 > 5.1 {
		t.Errorf("an endurance regeneration buff of 5 added %v to a tick, want 5", e2-e1)
	}
	_, p1 := gain(h)
	castAndLand(r, h, powBuff, h)
	_, p2 := gain(h)
	if p2-p1 < 2.9 || p2-p1 > 3.1 {
		t.Errorf("a power regeneration buff of 3 added %v to a tick, want 3", p2-p1)
	}
}

func TestCureEndsAMesmerize(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	cure := buffSpell(r, h, &SpellDef{ID: 9040, Name: "Spirit Purge", Type: rules.SpellCureMezz, Target: "Realm", Range: 1500, CastTime: 1, Power: -6})
	n := r.z.npcs[0]
	r.z.addEffect(w, n, &SpellDef{ID: 9041, Name: "Mesmerize", Type: rules.SpellMesmerize, Duration: 60}, effMez)
	if !held(w) {
		t.Fatal("the Warrior is not held")
	}
	castAndLand(r, h, cure, w)
	if held(w) {
		t.Error("the cure did not end the mesmerize")
	}
	sentOf[protocol.Err](h, protocol.TErr)
	castAndLand(r, h, cure, w) // nothing left to cure
	if !said(h, "nothing for Spirit Purge to cure") {
		t.Error("a cure with nothing to end was not told so")
	}
}

func TestSpreadHealSharesByNeed(t *testing.T) {
	r, h, w, _, _ := concRig(t)
	spread := buffSpell(r, h, &SpellDef{ID: 9050, Name: "Gift of the Host", Type: rules.SpellSpreadHeal, Target: "Self", Range: 2000, CastTime: 1, Value: 187, Power: 5})
	r.w.omu.Lock()
	for _, o := range []*Player{h, w} {
		r.w.online[o.Char.ID] = o
	}
	r.w.omu.Unlock()
	r.w.command(r.z, h, "invite Warrior")
	r.w.command(r.z, w, "accept")
	h.hp, w.hp = float64(h.maxHP())-100, float64(w.maxHP())-300
	hh, wh := h.hp, w.hp
	castAndLand(r, h, spread, h)
	gotH, gotW := h.hp-hh, w.hp-wh
	if gotH <= 0 || gotW <= gotH {
		t.Errorf("the Warrior missing 300 gained %v and the Healer missing 100 gained %v: the more hurt should gain more", gotW, gotH)
	}
	if ratio := gotW / gotH; ratio < 2.5 || ratio > 3.5 {
		t.Errorf("the Warrior gained %v and the Healer %v: shares should follow what is missing (300 to 100), ratio %v", gotW, gotH, ratio)
	}
	if total := gotH + gotW; total > 187 {
		t.Errorf("the group gained %v, more than the spell's 187", total)
	}
}
