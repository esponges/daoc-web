package world

import (
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// duelRig is two armed players of one realm, face to face.
func duelRig(t *testing.T) (*rig, *Player, *Player) {
	r := newRig(t, wolf(1, 36000, 30000, 8))
	a := r.player("Thora", 20, 30000, 30000, true)
	b := r.player("Bjorn", 20, 30060, 30000, true)
	a.Body.Yaw = angleTo(a.Body.X, a.Body.Y, b.Body.X, b.Body.Y)
	b.Body.Yaw = angleTo(b.Body.X, b.Body.Y, a.Body.X, a.Body.Y)
	return r, a, b
}

func startDuel(t *testing.T, r *rig, a, b *Player) {
	t.Helper()
	a.target = b.ID
	r.z.Duel(a)
	if !r.z.answerDuel(b, true) || a.duel == nil || a.duel != b.duel {
		t.Fatal("the duel did not begin")
	}
}

func TestDuelAcceptedAndDeclined(t *testing.T) {
	r, a, b := duelRig(t)
	a.target = b.ID
	r.z.Duel(a)
	r.z.answerDuel(b, false)
	if a.duel != nil || !containsLine(lines(a), "declines your challenge") {
		t.Error("a declined challenge started a duel or was not reported")
	}
	startDuel(t, r, a, b)
	if !containsLine(lines(b), "has begun") {
		t.Error("the duel did not announce itself")
	}
}

func TestChallengeLapsesWithSilence(t *testing.T) {
	r, a, b := duelRig(t)
	a.target = b.ID
	r.z.Duel(a)
	r.z.now += DuelLapse + 1
	lines(b)
	r.z.answerDuel(b, true)
	if a.duel != nil || !containsLine(lines(b), "lapsed") {
		t.Error("a challenge answered after it lapsed began a duel")
	}
}

func TestDuelistsHarmEachOtherAndNobodyDies(t *testing.T) {
	r, a, b := duelRig(t)
	startDuel(t, r, a, b)
	lines(a)
	r.z.SetAttack(a, true)
	if !a.attacking {
		t.Fatal("a duellist could not attack their opponent")
	}
	hp0 := b.hp
	for i := 0; i < 30; i++ { // drain as we go: a second of snapshots is most of a queue
		r.ticks(TickRate)
		lines(a)
		lines(b)
	}
	if b.hp >= hp0 {
		t.Errorf("Bjorn took no damage in 30 seconds: %v of %v", b.hp, hp0)
	}
	// Bjorn is at death's door: the next blows leave him at 1 and end the duel.
	b.hp = 3
	var told bool
	for i := 0; i < 60 && b.duel != nil; i++ {
		r.ticks(TickRate)
		lines(a)
		if containsLine(lines(b), "Thora has defeated Bjorn in a duel!") {
			told = true
		}
	}
	if b.dead || b.hp < 1 {
		t.Fatalf("the loser died or was left below 1: dead %v hp %v", b.dead, b.hp)
	}
	if a.duel != nil || b.duel != nil {
		t.Error("the duel did not end with the defeat")
	}
	if !told {
		t.Error("the zone was not told of the defeat")
	}
	if b.Char.XP != xpAt(20) || a.Char.XP != xpAt(20) {
		t.Error("a duel gave or took experience")
	}
}

func xpAt(level int) int64 { return rules.XPToReach(level) }

func TestStrangersCannotFightWithoutADuel(t *testing.T) {
	r, a, b := duelRig(t)
	a.target = b.ID
	r.z.SetAttack(a, true)
	if a.attacking {
		t.Error("a player attacked another with no duel")
	}
	spell := &SpellDef{ID: 9200, Name: "Bolt", Type: "DirectDamage", Target: "Enemy", Damage: 50, Range: 1500}
	if r.z.castTarget(a, spell, b.ID) != nil && r.z.mayHarmPlayer(a, b) {
		t.Error("a hostile spell could be aimed at a player outside a duel")
	}
	hp := b.hp
	r.z.hurt(b, a, 40)
	if b.hp != hp {
		t.Error("a player's damage reached another outside a duel")
	}
}

func TestACreatureJoiningEndsTheDuel(t *testing.T) {
	r, a, b := duelRig(t)
	startDuel(t, r, a, b)
	n := r.z.npcs[0]
	r.z.hurt(a, n, 5)
	if a.duel != nil || b.duel != nil {
		t.Error("a creature's blow left the duel standing")
	}
}

func TestDuelAbandonedWhenApart(t *testing.T) {
	r, a, b := duelRig(t)
	startDuel(t, r, a, b)
	b.Body.X = a.Body.X + DuelAbandon + 100
	r.ticks(2)
	if a.duel != nil || b.duel != nil {
		t.Error("the duel survived the two being 2600 units apart")
	}
}

func TestDuelNeedsSameRealmAndNoCombat(t *testing.T) {
	r, a, b := duelRig(t)
	b.Arch.Realm = a.Arch.Realm + 1
	a.target = b.ID
	r.z.Duel(a)
	if b.duelOffer != nil {
		t.Error("a duel was offered across realms")
	}
	b.Arch.Realm = a.Arch.Realm
	a.combatUntil = r.z.now + 10
	r.z.Duel(a)
	if b.duelOffer != nil {
		t.Error("a duel was offered to someone in combat")
	}
	_ = protocol.TErr
}
