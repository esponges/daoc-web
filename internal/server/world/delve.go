package world

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
)

// Delve answers a right-click: the details of an item, spell or style the
// player can see. The server only describes what the player holds, is shown in
// the open shop or on a corpse in reach, or has in their book.
func (z *Zone) Delve(p *Player, m *protocol.Delve) {
	cat := z.W.Cat
	var out *protocol.DelveInfo
	switch m.From {
	case "inv":
		it := p.inv[m.Slot]
		if it == nil && inVault(m.Slot) {
			it = p.vault[m.Slot]
		}
		if it != nil {
			if t := cat.Items[it.ID]; t != nil {
				out = itemInfo(cat, t, it)
			}
		}
	case "shop":
		if n := z.byID[p.shop]; n != nil && p.shop != 0 {
			for _, id := range cat.Stock[n.Spawn.ItemsList] {
				if id == m.ID && cat.Items[id] != nil {
					out = itemInfo(cat, cat.Items[id], nil)
				}
			}
		}
	case "loot":
		n := z.byID[m.Corpse]
		if n != nil && n.dead() && !n.gone && m.Slot >= 0 && m.Slot < len(n.loot) && n.inReach(p) {
			if t := cat.Items[n.loot[m.Slot].ID]; t != nil {
				out = itemInfo(cat, t, nil)
			}
		}
	case "spell":
		id, _ := strconv.Atoi(m.ID)
		for _, b := range p.book {
			if b.Spell.ID == id {
				out = spellInfo(p, b.Spell, b.Level)
				break
			}
		}
	case "style":
		id, _ := strconv.Atoi(m.ID)
		if s := cat.StyleByID(p.Char.Class, id); s != nil {
			out = styleInfo(s)
		}
	}
	if out == nil {
		p.errMsg("delve", "There is nothing to see there.")
		return
	}
	out.T = protocol.TDelveInfo
	p.send(*out)
}

// damageName names a damage type as the item and spell tables number it.
func damageName(t int) string {
	switch t {
	case rules.DamageCrush:
		return "crush"
	case rules.DamageSlash:
		return "slash"
	case rules.DamageThrus:
		return "thrust"
	case rules.DamageBody:
		return "body"
	case rules.DamageCold:
		return "cold"
	case rules.DamageEnergy:
		return "energy"
	case rules.DamageHeat:
		return "heat"
	case rules.DamageMatter:
		return "matter"
	case rules.DamageSpirit:
		return "spirit"
	}
	return ""
}

// armourKind names an armour piece's material by its object type.
var armourKind = map[int]string{32: "Cloth", 33: "Leather", 34: "Studded", 35: "Chain", 36: "Plate", 37: "Reinforced", 38: "Scale"}

// bonusLines words the stat, skill and resist bonuses an item gives. A bonus
// type the table has that this server does not use is left out rather than guessed.
func bonusLines(it *Item) []string {
	resistOf := map[int]int{} // item bonus type -> damage type
	for dmg, bonus := range resistBonus {
		resistOf[bonus] = dmg
	}
	var out []string
	for _, b := range it.Bonuses {
		_, isStat := statBonus[b.Type]
		switch {
		case isStat:
			name := statBonus[b.Type].String()
			out = append(out, fmt.Sprintf("%s: +%d", strings.ToUpper(name[:1])+name[1:], b.Amount))
		case resistOf[b.Type] != 0 && damageName(resistOf[b.Type]) != "":
			out = append(out, fmt.Sprintf("%s resist: +%d%%", damageName(resistOf[b.Type]), b.Amount))
		case len(skillBonusSpecs[b.Type]) == 1:
			out = append(out, fmt.Sprintf("%s: +%d", skillBonusSpecs[b.Type][0], b.Amount))
		case len(skillBonusSpecs[b.Type]) > 1:
			out = append(out, fmt.Sprintf("All melee weapon skills: +%d", b.Amount))
		}
	}
	sort.Strings(out)
	return out
}

