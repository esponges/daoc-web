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
	THello    = "hello"  // first message: protocol version and the character to play
	TInput    = "input"  // a run of fixed movement steps
	TChat     = "chat"   // say or broadcast
	TTarget   = "target" // select an entity
	TAttack   = "attack" // auto-attack on or off
	TStyle    = "style"  // use a style by its slot on the bar
	TLoot     = "loot"   // take an item from a corpse
	TEquip    = "equip"  // move an item between slots
	TUse      = "use"    // use a zone exit, etc.
	TGM       = "gm"     // an administrator's command
	TPing     = "ping"
	TRelease  = "release"  // a dead character asks to be stood up now rather than when the timer ends
	TInteract = "interact" // ask an NPC to do what it does (open a merchant's shop)
	TBuy      = "buy"      // buy an item from the open shop
	TSell     = "sell"     // sell a pack item to the open shop
	TTrain    = "train"    // raise a spec at the trainer
	TRestore  = "restore"  // pay a healer to restore lost constitution
	TPray     = "pray"     // pray at the targeted gravestone
	TGroupOp  = "groupop"  // invite, accept, decline, leave or remove from a group
)

// Message types, server to client.
const (
	TWelcome   = "welcome"
	TSnap      = "snap"
	TChatMsg   = "chatmsg"
	TErr       = "err"
	TZone      = "zone" // the character moved zone: reload at the given one
	TInv       = "inv"  // the character's inventory and equipment
	TCorpse    = "corpse"
	TPong      = "pong"
	TStyleBar  = "bar"
	TShop      = "shop"      // a merchant's stock, sent when interaction opens it
	TShopClose = "shopclose" // the open shop is closed (walked away, died, moved zone)
	TStats     = "stats"     // the character sheet
	TTrainer   = "trainer"   // a trainer's offer: the character's specs and their costs
	THealer    = "healer"    // a healer's offer to restore lost constitution
	TEffects   = "effects"   // the character's own lasting effects, and the speed they leave
	TGroup     = "group"     // the character's group: its members and how they fare
	TGroupInv  = "groupinv"  // an invitation to join a group
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
	FlagRun = 1 << iota // run rather than walk (backing and strafing always walk)
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
	Kind     string   `json:"ik,omitempty"` // what interacting does: "merchant", "trainer", "healer", "stable", "notclassic"; absent for nothing
	Con      *int     `json:"cc,omitempty"` // a creature's con colour to you, -3 grey to 3 purple
	Own      bool     `json:"own,omitempty"`
	Ride     bool     `json:"rd,omitempty"` // a player on a horse
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

// ShopClose says the open shop is gone.
type ShopClose struct {
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
}

// SpecRow is one spec on the character sheet or a trainer's list.
type SpecRow struct {
	Key   string `json:"key"`
	Name  string `json:"nm"`
	Level int    `json:"lv"`
	Bonus int    `json:"bonus,omitempty"` // from items, as combat counts it
	Cost  int    `json:"cost,omitempty"`  // of the next level, on a trainer's list
	Can   bool   `json:"can,omitempty"`   // the next level can be trained now
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
	Level     int       `json:"lv"`
	Half      bool      `json:"half,omitempty"` // past the midpoint of the level, from 40
	Base      [8]int    `json:"base"`           // strength, constitution, dexterity, quickness, intelligence, piety, empathy, charisma
	Stats     [8]int    `json:"stats"`          // with level gains, capped items and constitution lost
	MaxHP     int       `json:"mhp"`
	MaxEnd    int       `json:"mend"`
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
	case TGM:
		v = &GM{}
	case TTrain:
		v = &Train{}
	case TRestore:
		v = &Restore{}
	case TGroupOp:
		v = &GroupOp{}
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
	Kind string  `json:"k"`    // "dot", "slow", "stun", "mez", "proc": what it does, which picks its icon
	Icon int     `json:"ic"`   // the spell's icon number in the data; not drawn, as icon numbers do not match this install
	Left float64 `json:"left"` // seconds to run
	Of   float64 `json:"of"`   // seconds it lasts in all
}

// Effects is the character's own lasting effects, sent whenever they change,
// with the speed factor they leave (1 free, 0 held), which the client's
// predicted movement applies.
type Effects struct {
	T     string       `json:"t"`
	List  []EffectView `json:"list"`
	Speed float64      `json:"speed"`
	Held  bool         `json:"held,omitempty"` // stunned or mesmerized: no moving, swinging or styles
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
