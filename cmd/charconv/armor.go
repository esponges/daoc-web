package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"daocweb/internal/gamedata"
)

// The install does not say which art the database's armour numbers mean (see
// openspec/changes/archive/2026-10-19-npc-gear-mapping/design.md), so armour is dressed by tier.
// What is exact is the kind of armour each number is: the imported item table
// says model 47 is the legs of Albion plate. A piece therefore becomes a level,
// 0 cloth to 4 plate, and a realm, and the client dresses the NPC's matching
// slot in that level's tier mesh and a texture family of that realm.

// Armour slots in the database's numbering, and the shape slots each dresses.
var armourSlotNames = map[int]string{22: "gloves", 23: "boots", 25: "torso", 27: "legs", 28: "arms"}

// armourLevel maps the database's armour object types (32 cloth, 33 leather,
// 34 studded, 35 chain, 36 plate, 37 reinforced, 38 scale) to a level 0 to 4.
func armourLevel(objectType int) (int, bool) {
	switch objectType {
	case 32:
		return 0, true
	case 33:
		return 1, true
	case 34, 37:
		return 2, true
	case 35, 38:
		return 3, true
	case 36:
		return 4, true
	}
	return 0, false
}

// armourPiece is what a worn armour number is.
type armourPiece struct {
	Level int `json:"level"`
	Realm int `json:"realm,omitempty"` // 1 Albion, 2 Midgard, 3 Hibernia; 0 when the items using it name none
}

// armourSet names the texture file for each shape slot, under armor/.
type armourSet struct {
	Body   string `json:"body"`
	BodyF  string `json:"bodyF"`
	LBody  string `json:"lbody"`
	Arms   string `json:"arms"`
	Gloves string `json:"gloves"`
	Legs   string `json:"legs"`
	Boots  string `json:"boots"`
}

// armourData is the armour part of gear.json.
type armourData struct {
	Pieces   map[string]armourPiece `json:"pieces"`   // "slot:model"
	Textures map[string]armourSet   `json:"textures"` // "realm letter:level"
	Tints    map[string][3]int      `json:"tints"`    // colour number -> RGB
}

var realmLetter = map[int]string{1: "b", 2: "n", 3: "h"}

// familyCodes are the texture family names for a level, in the order a realm
// prefers them: Hibernia's third and fourth tiers are reinforced and scale,
// the other realms' studded and chain.
func familyCodes(realm string, level int) []string {
	switch level {
	case 0:
		return []string{"cth"}
	case 1:
		return []string{"lea"}
	case 2:
		if realm == "h" {
			return []string{"ren", "std"}
		}
		return []string{"std", "ren"}
	case 3:
		if realm == "h" {
			return []string{"scl", "chn"}
		}
		return []string{"chn", "scl"}
	case 4:
		return []string{"plt"}
	}
	return nil
}

var armourParts = []string{"body1m", "body1f", "lbody1", "arm1", "glov1", "leg1", "boot1"}

// pickSet finds the texture set for a realm letter and level. A family is a
// letter, code and number such as "n_chn1"; it is usable when it has every
// part, and the lowest such number wins. A realm without that level (Midgard
// has no plate) falls back to the nearest lighter one.
func pickSet(skins map[string]bool, realm string, level int) (armourSet, int, bool) {
	for lv := level; lv >= 0; lv-- {
		for _, code := range familyCodes(realm, lv) {
			for n := 0; n <= 9; n++ {
				fam := fmt.Sprintf("%s_%s%d", realm, code, n)
				ok := true
				for _, p := range armourParts {
					ok = ok && skins[fam+"_"+p]
				}
				if ok {
					return armourSet{
						Body: fam + "_body1m", BodyF: fam + "_body1f", LBody: fam + "_lbody1",
						Arms: fam + "_arm1", Gloves: fam + "_glov1", Legs: fam + "_leg1", Boots: fam + "_boot1",
					}, lv, true
				}
			}
		}
	}
	return armourSet{}, 0, false
}

// files lists every texture of a set.
func (a armourSet) files() []string {
	return []string{a.Body, a.BodyF, a.LBody, a.Arms, a.Gloves, a.Legs, a.Boots}
}

// pickMajority returns the key with the highest count, the lowest key on a tie.
func pickMajority(counts map[int]int) (int, bool) {
	best, bestN, found := 0, 0, false
	for k, n := range counts {
		if n > bestN || (n == bestN && found && k < best) {
			best, bestN, found = k, n, true
		}
	}
	return best, found
}

