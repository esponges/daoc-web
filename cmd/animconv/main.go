// Command animconv converts DAoC's recorded animations for a character.
//
// The .kfa files in anims/ are not a separate format: they are NetImmerse
// 4.2.1.0 NIFs, the same generation as the scenery, holding nothing but
// NiKeyframeController and NiKeyframeData blocks. What makes them unusual is
// how a track names the bone it drives.
//
// The nodes in a .kfa are unnamed. Instead the file's single real node carries
// two linked lists that run in lockstep: a chain of NiStringExtraData and a
// chain of NiKeyframeController. The n-th string belongs to the n-th
// controller, and its value is a number -- "0", "1", "2", "6" -- that indexes
// figures/animnode.dat, a 178-line list of 3ds Max Biped bone names. So
// resolving a track to a bone means walking both chains together and looking
// the index up:
//
//	chain position 5  ->  string "6"  ->  animnode.dat line 6  ->  "Bip01 Neck"
//
// That indirection is why the animations are portable. They name no skeleton
// of their own, so any model whose bones carry Biped names can play any of
// them, which is how one set of humanoid animations serves every player race.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"daocweb/internal/gamedata"
	"daocweb/internal/kfa"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

// gaitBones are the joints a locomotion cycle has to drive to be worth using.
// An animation that moves only the spine and arms is an emote, not a walk.
var gaitBones = []string{
	"Bip01 L Thigh", "Bip01 R Thigh",
	"Bip01 L Calf", "Bip01 R Calf",
	"Bip01 L Foot", "Bip01 R Foot",
}

// The reading is internal/kfa's, shared with the census.
type (
	track = kfa.Track
	anim  = kfa.Anim
)

var (
	readAnim      = kfa.Read
	loadAnimNodes = kfa.Nodes
)

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	char := flag.String("char", filepath.Join("web", "data", "char", "norseman"), "converted character directory")
	list := flag.Bool("list", false, "scan anims/ and rank animations by how well they fit the character")
	gait := flag.Bool("gait", false, "with -list, show only animations that drive the legs")
	top := flag.Int("top", 30, "with -list, how many to print")
	names := flag.String("anims", "", "comma-separated .kfa base names to convert")
	as := flag.String("as", "", "comma-separated output names, parallel to -anims")
	flag.Parse()

	if err := run(*game, *char, *list, *gait, *top, *names, *as); err != nil {
		fmt.Fprintln(os.Stderr, "animconv:", err)
		os.Exit(1)
	}
}

func run(game, char string, list, gaitOnly bool, top int, names, as string) error {
	nodes, err := loadAnimNodes(game)
	if err != nil {
		return err
	}
	skel, err := loadCharBones(char)
	if err != nil {
		return err
	}
	fmt.Printf("animnode.dat: %d bone slots; character: %d bones\n", len(nodes), len(skel))

	if list {
		return scan(game, nodes, skel, gaitOnly, top)
	}
	if names == "" {
		jobs, err := fromAnimSet(game, char)
		if err != nil {
			return err
		}
		return convert(game, char, nodes, skel, jobs)
	}
	// Named clips are taken to loop; they are almost always locomotion.
	src := strings.Split(names, ",")
	dst := src
	if as != "" {
		if dst = strings.Split(as, ","); len(dst) != len(src) {
			return fmt.Errorf("-as has %d names for %d animations", len(dst), len(src))
		}
	}
	t, _ := gamedata.Load(game) // optional here: only the rates come from it
	jobs := make([]job, len(src))
	for i := range src {
		jobs[i] = job{src: strings.TrimSpace(src[i]), out: strings.TrimSpace(dst[i]), rate: 1, loop: true}
		// A clip named by hand still has a row in animnifs.csv, which says
		// how fast the game plays it.
		if t != nil {
			if an, ok := t.ByFile(jobs[i].src + ".kfa"); ok {
				jobs[i].rate = an.Rate()
				jobs[i].length = an.Length()
			}
		}
	}
	return convert(game, char, nodes, skel, jobs)
}

