package world

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// PetIDBase keeps pet entity ids apart from spawns, players and gravestones.
const PetIDBase = 1_500_000_000

// Pet orders.
const (
	modeFollow = "follow"
	modeStay   = "stay"
	modeAttack = "attack"
)

// petState is what makes a creature somebody's pet: a creature the player
// summoned, or one they charmed (which keeps its own spawn and goes back to it).
type petState struct {
	owner      *Player
	leader     *NPC   // a minion's commander
	minions    []*NPC // a commander's minions
	maxMinions int
	mode       string
	stance     string
	goal       *NPC
	charmed    bool
	released   bool
	beyondAt   float64 // when the owner first went beyond the leash; 0 when within it
	sent       string  // the last pet message sent, as text, to send only changes
	attacker   *NPC    // the creature that last struck it
}

// isPet says whether a creature belongs to a player.
func (n *NPC) isPet() bool { return n.pet != nil && !n.pet.released }

// ownerOf is the player a combatant fights for: a player themselves, a pet's owner.
func ownerOf(c combatant) *Player {
	switch t := c.(type) {
	case *Player:
		return t
	case *NPC:
		if t.isPet() {
			return t.pet.owner
		}
	}
	return nil
}

// castPetSpell handles the summon and charm spell types; it returns whether the
// type was one of them, and whether the spell took effect.
func (z *Zone) castPetSpell(caster, target combatant, sp *SpellDef) (handled, landed bool) {
	p, ok := caster.(*Player)
	switch sp.Type {
	case rules.SpellSummonSpirit, rules.SpellSummonHunterPet, rules.SpellSummonCommander, rules.SpellSummonMinion:
		if !ok {
			return true, false
		}
		return true, z.summonPet(p, sp)
	case rules.SpellCharm:
		n, isNPC := target.(*NPC)
		if !ok || !isNPC {
			return true, false
		}
		return true, z.charmCreature(p, n, sp)
	}
	return false, false
}

// petGear reads the equipment a pet template names: an equipment template id,
// or the inline form "slot:model|model;slot:model" (the first model of each).
func petGear(cat *Catalog, equip string) [][4]int {
	if equip == "" {
		return nil
	}
	if g, ok := cat.Gear[equip]; ok {
		return g
	}
	if !strings.Contains(equip, ":") {
		return nil
	}
	var out [][4]int
	for _, part := range strings.Split(equip, ";") {
		slot, models, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		s, err := strconv.Atoi(strings.TrimSpace(slot))
		if err != nil {
			continue
		}
		for _, m := range strings.Split(models, "|") {
			if v, err := strconv.Atoi(strings.TrimSpace(m)); err == nil && v > 0 {
				out = append(out, [4]int{s, v, 0, 0})
				break
			}
		}
	}
	return out
}

