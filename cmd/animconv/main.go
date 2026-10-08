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
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"daocweb/internal/gamedata"
	"daocweb/internal/nif"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

// gaitBones are the joints a locomotion cycle has to drive to be worth using.
// An animation that moves only the spine and arms is an emote, not a walk.
var gaitBones = []string{
	"Bip01 L Thigh", "Bip01 R Thigh",
	"Bip01 L Calf", "Bip01 R Calf",
	"Bip01 L Foot", "Bip01 R Foot",
}

type track struct {
	Bone      string
	Rotations []nif.QuatKey
	Trans     []nif.VecKey
	Scales    []nif.FloatKey
	Start     float32
	Stop      float32
}

type anim struct {
	Name     string
	Duration float32
	Tracks   []track
}

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
	jobs := make([]job, len(src))
	for i := range src {
		jobs[i] = job{strings.TrimSpace(src[i]), strings.TrimSpace(dst[i]), true}
	}
	return convert(game, char, nodes, skel, jobs)
}

// A job is one clip to convert: the .kfa base name, the name the viewer will
// know it by, and whether it loops or plays once and holds.
type job struct {
	src, out string
	loop     bool
}

// A tableRole is one clip taken from a model's tables when none are named:
// the name the viewer uses, and the column it comes from -- in anims.csv for
// movement and death, canims.csv for combat.
type tableRole struct {
	out, column string
	combat      bool
}

var tableRoles = []tableRole{
	{"idle", "idle", false},
	{"walk", "walk", false},
	{"run", "run", false},
	{"death", "death", false},
	{"cidle", "c-idle", true},
	{"flinch", "flinch", true},
	{"attack1", "att med", true},
	{"attack2", "att high", true},
	{"attack3", "att low", true},
}

// unarmedAttacks replace the three attacks for a player-race figure. Their
// "att" columns are one-handed weapon swings, and nothing here holds a
// weapon yet; the hand-to-hand set is what an empty-handed fighter uses.
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
	var jobs []job
	for _, r := range tableRoles {
		col := r.column
		if humanoid && unarmedAttacks[r.out] != "" {
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
		jobs = append(jobs, job{strings.TrimSuffix(an.File, filepath.Ext(an.File)), r.out, an.Loop})
		mode := "once"
		if an.Loop {
			mode = "loop"
		}
		fmt.Printf("  anim set %d %-9s %-10s -> %-18s %s  %q\n", m.AnimSet, col, r.out, an.File, mode, an.Name)
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("anim set %d resolves no clips", m.AnimSet)
	}
	return jobs, nil
}

// loadAnimNodes reads the bone registry. Line N is the name of bone index N.
func loadAnimNodes(game string) ([]string, error) {
	path := filepath.Join(game, "figures", "animnode.dat")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		out = append(out, strings.TrimSpace(sc.Text()))
	}
	return out, sc.Err()
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

// readAnim walks the two parallel chains and resolves every track to a bone.
func readAnim(path string, nodes []string) (*anim, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := nif.Parse(raw)
	if err != nil {
		return nil, err
	}
	if len(f.Roots) == 0 {
		return nil, fmt.Errorf("no root block")
	}
	root, _ := f.Block(f.Roots[0]).(*nif.Node)
	if root == nil || len(root.ExtraData) == 0 {
		return nil, fmt.Errorf("root is not a node carrying the track chains")
	}
	a := &anim{Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))}
	s, c := root.ExtraData[0], root.Controller
	for s >= 0 && c >= 0 {
		se, _ := f.Block(s).(*nif.StringExtra)
		kc, _ := f.Block(c).(*nif.KeyframeController)
		if se == nil || kc == nil {
			break
		}
		idx, err := strconv.Atoi(strings.TrimSpace(se.Value))
		if err != nil {
			return nil, fmt.Errorf("track names bone %q, which is not an index", se.Value)
		}
		if idx < 0 || idx >= len(nodes) {
			return nil, fmt.Errorf("track names bone index %d, outside animnode.dat's %d slots", idx, len(nodes))
		}
		t := track{Bone: nodes[idx], Start: kc.StartTime, Stop: kc.StopTime}
		if kd, _ := f.Block(kc.Data).(*nif.KeyframeData); kd != nil {
			t.Rotations, t.Trans, t.Scales = kd.Rotations, kd.Translations, kd.Scales
		}
		if t.Stop > a.Duration {
			a.Duration = t.Stop
		}
		a.Tracks = append(a.Tracks, t)
		s, c = se.Next, kc.Next
	}
	if len(a.Tracks) == 0 {
		return nil, fmt.Errorf("no tracks")
	}
	return a, nil
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
	for _, jb := range jobs {
		name, out := jb.src, jb.out
		path, err := findCase(filepath.Join(game, "anims"), name+".kfa")
		if err != nil {
			return err
		}
		a, err := readAnim(path, nodes)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		ao := animOut{Name: out, Source: filepath.Base(path), Duration: a.Duration, Loop: jb.loop}
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
			return fmt.Errorf("%s: no track matched the character's skeleton", name)
		}

		js, err := json.Marshal(ao)
		if err != nil {
			return err
		}
		file := out + ".json"
		if err := os.WriteFile(filepath.Join(dir, file), js, 0o644); err != nil {
			return err
		}
		entry := indexEntry{out, file, ao.Source, ao.Duration, ao.Loop, len(ao.Tracks)}
		replaced := false
		for k := range idx.Animations {
			if idx.Animations[k].Name == out {
				idx.Animations[k], replaced = entry, true
			}
		}
		if !replaced {
			idx.Animations = append(idx.Animations, entry)
		}

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
