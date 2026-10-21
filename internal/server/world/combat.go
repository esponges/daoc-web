package world

import (
	"fmt"
	"math"
	"sort"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
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

// swingTime is how long a player's swing takes: the weapon's speed, quicker
// with quickness.
func (p *Player) swingTime() float64 {
	speed := unarmed.Speed
	if p.weapon != nil && p.weapon.Speed > 0 {
		speed = p.weapon.Speed
	}
	return rules.SwingTime(speed, p.sh.stats[rules.QUI])
}

// weaponSpeed is the held weapon's speed, for style costs.
func (p *Player) weaponSpeed() float64 {
	if p.weapon != nil && p.weapon.Speed > 0 {
		return p.weapon.Speed
	}
	return unarmed.Speed
}

// styleCost is a style's endurance with the weapon held.
func (p *Player) styleCost(s *StyleDef) int { return rules.StyleCost(p.weaponSpeed(), s.Endurance) }

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

// computeBar chooses the bar without sending it. Styles the player has already
// placed keep their order; new ones follow in the default order. The styles
// they have learned but cannot use with what they hold are kept in p.barAlso.
func (z *Zone) computeBar(p *Player) {
	prev := make([]int, len(p.bar))
	for i, s := range p.bar {
		prev[i] = s.ID
	}
	p.bar, p.barAlso = p.bar[:0], p.barAlso[:0]
	if z.W.Cat == nil {
		return
	}
	for _, s := range z.W.Cat.StylesFor(p.Char.Class) {
		if s.SpecLevel > p.specLevel(s.Spec) {
			continue
		}
		switch {
		case s.WeaponReq == ObjShield:
			if p.shield == nil {
				p.barAlso = append(p.barAlso, s)
				continue
			}
		case p.weapon == nil || p.weapon.ObjectType != s.WeaponReq:
			p.barAlso = append(p.barAlso, s)
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
	// The player's own order first.
	pos := map[int]int{}
	for i, id := range prev {
		pos[id] = i
	}
	sort.SliceStable(p.bar, func(i, j int) bool {
		pi, iok := pos[p.bar[i].ID]
		pj, jok := pos[p.bar[j].ID]
		if iok && jok {
			return pi < pj
		}
		return iok && !jok
	})
}

// SetBarSlot puts a style the player has on their bar into a slot, swapping it
// with the style that was there, so the key for that slot uses it.
func (z *Zone) SetBarSlot(p *Player, slot, id int) {
	from := -1
	for i := range p.bar {
		if p.bar[i].ID == id {
			from = i
		}
	}
	if from < 0 {
		p.errMsg("bar", "You cannot use that style with what you hold.")
		return
	}
	if slot < 0 || slot >= len(p.bar) {
		p.errMsg("bar", "There is no such slot on your bar.")
		return
	}
	p.pending = -1
	p.bar[from], p.bar[slot] = p.bar[slot], p.bar[from]
	p.sendBar(z.W.Cat)
}

func (p *Player) sendBar(c *Catalog) {
	b := protocol.Bar{T: protocol.TStyleBar}
	for i := range p.bar {
		s := &p.bar[i]
		b.Slots = append(b.Slots, protocol.BarSlot{Slot: i, ID: s.ID, Name: s.Name, Icon: s.Icon, Spec: s.Spec, Cost: p.styleCost(s),
			Opens: c.openingText(s), Level: s.SpecLevel})
	}
	for i := range p.barAlso {
		s := &p.barAlso[i]
		b.Also = append(b.Also, protocol.BarSlot{Slot: -1, ID: s.ID, Name: s.Name, Icon: s.Icon, Spec: s.Spec, Cost: p.styleCost(s),
			Opens: c.openingText(s), Level: s.SpecLevel, Locked: true})
	}
	p.send(b)
}

// styleReady says whether a style can be used on the current target now.
func (z *Zone) styleReady(p *Player, s *StyleDef) (bool, string) {
	if p.end < float64(p.styleCost(s)) {
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
	if held(p) {
		p.errMsg("style", "You cannot act while you are held.")
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
	if id != p.ab.engage {
		p.ab.engage = 0
	}
	if id == 0 {
		p.target, p.attacking, p.pending = 0, false, -1
		return
	}
	if id >= GraveIDBase {
		if g := z.graves[id-GraveIDBase]; g != nil {
			p.target, p.attacking, p.pending = id, false, -1
		}
		return
	}
	// Another player can be targeted, to invite them; not to fight.
	if o := z.players[id]; o != nil && o != p {
		p.target, p.attacking, p.pending = id, false, -1
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
	if p.ride != nil {
		p.errMsg("target", "You cannot fight while riding.")
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
	p.attacking, p.ab.engage = true, 0
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
	p.swingAt = z.now + p.swingTime()
	z.playerSwing(p, n)
}

func (z *Zone) playerSwing(p *Player, n *NPC) {
	z.interruptCasting(n) // any attack, hit or not
	var style *StyleDef
	if p.pending >= 0 && p.pending < len(p.bar) {
		def := p.bar[p.pending]
		if ok, _ := z.styleReady(p, &def); ok {
			style = &def
		}
	}
	p.pending = -1
	p.combatUntil = z.now + CombatLinger
	p.styleDefence = 0 // a style's defence bonus lasts until the next swing
	ev := protocol.Event{K: "swing", From: p.ID, To: n.ID}
	if style != nil {
		ev.Style = style.Anim
		ev.Text = style.Name
	}
	z.event(ev, p.Body.X, p.Body.Y)

	sw := rules.Swing{Attacker: p.combatant(z.W.Cat), Defender: &n.cb, Attackers: 1,
		InFront: angleOff(n.Body.Yaw, angleTo(n.Body.X, n.Body.Y, p.Body.X, p.Body.Y)) <= frontArc}
	if style != nil {
		sw.Style = &rules.Style{GrowthRate: style.GrowthRate, Spec: p.combatSpec(style.Spec), BonusToHit: style.BonusToHit, Location: rules.LocNone}
	}
	o := rules.Resolve(sw, z.rng)
	if o.Result != rules.Hit {
		if style != nil {
			p.end = math.Max(0, p.end-float64(rules.FailedStyleCost(p.weaponSpeed(), style.Endurance)))
		}
		if o.Result == rules.Missed {
			z.event(protocol.Event{K: "miss", From: p.ID, To: n.ID}, n.Body.X, n.Body.Y)
		} else {
			z.event(protocol.Event{K: o.Result.String(), From: n.ID, To: p.ID}, n.Body.X, n.Body.Y)
		}
		p.lastHit, p.lastHitAt, p.lastStyle = false, z.now, 0
		z.provoke(n, p, 0)
		return
	}
	if style != nil {
		p.end = math.Max(0, p.end-float64(p.styleCost(style)))
		p.lastStyle, p.lastStyleAt = style.ID, z.now
		p.styleDefence = style.BonusToDef
	} else {
		p.lastStyle = 0
	}
	z.wearItem(p, SlotWeapon)
	amount := o.Damage
	p.lastHit, p.lastHitAt = true, z.now
	z.event(protocol.Event{K: "hit", From: p.ID, To: n.ID, Amount: amount, Crit: o.Critical > 0, Text: ev.Text}, n.Body.X, n.Body.Y)
	n.HP -= amount
	n.hurtBy[p.ID] += amount
	z.damaged(n)
	if n.tagger == 0 {
		n.tagger = p.ID
	}
	z.provoke(n, p, float64(amount))
	if n.HP <= 0 {
		z.killNPC(n)
		return
	}
	if add := p.damageAdd(); add > 0 {
		z.event(protocol.Event{K: "spell", From: p.ID, To: n.ID, Text: "damage add", Amount: add}, n.Body.X, n.Body.Y)
		z.hurt(n, p, add)
		if n.dead() {
			return
		}
	}
	z.playerProcs(p, n, style)
}

// provoke makes a creature hate the one who struck at it by the damage done,
// and fight them if it was not fighting already.
func (z *Zone) provoke(n *NPC, by combatant, damage float64) {
	if n.dead() || n.Friendly && !n.Guard || n.state == npcReturn {
		return
	}
	n.attackedAt = z.now
	if p, ok := by.(*Player); ok && damage > 0 {
		damage = z.protectShift(n, p, damage)
	}
	n.addHate(by, damage)
	n.engage(by, z)
}

// attackersOf counts the creatures fighting a player within reach of them.
func (z *Zone) attackersOf(p *Player) int {
	n := 0
	for _, c := range z.npcs {
		if c.target == p && !c.dead() && math.Hypot(c.Body.X-p.Body.X, c.Body.Y-p.Body.Y) <= c.Reach+60 {
			n++
		}
	}
	return max(1, n)
}

// npcSwing resolves a creature's blow against whatever it fights.
func (z *Zone) npcSwing(n *NPC, c combatant) {
	switch t := c.(type) {
	case *Player:
		z.npcHitsPlayer(n, t)
	case *NPC:
		z.npcHitsNPC(n, t)
	}
}

// npcHitsNPC resolves one creature's blow against another, which then hates
// the striker and fights back.
func (z *Zone) npcHitsNPC(n, t *NPC) {
	z.interruptCasting(t)
	z.event(protocol.Event{K: "swing", From: n.ID, To: t.ID}, n.Body.X, n.Body.Y)
	sw := rules.Swing{Attacker: &n.cb, Defender: &t.cb, Attackers: 1,
		InFront: angleOff(t.Body.Yaw, angleTo(t.Body.X, t.Body.Y, n.Body.X, n.Body.Y)) <= frontArc}
	o := rules.Resolve(sw, z.rng)
	switch o.Result {
	case rules.Missed:
		z.event(protocol.Event{K: "miss", From: n.ID, To: t.ID}, t.Body.X, t.Body.Y)
		z.provoke(t, n, 0)
		return
	case rules.Evaded, rules.Parried, rules.Blocked:
		z.event(protocol.Event{K: o.Result.String(), From: t.ID, To: n.ID}, t.Body.X, t.Body.Y)
		z.provoke(t, n, 0)
		return
	}
	z.event(protocol.Event{K: "hit", From: n.ID, To: t.ID, Amount: o.Damage}, t.Body.X, t.Body.Y)
	t.HP -= o.Damage
	unsharedDamage(n, t, o.Damage)
	z.damaged(t)
	z.provoke(t, n, float64(o.Damage))
	if t.HP <= 0 {
		z.killNPC(t)
	}
	if !t.dead() {
		z.creatureProcs(n, t)
	}
}

// npcHitsPlayer resolves a creature's blow against a player.
func (z *Zone) npcHitsPlayer(n *NPC, p *Player) {
	z.event(protocol.Event{K: "swing", From: n.ID, To: p.ID}, n.Body.X, n.Body.Y)
	member := p
	w := z.interceptFor(p)
	if w != nil {
		p = w
	}
	p.combatUntil = z.now + CombatLinger
	cb := p.combatant(z.W.Cat)
	inFront := angleOff(p.Body.Yaw, angleTo(p.Body.X, p.Body.Y, n.Body.X, n.Body.Y)) <= frontArc
	if w == nil {
		z.engageBoost(n, p, cb)
	}
	sw := rules.Swing{Attacker: &n.cb, Defender: cb, Attackers: z.attackersOf(p), InFront: inFront, NoDefence: w != nil}
	o := rules.Resolve(sw, z.rng)
	if w == nil && (o.Result == rules.Hit || o.Result == rules.Missed) && z.guardBlocks(n, member) {
		return
	}
	z.wearBlow(p, o)
	switch o.Result {
	case rules.Missed:
		z.event(protocol.Event{K: "miss", From: n.ID, To: p.ID}, p.Body.X, p.Body.Y)
		return
	case rules.Evaded, rules.Parried, rules.Blocked:
		kind := o.Result.String()
		z.event(protocol.Event{K: kind, From: p.ID, To: n.ID}, p.Body.X, p.Body.Y)
		p.lastDef = defence{kind: kind, at: z.now}
		return
	}
	z.event(protocol.Event{K: "hit", From: n.ID, To: p.ID, Amount: o.Damage}, p.Body.X, p.Body.Y)
	p.hp -= float64(o.Damage)
	z.damaged(p)
	if p.hp <= 0 {
		z.killPlayer(p, n)
	}
	if !p.dead {
		z.creatureProcs(n, p)
	}
}

func ifSeen(seen bool, v float64) float64 {
	if seen {
		return v
	}
	return 0
}

// killPlayer ends a player's life until they return to their bind point.
func (z *Zone) killPlayer(p *Player, by *NPC) {
	z.closeShop(p)
	z.cancelPlayerCast(p, "")
	p.hp = 0
	p.dead = true
	p.deadAt = z.now
	p.attacking, p.target, p.pending = false, 0, -1
	p.queue, p.queuedSteps = nil, 0
	p.prayUntil = 0
	z.clearEffects(p)
	z.event(protocol.Event{K: "death", To: p.ID}, p.Body.X, p.Body.Y)
	for _, n := range z.npcs {
		n.forget(p)
	}
	z.deathPenalty(p, by)
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
	z.releasePenalty(p)
	bz, x, y, zz, yaw := z.W.bindFor(p)
	p.dead = false
	p.hp = float64(p.maxHP())
	p.end = float64(rules.MaxEndurance)
	p.combatUntil = 0
	z.W.savePlayer(p)
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
	n.target, n.hate = nil, nil
	z.clearEffects(n)
	n.speedNow = 0
	z.event(protocol.Event{K: "death", To: n.ID}, n.Body.X, n.Body.Y)

	total := 0
	for _, d := range n.hurtBy {
		total += d
	}
	// A creature grey to anyone who fought it is worth nothing to anyone.
	grey := false
	for id := range n.hurtBy {
		if p := z.players[id]; p != nil && rules.Grey(p.Char.Level, n.Level) {
			grey = true
		}
	}
	var top *Player
	best := 0
	// A group's members count as one damager; the group's share is split
	// among those in range by the group rule.
	groupDamage := map[int64]int{} // by leader
	groupMembers := map[int64][]int64{}
	for _, id := range sortedIDs(n.hurtBy) {
		p := z.players[id]
		if p == nil || p.closed.Load() {
			continue
		}
		if n.hurtBy[id] > best {
			top, best = p, n.hurtBy[id]
		}
		if p.target == n.ID {
			p.attacking, p.pending = false, -1
		}
		if members := z.W.groupOf(p.Char.ID); members != nil {
			groupDamage[members[0]] += n.hurtBy[id]
			groupMembers[members[0]] = members
			continue
		}
		share := float64(n.hurtBy[id]) / math.Max(1, float64(total))
		if xp := rules.KillXP(n.Level, p.Char.Level, share); xp > 0 && !grey {
			z.award(p, xp)
		}
	}
	for _, leader := range sortedIDs(groupDamage) {
		var near []*Player
		var levels []int
		for _, c := range groupMembers[leader] {
			p := z.players[PlayerIDBase+c]
			if p == nil || p.closed.Load() || p.dead || math.Hypot(p.Body.X-n.Body.X, p.Body.Y-n.Body.Y) > GroupRange {
				continue
			}
			near, levels = append(near, p), append(levels, p.Char.Level)
		}
		share := float64(groupDamage[leader]) / math.Max(1, float64(total))
		for i, xp := range rules.GroupKillXP(n.Level, share, levels) {
			if xp > 0 && !grey {
				z.award(near[i], xp)
			}
		}
	}
	for _, p := range z.players {
		if p.target == n.ID && p.attacking {
			p.attacking = false
		}
	}
	if top != nil {
		n.lootOwner = top.ID
		n.money = int64(float64(rules.KillMoney(n.Level)) * (0.7 + 0.6*z.rng.Float64()))
		n.loot = z.rollLoot(n)
		if len(n.loot) > 0 || n.money > 0 {
			z.sendCorpse(top, n)
		}
	}
}

// award gives experience and raises the level when it is earned.
func (z *Zone) award(p *Player, xp int64) {
	c := p.Char
	if c.Level >= MaxLevel {
		return
	}
	wasHalf := p.sh.half
	c.XP += xp
	z.event(protocol.Event{K: "xp", To: p.ID, Amount: int(xp)}, p.Body.X, p.Body.Y)
	levelled := false
	for c.Level < MaxLevel && c.XP >= rules.XPToReach(c.Level+1) {
		z.levelUp(p)
		levelled = true
	}
	if c.Level >= MaxLevel {
		c.XP = rules.XPToReach(MaxLevel)
	}
	z.W.recalc(p)
	if !levelled && !wasHalf && p.sh.half {
		pts := rules.SpecPoints(p.sh.class, c.Level, true) - rules.SpecPoints(p.sh.class, c.Level, false)
		z.event(protocol.Event{K: "level", To: p.ID, Amount: c.Level,
			Text: fmt.Sprintf("You have reached level %d.5! You gain %d specialization points.", c.Level, pts)}, p.Body.X, p.Body.Y)
	}
	z.W.savePlayer(p)
}

// levelUp raises a character one level and announces what it brought.
func (z *Zone) levelUp(p *Player) {
	c := p.Char
	before, hpBefore, ptsBefore := p.sh.stats, p.sh.maxHP, p.sh.unspent
	c.Level++
	c.DeathCount = 0
	z.W.recalc(p)
	p.hp = float64(p.maxHP())
	p.end = float64(rules.MaxEndurance)
	text := fmt.Sprintf("You have reached level %d!", c.Level)
	for st := rules.Stat(0); st < rules.NumStats; st++ {
		if d := p.sh.stats[st] - before[st]; d > 0 {
			text += fmt.Sprintf(" Your %s rises by %d.", st, d)
		}
	}
	if d := p.sh.maxHP - hpBefore; d > 0 {
		text += fmt.Sprintf(" You gain %d hit points.", d)
	}
	if d := p.sh.unspent - ptsBefore; d > 0 {
		text += fmt.Sprintf(" You gain %d specialization points.", d)
	}
	z.event(protocol.Event{K: "level", To: p.ID, Amount: c.Level, Text: text}, p.Body.X, p.Body.Y)
	z.buildBar(p)
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
	if len(out) == 0 && z.rng.Intn(100) < rules.GenericDropPercent(n.Level) {
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