// summonPet makes a pet from the spell's template, beside the caster.
func (z *Zone) summonPet(p *Player, sp *SpellDef) bool {
	t := z.W.Cat.Templates[sp.LifeDrain]
	if t == nil {
		sys(p, fmt.Sprintf("The creature %s calls for is not in this world.", sp.Name))
		return false
	}
	var commander *NPC
	switch sp.Type {
	case rules.SpellSummonMinion:
		commander = p.pet
		if commander == nil || !commander.isPet() || commander.pet.maxMinions == 0 && len(commander.pet.minions) == 0 || commander.pet.leader != nil {
			sys(p, "You need a commander to summon this minion.")
			return false
		}
		if len(commander.pet.minions) >= commander.pet.maxMinions {
			sys(p, "Your commander cannot control another minion.")
			return false
		}
	default:
		if p.pet != nil {
			z.releasePet(p.pet, "You release your "+p.pet.title()+".")
		}
	}
	level := rules.PetLevel(sp.Damage, sp.Value, p.Char.Level)
	z.petSeq++
	spawn := Spawn{ID: PetIDBase + z.petSeq, Zone: z.ID, Name: t.Name, Model: t.Model, Size: t.Size, Speed: t.Speed, Template: t.ID,
		Level: level, Equip: t.Equip, Realm: 2}
	n := newNPC(spawn, z.W.Cat, z.rng)
	n.Gear = petGear(z.W.Cat, t.Equip)
	n.Friendly = true
	n.Aggro, n.Roam = 0, 0
	n.pet = &petState{owner: p, mode: modeFollow, stance: rules.StanceDefensive}
	if sp.Type == rules.SpellSummonCommander {
		n.pet.maxMinions = rules.MinionLimit(sp.Pulse)
	}
	x, y := p.Body.X+math.Sin(p.Body.Yaw)*64, p.Body.Y-math.Cos(p.Body.Yaw)*64
	x, y = z.G.Clamp(x, y)
	n.HX, n.HY = x, y
	n.Body = Body{X: x, Y: y, Z: z.G.GroundAt(x, y, p.Body.Z), Yaw: p.Body.Yaw + math.Pi}
	n.state = npcIdle
	n.HP = n.MaxHP
	z.npcs = append(z.npcs, n)
	z.byID[n.ID] = n
	if commander != nil {
		n.pet.leader = commander
		n.pet.stance, n.pet.mode = commander.pet.stance, commander.pet.mode
		commander.pet.minions = append(commander.pet.minions, n)
	} else {
		p.pet = n
	}
	z.event(protocol.Event{K: "effect", From: p.ID, To: n.ID, Text: sp.Name}, x, y)
	sys(p, fmt.Sprintf("The %s is now under your control.", n.title()))
	z.sendPet(p)
	return true
}

// charmCreature takes a creature as a pet if the spell may.
func (z *Zone) charmCreature(p *Player, n *NPC, sp *SpellDef) bool {
	if n.Friendly || n.dead() || n.isPet() || n.Guard {
		sys(p, "You cannot charm that.")
		return false
	}
	body := 0
	if t := z.W.Cat.Templates[n.Spawn.Template]; t != nil {
		body = t.BodyType
	}
	if !rules.CharmKindOK(sp.AmnesiaChance, body) {
		sys(p, "This spell does not charm that type of monster!")
		return false
	}
	if !rules.CharmLevelOK(sp.Value, p.Char.Level, n.Level) {
		sys(p, fmt.Sprintf("%s is too strong for you to charm!", n.title()))
		return false
	}
	if p.pet != nil {
		z.releasePet(p.pet, "You release your "+p.pet.title()+".")
	}
	n.pet = &petState{owner: p, mode: modeFollow, stance: rules.StanceDefensive, charmed: true}
	n.Friendly = true
	n.Aggro = 0
	n.hate, n.target = nil, nil
	n.state = npcIdle
	n.casting = nil
	p.pet = n
	sys(p, fmt.Sprintf("%s is now under your control.", n.title()))
	z.sendPet(p)
	return true
}

// releasePet lets a pet go: a summoned one vanishes, a charmed creature goes
// back to being a creature, hostile to the one who held it. A commander's
// minions go with it.
func (z *Zone) releasePet(n *NPC, text string) {
	ps := n.pet
	if ps == nil || ps.released {
		return
	}
	ps.released = true
	minions := ps.minions
	ps.minions = nil
	for _, m := range minions {
		z.releasePet(m, "")
	}
	if ps.leader != nil {
		for i, m := range ps.leader.pet.minions {
			if m == n {
				ps.leader.pet.minions = append(ps.leader.pet.minions[:i], ps.leader.pet.minions[i+1:]...)
				break
			}
		}
	}
	o := ps.owner
	if o != nil && o.pet == n {
		o.pet = nil
	}
	if o != nil && text != "" {
		sys(o, text)
	}
	n.casting = nil
	z.clearEffects(n)
	if ps.charmed {
		// A charmed creature is its own again: it was hostile all along.
		n.pet = nil
		n.Friendly = n.Spawn.Realm != 0
		n.Aggro = 0
		if o != nil && o.fightable(z) && !n.Friendly {
			n.engage(o, z)
		} else {
			n.goHome()
		}
	} else {
		n.state = npcDead
		n.gone = true
		n.target, n.hate = nil, nil
		n.speedNow = 0
		z.event(protocol.Event{K: "death", To: n.ID}, n.Body.X, n.Body.Y)
	}
	if o != nil {
		z.sendPet(o)
	}
}

