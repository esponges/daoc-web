// Package protocol is the wire format between the browser and the game
// server: JSON over a WebSocket, each message an object with a "t" naming
// its type. The client sends its inputs and requests; the server decides and
// answers with snapshots. docs/protocol.md describes it for people.
//
// Unknown types are ignored by both ends, so either side can grow new
// messages without breaking an older peer; a change that would break one
// raises Version.
package protocol

import (
	"encoding/json"
	"fmt"
)

// Version is the protocol revision. A client that sends another is refused.
const Version = 1

// StepsPerSecond is the fixed simulation step the client's predicted
// movement and the server's authoritative movement both use. Inputs are sent
// as runs of these steps, so the two integrate identically.
const StepsPerSecond = 60

// Message types, client to server.
const (
	THello        = "hello"  // first message: protocol version and the character to play
	TInput        = "input"  // a run of fixed movement steps
	TChat         = "chat"   // say or broadcast
	TTarget       = "target" // select an entity
	TAttack       = "attack" // auto-attack on or off
	TStyle        = "style"  // use a style by its slot on the bar
	TLoot         = "loot"   // take an item from a corpse
	TEquip        = "equip"  // move an item between slots
	TUse          = "use"    // use a zone exit, etc.
	TGM           = "gm"     // an administrator's command
	TPing         = "ping"
	TRelease      = "release"      // a dead character asks to be stood up now rather than when the timer ends
	TInteract     = "interact"     // ask an NPC to do what it does (open a merchant's shop)
	TBuy          = "buy"          // buy an item from the open shop
	TSell         = "sell"         // sell a pack item to the open shop
	TBarSet       = "barset"       // put a style on a bar slot
	TRepair       = "repair"       // ask the open blacksmith for a repair price, or accept it
	TTrain        = "train"        // raise a spec at the trainer
	TRestore      = "restore"      // pay a healer to restore lost constitution
	TPray         = "pray"         // pray at the targeted gravestone
	TGroupOp      = "groupop"      // invite, accept, decline, leave or remove from a group
	TCast         = "cast"         // cast a spell from the spellbook
	TRezAnswer    = "rez-answer"   // accept or decline a resurrection offer
	TConcCancel   = "conc-cancel"  // let go of one of the buffs the caster holds
	TGroundTarget = "gt"           // set or clear the point abilities that need a ground target aim at
	TPetCmd       = "pet-cmd"      // an order to your pet: attack, follow, stay, release, or a stance
	TAskTask      = "ask-task"     // talk to an NPC about tasks: ask for one, or hand one in
	TAbandonTask  = "abandon-task" // give up the task you hold
	TUseItem      = "use-item"     // use the item in a pack slot (a poison goes on a weapon)
	TShoot        = "shoot"        // draw a bow at the held target, switch Critical Shot or Rapid Fire, or stop
	TDelve        = "delve"        // right-click: ask for the details of an item, spell or style
)

// Message types, server to client.
const (
	TWelcome     = "welcome"
	TSnap        = "snap"
	TChatMsg     = "chatmsg"
	TErr         = "err"
	TZone        = "zone" // the character moved zone: reload at the given one
	TInv         = "inv"  // the character's inventory and equipment
	TCorpse      = "corpse"
	TPong        = "pong"
	TStyleBar    = "bar"
	TBook        = "book"        // the spellbook, sent when it changes
	TShop        = "shop"        // a merchant's stock, sent when interaction opens it
	TShopClose   = "shopclose"   // the open shop is closed (walked away, died, moved zone)
	TTrainClose  = "trainclose"  // the open trainer is closed (walked away, died, moved zone)
	TSmith       = "smith"       // what a blacksmith can repair for the player, with prices
	TSmithClose  = "smithclose"  // the repair window is closed
	TRepairQuote = "repairquote" // the price of repairing one item, to be accepted
	TClock       = "clock"       // the game time of day, on entering and once a minute
	TVault       = "vault"       // the character's own vault, sent when a vault keeper opens it
	TVaultClose  = "vaultclose"  // the open vault is closed
	TTrade       = "trade"       // the state of a trade you are in: both offers and both accepts
	TTradeClose  = "tradeclose"  // your trade ended, with the reason
	TGuild       = "guild"       // your guild, or an empty one when you have none
	TDuelAsk     = "duelask"     // you have been challenged to a duel; answer with /accept or /decline
	TQuestOp     = "questop"     // accept or decline an offered quest, abandon one, choose a reward, ask for the journal
	TQuests      = "quests"      // the quest journal: active quests with their step, and those done
	TQuestOffer  = "questoffer"  // an NPC offers a quest
	TQuestReward = "questreward" // a finished quest offers a choice of items
	TCraftOp     = "craftop"     // open the crafting window, join a profession, make an item or cancel
	TCraft       = "craft"       // the crafting window: your skills and what you can make
	TCraftProg   = "craftprog"   // a craft began: how long it takes
	TCraftDone   = "craftdone"   // a craft ended: made, failed or cancelled
	TFriends     = "friends"     // your friends list, with who is in the game
	TTradeOp     = "tradeop"     // a request in a trade: open, add, remove, money, accept or cancel
	TStats       = "stats"       // the character sheet
	TTrainer     = "trainer"     // a trainer's offer: the character's specs and their costs
	THealer      = "healer"      // a healer's offer to restore lost constitution
	TEffects     = "effects"     // the character's own lasting effects, and the speed they leave
	TGroup       = "group"       // the character's group: its members and how they fare
	TGroupInv    = "groupinv"    // an invitation to join a group
	TRezOffer    = "rez-offer"   // a resurrection offered to a dead character, or withdrawn
	TConc        = "conc"        // the concentration a caster holds: pool, use and the buffs
	TPet         = "pet"         // your pet: its name, level, health, stance and orders; closed when it is gone
	TTask        = "task"        // your task: what it is, its progress and time left; on false when you hold none
	TRecast      = "recast"      // a spell you cast has a recast: its id and the seconds until it is ready
	TDelveInfo   = "delveinfo"   // the details a delve asked for
)

