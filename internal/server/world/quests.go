package world

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// Quests, from the quests capability. A quest is a data file written by this
// project (server/quests/*.json): who gives it, what its steps ask, what it
// pays. Progress is a row per quest in players.db.

// Step kinds.
const (
	StepTalk    = "talk"    // speak to an NPC
	StepKill    = "kill"    // kill a number of a creature
	StepItem    = "item"    // a creature drops a quest item, to hold a number of
	StepDeliver = "deliver" // hand a held item to an NPC
	StepGoto    = "goto"    // reach a place
)

// QuestDef is one quest as its file says.
type QuestDef struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Source     string      `json:"source"` // where the quest is described publicly; the wording here is our own
	MinLevel   int         `json:"min_level"`
	MaxLevel   int         `json:"max_level"` // 0 for no limit
	Classes    []string    `json:"classes"`   // class names; empty for any
	Requires   string      `json:"requires"`  // a quest that must be done first
	Repeatable bool        `json:"repeatable"`
	Giver      string      `json:"giver"` // the NPC who offers it
	Offer      string      `json:"offer"`
	Complete   string      `json:"complete"`
	Items      []QuestItem `json:"items"`
	Steps      []QuestStep `json:"steps"`
	Rewards    QuestReward `json:"rewards"`
}

// QuestItem is an item that exists for a quest: it is not in the database.
type QuestItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// QuestStep is one thing a quest asks.
type QuestStep struct {
	Kind   string  `json:"kind"`
	Text   string  `json:"text"`   // the journal line
	NPC    string  `json:"npc"`    // talk, deliver: who
	Say    string  `json:"say"`    // talk, deliver: what they answer
	Target string  `json:"target"` // kill, item: the creature's name
	Item   string  `json:"item"`   // item, deliver: a QuestItem id
	Count  int     `json:"count"`
	Chance int     `json:"chance"` // item: percent per kill
	Zone   int     `json:"zone"`   // goto
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Radius float64 `json:"radius"`
}

// QuestReward is what finishing pays. Choose lists items of which the player
// takes one.
type QuestReward struct {
	XPPct  float64  `json:"xp_pct"` // percent of the experience the player's current level takes
	Money  int64    `json:"money"`  // copper
	Items  []string `json:"items"`
	Choose []string `json:"choose"`
}

// Quests is the loaded set, by id, in file order.
type Quests struct {
	byID  map[string]*QuestDef
	order []*QuestDef
}

// LoadQuests reads every *.json file under dir (and its folders). A missing
// directory is no quests, not an error.
func LoadQuests(dir string) (*Quests, error) {
	q := &Quests{byID: map[string]*QuestDef{}}
	if dir == "" {
		return q, nil
	}
	var files []string
	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".json") {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var d QuestDef
		dec := json.NewDecoder(strings.NewReader(string(b)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&d); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if err := d.Check(); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if q.byID[d.ID] != nil {
			return nil, fmt.Errorf("%s: quest id %q is used twice", f, d.ID)
		}
		q.byID[d.ID] = &d
		q.order = append(q.order, &d)
	}
	for _, d := range q.order {
		if d.Requires != "" && q.byID[d.Requires] == nil {
			return nil, fmt.Errorf("quest %q requires %q, which does not exist", d.ID, d.Requires)
		}
	}
	return q, nil
}

// Get is a quest by id.
func (q *Quests) Get(id string) *QuestDef {
	if q == nil {
		return nil
	}
	return q.byID[id]
}

// All lists the quests in file order.
func (q *Quests) All() []*QuestDef {
	if q == nil {
		return nil
	}
	return q.order
}

