package world

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"daocweb/internal/importer"
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// Crafting, from the crafting capability: a character joins one primary
// profession at its master, then makes what the recipes of its skills allow
// from materials in the pack. A craft takes time and is cancelled by moving or
// by a fight. Stations, spellcraft imbuing and salvage are not built yet.

// KindCraftMaster is a master of a profession, who teaches it and opens the
// crafting window.
const KindCraftMaster = "craftmaster"

// craftMoveLimit is how far a crafter may drift before the craft is lost.
const craftMoveLimit = 8.0

// craftState is a player's tradeskills and the craft they are making.
type craftState struct {
	primary int
	skills  map[int]int
	making  *making
	master  int64 // the craft master last talked to, whom "join" means
}

// making is a craft under way.
type making struct {
	rec  *Recipe
	ends float64
	x, y float64
}

// loadCraft reads a character's tradeskills at login.
func (w *World) loadCraft(p *Player) {
	c, err := w.Store.LoadCraft(p.Char.ID)
	if err != nil {
		w.Log("loading tradeskills of %s: %v", p.Char.Name, err)
	}
	p.craft = &craftState{primary: c.Primary, skills: c.Skills}
	if p.craft.skills == nil {
		p.craft.skills = map[int]int{}
	}
}

func (cs *craftState) isPrimary(skill int) bool { return skill == cs.primary }

// capOf is the most a skill may reach now.
func (cs *craftState) capOf(skill int) int {
	if cs.isPrimary(skill) {
		return rules.MaxSkill
	}
	return rules.SecondaryCap(cs.skills[cs.primary])
}

// Craft runs a crafting request; it is called in the player's zone.
func (z *Zone) Craft(p *Player, m *protocol.CraftOp) {
	if p.craft == nil || p.dead {
		return
	}
	switch m.Op {
	case "open":
		z.sendCraft(p, z.masterNear(p))
	case "join":
		z.joinCraft(p)
	case "make":
		z.startCraft(p, m.ID)
	case "cancel":
		z.cancelCraft(p, "You stop crafting.")
	}
}

// masterNear is the craft master the player last talked to, if still in reach,
// or any craft master in reach.
func (z *Zone) masterNear(p *Player) *NPC {
	if n := z.byID[p.craft.master]; n != nil && n.Kind == KindCraftMaster && !n.gone && !n.dead() && n.inReach(p) {
		return n
	}
	return z.nearNPC(p, KindCraftMaster)
}

func masterSkill(n *NPC) int {
	if n == nil {
		return 0
	}
	return importer.CraftMasterSkill(n.Spawn.ClassType)
}

// openMaster is what talking to a craft master does.
func (z *Zone) openMaster(p *Player, n *NPC) {
	if p.craft == nil {
		return
	}
	p.craft.master = n.ID
	prof, ok := rules.Professions[masterSkill(n)]
	switch {
	case !ok:
		sys(p, n.title()+" says: \"I have nothing to teach you.\"")
	case p.craft.primary == 0:
		sys(p, n.title()+" says: \"I can teach you "+prof.Name+". You may follow only one craft as your own.\"")
	case p.craft.primary == prof.Primary:
		sys(p, n.title()+" says: \"Show me what you can make.\"")
	default:
		sys(p, n.title()+" says: \"You already follow another craft.\"")
	}
	z.sendCraft(p, n)
}

// joinCraft makes the player a crafter of the profession of the master they
// stand by.
func (z *Zone) joinCraft(p *Player) {
	n := z.masterNear(p)
	prof, ok := rules.Professions[masterSkill(n)]
	switch {
	case n == nil || !ok:
		p.errMsg("craft", "You must be near a craft master.")
		return
	case p.craft.primary != 0:
		p.errMsg("craft", "You already follow a craft.")
		return
	}
	cs := p.craft
	cs.primary = prof.Primary
	cs.skills[prof.Primary] = rules.StartSkill
	for _, s := range prof.Secondary {
		cs.skills[s] = rules.StartSkill
	}
	if !z.saveCraft(p) {
		cs.primary = 0
		cs.skills = map[int]int{}
		p.errMsg("craft", "Could not save that; try again.")
		return
	}
	sys(p, "You are now a "+prof.Name+" crafter.")
	z.sendCraft(p, n)
}

func (z *Zone) saveCraft(p *Player) bool {
	err := z.W.Store.Tx(func(tx *store.Tx) error {
		return tx.SaveCraft(p.Char.ID, store.Craft{Primary: p.craft.primary, Skills: p.craft.skills})
	})
	if err != nil {
		z.W.Log("saving tradeskills of %s: %v", p.Char.Name, err)
	}
	return err == nil
}

