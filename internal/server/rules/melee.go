package rules

import "math"

// Rolls is where the rules take their dice from: uniform numbers in [0, 1).
type Rolls interface{ Float64() float64 }

// Location is where a blow lands on a player.
type Location int

const (
	LocNone Location = iota - 1 // a creature, or no style location
	LocTorso
	LocLegs
	LocArms
	LocHead
	LocHands
	LocFeet
	NumLocations
)

var locationNames = [NumLocations]string{"torso", "legs", "arms", "head", "hands", "feet"}

func (l Location) String() string {
	if l < 0 || l >= NumLocations {
		return "none"
	}
	return locationNames[l]
}

// locationWeights are the chances, in percent, of a hit landing on each place.
var locationWeights = [NumLocations]int{40, 25, 15, 10, 5, 5}

// RollLocation picks where an unstyled hit lands on a player.
func RollLocation(r float64) Location {
	n := int(r * 100)
	for l, w := range locationWeights {
		if n < w {
			return Location(l)
		}
		n -= w
	}
	return LocFeet
}

// Weapon is what an attacker strikes with, as the rules use it.
type Weapon struct {
	DPS        float64 // damage per second
	Speed      float64 // seconds a swing
	ObjectType int
	TwoHanded  bool
	Quality    float64 // 0..1
	Condition  float64 // 0..1
	DamageType int     // DamageCrush, DamageSlash, DamageThrus
	Bonus      int     // enchantment, percent
}

// Piece is the armour worn on one location.
type Piece struct {
	AF        float64
	Absorb    float64 // 0..1
	Cloth     bool
	Quality   float64
	Condition float64
	Bonus     int
}

// Shield is the shield a player holds.
type Shield struct {
	Size      int // 1 small, 2 medium, 3 large
	Quality   float64
	Condition float64
}

// Combatant is one side of a swing: a player or a creature.
type Combatant struct {
	Player bool
	Level  int
	Stats  Stats

	// Offence.
	WeaponSkillBase int // the class's (players) or 440 (creatures)
	Weapon          Weapon
	Unarmed         bool // a player with nothing in hand, who cannot parry
	Spec            int  // combat spec in the held weapon's line

	// Defence. Players: abilities and specs; creatures: template percents.
	EvadeLevel int
	ParrySpec  int // 0 when the player has no parry
	ShieldSpec int
	Shield     *Shield
	EvadePct   float64
	ParryPct   float64
	BlockPct   float64

	// Armour: players by location; creatures one value for all.
	Armour         [NumLocations]*Piece
	CreatureAF     float64
	CreatureAbsorb float64
	Resists        map[int]float64 // damage type -> percent

	// The defence bonus of this combatant's last styled attack, while it holds.
	StyleDefence int

	// EngageBlock is added to a player's block chance while they engage the attacker.
	EngageBlock float64
}

// Style is a style as a swing uses it.
type Style struct {
	GrowthRate float64
	Spec       int // the attacker's combat spec in the style's line
	BonusToHit int
	Location   Location
}

// Swing is one attack to resolve.
type Swing struct {
	Attacker, Defender *Combatant
	Style              *Style
	InFront            bool // the attacker is within 60 degrees of the defender's facing
	Attackers          int  // how many are attacking the defender, at least 1
	NoDefence          bool // the blow cannot be evaded, parried, blocked or missed (an intercepted one)
}

// Result is how a swing ended.
type Result int

const (
	Hit Result = iota
	Evaded
	Parried
	Blocked
	Missed
)

func (r Result) String() string {
	return [...]string{"hit", "evade", "parry", "block", "miss"}[r]
}

// Outcome is everything a swing decided.
type Outcome struct {
	Result     Result
	Location   Location
	Damage     int // dealt, styled and critical damage included
	Critical   int // the critical part of Damage
	StyleFired bool
	Chance     struct{ Evade, Parry, Block, Miss float64 }
}

