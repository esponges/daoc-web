package importer

import (
	"fmt"
	"sort"
	"strings"
)

// Added merchant stock (midgard-to-fifty, part B). The community data has no
// merchant gear above level 30, so this project's own table extends a few of
// the data's Midgard merchant lists with generic gear for levels 31 to 50. The
// data is not edited: the extra rows are written beside it, marked with
// AddedReason, and chosen by the rule in AddedStock.

// AddedReason is written beside every row this table adds.
const AddedReason = "levelling gear 31-50 (project table, midgard-to-fifty)"

// Added stock covers these levels, in bands of five.
const (
	AddedFrom = 31
	AddedTo   = 50
)

// addedLists names, for each object type of Midgard gear, the data's merchant
// list the gear is added to. They are Jordheim's: Jordheim sells each kind
// already, so a player knows where to look, and the lists the frontier towns
// share with it carry the stock along.
var addedLists = map[int]string{
	8:  "MidStaff",
	11: "MidWeaponMerchant",
	12: "MidWeaponMerchant",
	13: "MidWeaponMerchant",
	14: "MidSpears",
	15: "MidBows",
	16: "MidThrowingWeapons",
	17: "MidAxesClaws",
	25: "MidAxesClaws",
	32: "MidClothPadded",
	33: "MidLeatherMjuklaedar",
	34: "MidStuddedMerchant",
	35: "32a8b069-1f72-469c-af9a-189fdadf4d45",
	42: "MidShields",
}

// AddedKinds are the object types the table covers, sorted.
func AddedKinds() []int {
	var out []int
	for k := range addedLists {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// AddedList is the merchant list an object type's added gear goes to, or "".
func AddedList(objectType int) string { return addedLists[objectType] }

// merchantQuality is the quality the data's own merchant gear has.
const merchantQuality = 85

// StockCandidate is the part of an item template the rule looks at.
type StockCandidate struct {
	ID         string
	Name       string
	Level      int
	Quality    int
	Price      int
	Realm      int
	ObjectType int
	ItemType   int
	Hand       int
	Plain      bool // no bonus level and no stat, skill or resist bonus
	Classes    string
}

// CandidateOf reads a candidate from an itemtemplate row of the dump.
func CandidateOf(r *Row) StockCandidate {
	plain := r.Int("Bonus") == 0 && r.Int("ProcSpellID") == 0
	for i := 1; i <= 10 && plain; i++ {
		if r.Int(fmt.Sprintf("Bonus%dType", i)) != 0 && r.Int(fmt.Sprintf("Bonus%d", i)) != 0 {
			plain = false
		}
	}
	return StockCandidate{
		ID: r.Str("Id_nb"), Name: r.Str("Name"), Level: r.Int("Level"), Quality: r.Int("Quality"),
		Price: r.Int("Price"), Realm: r.Int("Realm"), ObjectType: r.Int("Object_Type"),
		ItemType: r.Int("Item_Type"), Hand: r.Int("Hand"), Plain: plain, Classes: r.Str("AllowedClasses"),
	}
}

// StockLine is one item added to one merchant list.
type StockLine struct {
	List string
	Item string
}

// eligible says whether an item may be added: a generic Midgard or neutral
// piece of a covered kind in the levels 31 to 50 — it sells for something, has
// no bonuses, is not locked to a class, and is not a named item (the data
// writes generic items in lower case, and named ones, such as a drop's, with
// capitals).
func (c StockCandidate) eligible() bool {
	if c.Level < AddedFrom || c.Level > AddedTo || c.Price <= 0 || c.ItemType == 0 || !c.Plain {
		return false
	}
	if c.Realm != 0 && c.Realm != 2 {
		return false
	}
	if _, ok := addedLists[c.ObjectType]; !ok {
		return false
	}
	if (c.Classes != "" && c.Classes != "0") || c.Name != strings.ToLower(c.Name) || c.Quality < merchantQuality {
		return false
	}
	return true
}

// groupKey groups interchangeable gear: the same kind of piece in the same
// band of five levels.
type groupKey struct {
	obj, slot, hand, band int
}

func keyOf(c StockCandidate) groupKey {
	return groupKey{c.ObjectType, c.ItemType, min(c.Hand, 1), (c.Level - AddedFrom) / 5}
}

// AddedStock picks the items to add. For each kind of piece and each band of
// five levels it takes the eligible items of the merchant quality (85) and,
// where the data has none of that quality in the band, the lowest quality it
// has. The result is ordered by list, then by kind, level and id, so a
// re-import adds the same rows.
func AddedStock(items []StockCandidate) []StockLine {
	groups := map[groupKey][]StockCandidate{}
	for _, c := range items {
		if c.eligible() {
			k := keyOf(c)
			groups[k] = append(groups[k], c)
		}
	}
	var picked []StockCandidate
	for _, g := range groups {
		best := g[0].Quality
		for _, c := range g {
			best = min(best, c.Quality)
		}
		for _, c := range g {
			if c.Quality == best {
				picked = append(picked, c)
			}
		}
	}
	sort.Slice(picked, func(i, j int) bool {
		a, b := picked[i], picked[j]
		if la, lb := addedLists[a.ObjectType], addedLists[b.ObjectType]; la != lb {
			return la < lb
		}
		if a.ObjectType != b.ObjectType {
			return a.ObjectType < b.ObjectType
		}
		if a.Level != b.Level {
			return a.Level < b.Level
		}
		return a.ID < b.ID
	})
	out := make([]StockLine, 0, len(picked))
	seen := map[StockLine]bool{}
	for _, c := range picked {
		l := StockLine{addedLists[c.ObjectType], c.ID}
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	return out
}

// StockPage is where an added row sits on the list's pages: after the data's
// own last page, thirty to a page.
const stockPerPage = 30

// PlaceStock gives each added line a page and slot after the last page the
// data's list uses. last maps a list to its highest page (-1 or absent: the
// list has no rows and is skipped, since an added row would make a merchant of
// nothing).
func PlaceStock(lines []StockLine, last map[string]int) []PlacedStock {
	n := map[string]int{}
	var out []PlacedStock
	for _, l := range lines {
		lp, ok := last[l.List]
		if !ok {
			continue
		}
		i := n[l.List]
		n[l.List]++
		out = append(out, PlacedStock{l.List, l.Item, lp + 1 + i/stockPerPage, i % stockPerPage})
	}
	return out
}

// PlacedStock is an added line with its page and slot.
type PlacedStock struct {
	List, Item string
	Page, Slot int
}