// have counts what the pack holds of an item.
func (p *Player) have(id string) int {
	n := 0
	for slot := 0; slot < PackSize; slot++ {
		if it := p.inv[slot]; it != nil && it.ID == id {
			n += max(1, it.Count)
		}
	}
	return n
}

// reach is the recipes the player's skills could make now, easiest first.
func (z *Zone) reach(p *Player) []*Recipe {
	var out []*Recipe
	for sk, v := range p.craft.skills {
		for _, r := range z.W.Cat.RecipesBySkill[sk] {
			if rules.SuccessChance(v, r.Level) > 0 {
				out = append(out, r)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Skill != out[j].Skill {
			return out[i].Skill < out[j].Skill
		}
		if out[i].Level != out[j].Level {
			return out[i].Level < out[j].Level
		}
		return out[i].Product < out[j].Product
	})
	return out
}

// sendCraft sends the crafting window. A master by the player who teaches a
// profession the player lacks adds that offer.
func (z *Zone) sendCraft(p *Player, master *NPC) {
	cs, cat := p.craft, z.W.Cat
	m := protocol.Craft{T: protocol.TCraft, Skills: []protocol.CraftSkill{}, Recipes: []protocol.CraftRecipe{}}
	if cs.primary != 0 {
		m.Prof, m.ProfName = cs.primary, rules.Professions[cs.primary].Name
	} else if prof, ok := rules.Professions[masterSkill(master)]; ok {
		m.Offer, m.OfferName = prof.Primary, prof.Name
	}
	ids := make([]int, 0, len(cs.skills))
	for sk := range cs.skills {
		ids = append(ids, sk)
	}
	sort.Ints(ids)
	for _, sk := range ids {
		m.Skills = append(m.Skills, protocol.CraftSkill{ID: sk, Name: rules.SkillName(sk), Value: cs.skills[sk], Cap: cs.capOf(sk)})
	}
	for _, r := range z.reach(p) {
		it := cat.Items[r.Product]
		if it == nil {
			continue
		}
		cr := protocol.CraftRecipe{ID: r.Product, Name: it.Name, Skill: r.Skill, Level: r.Level,
			Chance: rules.SuccessChance(cs.skills[r.Skill], r.Level), Secs: rules.CraftSeconds(r.Level, r.MatCount()), Mats: []protocol.CraftMat{}}
		for _, mt := range r.Mats {
			name := mt.Item
			if x := cat.Items[mt.Item]; x != nil {
				name = x.Name
			}
			cr.Mats = append(cr.Mats, protocol.CraftMat{ID: mt.Item, Name: name, N: mt.Count, Have: p.have(mt.Item)})
		}
		m.Recipes = append(m.Recipes, cr)
	}
	p.send(m)
}

// startCraft begins making the recipe of a product.
func (z *Zone) startCraft(p *Player, product string) {
	cs := p.craft
	r := z.W.Cat.Recipes[product]
	switch {
	case r == nil:
		p.errMsg("craft", "You do not know how to make that.")
		return
	case cs.making != nil:
		p.errMsg("craft", "You are already crafting.")
		return
	case z.now < p.combatUntil || p.casting != nil:
		p.errMsg("craft", "You cannot craft while fighting or casting.")
		return
	}
	v, has := cs.skills[r.Skill]
	if !has {
		p.errMsg("craft", "You do not have the skill to make that.")
		return
	}
	if rules.SuccessChance(v, r.Level) == 0 {
		p.errMsg("craft", "That is too far beyond your skill.")
		return
	}
	for _, mt := range r.Mats {
		if p.have(mt.Item) < mt.Count {
			p.errMsg("craft", "You lack the materials.")
			return
		}
	}
	secs := rules.CraftSeconds(r.Level, r.MatCount())
	cs.making = &making{rec: r, ends: z.now + float64(secs), x: p.Body.X, y: p.Body.Y}
	p.send(protocol.CraftProg{T: protocol.TCraftProg, ID: product, Secs: secs})
}

// cancelCraft stops a craft under way, using nothing.
func (z *Zone) cancelCraft(p *Player, text string) {
	cs := p.craft
	if cs == nil || cs.making == nil {
		return
	}
	id := cs.making.rec.Product
	cs.making = nil
	p.send(protocol.CraftDone{T: protocol.TCraftDone, ID: id, Text: text})
	sys(p, text)
}

// craftTick finishes or cancels the craft under way; called every tick.
func (z *Zone) craftTick(p *Player) {
	cs := p.craft
	if cs == nil || cs.making == nil {
		return
	}
	m := cs.making
	switch {
	case math.Hypot(p.Body.X-m.x, p.Body.Y-m.y) > craftMoveLimit:
		z.cancelCraft(p, "You move and stop crafting.")
	case z.now < p.combatUntil || p.casting != nil:
		z.cancelCraft(p, "You are interrupted and stop crafting.")
	case z.now >= m.ends:
		z.finishCraft(p)
	}
}

var errNoRoom = errors.New("your pack is full")

// finishCraft rolls the craft: what is used up, the item, and the skill.
func (z *Zone) finishCraft(p *Player) {
	cs := p.craft
	r := cs.making.rec
	cs.making = nil
	cat := z.W.Cat
	skill := cs.skills[r.Skill]
	ok := z.rng.Intn(100) < rules.SuccessChance(skill, r.Level)
	q := 0
	if ok {
		q = rules.Quality(skill, r.Level, z.rng.Float64)
	}
	gain := ok && z.rng.Intn(100) < rules.GainChance(skill, r.Level) &&
		rules.MayRaise(r.Skill, skill, cs.isPrimary(r.Skill), cs.skills[cs.primary])

	// What is used up: everything on a success, half of each stack on a failure.
	use := map[string]int{}
	for _, mt := range r.Mats {
		if ok {
			use[mt.Item] += mt.Count
		} else {
			use[mt.Item] += rules.FailureLoss(mt.Count)
		}
	}
	var newSlot = -1
	err := z.W.Store.Tx(func(tx *store.Tx) error {
		for id, n := range use {
			if err := takeMaterials(tx, p, id, n); err != nil {
				return err
			}
		}
		if ok {
			slot, err := tx.FreeSlot(p.Char.ID, PackSize)
			if err != nil {
				return err
			}
			if slot < 0 {
				return errNoRoom
			}
			newSlot = slot
			if err := tx.AddCrafted(p.Char.ID, r.Product, slot, 1, q, p.Char.Name); err != nil {
				return err
			}
		}
		if gain {
			cs.skills[r.Skill] = skill + 1
			return tx.SaveCraft(p.Char.ID, store.Craft{Primary: cs.primary, Skills: cs.skills})
		}
		return nil
	})
	// The store rolled the move back on an error; the pack in memory must too.
	if err != nil {
		cs.skills[r.Skill] = skill
		z.reloadPack(p)
		text := "Could not finish that; try again."
		if errors.Is(err, errNoRoom) {
			text = "Your pack is full, and the craft is lost."
		}
		p.send(protocol.CraftDone{T: protocol.TCraftDone, ID: r.Product, Text: text})
		sys(p, text)
		return
	}
	// Mirror the store in memory.
	for id, n := range use {
		dropMaterials(p, id, n)
	}
	name := r.Product
	if it := cat.Items[r.Product]; it != nil {
		name = it.Name
	}
	done := protocol.CraftDone{T: protocol.TCraftDone, ID: r.Product, OK: ok, Q: q, Gain: gain}
	if ok {
		p.inv[newSlot] = &invItem{ID: r.Product, Slot: newSlot, Count: 1, Cond: cat.newCondition(r.Product), Quality: q, Crafter: p.Char.Name}
		done.Text = fmt.Sprintf("You make a %s of %d%% quality.", name, q)
		if rules.Masterpiece(q) {
			done.Text = "You make a masterpiece: " + name + "!"
		}
	} else {
		done.Text = "You fail to make the " + name + " and lose some of the materials."
	}
	if gain {
		done.Text += fmt.Sprintf(" Your %s rises to %d.", rules.SkillName(r.Skill), cs.skills[r.Skill])
	}
	p.send(done)
	sys(p, done.Text)
	p.sendInv(cat, p.Char.Money)
	z.sendCraft(p, z.masterNear(p))
}

// takeMaterials removes n of an item from the pack's stacks in the store.
func takeMaterials(tx *store.Tx, p *Player, id string, n int) error {
	for _, slot := range sortedSlots(p.inv) {
		it := p.inv[slot]
		if slot >= PackSize || it.ID != id || n <= 0 {
			continue
		}
		have := max(1, it.Count)
		if have <= n {
			if err := tx.RemoveItem(p.Char.ID, slot); err != nil {
				return err
			}
			n -= have
		} else {
			if err := tx.SetCount(p.Char.ID, slot, have-n); err != nil {
				return err
			}
			n = 0
		}
	}
	if n > 0 {
		return errors.New("materials are gone")
	}
	return nil
}

// dropMaterials is takeMaterials on the pack in memory, after the store agreed.
func dropMaterials(p *Player, id string, n int) {
	for _, slot := range sortedSlots(p.inv) {
		it := p.inv[slot]
		if slot >= PackSize || it.ID != id || n <= 0 {
			continue
		}
		have := max(1, it.Count)
		if have <= n {
			delete(p.inv, slot)
			n -= have
		} else {
			it.Count = have - n
			n = 0
		}
	}
}

// reloadPack rereads the pack from the store after a rolled-back change.
func (z *Zone) reloadPack(p *Player) {
	z.W.loadInventory(p)
	p.sendInv(z.W.Cat, p.Char.Money)
}