// itemInfo describes an item template, and the instance a player holds when
// there is one: its condition, its maker, its poison.
func itemInfo(cat *Catalog, it *Item, held *invItem) *protocol.DelveInfo {
	v := itemView(0, it, 1)
	d := &protocol.DelveInfo{Name: it.Name, Level: it.Level, Con: v.Kind == "weapon" || v.Kind == "shield" || v.Kind == "armour", Lines: []string{}}
	add := func(f string, a ...interface{}) { d.Lines = append(d.Lines, fmt.Sprintf(f, a...)) }
	if it.Level > 0 {
		add("Level %d", it.Level)
	}
	switch v.Kind {
	case "weapon":
		if sp := weaponSpec(it.ObjectType); sp != "" {
			add("Weapon: %s", sp)
		}
		hands := "one-handed"
		switch {
		case it.TwoHanded():
			hands = "two-handed"
		case it.Hand == rules.HandLeft:
			hands = "left-hand"
		}
		add("Damage: %.1f DPS, %.1f s swing, %s", it.DPS, it.Speed, hands)
		if n := damageName(it.DamageType); n != "" {
			add("Damage type: %s", n)
		}
	case "shield":
		add("Shield")
	case "armour":
		if k := armourKind[it.ObjectType]; k != "" {
			add("Armour: %s", k)
		}
		add("Armour factor: %d", it.AF)
		if it.Absorb > 0 {
			add("Absorption: %d%%", it.Absorb)
		}
	case "poison":
		add("A poison for a weapon")
		if it.PoisonCharges > 0 {
			add("Lasts %d hits", it.PoisonCharges)
		}
	}
	q := it.Quality
	if held != nil && held.Quality > 0 {
		q = held.Quality
	}
	if q > 0 && repairable(it) {
		add("Quality: %d%%", q)
	}
	if held != nil && repairable(it) {
		add("Condition: %.0f%%", math.Floor(float64(held.Cond)/100))
	}
	if held != nil && held.Crafter != "" {
		add("Made by %s", held.Crafter)
	}
	if held != nil && held.Poison != 0 {
		name := "poison"
		if sp := cat.Spells[held.Poison]; sp != nil {
			name = sp.Name
		}
		add("Poisoned with %s (%d hits left)", name, held.Charges)
	}
	d.Lines = append(d.Lines, bonusLines(it)...)
	if it.Price > 0 {
		add("Worth: %s", moneyText(it.Price))
	}
	return d
}

// spellInfo describes a spell in the player's book.
func spellInfo(p *Player, sp *SpellDef, level int) *protocol.DelveInfo {
	d := &protocol.DelveInfo{Name: sp.Name, Level: level, Lines: []string{}}
	add := func(f string, a ...interface{}) { d.Lines = append(d.Lines, fmt.Sprintf(f, a...)) }
	if level > 0 {
		add("Level %d", level)
	}
	if sp.Target != "" {
		add("Target: %s", sp.Target)
	}
	if sp.Range > 0 {
		add("Range: %d", int(sp.Range))
	}
	if sp.Radius > 0 {
		add("Radius: %d", sp.Radius)
	}
	if sp.Damage > 0 {
		if n := damageName(sp.DamageType); n != "" {
			add("Damage: %.0f %s", sp.Damage, n)
		} else {
			add("Damage: %.0f", sp.Damage)
		}
	} else if sp.Value > 0 {
		add("Effect: %s, %.0f", sp.Type, sp.Value)
	} else if sp.Type != "" {
		add("Effect: %s", sp.Type)
	}
	if sp.CastTime > 0 {
		add("Cast time: %.1f s", sp.CastTime)
	} else {
		add("Cast time: instant")
	}
	if sp.Recast > 0 {
		add("Recast: %.0f s", sp.Recast)
	}
	if sp.Duration > 0 {
		add("Duration: %s", durationText(sp.Duration))
	}
	if sp.Power != 0 {
		if rules.PaysHealth(sp.Type) {
			add("Health cost: %d", p.powerCost(sp))
		} else {
			add("Power cost: %d", p.powerCost(sp))
		}
	}
	if sp.pulsing() {
		add("A song or chant: it pulses while it is on")
	}
	return d
}

// durationText words a duration in seconds.
func durationText(s int) string {
	switch {
	case s >= 120 && s%60 == 0:
		return fmt.Sprintf("%d minutes", s/60)
	case s >= 60:
		return fmt.Sprintf("%d min %d s", s/60, s%60)
	}
	return fmt.Sprintf("%d s", s)
}

// styleInfo describes a style.
func styleInfo(s *StyleDef) *protocol.DelveInfo {
	d := &protocol.DelveInfo{Name: s.Name, Level: s.SpecLevel, Lines: []string{}}
	add := func(f string, a ...interface{}) { d.Lines = append(d.Lines, fmt.Sprintf(f, a...)) }
	add("%s style, level %d", s.Spec, s.SpecLevel)
	add("Needs %s", rules.StyleWeaponNeeds(s.WeaponReq, s.Spec))
	if s.Endurance > 0 {
		add("Endurance: %d", s.Endurance)
	}
	if s.Stealth {
		add("Only from stealth")
	}
	if s.BonusToHit != 0 {
		add("To hit: %+d%%", s.BonusToHit)
	}
	if s.BonusToDef != 0 {
		add("Defence: %+d%%", s.BonusToDef)
	}
	return d
}
