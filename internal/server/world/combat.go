package world

import (
	"fmt"
	"math"
	"sort"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
)

// Result codes a style's requirement refers to (the style table's numbers).
const (
	resAny    = 0
	resMiss   = 1
	resHit    = 2
	resParry  = 3
	resBlock  = 4
	resEvade  = 5
	resStyle  = 7
	openAny   = 0
	openDef   = 1
	openPos   = 2
	posBack   = 0
	posSide   = 1
	posFront  = 2
	unarmedSp = 2.2
)

func (p *Player) hasWeapon() bool { return p.weapon != nil }
func (p *Player) swingSpeed() float64 {
	if p.weapon != nil && p.weapon.Speed > 0 {
		return p.weapon.Speed
	}
	return unarmedSp
}
func (p *Player) weaponDPS() float64 {
	if p.weapon != nil && p.weapon.DPS > 0 {
		return p.weapon.DPS
	}
	return 1.2
}

// posOf is where the attacker stands relative to a defender's facing.
func posOf(defender Body, ax, ay float64) int {
	a := angleOff(defender.Yaw, angleTo(defender.X, defender.Y, ax, ay))
	switch {
	case a >= backArc:
		return posBack
	case a <= frontArc:
		return posFront
	default:
		return posSide
	}
}

// openingText describes a style's requirement for the bar.
func (c *Catalog) openingText(s *StyleDef) string {
	switch s.OpenType {
	case openDef:
		switch s.ResultReq {
		case resParry:
			return "after a parry"
		case resBlock:
			return "after a block"
		case resEvade:
			return "after an evade"
		}
		return "after a defence"
	case openPos:
		switch s.OpenValue {
		case posBack:
			return "from behind"
		case posSide:
			return "from the side"
		}
		return "from the front"
	}
	switch s.ResultReq {
	case resStyle:
		for i := range c.Styles {
			if c.Styles[i].Class == s.Class && c.Styles[i].ID == s.OpenValue {
				return "after " + c.Styles[i].Name
			}
		}
		return "after another style"
	case resHit:
		return "after a hit"
	case resMiss:
		return "after a miss"
	}
	return "anytime"
}

// buildBar chooses the styles on a player's hotbar: the lines for what they
// hold, up to the level they have reached.
func (z *Zone) buildBar(p *Player) {
	z.computeBar(p)
	p.sendBar(z.W.Cat)
}

// computeBar chooses the bar without sending it.
func (z *Zone) computeBar(p *Player) {
	p.bar = p.bar[:0]
	if z.W.Cat == nil {
		return
	}
	for _, s := range z.W.Cat.StylesFor(p.Char.Class) {
		if s.SpecLevel > p.Char.Level+1 {
			continue
		}
		switch {
		case s.WeaponReq == ObjShield:
			if p.shield == nil {
				continue
			}
		case p.weapon == nil || p.weapon.ObjectType != s.WeaponReq:
			continue
		}
		p.bar = append(p.bar, s)
	}
	sort.SliceStable(p.bar, func(i, j int) bool {
		a, b := p.bar[i], p.bar[j]
		if (a.WeaponReq == ObjShield) != (b.WeaponReq == ObjShield) {
			return b.WeaponReq == ObjShield
		}
		return a.SpecLevel < b.SpecLevel
	})
}

func (p *Player) sendBar(c *Catalog) {
	b := protocol.Bar{T: protocol.TStyleBar}
	for i := range p.bar {
		s := &p.bar[i]
		b.Slots = append(b.Slots, protocol.BarSlot{Slot: i, ID: s.ID, Name: s.Name, Icon: s.Icon, Spec: s.Spec, Cost: s.Endurance,
			Opens: c.openingText(s), Level: s.SpecLevel})
	}
	p.send(b)
}

// styleReady says whether a style can be used on the current target now.
func (z *Zone) styleReady(p *Player, s *StyleDef) (bool, string) {
	if p.end < float64(s.Endurance) {
		return false, "not enough endurance"
	}
	switch s.OpenType {
	case openDef:
		if p.lastDef.kind == "" || z.now-p.lastDef.at > DefensiveWindow || defCode(p.lastDef.kind) != s.ResultReq {
			return false, "needs " + z.W.Cat.openingText(s)
		}
	case openPos:
		n := z.byID[p.target]
		if n == nil || n.dead() {
			return false, "no target"
		}
		if posOf(n.Body, p.Body.X, p.Body.Y) != s.OpenValue {
			return false, "needs " + z.W.Cat.openingText(s)
		}
	default:
		switch s.ResultReq {
		case resStyle:
			if p.lastStyle != s.OpenValue || z.now-p.lastStyleAt > ChainWindow {
				return false, "needs " + z.W.Cat.openingText(s)
			}
		case resHit:
			if !p.lastHit || z.now-p.lastHitAt > ChainWindow {
				return false, "needs a hit first"
			}
		case resMiss:
			if p.lastHit || z.now-p.lastHitAt > ChainWindow {
				return false, "needs a miss first"
			}
		}
	}
	return true, ""
}

