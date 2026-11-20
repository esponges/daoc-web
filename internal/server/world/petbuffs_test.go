package world

import (
	"strings"
	"testing"

	"daocweb/internal/server/rules"
)

// petWithBuffSpell is a rig with a summoned pet and a spell for the owner to
// cast on it.
func petWithBuffSpell(t *testing.T, sp *SpellDef) (r *rig, h *Player, pet *NPC, buff *SpellDef) {
	r, h, _, summon := petRig(t)
	summon.Damage = 50
	castAndLand(r, h, summon, h)
	pet = h.pet
	if pet == nil {
		t.Fatal("no pet after the summon")
	}
	sp.Target, sp.Range, sp.CastTime = "Pet", 1500, 1
	buff = buffSpell(r, h, sp)
	return
}

func TestASpellOnThePetNeedsAPet(t *testing.T) {
	r, h, _, _ := petRig(t)
	buff := buffSpell(r, h, &SpellDef{ID: 9310, Name: "Strength of Valhalla", Type: rules.SpellStrConBuff, Target: "Pet", Range: 1500, CastTime: 1, Value: 20, Duration: 1200})
	if !spellSupported(buff) {
		t.Fatal("a buff on the pet is not castable")
	}
	r.z.Cast(h, buff.ID, 0)
	if h.casting != nil || !strings.Contains(lastErr(t, h), "pet") {
		t.Errorf("a pet spell with no pet: casting %v, error %q", h.casting != nil, lastErr(t, h))
	}
}

func TestStrengthConstitutionAndHasteOnThePet(t *testing.T) {
	r, h, pet, _ := petWithBuffSpell(t, &SpellDef{ID: 9311, Name: "Strength of Valhalla", Type: rules.SpellStrConBuff, Value: 18, Duration: 1200})
	str0, con0, swing0 := pet.cb.Stats[rules.STR], pet.cb.Stats[rules.CON], pet.Swing
	castAndLand(r, h, r.w.Cat.Spells[9311], h)
	capv := rules.BuffStatCap(pet.Level)
	if got := pet.cb.Stats[rules.STR] - str0; got != min(18, capv) {
		t.Errorf("a strength buff of 18 raised the pet's strength by %d, want %d", got, min(18, capv))
	}
	if got := pet.cb.Stats[rules.CON] - con0; got != min(18, capv) {
		t.Errorf("and its constitution by %d, want %d", got, min(18, capv))
	}
	haste := buffSpell(r, h, &SpellDef{ID: 9312, Name: "Spirit of Melee", Type: rules.SpellCombatSpeedBuff, Target: "Pet", Range: 1500, CastTime: 1, Value: 10, Duration: 1200})
	swing1 := pet.Swing
	castAndLand(r, h, haste, h)
	if pet.Swing >= swing1 {
		t.Errorf("a combat speed buff left the swing at %v (was %v)", pet.Swing, swing1)
	}
	if swing0 < swing1 {
		t.Errorf("the strength buff slowed the swing: %v -> %v", swing0, swing1)
	}
}

func TestPetDamageAddShieldAndAblative(t *testing.T) {
	r, h, pet, _ := petWithBuffSpell(t, &SpellDef{ID: 9313, Name: "Spirit's Revenge", Type: rules.SpellDamageShield, Damage: 12, DamageType: rules.DamageSpirit, Value: 0, Duration: 300})
	wolf := r.z.npcs[0]
	wolf.Aggro, wolf.Roam = 0, 0
	wolf.setLevel(8)
	wolf.HP = wolf.MaxHP
	pet.HP, pet.MaxHP = 5000, 5000

	// A damage shield hurts whoever hits the pet in melee.
	castAndLand(r, h, r.w.Cat.Spells[9313], h)
	if strongestSpell(pet, effShield) == nil {
		t.Fatal("the shield is not on the pet")
	}
	before := wolf.HP
	r.z.creatureShieldHits(pet, wolf)
	if wolf.HP >= before {
		t.Errorf("the damage shield did nothing: %d -> %d", before, wolf.HP)
	}

	// A damage add rides on the pet's blows.
	add := buffSpell(r, h, &SpellDef{ID: 9314, Name: "Spirit Fury", Type: rules.SpellDamageAdd, Target: "Pet", Range: 1500, CastTime: 1, Damage: 20, DamageType: rules.DamageSpirit, Duration: 1200})
	if r.z.creatureDamageAdd(pet, wolf, 30) != 0 {
		t.Error("a damage add with no spell")
	}
	castAndLand(r, h, add, h)
	if got := r.z.creatureDamageAdd(pet, wolf, 30); got <= 0 {
		t.Errorf("the damage add gave %d on a blow of 30", got)
	}

	// An ablative takes its share of a blow until its pool is gone.
	ward := buffSpell(r, h, &SpellDef{ID: 9315, Name: "Spirit Ward", Type: rules.SpellAblativeArmor, Target: "Pet", Range: 1500, CastTime: 1, Damage: 50, Value: 60, Duration: 1200})
	castAndLand(r, h, ward, h)
	if got := creatureWard(pet, 40); got != 20 {
		t.Errorf("a 50 percent ablative let %d of a 40 blow through, want 20", got)
	}
	if got := creatureWard(pet, 100); got != 60 { // 20 taken so far; half of 100 is 50, but only 40 is left in the pool
		t.Errorf("the pool ran out: %d of a 100 blow through, want 60", got)
	}
	if strongestSpell(pet, effAblative) != nil {
		t.Error("the ablative stayed on after its pool was spent")
	}
}

