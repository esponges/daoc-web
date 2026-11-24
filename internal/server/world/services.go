package world

import (
	"fmt"
	"sort"
	"strings"

	"daocweb/internal/importer"
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// What interacting with a trainer or a healer does.
const (
	KindTrainer = "trainer"
	KindHealer  = "healer"
)

// --- trainers ---

// openTrainer answers a trainer's interaction: the character's specs, or a
// refusal if the trainer teaches another class.
func (z *Zone) openTrainer(p *Player, n *NPC) {
	if _, ok := rules.Classes[p.Char.Class]; !ok || !classOf(p.Char.Class).TaughtBy(n.Teaches) {
		msg := n.title() + " cannot teach you."
		if seek := seekTrainers(p.Char.Class); seek != "" {
			msg += " Seek " + seek + "."
		}
		p.errMsg("train", msg)
		return
	}
	p.trainer = n.ID
	m := z.trainerMsg(p, n)
	m.Say = trainerGreeting(p.Char.Name, p.sh.unspent)
	p.send(m)
}

// seekTrainers names the trainers that teach a class: its own and its base
// class's, the pair TaughtBy accepts ("a Savage or Viking trainer").
func seekTrainers(class int) string {
	c, ok := rules.Classes[class]
	if !ok || c.Name == "" {
		return ""
	}
	if c.Base != "" && c.Base != c.Name {
		return article(c.Name) + " " + c.Name + " or " + c.Base + " trainer"
	}
	return article(c.Name) + " " + c.Name + " trainer"
}

// spendHint is what the level-up line adds when points are gained.
func spendHint(class int) string {
	c, ok := rules.Classes[class]
	if !ok || c.Name == "" {
		return " Visit your trainer to spend them."
	}
	return " Visit " + article(c.Name) + " " + c.Name + " trainer to spend them."
}

func article(word string) string {
	if word != "" && strings.ContainsRune("AEIOU", rune(word[0])) {
		return "an"
	}
	return "a"
}

// trainerGreeting is what a trainer says on being opened (our words).
func trainerGreeting(name string, unspent int) string {
	switch {
	case unspent <= 0:
		return fmt.Sprintf("Welcome, %s. You have no specialization points to spend; you gain more as you level.", name)
	case unspent == 1:
		return fmt.Sprintf("Welcome, %s. You have 1 specialization point to spend. Choose a specialization and press TRAIN.", name)
	}
	return fmt.Sprintf("Welcome, %s. You have %d specialization points to spend. Choose a specialization and press TRAIN.", name, unspent)
}

func (z *Zone) trainerMsg(p *Player, n *NPC) protocol.Trainer {
	m := protocol.Trainer{T: protocol.TTrainer, NPC: n.ID, Name: n.title(), Unspent: p.sh.unspent, Specs: []protocol.SpecRow{}}
	for _, k := range z.W.Cat.TrainableSpecs(p.Char.Class) {
		lv := p.specLevel(k)
		cost := rules.TrainCost(lv, lv+1)
		m.Specs = append(m.Specs, protocol.SpecRow{Key: k, Name: z.W.Cat.SpecName(k), Level: lv, Cost: cost,
			Can: lv < p.Char.Level && cost <= p.sh.unspent, Gives: z.W.Cat.SpecUnlocks(p.Char.Class, k, lv, 5)})
	}
	return m
}

// SpecUnlocks lists, in level order, up to max styles, abilities and spells a
// class gains from a spec above level: the data the Styles, Skills and Spells
// windows read once they are trained.
func (c *Catalog) SpecUnlocks(class int, spec string, level, max int) []protocol.Unlock {
	var out []protocol.Unlock
	for _, s := range c.StylesFor(class) {
		if s.Spec == spec && s.SpecLevel > level {
			out = append(out, protocol.Unlock{Level: s.SpecLevel, Name: s.Name, Kind: "style"})
		}
	}
	for _, g := range c.Grants[class] {
		if g.Source == spec && g.Level > level {
			out = append(out, protocol.Unlock{Level: g.Level, Name: g.Ability, Kind: "ability"})
		}
	}
	for _, l := range c.SpellsFor(class, spec) {
		if l.Level > level {
			name := ""
			if sp := c.Spells[l.Spell]; sp != nil {
				name = sp.Name
			}
			if name != "" {
				out = append(out, protocol.Unlock{Level: l.Level, Name: name, Kind: "spell"})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Level < out[j].Level })
	if len(out) > max {
		out = out[:max]
	}
	return out
}

// trainerInReach is a living trainer of the player's class within reach.
func (z *Zone) trainerInReach(p *Player) *NPC {
	class := classOf(p.Char.Class)
	for _, n := range z.npcs {
		if n.Kind == KindTrainer && class.TaughtBy(n.Teaches) && !n.dead() && !n.gone && n.inReach(p) {
			return n
		}
	}
	return nil
}

// Train raises one of the player's specs by levels at a trainer in reach,
// saved with the character in one transaction.
func (z *Zone) Train(p *Player, spec string, levels int) {
	if p.dead {
		return
	}
	n := z.trainerInReach(p)
	if n == nil {
		p.errMsg("train", "You need to be near a trainer of your class.")
		return
	}
	known := false
	for _, k := range z.W.Cat.TrainableSpecs(p.Char.Class) {
		known = known || k == spec
	}
	if !known {
		p.errMsg("train", "You cannot train that.")
		return
	}
	levels = max(1, levels)
	cur := p.specLevel(spec)
	if _, err := rules.Train(p.Char.Level, cur, levels, p.sh.unspent); err != nil {
		p.errMsg("train", "You cannot train "+z.W.Cat.SpecName(spec)+": "+err.Error()+".")
		return
	}
	p.Char.Specs[spec] = cur + levels
	if err := z.W.Store.Tx(func(tx *store.Tx) error { return tx.SaveCharacter(p.Char) }); err != nil {
		p.Char.Specs[spec] = cur
		p.errMsg("train", "Could not save that; try again.")
		return
	}
	z.W.recalc(p)
	z.buildBar(p)
	z.event(protocol.Event{K: "train", To: p.ID, Amount: cur + levels, Text: z.W.Cat.SpecName(spec)}, p.Body.X, p.Body.Y)
	p.send(z.trainerMsg(p, n))
}

// --- healers ---

func (z *Zone) openHealer(p *Player, n *NPC) {
	if p.Char.ConLost <= 0 {
		p.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "sys", From: n.title(), Text: "You have no lost constitution to restore."})
		return
	}
	price := rules.RestorePrice(p.Char.Level)
	p.send(protocol.Healer{T: protocol.THealer, NPC: n.ID, Name: n.title(), Lost: p.Char.ConLost, Price: price,
		Total: price * int64(p.Char.ConLost)})
}

