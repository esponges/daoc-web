package world

import (
	"math"

	"daocweb/internal/server/rules"
)

// This file holds the server's layout and timing constants. The game's rules
// -- hit points, damage, experience and the rest -- are in the rules package,
// written from the rule book.

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
const MaxLevel = rules.MaxLevel

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

func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// Arc tests, with the angle measured from where the defender faces.
const (
	frontArc = math.Pi / 3     // within 60 degrees: in front
	backArc  = math.Pi * 2 / 3 // beyond 120 degrees: behind
)
