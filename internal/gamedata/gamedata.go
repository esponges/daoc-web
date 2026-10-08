// Package gamedata reads the tables in gamedata.mpk that tie a creature's
// model ID to everything needed to draw and animate it.
//
// The server only ever names a creature by number. Everything else the client
// looks up through a chain of CSV files:
//
//	monsters.csv  model ID -> figure, a skin per slot, scale
//	monnifs.csv   figure   -> .nif base name, animation set
//	anims.csv     anim set -> one clip per role: walk, run, idle, death, ...
//	canims.csv    anim set -> combat clips: attacks, flinch, combat idle
//	animnifs.csv  clip     -> .kfa file, frame count, fps, loop or clamp
//	skins.csv     skin     -> texture name, and which skinNNN.mpk holds it
//
// Each file opens with two header rows: a group label above a column name, so
// "Body" under "Skins" is the body skin while "Body" under "Cata Skin" is
// something else entirely. The columns read here are taken by position, from
// the left where those groups are unambiguous.
package gamedata

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"daocweb/internal/mpak"
)

// Slots is the order of the skin columns in monsters.csv, starting at column 3.
var Slots = []string{"Body", "Head", "Arms", "Gloves", "LBody", "Legs", "Boots", "Cloak", "Helm"}

// Monster is one row of monsters.csv.
type Monster struct {
	ID     int
	Name   string
	Figure int            // monnifs.csv row
	Skins  map[string]int // slot -> skins.csv row; slots with 0 are left out
	Scale  float64        // 1.0 is the model as authored
}

// Figure is one row of monnifs.csv.
type Figure struct {
	ID      int
	Name    string
	File    string // base name in figures/, no extension, case as written
	AnimSet int
	// Type is 1 for the player-race figures, whose combat sets carry the
	// humanoid weapon and hand-to-hand clips, and 0 for creatures.
	Type int
}

// Skin is one row of skins.csv.
type Skin struct {
	ID      int
	Name    string
	File    string // as the table writes it, usually .tga; the archives hold .dds
	Archive int    // figures/skins/skinNNN.mpk
}

// DDS is the texture's name as it is actually stored.
func (s Skin) DDS() string {
	ext := filepath.Ext(s.File)
	return strings.TrimSuffix(s.File, ext) + ".dds"
}

// Anim is one row of animnifs.csv.
type Anim struct {
	ID     int
	Name   string
	File   string // .kfa in anims/
	Frames int
	FPS    int // the rate the game plays it at
	Base   int // the rate it was authored at; FPS/Base is the playback speed
	Loop   bool
}

// Tables holds the parsed files.
type Tables struct {
	Monsters map[int]Monster
	Figures  map[int]Figure
	Skins    map[int]Skin
	Anims    map[int]Anim
	// AnimSets maps an anim set to its clips by role, keyed by the
	// lower-cased column name in anims.csv: "walk", "run", "idle", ...
	AnimSets map[int]map[string]int
	// CombatSets is the same for canims.csv, which shares the anim-set
	// numbering: "att med", "flinch", "c-idle", "h2h att m", ...
	CombatSets map[int]map[string]int
}

// Load reads the tables from the install at game.
func Load(game string) (*Tables, error) {
	a, err := mpak.Open(filepath.Join(game, "gamedata.mpk"))
	if err != nil {
		return nil, err
	}
	t := &Tables{
		Monsters: map[int]Monster{}, Figures: map[int]Figure{}, Skins: map[int]Skin{},
		Anims: map[int]Anim{}, AnimSets: map[int]map[string]int{}, CombatSets: map[int]map[string]int{},
	}

	rows, _, err := table(a, "monsters.csv")
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		m := Monster{ID: num(r, 0), Name: col(r, 1), Figure: num(r, 2), Skins: map[string]int{}}
		for i, s := range Slots {
			if v := num(r, 3+i); v > 0 {
				m.Skins[s] = v
			}
		}
		m.Scale = float64(num(r, 12)) / 100
		if m.Scale <= 0 {
			m.Scale = 1
		}
		t.Monsters[m.ID] = m
	}

	if rows, _, err = table(a, "monnifs.csv"); err != nil {
		return nil, err
	}
	for _, r := range rows {
		f := Figure{ID: num(r, 0), Name: col(r, 1), File: col(r, 2), AnimSet: num(r, 3), Type: num(r, 10)}
		t.Figures[f.ID] = f
	}

	if rows, _, err = table(a, "skins.csv"); err != nil {
		return nil, err
	}
	for _, r := range rows {
		s := Skin{ID: num(r, 0), Name: col(r, 1), File: col(r, 2), Archive: num(r, 4)}
		t.Skins[s.ID] = s
	}

	if rows, _, err = table(a, "animnifs.csv"); err != nil {
		return nil, err
	}
	for _, r := range rows {
		an := Anim{ID: num(r, 0), Name: col(r, 1), File: col(r, 2),
			Frames: num(r, 3), FPS: num(r, 4), Base: num(r, 5), Loop: strings.EqualFold(col(r, 6), "loop")}
		t.Anims[an.ID] = an
	}

	if t.AnimSets, err = roleTable(a, "anims.csv"); err != nil {
		return nil, err
	}
	if t.CombatSets, err = roleTable(a, "canims.csv"); err != nil {
		return nil, err
	}
	return t, nil
}

