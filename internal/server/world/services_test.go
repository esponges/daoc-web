package world

import (
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

func npcAt(id int64, name, classType string, x, y float64) Spawn {
	return Spawn{ID: id, Name: name, Model: 56, X: x, Y: y, Z: 5000, Level: 50, Size: 50, Speed: 191, Realm: 2, ClassType: classType}
}

func errsOf(p *Player) (out []string) {
	for _, e := range sentOf[protocol.Err](p, protocol.TErr) {
		out = append(out, e.Text)
	}
	return
}

func TestTrainingAtAWarriorTrainer(t *testing.T) {
	r := newRig(t, npcAt(1, "Asmund", "DOL.GS.Trainer.WarriorTrainer", 30050, 30000),
		npcAt(2, "Hella", "DOL.GS.Trainer.ThaneTrainer", 30000, 30050),
		npcAt(3, "Krek", "DOL.GS.Trainer.VikingTrainer", 41000, 30000))
	p := r.player("Pupil", 10, 30000, 30000, true)
	if r.z.byID[1].Kind != KindTrainer || r.z.byID[1].Teaches != "Warrior" {
		t.Fatalf("trainer kind %q teaches %q", r.z.byID[1].Kind, r.z.byID[1].Teaches)
	}
	sentOf[protocol.Err](p, "")
	r.z.Interact(p, 1)
	tr := sentOf[protocol.Trainer](p, protocol.TTrainer)
	if len(tr) != 1 || tr[0].Unspent != p.sh.unspent || len(tr[0].Specs) == 0 {
		t.Fatalf("trainer offer %+v", tr)
	}
	pts := p.sh.unspent
	r.z.Train(p, "Hammer", 2)
	if p.specLevel("Hammer") != 3 || p.sh.unspent != pts-rules.TrainCost(1, 3) {
		t.Errorf("hammer %d, points %d -> %d", p.specLevel("Hammer"), pts, p.sh.unspent)
	}
	saved, _ := r.st.Character(p.Acct.ID, p.Char.ID)
	if saved.Specs["Hammer"] != 3 {
		t.Errorf("saved specs %v", saved.Specs)
	}
	if len(sentOf[protocol.Trainer](p, protocol.TTrainer)) == 0 {
		t.Error("the trainer's offer was not resent")
	}
	// Above level is refused.
	r.z.Train(p, "Hammer", 8)
	if p.specLevel("Hammer") != 3 {
		t.Error("trained above level")
	}
	// Another class's trainer will not teach.
	errsOf(p)
	r.z.Interact(p, 2)
	if errs := errsOf(p); len(errs) == 0 || !strings.Contains(errs[0], "cannot teach you") {
		t.Errorf("thane trainer: %v", errs)
	}
	// Away from a trainer.
	p.Body.X = 38000
	r.z.Train(p, "Sword", 1)
	if p.specLevel("Sword") != 1 {
		t.Error("trained away from a trainer")
	}
	// The base class's trainer teaches a Warrior too.
	p.Body.X = 40950
	r.z.Interact(p, 3)
	r.z.Train(p, "Sword", 1)
	if p.specLevel("Sword") != 2 {
		t.Error("a Viking trainer did not teach a Warrior")
	}
}

func TestDeathCostsExperienceGravestoneAndConstitution(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 12))
	p := r.player("Mortal", 10, 30000, 30000, true)
	n := r.z.byID[1]
	p.Char.XP = 300000
	r.w.recalc(p)
	con := p.sh.stats[rules.CON]
	p.Body.X, p.Body.Y = 30500, 30500
	r.z.killPlayer(p, n)
	if p.Char.XP != 300000-3328 {
		t.Fatalf("after the first death xp %d, want %d", p.Char.XP, 300000-3328)
	}
	r.z.ReleasePlayer(p)
	if p.Char.XP != 300000-2*3328 {
		t.Errorf("after release xp %d, want %d", p.Char.XP, 300000-2*3328)
	}
	g := r.z.graves[p.Char.ID]
	if g == nil || g.xp != 3328 || g.x != 30500 || g.y != 30500 {
		t.Fatalf("grave %+v", g)
	}
	if p.Char.ConLost != 1 || p.sh.stats[rules.CON] != con-1 {
		t.Errorf("constitution lost %d, con %d -> %d", p.Char.ConLost, con, p.sh.stats[rules.CON])
	}
	if p.dead || p.hp != float64(p.maxHP()) {
		t.Errorf("back at the bind stone with %v of %d", p.hp, p.maxHP())
	}
	// A second death and release: the old grave goes, the new holds only the new loss.
	p.Body.X, p.Body.Y = 31000, 31000
	r.z.killPlayer(p, n)
	second := int64(256000 * rules.LossPermille(10, 1) / 1000)
	r.z.ReleasePlayer(p)
	g = r.z.graves[p.Char.ID]
	if len(r.z.graves) != 1 || g.x != 31000 || g.xp != second {
		t.Errorf("second grave %+v, want %d at 31000", g, second)
	}
	if p.Char.ConLost != 3 {
		t.Errorf("constitution lost %d after two deaths, want 3", p.Char.ConLost)
	}
	saved, _ := r.st.Character(p.Acct.ID, p.Char.ID)
	if saved.Grave == nil || saved.Grave.XP != second || saved.ConLost != 3 {
		t.Errorf("saved %+v grave %+v", saved, saved.Grave)
	}
}