// Envelope reads just the type of a message.
type Envelope struct {
	T string `json:"t"`
}

// Hello is the client's first message.
type Hello struct {
	T    string `json:"t"`
	V    int    `json:"v"`
	Char int64  `json:"char"`
}

// Move flag bits in a step run.
const (
	FlagRun  = 1 << iota // run rather than walk (backing and strafing always walk)
	FlagJump             // jump, if on the ground and free to
)

// Input is a batch of movement. Steps is a list of runs, each
// [count, fwd, strafe, turn, flags]: that many fixed steps with those keys
// held, fwd and strafe in -1..1 and turn -1 (left), 0 or 1 (right). Seq is
// the cumulative number of steps the client has sent including these, which
// the server echoes as Ack once applied. PX and PY are the client's
// predicted position after them, which the server only compares.
type Input struct {
	T     string      `json:"t"`
	Seq   int64       `json:"seq"`
	Steps [][5]int    `json:"steps"`
	PX    float64     `json:"px"`
	PY    float64     `json:"py"`
	Extra interface{} `json:"-"`
}

// Chat is a chat line.
type Chat struct {
	T    string `json:"t"`
	Ch   string `json:"ch"` // "say" or "all"
	Text string `json:"text"`
}

// Target selects an entity (0 clears it).
type Target struct {
	T  string `json:"t"`
	ID int64  `json:"id"`
}

// Attack switches auto-attack.
type Attack struct {
	T  string `json:"t"`
	On bool   `json:"on"`
}

// Style asks to use the style on a bar slot.
type Style struct {
	T    string `json:"t"`
	Slot int    `json:"slot"`
	ID   int    `json:"id,omitempty"` // when set, the style by its id wherever it sits (a quickbar slot)
}

// Loot takes an item from a corpse; Index -1 takes the money.
type Loot struct {
	T      string `json:"t"`
	Corpse int64  `json:"corpse"`
	Index  int    `json:"index"`
}

// Equip moves an item from one slot to another (equipping is a move into a worn slot). To -1 destroys it, -2 wears it where it belongs.
// (Equip, continued:)
// move into an equipment slot).
type Equip struct {
	T    string `json:"t"`
	From int    `json:"from"`
	To   int    `json:"to"`
}

// Interact asks the NPC with this id to do what it does.
type Interact struct {
	T  string `json:"t"`
	ID int64  `json:"id"`
}

// Buy takes one of the open shop's items, named by its item id.
type Buy struct {
	T  string `json:"t"`
	ID string `json:"id"`
}

// Sell sells the item in a pack slot to the open shop.
// Repair asks the open blacksmith to repair the item in a slot: without Accept
// for the price, with it to pay the Price the player was shown.
type Repair struct {
	T      string `json:"t"`
	Slot   int    `json:"slot"`
	Accept bool   `json:"accept,omitempty"`
	Price  int64  `json:"price,omitempty"`
}

type Sell struct {
	T    string `json:"t"`
	Slot int    `json:"slot"`
}

// Train raises a spec by N levels (1 when absent) at the trainer in reach.
type Train struct {
	T    string `json:"t"`
	Spec string `json:"spec"`
	N    int    `json:"n,omitempty"`
}

// Restore asks the healer with this id to restore lost constitution.
type Restore struct {
	T   string `json:"t"`
	NPC int64  `json:"npc"`
}

// Use asks to use the zone exit the player stands at.
type Use struct {
	T string `json:"t"`
}

// GM is an administrator's command line, e.g. "goto 103".
type GM struct {
	T    string `json:"t"`
	Line string `json:"line"`
}

// --- server to client ---