// Check says what is wrong with a quest on its own: the shape, not whether the
// world has what it names (see CheckWorld).
func (d *QuestDef) Check() error {
	bad := func(f string, a ...any) error { return fmt.Errorf("quest %q: "+f, append([]any{d.ID}, a...)...) }
	switch {
	case d.ID == "" || d.Title == "" || d.Giver == "" || d.Offer == "":
		return bad("needs an id, title, giver and offer text")
	case d.Source == "":
		return bad("must name its source")
	case len(d.Steps) == 0:
		return bad("has no steps")
	case d.MinLevel < 1:
		return bad("min_level must be at least 1")
	}
	have := map[string]bool{}
	for _, it := range d.Items {
		if it.ID == "" || it.Name == "" || have[it.ID] {
			return bad("an item needs a unique id and a name")
		}
		have[it.ID] = true
	}
	for i, s := range d.Steps {
		if s.Text == "" {
			return bad("step %d has no journal text", i+1)
		}
		switch s.Kind {
		case StepTalk:
			if s.NPC == "" {
				return bad("step %d: talk needs an npc", i+1)
			}
		case StepKill:
			if s.Target == "" || s.Count < 1 {
				return bad("step %d: kill needs a target and a count", i+1)
			}
		case StepItem:
			if s.Target == "" || s.Count < 1 || s.Chance < 1 || s.Chance > 100 || !have[s.Item] {
				return bad("step %d: item needs a target, a count, a chance of 1 to 100 and one of the quest's items", i+1)
			}
		case StepDeliver:
			if s.NPC == "" || s.Count < 1 || !have[s.Item] {
				return bad("step %d: deliver needs an npc, a count and one of the quest's items", i+1)
			}
		case StepGoto:
			if s.Zone == 0 || s.Radius <= 0 {
				return bad("step %d: goto needs a zone and a radius", i+1)
			}
		default:
			return bad("step %d has an unknown kind %q", i+1, s.Kind)
		}
	}
	if last := d.Steps[len(d.Steps)-1].Kind; last != StepTalk && last != StepDeliver {
		return bad("the last step must be to talk to or deliver to someone, who pays")
	}
	for _, c := range d.Classes {
		if classByName(c) == 0 {
			return bad("unknown class %q", c)
		}
	}
	return nil
}

// classByName is the class id for a name, 0 if there is none.
func classByName(name string) int {
	for id, c := range rules.Classes {
		if strings.EqualFold(c.Name, name) {
			return id
		}
	}
	return 0
}

// CheckWorld lists what a quest names that the world data lacks: NPCs,
// creatures and reward items.
func (d *QuestDef) CheckWorld(c *Catalog) []string {
	var out []string
	npc := func(name string) {
		if name != "" && !c.HasSpawnNamed(name) {
			out = append(out, fmt.Sprintf("quest %q: no NPC named %q", d.ID, name))
		}
	}
	npc(d.Giver)
	for _, s := range d.Steps {
		npc(s.NPC)
		if s.Target != "" && !c.HasSpawnNamed(s.Target) {
			out = append(out, fmt.Sprintf("quest %q: no creature named %q", d.ID, s.Target))
		}
	}
	for _, id := range append(append([]string{}, d.Rewards.Items...), d.Rewards.Choose...) {
		if c.Items[id] == nil {
			out = append(out, fmt.Sprintf("quest %q: no item %q", d.ID, id))
		}
	}
	return out
}

// HasSpawnNamed says whether the world places an NPC or creature of that name
// (compared without case).
func (c *Catalog) HasSpawnNamed(name string) bool {
	if c.db == nil {
		return true // a test catalogue has no store to ask
	}
	var n int
	c.db.QueryRow(`SELECT count(*) FROM spawn WHERE lower(name)=lower(?)`, name).Scan(&n)
	return n > 0
}

// ensureQuestItems puts the items quests define into the catalogue.
func ensureQuestItems(c *Catalog, q *Quests) {
	for _, d := range q.All() {
		for _, it := range d.Items {
			if c.Items[it.ID] == nil {
				c.Items[it.ID] = &Item{ID: it.ID, Name: it.Name, ObjectType: 0, ItemType: 40, Level: 1, MaxCount: 20}
			}
		}
	}
}

// --- the player's side ---

