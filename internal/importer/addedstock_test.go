package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func piece(id string, level, quality, obj, slot int) StockCandidate {
	return StockCandidate{ID: id, Name: id, Level: level, Quality: quality, Price: 1000, Realm: 2,
		ObjectType: obj, ItemType: slot, Plain: true}
}

func TestAddedStockRule(t *testing.T) {
	good := piece("steel_sword", 33, 85, 11, 10)
	cases := []struct {
		name string
		c    StockCandidate
		want bool
	}{
		{"generic piece", good, true},
		{"below the band", func() StockCandidate { c := good; c.Level = 30; return c }(), false},
		{"above the band", func() StockCandidate { c := good; c.Level = 51; return c }(), false},
		{"has a bonus", func() StockCandidate { c := good; c.Plain = false; return c }(), false},
		{"named item", func() StockCandidate { c := good; c.Name = "Baln's Cleaver"; return c }(), false},
		{"class locked", func() StockCandidate { c := good; c.Classes = "22;"; return c }(), false},
		{"no price", func() StockCandidate { c := good; c.Price = 0; return c }(), false},
		{"Albion", func() StockCandidate { c := good; c.Realm = 1; return c }(), false},
		{"neutral", func() StockCandidate { c := good; c.Realm = 0; return c }(), true},
		{"poor quality", func() StockCandidate { c := good; c.Quality = 70; return c }(), false},
		{"not gear", func() StockCandidate { c := good; c.ObjectType = 41; return c }(), false},
	}
	for _, tc := range cases {
		got := AddedStock([]StockCandidate{tc.c})
		if (len(got) == 1) != tc.want {
			t.Errorf("%s: added %v, want added=%v", tc.name, got, tc.want)
		}
	}
}

// Merchant quality wins where the band has it; else the lowest there is.
func TestAddedStockQuality(t *testing.T) {
	got := AddedStock([]StockCandidate{
		piece("a85", 33, 85, 11, 10), piece("a100", 34, 100, 11, 10),
		piece("b100", 38, 100, 11, 10), piece("b99", 38, 99, 11, 10),
	})
	var ids []string
	for _, l := range got {
		ids = append(ids, l.Item)
	}
	if !slices.Equal(ids, []string{"a85", "b99"}) {
		t.Errorf("picked %v, want [a85 b99]", ids)
	}
	if got[0].List != "MidWeaponMerchant" {
		t.Errorf("a sword goes to %q", got[0].List)
	}
}

func TestPlaceStock(t *testing.T) {
	var lines []StockLine
	for i := 0; i < 65; i++ {
		lines = append(lines, StockLine{"L", fmt.Sprint("i", i)})
	}
	lines = append(lines, StockLine{"Absent", "x"})
	got := PlaceStock(lines, map[string]int{"L": 2})
	if len(got) != 65 {
		t.Fatalf("placed %d, want 65 (a list with no rows is skipped)", len(got))
	}
	if got[0].Page != 3 || got[0].Slot != 0 || got[29].Page != 3 || got[29].Slot != 29 || got[30].Page != 4 || got[64].Page != 5 {
		t.Errorf("pages and slots %+v %+v %+v", got[0], got[30], got[64])
	}
}

// Against the real community data: every band from 31 to 50 has something for
// each kind of gear that the data has at all, and a re-run picks the same rows.
func TestAddedStockCoversEveryBand(t *testing.T) {
	dump := filepath.Join("..", "..", "server", "data", "community", "opendaoc-db-core", "itemtemplate.sql")
	if _, err := os.Stat(dump); err != nil {
		t.Skip("no community database checked out in server/data/community")
	}
	var all []StockCandidate
	if _, err := ReadTable(dump, func(r *Row) error { all = append(all, CandidateOf(r)); return nil }); err != nil {
		t.Fatal(err)
	}
	lines := AddedStock(all)
	if again := AddedStock(all); !slices.Equal(lines, again) {
		t.Error("two runs picked different rows")
	}
	byID := map[string]StockCandidate{}
	for _, c := range all {
		byID[c.ID] = c
	}
	got := map[[2]int]bool{} // object type, band
	for _, l := range lines {
		c := byID[l.Item]
		if l.List != AddedList(c.ObjectType) {
			t.Errorf("%s went to %s", c.ID, l.List)
		}
		got[[2]int{c.ObjectType, (c.Level - AddedFrom) / 5}] = true
	}
	// The kinds a Warrior wears and wields must have every band; the others
	// are reported when the data has the kind in a band and nothing was added.
	for _, obj := range []int{11, 12, 13, 14, 15, 32, 33, 34, 35, 42} {
		for band := 0; band < 4; band++ {
			if !got[[2]int{obj, band}] {
				t.Errorf("object type %d has nothing added for levels %d-%d", obj, AddedFrom+5*band, AddedFrom+5*band+4)
			}
		}
	}
}