// A job is one clip to convert: the .kfa base name, the name the viewer will
// know it by, and whether it loops or plays once and holds.
type job struct {
	src, out string
	rate     float64 // playback speed relative to the keys; 1 if unknown
	loop     bool
	length   float64 // seconds the tables say the clip lasts; 0 if unknown
	// alias marks a clip another job may already have converted: it is then
	// listed under its own name but shares the file, and a failure is not fatal.
	alias bool
}

// A tableRole is one clip taken from a model's tables when none are named:
// the name the viewer uses, and the column it comes from -- in anims.csv for
// movement and death, canims.csv for combat.
type tableRole struct {
	out, column string
	combat      bool
	// player marks a role only a humanoid figure gets, and written lazy: the
	// postures and emotes of the player races.
	player bool
}

var tableRoles = []tableRole{
	{"idle", "idle", false, false},
	{"walk", "walk", false, false},
	{"run", "run", false, false},
	{"back", "back", false, false},
	{"strafeL", "slide left", false, false},
	{"strafeR", "slide right", false, false},
	{"death", "death", false, false},
	{"cidle", "c-idle", true, false},
	{"flinch", "flinch", true, false},
	{"attack1", "att med", true, false},
	{"attack2", "att high", true, false},
	{"attack3", "att low", true, false},
	// Only for a figure with a shield on its arm; see shieldOnly.
	{"block", "block h", true, false},
	// Held while a creature casts; the server says when it ends.
	{"cast", "cast mid", false, false},
}

// playerRoles are the clips of a player's body beyond walking and fighting:
// jumping, sitting, swimming, and the classic emotes. An emote is called
// em_<name>, the name its command has (server world/emotes.go).
var playerRoles = []tableRole{
	{"jump", "jump", false, true},
	{"land", "land", false, true},
	{"sitdown", "sit down", false, true},
	{"sit", "sit idle", false, true},
	{"standup", "stand up", false, true},
	{"swim", "crawl", false, true},
	{"em_beckon", "em beckon", false, true},
	{"em_blush", "em blush", false, true},
	{"em_bow", "em bow", false, true},
	{"em_cheer", "em cheer", false, true},
	{"em_clap", "em clap", false, true},
	{"em_cry", "em cry", false, true},
	{"em_curtsey", "em curtsey", false, true},
	{"em_dance", "em dance", false, true},
	{"em_flex", "em flex", false, true},
	{"em_kiss", "em kiss", false, true},
	{"em_laugh", "em laugh", false, true},
	{"em_point", "em point", false, true},
	{"em_salute", "em salute", false, true},
	{"em_bang", "em shield", false, true},
	{"em_victory", "em victory", false, true},
	{"em_wave", "em wave", false, true},
}

// unarmedAttacks replace the three attacks for an empty-handed player-race
// figure. Its "att" columns are one-handed weapon swings; with nothing in
// the hand, the hand-to-hand set is what fits.
var unarmedAttacks = map[string]string{
	"attack1": "h2h att m",
	"attack2": "h2h att h",
	"attack3": "h2h att l",
}