func defCode(kind string) int {
	switch kind {
	case "parry":
		return resParry
	case "block":
		return resBlock
	case "evade":
		return resEvade
	}
	return -1
}

// requestStyle queues the style on a bar slot for the next swing.
func (z *Zone) requestStyle(p *Player, slot int) {
	if p.dead || slot < 0 || slot >= len(p.bar) {
		return
	}
	if !p.attacking {
		p.errMsg("style", "Start attacking first.")
		return
	}
	if ok, why := z.styleReady(p, &p.bar[slot]); !ok {
		p.errMsg("style", p.bar[slot].Name+": "+why+".")
		return
	}
	p.pending = slot
}

// SetTarget selects an NPC (id 0 clears it).
func (z *Zone) SetTarget(p *Player, id int64) {
	if id == 0 {
		p.target, p.attacking, p.pending = 0, false, -1
		return
	}
	n := z.byID[id]
	if n == nil || n.gone {
		return
	}
	if p.target != id {
		p.attacking, p.pending = false, -1
	}
	p.target = id
	if n.dead() && n.lootOwner == p.ID {
		z.sendCorpse(p, n)
	}
}

// SetAttack switches auto-attack.
func (z *Zone) SetAttack(p *Player, on bool) {
	if !on {
		p.attacking, p.pending = false, -1
		return
	}
	n := z.byID[p.target]
	if p.dead || n == nil || n.dead() {
		p.errMsg("target", "You need a target.")
		return
	}
	if n.Friendly {
		p.errMsg("target", "You cannot attack "+n.title()+".")
		return
	}
	p.attacking = true
	if p.swingAt < z.now {
		p.swingAt = z.now + 0.4
	}
}

// combatPlayer runs a player's side of a fight for the tick.
func (z *Zone) combatPlayer(p *Player) {
	if !p.attacking {
		return
	}
	n := z.byID[p.target]
	if n == nil || n.dead() || n.gone {
		p.attacking, p.pending = false, -1
		return
	}
	d := math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y)
	reach := 100 + n.Reach*0.5
	if d > reach+30 {
		return
	}
	if angleOff(p.Body.Yaw, angleTo(p.Body.X, p.Body.Y, n.Body.X, n.Body.Y)) > frontArc*1.2 {
		return
	}
	if z.now < p.swingAt {
		return
	}
	p.swingAt = z.now + p.swingSpeed()
	z.playerSwing(p, n)
}

func (z *Zone) playerSwing(p *Player, n *NPC) {
	c := p.Char
	var style *StyleDef
	if p.pending >= 0 && p.pending < len(p.bar) {
		def := p.bar[p.pending]
		if ok, _ := z.styleReady(p, &def); ok {
			style = &def
			p.end -= float64(def.Endurance)
		}
	}
	p.pending = -1
	p.combatUntil = z.now + CombatLinger
	ev := protocol.Event{K: "swing", From: p.ID, To: n.ID}
	if style != nil {
		ev.Style = style.Anim
		ev.Text = style.Name
	}
	z.event(ev, p.Body.X, p.Body.Y)

	bonus := 0
	if style != nil {
		bonus = style.BonusToHit
	}
	if z.rng.Float64() >= playerHitChance(c.Level, n.Level, bonus) {
		z.event(protocol.Event{K: "miss", From: p.ID, To: n.ID}, n.Body.X, n.Body.Y)
		p.lastHit, p.lastHitAt, p.lastStyle = false, z.now, 0
		z.provoke(n, p)
		return
	}
	// The creature's defences. Block and parry only cover what it can see.
	seen := angleOff(n.Body.Yaw, angleTo(n.Body.X, n.Body.Y, p.Body.X, p.Body.Y)) <= frontArc*1.3
	for _, d := range []struct {
		kind   string
		chance float64
		ok     bool
	}{{"evade", n.Evade, true}, {"parry", n.Parry, seen}, {"block", n.Block, seen}} {
		if d.ok && d.chance > 0 && z.rng.Float64() < d.chance {
			z.event(protocol.Event{K: d.kind, From: n.ID, To: p.ID}, n.Body.X, n.Body.Y)
			p.lastHit, p.lastHitAt, p.lastStyle = false, z.now, 0
			z.provoke(n, p)
			return
		}
	}
	dmg := weaponDamage(p.weaponDPS(), p.swingSpeed(), c.Level) * (0.75 + 0.5*z.rng.Float64())
	crit := false
	if style != nil {
		dmg = dmg*(1+style.GrowthRate*1.4) + style.GrowthOff
		p.lastStyle, p.lastStyleAt = style.ID, z.now
	} else {
		p.lastStyle = 0
		if z.rng.Float64() < 0.06 {
			dmg *= 1.5
			crit = true
		}
	}
	amount := int(math.Max(1, math.Round(dmg)))
	p.lastHit, p.lastHitAt = true, z.now
	z.event(protocol.Event{K: "hit", From: p.ID, To: n.ID, Amount: amount, Crit: crit, Text: ev.Text}, n.Body.X, n.Body.Y)
	n.HP -= amount
	n.hurtBy[p.ID] += amount
	if n.tagger == 0 {
		n.tagger = p.ID
	}
	z.provoke(n, p)
	if n.HP <= 0 {
		z.killNPC(n)
	}
}

