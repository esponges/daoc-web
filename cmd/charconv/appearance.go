package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"daocweb/internal/gamedata"
)

// What a player may choose at creation: a face and a hair colour. The faces
// are the skins named like the race's own ("v_mhed1" .. "v_mhed6"); the hair
// colours are the hair skins (hair_black, ...; the trolls have t_hair1..4).
// The server's rules/appearance.go counts them; these lists give the client
// the picture for each number (1 up).

// hairColours are the human hair skins, in the order a number picks them.
var hairColours = []string{"hair_black", "hair_brown", "hair_dkbrown", "hair_auburn", "hair_red",
	"hair_brightred", "hair_blonde", "hair_dirtblonde", "hair_gray", "hair_white"}

var faceName = regexp.MustCompile(`^(.*hed)\d(.*)\.dds$`)

type looks struct {
	Faces []string `json:"faces"`
	Hair  []string `json:"hair"`
}

type appearanceFile struct {
	Races map[string]map[string]looks `json:"races"` // race number -> "m" / "f" -> looks
}

// convertAppearance writes web/data/appearance/*.png and appearance.json for
// the races in races.json.
func convertAppearance(game, dataDir string) error {
	t, err := gamedata.Load(game)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Join(dataDir, "races.json"))
	if err != nil {
		return fmt.Errorf("convert the races first (-races): %w", err)
	}
	var rf racesFile
	if err := json.Unmarshal(b, &rf); err != nil {
		return err
	}
	dir := filepath.Join(dataDir, "appearance")
	out := appearanceFile{Races: map[string]map[string]looks{}}
	have := map[string]bool{}
	grab := func(stem string) bool {
		if have[stem] {
			return true
		}
		if extractTextures(game, dir, map[string]bool{stem + ".dds": true}) != nil {
			return false
		}
		have[stem] = true
		return true
	}
	for rid, rm := range rf.Races {
		out.Races[rid] = map[string]looks{}
		for g, key := range map[string]string{"m": rm.Male, "f": rm.Female} {
			if key == "" {
				continue
			}
			id, _ := strconv.Atoi(strings.TrimPrefix(key, "m"))
			m := t.Monsters[id]
			face, ok := t.Skins[m.Face]
			if !ok {
				continue
			}
			var l looks
			if sm := faceName.FindStringSubmatch(strings.ToLower(face.DDS())); sm != nil {
				for n := 1; n <= 6; n++ {
					if s := sm[1] + strconv.Itoa(n) + sm[2]; grab(s) {
						l.Faces = append(l.Faces, s)
					}
				}
			}
			hair := hairColours
			if hs, ok := t.Skins[m.Skins["Head"]]; ok && strings.HasPrefix(strings.ToLower(hs.DDS()), "t_hair") {
				hair = []string{"t_hair1", "t_hair2", "t_hair3", "t_hair4"}
			}
			for _, h := range hair {
				if grab(h) {
					l.Hair = append(l.Hair, h)
				}
			}
			out.Races[rid][g] = l
			fmt.Printf("race %s %s: %d faces, %d hair colours\n", rid, g, len(l.Faces), len(l.Hair))
		}
	}
	js, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "appearance.json"), js, 0o644)
}
