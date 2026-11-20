package world

import (
	"fmt"
	"math"
	"strings"
	"time"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// Tasks, from the rules-tasks capability: a player asks a town NPC for a kill
// task or a delivery task, holds one at a time under a two hour limit, and is
// paid on finishing it. The state is saved with the character in the task table.

const taskItemID = "task_item" // the one catalog item every task item is, named per task

var taskObjects = []string{"Potion", "Flask", "Necklace", "Ring", "Dyes", "Green Gem", "Note", "Scroll", "Book", "Hammer", "Crown",
	"Blue orb", "Red orb", "Silver orb", "Wand", "Bottle", "Gold key", "Silver key", "Mirror", "Scroll tube", "Four leaf clover",
	"Vine", "Silver earring", "Gold bracelet", "Gold necklace", "Amulet", "Letter"}
var taskSizes = []string{"Little", "Small", "Large", "Big", "Thin"}

// wall is the time now for task limits, which run in real time.
func (w *World) wall() time.Time {
	if w.Clock != nil {
		return w.Clock()
	}
	return time.Now()
}

// ensureTaskItem puts the task item in the catalog if the data lacks it.
func ensureTaskItem(c *Catalog) {
	if c != nil && c.Items != nil && c.Items[taskItemID] == nil {
		c.Items[taskItemID] = &Item{ID: taskItemID, Name: "task item", ObjectType: 0, ItemType: 40, Level: 1}
	}
}

// loadTask reads a player's task row at login.
func (w *World) loadTask(p *Player) {
	t, err := w.Store.Task(p.Char.ID)
	if err != nil {
		w.Log("loading the task of %s: %v", p.Char.Name, err)
	}
	if t == nil {
		t = &store.Task{CharID: p.Char.ID}
	}
	p.task = t
}

func (z *Zone) saveTask(p *Player) {
	z.W.Store.Tx(func(tx *store.Tx) error { return tx.SaveTask(p.task) })
}

// taskView is the task message for a player.
func (z *Zone) taskView(p *Player) protocol.Task {
	z.ensureTask(p)
	t := p.task
	m := protocol.Task{T: protocol.TTask, Finish: t.Done, Max: rules.MaxTasks(p.Char.Level)}
	if !t.Active() {
		return m
	}
	m.On, m.Kind, m.Giver, m.Done = true, t.Kind, t.Giver, t.Killed
	m.Left = max(0, int(t.Expires-z.W.wall().Unix()))
	switch t.Kind {
	case "kill":
		m.Name = "Kill Task"
		if t.Killed {
			m.Text = "Return to " + t.Giver + " for your reward!"
		} else {
			m.Text = "Find a " + t.Target + " and kill it, then return to " + t.Giver + " for your reward."
		}
	default:
		m.Name = "Delivery Task"
		m.Text = "Bring the " + t.ItemName + " to " + t.Receiver + " in " + t.Place + "."
	}
	return m
}

func (z *Zone) sendTask(p *Player) { p.send(z.taskView(p)) }

// compass is the direction word from one place to another (north is up the map).
func compass(dx, dy float64) string {
	names := []string{"North", "North East", "East", "South East", "South", "South West", "West", "North West"}
	a := math.Atan2(dx, -dy) // 0 is north, clockwise
	i := int(math.Round(a/(math.Pi/4))+8) % 8
	return names[i]
}

// AskTask is the player's talk with an NPC about tasks.
func (z *Zone) AskTask(p *Player, id int64) {
	z.ensureTask(p)
	n := z.byID[id]
	if p.dead || n == nil || n.gone || n.dead() {
		p.errMsg("task", "There is nothing there.")
		return
	}
	if !n.inReach(p) {
		p.errMsg("task", "You are too far away to talk to "+n.title()+".")
		return
	}
	if z.taskTouch(p, n) {
		return
	}
	t := p.task
	if t.Active() {
		sys(p, n.title()+" says: \"You already have a task. Ask yourself with /task.\"")
		return
	}
	if p.Char.Level > rules.TaskMaxLevel {
		sys(p, fmt.Sprintf("Tasks are only available to players up to level %d.", rules.TaskMaxLevel))
		return
	}
	if t.Done >= rules.MaxTasks(p.Char.Level) {
		sys(p, fmt.Sprintf("You cannot do more than %d tasks at your level!", rules.MaxTasks(p.Char.Level)))
		return
	}
	kill := (n.Friendly || n.Guard) && rules.KillGiver(n.Name, n.Spawn.Guild)
	delivery := !kill && n.Friendly && rules.DeliveryGiver(n.Kind == KindTrainer, n.Kind == KindMerchant, n.Name)
	if !kill && !delivery {
		sys(p, n.title()+" says: \"I have no task for you.\"")
		return
	}
	if z.now < p.taskPauseUntil {
		sys(p, n.title()+" says: \"I have no tasks for you at the moment. Come back later, perhaps then you can help me with something.\"")
		return
	}
	if delivery && z.rng.Float64() >= rules.DeliveryChance {
		p.taskPauseUntil = z.now + rules.TaskAskPause
		sys(p, n.title()+" says: \"I have no tasks for you at the moment. Come back later, perhaps then you can help me with something.\"")
		return
	}
	if kill {
		z.giveKillTask(p, n)
	} else {
		z.giveDeliveryTask(p, n)
	}
}

func (z *Zone) giveKillTask(p *Player, giver *NPC) {
	lo, hi := rules.KillTargetLevels(p.Char.Level)
	var pool []*NPC
	for _, o := range z.npcs {
		if o.Friendly || o.Guard || o.isPet() || o.dead() || o.gone || o.Kind != "" || o.Name == "" || o.Name != strings.ToLower(o.Name) {
			continue
		}
		if o.Level >= lo && o.Level <= hi {
			pool = append(pool, o)
		}
	}
	if len(pool) == 0 {
		sys(p, giver.title()+" says: \"Sorry, I couldn't find any kill order for you. Come back later!\"")
		return
	}
	target := pool[z.rng.Intn(len(pool))]
	p.task = &store.Task{CharID: p.Char.ID, Kind: "kill", Giver: giver.title(), GiverID: giver.ID, Target: target.Name,
		Expires: z.W.wall().Unix() + rules.TaskSeconds, Done: p.task.Done}
	z.saveTask(p)
	dir := compass(target.Body.X-p.Body.X, target.Body.Y-p.Body.Y)
	sys(p, fmt.Sprintf("%s says: \"Very well, %s, it's good to see adventurers willing to help out the realm. Search to the %s and kill a %s and return to me for your reward. Good luck!\"",
		giver.title(), p.Char.Name, dir, target.Name))
	z.sendTask(p)
}

func (z *Zone) giveDeliveryTask(p *Player, giver *NPC) {
	var pool []*NPC
	for _, o := range z.npcs {
		if o == giver || !o.Friendly || o.Guard || o.isPet() || o.dead() || o.gone || o.Kind == KindNotClassic || o.Name == "" || o.Name == strings.ToLower(o.Name) {
			continue
		}
		pool = append(pool, o)
	}
	if len(pool) == 0 {
		sys(p, giver.title()+" says: \"I have no task for you, come back some time later.\"")
		return
	}
	recv := pool[z.rng.Intn(len(pool))]
	item := fmt.Sprintf("%s's %s %s", recv.title(), taskSizes[z.rng.Intn(len(taskSizes))], taskObjects[z.rng.Intn(len(taskObjects))])
	if _, err := z.W.addItem(p, taskItemID, 1); err != nil {
		sys(p, giver.title()+" says: \"Your pack is full; come back with room for the item.\"")
		return
	}
	p.task = &store.Task{CharID: p.Char.ID, Kind: "delivery", Giver: giver.title(), GiverID: giver.ID, Receiver: recv.title(), ReceiverID: recv.ID,
		Place: z.Name, ItemName: item, Expires: z.W.wall().Unix() + rules.TaskSeconds, Done: p.task.Done}
	z.saveTask(p)
	sys(p, fmt.Sprintf("%s says: \"Bring the %s to %s in %s.\"", giver.title(), item, recv.title(), z.Name))
	p.sendInv(z.W.Cat, p.Char.Money)
	z.sendTask(p)
}

// taskTouch finishes a task when the player talks to the NPC it ends at: the
// giver of a finished kill task, or the receiver of a delivery they carry the
// item for. It returns whether it did.
func (z *Zone) taskTouch(p *Player, n *NPC) bool {
	t := p.task
	if !t.Active() || z.expireTask(p) {
		return false
	}
	switch {
	case t.Kind == "kill" && t.Killed && n.title() == t.Giver:
		sys(p, n.title()+" says: \"Good work, "+p.Char.Name+". Here is your reward as promised.\"")
		z.finishTask(p)
		return true
	case t.Kind == "delivery" && n.title() == t.Receiver:
		slot := taskItemSlot(p)
		if slot < 0 {
			sys(p, n.title()+" says: \"You do not have what I am waiting for.\"")
			return true
		}
		z.W.Store.Tx(func(tx *store.Tx) error { return tx.RemoveItem(p.Char.ID, slot) })
		delete(p.inv, slot)
		sys(p, n.title()+" says: \"Good work, "+p.Char.Name+". Here is your reward as promised.\"")
		z.finishTask(p)
		return true
	}
	return false
}

// taskItemSlot is the pack slot of the task item, or -1.
func taskItemSlot(p *Player) int {
	for s := 0; s < PackSize; s++ {
		if it := p.inv[s]; it != nil && it.ID == taskItemID {
			return s
		}
	}
	return -1
}

// finishTask pays a finished task and clears it.
func (z *Zone) finishTask(p *Player) {
	t := p.task
	level := p.Char.Level
	delivery := t.Kind == "delivery"
	name := map[bool]string{false: "Kill Task", true: "Delivery Task"}[delivery]
	money := rules.TaskMoney(level, delivery, z.rng.Float64())
	t.Kind, t.Giver, t.Target, t.Receiver, t.Place, t.ItemName, t.Killed, t.Expires = "", "", "", "", "", "", false, 0
	t.GiverID, t.ReceiverID = 0, 0
	t.Done++
	p.Char.Money += int64(money)
	z.saveTask(p)
	sys(p, "You finish the "+name+"!")
	if money > 0 {
		sys(p, "You receive "+moneyText(int64(money))+" for completing your task.")
	}
	z.award(p, rules.TaskXP(level))
	z.W.savePlayer(p)
	p.sendInv(z.W.Cat, p.Char.Money)
	z.sendTask(p)
}

func moneyText(c int64) string {
	g, s, cu := c/10000, (c%10000)/100, c%100
	var parts []string
	if g > 0 {
		parts = append(parts, fmt.Sprintf("%d gold", g))
	}
	if g > 0 || s > 0 {
		parts = append(parts, fmt.Sprintf("%d silver", s))
	}
	parts = append(parts, fmt.Sprintf("%d copper", cu))
	return strings.Join(parts, " ")
}

// AbandonTask gives up the task: its item is removed and nothing is paid.
func (z *Zone) AbandonTask(p *Player) {
	z.ensureTask(p)
	if !p.task.Active() {
		sys(p, "You have no task.")
		return
	}
	z.dropTask(p, "You abandon your task.")
}

// expireTask ends a task whose time is up (or whose holder is past the level
// cap), returning whether it did.
func (z *Zone) expireTask(p *Player) bool {
	t := p.task
	if !t.Active() {
		return false
	}
	if z.W.wall().Unix() <= t.Expires && p.Char.Level <= rules.TaskMaxLevel {
		return false
	}
	z.dropTask(p, "Your task has expired!")
	return true
}

// dropTask clears a held task and its item, with a line to the player.
func (z *Zone) dropTask(p *Player, text string) {
	t := p.task
	if t.Kind == "delivery" {
		if slot := taskItemSlot(p); slot >= 0 {
			z.W.Store.Tx(func(tx *store.Tx) error { return tx.RemoveItem(p.Char.ID, slot) })
			delete(p.inv, slot)
			sys(p, "Your task related item has been removed from your inventory.")
			p.sendInv(z.W.Cat, p.Char.Money)
		}
	}
	t.Kind, t.Giver, t.Target, t.Receiver, t.Place, t.ItemName, t.Killed, t.Expires = "", "", "", "", "", "", false, 0
	t.GiverID, t.ReceiverID = 0, 0
	z.saveTask(p)
	sys(p, text)
	z.sendTask(p)
}

// taskKilled marks a kill task done for those who fought a creature of its kind
// and their groups.
func (z *Zone) taskKilled(n *NPC) {
	seen := map[int64]bool{}
	mark := func(p *Player) {
		if p == nil || seen[p.ID] {
			return
		}
		seen[p.ID] = true
		t := p.task
		if t == nil || t.Kind != "kill" || t.Killed || t.Target != n.Name {
			return
		}
		t.Killed = true
		z.saveTask(p)
		sys(p, "You must now return to "+t.Giver+" to receive your reward!")
		z.sendTask(p)
	}
	for _, id := range sortedIDs(n.hurtBy) {
		p := z.players[id]
		mark(p)
		if p == nil {
			continue
		}
		for _, m := range z.W.groupOf(p.Char.ID) {
			for _, o := range z.players {
				if o.Char.ID == m {
					mark(o)
				}
			}
		}
	}
}

// tickTask checks a player's task limit once in a while.
func (z *Zone) tickTask(p *Player) {
	if p.task != nil {
		z.expireTask(p)
	}
}

// taskLevelUp resets the count of tasks done, as a level does, and ends a task
// a player has outlevelled.
func (z *Zone) taskLevelUp(p *Player) {
	if p.task == nil {
		return
	}
	p.task.Done = 0
	if p.task.Active() && p.Char.Level > rules.TaskMaxLevel {
		z.dropTask(p, "Your task has expired!")
		return
	}
	z.saveTask(p)
	z.sendTask(p)
}

// TaskCommand is the /task chat command: "/task" says where the player stands,
// "/task ask" talks about tasks to the NPC they target, "/task abandon" gives
// the task up.
func (z *Zone) TaskCommand(p *Player, args []string) {
	if p.task == nil {
		p.task = &store.Task{CharID: p.Char.ID}
	}
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "ask":
			z.AskTask(p, p.target)
			return
		case "abandon", "abort":
			z.AbandonTask(p)
			return
		}
	}
	v := z.taskView(p)
	if !v.On {
		sys(p, fmt.Sprintf("You have no task. You have done %d of the %d tasks your level allows.", v.Finish, v.Max))
	} else {
		sys(p, fmt.Sprintf("%s: %s (%d minutes left; %d of %d tasks done at this level)", v.Name, v.Text, v.Left/60, v.Finish, v.Max))
	}
	z.sendTask(p)
}

// ensureTask gives a player an empty task record if none was loaded.
func (z *Zone) ensureTask(p *Player) {
	if p.task == nil {
		p.task = &store.Task{CharID: p.Char.ID}
	}
}