func TestPetProcSnaresWhatItHits(t *testing.T) {
	r, h, pet, _ := petWithBuffSpell(t, &SpellDef{ID: 9316, Name: "Hinder Prey", Type: rules.SpellOffensiveProc, Value: 9317, Frequency: 100, Duration: 1200})
	r.w.Cat.Spells[9317] = &SpellDef{ID: 9317, Name: "Hinder Prey", Type: rules.SpellUnbreakableSlow, Target: "Enemy", Range: 500, Value: 35, Duration: 10, DamageType: rules.DamageSpirit}
	wolf := r.z.npcs[0]
	wolf.Aggro, wolf.Roam = 0, 0
	wolf.setLevel(8)
	wolf.HP = wolf.MaxHP
	r.z.creatureProcs(pet, wolf)
	if hasKind(wolf, effSlow) {
		t.Fatal("a proc fired before the buff was cast")
	}
	castAndLand(r, h, r.w.Cat.Spells[9316], h)
	if strongestSpell(pet, effProc) == nil {
		t.Fatal("the proc buff is not on the pet")
	}
	for i := 0; i < 20 && !hasKind(wolf, effSlow); i++ {
		r.z.creatureProcs(pet, wolf)
	}
	if !hasKind(wolf, effSlow) {
		t.Error("the pet's proc never slowed what it hit")
	}
}

func TestMesmerizeDurationBuffShortensMezOnThePet(t *testing.T) {
	r, h, pet, _ := petWithBuffSpell(t, &SpellDef{ID: 9318, Name: "Tenacious Warrior Spirit", Type: rules.SpellMezDurationBuff, Value: 50, Duration: 1200})
	wolf := r.z.npcs[0]
	mez := &SpellDef{ID: 9319, Name: "Mesmerize", Type: rules.SpellMesmerize, Target: "Enemy", Range: 1000, Duration: 20}
	r.z.addEffect(pet, wolf, mez, effMez)
	plain := r.z.now
	var ends float64
	for _, e := range pet.effects {
		if e.kind == effMez {
			ends = e.ends - plain
		}
	}
	pet.effects = nil
	castAndLand(r, h, r.w.Cat.Spells[9318], h)
	r.z.addEffect(pet, wolf, mez, effMez)
	var short float64
	for _, e := range pet.effects {
		if e.kind == effMez {
			short = e.ends - r.z.now
		}
	}
	if ends != 20 || short != 10 {
		t.Errorf("a 20 second mesmerize lasted %v on the pet, and %v with a 50 percent mesmerize duration buff", ends, short)
	}
}

// river is flat ground with a river across x 30300 to 30600.
type river struct{ flat }

func (river) IsWet(x, y float64) bool { return x > 30300 && x < 30600 }

func TestAPetGivesUpOnATargetItCannotReach(t *testing.T) {
	r, h, _, sp := petRig(t)
	sp.Damage = 50
	castAndLand(r, h, sp, h)
	pet := h.pet
	r.z.G = river{}
	wolf := r.z.npcs[0] // across the river, at 31000
	wolf.Aggro, wolf.Roam = 0, 0
	h.target = wolf.ID
	r.z.PetCommand(h, "attack")
	if pet.pet.mode != modeAttack {
		t.Fatal("the attack order was not taken")
	}
	r.ticks(int((rules.PetStuckGrace + 12) * TickRate))
	if pet.pet.mode != modeFollow || pet.pet.goal != nil {
		t.Errorf("the pet kept trying: mode %q goal %v", pet.pet.mode, pet.pet.goal != nil)
	}
	if !said(h, "cannot reach") {
		t.Error("the owner was not told the pet cannot reach its target")
	}
	if pet.pet.ignored != wolf {
		t.Error("the pet does not remember the creature it gave up on")
	}
}
