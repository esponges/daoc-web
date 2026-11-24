package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// questRig is a zone with a giver, a receiver and a wolf pack, and a quest that
// asks for 2 wolf kills, then a pelt from the pack, then its delivery.
func questRig(t *testing.T) (*rig, *QuestDef) {
	t.Helper()
	r := newRig(t,
		Spawn{ID: 980, Name: "Ragna", Model: 56, X: 30000, Y: 30000, Z: 5000, Level: 30, Size: 50, Speed: 191, Realm: 2},
		Spawn{ID: 981, Name: "Svala", Model: 56, X: 30200, Y: 30000, Z: 5000, Level: 30, Size: 50, Speed: 191, Realm: 2},
		wolf(990, 30100, 30500, 1), wolf(991, 30120, 30500, 1), wolf(992, 30140, 30500, 1))
	d := &QuestDef{
		ID: "wolves", Title: "Wolves at the Gate", Source: "test", MinLevel: 1, MaxLevel: 10, Classes: []string{"Warrior"},
		Giver: "Ragna", Offer: "Thin the wolves for me.", Complete: "Well done.",
		Items: []QuestItem{{ID: "quest_pelt", Name: "a fine pelt"}},
		Steps: []QuestStep{
			{Kind: StepKill, Target: "wolf", Count: 2, Text: "Kill two wolves."},
			{Kind: StepItem, Target: "wolf", Item: "quest_pelt", Count: 1, Chance: 100, Text: "Find a fine pelt."},
			{Kind: StepDeliver, NPC: "Svala", Item: "quest_pelt", Count: 1, Text: "Bring it to Svala.", Say: "Thank you."},
		},
		Rewards: QuestReward{XPPct: 10, Money: 500, Items: []string{"shield"}},
	}
	if err := d.Check(); err != nil {
		t.Fatal(err)
	}
	r.w.Quests = &Quests{byID: map[string]*QuestDef{d.ID: d}, order: []*QuestDef{d}}
	ensureQuestItems(r.w.Cat, r.w.Quests)
	return r, d
}

func questPlayer(t *testing.T, r *rig, name string) *Player {
	p := r.player(name, 2, 30050, 30000, false)
	r.w.loadQuests(p)
	drain(p)
	return p
}

func kill(z *Zone, p *Player, id int64) {
	n := z.byID[id]
	n.hurtBy[p.ID]++
	n.HP = 0
	z.killNPC(n)
}

func lastOffer(p *Player) (protocol.QuestOffer, bool) {
	all := sentOf[protocol.QuestOffer](p, protocol.TQuestOffer)
	if len(all) == 0 {
		return protocol.QuestOffer{}, false
	}
	return all[len(all)-1], true
}

func TestQuestOfferAndAccept(t *testing.T) {
	r, d := questRig(t)
	p := questPlayer(t, r, "Warriorone")
	r.z.Interact(p, 980)
	off, ok := lastOffer(p)
	if !ok || off.ID != d.ID || off.Title != d.Title || off.Reward == "" {
		t.Fatalf("offer: %+v %v", off, ok)
	}
	// A different class sees no offer, nor does a level beyond the quest's.
	other := r.player("Healerone", 2, 30060, 30000, false)
	other.Char.Class = 10
	r.w.loadQuests(other)
	drain(other)
	r.z.Interact(other, 980)
	if _, ok := lastOffer(other); ok {
		t.Error("a class the quest is not for was offered it")
	}
	high := r.player("Veteran", 20, 30070, 30000, false)
	r.w.loadQuests(high)
	drain(high)
	r.z.Interact(high, 980)
	if _, ok := lastOffer(high); ok {
		t.Error("a level 20 was offered a level 1 to 10 quest")
	}
	r.z.QuestOp(p, &protocol.QuestOp{Op: "accept", ID: d.ID})
	if q := p.quests[d.ID]; q == nil || q.Step != 0 || q.Done {
		t.Fatalf("accepted: %+v", p.quests[d.ID])
	}
	j := sentOf[protocol.Quests](p, protocol.TQuests)
	if len(j) == 0 || len(j[len(j)-1].Active) != 1 || j[len(j)-1].Active[0].Text != "Kill two wolves." {
		t.Errorf("journal: %+v", j)
	}
	drain(p)
	r.z.Interact(p, 980)
	if _, ok := lastOffer(p); ok {
		t.Error("offered a quest already taken")
	}
}