// loadQuests reads a character's quests at login.
func (w *World) loadQuests(p *Player) {
	rows, err := w.Store.Quests(p.Char.ID)
	if err != nil {
		w.Log("loading quests of %s: %v", p.Char.Name, err)
	}
	p.quests = map[string]*store.Quest{}
	for i := range rows {
		q := rows[i]
		p.quests[q.Quest] = &q
	}
}

func (z *Zone) saveQuest(p *Player, q *store.Quest) {
	z.W.Store.Tx(func(tx *store.Tx) error { return tx.SaveQuest(*q) })
}

// available says whether a player may be offered a quest: not started or done
// (unless repeatable), and meeting its conditions.
func (z *Zone) available(p *Player, d *QuestDef) bool {
	if q := p.quests[d.ID]; q != nil && !(q.Done && d.Repeatable) {
		return false
	}
	lv := p.Char.Level
	if lv < d.MinLevel || (d.MaxLevel > 0 && lv > d.MaxLevel) {
		return false
	}
	if len(d.Classes) > 0 {
		ok := false
		for _, c := range d.Classes {
			if classByName(c) == p.Char.Class {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}
	if d.Requires != "" {
		if r := p.quests[d.Requires]; r == nil || !r.Done {
			return false
		}
	}
	return true
}

// active is the quest rows a player is on.
func (p *Player) activeQuests() []*store.Quest {
	var out []*store.Quest
	for _, q := range p.quests {
		if !q.Done {
			out = append(out, q)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Quest < out[j].Quest })
	return out
}

// questHeld counts a quest item in the pack.
func (p *Player) questHeld(item string) int { return p.have(item) }

// stepOf is the step a quest row is on, or nil when it is past the last
// (waiting on a reward choice).
func stepOf(d *QuestDef, q *store.Quest) *QuestStep {
	if q.Step < 0 || q.Step >= len(d.Steps) {
		return nil
	}
	return &d.Steps[q.Step]
}

// questTouch runs what talking to an NPC does for quests: it advances talk
// and deliver steps that end here, and shows the quests this NPC offers.
// It returns whether it consumed the interaction.
func (z *Zone) questTouch(p *Player, n *NPC) bool {
	qs := z.W.Quests
	if qs == nil || p.quests == nil {
		return false
	}
	handled := false
	for _, q := range p.activeQuests() {
		d := qs.Get(q.Quest)
		if d == nil {
			continue
		}
		s := stepOf(d, q)
		if s == nil || !strings.EqualFold(s.NPC, n.Name) {
			continue
		}
		if q.Step == len(d.Steps)-1 && freeSlots(p) < rewardSlots(d) {
			// The last step pays: make room first, so nothing is lost.
			sys(p, fmt.Sprintf("%s says: \"Come back with %d free pack slot(s) for your reward.\"", n.title(), rewardSlots(d)))
			handled = true
			continue
		}
		switch s.Kind {
		case StepTalk:
			if s.Say != "" {
				sys(p, n.title()+" says: \""+s.Say+"\"")
			}
			z.questStepDone(p, d, q, n)
			handled = true
		case StepDeliver:
			if p.questHeld(s.Item) < s.Count {
				sys(p, n.title()+" says: \"You do not have what I am waiting for.\"")
				handled = true
				continue
			}
			if s.Say != "" {
				sys(p, n.title()+" says: \""+s.Say+"\"")
			}
			if z.takeQuestItems(p, s.Item, s.Count) {
				z.questStepDone(p, d, q, n)
			}
			handled = true
		}
	}
	if handled {
		return true
	}
	for _, d := range qs.All() {
		if strings.EqualFold(d.Giver, n.Name) && z.available(p, d) {
			p.questOffer = d.ID
			p.send(protocol.QuestOffer{T: protocol.TQuestOffer, ID: d.ID, Title: d.Title, NPC: n.title(), Text: d.Offer, Reward: rewardText(p, d)})
			sys(p, n.title()+" says: \""+d.Offer+"\"")
			return false // the NPC's own services stay open beside the offer
		}
	}
	return false
}

func rewardText(p *Player, d *QuestDef) string {
	var parts []string
	if d.Rewards.XPPct > 0 {
		parts = append(parts, fmt.Sprintf("%.1f%% of a level", d.Rewards.XPPct))
	}
	if d.Rewards.Money > 0 {
		parts = append(parts, moneyText(d.Rewards.Money))
	}
	if n := len(d.Rewards.Items); n > 0 {
		parts = append(parts, fmt.Sprintf("%d item(s)", n))
	}
	if len(d.Rewards.Choose) > 0 {
		parts = append(parts, "a choice of items")
	}
	return strings.Join(parts, ", ")
}

// takeQuestItems removes count of an item from the pack, in the store first.
func (z *Zone) takeQuestItems(p *Player, item string, count int) bool {
	err := z.W.Store.Tx(func(tx *store.Tx) error { return takeMaterials(tx, p, item, count) })
	if err != nil {
		z.W.Log("taking quest items from %s: %v", p.Char.Name, err)
		return false
	}
	dropMaterials(p, item, count)
	p.sendInv(z.W.Cat, p.Char.Money)
	return true
}

// QuestOp runs a quest request; it is called in the player's zone.
func (z *Zone) QuestOp(p *Player, m *protocol.QuestOp) {
	if p.quests == nil || z.W.Quests == nil {
		return
	}
	switch m.Op {
	case "accept":
		z.acceptQuest(p, m.ID)
	case "decline":
		if p.questOffer == m.ID {
			p.questOffer = ""
		}
	case "abandon":
		z.abandonQuest(p, m.ID)
	case "choose":
		z.chooseReward(p, m.ID, m.N)
	case "journal":
		z.sendQuests(p)
	}
}

// acceptQuest starts the quest the player was offered, if its giver is by.
func (z *Zone) acceptQuest(p *Player, id string) {
	d := z.W.Quests.Get(id)
	if d == nil || p.questOffer != id || !z.available(p, d) {
		p.errMsg("quest", "That quest is not on offer.")
		return
	}
	giver := false
	for _, n := range z.npcs {
		if strings.EqualFold(n.Name, d.Giver) && !n.dead() && !n.gone && n.inReach(p) {
			giver = true
		}
	}
	if !giver {
		p.errMsg("quest", "You are too far from "+d.Giver+".")
		return
	}
	q := &store.Quest{CharID: p.Char.ID, Quest: id}
	p.quests[id] = q
	p.questOffer = ""
	z.saveQuest(p, q)
	sys(p, "You begin the quest: "+d.Title+".")
	z.questProgress(p, d, q) // a step may be met already (items held)
	z.sendQuests(p)
}

// abandonQuest drops an active quest, taking its items.
func (z *Zone) abandonQuest(p *Player, id string) {
	q, d := p.quests[id], z.W.Quests.Get(id)
	if q == nil || d == nil || q.Done {
		p.errMsg("quest", "You are not on that quest.")
		return
	}
	for _, it := range d.Items {
		if n := p.have(it.ID); n > 0 {
			z.takeQuestItems(p, it.ID, n)
		}
	}
	z.W.Store.Tx(func(tx *store.Tx) error { return tx.DeleteQuest(p.Char.ID, id) })
	delete(p.quests, id)
	sys(p, "You abandon the quest: "+d.Title+".")
	z.sendQuests(p)
}

// questStepDone moves a quest past its current step, finishing it after the last.
func (z *Zone) questStepDone(p *Player, d *QuestDef, q *store.Quest, giver *NPC) {
	q.Step++
	q.Count = 0
	z.saveQuest(p, q)
	if q.Step >= len(d.Steps) {
		z.finishQuest(p, d, q, giver)
		return
	}
	if t := d.Steps[q.Step].Text; t != "" {
		sys(p, "Quest: "+d.Title+" - "+t)
	}
	z.questProgress(p, d, q)
	z.sendQuests(p)
}

// questProgress completes steps that are already met: items held, a place reached.
func (z *Zone) questProgress(p *Player, d *QuestDef, q *store.Quest) {
	for {
		s := stepOf(d, q)
		if s == nil {
			return
		}
		switch {
		case s.Kind == StepItem && p.questHeld(s.Item) >= s.Count:
		case s.Kind == StepGoto && z.ID == s.Zone && math.Hypot(p.Body.X-s.X, p.Body.Y-s.Y) <= s.Radius:
		default:
			return
		}
		q.Step++
		q.Count = 0
		z.saveQuest(p, q)
		if t := stepTextAt(d, q.Step); t != "" {
			sys(p, "Quest: "+d.Title+" - "+t)
		}
	}
}

func stepTextAt(d *QuestDef, i int) string {
	if i < 0 || i >= len(d.Steps) {
		return ""
	}
	return d.Steps[i].Text
}

// questKilled counts a kill for the players who fought the creature and their
// groups: kill steps count it, and item steps may drop their item.
func (z *Zone) questKilled(n *NPC) {
	qs := z.W.Quests
	if qs == nil {
		return
	}
	seen := map[int64]bool{}
	credit := func(p *Player) {
		if p == nil || seen[p.ID] || p.quests == nil {
			return
		}
		seen[p.ID] = true
		changed := false
		for _, q := range p.activeQuests() {
			d := qs.Get(q.Quest)
			s := stepOf(d, q)
			if d == nil || s == nil || !strings.EqualFold(s.Target, n.Name) {
				continue
			}
			switch s.Kind {
			case StepKill:
				q.Count++
				if q.Count >= s.Count {
					z.questStepDone(p, d, q, nil)
				} else {
					z.saveQuest(p, q)
					sys(p, fmt.Sprintf("%s: %d of %d.", d.Title, q.Count, s.Count))
				}
				changed = true
			case StepItem:
				if z.rng.Intn(100) >= s.Chance || p.questHeld(s.Item) >= s.Count {
					continue
				}
				if _, err := z.W.addItem(p, s.Item, 1); err != nil {
					sys(p, "Your pack is full; the quest item is lost.")
					continue
				}
				name := s.Item
				if it := z.W.Cat.Items[s.Item]; it != nil {
					name = it.Name
				}
				sys(p, fmt.Sprintf("You find %s (%d of %d).", name, p.questHeld(s.Item), s.Count))
				z.questProgress(p, d, q)
				changed = true
			}
		}
		if changed {
			z.sendQuests(p)
		}
	}
	for _, id := range sortedIDs(n.hurtBy) {
		p := z.players[id]
		credit(p)
		if p == nil {
			continue
		}
		for _, m := range z.W.groupOf(p.Char.ID) {
			for _, o := range z.players {
				if o.Char.ID == m {
					credit(o)
				}
			}
		}
	}
}

// questTick completes goto steps; called every tick for a player on a quest.
func (z *Zone) questTick(p *Player) {
	qs := z.W.Quests
	if qs == nil || len(p.quests) == 0 {
		return
	}
	for _, q := range p.activeQuests() {
		d := qs.Get(q.Quest)
		if d == nil {
			continue
		}
		if s := stepOf(d, q); s != nil && s.Kind == StepGoto && z.ID == s.Zone && math.Hypot(p.Body.X-s.X, p.Body.Y-s.Y) <= s.Radius {
			z.questStepDone(p, d, q, nil)
		}
	}
}

// finishQuest pays a quest whose last step is done. With a choice of items it
// waits for the player to pick; a pack without room holds the reward back.
func (z *Zone) finishQuest(p *Player, d *QuestDef, q *store.Quest, giver *NPC) {
	if d.Complete != "" && giver != nil {
		sys(p, giver.title()+" says: \""+d.Complete+"\"")
	}
	need := rewardSlots(d)
	if free := freeSlots(p); free < need {
		// Hold the quest on its last step; the player comes back with room.
		q.Step = len(d.Steps) - 1
		z.saveQuest(p, q)
		sys(p, fmt.Sprintf("Your pack needs %d free slot(s) for the reward.", need))
		return
	}
	z.payQuest(p, d, q)
	if len(d.Rewards.Choose) > 0 {
		q.Step = len(d.Steps)
		z.saveQuest(p, q)
		m := protocol.QuestReward{T: protocol.TQuestReward, ID: d.ID, Title: d.Title}
		for i, id := range d.Rewards.Choose {
			name := id
			if it := z.W.Cat.Items[id]; it != nil {
				name = it.Name
			}
			m.Choose = append(m.Choose, protocol.QuestChoice{N: i, ID: id, Name: name})
		}
		p.send(m)
		return
	}
	z.closeQuest(p, d, q)
}

// rewardSlots is the pack room a quest's reward needs.
func rewardSlots(d *QuestDef) int {
	n := len(d.Rewards.Items)
	if len(d.Rewards.Choose) > 0 {
		n++
	}
	return n
}

// freeSlots counts empty pack slots.
func freeSlots(p *Player) int {
	n := 0
	for s := 0; s < PackSize; s++ {
		if p.inv[s] == nil {
			n++
		}
	}
	return n
}

// payQuest gives experience, money and the fixed items.
func (z *Zone) payQuest(p *Player, d *QuestDef, q *store.Quest) {
	r := d.Rewards
	if r.Money > 0 {
		p.Char.Money += r.Money
		sys(p, "You receive "+moneyText(r.Money)+".")
	}
	for _, id := range r.Items {
		if _, err := z.W.addItem(p, id, 1); err != nil {
			z.W.Log("quest %s reward %s to %s: %v", d.ID, id, p.Char.Name, err)
		}
	}
	if r.XPPct > 0 && p.Char.Level < MaxLevel {
		lv := p.Char.Level
		z.award(p, int64(float64(rules.XPToReach(lv+1)-rules.XPToReach(lv))*r.XPPct/100))
	}
	z.W.savePlayer(p)
	p.sendInv(z.W.Cat, p.Char.Money)
}

// chooseReward gives the item the player picked from a finished quest.
func (z *Zone) chooseReward(p *Player, id string, n int) {
	q, d := p.quests[id], z.W.Quests.Get(id)
	if q == nil || d == nil || q.Done || q.Step < len(d.Steps) || n < 0 || n >= len(d.Rewards.Choose) {
		p.errMsg("quest", "There is no such choice.")
		return
	}
	if _, err := z.W.addItem(p, d.Rewards.Choose[n], 1); err != nil {
		p.errMsg("quest", "Your pack is full.")
		return
	}
	z.closeQuest(p, d, q)
}

func (z *Zone) closeQuest(p *Player, d *QuestDef, q *store.Quest) {
	q.Done = true
	z.saveQuest(p, q)
	sys(p, "You have completed the quest: "+d.Title+"!")
	z.sendQuests(p)
}

// sendQuests sends the journal: active quests with the current step, and
// the titles of those done.
func (z *Zone) sendQuests(p *Player) {
	m := protocol.Quests{T: protocol.TQuests, Active: []protocol.QuestLine{}, Done: []string{}}
	qs := z.W.Quests
	for _, q := range p.activeQuests() {
		d := qs.Get(q.Quest)
		if d == nil {
			continue
		}
		l := protocol.QuestLine{ID: d.ID, Title: d.Title}
		if s := stepOf(d, q); s != nil {
			l.Text, l.Need = s.Text, s.Count
			switch s.Kind {
			case StepKill:
				l.Have = q.Count
			case StepItem, StepDeliver:
				l.Have = p.questHeld(s.Item)
			}
		} else {
			l.Text = "Choose your reward."
		}
		m.Active = append(m.Active, l)
	}
	var done []string
	for id, q := range p.quests {
		if d := qs.Get(id); q.Done && d != nil {
			done = append(done, d.Title)
		}
	}
	sort.Strings(done)
	m.Done = append(m.Done, done...)
	p.send(m)
}