// releaseOwned lets go of every pet a player has, for a death or a departure.
func (z *Zone) releaseOwned(p *Player, text string) {
	if p.pet != nil {
		z.releasePet(p.pet, text)
	}
}

// sweepPets takes released pets out of the zone.
func (z *Zone) sweepPets() {
	kept := z.npcs[:0]
	for _, n := range z.npcs {
		if n.pet != nil && n.pet.released && n.gone {
			delete(z.byID, n.ID)
			continue
		}
		kept = append(kept, n)
	}
	for i := len(kept); i < len(z.npcs); i++ {
		z.npcs[i] = nil
	}
	z.npcs = kept
}

// hostileTo says whether a pet may fight a creature.
func (z *Zone) hostileTo(o *NPC) bool {
	return !o.Friendly && !o.Guard && !o.isPet() && o.fightable(z)
}

// petUpdate runs one tick of a pet's life: lose it if it must go, choose
// whom to fight by the orders and stance, then fight or follow.
func (z *Zone) petUpdate(n *NPC) {
	ps := n.pet
	o := ps.owner
	if ps.released || n.dead() {
		return
	}
	if o == nil || o.closed.Load() || o.Zone != z || o.dead {
		reason := "Your pet is released."
		if o != nil && o.dead {
			reason = "Your " + n.title() + " is released as you fall."
		}
		z.releasePet(n, reason)
		return
	}
	// A leader the pet follows: a minion follows its commander.
	anchor := combatant(o)
	if ps.leader != nil {
		anchor = ps.leader
	}
	ax, ay := anchor.position()
	if math.Hypot(o.Body.X-n.Body.X, o.Body.Y-n.Body.Y) > rules.PetLeash {
		if ps.beyondAt == 0 {
			ps.beyondAt = z.now
		} else if z.now-ps.beyondAt > rules.PetLeashGrace {
			z.releasePet(n, "Your "+n.title()+" is lost, too far from you.")
			return
		}
	} else {
		ps.beyondAt = 0
	}
	if held(n) {
		n.speedNow = 0
		return
	}
	// Whom to fight.
	var goal *NPC
	switch {
	case ps.mode == modeAttack && ps.goal != nil && z.hostileTo(ps.goal):
		goal = ps.goal
	case ps.mode == modeAttack:
		ps.mode, ps.goal = modeFollow, nil
		z.sendPet(o)
	}
	if goal == nil && ps.leader != nil && ps.leader.pet.goal != nil && z.hostileTo(ps.leader.pet.goal) {
		goal = ps.leader.pet.goal
	}
	if goal == nil {
		goal = z.petFindGoal(n, ps)
	}
	if goal != nil {
		n.target = goal
		z.petFight(n, goal, ps.mode == modeStay)
		return
	}
	n.target = nil
	n.casting = nil
	if ps.mode == modeStay {
		n.speedNow = 0
		return
	}
	d := math.Hypot(ax-n.Body.X, ay-n.Body.Y)
	if d <= rules.PetFollowDist+30 {
		n.speedNow = 0
		return
	}
	speed := math.Max(n.RunSpd, 250)
	if d < 300 {
		speed = n.RunSpd * 0.7
	}
	n.stepToward(z, ax, ay, speed)
}

// petFindGoal is the creature a pet picks for itself by its stance: a
// defensive one takes whatever is fighting its owner or it, an aggressive one
// anything hostile near it.
func (z *Zone) petFindGoal(n *NPC, ps *petState) *NPC {
	if ps.stance == rules.StancePassive {
		return nil
	}
	if a := ps.attacker; a != nil && z.hostileTo(a) {
		return a
	}
	var best *NPC
	bd := math.MaxFloat64
	for _, o := range z.npcs {
		if !z.hostileTo(o) {
			continue
		}
		d := math.Hypot(o.Body.X-n.Body.X, o.Body.Y-n.Body.Y)
		if o.target == combatant(ps.owner) || o.target == combatant(n) {
			if d < 1500 && d < bd {
				best, bd = o, d
			}
			continue
		}
		if ps.stance == rules.StanceAggressive && d <= rules.PetAggroRange && best == nil && o.target == nil {
			best, bd = o, d
		}
	}
	return best
}