// Welcome answers a good hello.
type Welcome struct {
	T      string  `json:"t"`
	V      int     `json:"v"`
	You    int64   `json:"you"` // your entity id
	Name   string  `json:"name"`
	Char   string  `json:"char"`         // the converted character to draw you with
	Race   int     `json:"rc,omitempty"` // your race number, for the race-and-gender base model
	Female bool    `json:"fm,omitempty"`
	Face   int     `json:"fc,omitempty"` // the face you chose, 1 up; absent is the race's default
	Hair   int     `json:"hc,omitempty"` // the hair colour you chose, 1 up
	Zone   int     `json:"zone"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	Yaw    float64 `json:"yaw"`
	Level  int     `json:"level"`
	Admin  bool    `json:"admin,omitempty"`
}

// Entity is one thing in a snapshot.
type Entity struct {
	ID       int64    `json:"id"`
	K        string   `json:"k"`            // "p" player, "n" npc
	Name     string   `json:"nm"`           // as it is shown
	Model    int      `json:"m,omitempty"`  // an NPC's model number
	Equip    [][4]int `json:"eq,omitempty"` // an NPC's worn pieces: [slot, model, colour, extension]
	Guild    string   `json:"gd,omitempty"` // an NPC's guild or description, shown under its name
	Char     string   `json:"ch,omitempty"` // a player's character
	Race     int      `json:"rc,omitempty"` // a player's race number
	Female   bool     `json:"fm,omitempty"`
	Face     int      `json:"fc,omitempty"` // a player's chosen face and hair colour, 1 up
	Hair     int      `json:"hc,omitempty"`
	GearSet  bool     `json:"gs,omitempty"` // a player's eq is present even when empty: it went out because it changed
	Level    int      `json:"lv,omitempty"`
	X        float64  `json:"x"`
	Y        float64  `json:"y"`
	Z        float64  `json:"z"`
	Yaw      float64  `json:"h"`
	Speed    float64  `json:"sp,omitempty"` // current ground speed, units a second
	HP       float64  `json:"hp"`           // 0..1
	State    string   `json:"st,omitempty"` // "", "fight", "dead"
	Scale    float64  `json:"sc,omitempty"`
	Aggro    bool     `json:"ag,omitempty"`
	Loot     bool     `json:"lt,omitempty"` // a corpse with something to take
	Armed    bool     `json:"ar,omitempty"`
	Shld     bool     `json:"sh,omitempty"`
	Back     bool     `json:"bk,omitempty"`
	Side     int      `json:"sd,omitempty"`
	Friendly bool     `json:"fr,omitempty"`
	Kind     string   `json:"ik,omitempty"` // what interacting does: "merchant", "trainer", "healer", "stable", "vault", "smith", "notclassic"; absent for nothing
	Con      *int     `json:"cc,omitempty"` // a creature's con colour to you, -3 grey to 3 purple
	Own      bool     `json:"own,omitempty"`
	Ride     bool     `json:"rd,omitempty"`  // a player on a horse
	Owner    string   `json:"ow,omitempty"`  // the name of a pet's owner
	Posture  string   `json:"ps,omitempty"`  // a player's posture: "sit", "air" or "swim"; absent standing
	Foe      bool     `json:"foe,omitempty"` // the player you are in a duel with
	VZ       float64  `json:"vz,omitempty"`  // a player in the air: vertical speed, units a second
	Torch    bool     `json:"tc,omitempty"`  // a player whose torch is lit: a light around them at night
}

// Event is something that happened this tick, for the log and the animations.
type Event struct {
	K      string  `json:"k"` // swing, hit, miss, block, parry, evade, style, death, level, xp, loot, spell, resist, effect
	From   int64   `json:"from,omitempty"`
	To     int64   `json:"to,omitempty"`
	Amount int     `json:"n,omitempty"`
	Text   string  `json:"x,omitempty"`
	Style  int     `json:"s,omitempty"`
	Crit   bool    `json:"c,omitempty"`
	At     float64 `json:"at,omitempty"`
	Dur    float64 `json:"dur,omitempty"` // a cast's time, in seconds
	FX     int     `json:"fx,omitempty"`  // the client effect of the spell, on cast, spell, effect and resist events: how the spell looks
}

// Self is the part of a snapshot only the player gets.
type Self struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Z         float64 `json:"z"`
	Yaw       float64 `json:"h"`
	HP        int     `json:"hp"`
	MaxHP     int     `json:"mhp"`
	End       int     `json:"end"`
	MaxEnd    int     `json:"mend"`
	Power     int     `json:"pw,omitempty"` // spell power, casters only
	MaxPower  int     `json:"mpw,omitempty"`
	Level     int     `json:"lv"`
	XP        int64   `json:"xp"`
	XPNext    int64   `json:"xpn"`
	Money     int64   `json:"mo"`
	Target    int64   `json:"tg"`
	Attacking bool    `json:"at"`
	Dead      bool    `json:"dead,omitempty"`
	Correct   bool    `json:"fix,omitempty"` // the client's prediction was wrong: snap, do not smooth
	Back      bool    `json:"bk,omitempty"`  // moving backwards, for the clip
	Side      int     `json:"sd,omitempty"`
	Speed     float64 `json:"sp"`
	Ride      bool    `json:"rd,omitempty"` // on a horse route: inputs do nothing
	Fighting  bool    `json:"fg,omitempty"`
	Avail     []int   `json:"av,omitempty"` // style bar slots usable right now
	// The body's vertical state, so the client's prediction can pick up where
	// the server is: in the air, its speed up, its highest point, and whether
	// the air began with a jump.
	Air    bool      `json:"air,omitempty"`
	VZ     float64   `json:"vz,omitempty"`
	Peak   float64   `json:"pk,omitempty"`
	Jumped bool      `json:"jp,omitempty"`
	Seated bool      `json:"sit,omitempty"` // you are sitting
	Ground []float64 `json:"gt,omitempty"`  // your ground target, x y z, when one is set
	Follow bool      `json:"fl,omitempty"`  // the server is walking you after who you follow: do not predict
	Torch  bool      `json:"tc,omitempty"`  // your torch is lit
}

// Snap is the world as one player sees it.
type Snap struct {
	T    string   `json:"t"`
	Tick int64    `json:"tick"`
	Ack  int64    `json:"ack"` // input steps applied so far
	You  Self     `json:"you"`
	Ents []Entity `json:"ents"`
	Ev   []Event  `json:"ev,omitempty"`
}

// ChatMsg is a chat line to show.
type ChatMsg struct {
	T    string `json:"t"`
	Ch   string `json:"ch"` // say, all, tell, group or sys
	From string `json:"from"`
	To   string `json:"to,omitempty"` // on your own tell: whom it went to
	Text string `json:"text"`
}

// Err is a message for the player, with a machine-readable code.
type Err struct {
	T    string `json:"t"`
	Code string `json:"code"`
	Text string `json:"text"`
}

// ZoneChange tells the client to load another zone.
type ZoneChange struct {
	T    string  `json:"t"`
	Zone int     `json:"zone"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

// InvItem is one stack for the client's inventory panel.
type InvItem struct {
	Slot   int    `json:"slot"`
	ID     string `json:"id"`
	Name   string `json:"nm"`
	Count  int    `json:"n"`
	Level  int    `json:"lv,omitempty"`
	Model  int    `json:"m,omitempty"`
	Kind   string `json:"k,omitempty"`    // weapon, shield, armour, misc
	Worn   int    `json:"w,omitempty"`    // the worn slot an armour piece belongs in
	Line   string `json:"line,omitempty"` // a weapon's line: Sword, Hammer, Axe
	Detail string `json:"d,omitempty"`
	Price  int64  `json:"pr,omitempty"` // in copper, on a shop's items
	Poison int    `json:"ps,omitempty"` // a poisoned weapon: the hits the poison has left
	// What the item is in the item table's own numbers, so the client can pick its picture:
	// ot the object type (11 sword, 12 hammer, 13 axe, 14 spear, 15 composite bow, 43 arrow, ...),
	// ty the item type (12 for a two-handed weapon, 21 to 28 the armour slots, ...) and sz a
	// shield's size, 1 to 3.
	ObjType  int `json:"ot,omitempty"`
	ItemType int `json:"ty,omitempty"`
	Size     int `json:"sz,omitempty"`
}

// Inv is a full inventory.
type Inv struct {
	T     string    `json:"t"`
	Items []InvItem `json:"items"`
	Money int64     `json:"mo"`
	Eq    [][4]int  `json:"eq"` // what you wear, as [slot, model, colour, extension], for drawing you
}

// Shop is a merchant's stock. Each item's Slot is its place in the list, in
// the order to show; buying names the item by ID.
type Shop struct {
	T     string    `json:"t"`
	NPC   int64     `json:"npc"`
	Name  string    `json:"nm"`
	Items []InvItem `json:"items"`
}

// Smith lists what a blacksmith can repair for the player: each item with
// its repair price in Price (copper) and its condition in Detail.
type Smith struct {
	T     string    `json:"t"`
	NPC   int64     `json:"npc"`
	Name  string    `json:"nm"`
	Items []InvItem `json:"items"`
}

// SmithClose says the repair window is gone.
type SmithClose struct {
	T string `json:"t"`
}

// RepairQuote is the price of repairing one item, which the player accepts or
// not.
type RepairQuote struct {
	T     string `json:"t"`
	Slot  int    `json:"slot"`
	Name  string `json:"nm"`
	Price int64  `json:"price"`
}

// Clock is the game time of day, in seconds since midnight. A game day lasts
// 60 real minutes, so the client runs it at 24 times real time between messages.
type Clock struct {
	T  string `json:"t"`
	GS int    `json:"gs"`
}

// Vault is a character's vault: its items, each with its slot (1000 to 1039,
// the numbers an equip move names). Sent only to its owner.
type Vault struct {
	T     string    `json:"t"`
	NPC   int64     `json:"npc"`
	Name  string    `json:"nm"`
	Items []InvItem `json:"items"`
}

// VaultClose says the open vault is gone.
type VaultClose struct {
	T string `json:"t"`
}

// ShopClose says the open shop is gone.
type ShopClose struct {
	T string `json:"t"`
}

// TrainClose says the open trainer window is closed.
type TrainClose struct {
	T string `json:"t"`
}

// Corpse lists what a dead NPC dropped, for its killer.
type Corpse struct {
	T     string    `json:"t"`
	ID    int64     `json:"id"`
	Name  string    `json:"nm"`
	Items []InvItem `json:"items"`
	Money int64     `json:"mo"`
}

// BarSlot is one style on the hotbar.
type BarSlot struct {
	Slot   int    `json:"slot"`
	ID     int    `json:"id"`
	Name   string `json:"nm"`
	Icon   int    `json:"icon"`
	Spec   string `json:"spec,omitempty"` // the weapon line: "Sword", "Hammer", "Axe", ...
	Cost   int    `json:"cost"`
	Opens  string `json:"opens"` // "anytime", "after block", ...
	Level  int    `json:"lv"`
	Locked bool   `json:"locked,omitempty"`
	Needs  string `json:"nd,omitempty"` // a locked style: what it needs in the hands ("a hand to hand weapon in each hand")
}

// SpecRow is one spec on the character sheet or a trainer's list.
type SpecRow struct {
	Key   string   `json:"key"`
	Name  string   `json:"nm"`
	Level int      `json:"lv"`
	Bonus int      `json:"bonus,omitempty"` // from items, as combat counts it
	Cost  int      `json:"cost,omitempty"`  // of the next level, on a trainer's list
	Can   bool     `json:"can,omitempty"`   // the next level can be trained now
	Gives []Unlock `json:"gives,omitempty"` // on a trainer's list: up to five unlocks above the level
}

// Ability is one ability a character knows; On is false for one the server
// does not implement yet.
type Ability struct {
	Name  string `json:"nm"`
	Level int    `json:"lv"`
	On    bool   `json:"on,omitempty"`
}

// Stats is the character sheet, sent to its owner only.
type Stats struct {
	T         string    `json:"t"`
	Class     string    `json:"cls"`
	Race      string    `json:"race,omitempty"` // the race's name, when the store has it
	Level     int       `json:"lv"`
	Half      bool      `json:"half,omitempty"` // past the midpoint of the level, from 40
	Base      [8]int    `json:"base"`           // strength, constitution, dexterity, quickness, intelligence, piety, empathy, charisma
	Stats     [8]int    `json:"stats"`          // with level gains, capped items and constitution lost
	MaxHP     int       `json:"mhp"`
	MaxEnd    int       `json:"mend"`
	MaxPower  int       `json:"mpw,omitempty"`
	MaxConc   int       `json:"mcn,omitempty"` // the concentration pool, for a class that casts
	ConcUsed  int       `json:"cnu,omitempty"` // and what the caster's buffs hold of it
	Specs     []SpecRow `json:"specs"`
	Unspent   int       `json:"pts"`
	Abilities []Ability `json:"ab"`
	WS        int       `json:"ws"` // weapon skill of the held weapon, as shown
	AF        int       `json:"af"` // armour factor, averaged over the body
	ConLost   int       `json:"conlost,omitempty"`
}

// Trainer is a trainer's offer.
type Trainer struct {
	T       string    `json:"t"`
	NPC     int64     `json:"npc"`
	Name    string    `json:"nm"`
	Specs   []SpecRow `json:"specs"`
	Unspent int       `json:"pts"`
	Say     string    `json:"say,omitempty"` // the trainer's greeting, on opening only
}

// Unlock is something a spec grants at a level: a style, an ability or a spell.
type Unlock struct {
	Level int    `json:"lv"`
	Name  string `json:"nm"`
	Kind  string `json:"k"` // "style", "ability" or "spell"
}

// Healer is a healer's offer to restore lost constitution.
type Healer struct {
	T     string `json:"t"`
	NPC   int64  `json:"npc"`
	Name  string `json:"nm"`
	Lost  int    `json:"lost"`
	Price int64  `json:"price"` // per point, in copper
	Total int64  `json:"total"`
}

// Bar is the style hotbar.
type Bar struct {
	T     string    `json:"t"`
	Slots []BarSlot `json:"slots"`
	// Also are the styles the character has learned but cannot use with what
	// they hold, by weapon line, for the styles window.
	Also []BarSlot `json:"also,omitempty"`
}

// BarSet puts the style with an ID into a bar slot, swapping with what was there.
type BarSet struct {
	T    string `json:"t"`
	Slot int    `json:"slot"`
	ID   int    `json:"id"`
}

// Decode reads a message's type and, for those the server accepts, its body.
// An unknown type returns ("", nil, nil): ignored, not an error.
func Decode(data []byte) (string, interface{}, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return "", nil, fmt.Errorf("protocol: not JSON: %w", err)
	}
	var v interface{}
	switch env.T {
	case THello:
		v = &Hello{}
	case TInput:
		v = &Input{}
	case TChat:
		v = &Chat{}
	case TTarget:
		v = &Target{}
	case TAttack:
		v = &Attack{}
	case TStyle:
		v = &Style{}
	case TLoot:
		v = &Loot{}
	case TEquip:
		v = &Equip{}
	case TUse:
		v = &Use{}
	case TInteract:
		v = &Interact{}
	case TBuy:
		v = &Buy{}
	case TSell:
		v = &Sell{}
	case TDelve:
		v = &Delve{}
	case TRepair:
		v = &Repair{}
	case TBarSet:
		v = &BarSet{}
	case TGM:
		v = &GM{}
	case TTrain:
		v = &Train{}
	case TRestore:
		v = &Restore{}
	case TGroupOp:
		v = &GroupOp{}
	case TCast:
		v = &Cast{}
	case TRezAnswer:
		v = &RezAnswer{}
	case TGroundTarget:
		v = &GroundTarget{}
	case TCraftOp:
		v = &CraftOp{}
	case TQuestOp:
		v = &QuestOp{}
	case TTradeOp:
		v = &TradeOp{}
	case TConcCancel:
		v = &ConcCancel{}
	case TPetCmd:
		v = &PetCmd{}
	case TShoot:
		v = &Shoot{}
	case TUseItem:
		v = &UseItem{}
	case TAskTask:
		v = &AskTask{}
	case TAbandonTask:
		return TAbandonTask, nil, nil
	case TPray:
		return TPray, nil, nil
	case TPing:
		return TPing, nil, nil
	case TRelease:
		return TRelease, nil, nil
	default:
		return "", nil, nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return env.T, nil, fmt.Errorf("protocol: bad %s: %w", env.T, err)
	}
	return env.T, v, nil
}