// roleTable reads a table with one row per anim set and one column per role,
// each cell naming a row of animnifs.csv.
func roleTable(a *mpak.Archive, name string) (map[int]map[string]int, error) {
	rows, header, err := table(a, name)
	if err != nil {
		return nil, err
	}
	out := map[int]map[string]int{}
	for _, r := range rows {
		set := map[string]int{}
		for c := 2; c < len(r) && c < len(header); c++ {
			role := strings.ToLower(strings.TrimSpace(header[c]))
			if role == "" {
				continue
			}
			if _, seen := set[role]; seen {
				continue
			}
			if v := num(r, c); v > 0 {
				set[role] = v
			}
		}
		out[num(r, 0)] = set
	}
	return out, nil
}

// Clip resolves one role in a figure's anim set.
func (t *Tables) Clip(fig Figure, role string) (Anim, bool) {
	return t.SetClip(fig.AnimSet, role)
}

// SetClip resolves one role in an anim set: walk, run, idle, death, ...
func (t *Tables) SetClip(set int, role string) (Anim, bool) {
	return t.resolve(t.AnimSets[set][role])
}

// CombatClip resolves one role in an anim set's combat row: "att med",
// "flinch", "c-idle", "h2h att m", ...
func (t *Tables) CombatClip(set int, role string) (Anim, bool) {
	return t.resolve(t.CombatSets[set][role])
}

func (t *Tables) resolve(id int) (Anim, bool) {
	if id <= 0 {
		return Anim{}, false
	}
	an, ok := t.Anims[id]
	return an, ok && an.File != ""
}

// table returns a CSV's data rows and its column-name row (the second header).
// Rows whose first field is not a number -- blank lines, notes -- are skipped.
func table(a *mpak.Archive, name string) (rows [][]string, header []string, err error) {
	raw, err := a.Read(name)
	if err != nil {
		return nil, nil, err
	}
	r := csv.NewReader(bytes.NewReader(raw))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	all, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	if len(all) < 2 {
		return nil, nil, fmt.Errorf("%s: no header", name)
	}
	header = all[1]
	for _, row := range all[2:] {
		if _, err := strconv.Atoi(col(row, 0)); err == nil {
			rows = append(rows, row)
		}
	}
	return rows, header, nil
}

func col(r []string, i int) string {
	if i >= len(r) {
		return ""
	}
	return strings.TrimSpace(r[i])
}

func num(r []string, i int) int {
	v, _ := strconv.Atoi(col(r, i))
	return v
}

// Rate is how fast the game plays a clip relative to how it was authored:
// the fps column over the base fps column. Most are 15 over 15. The idles
// are the exception -- the troll's is 2 over 15, so the game plays it at a
// seventh and a half of its keyed speed, a slow breath rather than a pant.
func (a Anim) Rate() float64 {
	if a.FPS <= 0 || a.Base <= 0 {
		return 1
	}
	return float64(a.FPS) / float64(a.Base)
}

// Length is how long the clip lasts in seconds of its own keys: frames at
// the base rate. 0 if the row does not say.
func (a Anim) Length() float64 {
	if a.Frames <= 0 || a.Base <= 0 {
		return 0
	}
	return float64(a.Frames) / float64(a.Base)
}

// ByFile finds the first animnifs.csv row that plays a .kfa, for clips named
// by file rather than reached through an anim set.
func (t *Tables) ByFile(file string) (Anim, bool) {
	best := Anim{}
	for _, an := range t.Anims {
		if strings.EqualFold(an.File, file) && (best.ID == 0 || an.ID < best.ID) {
			best = an
		}
	}
	return best, best.ID != 0
}