// petFight casts, closes on and strikes a pet's goal.
func (z *Zone) petFight(n *NPC, goal *NPC, hold bool) {
	if n.casting != nil {
		z.continueCasting(n)
		return
	}
	if z.startCasting(n) {
		return
	}
	tx, ty := goal.position()
	dx, dy := tx-n.Body.X, ty-n.Body.Y
	d := math.Hypot(dx, dy)
	reach := n.Reach + goal.Reach*0.3 + 20
	n.Body.Yaw = turnToward(n.Body.Yaw, math.Atan2(dx, -dy), 6*TickDT)
	if d > reach {
		if hold {
			n.speedNow = 0
			return
		}
		n.stepToward(z, tx, ty, math.Max(n.RunSpd, 250))
		return
	}
	n.speedNow = 0
	if z.now >= n.swingAt {
		n.swingAt = z.now + n.Swing
		z.npcSwing(n, goal)
	}
}

// PetCommand is the owner's order to their pet.
func (z *Zone) PetCommand(p *Player, cmd string) {
	n := p.pet
	if n == nil || !n.isPet() || n.pet.owner != p {
		sys(p, "You have no pet.")
		return
	}
	ps := n.pet
	apply := func(f func(*petState)) {
		f(ps)
		for _, m := range ps.minions {
			f(m.pet)
		}
	}
	switch cmd {
	case "release":
		z.releasePet(n, "You release your "+n.title()+".")
		return
	case "follow", "come":
		apply(func(s *petState) { s.mode, s.goal = modeFollow, nil })
	case "stay":
		apply(func(s *petState) { s.mode, s.goal = modeStay, nil })
	case "attack":
		t := z.byID[p.target]
		if t == nil || !z.hostileTo(t) {
			sys(p, "Your pet cannot attack that.")
			return
		}
		if dx, dy := t.Body.X-p.Body.X, t.Body.Y-p.Body.Y; math.Hypot(dx, dy) > 2500 {
			sys(p, "That is too far away.")
			return
		}
		apply(func(s *petState) { s.mode, s.goal = modeAttack, t })
		sys(p, fmt.Sprintf("Your %s attacks %s.", n.title(), t.title()))
	case rules.StancePassive, rules.StanceDefensive, rules.StanceAggressive:
		apply(func(s *petState) { s.stance = cmd })
	default:
		sys(p, "You can tell your pet: attack, follow, stay, release, passive, defensive or aggressive.")
		return
	}
	z.sendPet(p)
}

// petMsg is the pet message for a player: their pet as it is now, or closed.
func (z *Zone) petMsg(p *Player) protocol.Pet {
	n := p.pet
	if n == nil || !n.isPet() {
		return protocol.Pet{T: protocol.TPet, Closed: true}
	}
	m := protocol.Pet{T: protocol.TPet, ID: n.ID, Name: n.title(), Level: n.Level, HP: math.Max(0, float64(n.HP)/float64(max(1, n.MaxHP))),
		Stance: n.pet.stance, Mode: n.pet.mode}
	for _, mm := range n.pet.minions {
		m.Minions = append(m.Minions, protocol.PetMinion{ID: mm.ID, Name: mm.title(), Level: mm.Level, HP: math.Max(0, float64(mm.HP)/float64(max(1, mm.MaxHP)))})
	}
	return m
}

// sendPet sends a player their pet's state if it has changed.
func (z *Zone) sendPet(p *Player) {
	m := z.petMsg(p)
	key := fmt.Sprintf("%+v", m)
	if key == p.petSent {
		return
	}
	p.petSent = key
	p.send(m)
}

// tickPets keeps each owner's pet window current: health changes as they fight.
func (z *Zone) tickPets() {
	for _, p := range z.players {
		if p.pet != nil || p.petSent != "" {
			z.sendPet(p)
		}
	}
}