// EffectView is one lasting effect on the character.
type EffectView struct {
	Name string  `json:"nm"`
	Kind string  `json:"k"`            // "dot", "slow", "stun", "mez", "proc", "buff", "dmgadd", "speed", "regen", "resist", "sick", "haste", "shield", "endregen", "powregen": what it does, which picks its icon
	Icon int     `json:"ic"`           // the spell's icon number in the data; not drawn, as icon numbers do not match this install
	Left float64 `json:"left"`         // seconds to run
	Of   float64 `json:"of"`           // seconds it lasts in all
	Conc bool    `json:"cc,omitempty"` // a concentration buff: no timer, left and of are 0
}

// Effects is the character's own lasting effects, sent whenever they change,
// with the speed factor they leave (1 free, 0 held), which the client's
// predicted movement applies.
type Effects struct {
	T     string       `json:"t"`
	List  []EffectView `json:"list"`
	Speed float64      `json:"speed"`
	Held  bool         `json:"held,omitempty"` // stunned or mesmerized: no moving, swinging or styles
	Song  int          `json:"song,omitempty"` // the spell id of the song or chant you have on; absent when none
	Spr   bool         `json:"spr,omitempty"`  // sprint is on: prediction runs faster
}

// Group operations a client may ask for.
const (
	GroupInvite  = "invite"  // invite the player with id ID, or else named Name
	GroupAccept  = "accept"  // accept the pending invitation
	GroupDecline = "decline" // decline it
	GroupLeave   = "leave"   // leave the group (/disband)
	GroupRemove  = "remove"  // the leader removes the member named Name
)

