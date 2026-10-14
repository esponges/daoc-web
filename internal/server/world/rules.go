package world

import "math"

// This file holds the numbers the server's combat and progression use.
// They are this project's own, chosen so that fights in the starter zones
// feel like a fight, informed by how DAoC plays but not copied from any
// server's source. Imported data supplies the weapon, style and NPC stats;
// what turns those into hit points and damage is here.

// Inventory layout: slots below PackSize are the backpack; the rest are worn.
const (
	PackSize   = 40
	SlotWeapon = 40
	SlotShield = 41
	SlotTorso  = 42
	SlotLegs   = 43
	SlotArms   = 44
	SlotFeet   = 45
	SlotHands  = 46
	SlotHead   = 47
	SlotCount  = 48
)

// MaxLevel is the highest level a character reaches.
const MaxLevel = 50

// Timing, in seconds.
const (
	TickRate         = 20
	TickDT           = 1.0 / TickRate
	CombatLinger     = 6.0  // out of combat after this long without a blow
	NPCCorpseSeconds = 45.0 // a dead NPC lies this long (shorter once looted out)
	LootExclusive    = 30.0 // only the killer may loot for this long
	PlayerRespawnS   = 5.0
	DefensiveWindow  = 6.0 // a block, parry or evade opens styles this long
	ChainWindow      = 5.0 // a style opens its follow-ups this long
	DisconnectGrace  = 8.0 // a fighter stays in the world this long after the tab closes
	InterestRadius   = 14000.0
	SayRange         = 3000.0
)

func playerMaxHP(level int) int { return 120 + 55*level }
func playerMaxEnd() int         { return 100 }

// npcMaxHP is a creature's hit points by level.
func npcMaxHP(level int) int {
	l := float64(level)
	return int(10 + 12*l + 0.8*l*l)
}

// npcDamage is the middle of a creature's blow; each hit varies +-25%.
func npcDamage(level int) float64 { return 2.5 + 2.1*float64(level) + 0.04*float64(level*level) }

// npcSwing is how often a creature strikes, by its size: bigger swings slower.
func npcSwing(size int) float64 {
	s := 2.0 + float64(size-50)/50
	return math.Max(1.8, math.Min(4.0, s))
}

// xpToReach is the experience a character holds on reaching a level.
func xpToReach(level int) int64 {
	if level <= 1 {
		return 0
	}
	n := float64(level - 1)
	return int64(40*n*n + 20*n)
}

// killXP is the experience for a kill by a character of the given level.
func killXP(mobLevel, level int) int64 {
	if level >= MaxLevel {
		return 0
	}
	base := float64(xpToReach(level+1)-xpToReach(level)) * 0.14
	var f float64
	switch d := mobLevel - level; {
	case d >= 3:
		f = 1.5
	case d == 2:
		f = 1.3
	case d == 1:
		f = 1.15
	case d == 0:
		f = 1.0
	case d == -1:
		f = 0.85
	case d == -2:
		f = 0.7
	case d == -3:
		f = 0.55
	case d == -4:
		f = 0.4
	case d == -5:
		f = 0.3
	case d == -6:
		f = 0.2
	case d == -7:
		f = 0.12
	case d == -8:
		f = 0.06
	default:
		f = 0
	}
	return int64(math.Round(base * f))
}

// killMoney is the copper a creature of a level carries.
func killMoney(level int) int64 {
	if level < 1 {
		level = 1
	}
	return int64(4 + level*level*2 + level*5)
}

func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// playerHitChance is a character's chance to land a blow.
func playerHitChance(level, mobLevel, bonus int) float64 {
	return clamp(0.88+0.015*float64(level-mobLevel)+float64(bonus)/100, 0.55, 0.97)
}

// npcHitChance is a creature's chance to land a blow.
func npcHitChance(mobLevel, level int) float64 {
	return clamp(0.85-0.02*float64(level-mobLevel), 0.5, 0.95)
}

// weaponDamage is a character's blow with a weapon, before variation: the
// weapon's damage per second times its speed, scaled up with level.
func weaponDamage(dps, speed float64, level int) float64 {
	return dps * speed * (1.8 + 0.1*float64(level))
}

// armourReduction is the share of damage armour turns aside.
func armourReduction(af, attackerLevel int) float64 {
	return clamp(float64(af)/(float64(af)+100+float64(attackerLevel)*8), 0, 0.5)
}

// Defence chances (0..1) of a character.
func playerParry(level int, hasWeapon bool) float64 {
	if !hasWeapon {
		return 0
	}
	return clamp(0.06+0.004*float64(level), 0, 0.25)
}
func playerEvade(level int) float64 { return clamp(0.05+0.002*float64(level), 0, 0.15) }
func playerBlock(level int, hasShield bool) float64 {
	if !hasShield {
		return 0
	}
	return clamp(0.18+0.004*float64(level), 0, 0.35)
}

// Arc tests, with the angle measured from where the defender faces.
const (
	frontArc = math.Pi / 3     // within 60 degrees: in front
	backArc  = math.Pi * 2 / 3 // beyond 120 degrees: behind
)
