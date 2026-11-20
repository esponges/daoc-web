package world

import (
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// taskRig is a town: a kill-task giver (a named guard figure), a trainer and a
// merchant who give delivery tasks, a plain named NPC to deliver to, a wolf, and
// a level 8 player beside the givers.
func taskRig(t *testing.T) (*rig, *Player) {
	t.Helper()
	giver := npcAt(901, "Sir Gudrun", "DOL.GS.GameGuard", 30050, 30000)
	trainer := npcAt(902, "Asmund", "DOL.GS.Trainer.WarriorTrainer", 30000, 30050)
	recv := npcAt(903, "Hella", "DOL.GS.GameNPC", 30080, 30080)
	w := wolf(904, 34000, 30000, 8)
	r := newRig(t, giver, trainer, recv, w)
	r.z.rng = rand.New(zeroSource{})
	r.w.Clock = func() time.Time { return time.Unix(1_700_000_000, 0) }
	n := r.z.byID[904]
	n.Aggro, n.Roam = 0, 0
	p := r.player("Tasker", 8, 30000, 30000, true)
	r.ticks(TickRate)
	sentOf[protocol.Task](p, protocol.TTask)
	return r, p
}

func lastTask(p *Player) (protocol.Task, bool) {
	ms := sentOf[protocol.Task](p, protocol.TTask)
	if len(ms) == 0 {
		return protocol.Task{}, false
	}
	return ms[len(ms)-1], true
}

func TestAskingAGuardFigureForAKillTask(t *testing.T) {
	r, p := taskRig(t)
	r.z.AskTask(p, 901)
	if !said(p, "Search to the") {
		t.Error("the giver gave no direction")
	}
	m := r.z.taskView(p)
	if !m.On || m.Kind != "kill" || m.Max != 8 || !strings.Contains(m.Text, "wolf") || m.Left < 7000 {
		t.Fatalf("task message %+v", m)
	}
	// Asking again while holding it is refused.
	r.z.AskTask(p, 901)
	if !said(p, "already have a task") {
		t.Error("a second ask was not refused")
	}
}

func TestKillingTheTargetAndReturningPays(t *testing.T) {
	r, p := taskRig(t)
	r.z.AskTask(p, 901)
	w := r.z.byID[904]
	w.hurtBy[p.ID] = 10
	xp, money := p.Char.XP, p.Char.Money
	r.z.killNPC(w)
	if !p.task.Killed {
		t.Fatal("the kill did not mark the task")
	}
	if m, _ := lastTask(p); !m.Done || !strings.Contains(m.Text, "Return to Sir Gudrun") {
		t.Errorf("task window after the kill: %+v", m)
	}
	// Away from the giver nothing is paid; back with them it is.
	r.z.Interact(p, 901)
	if p.task.Active() {
		t.Fatal("the task did not finish on return")
	}
	lo, hi := rules.TaskMoney(8, false, 0), rules.TaskMoney(8, false, 0.999)
	if got := p.Char.Money - money; got < int64(lo) || got > int64(hi) {
		t.Errorf("paid %d copper, want %d to %d", got, lo, hi)
	}
	if got := p.Char.XP - xp; got < rules.TaskXP(8) {
		t.Errorf("paid %d experience, want at least %d", got, rules.TaskXP(8))
	}
	if p.task.Done != 1 {
		t.Errorf("tasks done %d, want 1", p.task.Done)
	}
	if m, _ := lastTask(p); m.On || m.Finish != 1 {
		t.Errorf("the window after finishing: %+v", m)
	}
}

func TestTheLimitAndTheLevelCap(t *testing.T) {
	r, p := taskRig(t)
	p.task = nil
	r.z.ensureTask(p)
	p.task.Done = 8
	r.z.AskTask(p, 901)
	if p.task.Active() || !said(p, "cannot do more than 8") {
		t.Error("a ninth task at level 8 was given")
	}
	p.task.Done = 0
	p.Char.Level = 21
	r.z.AskTask(p, 901)
	if p.task.Active() || !said(p, "up to level 20") {
		t.Error("a task was given above level 20")
	}
}

func TestADeliveryTaskCarriesAnItemToAnotherNPC(t *testing.T) {
	r, p := taskRig(t)
	r.z.AskTask(p, 902)
	if !p.task.Active() || p.task.Kind != "delivery" || p.task.Receiver != "Hella" {
		t.Fatalf("task %+v (%v)", p.task, errsOf(p))
	}
	slot := taskItemSlot(p)
	if slot < 0 {
		t.Fatal("no task item in the pack")
	}
	items := sentOf[protocol.Inv](p, protocol.TInv)
	_ = items
	p.sendInv(r.w.Cat, 0)
	named := false
	for _, m := range sentOf[protocol.Inv](p, protocol.TInv) {
		for _, it := range m.Items {
			named = named || it.Slot == slot && strings.HasPrefix(it.Name, "Hella's ")
		}
	}
	if !named {
		t.Error("the task item is not named for its receiver")
	}
	// It cannot be thrown away.
	r.z.DropItem(p, slot)
	if p.inv[slot] == nil {
		t.Error("a task item was destroyed")
	}
	money := p.Char.Money
	r.z.Interact(p, 903)
	if p.task.Active() || taskItemSlot(p) >= 0 {
		t.Fatal("the delivery did not finish")
	}
	if p.Char.Money <= money {
		t.Error("a delivery paid nothing")
	}
}

func TestAbandoningAndExpiryRemoveTheItemAndPayNothing(t *testing.T) {
	r, p := taskRig(t)
	r.z.AskTask(p, 902)
	money := p.Char.Money
	r.z.AbandonTask(p)
	if p.task.Active() || taskItemSlot(p) >= 0 || p.Char.Money != money {
		t.Error("abandoning did not clear the task and its item")
	}
	if m, _ := lastTask(p); m.On {
		t.Error("the window still shows a task")
	}
	r.z.AskTask(p, 902)
	if !p.task.Active() {
		t.Fatal("a new task was refused after abandoning")
	}
	// Two hours and a second later.
	r.w.Clock = func() time.Time { return time.Unix(1_700_000_000+rules.TaskSeconds+1, 0) }
	r.ticks(TickRate * 2)
	if p.task.Active() || taskItemSlot(p) >= 0 || !said(p, "expired") {
		t.Error("an old task did not expire")
	}
	if p.task.Done != 0 {
		t.Error("an expired task counted as done")
	}
}

func TestATaskSurvivesAReconnect(t *testing.T) {
	r, p := taskRig(t)
	r.z.AskTask(p, 901)
	w := r.z.byID[904]
	w.hurtBy[p.ID] = 5
	r.z.killNPC(w)
	p.task = nil
	r.w.loadTask(p) // what logging in does
	if !p.task.Active() || p.task.Kind != "kill" || !p.task.Killed || p.task.Target != "wolf" {
		t.Errorf("after a reload: %+v", p.task)
	}
	if m := r.z.taskView(p); !m.On || !m.Done || m.Left < 7000 {
		t.Errorf("the reloaded window: %+v", m)
	}
}

func TestALevelResetsTasksDone(t *testing.T) {
	r, p := taskRig(t)
	r.z.ensureTask(p)
	p.task.Done = 5
	r.z.levelUp(p)
	if p.task.Done != 0 {
		t.Errorf("tasks done after a level: %d", p.task.Done)
	}
}

// On the imported data, the Norseman start zone has a giver of each kind, and
// kill targets for every level of a new character's first levels.
func TestTheStartZoneHasTaskGiversAndTargets(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	spawns, err := cat.SpawnsIn(101)
	if err != nil {
		t.Fatal(err)
	}
	kill, deliver, named := 0, 0, 0
	targets := map[int]int{}
	for _, s := range spawns {
		n := newNPC(s, cat, rand.New(rand.NewSource(1)))
		if (n.Friendly || n.Guard) && rules.KillGiver(n.Name, n.Spawn.Guild) {
			kill++
		}
		if n.Friendly && rules.DeliveryGiver(n.Kind == KindTrainer, n.Kind == KindMerchant, n.Name) {
			deliver++
		}
		if n.Friendly && !n.Guard && n.Kind != KindNotClassic && n.Name != "" && n.Name != strings.ToLower(n.Name) {
			named++
		}
		if !n.Friendly && !n.Guard && n.Kind == "" && n.Name != "" && n.Name == strings.ToLower(n.Name) {
			targets[n.Level]++
		}
	}
	t.Logf("East Svealand: %d kill givers, %d delivery givers, %d NPCs to deliver to; kill targets by level %v", kill, deliver, named, targets)
	if kill == 0 || deliver == 0 || named < 2 {
		t.Errorf("the start zone lacks givers: %d kill, %d delivery, %d receivers", kill, deliver, named)
	}
	for level := 1; level <= 8; level++ {
		lo, hi := rules.KillTargetLevels(level)
		found := false
		for l := lo; l <= hi; l++ {
			found = found || targets[l] > 0
		}
		if !found {
			t.Errorf("a level %d player has no kill target in the start zone (levels %d to %d)", level, lo, hi)
		}
	}
}