// provoke makes a creature fight the one who struck at it.
func (z *Zone) provoke(n *NPC, p *Player) {
	if n.dead() {
		return
	}
	if n.target == nil && n.state != npcReturn {
		n.engage(p, z)
	}
}

// npcSwing resolves a creature's blow against a player.
func (z *Zone) npcSwing(n *NPC, p *Player) {
	z.event(protocol.Event{K: "swing", From: n.ID, To: p.ID}, n.Body.X, n.Body.Y)
	p.combatUntil = z.now + CombatLinger
	if z.rng.Float64() >= npcHitChance(n.Level, p.Char.Level) {
		z.event(protocol.Event{K: "miss", From: n.ID, To: p.ID}, p.Body.X, p.Body.Y)
		return
	}
	seen := angleOff(p.Body.Yaw, angleTo(p.Body.X, p.Body.Y, n.Body.X, n.Body.Y)) <= frontArc*1.3
	lv := p.Char.Level
	for _, d := range []struct {
		kind   string
		chance float64
	}{
		{"evade", playerEvade(lv)},
		{"parry", ifSeen(seen, playerParry(lv, p.hasWeapon()))},
		{"block", ifSeen(seen, playerBlock(lv, p.shield != nil))},
	} {
		if d.chance > 0 && z.rng.Float64() < d.chance {
			z.event(protocol.Event{K: d.kind, From: p.ID, To: n.ID}, p.Body.X, p.Body.Y)
			p.lastDef = defence{kind: d.kind, at: z.now}
			return
		}
	}
	dmg := n.Damage * (0.75 + 0.5*z.rng.Float64()) * (1 - armourReduction(p.armour, n.Level))
	amount := int(math.Max(1, math.Round(dmg)))
	z.event(protocol.Event{K: "hit", From: n.ID, To: p.ID, Amount: amount}, p.Body.X, p.Body.Y)
	p.hp -= float64(amount)
	if p.hp <= 0 {
		z.killPlayer(p)
	}
}

func ifSeen(seen bool, v float64) float64 {
	if seen {
		return v
	}
	return 0
}

// killPlayer ends a player's life until they return to their bind point.
func (z *Zone) killPlayer(p *Player) {
	p.hp = 0
	p.dead = true
	p.deadAt = z.now
	p.attacking, p.target, p.pending = false, 0, -1
	p.queue, p.queuedSteps = nil, 0
	z.event(protocol.Event{K: "death", To: p.ID}, p.Body.X, p.Body.Y)
	for _, n := range z.npcs {
		if n.target == p {
			n.goHome()
		}
	}
}

// ReleasePlayer stands a dead player up now, at their bind point, without
// waiting for the respawn timer. It does nothing to a living one.
func (z *Zone) ReleasePlayer(p *Player) {
	if p.dead {
		z.respawnPlayer(p)
	}
}

// respawnPlayer stands a dead player up at their bind point.
func (z *Zone) respawnPlayer(p *Player) {
	bz, x, y, zz, yaw := z.W.bindFor(p)
	p.dead = false
	p.hp = float64(p.maxHP()) * 0.6
	p.end = float64(playerMaxEnd())
	p.combatUntil = 0
	if bz != z.ID {
		z.W.moveToZone(z, p, bz, x, y, zz)
		return
	}
	p.Body = Body{X: x, Y: y, Z: z.G.GroundAt(x, y, zz), Yaw: yaw}
	p.correct = true
}