// Resolve settles a swing. It always draws the same eight rolls in the same
// order (evade, parry, block, location, miss, variance, critical, critical
// size), so a sequence of rolls describes a swing completely.
func Resolve(s Swing, r Rolls) Outcome {
	var dice [8]float64
	for i := range dice {
		dice[i] = r.Float64()
	}
	a, d := s.Attacker, s.Defender
	var o Outcome
	o.Location = LocNone
	o.Chance.Evade = EvadeChance(d, s.InFront, s.Attackers)
	o.Chance.Parry = ParryChance(d, s.InFront, s.Attackers, a.Weapon.TwoHanded)
	o.Chance.Block = BlockChance(d, s.InFront)
	switch {
	case s.NoDefence:
	case dice[0] < o.Chance.Evade:
		o.Result = Evaded
		return o
	case dice[1] < o.Chance.Parry:
		o.Result = Parried
		return o
	case dice[2] < o.Chance.Block:
		o.Result = Blocked
		return o
	}
	if d.Player {
		o.Location = RollLocation(dice[3])
		if s.Style != nil && s.Style.Location >= 0 {
			o.Location = s.Style.Location
		}
	}
	o.Chance.Miss = MissChance(s, o.Location)
	if !s.NoDefence && dice[4] < o.Chance.Miss {
		o.Result = Missed
		return o
	}
	o.Result = Hit
	raw, styled, fired := HitDamage(s, o.Location, dice[5])
	o.StyleFired = fired
	o.Damage = int(raw + styled)
	if a.Player && dice[6] < CriticalChance {
		top := 1.0
		if d.Player {
			top = 0.5
		}
		o.Critical = int(float64(o.Damage) * (0.1 + dice[7]*(top-0.1)))
		o.Damage += o.Critical
	}
	return o
}

// --- defences ---

// perMille turns tenths of a percent into a chance.
func perMille(n int) float64 { return float64(n) / 1000 }

// EvadeChance is the defender's chance to evade this blow.
func EvadeChance(d *Combatant, inFront bool, attackers int) float64 {
	if !inFront {
		return 0
	}
	if !d.Player {
		return d.EvadePct / 100
	}
	if d.EvadeLevel <= 0 {
		return 0
	}
	c := perMille((900 + d.Stats[QUI] + d.Stats[DEX]) * d.EvadeLevel * 5 / 100)
	return c / (1 + float64(max(1, attackers)-1)/3)
}

// skillDefence is the parry and block base in tenths of a percent.
func skillDefence(dex, spec int) int { return (dex*2-100)/4 + (spec-1)*5 + 50 }

// ParryChance is the defender's chance to parry this blow.
func ParryChance(d *Combatant, inFront bool, attackers int, twoHanded bool) float64 {
	if !inFront {
		return 0
	}
	if !d.Player {
		return d.ParryPct / 100
	}
	if d.ParrySpec <= 0 || d.Unarmed {
		return 0
	}
	c := perMille(skillDefence(d.Stats[DEX], d.ParrySpec)) / float64(max(1, attackers))
	if twoHanded {
		c /= 2
	}
	return math.Max(0, c)
}

// BlockChance is the defender's chance to block this blow.
func BlockChance(d *Combatant, inFront bool) float64 {
	if !inFront {
		return 0
	}
	if !d.Player {
		return d.BlockPct / 100
	}
	if d.Shield == nil || d.ShieldSpec <= 0 || d.Weapon.TwoHanded {
		return 0
	}
	c := perMille(skillDefence(d.Stats[DEX], d.ShieldSpec)) * d.Shield.Quality * d.Shield.Condition
	return math.Max(0, c+d.EngageBlock)
}

// --- miss ---

// BaseMiss is the miss chance between equals.
const BaseMiss = 18.0

// ItemBonusCap is the most enchantment bonus that counts at a level.
func ItemBonusCap(level int) int {
	if level < 15 {
		return 0
	}
	return min(35, (level/5-2)*5)
}

