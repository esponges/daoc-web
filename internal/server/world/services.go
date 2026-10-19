package world

import (
	"fmt"

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
		p.errMsg("train", n.title()+" cannot teach you.")
		return
	}
	p.send(z.trainerMsg(p, n))
}

func (z *Zone) trainerMsg(p *Player, n *NPC) protocol.Trainer {
	m := protocol.Trainer{T: protocol.TTrainer, NPC: n.ID, Name: n.title(), Unspent: p.sh.unspent, Specs: []protocol.SpecRow{}}
	for _, k := range z.W.Cat.TrainableSpecs(p.Char.Class) {
		lv := p.specLevel(k)
		cost := rules.TrainCost(lv, lv+1)
		m.Specs = append(m.Specs, protocol.SpecRow{Key: k, Name: z.W.Cat.SpecName(k), Level: lv, Cost: cost,
			Can: lv < p.Char.Level && cost <= p.sh.unspent})
	}
	return m
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