// fromAnimSet picks the clips the game itself plays for this model, from the
// anim set charconv recorded out of monnifs.csv.
func fromAnimSet(game, char string) ([]job, error) {
	b, err := os.ReadFile(filepath.Join(char, "char.json"))
	if err != nil {
		return nil, err
	}
	var m struct {
		Model   int `json:"model"`
		AnimSet int `json:"animSet"`
		Equip   []struct {
			Slot string `json:"slot"`
		} `json:"equip"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.AnimSet == 0 {
		return nil, fmt.Errorf("%s has no anim set; give -anims, or -list to search", char)
	}
	t, err := gamedata.Load(game)
	if err != nil {
		return nil, err
	}
	humanoid := t.Figures[t.Monsters[m.Model].Figure].Type == 1
	// What it holds decides how it fights: a weapon in the right hand
	// takes the one-handed swings, a shield on the left adds a block.
	armed, shield := false, false
	for _, e := range m.Equip {
		armed = armed || e.Slot == "right"
		shield = shield || e.Slot == "left"
	}
	var jobs []job
	for _, r := range append(append([]tableRole{}, tableRoles...), playerRoles...) {
		col := r.column
		if r.player && !playerFigure(char) {
			continue
		}
		if r.out == "block" && !shield {
			continue
		}
		if humanoid && !armed && unarmedAttacks[r.out] != "" {
			col = unarmedAttacks[r.out]
		}
		var an gamedata.Anim
		var ok bool
		if r.combat {
			an, ok = t.CombatClip(m.AnimSet, col)
		} else {
			an, ok = t.SetClip(m.AnimSet, col)
		}
		if !ok {
			fmt.Printf("  anim set %d has no %s clip\n", m.AnimSet, col)
			continue
		}
		// Some sets fill the step columns with the forward walk -- the
		// badger's "back" is its walk -- which would carry it the wrong
		// way. Without a clip of its own the viewer makes do with the walk.
		if r.out == "back" || r.out == "strafeL" || r.out == "strafeR" {
			if w, ok := t.SetClip(m.AnimSet, "walk"); ok && strings.EqualFold(w.File, an.File) {
				fmt.Printf("  anim set %d %s is just the walk; skipped\n", m.AnimSet, col)
				continue
			}
		}
		jobs = append(jobs, job{src: strings.TrimSuffix(an.File, filepath.Ext(an.File)), out: r.out, rate: an.Rate(), loop: an.Loop, length: an.Length(), alias: r.player})
		mode := "once"
		if an.Loop {
			mode = "loop"
		}
		fmt.Printf("  anim set %d %-9s %-10s -> %-18s %s  %q\n", m.AnimSet, col, r.out, an.File, mode, an.Name)
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("anim set %d resolves no clips", m.AnimSet)
	}
	jobs = append(jobs, castJobs(t, m.AnimSet)...)
	if playerFigure(char) {
		jobs = append(jobs, wieldJobs(t, m.AnimSet)...)
		jobs = append(jobs, styleJobs(t, m.AnimSet)...)
		if err := writeStyleMap(filepath.Join(filepath.Dir(filepath.Dir(filepath.Clean(char))), "styles.json"), t); err != nil {
			fmt.Printf("  styles.json: %v\n", err)
		}
	}
	return jobs, nil
}

// wieldSets are a player body's plain swings and combat stance for each kind of
// weapon it can hold, from canims.csv: the server names the kind a player holds
// (an entity's wd), and the viewer plays att_<kind>1..3 and cidle_<kind>. The
// body's own attack1..3 are whatever it was converted holding, which for a race
// base is nothing: the hand-to-hand swings.
var wieldSets = []struct {
	kind    string
	attacks [3]string // medium, high, low
	idle    string    // "" uses the one-handed stance
}{
	{"1h", [3]string{"att med", "att high", "att low"}, "c-idle"},
	{"2h", [3]string{"att 2h med", "att 2h high", "att 2h low"}, "c-idle 2h"},
	{"pole", [3]string{"att pole med", "att pole h", "att pole l"}, "c-idle pole"},
	{"staff", [3]string{"att staff-m", "att staff h", "att staff l"}, "c-idle staff"},
	{"dual", [3]string{"att dual m", "att dual h", "att dual l"}, ""},
	{"h2h", [3]string{"h2h att m", "h2h att h", "h2h att l"}, "h2h cidle"},
}

// wieldJobs are the clips of wieldSets an anim set has, written lazy.
func wieldJobs(t *gamedata.Tables, set int) []job {
	var jobs []job
	add := func(col, out string) {
		an, ok := t.CombatClip(set, col)
		if !ok {
			return
		}
		jobs = append(jobs, job{src: strings.TrimSuffix(an.File, filepath.Ext(an.File)), out: out,
			rate: an.Rate(), loop: an.Loop, length: an.Length(), alias: true})
	}
	for _, w := range wieldSets {
		for i, col := range w.attacks {
			add(col, fmt.Sprintf("att_%s%d", w.kind, i+1))
		}
		if w.idle != "" {
			add(w.idle, "cidle_"+w.kind)
		}
	}
	return jobs
}

// styleNumbers lists the combat animation numbers any style uses, sorted.
func styleNumbers(t *gamedata.Tables) []int {
	seen := map[int]bool{}
	for _, n := range t.StyleAnims {
		seen[n] = true
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// styleJobs are a player body's style animations, each called style<number>
// after its column in canims.csv and written lazy, like the cast clips.
func styleJobs(t *gamedata.Tables, set int) []job {
	var jobs []job
	for _, n := range styleNumbers(t) {
		an, ok := t.CombatClipAt(set, n)
		if !ok {
			continue
		}
		jobs = append(jobs, job{src: strings.TrimSuffix(an.File, filepath.Ext(an.File)), out: fmt.Sprintf("style%d", n),
			rate: an.Rate(), loop: an.Loop, length: an.Length(), alias: true})
	}
	return jobs
}

// writeStyleMap writes styles.json: each client style number (what the server
// sends as a swing's s) to the canims.csv number of its animation, so the
// viewer can play style<number>.
func writeStyleMap(path string, t *gamedata.Tables) error {
	m := map[string]int{}
	for id, n := range t.StyleAnims {
		m[strconv.Itoa(id)] = n
	}
	b, err := json.Marshal(struct {
		Styles map[string]int `json:"styles"`
	}{m})
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// playerFigure says whether a converted character directory is a player's
// body: an outfit named by hand (norseman, troll-plate) or one of the race
// bases web/data/races.json lists (m137, the male troll). A creature model
// (m100) has no use for sitting or emotes.
func playerFigure(char string) bool {
	clean := filepath.Clean(char)
	base := filepath.Base(clean)
	if !(len(base) > 1 && base[0] == 'm' && strings.Trim(base[1:], "0123456789") == "") {
		return true
	}
	b, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(clean)), "races.json"))
	if err != nil {
		return false
	}
	var rs struct {
		Races map[string]map[string]string `json:"races"`
	}
	if json.Unmarshal(b, &rs) != nil {
		return false
	}
	for _, r := range rs.Races {
		for _, model := range r {
			if model == base {
				return true
			}
		}
	}
	return false
}

// castNumbers lists the animation numbers any spell uses for its caster, sorted.
func castNumbers(t *gamedata.Tables) []int {
	seen := map[int]bool{}
	for _, s := range t.Spells {
		for _, n := range s.Anim {
			if n > 0 {
				seen[n] = true
			}
		}
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// castJobs are the clips the spell table names for a caster in this anim set,
// each called cast<number> after its column in anims.csv. Sets that fill them
// with a stand-in share its file instead of adding copies.
func castJobs(t *gamedata.Tables, set int) []job {
	var jobs []job
	for _, n := range castNumbers(t) {
		an, ok := t.SetClipAt(set, n)
		if !ok {
			continue
		}
		jobs = append(jobs, job{src: strings.TrimSuffix(an.File, filepath.Ext(an.File)), out: fmt.Sprintf("cast%d", n),
			rate: an.Rate(), loop: an.Loop, length: an.Length(), alias: true})
	}
	return jobs
}

// loadCharBones reads the bone names out of a converted character's manifest.
func loadCharBones(dir string) (map[string]bool, error) {
	path := filepath.Join(dir, "char.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s (run charconv first): %w", path, err)
	}
	var m struct {
		Bones []struct {
			Name string `json:"name"`
		} `json:"bones"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, bn := range m.Bones {
		out[strings.ToLower(bn.Name)] = true
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s lists no bones", path)
	}
	return out, nil
}

type scored struct {
	name     string
	tracks   int
	matched  int
	legs     int
	duration float32
	keys     int
}

func scan(game string, nodes []string, skel map[string]bool, gaitOnly bool, top int) error {
	dir := filepath.Join(game, "anims")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var rows []scored
	failed := 0
	for _, e := range ents {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".kfa") {
			continue
		}
		a, err := readAnim(filepath.Join(dir, e.Name()), nodes)
		if err != nil {
			failed++
			continue
		}
		r := scored{name: a.Name, tracks: len(a.Tracks), duration: a.Duration}
		for _, t := range a.Tracks {
			if skel[strings.ToLower(t.Bone)] {
				r.matched++
			}
			for _, g := range gaitBones {
				if strings.EqualFold(g, t.Bone) && len(t.Rotations) > 1 {
					r.legs++
					break
				}
			}
			r.keys += len(t.Rotations) + len(t.Trans) + len(t.Scales)
		}
		rows = append(rows, r)
	}
	fmt.Printf("scanned %d animations, %d unreadable\n\n", len(rows), failed)

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].legs != rows[j].legs {
			return rows[i].legs > rows[j].legs
		}
		if rows[i].matched != rows[j].matched {
			return rows[i].matched > rows[j].matched
		}
		return rows[i].name < rows[j].name
	})

	fmt.Printf("%-34s %6s %8s %5s %8s %7s\n", "animation", "tracks", "on model", "legs", "duration", "keys")
	shown := 0
	for _, r := range rows {
		if gaitOnly && r.legs < len(gaitBones) {
			continue
		}
		if shown >= top {
			break
		}
		fmt.Printf("%-34s %6d %8d %5d %7.2fs %7d\n",
			r.name, r.tracks, r.matched, r.legs, r.duration, r.keys)
		shown++
	}
	full := 0
	for _, r := range rows {
		if r.legs == len(gaitBones) {
			full++
		}
	}
	fmt.Printf("\n%d animations drive all six gait joints\n", full)
	return nil
}

