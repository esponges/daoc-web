package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"daocweb/internal/gamedata"
)

// A player race is named in the install's monsters.csv: "Troll Male" and
// "Troll Female" are rows like any creature's, with the figure, the skin for
// each body slot and the race's scale. So a race and a gender pick a base
// model by the race's name; no number is offset or guessed.
//
// What the install cannot say is which race number the community database
// uses for which race name. That list is the game's own, fixed by its races,
// and is kept here (the project's archetypes already use Norse 5 and Troll 6).
type race struct {
	ID    int
	Name  string // the prefix monsters.csv uses: "<Name> Male", "<Name> Female"
	Realm int    // 1 Albion, 2 Midgard, 3 Hibernia
}

var races = []race{
	{1, "Briton", 1}, {2, "Avalonian", 1}, {3, "High", 1}, {4, "Saracen", 1},
	{5, "Norse", 2}, {6, "Troll", 2}, {7, "Dwarf", 2}, {8, "Kobold", 2},
	{9, "Celt", 3}, {10, "Firbolg", 3}, {11, "Elf", 3}, {12, "Lurikeen", 3},
	{13, "Inconnu", 1}, {14, "Valkyn", 2}, {15, "Sylvan", 3},
	{16, "Half Ogre", 1}, {17, "Frostalf", 2}, {18, "Shar", 3},
}

// raceModel is the pair of monsters.csv rows a race is drawn from.
type raceModel struct {
	Male   string `json:"m"` // converted character directory, "m<row>"
	Female string `json:"f"`
}

// racesFile is web/data/races.json: database race number to the converted
// base model for each gender. Generated locally from the install, never
// committed. A race absent from the file is drawn with the archetype's outfit.
type racesFile struct {
	Races map[string]raceModel `json:"races"`
}

// raceRows finds, for each race, the first monsters.csv row named exactly
// "<Race> Male" and "<Race> Female". The numbered rows ("Troll Male 2") and the
// merchant and shopkeeper rows differ only in head or clothes and are skipped.
// Returned rows are monsters.csv IDs, zero where the install has none.
func raceRows(monsters map[int]gamedata.Monster) map[int][2]int {
	ids := make([]int, 0, len(monsters))
	for id := range monsters {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	// The install names some races' rows "Troll Male" and others "Kobold Male 1".
	byName := map[string]int{}
	for _, id := range ids {
		n := strings.TrimSpace(strings.TrimRight(monsters[id].Name, "0123456789"))
		if byName[n] == 0 {
			byName[n] = id
		}
	}
	out := map[int][2]int{}
	for _, r := range races {
		m, f := byName[r.Name+" Male"], byName[r.Name+" Female"]
		if m != 0 || f != 0 {
			out[r.ID] = [2]int{m, f}
		}
	}
	return out
}

// buildRaces makes the races file for the given realms (0 for all) and returns
// the monsters.csv rows to convert with -model.
func buildRaces(monsters map[int]gamedata.Monster, realm int) (racesFile, []int) {
	rows := raceRows(monsters)
	f := racesFile{Races: map[string]raceModel{}}
	var convert []int
	for _, r := range races {
		if realm != 0 && r.Realm != realm {
			continue
		}
		pair, ok := rows[r.ID]
		if !ok {
			continue
		}
		var rm raceModel
		if pair[0] != 0 {
			rm.Male = "m" + strconv.Itoa(pair[0])
			convert = append(convert, pair[0])
		}
		if pair[1] != 0 {
			rm.Female = "m" + strconv.Itoa(pair[1])
			convert = append(convert, pair[1])
		}
		f.Races[strconv.Itoa(r.ID)] = rm
	}
	return f, convert
}

// convertRaces writes races.json for a realm and lists the rows to convert.
// Converting them is the usual two commands per row (the README recipe), so
// the list is printed rather than run here.
func convertRaces(game, dataDir string, realm int) error {
	t, err := gamedata.Load(game)
	if err != nil {
		return err
	}
	f, convert := buildRaces(t.Monsters, realm)
	if len(convert) == 0 {
		return fmt.Errorf("no player race rows found in monsters.csv")
	}
	b, err := json.MarshalIndent(f, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dataDir, "races.json"), b, 0o644); err != nil {
		return err
	}
	fmt.Printf("races.json: %d races\nconvert with:\n", len(f.Races))
	for _, id := range convert {
		fmt.Printf("  go run ./cmd/charconv -model %d -name m%d && go run ./cmd/animconv -char web/data/char/m%d\n", id, id, id)
	}
	return nil
}