func TestLowLevelDeathsCostNothing(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 12))
	p := r.player("Novice", 3, 30000, 30000, true)
	xp := p.Char.XP + 100
	p.Char.XP = xp
	r.z.killPlayer(p, r.z.byID[1])
	r.z.ReleasePlayer(p)
	if p.Char.XP != xp || p.Char.ConLost != 0 || len(r.z.graves) != 0 || p.dead {
		t.Errorf("level 3 death: xp %d (was %d), con lost %d, graves %d", p.Char.XP, xp, p.Char.ConLost, len(r.z.graves))
	}
}

func TestPrayingAtYourGrave(t *testing.T) {
	r := newRig(t, wolf(1, 39000, 39000, 12))
	p := r.player("Pious", 10, 30000, 30000, true)
	other := r.player("Stranger", 10, 30100, 30000, true)
	p.Char.XP = 300000
	r.z.killPlayer(p, r.z.byID[1])
	r.z.ReleasePlayer(p)
	gid := GraveIDBase + p.Char.ID
	// Someone else's grave.
	r.z.SetTarget(other, gid)
	errsOf(other)
	r.z.Pray(other)
	if errs := errsOf(other); len(errs) == 0 || !strings.Contains(errs[0], "not your grave") {
		t.Errorf("stranger praying: %v", errs)
	}
	// Walk to the grave (the bind point is elsewhere), target it, pray.
	g := r.z.graves[p.Char.ID]
	p.Body.X, p.Body.Y = g.x+100, g.y
	r.z.SetTarget(p, gid)
	if p.target != gid {
		t.Fatal("could not target the grave")
	}
	// Moving cancels.
	r.z.Pray(p)
	r.ticks(TickRate)
	p.Body.X += 50
	r.ticks(2)
	if p.prayUntil != 0 || r.z.graves[p.Char.ID] == nil {
		t.Fatal("moving did not cancel the prayer")
	}
	xp := p.Char.XP
	r.z.Pray(p)
	r.ticks(int(rules.PraySeconds*TickRate) + 2)
	if p.Char.XP != xp+3328 || r.z.graves[p.Char.ID] != nil || p.Char.Grave != nil {
		t.Errorf("after praying xp %d (was %d), grave %v", p.Char.XP, xp, r.z.graves[p.Char.ID])
	}
}

func TestGravesSurviveARestart(t *testing.T) {
	r := newRig(t, wolf(1, 39000, 39000, 12))
	p := r.player("Haunted", 10, 30000, 30000, true)
	p.Char.XP = 300000
	r.z.killPlayer(p, r.z.byID[1])
	r.z.ReleasePlayer(p)
	z2 := NewZone(r.w, 100, "Test", flat{}, nil, nil)
	g := z2.graves[p.Char.ID]
	if g == nil || g.xp != 3328 || g.name != "Haunted" {
		t.Errorf("reloaded zone grave %+v", g)
	}
	e := graveEntity(g, p)
	if e.K != "g" || !e.Own || e.Name != "grave of Haunted" {
		t.Errorf("grave entity %+v", e)
	}
}

func TestHealersRestoreConstitution(t *testing.T) {
	r := newRig(t, npcAt(1, "Hilda", "DOL.GS.GameHealer", 30050, 30000))
	p := r.player("Patient", 12, 30000, 30000, true)
	if r.z.byID[1].Kind != KindHealer {
		t.Fatalf("healer kind %q", r.z.byID[1].Kind)
	}
	// Nothing to restore.
	sentOf[protocol.ChatMsg](p, protocol.TChatMsg)
	r.z.Interact(p, 1)
	if msgs := sentOf[protocol.ChatMsg](p, protocol.TChatMsg); len(msgs) == 0 || !strings.Contains(msgs[0].Text, "no lost constitution") {
		t.Errorf("nothing to restore: %v", msgs)
	}
	p.Char.ConLost, p.Char.Money = 3, 500
	r.w.recalc(p)
	r.z.Interact(p, 1)
	if h := sentOf[protocol.Healer](p, protocol.THealer); len(h) != 1 || h[0].Total != 1056 || h[0].Lost != 3 {
		t.Fatalf("healer offer %+v", h)
	}
	r.z.Restore(p, 1)
	if p.Char.ConLost != 2 || p.Char.Money != 148 {
		t.Errorf("with 500 copper: lost %d, money %d; want 2 and 148", p.Char.ConLost, p.Char.Money)
	}
	p.Char.ConLost, p.Char.Money = 3, 2000
	r.z.Restore(p, 1)
	if p.Char.ConLost != 0 || p.Char.Money != 2000-1056 {
		t.Errorf("with 2000 copper: lost %d, money %d", p.Char.ConLost, p.Char.Money)
	}
	saved, _ := r.st.Character(p.Acct.ID, p.Char.ID)
	if saved.ConLost != 0 || saved.Money != 944 {
		t.Errorf("saved lost %d money %d", saved.ConLost, saved.Money)
	}
}