// killNPC finishes a creature: experience, loot and the clock to its return.
func (z *Zone) killNPC(n *NPC) {
	n.state = npcDead
	n.diedAt = z.now
	rs := float64(n.Spawn.RespawnS)
	n.respawnAt = z.now + clamp(rs, 25, 120)
	n.target = nil
	n.speedNow = 0
	z.event(protocol.Event{K: "death", To: n.ID}, n.Body.X, n.Body.Y)

	total := 0
	for _, d := range n.hurtBy {
		total += d
	}
	var top *Player
	best := 0
	for _, id := range sortedIDs(n.hurtBy) {
		p := z.players[id]
		if p == nil || p.closed.Load() {
			continue
		}
		share := float64(n.hurtBy[id]) / math.Max(1, float64(total))
		if n.hurtBy[id] > best {
			top, best = p, n.hurtBy[id]
		}
		if xp := int64(math.Round(float64(killXP(n.Level, p.Char.Level)) * share)); xp > 0 {
			z.award(p, xp)
		}
		if p.target == n.ID {
			p.attacking, p.pending = false, -1
		}
	}
	for _, p := range z.players {
		if p.target == n.ID && p.attacking {
			p.attacking = false
		}
	}
	if top != nil {
		n.lootOwner = top.ID
		n.money = int64(float64(killMoney(n.Level)) * (0.7 + 0.6*z.rng.Float64()))
		n.loot = z.rollLoot(n)
		if len(n.loot) > 0 || n.money > 0 {
			z.sendCorpse(top, n)
		}
	}
}

// award gives experience and raises the level when it is earned.
func (z *Zone) award(p *Player, xp int64) {
	c := p.Char
	c.XP += xp
	z.event(protocol.Event{K: "xp", To: p.ID, Amount: int(xp)}, p.Body.X, p.Body.Y)
	for c.Level < MaxLevel && c.XP >= xpToReach(c.Level+1) {
		c.Level++
		p.hp = float64(p.maxHP())
		p.end = float64(playerMaxEnd())
		z.event(protocol.Event{K: "level", To: p.ID, Amount: c.Level, Text: fmt.Sprintf("You have reached level %d!", c.Level)}, p.Body.X, p.Body.Y)
		z.buildBar(p)
	}
	z.W.savePlayer(p)
}

// rollLoot decides what a dead creature dropped.
func (z *Zone) rollLoot(n *NPC) []lootItem {
	var out []lootItem
	cat := z.W.Cat
	for _, e := range cat.LootFor(n.Name) {
		if cat.Items[e.Item] == nil || z.rng.Intn(100) >= e.Chance {
			continue
		}
		cnt := e.Count
		if cnt < 1 {
			cnt = 1
		}
		out = append(out, lootItem{ID: e.Item, Count: cnt})
	}
	if len(out) == 0 && z.rng.Float64() < 0.22 {
		if id := cat.randomDrop(z.rng, n.Level); id != "" {
			out = append(out, lootItem{ID: id, Count: 1})
		}
	}
	return out
}

func (z *Zone) sendCorpse(p *Player, n *NPC) {
	m := protocol.Corpse{T: protocol.TCorpse, ID: n.ID, Name: n.title(), Money: n.money}
	for i, l := range n.loot {
		it := z.W.Cat.Items[l.ID]
		if it == nil {
			continue
		}
		m.Items = append(m.Items, itemView(i, it, l.Count))
	}
	p.send(m)
}

// takeLoot moves something from a corpse to the player.
func (z *Zone) takeLoot(p *Player, id int64, index int) {
	n := z.byID[id]
	if n == nil || !n.dead() || n.gone {
		p.errMsg("loot", "There is nothing to loot.")
		return
	}
	if math.Hypot(n.Body.X-p.Body.X, n.Body.Y-p.Body.Y) > 450 {
		p.errMsg("loot", "You are too far away.")
		return
	}
	if n.lootOwner != p.ID && z.now-n.diedAt < LootExclusive {
		p.errMsg("loot", "That belongs to someone else.")
		return
	}
	if index < 0 {
		if n.money == 0 {
			return
		}
		got := n.money
		err := z.W.Store.Tx(func(tx *store.Tx) error {
			p.Char.Money += got
			return tx.SaveCharacter(p.Char)
		})
		if err != nil {
			p.Char.Money -= got
			p.errMsg("loot", "Could not save that; try again.")
			return
		}
		n.money = 0
		z.event(protocol.Event{K: "loot", To: p.ID, Amount: int(got), Text: "money"}, p.Body.X, p.Body.Y)
		z.sendCorpse(p, n)
		p.sendInv(z.W.Cat, p.Char.Money)
		return
	}
	if index >= len(n.loot) {
		return
	}
	l := n.loot[index]
	slot, err := z.W.addItem(p, l.ID, l.Count)
	if err != nil {
		p.errMsg("loot", err.Error())
		return
	}
	_ = slot
	n.loot = append(n.loot[:index], n.loot[index+1:]...)
	z.event(protocol.Event{K: "loot", To: p.ID, Text: z.W.Cat.Items[l.ID].Name}, p.Body.X, p.Body.Y)
	z.sendCorpse(p, n)
}