// MissChance is the attacker's chance to miss, given where the blow would land.
func MissChance(s Swing, loc Location) float64 {
	a, d := s.Attacker, s.Defender
	m := BaseMiss - float64(a.Level-d.Level)*4/3
	if s.Style != nil {
		m -= float64(s.Style.BonusToHit)
	}
	m += float64(d.StyleDefence)
	armour := 0
	if d.Player && loc >= 0 && d.Armour[loc] != nil {
		armour = min(d.Armour[loc].Bonus, ItemBonusCap(d.Level))
	}
	weapon := min(a.Weapon.Bonus, ItemBonusCap(a.Level))
	m += m * float64(armour-weapon) / 100
	return math.Max(0, math.Min(100, m)) / 100
}

// --- damage ---

// CriticalChance is a player's chance to land a critical hit.
const CriticalChance = 0.10

// DPSCap is the most weapon damage per second a level can use.
func DPSCap(level int) float64 { return 1.2 + 0.3*float64(level) }

// BaseDamage is a blow's damage before weapon skill and armour, and its cap.
func BaseDamage(a *Combatant) (base, cap float64) {
	w := a.Weapon
	if !a.Player {
		l := float64(a.Level)
		base = (1 + l/2.5 + l*l/120) * w.Speed
		return base, 3 * base
	}
	base = math.Min(w.DPS, DPSCap(a.Level)) * w.Speed * (1 + (w.Speed*10-20)*0.003)
	if w.TwoHanded {
		base *= 1.1 + 0.005*float64(a.Spec)
	}
	return base * w.Quality * w.Condition, 3 * base
}

// WeaponStat is the stat a weapon's skill grows with.
func WeaponStat(a *Combatant) int {
	switch a.Weapon.ObjectType {
	case ObjStaff, ObjCompBow, ObjThrown:
		return a.Stats[DEX]
	case ObjSpear:
		return (a.Stats[STR] + a.Stats[DEX]) / 2
	}
	return a.Stats[STR]
}

// WeaponSkill is the attacker's weapon skill before variance.
func WeaponSkill(a *Combatant) float64 {
	return float64(a.Level) * float64(a.WeaponSkillBase) * 0.005 * (1 + float64(WeaponStat(a))*0.005)
}

// Variance is the weapon skill multiplier for a roll r.
func Variance(a *Combatant, targetLevel int, r float64) float64 {
	if !a.Player {
		return 0.9 + 0.2*r
	}
	ratio := math.Max(0, math.Min(1, float64(a.Spec-1)/float64(targetLevel+1)))
	return 0.75 + 0.5*ratio + 0.5*r
}

// LocationArmour is the armour factor and absorption a blow meets.
func LocationArmour(d *Combatant, loc Location) (af, absorb float64) {
	if !d.Player {
		return d.CreatureAF, d.CreatureAbsorb
	}
	l := float64(d.Level)
	extra := 0.4 * l
	if loc < 0 || d.Armour[loc] == nil {
		return extra, 0
	}
	p := d.Armour[loc]
	cap := 2 * l
	if p.Cloth {
		cap = l
	}
	af = math.Min(math.Min(p.AF, cap)*p.Quality*p.Condition, 2*l)
	return af + extra, p.Absorb
}

// ArmourTerm is what weapon skill is divided by.
func ArmourTerm(af, absorb float64) float64 {
	if absorb >= 1 {
		return math.MaxFloat64
	}
	return math.Max(1, af/(1-absorb))
}

// HitDamage is a landed blow's unstyled and styled damage, untruncated, and
// whether the style fired.
func HitDamage(s Swing, loc Location, variance float64) (raw, styled float64, fired bool) {
	a, d := s.Attacker, s.Defender
	base, cap := BaseDamage(a)
	af, absorb := LocationArmour(d, loc)
	raw = base * WeaponSkill(a) * Variance(a, d.Level, variance) / ArmourTerm(af, absorb)
	resist := 1 - d.Resists[a.Weapon.DamageType]/100
	if s.Style != nil {
		fired = true
		g := s.Style.GrowthRate * float64(s.Style.Spec) * a.Weapon.Speed
		styled = raw * g / cap * resist
		if s.Style.GrowthRate > 0 && styled < 1 {
			styled = 1
		} else {
			styled = math.Min(styled, g)
		}
	}
	raw = math.Min(raw*resist, cap)
	return raw, styled, fired
}

// --- timing and endurance ---