func TestQuestWholeRoad(t *testing.T) {
	r, d := questRig(t)
	p := questPlayer(t, r, "Warriorone")
	r.z.Interact(p, 980)
	r.z.QuestOp(p, &protocol.QuestOp{Op: "accept", ID: d.ID})
	xp0, money0 := p.Char.XP, p.Char.Money

	kill(r.z, p, 990)
	if q := p.quests[d.ID]; q.Step != 0 || q.Count != 1 {
		t.Fatalf("after one kill: %+v", q)
	}
	kill(r.z, p, 991)
	q := p.quests[d.ID]
	if q.Step != 1 {
		t.Fatalf("after two kills: %+v", q)
	}
	if p.have("quest_pelt") != 0 {
		t.Fatal("the item dropped before its step")
	}
	kill(r.z, p, 992)
	if p.have("quest_pelt") != 1 || q.Step != 2 {
		t.Fatalf("the pelt: held %d, step %d", p.have("quest_pelt"), q.Step)
	}
	r.z.Interact(p, 980)
	if p.have("quest_pelt") != 1 {
		t.Error("the giver took the pelt")
	}
	r.z.Interact(p, 981)
	if !q.Done || p.have("quest_pelt") != 0 {
		t.Fatalf("not finished: %+v, pelt %d", q, p.have("quest_pelt"))
	}
	if p.Char.Money != money0+500 || p.Char.XP <= xp0 || !packHas(p, "shield") {
		t.Errorf("rewards: money %d (was %d), xp %d (was %d), shield %v", p.Char.Money, money0, p.Char.XP, xp0, packHas(p, "shield"))
	}
	again := &Player{Char: p.Char}
	r.w.loadQuests(again)
	if a := again.quests[d.ID]; a == nil || !a.Done {
		t.Errorf("saved as %+v", again.quests[d.ID])
	}
	drain(p)
	r.z.Interact(p, 980)
	if _, ok := lastOffer(p); ok {
		t.Error("a finished quest was offered again")
	}
}

func TestQuestItemOnlyDropsOnTheQuest(t *testing.T) {
	r, d := questRig(t)
	p := questPlayer(t, r, "Warriorone")
	kill(r.z, p, 990)
	if p.have("quest_pelt") != 0 {
		t.Error("a quest item dropped for someone not on the quest")
	}
	r.z.Interact(p, 980)
	r.z.QuestOp(p, &protocol.QuestOp{Op: "accept", ID: d.ID})
	p.quests[d.ID].Step = 1
	kill(r.z, p, 991)
	kill(r.z, p, 992)
	if p.have("quest_pelt") != 1 {
		t.Errorf("held %d pelts, want exactly the 1 the step needs", p.have("quest_pelt"))
	}
}

func TestQuestAbandonAndRepeat(t *testing.T) {
	r, d := questRig(t)
	p := questPlayer(t, r, "Warriorone")
	r.z.Interact(p, 980)
	r.z.QuestOp(p, &protocol.QuestOp{Op: "accept", ID: d.ID})
	p.quests[d.ID].Step = 2
	r.w.addItem(p, "quest_pelt", 1)
	r.z.QuestOp(p, &protocol.QuestOp{Op: "abandon", ID: d.ID})
	if p.quests[d.ID] != nil || p.have("quest_pelt") != 0 {
		t.Error("abandoning kept the quest or its item")
	}
	drain(p)
	r.z.Interact(p, 980)
	if _, ok := lastOffer(p); !ok {
		t.Error("an abandoned quest cannot be taken again")
	}
	r.z.QuestOp(p, &protocol.QuestOp{Op: "decline", ID: d.ID})
	r.z.QuestOp(p, &protocol.QuestOp{Op: "accept", ID: d.ID})
	if p.quests[d.ID] != nil {
		t.Error("accepted after declining")
	}
}

func TestQuestRewardChoiceAndFullPack(t *testing.T) {
	r, d := questRig(t)
	d.Rewards = QuestReward{Money: 100, Choose: []string{"shield", "helm"}}
	d.Steps = d.Steps[2:]
	p := questPlayer(t, r, "Warriorone")
	r.z.Interact(p, 980)
	r.z.QuestOp(p, &protocol.QuestOp{Op: "accept", ID: d.ID})
	r.w.addItem(p, "quest_pelt", 1)
	for i := 0; i < PackSize-2; i++ {
		if _, err := r.w.addItem(p, "pelt", 1); err != nil {
			t.Fatal(err)
		}
	}
	if freeSlots(p) != 1 {
		t.Fatalf("setup: %d free", freeSlots(p))
	}
	// One free slot is enough for one chosen item.
	drain(p)
	r.z.Interact(p, 981)
	rw := sentOf[protocol.QuestReward](p, protocol.TQuestReward)
	if len(rw) != 1 || len(rw[0].Choose) != 2 {
		t.Fatalf("no choice offered: %+v", rw)
	}
	if p.quests[d.ID].Done {
		t.Fatal("done before choosing")
	}
	r.z.QuestOp(p, &protocol.QuestOp{Op: "choose", ID: d.ID, N: 5})
	if p.quests[d.ID].Done {
		t.Error("a choice that does not exist was accepted")
	}
	r.z.QuestOp(p, &protocol.QuestOp{Op: "choose", ID: d.ID, N: 1})
	if !p.quests[d.ID].Done || !packHas(p, "helm") || packHas(p, "shield") {
		t.Errorf("chose the helm: done %v helm %v shield %v", p.quests[d.ID].Done, packHas(p, "helm"), packHas(p, "shield"))
	}
}

