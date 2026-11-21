// Command effectconv converts the spell effects the game shows: what a caster
// does while casting, what flies to the target, and what plays on it.
//
// The client's own tables say all of it. spells.csv gives each client effect
// (the number the community data keeps as a spell's ClientEffect) its three
// caster animations, its effects on the caster, its projectile and its target
// effect; speffects.csv gives each effect its model, the bone it hangs on and
// how long it plays; spnifs.csv names the models, which sit in the install's
// effects directory. See docs/formats.md, "How a spell looks".
//
//	effectconv -world server/data/world.db
//
// It converts only the effects the spells of that world store use, with the
// same model and particle reading the scenery has, into one pack
// (web/data/effects/pack: props.bin, props.json and textures) and an
// index.json that ties spells to effects and effects to models. Effects it
// cannot draw are listed, and the rest of the spell's effects are still
// written. Nothing it writes is committed.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"

	"daocweb/internal/gamedata"
	"daocweb/internal/nif"
	"daocweb/internal/scenery"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	world := flag.String("world", filepath.Join("server", "data", "world.db"), "world store: its spells say which client effects to convert")
	out := flag.String("out", filepath.Join("web", "data", "effects"), "output directory")
	all := flag.Bool("all", false, "convert every client effect in the spell table, not only the world's")
	flag.Parse()
	if err := run(*game, *world, *out, *all); err != nil {
		fmt.Fprintln(os.Stderr, "effectconv:", err)
		os.Exit(1)
	}
}

// Index is index.json: how each client effect looks, and what each effect is.
type Index struct {
	// Spells maps a client effect (as text, for JSON) to its parts.
	Spells map[string]SpellOut `json:"spells"`
	// Effects maps an effect number to its model in the pack and its timing.
	Effects map[string]EffectOut `json:"effects"`
}

// SpellOut is one client effect: the caster's animation numbers (columns of
// anims.csv) and the effects, each an effect number, 0 for none.
type SpellOut struct {
	Name       string `json:"name"`
	Anims      [3]int `json:"anims"`
	Caster     []int  `json:"caster,omitempty"`
	Projectile int    `json:"projectile,omitempty"`
	Target     int    `json:"target,omitempty"`
	Buildup    int    `json:"buildup,omitempty"`
	Pulse      int    `json:"pulse,omitempty"`
	Resist     int    `json:"resist,omitempty"`
}

// EffectOut is one effect: Model indexes the pack's models, and is -1 for an
// effect with nothing to draw (a blank model, or one that could not be read).
type EffectOut struct {
	Name  string  `json:"name"`
	Model int     `json:"model"`
	Node  string  `json:"node"`
	Ms    int     `json:"ms"`
	Rate  int     `json:"rate"`
	Scale float64 `json:"scale"`
}

// Problem is an effect that could not be drawn, and why.
type Problem struct {
	Effect int
	Name   string
	File   string
	Why    string
}

func run(game, world, out string, all bool) error {
	t, err := gamedata.Load(game)
	if err != nil {
		return err
	}
	var want []int
	if all {
		for id := range t.Spells {
			want = append(want, id)
		}
	} else if want, err = worldEffects(world); err != nil {
		return err
	}
	sort.Ints(want)
	idx, need := Plan(t, want)

	dirs := []string{filepath.Join(game, "effects")}
	var (
		verts   []scenery.Vertex
		indices []uint32
		models  []scenery.Model
		skipped []string
		probs   []Problem
		partial []Problem
	)
	ids := make([]int, 0, len(need))
	for id := range need {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	fileSlot := map[string]int{} // model file -> its slot in models, -1 for none
	for _, id := range ids {
		e := t.Effects[id]
		eo := EffectOut{Name: e.Name, Model: -1, Node: e.Node, Ms: e.Duration, Rate: e.Rate, Scale: e.Scale}
		file, ok := t.EffectNIFs[e.NIF]
		switch {
		case !ok:
			probs = append(probs, Problem{id, e.Name, "", fmt.Sprintf("no model %d in spnifs.csv", e.NIF)})
		case strings.EqualFold(file, "empty"):
			// A deliberate blank: the effect exists to be a placeholder.
		default:
			slot, seen := fileSlot[strings.ToLower(file)]
			if !seen {
				slot = -1
				name := onDisk(dirs[0], file)
				if name == "" {
					probs = append(probs, Problem{id, e.Name, file, "model not in the effects directory"})
				} else if mo, err := scenery.ConvertModel(dirs, e.Name, name, &verts, &indices); err != nil {
					probs = append(probs, Problem{id, e.Name, file, err.Error()})
					skipped = append(skipped, file+": "+err.Error())
				} else if len(mo.Groups) == 0 && len(mo.Particles) == 0 {
					probs = append(probs, Problem{id, e.Name, file, "nothing to draw"})
				} else {
					slot = len(models)
					models = append(models, *mo)
					if parts := undrawn(dirs, name); len(parts) > 0 {
						partial = append(partial, Problem{id, e.Name, file, "not drawn: " + strings.Join(parts, ", ")})
					}
				}
				fileSlot[strings.ToLower(file)] = slot
			}
			eo.Model = slot
		}
		idx.Effects[fmt.Sprint(id)] = eo
	}

	pack := filepath.Join(out, "pack")
	if err := scenery.WriteProps(pack, 0, dirs, models, verts, indices, nil, skipped); err != nil {
		return err
	}
	b, err := json.MarshalIndent(idx, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "index.json"), b, 0o644); err != nil {
		return err
	}
	fmt.Printf("%d client effects, %d effects used, %d models converted -> %s\n", len(idx.Spells), len(need), len(models), out)
	fmt.Printf("%d effects have nothing to draw:\n", len(probs))
	for _, p := range probs {
		fmt.Printf("  effect %d %q (%s): %s\n", p.Effect, p.Name, p.File, p.Why)
	}
	fmt.Printf("%d effects draw only in part:\n", len(partial))
	for _, p := range partial {
		fmt.Printf("  effect %d %q (%s): %s\n", p.Effect, p.Name, p.File, p.Why)
	}
	return nil
}