// trackOut is one bone's channels. Rotations are quaternions in NIF's own
// w,x,y,z order; the viewer slerps between them. A channel with no keys means
// the bone keeps its bind value for that component, so an empty array is
// meaningful rather than missing data.
type trackOut struct {
	Bone  string       `json:"bone"`
	Rot   [][5]float32 `json:"rot,omitempty"`   // t, w, x, y, z
	Trans [][4]float32 `json:"trans,omitempty"` // t, x, y, z
	Scale [][2]float32 `json:"scale,omitempty"` // t, s
}

type animOut struct {
	Name     string     `json:"name"`
	Source   string     `json:"source"`
	Duration float32    `json:"duration"`
	Loop     bool       `json:"loop"` // false: play once and hold the last frame
	Rate     float64    `json:"rate"` // playback speed: animnifs.csv fps over base fps
	Tracks   []trackOut `json:"tracks"`
	Unmapped []string   `json:"unmapped,omitempty"`
}

type indexEntry struct {
	Name     string  `json:"name"`
	File     string  `json:"file"`
	Source   string  `json:"source"`
	Duration float32 `json:"duration"`
	Loop     bool    `json:"loop"`
	Tracks   int     `json:"tracks"`
	// Lazy marks a clip the client fetches only when it is first wanted: the
	// casting animations, of which a figure may have dozens.
	Lazy bool `json:"lazy,omitempty"`
}

