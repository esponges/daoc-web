package world

import (
	"fmt"
	"math"
	"sort"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

// sheet is a player's numbers under the rules, derived from what they are
// (class, level, base stats, specs, constitution lost) and what they wear. It
// is recomputed whenever any of those change; nothing on it is saved.
type sheet struct {
	class      rules.Class
	stats      rules.Stats    // with level gains, capped items and constitution lost
	itemStats  rules.Stats    // uncapped, as worn
	itemHP     int            //
	skillBonus map[string]int // spec key -> item bonus, before the cap
	abilities  map[string]int
	maxHP      int
	buffAF     int // armour factor spell buffs add
	maxPower   int // 0 for a class that does not cast
	unspent    int
	half       bool
}

// classOf is the rules' class for a character; anything the rules do not
// cover yet plays as a Warrior, the class they do.
func classOf(class int) rules.Class {
	if c, ok := rules.Classes[class]; ok {
		return c
	}
	return rules.Classes[rules.ClassWarrior]
}

func raceOf(race int) rules.Race {
	if r, ok := rules.Races[race]; ok {
		return r
	}
	return rules.Races[5]
}

// oldXPToReach is the experience curve characters were saved on before the
// rule book; it is kept only to move them onto the real table.
func oldXPToReach(level int) int64 {
	if level <= 1 {
		return 0
	}
	n := float64(level - 1)
	return int64(40*n*n + 20*n)
}

// prepareCharacter brings a character saved before the rule book up to date:
// the default build for stats it never chose, and its experience moved to the
// same fraction of its level on the real table. It reports whether anything
// changed.
func prepareCharacter(c *store.Character) bool {
	changed := false
	if !c.HasStats {
		cl := classOf(c.Class)
		st, err := rules.Create(raceOf(c.Race), cl, rules.DefaultBuild(cl))
		if err != nil {
			st = raceOf(c.Race).Base
		}
		c.Stats, c.HasStats = st, true
		changed = true
	}
	if c.XPScheme != store.XPSchemeRules {
		c.Level = max(1, min(rules.MaxLevel, c.Level))
		if c.Level >= rules.MaxLevel {
			c.XP = rules.XPToReach(rules.MaxLevel)
		} else {
			lo, hi := oldXPToReach(c.Level), oldXPToReach(c.Level+1)
			frac := math.Max(0, math.Min(0.999999, float64(c.XP-lo)/float64(hi-lo)))
			nlo, nhi := rules.XPToReach(c.Level), rules.XPToReach(c.Level+1)
			c.XP = nlo + int64(frac*float64(nhi-nlo))
		}
		c.XPScheme = store.XPSchemeRules
		changed = true
	}
	if c.Specs == nil {
		c.Specs = map[string]int{}
	}
	return changed
}

// specLevel is a trained spec's level; untrained specs are 1.
func (p *Player) specLevel(key string) int { return max(1, p.Char.Specs[key]) }

// weaponSpec is the spec that governs a weapon object type.
func weaponSpec(objectType int) string {
	switch objectType {
	case rules.ObjSword:
		return "Sword"
	case rules.ObjHammer:
		return "Hammer"
	case rules.ObjAxe:
		return "Axe"
	case rules.ObjLeftAxe:
		return "Left Axe"
	case rules.ObjSpear:
		return "Spear"
	case rules.ObjThrown:
		return "Thrown Weapons"
	case rules.ObjStaff:
		return "Staff"
	case ObjShield:
		return "Shields"
	}
	return ""
}

// skillBonusSpecs are the specs each item skill bonus type adds to.
var skillBonusSpecs = map[int][]string{
	40: {"Parry"}, 43: {"Shields"}, 52: {"Sword"}, 53: {"Hammer"}, 54: {"Axe"}, 55: {"Left Axe"}, 91: {"Thrown Weapons"},
	164: {"Sword", "Hammer", "Axe", "Left Axe", "Spear", "Thrown Weapons", "Staff"}, // all melee weapon skills
}

// statBonus maps the item table's stat bonus types to the rules' stats.
var statBonus = map[int]rules.Stat{1: rules.STR, 2: rules.DEX, 3: rules.CON, 4: rules.QUI, 5: rules.INT, 6: rules.PIE, 7: rules.EMP, 8: rules.CHA}

// combatSpec is a spec as combat counts it: trained plus capped item bonus.
func (p *Player) combatSpec(key string) int {
	return rules.CombatSpec(p.specLevel(key), p.sh.skillBonus[key], p.Char.Level)
}

// trainedSpent is the points spent on a character's specs.
func (w *World) trainedSpent(c *store.Character) int {
	spent := 0
	for _, k := range w.Cat.TrainableSpecs(c.Class) {
		spent += rules.TrainCost(1, max(1, c.Specs[k]))
	}
	return spent
}

// recalc rebuilds what a player holds and wears and their sheet, and sends
// the sheet when it changed.
func (w *World) recalc(p *Player) {
	w.recalcGear(p)
	c := p.Char
	sh := sheet{class: classOf(c.Class), skillBonus: map[string]int{}}
	for s := SlotWeapon; s < SlotCount; s++ {
		it := p.inv[s]
		if it == nil {
			continue
		}
		x := w.Cat.Items[it.ID]
		if x == nil {
			continue
		}
		for _, b := range x.Bonuses {
			if st, ok := statBonus[b.Type]; ok {
				sh.itemStats[st] += b.Amount
			} else if b.Type == 10 {
				sh.itemHP += b.Amount
			} else {
				for _, k := range skillBonusSpecs[b.Type] {
					sh.skillBonus[k] += b.Amount
				}
			}
		}
	}
	sh.stats = rules.Modified(c.Stats, sh.class, c.Level, sh.itemStats, c.ConLost)
	buffs, buffAF := p.buffBonuses()
	for s := range sh.stats {
		sh.stats[s] += min(buffs[s], rules.BuffStatCap(c.Level))
	}
	sh.buffAF = min(buffAF, rules.BuffAFCap(c.Level))
	sh.maxHP = rules.MaxHP(sh.class, c.Level, sh.stats[rules.CON], sh.itemHP)
	sh.maxPower = rules.MaxPower(sh.class, c.Level, sh.stats[sh.class.Acuity])
	sh.half = c.Level >= 40 && c.XP >= rules.HalfLevel(c.Level)
	sh.unspent = rules.SpecPoints(sh.class, c.Level, sh.half) - w.trainedSpent(c)
	trained := map[string]int{}
	for _, k := range w.Cat.TrainableSpecs(c.Class) {
		trained[k] = p.specLevel(k)
	}
	sh.abilities = rules.Abilities(w.Cat.Grants[c.Class], w.Cat.Careers(c.Class), c.Level, trained)
	p.sh = sh
	w.refreshBook(p)
	if p.hp > float64(sh.maxHP) {
		p.hp = float64(sh.maxHP)
	}
	if p.power > float64(sh.maxPower) || !p.powerSet {
		// A character who has not been given a pool yet starts with a full one.
		p.power, p.powerSet = float64(sh.maxPower), true
	}
	if !p.welcomed {
		return // join sends the sheet after the welcome
	}
	if m := p.statsMsg(w.Cat); !sameStats(m, p.statsSent) {
		p.statsSent = m
		p.send(m)
	}
}

// statsMsg is the character sheet as the protocol carries it.
func (p *Player) statsMsg(cat *Catalog) protocol.Stats {
	c, sh := p.Char, p.sh
	m := protocol.Stats{T: protocol.TStats, Class: sh.class.Name, Race: cat.RaceNames[c.Race], Level: c.Level, Half: sh.half, Base: c.Stats, Stats: sh.stats,
		MaxHP: sh.maxHP, MaxEnd: rules.MaxEndurance, MaxPower: sh.maxPower, Unspent: sh.unspent, ConLost: c.ConLost}
	for _, k := range cat.TrainableSpecs(c.Class) {
		m.Specs = append(m.Specs, protocol.SpecRow{Key: k, Name: cat.SpecName(k), Level: p.specLevel(k),
			Bonus: p.combatSpec(k) - p.specLevel(k)})
	}
	for k, lv := range sh.abilities {
		m.Abilities = append(m.Abilities, protocol.Ability{Name: k, Level: lv, On: rules.ActiveAbility(k)})
	}
	sort.Slice(m.Abilities, func(i, j int) bool { return m.Abilities[i].Name < m.Abilities[j].Name })
	if m.Specs == nil {
		m.Specs = []protocol.SpecRow{}
	}
	if m.Abilities == nil {
		m.Abilities = []protocol.Ability{}
	}
	cb := p.combatant(cat)
	m.WS = int(math.Round(rules.WeaponSkill(cb)))
	af := 0.0
	for l := rules.Location(0); l < rules.NumLocations; l++ {
		a, _ := rules.LocationArmour(cb, l)
		af += a
	}
	m.AF = int(af / float64(rules.NumLocations))
	return m
}

func sameStats(a, b protocol.Stats) bool { return fmt.Sprint(a) == fmt.Sprint(b) }

// itemFraction turns a quality percent into a multiplier; unknown counts as whole.
func itemFraction(pct int) float64 {
	if pct <= 0 || pct > 100 {
		return 1
	}
	return float64(pct) / 100
}

// condition is an item's condition as a multiplier; an item that does not
// say counts as whole.
func condition(c float64) float64 {
	if c <= 0 || c > 1 {
		return 1
	}
	return c
}

// unarmed is what a player without a weapon strikes with.
var unarmed = rules.Weapon{DPS: 1.2, Speed: 2.2, Quality: 1, Condition: 1, DamageType: rules.DamageCrush}

// weaponOf describes an item for a swing.
func weaponOf(it *Item) rules.Weapon {
	if it == nil {
		return unarmed
	}
	return rules.Weapon{DPS: it.DPS, Speed: it.Speed, ObjectType: it.ObjectType, TwoHanded: it.TwoHanded(),
		Quality: itemFraction(it.Quality), Condition: condition(it.Condition), DamageType: it.DamageType, Bonus: it.Bonus}
}

// locationSlot is the worn slot that covers each place a blow lands.
var locationSlot = [rules.NumLocations]int{SlotTorso, SlotLegs, SlotArms, SlotHead, SlotHands, SlotFeet}

// combatant describes the player for a swing.
func (p *Player) combatant(cat *Catalog) *rules.Combatant {
	sh := p.sh
	cb := &rules.Combatant{Player: true, Level: p.Char.Level, Stats: sh.stats, WeaponSkillBase: sh.class.WeaponSkillBase,
		Weapon: weaponOf(p.weapon), EvadeLevel: sh.abilities[rules.AbEvade], StyleDefence: p.styleDefence, BuffAF: float64(sh.buffAF)}
	if p.weapon != nil {
		cb.Weapon.Condition = p.condOf(SlotWeapon)
	}
	if p.weapon != nil {
		cb.Spec = p.combatSpec(weaponSpec(p.weapon.ObjectType))
	} else {
		cb.Unarmed = true
	}
	if p.hasSpec(cat, "Parry") {
		cb.ParrySpec = p.combatSpec("Parry")
	}
	if s := p.shield; s != nil && sh.abilities[rules.AbShield] > 0 {
		cb.ShieldSpec = p.combatSpec("Shields")
		cb.Shield = &rules.Shield{Size: max(1, s.DamageType), Quality: itemFraction(s.Quality), Condition: p.condOf(SlotShield)}
	}
	for l, slot := range locationSlot {
		if it := p.inv[slot]; it != nil {
			if x := cat.Items[it.ID]; x != nil && x.IsArmour() {
				cb.Armour[l] = &rules.Piece{AF: float64(x.AF), Absorb: float64(x.Absorb) / 100, Cloth: x.ObjectType == rules.ObjCloth,
					Quality: itemFraction(x.Quality), Condition: p.condOf(slot), Bonus: x.Bonus}
			}
		}
	}
	return cb
}

// hasSpec says whether the player's class has a spec.
func (p *Player) hasSpec(cat *Catalog, key string) bool {
	for _, k := range cat.TrainableSpecs(p.Char.Class) {
		if k == key {
			return true
		}
	}
	return false
}
