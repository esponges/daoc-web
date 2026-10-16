package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"

	"daocweb/internal/gamedata"
)

// gearEntry is one database model number the client can draw as a held item.
type gearEntry struct {
	Item string `json:"item"` // directory under items/, lower case
	Kind string `json:"kind"` // gamedata.Kind*
}

// gearFile is web/data/gear.json: the lookup from the community database's
// held-item model numbers to the shared item models converted from the
// install. It describes EA-owned content, so like everything under web/data/
// it is generated locally and never committed.
type gearFile struct {
	Items map[string]gearEntry `json:"items"`
	Armor *armourData          `json:"armor,omitempty"`
}

// convertGear converts every held item the imported NPCs carry, once each.
//
// The database names a weapon by a model number; items.csv in the install
// says which .nif that number draws. A number is kept only when it resolves
// and its kind fits at least one held slot it is worn in, so a bow in a hand
// slot or a shield in the right hand never reaches the client.
func convertGear(game, worldDB, outDir string) error {
	items, err := gamedata.Items(game)
	if err != nil {
		return fmt.Errorf("items.csv: %w", err)
	}
	db, err := sql.Open("sqlite", "file:"+worldDB+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT DISTINCT model, slot FROM npc_equip WHERE slot IN (10,11,12,13) AND model > 0`)
	if err != nil {
		return fmt.Errorf("%s: %w (import the world with cmd/dbimport first)", worldDB, err)
	}
	slotsOf := map[int][]int{}
	for rows.Next() {
		var model, slot int
		if err := rows.Scan(&model, &slot); err != nil {
			rows.Close()
			return err
		}
		slotsOf[model] = append(slotsOf[model], slot)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	// Players wear items from the same table, so the held ones a Midgard
	// character can get (the starter kit and what creatures drop) are drawn too.
	prows, err := db.Query(`SELECT DISTINCT model, item_type FROM item WHERE item_type IN (10,11,12,13) AND model > 0 AND realm IN (0,2) AND (droppable = 1 OR level <= 5)`)
	if err != nil {
		return fmt.Errorf("%s: %w", worldDB, err)
	}
	for prows.Next() {
		var model, slot int
		if err := prows.Scan(&model, &slot); err != nil {
			prows.Close()
			return err
		}
		if !containsInt(slotsOf[model], slot) {
			slotsOf[model] = append(slotsOf[model], slot)
		}
	}
	prows.Close()
	if err := prows.Err(); err != nil {
		return err
	}

	var models []int
	for m := range slotsOf {
		models = append(models, m)
	}
	sort.Ints(models)

	out := gearFile{Items: map[string]gearEntry{}}
	var absent, wrongKind, noModel int
	done := map[string]bool{}
	for _, m := range models {
		it, ok := items[m]
		if !ok {
			absent++
			continue
		}
		fits := false
		for _, s := range slotsOf[m] {
			fits = fits || gamedata.KindFitsSlot(it.Kind, s)
		}
		if !fits {
			wrongKind++
			continue
		}
		dirName := strings.ToLower(it.NIF)
		if !done[dirName] {
			done[dirName] = true
			// The slot name is only a label inside the manifest; the socket is
			// chosen by the client from the database slot.
			e := equip{Slot: "item", Bone: "HELD", Item: it.NIF, Name: it.Name}
			if err := convertItem(game, e, filepath.Join(outDir, "items", dirName)); err != nil {
				fmt.Fprintf(os.Stderr, "  skip model %d (%s): %v\n", m, it.NIF, err)
				done[dirName] = false
				noModel++
				continue
			}
		} else if !fileExists(filepath.Join(outDir, "items", dirName, "char.json")) {
			noModel++
			continue
		}
		out.Items[fmt.Sprint(m)] = gearEntry{Item: dirName, Kind: it.Kind}
	}
	if out.Armor, err = convertArmour(game, db, outDir); err != nil {
		return fmt.Errorf("armour: %w", err)
	}
	js, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "gear.json"), js, 0o644); err != nil {
		return err
	}
	fmt.Printf("gear: %d held-item numbers in the database; %d drawn, %d absent from items.csv, %d wrong kind for their slot, %d would not convert\n",
		len(models), len(out.Items), absent, wrongKind, noModel)
	return nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