// SwingTime is a player's seconds between swings.
func SwingTime(speed float64, quickness int) float64 {
	return math.Max(1.5, speed*(1-float64(min(quickness, 250)-50)*0.002))
}

// CreatureSwingTime is a creature's seconds between swings.
func CreatureSwingTime(speed float64, quickness int) float64 {
	return math.Max(0.5, speed*(1-float64(quickness-50)*0.002))
}

// StyleCost is a style's endurance with a weapon of this speed.
func StyleCost(speed float64, endurance int) int {
	tenths := int(math.Round(speed * 10))
	c := tenths * endurance / 40
	if tenths < 40 {
		c++
	}
	return max(1, c)
}

// FailedStyleCost is what a style costs when its swing does not land.
func FailedStyleCost(speed float64, endurance int) int { return max(1, StyleCost(speed, endurance)/2) }

func init() {
	register(
		Rule{ID: "melee.order", Capability: "rules-melee", Requirement: "A swing resolves in a fixed order",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A parried swing"}},
		Rule{ID: "melee.evade", Capability: "rules-melee", Requirement: "Evade",
			Status: Documented, Sources: []string{"disorder", "grabbag", "opendaoc"}, Scenarios: []string{"Evade I", "From behind"}},
		Rule{ID: "melee.parry", Capability: "rules-melee", Requirement: "Parry",
			Status: Documented, Sources: []string{"disorder", "grabbag"}, Scenarios: []string{"Parry 20", "Two attackers", "Untrained"}},
		Rule{ID: "melee.block", Capability: "rules-melee", Requirement: "Block",
			Status: Documented, Sources: []string{"disorder", "grabbag"}, Scenarios: []string{"Shields 1", "Worn shield"}},
		Rule{ID: "melee.miss", Capability: "rules-melee", Requirement: "Miss chance",
			Status: Documented, Sources: []string{"disorder", "opendaoc"}, Scenarios: []string{"Even fight", "Higher creature"}},
		Rule{ID: "melee.miss-bonuses", Capability: "rules-melee", Requirement: "Bonuses move the miss chance",
			Status: Reference, Sources: []string{"disorder", "opendaoc"}, Scenarios: []string{"A style with a to-hit bonus", "An enchanted weapon"}},
		Rule{ID: "melee.location", Capability: "rules-melee", Requirement: "Hit location on a player",
			Status: Reference, Sources: []string{"opendaoc", "disorder"}, Scenarios: []string{"Bare legs"}},
		Rule{ID: "melee.base-damage", Capability: "rules-melee", Requirement: "Weapon damage and its cap",
			Status: Documented, Sources: []string{"herald", "opendaoc"}, Scenarios: []string{"Over-level weapon"}},
		Rule{ID: "melee.weapon-skill", Capability: "rules-melee", Requirement: "Weapon skill against armour",
			Status: Reference, Sources: []string{"herald", "opendaoc"}, Scenarios: []string{"Mid roll on an even creature", "The spread", "Untrained weapon", "Two-handed"}},
		Rule{ID: "melee.critical", Capability: "rules-melee", Requirement: "Critical hits",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"A big crit"}},
		Rule{ID: "melee.style-damage", Capability: "rules-melee", Requirement: "Styles add growth damage",
			Status: Documented, Sources: []string{"herald", "opendaoc"}, Scenarios: []string{"A styled hit"}},
		Rule{ID: "melee.style-cost", Capability: "rules-melee", Requirement: "Styles cost endurance by weapon speed",
			Status: Reference, Sources: []string{"opendaoc"}, Scenarios: []string{"Quick sword", "Slow axe"}},
		Rule{ID: "melee.swing-speed", Capability: "rules-melee", Requirement: "Swing speed",
			Status: Documented, Sources: []string{"disorder"}, Scenarios: []string{"A slow Troll", "Floor"}},
		Rule{ID: "melee.armour", Capability: "rules-melee", Requirement: "Armour factor of a player's location",
			Status: Documented, Sources: []string{"disorder", "opendaoc", "data"}, Scenarios: []string{"Studded torso"}},
	)
}