// Restore pays a healer in reach for as much lost constitution as the
// player's money covers, money and constitution saved together.
func (z *Zone) Restore(p *Player, id int64) {
	n := z.byID[id]
	if p.dead || n == nil || n.Kind != KindHealer || n.dead() || n.gone || !n.inReach(p) {
		p.errMsg("restore", "You need to be near a healer.")
		return
	}
	c := p.Char
	if c.ConLost <= 0 {
		p.errMsg("restore", "You have no lost constitution to restore.")
		return
	}
	points, cost := rules.Restore(c.Level, c.ConLost, c.Money)
	if points == 0 {
		p.errMsg("restore", fmt.Sprintf("You need %d copper to restore a point of constitution.", rules.RestorePrice(c.Level)))
		return
	}
	c.Money -= cost
	c.ConLost -= points
	if err := z.W.Store.Tx(func(tx *store.Tx) error { return tx.SaveCharacter(c) }); err != nil {
		c.Money += cost
		c.ConLost += points
		p.errMsg("restore", "Could not save that; try again.")
		return
	}
	z.W.recalc(p)
	p.sendInv(z.W.Cat, c.Money)
	p.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "sys", From: n.title(),
		Text: fmt.Sprintf("%s restores %d constitution for %d copper.", n.title(), points, cost)})
}

// teaches is the class a trainer's class type teaches, "" for anything else.
func teaches(classType string) string { return importer.TrainerClass(classType) }

// checkTrainer closes the open trainer window when its trainer is out of
// reach, dead or gone, or the player is dead. Run each tick.
func (z *Zone) checkTrainer(p *Player) {
	if p.trainer == 0 {
		return
	}
	n := z.byID[p.trainer]
	if p.dead || n == nil || n.gone || n.dead() || !n.inReach(p) {
		p.trainer = 0
		p.send(protocol.TrainClose{T: protocol.TTrainClose})
	}
}