// Plan builds the index's spell part for the wanted client effects, and the set
// of effects those spells use. Client effects the table lacks are left out;
// the effects table is filled in by the caller.
func Plan(t *gamedata.Tables, want []int) (*Index, map[int]bool) {
	idx := &Index{Spells: map[string]SpellOut{}, Effects: map[string]EffectOut{}}
	need := map[int]bool{}
	add := func(id int) {
		if _, ok := t.Effects[id]; ok && id > 0 {
			need[id] = true
		}
	}
	for _, fx := range want {
		s, ok := t.Spells[fx]
		if !ok {
			continue
		}
		so := SpellOut{Name: s.Name, Anims: s.Anim, Caster: s.Caster, Projectile: s.Projectile, Target: s.Target,
			Buildup: s.Buildup, Pulse: s.Pulse, Resist: s.Resist}
		idx.Spells[fmt.Sprint(fx)] = so
		for _, id := range s.Caster {
			add(id)
		}
		for _, id := range []int{s.Projectile, s.Target, s.Buildup, s.Pulse, s.Resist} {
			add(id)
		}
	}
	return idx, need
}

// worldEffects lists the client effects the world store's spells use.
func worldEffects(path string) ([]int, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT DISTINCT client_effect FROM spell WHERE client_effect > 0`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w (run cmd/dbimport again for the client effect column)", path, err)
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// onDisk finds a model's file in dir whatever its case and extension: the
// table says "cast1", the install holds CAST1.NIF.
func onDisk(dir, base string) string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range ents {
		n := e.Name()
		if strings.EqualFold(strings.TrimSuffix(n, filepath.Ext(n)), base) && strings.EqualFold(filepath.Ext(n), ".nif") {
			return n
		}
	}
	return ""
}

// notDrawn are the block types the model reader parses exactly but the baking
// does not turn into anything on screen: how the colour, alpha, visibility,
// flipbook frame or position of a part changes over time. An effect holding
// one still draws, only more plainly.
var notDrawn = map[string]bool{
	"NiAlphaController": true, "NiMaterialColorController": true, "NiVisController": true,
	"NiPathController": true, "NiTextureTransformController": true, "NiFlipController": true,
	"NiGeomMorpherController": true, "NiLightColorController": true,
}

// undrawn names the kinds of animation a model has that are not drawn, with
// how many of each, sorted; nil when there are none or the file cannot be read.
func undrawn(dirs []string, file string) []string {
	raw, err := scenery.FindModel(dirs, file)
	if err != nil {
		return nil
	}
	f, err := nif.Parse(raw)
	if err != nil {
		return nil
	}
	return UndrawnTypes(f.TypeOf)
}

// UndrawnTypes counts the not-drawn block types in a list of block type names.
func UndrawnTypes(types []string) []string {
	n := map[string]int{}
	for _, t := range types {
		if notDrawn[t] {
			n[t]++
		}
	}
	var out []string
	for t, c := range n {
		out = append(out, fmt.Sprintf("%s x%d", t, c))
	}
	sort.Strings(out)
	return out
}
