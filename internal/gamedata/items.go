package gamedata

import (
	"path/filepath"
	"strings"

	"daocweb/internal/mpak"
)

// Kinds of held item, read from an items.csv model's file name.
const (
	KindShield   = "shield"
	KindRanged   = "ranged"
	KindOneHand  = "onehand"
	KindTwoHand  = "twohand"
	KindMelee    = "melee" // a hand weapon whose name does not say one or two hands
	KindUnknown  = ""
	slotRight    = 10
	slotLeft     = 11
	slotTwoHand  = 12
	slotRanged   = 13
	maxItemModel = 1 << 20
)

// Item is one row of items.csv: a model number the game server can name, and
// the .nif in items/ that draws it.
type Item struct {
	ID   int
	Name string // the "textual name" column
	NIF  string // file name in items/, no extension
	Kind string // see the Kind constants; "" when the name does not say
}

// ParseItems builds the model-number table from items.csv's data rows
// (column 0 the number, 1 the textual name, 2 the .nif name).
func ParseItems(rows [][]string) map[int]Item {
	out := map[int]Item{}
	for _, r := range rows {
		id := num(r, 0)
		nif := col(r, 2)
		if id <= 0 || id > maxItemModel || nif == "" {
			continue
		}
		out[id] = Item{ID: id, Name: col(r, 1), NIF: nif, Kind: ItemKind(nif)}
	}
	return out
}

// Items reads items.csv from the install at game.
func Items(game string) (map[int]Item, error) {
	a, err := mpak.Open(filepath.Join(game, "gamedata.mpk"))
	if err != nil {
		return nil, err
	}
	rows, _, err := table(a, "items.csv")
	if err != nil {
		return nil, err
	}
	return ParseItems(rows), nil
}

// ItemKind classifies a model by its file name. The names follow a pattern:
// an optional realm letter, then a token for the weapon line (cr crush, sl
// slash, th thrust, bl blunt, st staff, po polearm, sc scythe, bo bow, hth
// hand-to-hand), or an explicit 1h or 2h, or "shield". Only whole tokens
// count, so props such as "statue" never read as a staff.
func ItemKind(nif string) string {
	toks := strings.Split(strings.ToLower(nif), "_")
	has := func(want ...string) bool {
		for _, t := range toks {
			for _, w := range want {
				if t == w {
					return true
				}
			}
		}
		return false
	}
	switch {
	case has("shield", "sh"):
		return KindShield
	case has("bo", "bow") || strings.Contains(strings.ToLower(nif), "bow"):
		return KindRanged
	case has("2h", "st", "po", "sc"):
		return KindTwoHand
	case has("1h"):
		return KindOneHand
	case has("cr", "sl", "th", "hth", "bl"):
		return KindMelee
	}
	return KindUnknown
}

// KindFitsSlot reports whether an item of this kind may be drawn for the
// database's equipment slot: right hand 10, left hand 11, two-hand 12,
// ranged 13. The database does put two-handers in the right hand, so slot 10
// takes them; a shield belongs only in the left hand, and a bow only in 13.
func KindFitsSlot(kind string, slot int) bool {
	switch slot {
	case slotRight:
		return kind == KindOneHand || kind == KindTwoHand || kind == KindMelee
	case slotLeft:
		return kind == KindShield || kind == KindOneHand || kind == KindMelee
	case slotTwoHand:
		return kind == KindTwoHand || kind == KindMelee
	case slotRanged:
		return kind == KindRanged
	}
	return false
}

// Tints reads tints.csv: the RGB the game tints a piece of equipment with,
// by the colour number the database stores. Columns 2 to 4 are red, green and
// blue (the later columns repeat them for expansions).
func Tints(game string) (map[int][3]int, error) {
	a, err := mpak.Open(filepath.Join(game, "gamedata.mpk"))
	if err != nil {
		return nil, err
	}
	rows, _, err := table(a, "tints.csv")
	if err != nil {
		return nil, err
	}
	return ParseTints(rows), nil
}

// ParseTints is Tints over already-read rows.
func ParseTints(rows [][]string) map[int][3]int {
	out := map[int][3]int{}
	for _, r := range rows {
		out[num(r, 0)] = [3]int{num(r, 2), num(r, 3), num(r, 4)}
	}
	return out
}

// PlayerSkins reads pskins.csv and returns the lower-cased texture names
// without their extension ("b_plt1_body1m"), which is what the armour
// families are made of.
func PlayerSkins(game string) (map[string]bool, error) {
	a, err := mpak.Open(filepath.Join(game, "gamedata.mpk"))
	if err != nil {
		return nil, err
	}
	rows, _, err := table(a, "pskins.csv")
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, r := range rows {
		n := strings.ToLower(col(r, 2))
		if i := strings.LastIndexByte(n, '.'); i > 0 {
			n = n[:i]
		}
		if n != "" {
			out[n] = true
		}
	}
	return out, nil
}