// armourPieces classifies the (slot, model) pairs the NPCs wear from the item
// rows. Each row is {item slot, model, object type, realm}. A number used by
// items of several types takes the commonest one.
func armourPieces(wanted map[[2]int]bool, rows [][4]int) map[string]armourPiece {
	type tally struct{ types, realms map[int]int }
	byKey := map[[2]int]*tally{}
	for _, r := range rows {
		key := [2]int{r[0], r[1]}
		if !wanted[key] {
			continue
		}
		if _, ok := armourLevel(r[2]); !ok {
			continue
		}
		t := byKey[key]
		if t == nil {
			t = &tally{map[int]int{}, map[int]int{}}
			byKey[key] = t
		}
		t.types[r[2]]++
		if r[3] >= 1 && r[3] <= 3 {
			t.realms[r[3]]++
		}
	}
	out := map[string]armourPiece{}
	for key, t := range byKey {
		typ, _ := pickMajority(t.types)
		lv, _ := armourLevel(typ)
		realm, _ := pickMajority(t.realms)
		out[fmt.Sprintf("%d:%d", key[0], key[1])] = armourPiece{Level: lv, Realm: realm}
	}
	return out
}

// convertArmour builds the armour section: it classifies the worn numbers
// from the world store, chooses a texture set for each realm and level in
// use, writes those textures to outDir/armor, and reads the tints.
func convertArmour(game string, db *sql.DB, outDir string) (*armourData, error) {
	skins, err := gamedata.PlayerSkins(game)
	if err != nil {
		return nil, fmt.Errorf("pskins.csv: %w", err)
	}
	tints, err := gamedata.Tints(game)
	if err != nil {
		return nil, fmt.Errorf("tints.csv: %w", err)
	}

	wanted := map[[2]int]bool{}
	rows, err := db.Query(`SELECT DISTINCT slot, model FROM npc_equip WHERE slot IN (22,23,25,27,28) AND model > 0`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var s, m int
		if err := rows.Scan(&s, &m); err != nil {
			rows.Close()
			return nil, err
		}
		wanted[[2]int{s, m}] = true
	}
	rows.Close()

	items, err := db.Query(`SELECT item_type, model, object_type, realm FROM item WHERE item_type IN (22,23,25,27,28) AND object_type BETWEEN 32 AND 38`)
	if err != nil {
		return nil, err
	}
	var irows [][4]int
	for items.Next() {
		var r [4]int
		if err := items.Scan(&r[0], &r[1], &r[2], &r[3]); err != nil {
			items.Close()
			return nil, err
		}
		irows = append(irows, r)
	}
	items.Close()

	out := &armourData{Pieces: armourPieces(wanted, irows), Textures: map[string]armourSet{}, Tints: map[string][3]int{}}

	// A texture set for every realm letter at every level: a piece whose items
	// name no realm is dressed in the creature's own, which the client picks.
	// pskins.csv lists families whose files are not all in the install, so a
	// family with a missing file is dropped and the pick repeated.
	dir := filepath.Join(outDir, "armor")
	for attempt := 0; attempt < 6; attempt++ {
		out.Textures = map[string]armourSet{}
		need := map[string]bool{}
		for _, letter := range []string{"b", "n", "h"} {
			for lv := 0; lv <= 4; lv++ {
				if set, _, ok := pickSet(skins, letter, lv); ok {
					out.Textures[fmt.Sprintf("%s:%d", letter, lv)] = set
					for _, f := range set.files() {
						need[f+".dds"] = true
					}
				}
			}
		}
		if err := extractTextures(game, dir, need); err != nil && attempt == 0 {
			fmt.Fprintln(os.Stderr, "  armour textures: some are not in the install; choosing other families")
		}
		bad := map[string]bool{}
		for _, set := range out.Textures {
			for _, f := range set.files() {
				if _, err := os.Stat(filepath.Join(dir, f+".png")); err != nil {
					bad[strings.TrimSuffix(set.Body, "_body1m")] = true
				}
			}
		}
		if len(bad) == 0 {
			break
		}
		for fam := range bad {
			for _, part := range armourParts {
				delete(skins, fam+"_"+part)
			}
		}
	}

	colours, err := db.Query(`SELECT DISTINCT color FROM npc_equip WHERE slot IN (22,23,25,27,28) AND color > 0`)
	if err != nil {
		return nil, err
	}
	var ids []int
	for colours.Next() {
		var c int
		if err := colours.Scan(&c); err != nil {
			colours.Close()
			return nil, err
		}
		ids = append(ids, c)
	}
	colours.Close()
	sort.Ints(ids)
	for _, c := range ids {
		if rgb, ok := tints[c]; ok {
			out.Tints[fmt.Sprint(c)] = rgb
		}
	}
	fmt.Printf("armour: %d worn numbers classified, %d texture sets, %d tints\n",
		len(out.Pieces), len(out.Textures), len(out.Tints))
	return out, nil
}
