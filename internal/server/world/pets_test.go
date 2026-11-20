package world

import (
	"encoding/json"
	"math"
	"math/rand"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// petRig is a world with a Healer who knows a summon spell (any caster class will
// do: the spell is the thing), a Warrior and a wolf.
func petRig(t *testing.T) (r *rig, h, w *Player, summon *SpellDef) {
	r, h, w, _, _ = concRig(t)
	r.w.Cat.Templates[801] = &Template{ID: 801, Name: "bone pet", Model: 920, Size: 45, Speed: 200, MeleeType: rules.DamageCrush, BodyType: rules.BodyUndead}
	summon = buffSpell(r, h, &SpellDef{ID: 9300, Name: "Summon Bone Pet", Type: rules.SpellSummonSpirit, Target: "Self", CastTime: 1,
		Damage: 50, Value: 0, LifeDrain: 801})
	r.z.rng = rand.New(rand.NewSource(11))
	return
}

func petsOf(z *Zone, p *Player) (out []*NPC) {
	for _, n := range z.npcs {
		if n.isPet() && n.pet.owner == p {
			out = append(out, n)
		}
	}
	return
}

func TestSummonCreatesAnOwnedPetAtTheSpellsLevel(t *testing.T) {
	r, h, _, sp := petRig(t)
	sp.Damage, sp.Value = -75, 15
	castAndLand(r, h, sp, h)
	pets := petsOf(r.z, h)
	if len(pets) != 1 {
		t.Fatalf("%d pets after a summon", len(pets))
	}
	n := pets[0]
	if want := rules.PetLevel(-75, 15, h.Char.Level); n.Level != want {
		t.Errorf("the pet is level %d, want %d", n.Level, want)
	}
	if n.Name != "bone pet" || n.Model != 920 || h.pet != n {
		t.Errorf("name %q model %d, owner holds %v", n.Name, n.Model, h.pet == n)
	}
	if d := math.Hypot(n.Body.X-h.Body.X, n.Body.Y-h.Body.Y); d > 100 {
		t.Errorf("the pet stands %v from its owner", d)
	}
	e := r.z.npcEntity(n, h)
	if e.Owner != h.Char.Name || !e.Friendly {
		t.Errorf("the entity %+v does not carry the owner and friendliness", e)
	}
	types, raws := sent(h)
	var m protocol.Pet
	found := false
	for i, ty := range types {
		if ty == protocol.TPet {
			found = true
			m = decodePet(raws[i])
		}
	}
	if !found || m.Name != "bone pet" || m.Stance != rules.StanceDefensive || m.Mode != modeFollow || m.HP != 1 {
		t.Errorf("the pet message %+v (sent: %v)", m, found)
	}
	// Summoning again releases the first.
	castAndLand(r, h, sp, h)
	if got := petsOf(r.z, h); len(got) != 1 || got[0] == n {
		t.Errorf("after a second summon: %d pets, the old one kept: %v", len(got), len(got) == 1 && got[0] == n)
	}
	if r.z.byID[n.ID] != nil {
		t.Error("the released pet is still in the zone")
	}
}

func TestSummonWithMissingTemplateFails(t *testing.T) {
	r, h, _, sp := petRig(t)
	sp.LifeDrain = 4242
	castAndLand(r, h, sp, h)
	if len(petsOf(r.z, h)) != 0 || !said(h, "not in this world") {
		t.Error("a summon of an unknown template made a pet or said nothing")
	}
}

func TestPetFollowsItsOwner(t *testing.T) {
	r, h, _, sp := petRig(t)
	castAndLand(r, h, sp, h)
	n := h.pet
	h.Body.X += 2000
	r.ticks(20 * TickRate)
	if d := math.Hypot(n.Body.X-h.Body.X, n.Body.Y-h.Body.Y); d > rules.PetFollowDist+80 {
		t.Errorf("after the owner ran 2000 units the pet is %v behind", d)
	}
	n.pet.mode = modeStay
	h.Body.X += 1000
	before := n.Body.X
	r.ticks(5 * TickRate)
	if n.Body.X != before {
		t.Error("a pet told to stay followed")
	}
}

func TestPetAttacksOnOrderAndTheOwnerGetsTheKill(t *testing.T) {
	r, h, _, sp := petRig(t)
	sp.Damage = 50
	castAndLand(r, h, sp, h)
	n := h.pet
	wolf := r.z.npcs[0]
	wolf.Aggro, wolf.Roam = 0, 0
	wolf.Level = 5
	wolf.setLevel(5)
	wolf.HP = wolf.MaxHP
	h.Char.Level = 10
	xp := h.Char.XP
	h.target = wolf.ID
	r.z.PetCommand(h, "attack")
	if n.pet.mode != modeAttack || n.pet.goal != wolf {
		t.Fatal("the attack order was not taken")
	}
	for i := 0; i < 90 && !wolf.dead(); i++ {
		r.ticks(TickRate)
	}
	if !wolf.dead() {
		t.Fatalf("the pet did not kill the creature in 90 seconds: hp %d of %d", wolf.HP, wolf.MaxHP)
	}
	if wolf.hurtBy[h.ID] == 0 {
		t.Error("the pet's damage was not counted as its owner's")
	}
	if h.Char.XP <= xp {
		t.Errorf("the owner gained no experience for the pet's kill: %d -> %d", xp, h.Char.XP)
	}
	if n.pet.mode != modeFollow {
		t.Errorf("after the kill the pet is on %q, want follow", n.pet.mode)
	}
}

func TestPetStances(t *testing.T) {
	for _, tc := range []struct {
		stance string
		fights bool
	}{{rules.StanceDefensive, true}, {rules.StancePassive, false}} {
		r, h, _, sp := petRig(t)
		sp.Damage = 50
		castAndLand(r, h, sp, h)
		_ = h.pet
		r.z.PetCommand(h, tc.stance)
		wolf := r.z.npcs[0]
		wolf.Aggro, wolf.Roam = 0, 0
		wolf.Body.X, wolf.Body.Y = h.Body.X+300, h.Body.Y
		wolf.engage(h, r.z)
		r.ticks(3 * TickRate)
		if got := wolf.hurtBy[h.ID] > 0 || wolf.dead(); got != tc.fights {
			t.Errorf("a %s pet whose owner is attacked fights: %v, want %v", tc.stance, got, tc.fights)
		}
	}
	// Aggressive: it picks a quiet creature nearby.
	r, h, _, sp := petRig(t)
	castAndLand(r, h, sp, h)
	r.z.PetCommand(h, rules.StanceAggressive)
	wolf := r.z.npcs[0]
	wolf.Aggro, wolf.Roam = 0, 0
	wolf.Body.X, wolf.Body.Y = h.pet.Body.X+300, h.pet.Body.Y
	r.ticks(2 * TickRate)
	if wolf.hurtBy[h.ID] == 0 && !wolf.dead() {
		t.Error("an aggressive pet left a creature in reach alone")
	}
	r.z.PetCommand(h, "bogus")
	if !said(h, "You can tell your pet") {
		t.Error("an unknown command was not answered")
	}
}

func TestPetOrdersAreTheOwnersAlone(t *testing.T) {
	r, h, w, sp := petRig(t)
	castAndLand(r, h, sp, h)
	r.z.PetCommand(w, "release")
	if h.pet == nil || !said(w, "You have no pet") {
		t.Error("a stranger released the pet")
	}
}

func TestPetIsReleasedWithItsOwner(t *testing.T) {
	r, h, w, sp := petRig(t)
	castAndLand(r, h, sp, h)
	n := h.pet
	r.z.killPlayer(h, nil)
	r.ticks(2)
	if r.z.byID[n.ID] != nil || h.pet != nil || len(petsOf(r.z, h)) != 0 {
		t.Error("the pet outlived its owner's death")
	}
	types, raws := sent(h)
	closed := false
	for i, ty := range types {
		if ty == protocol.TPet && decodePet(raws[i]).Closed {
			closed = true
		}
	}
	if !closed {
		t.Error("the owner was not told the pet window closes")
	}
	// A logout.
	castAndLand(r, w, buffSpell(r, h, &SpellDef{ID: 9301, Name: "Summon", Type: rules.SpellSummonSpirit, Target: "Self", Damage: 50, LifeDrain: 801}), w)
	_ = w
}

func TestPetIsReleasedWhenTheOwnerLogsOutOrIsTooFar(t *testing.T) {
	r, h, _, sp := petRig(t)
	castAndLand(r, h, sp, h)
	n := h.pet
	h.closed.Store(true)
	r.ticks(2)
	if r.z.byID[n.ID] != nil {
		t.Error("the pet outlived its owner's logout")
	}
	r, h, _, sp = petRig(t)
	castAndLand(r, h, sp, h)
	n = h.pet
	n.pet.mode = modeStay
	h.Body.X += rules.PetLeash + 500
	r.ticks(int(rules.PetLeashGrace-1) * TickRate)
	if n.pet.released {
		t.Error("the pet was lost before the grace time")
	}
	r.ticks(3 * TickRate)
	if !n.pet.released || !said(h, "lost") {
		t.Error("the pet was not lost beyond the leash")
	}
}

func TestACreatureThatHatesThePetHatesItsOwner(t *testing.T) {
	r, h, _, sp := petRig(t)
	sp.Damage = 5
	castAndLand(r, h, sp, h)
	n := h.pet
	wolf := r.z.npcs[0]
	wolf.Aggro, wolf.Roam = 0, 0
	wolf.Body.X, wolf.Body.Y = n.Body.X+60, n.Body.Y
	wolf.MaxHP, wolf.HP = 100000, 100000
	r.z.npcHitsNPC(n, wolf)
	for i := 0; i < 40 && wolf.hurtBy[h.ID] == 0; i++ {
		r.z.npcHitsNPC(n, wolf)
	}
	if !wolf.hates(n) || !wolf.hates(h) {
		t.Errorf("the creature hates the pet %v and the owner %v after the pet struck it", wolf.hates(n), wolf.hates(h))
	}
	if wolf.tagger != h.ID {
		t.Errorf("the creature was tagged by %d, want the owner %d", wolf.tagger, h.ID)
	}
}

func charmSpell(value float64, kind int) *SpellDef {
	return &SpellDef{ID: 9310, Name: "Call of Gleipnir", Type: rules.SpellCharm, Target: "Enemy", Range: 1500, CastTime: 1, Value: value, AmnesiaChance: kind}
}

func TestCharmTakesAnAnimalAndRefusesTheRest(t *testing.T) {
	r, h, _, _ := petRig(t)
	wolf := r.z.npcs[0]
	wolf.Aggro, wolf.Roam = 0, 0
	r.w.Cat.Templates[802] = &Template{ID: 802, Name: "wolf", BodyType: rules.BodyAnimal}
	wolf.Spawn.Template = 802
	wolf.Level = 8
	cs := buffSpell(r, h, charmSpell(20, 2))
	// Too high: the spell's limit.
	low := *cs
	low.Value = 5
	r.z.castSpell(h, wolf, &low)
	if wolf.isPet() || !said(h, "too strong") {
		t.Error("a creature above the spell's limit was charmed, or no reason was given")
	}
	// Wrong kind.
	r.w.Cat.Templates[802].BodyType = rules.BodyHumanoid
	r.z.castSpell(h, wolf, cs)
	if wolf.isPet() || !said(h, "does not charm that type") {
		t.Error("a humanoid was charmed by an animal charm")
	}
	r.w.Cat.Templates[802].BodyType = rules.BodyAnimal
	wolf.engage(h, r.z)
	r.z.castSpell(h, wolf, cs)
	if !wolf.isPet() || h.pet != wolf || wolf.target != nil || wolf.hates(h) {
		t.Fatal("the wolf was not taken, or still fights its new owner")
	}
	if e := r.z.npcEntity(wolf, h); e.Owner != h.Char.Name {
		t.Error("the charmed wolf does not show its owner")
	}
	// Released: the creature is itself again, and angry with its old owner.
	r.z.PetCommand(h, "release")
	if wolf.isPet() || h.pet != nil || wolf.target != combatant(h) || wolf.Friendly {
		t.Errorf("a released charm: pet %v, target %v, friendly %v", wolf.isPet(), wolf.target, wolf.Friendly)
	}
}

func TestCommanderAndMinions(t *testing.T) {
	r, h, _, _ := petRig(t)
	comm := buffSpell(r, h, &SpellDef{ID: 9320, Name: "Summon Bone Master", Type: rules.SpellSummonCommander, Target: "Self", CastTime: 1, Damage: 50, Pulse: 3, LifeDrain: 801})
	minion := buffSpell(r, h, &SpellDef{ID: 9321, Name: "Summon Bonemage", Type: rules.SpellSummonMinion, Target: "Self", CastTime: 1, Damage: 40, LifeDrain: 801})
	castAndLand(r, h, minion, h)
	if len(petsOf(r.z, h)) != 0 || !said(h, "need a commander") {
		t.Error("a minion was summoned without a commander")
	}
	castAndLand(r, h, comm, h)
	cmd := h.pet
	if cmd == nil || cmd.pet.maxMinions != 3 {
		t.Fatalf("no commander of three: %+v", cmd)
	}
	for i := 0; i < 3; i++ {
		castAndLand(r, h, minion, h)
	}
	if len(cmd.pet.minions) != 3 || h.pet != cmd {
		t.Fatalf("%d minions, owner's pet is still the commander: %v", len(cmd.pet.minions), h.pet == cmd)
	}
	castAndLand(r, h, minion, h)
	if len(cmd.pet.minions) != 3 || !said(h, "cannot control another") {
		t.Error("a fourth minion was accepted")
	}
	// Orders reach the minions, and the window lists them.
	wolf := r.z.npcs[0]
	wolf.Aggro, wolf.Roam = 0, 0
	wolf.Body.X, wolf.Body.Y = cmd.Body.X+400, cmd.Body.Y
	h.target = wolf.ID
	r.z.PetCommand(h, "attack")
	for _, m := range cmd.pet.minions {
		if m.pet.goal != wolf || m.pet.mode != modeAttack {
			t.Error("a minion did not take its commander's order")
		}
	}
	if m := r.z.petMsg(h); len(m.Minions) != 3 {
		t.Errorf("the pet message lists %d minions", len(m.Minions))
	}
	// The commander dies: the minions are released.
	minions := append([]*NPC(nil), cmd.pet.minions...)
	cmd.HP = 1
	r.z.killNPC(cmd)
	r.ticks(2)
	for _, m := range minions {
		if !m.pet.released || r.z.byID[m.ID] != nil {
			t.Error("a minion outlived its commander")
		}
	}
	if len(petsOf(r.z, h)) != 0 {
		t.Error("pets are left after the commander died")
	}
}

func TestPetDiesAndLeavesNoCorpse(t *testing.T) {
	r, h, _, sp := petRig(t)
	castAndLand(r, h, sp, h)
	n := h.pet
	r.z.killNPC(n)
	r.ticks(2)
	if r.z.byID[n.ID] != nil || h.pet != nil || !said(h, "dies") {
		t.Error("a pet that died left a corpse or its owner unaware")
	}
}

func decodePet(raw []byte) protocol.Pet {
	var m protocol.Pet
	if err := json.Unmarshal(raw, &m); err != nil {
		return protocol.Pet{}
	}
	return m
}