func TestQuestFullPackHoldsTheReward(t *testing.T) {
	r, d := questRig(t)
	p := questPlayer(t, r, "Warriorone")
	r.z.Interact(p, 980)
	r.z.QuestOp(p, &protocol.QuestOp{Op: "accept", ID: d.ID})
	p.quests[d.ID].Step = 2
	for i := 0; i < PackSize; i++ {
		r.w.addItem(p, "pelt", 1)
	}
	// The pack is full of pelts, and one of them is no quest item: the quest
	// item has no room, so put it in over a pelt.
	delete(p.inv, 0)
	r.st.Tx(func(tx *store.Tx) error { return tx.RemoveItem(p.Char.ID, 0) })
	r.w.addItem(p, "quest_pelt", 1)
	drain(p)
	r.z.Interact(p, 981)
	if p.quests[d.ID].Done || p.have("quest_pelt") != 1 {
		t.Fatalf("paid into a full pack: done %v, pelt %d", p.quests[d.ID].Done, p.have("quest_pelt"))
	}
}

func TestQuestGotoStep(t *testing.T) {
	r, d := questRig(t)
	d.Steps = []QuestStep{
		{Kind: StepGoto, Zone: 100, X: 31000, Y: 30000, Radius: 100, Text: "Go to the ford."},
		{Kind: StepTalk, NPC: "Svala", Text: "Tell Svala."},
	}
	p := questPlayer(t, r, "Warriorone")
	r.z.Interact(p, 980)
	r.z.QuestOp(p, &protocol.QuestOp{Op: "accept", ID: d.ID})
	r.ticks(2)
	if p.quests[d.ID].Step != 0 {
		t.Fatal("the place was reached from afar")
	}
	p.Body.X, p.Body.Y = 31010, 30000
	r.ticks(2)
	if p.quests[d.ID].Step != 1 {
		t.Errorf("at the ford, on step %d", p.quests[d.ID].Step)
	}
}

func TestQuestFilesAreChecked(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	talk := `{"kind":"talk","npc":"Svala","text":"Talk."}`
	good := `{"id":"a","title":"A","source":"s","min_level":1,"giver":"Ragna","offer":"o","steps":[` + talk + `],"rewards":{"money":1}}`
	write("a.json", good)
	if q, err := LoadQuests(dir); err != nil || q.Get("a") == nil {
		t.Fatalf("a good quest: %v", err)
	}
	for name, bad := range map[string]string{
		"no source":     strings.Replace(good, `"source":"s",`, ``, 1),
		"no steps":      strings.Replace(good, talk, ``, 1),
		"last is kill":  strings.Replace(good, talk, `{"kind":"kill","target":"wolf","count":1,"text":"Kill."}`, 1),
		"unknown field": strings.Replace(good, `"min_level":1,`, `"min_level":1,"colour":"red",`, 1),
		"bad class":     strings.Replace(good, `"min_level":1,`, `"min_level":1,"classes":["Jester"],`, 1),
		"unknown kind":  strings.Replace(good, `"kind":"talk"`, `"kind":"dance"`, 1),
		"missing prior": strings.Replace(good, `"min_level":1,`, `"min_level":1,"requires":"zzz",`, 1),
	} {
		write("a.json", bad)
		if _, err := LoadQuests(dir); err == nil {
			t.Errorf("%s: loaded", name)
		}
	}
	write("a.json", good)
	write("b.json", good)
	if _, err := LoadQuests(dir); err == nil {
		t.Error("a repeated id loaded")
	}
	if q, err := LoadQuests(filepath.Join(dir, "none")); err != nil || len(q.All()) != 0 {
		t.Error("a missing folder is no quests, not an error")
	}
}