type indexOut struct {
	Animations []indexEntry `json:"animations"`
}

func convert(game, char string, nodes []string, skel map[string]bool, jobs []job) error {
	dir := filepath.Join(char, "anim")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Merge into what is already there, so a second run can add clips or
	// replace some -- the tables' set first, then hand-picked locomotion
	// over the top of it -- without dropping the rest.
	var idx indexOut
	if b, err := os.ReadFile(filepath.Join(dir, "index.json")); err == nil {
		_ = json.Unmarshal(b, &idx)
	}
	// What is converted already, by source file, so that a clip another role
	// named is shared rather than written again.
	done := map[string]indexEntry{}
	for _, e := range idx.Animations {
		done[strings.ToLower(e.Source)] = e
	}
	// skip says a failed alias clip is left out rather than stopping the run.
	skip := func(jb job, err error) bool {
		if jb.alias {
			fmt.Printf("  %-8s skipped: %v\n", jb.out, err)
		}
		return jb.alias
	}
	for _, jb := range jobs {
		name, out := jb.src, jb.out
		if prior, ok := done[strings.ToLower(name+".kfa")]; ok && jb.alias {
			idx.Animations = withEntry(idx.Animations, indexEntry{out, prior.File, prior.Source, prior.Duration, prior.Loop, prior.Tracks, true})
			continue
		}
		path, err := findCase(filepath.Join(game, "anims"), name+".kfa")
		if err != nil {
			if skip(jb, err) {
				continue
			}
			return err
		}
		a, err := readAnim(path, nodes)
		if err != nil {
			if skip(jb, err) {
				continue
			}
			return fmt.Errorf("%s: %w", name, err)
		}

		ao := animOut{Name: out, Source: filepath.Base(path), Duration: a.Duration, Loop: jb.loop, Rate: jb.rate}
		// A stray key past the end would otherwise stretch the clip: the
		// right strafe has a second of standing still after its stride,
		// held by one key at 2s, where its row says 15 frames at 15. Only
		// a tail that is nothing but such a key is cut. The rows are not
		// always right the other way -- the wolf's walk is listed at 10
		// frames and keyed through 3.6s -- so keys that carry motion past
		// the row's length are trusted over it.
		if jb.length > 0 && float64(a.Duration) > jb.length+0.5/15 && strayTail(a, jb.length) {
			fmt.Printf("  %-8s keys run to %.2fs; the tables say %.2fs, so it ends there\n", out, a.Duration, jb.length)
			ao.Duration = float32(jb.length)
		}
		keys := 0
		for _, t := range a.Tracks {
			// A track for a bone this model does not have is dropped, not
			// an error: the animations are shared across every player race
			// and carry tracks for bones a given skeleton may lack.
			if !skel[strings.ToLower(t.Bone)] {
				ao.Unmapped = append(ao.Unmapped, t.Bone)
				continue
			}
			to := trackOut{Bone: t.Bone}
			for _, k := range t.Rotations {
				to.Rot = append(to.Rot, [5]float32{k.Time, k.Value[0], k.Value[1], k.Value[2], k.Value[3]})
			}
			for _, k := range t.Trans {
				to.Trans = append(to.Trans, [4]float32{k.Time, k.Value[0], k.Value[1], k.Value[2]})
			}
			for _, k := range t.Scales {
				to.Scale = append(to.Scale, [2]float32{k.Time, k.Value})
			}
			keys += len(to.Rot) + len(to.Trans) + len(to.Scale)
			ao.Tracks = append(ao.Tracks, to)
		}
		if len(ao.Tracks) == 0 {
			err := fmt.Errorf("%s: no track matched the character's skeleton", name)
			if skip(jb, err) {
				continue
			}
			return err
		}

		js, err := json.Marshal(ao)
		if err != nil {
			return err
		}
		file := out + ".json"
		if err := os.WriteFile(filepath.Join(dir, file), js, 0o644); err != nil {
			return err
		}
		entry := indexEntry{out, file, ao.Source, ao.Duration, ao.Loop, len(ao.Tracks), jb.alias}
		idx.Animations = withEntry(idx.Animations, entry)
		done[strings.ToLower(ao.Source)] = entry

		fmt.Printf("  %-8s <- %-16s %5.2fs  %3d tracks (%d dropped), %4d keys, %d bytes\n",
			out, ao.Source, ao.Duration, len(ao.Tracks), len(ao.Unmapped), keys, len(js))
	}

	js, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "index.json"), js, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %d animations to %s\n", len(idx.Animations), dir)
	return nil
}

// findCase resolves a filename case-insensitively; the animation names on disk
// are not consistently capitalised.
func findCase(dir, name string) (string, error) {
	p := filepath.Join(dir, name)
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range ents {
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("%s not found in %s", name, dir)
}

// strayTail reports whether every key after length seconds sits on the
// clip's very last instant: a lone end key with nothing between it and the
// stride, rather than motion the tables forgot to count.
func strayTail(a *anim, length float64) bool {
	const eps = 0.5 / 15
	late := func(t float32) bool {
		return float64(t) > length+eps && t < a.Duration-eps
	}
	for _, t := range a.Tracks {
		for _, k := range t.Rotations {
			if late(k.Time) {
				return false
			}
		}
		for _, k := range t.Trans {
			if late(k.Time) {
				return false
			}
		}
		for _, k := range t.Scales {
			if late(k.Time) {
				return false
			}
		}
	}
	return true
}

// withEntry adds an index entry, or replaces the one of the same name.
func withEntry(list []indexEntry, e indexEntry) []indexEntry {
	for k := range list {
		if list[k].Name == e.Name {
			list[k] = e
			return list
		}
	}
	return append(list, e)
}

var gamedataLoad = gamedata.Load
