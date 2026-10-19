package main

import (
	"fmt"
	"strings"

	"daocweb/internal/importer"
)

// GearRow is one five-level band: for each kind of Warrior gear, how many
// Midgard merchants (placements) sell at least one piece of it at a level in
// the band.
type GearRow struct {
	Band  string         `json:"band"`
	Kinds map[string]int `json:"kinds"`
}

// gearKinds are the columns of the levelling-gear table, in order.
var gearKinds = []string{"1h axe", "1h hammer", "1h sword", "2h axe", "2h hammer", "2h sword", "shield", "studded", "chain"}

// gearNeeded says whether a Warrior needs a kind in a band: studded armour
// until level 10, chain from it, and every weapon and the shield throughout.
func gearNeeded(kind string, lo int) bool {
	switch kind {
	case "studded":
		return lo < 10
	case "chain":
		return lo+4 >= 10
	}
	return true
}

// gearKind names the column a Midgard item falls in, or "" for none.
func gearKind(r *importer.Row) string {
	if realm := r.Int("Realm"); realm != 0 && realm != 2 {
		return ""
	}
	obj, slot := r.Int("Object_Type"), r.Int("Item_Type")
	two := "1h "
	if slot == 12 || r.Int("Hand") == 1 {
		two = "2h "
	}
	switch obj {
	case 11:
		return two + "sword"
	case 12:
		return two + "hammer"
	case 13:
		return two + "axe"
	case 42:
		return "shield"
	case 34:
		return "studded"
	case 35:
		return "chain"
	}
	return ""
}

// levellingGear reads which Warrior gear Midgard merchants sell, by band.
func levellingGear(read func(string, func(*importer.Row) error) error) ([]GearRow, error) {
	merchants := map[string]int{} // stock list -> placements selling it
	if err := read("mob", func(r *importer.Row) error {
		if l := r.Str("ItemsListTemplateID"); l != "" && midgardRegions[r.Int("Region")] {
			merchants[l]++
		}
		return nil
	}); err != nil {
		return nil, err
	}
	type piece struct {
		kind  string
		level int
	}
	items := map[string]piece{}
	if err := read("itemtemplate", func(r *importer.Row) error {
		if k := gearKind(r); k != "" {
			items[r.Str("Id_nb")] = piece{k, r.Int("Level")}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sells := map[string]map[string]bool{} // band/kind -> stock lists
	if err := read("merchantitem", func(r *importer.Row) error {
		l := r.Str("ItemListID")
		it, ok := items[r.Str("ItemTemplateID")]
		if merchants[l] == 0 || !ok || it.level < 1 || it.level > 50 {
			return nil
		}
		k := band(it.level) + "/" + it.kind
		if sells[k] == nil {
			sells[k] = map[string]bool{}
		}
		sells[k][l] = true
		return nil
	}); err != nil {
		return nil, err
	}
	var out []GearRow
	for lo := 1; lo <= 46; lo += 5 {
		row := GearRow{Band: band(lo), Kinds: map[string]int{}}
		for _, k := range gearKinds {
			for l := range sells[row.Band+"/"+k] {
				row.Kinds[k] += merchants[l]
			}
		}
		out = append(out, row)
	}
	return out, nil
}

func band(level int) string {
	lo := (level-1)/5*5 + 1
	return fmt.Sprintf("%d-%d", lo, lo+4)
}

// gearTable writes the levelling-gear table: merchants selling each kind in
// each band, "none" where a Warrior needs it and nobody sells it, and "-"
// where a Warrior does not need it.
func gearTable(b *strings.Builder, rows []GearRow) {
	b.WriteString("\n## Levelling gear\n\nMidgard merchants (placements) selling Warrior gear in each five-level band. ")
	b.WriteString("**none** marks a gap: a band where a Warrior needs that kind and no merchant sells it. Gaps are reported, not filled.\n\n")
	b.WriteString("| band | " + strings.Join(gearKinds, " | ") + " |\n|" + strings.Repeat(" --- |", len(gearKinds)+1) + "\n")
	for _, r := range rows {
		var lo int
		fmt.Sscanf(r.Band, "%d", &lo)
		cells := []string{r.Band}
		for _, k := range gearKinds {
			n := r.Kinds[k]
			switch {
			case n > 0:
				cells = append(cells, fmt.Sprint(n))
			case gearNeeded(k, lo):
				cells = append(cells, "**none**")
			default:
				cells = append(cells, "-")
			}
		}
		b.WriteString("| " + strings.Join(cells, " | ") + " |\n")
	}
}

// gearGaps counts the cells a Warrior needs that no merchant sells.
func gearGaps(rows []GearRow) (gaps int) {
	for _, r := range rows {
		var lo int
		fmt.Sscanf(r.Band, "%d", &lo)
		for _, k := range gearKinds {
			if r.Kinds[k] == 0 && gearNeeded(k, lo) {
				gaps++
			}
		}
	}
	return
}