// GroupOp asks for something to be done about groups.
type GroupOp struct {
	T    string `json:"t"`
	Op   string `json:"op"`
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Conc is a caster's concentration: the pool (max), what their buffs hold
// (used) and each buff held, sent whenever any of it changes.
type Conc struct {
	T    string     `json:"t"`
	Max  int        `json:"max"`
	Used int        `json:"used"`
	List []ConcView `json:"list"`
}

// ConcView is one buff a caster holds: its id (for cancelling), the spell, who
// it is on and what it costs.
type ConcView struct {
	ID     int    `json:"id"`
	Spell  int    `json:"sp"` // the spell's id, to mark it on the spell bar
	Name   string `json:"nm"`
	Target string `json:"tg"`
	Cost   int    `json:"cost"`
}

// ConcCancel lets go of the buff with this id.
type ConcCancel struct {
	T  string `json:"t"`
	ID int    `json:"id"`
}

// AskTask talks to the NPC with this entity id about tasks: a giver offers one, the
// giver of a finished kill task pays, and the receiver of a delivery takes the item.
type AskTask struct {
	T  string `json:"t"`
	ID int64  `json:"id"`
}

// UseItem uses the item in a pack slot; a poison is put on a worn weapon.
type UseItem struct {
	T    string `json:"t"`
	Slot int    `json:"slot"`
}

// Delve asks for the details of one thing, as right-click does in the stock
// client. From says where it is shown: "inv" (a pack, worn or vault slot, by
// Slot), "shop" (an item of the open shop, by ID), "loot" (a corpse's item,
// by Corpse and Slot as its index), "spell" or "style" (by ID).
type Delve struct {
	T      string `json:"t"`
	From   string `json:"from"`
	Slot   int    `json:"slot,omitempty"`
	ID     string `json:"id,omitempty"`
	Corpse int64  `json:"corpse,omitempty"`
}

// DelveInfo is what a delve found. Con is true for a piece of gear: the client
// colours its name by Level against the player's own, as it does a creature's.
type DelveInfo struct {
	T     string   `json:"t"`
	Name  string   `json:"nm"`
	Level int      `json:"lv,omitempty"`
	Con   bool     `json:"con,omitempty"`
	Lines []string `json:"ln"`
}

// Shoot is the archer's bow order. Mode "" draws at the held target and keeps
// shooting until told to stop; "critical" draws one Critical Shot; "rapid"
// turns Rapid Fire on or off; "stop" lowers the bow.
type Shoot struct {
	T    string `json:"t"`
	Mode string `json:"mode,omitempty"`
}

// PetCmd is an order to the sender's pet: "attack" (their target), "follow",
// "stay", "release", or a stance, "passive", "defensive" or "aggressive".
type PetCmd struct {
	T   string `json:"t"`
	Cmd string `json:"cmd"`
}

// Task is the sender's task, sent when it is taken, changes, is finished or
// lost, and at login. On false means none is held.
type Task struct {
	T      string `json:"t"`
	On     bool   `json:"on"`
	Kind   string `json:"k,omitempty"`    // kill or delivery
	Name   string `json:"nm,omitempty"`   // "Kill Task" or "Delivery Task"
	Text   string `json:"tx,omitempty"`   // what to do next
	Giver  string `json:"gv,omitempty"`   // who gave it
	Done   bool   `json:"dn,omitempty"`   // a kill task whose creature has been killed
	Left   int    `json:"left,omitempty"` // seconds until it expires
	Finish int    `json:"n"`              // tasks finished at this level
	Max    int    `json:"max"`            // and the most the level allows
}

// Pet is the sender's own pet, sent when it appears, changes or is lost.
// Closed says it is gone and the window should close.
type Pet struct {
	T       string       `json:"t"`
	ID      int64        `json:"id,omitempty"`
	Name    string       `json:"nm,omitempty"`
	Level   int          `json:"lv,omitempty"`
	HP      float64      `json:"hp"` // 0..1
	Stance  string       `json:"st,omitempty"`
	FX      []EffectView `json:"fx,omitempty"` // the pet's lasting effects, the same as yours; the client counts their time down
	Mode    string       `json:"md,omitempty"` // follow, stay or attack
	Minions []PetMinion  `json:"mn,omitempty"`
	Closed  bool         `json:"cl,omitempty"`
}

// PetMinion is one of a commander's minions.
type PetMinion struct {
	ID    int64   `json:"id"`
	Name  string  `json:"nm"`
	Level int     `json:"lv"`
	HP    float64 `json:"hp"`
}

// RezOffer tells a dead character that a resurrection is offered, by whom and
// with how many percent of their hit points and power, and for how many
// seconds it stands. Secs 0 withdraws the offer (it lapsed, or they released).
type RezOffer struct {
	T      string  `json:"t"`
	From   string  `json:"from,omitempty"`
	Health int     `json:"health,omitempty"`
	Power  int     `json:"power,omitempty"`
	Secs   float64 `json:"secs"`
}

// RezAnswer accepts or declines the pending resurrection offer.
type RezAnswer struct {
	T      string `json:"t"`
	Accept bool   `json:"accept"`
}

// GroupInv tells a player they are invited to a group, by whom, and how long
// the invitation stands, in seconds.
type GroupInv struct {
	T    string  `json:"t"`
	From string  `json:"from"`
	Secs float64 `json:"secs"`
}

// GroupMember is what a member sees of another (or of themselves): the
// leader comes first, then the members in the order they joined.
type GroupMember struct {
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Class   string `json:"class"`
	HP      int    `json:"hp"`  // percent
	End     int    `json:"end"` // percent
	Zone    int    `json:"zone"`
	ZoneN   string `json:"zoneName,omitempty"`
	InRange bool   `json:"inRange"` // in the same zone, within 6,000 units
	Dead    bool   `json:"dead,omitempty"`
	Gone    bool   `json:"gone,omitempty"` // not in the game right now
	You     bool   `json:"you,omitempty"`
}

// Group is a player's group as it stands; an empty list means no group.
type Group struct {
	T       string        `json:"t"`
	Members []GroupMember `json:"members"`
}

// Cast asks to cast a spell from the spellbook at an entity; without a target
// the one the player holds is used, and a spell on the caster needs none.
type Cast struct {
	T      string `json:"t"`
	Spell  int    `json:"sp"`
	Target int64  `json:"tg,omitempty"`
}

// GuildMember is one line of the guild window.
type GuildMember struct {
	Name  string `json:"nm"`
	Rank  int    `json:"rk"`
	Note  string `json:"nt,omitempty"`
	On    bool   `json:"on,omitempty"`
	Level int    `json:"lv,omitempty"` // online members only
	Class string `json:"cl,omitempty"`
}

// GuildInfo is a player's guild as the guild window shows it. Name is empty
// for a player in no guild. Perms is what your rank may do, as bits (see
// docs/protocol.md).
type GuildInfo struct {
	T        string        `json:"t"`
	Name     string        `json:"nm,omitempty"`
	Rank     int           `json:"rk"`
	Title    string        `json:"ti,omitempty"`
	Motd     string        `json:"motd,omitempty"`
	Omotd    string        `json:"omotd,omitempty"` // the officers' message, for ranks that hear officer chat
	Emblem   int           `json:"em,omitempty"`
	Alliance string        `json:"al,omitempty"` // the leading guild of your alliance
	Perms    int           `json:"pm"`
	Members  []GuildMember `json:"mem"`
}

// DuelAsk tells a player who challenged them.
type DuelAsk struct {
	T    string `json:"t"`
	From string `json:"from"`
	Secs int    `json:"secs"` // how long the challenge waits
}

// Friend is one entry of the friends list.
type Friend struct {
	Name string `json:"nm"`
	On   bool   `json:"on,omitempty"` // in the game now
}

// Friends is a player's friends list, sent when it changes or a friend comes or goes.
type Friends struct {
	T    string   `json:"t"`
	List []Friend `json:"list"`
}

// TradeOp is a request in a trade. Open names the player (ID, or your target
// when 0); add and remove name a pack Slot; money sets your offered Copper;
// accept and cancel need nothing.
type TradeOp struct {
	T      string `json:"t"`
	Op     string `json:"op"`
	ID     int64  `json:"id,omitempty"`
	Slot   int    `json:"slot,omitempty"`
	Copper int64  `json:"cu,omitempty"`
}

// TradeSide is one side of a trade: the items offered and the money.
type TradeSide struct {
	Items []InvItem `json:"items"`
	Money int64     `json:"mo"`
	OK    bool      `json:"ok,omitempty"` // this side has accepted
}

// Trade is a trade as one of its two players sees it: Mine and Theirs.
type Trade struct {
	T      string    `json:"t"`
	With   string    `json:"with"`
	Mine   TradeSide `json:"mine"`
	Theirs TradeSide `json:"theirs"`
}

// TradeClose says the trade is over: Done when it went through, else Why.
type TradeClose struct {
	T    string `json:"t"`
	Done bool   `json:"done,omitempty"`
	Why  string `json:"why,omitempty"`
}

// GroundTarget sets the player's ground target to a point, or clears it when
// Clear is set.
type GroundTarget struct {
	T     string  `json:"t"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Clear bool    `json:"clear,omitempty"`
}

// QuestOp is a quest request: Op is "accept" or "decline" (the offered quest,
// ID), "abandon" (ID), "choose" (a reward item, ID and N, an index of the
// offered choices) or "journal".
type QuestOp struct {
	T  string `json:"t"`
	Op string `json:"op"`
	ID string `json:"id,omitempty"`
	N  int    `json:"n,omitempty"`
}

// QuestLine is one active quest in the journal: its current step's text and,
// for a counted step, how many you have of how many it needs.
type QuestLine struct {
	ID    string `json:"id"`
	Title string `json:"nm"`
	Text  string `json:"text"`
	Have  int    `json:"have,omitempty"`
	Need  int    `json:"need,omitempty"`
}

// Quests is the journal.
type Quests struct {
	T      string      `json:"t"`
	Active []QuestLine `json:"active"`
	Done   []string    `json:"done"` // titles
}

// QuestOffer is a quest an NPC offers you; answer with questop accept or decline.
type QuestOffer struct {
	T      string `json:"t"`
	ID     string `json:"id"`
	Title  string `json:"nm"`
	NPC    string `json:"npc"`
	Text   string `json:"text"`
	Reward string `json:"reward,omitempty"`
}

// QuestChoice is one item offered as a reward.
type QuestChoice struct {
	N    int    `json:"n"`
	ID   string `json:"id"`
	Name string `json:"nm"`
}

// QuestReward offers a choice of items for a finished quest; answer with
// questop choose.
type QuestReward struct {
	T      string        `json:"t"`
	ID     string        `json:"id"`
	Title  string        `json:"nm"`
	Choose []QuestChoice `json:"choose"`
}

// CraftOp is a crafting request: Op is "open", "join" (the profession of the
// master you stand by), "make" (the recipe of product ID) or "cancel".
type CraftOp struct {
	T  string `json:"t"`
	Op string `json:"op"`
	ID string `json:"id,omitempty"`
}

// CraftSkill is one tradeskill of the crafter: its number, name, value and the
// most it can reach now.
type CraftSkill struct {
	ID    int    `json:"id"`
	Name  string `json:"nm"`
	Value int    `json:"v"`
	Cap   int    `json:"cap"`
}

// CraftMat is one material of a recipe and how many the crafter holds.
type CraftMat struct {
	ID   string `json:"id"`
	Name string `json:"nm"`
	N    int    `json:"n"`
	Have int    `json:"have"`
}

// CraftRecipe is one thing the crafter can try to make, with the chance it
// works and how long it takes.
type CraftRecipe struct {
	ID     string     `json:"id"` // the product's item id
	Name   string     `json:"nm"`
	Skill  int        `json:"sk"`
	Level  int        `json:"lv"`
	Chance int        `json:"ok"` // percent
	Secs   int        `json:"s"`
	Mats   []CraftMat `json:"m"`
}

// Craft is the crafting window. Prof is the primary profession's skill number
// (0 before you join one); Offer is the profession the master you stand by
// teaches, when you have none.
type Craft struct {
	T         string        `json:"t"`
	Prof      int           `json:"pr,omitempty"`
	ProfName  string        `json:"pn,omitempty"`
	Offer     int           `json:"offer,omitempty"`
	OfferName string        `json:"on,omitempty"`
	Skills    []CraftSkill  `json:"skills"`
	Recipes   []CraftRecipe `json:"recipes"`
}

// CraftProg says a craft began and how many seconds it takes.
type CraftProg struct {
	T    string `json:"t"`
	ID   string `json:"id"`
	Secs int    `json:"s"`
}

// CraftDone ends a craft. OK is a made item; Q its quality; Gain a point of
// skill won; Text the line for the log.
type CraftDone struct {
	T    string `json:"t"`
	ID   string `json:"id"`
	OK   bool   `json:"ok"`
	Q    int    `json:"q,omitempty"`
	Gain bool   `json:"gain,omitempty"`
	Text string `json:"text"`
}

// Book is the spellbook: the spells the character knows, by spec line.
type Book struct {
	T     string     `json:"t"`
	Lines []BookLine `json:"lines"`
}

// BookLine is one spec line's spells, lowest level first.
type BookLine struct {
	Spec   string      `json:"spec"`
	Name   string      `json:"nm"`
	Spells []BookSpell `json:"sp"`
}

// BookSpell is one spell. OK is false for a type the server cannot cast yet.
type BookSpell struct {
	ID     int     `json:"id"`
	Name   string  `json:"nm"`
	Level  int     `json:"lv"`
	Cost   int     `json:"pw"` // power it draws
	Cast   float64 `json:"ct"` // seconds to cast
	Range  int     `json:"rg"`
	Type   string  `json:"tp"`
	Target string  `json:"tg"`
	OK     bool    `json:"ok"`
	Icon   int     `json:"ic"`           // the data's number; not drawn
	Recast int     `json:"rc,omitempty"` // seconds before it can be cast again
	Pulse  bool    `json:"pl,omitempty"` // a song or chant: casting it turns it on, and casting it again off
	HP     bool    `json:"hp,omitempty"` // the cost is hit points, not power (a Savage spell)
}

// Recast tells a caster that a spell they cast cannot be cast again for Secs
// seconds.
type Recast struct {
	T    string  `json:"t"`
	ID   int     `json:"id"`
	Secs float64 `json:"s"`
}
