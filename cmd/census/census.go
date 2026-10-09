package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"daocweb/internal/kfa"
	"daocweb/internal/nif"
	"daocweb/internal/scenery"
	"daocweb/internal/zone"
)

// Report is the whole census.
type Report struct {
	Classes []*Class   `json:"classes"`
	Zones   ZoneReport `json:"zones"`
}

// Class is one asset class: how many items there are, where, how many read,
// and why the rest do not.
type Class struct {
	Name      string      `json:"name"`
	What      string      `json:"what"`
	Total     int         `json:"total"`
	Passed    int         `json:"passed"`
	Locations []*Location `json:"locations"`
	Failures  []*Failure  `json:"failures"`
	// ReadNotDrawn counts, for model classes, the files that carry
	// something the reader parses but the converters do not render.
	ReadNotDrawn []*Feature `json:"readNotDrawn,omitempty"`
}

// Feature is something read but not drawn, and how many files carry it.
type Feature struct {
	Feature string `json:"feature"`
	Files   int    `json:"files"`
}

// notDrawn maps the block types that are read exactly but not rendered to
// what they would show.
var notDrawn = map[string]string{
	"NiTextureEffect":              "projected and environment textures",
	"NiGeomMorpherController":      "vertex morphs",
	"NiAlphaController":            "alpha animation",
	"NiMaterialColorController":    "colour animation",
	"NiLightColorController":       "colour animation",
	"NiTextureTransformController": "texture transforms",
	"NiVisController":              "visibility animation",
	"NiPathController":             "path animation",
	"NiFlipController":             "texture flipbooks",
	"NiFogProperty":                "fog properties",
	"NiPlanarCollider":             "particle collisions",
	"NiSphericalCollider":          "particle collisions",
	"NiParticleBomb":               "particle bombs",
}

// features lists the read-not-drawn features a model carries, once each.
func features(f *nif.File) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range f.TypeOf {
		if ft, ok := notDrawn[t]; ok && !seen[ft] {
			seen[ft] = true
			out = append(out, ft)
		}
	}
	return out
}

// readNotDrawn tallies the features of a class's parsed models.
func readNotDrawn(keep *parsed, class string) []*Feature {
	n := map[string]int{}
	for _, m := range keep.files {
		if m.class == class {
			for _, ft := range m.features {
				n[ft]++
			}
		}
	}
	var out []*Feature
	for ft, c := range n {
		out = append(out, &Feature{ft, c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Files != out[j].Files {
			return out[i].Files > out[j].Files
		}
		return out[i].Feature < out[j].Feature
	})
	return out
}

// Location is one folder's share of a class.
type Location struct {
	Path   string `json:"path"`
	Total  int    `json:"total"`
	Passed int    `json:"passed"`
}

// Failure is one cause, how many items it stops, and a few of them.
type Failure struct {
	Cause    string   `json:"cause"`
	Count    int      `json:"count"`
	Examples []string `json:"examples"`
	// ByLocation splits Count by folder, in Locations' order.
	ByLocation []*LocCount `json:"byLocation"`
}

const maxExamples = 5

// item is one thing to read. read reports why it cannot be, or nil.
type item struct {
	loc  string // install-relative folder, forward slashes
	name string // file name within it
	read func() error
}

// runClass reads every item, workers at a time, and tallies the results.
func runClass(name, what string, items []item, workers int) *Class {
	errs := make([]error, len(items))
	parallel(len(items), workers, func(i int) { errs[i] = safely(items[i].read) })

	c := &Class{Name: name, What: what, Total: len(items)}
	locs := map[string]*Location{}
	fails := map[string]*Failure{}
	for i, it := range items {
		l := locs[it.loc]
		if l == nil {
			l = &Location{Path: it.loc}
			locs[it.loc] = l
		}
		l.Total++
		if errs[i] == nil {
			c.Passed++
			l.Passed++
			continue
		}
		cs := cause(errs[i])
		f := fails[cs]
		if f == nil {
			f = &Failure{Cause: cs}
			fails[cs] = f
		}
		f.Count++
		f.Examples = append(f.Examples, it.loc+"/"+it.name)
		if n := len(f.ByLocation); n == 0 || f.ByLocation[n-1].Path != it.loc {
			f.ByLocation = append(f.ByLocation, &LocCount{Path: it.loc})
		}
		f.ByLocation[len(f.ByLocation)-1].Count++
	}
	for _, l := range locs {
		c.Locations = append(c.Locations, l)
	}
	sort.Slice(c.Locations, func(i, j int) bool { return c.Locations[i].Path < c.Locations[j].Path })
	for _, f := range fails {
		sort.Slice(f.ByLocation, func(i, j int) bool { return f.ByLocation[i].Path < f.ByLocation[j].Path })
		sort.Strings(f.Examples)
		if len(f.Examples) > maxExamples {
			f.Examples = f.Examples[:maxExamples]
		}
		c.Failures = append(c.Failures, f)
	}
	sortFailures(c.Failures)
	return c
}

func sortFailures(fs []*Failure) {
	sort.Slice(fs, func(i, j int) bool {
		if fs[i].Count != fs[j].Count {
			return fs[i].Count > fs[j].Count
		}
		return fs[i].Cause < fs[j].Cause
	})
}

// parallel runs fn(0..n-1) on workers goroutines.
func parallel(n, workers int, fn func(i int)) {
	if workers < 1 {
		workers = 1
	}
	next := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				fn(i)
			}
		}()
	}
	for i := 0; i < n; i++ {
		next <- i
	}
	close(next)
	wg.Wait()
}

// panicError is a reader that panicked rather than returning an error.
type panicError struct{ v any }

func (p panicError) Error() string { return fmt.Sprint(p.v) }

// safely runs a read, turning a panic into a failure so one bad file cannot
// stop the census.
func safely(read func() error) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = panicError{v}
		}
	}()
	return read()
}

var (
	digits = regexp.MustCompile(`[0-9]+`)
	quoted = regexp.MustCompile(`"[^"]*"`)
)

// cause names a failure so that failures with one underlying reason group
// together: by kind and block type for the NIF reader's typed errors, and
// otherwise by the message with the numbers and quoted names taken out.
func cause(err error) string {
	var ne *nif.Error
	var pe panicError
	switch {
	case errors.As(err, &ne):
		switch {
		case ne.Kind == nif.Unsupported && ne.Type != "":
			return "unsupported block type " + ne.Type
		case ne.Kind == nif.Unsupported:
			return "unsupported NIF version"
		case ne.Type != "":
			return ne.Kind.String() + " in " + ne.Type
		}
		return ne.Kind.String()
	case errors.As(err, &pe):
		return "panic: " + normalise(pe.Error())
	case errors.Is(err, scenery.ErrNoModel):
		return "model file not found"
	case errors.Is(err, scenery.ErrNoTexture), errors.Is(err, fs.ErrNotExist):
		return "file not found"
	}
	return normalise(err.Error())
}

func normalise(s string) string {
	s = quoted.ReplaceAllString(s, `"…"`)
	return digits.ReplaceAllString(s, "N")
}

// --- enumeration ---------------------------------------------------------

// files lists a folder's files with one of the extensions, ignoring case,
// sorted. A missing folder is empty.
func files(dir string, exts ...string) []string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		x := strings.ToLower(filepath.Ext(e.Name()))
		for _, want := range exts {
			if x == want {
				out = append(out, e.Name())
				break
			}
		}
	}
	sort.Strings(out)
	return out
}

// models lists a model folder: one model per .npk, plus any loose .nif
// without an archive of the same name.
func models(dir string) []string {
	var out []string
	packed := map[string]bool{}
	for _, n := range files(dir, ".npk") {
		out = append(out, n)
		packed[strings.ToLower(strings.TrimSuffix(n, filepath.Ext(n)))] = true
	}
	for _, n := range files(dir, ".nif") {
		if !packed[strings.ToLower(strings.TrimSuffix(n, filepath.Ext(n)))] {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// parsed keeps the models that read, for the texture census.
type parsed struct {
	mu    sync.Mutex
	files []parsedModel
}

type parsedModel struct {
	class     string
	loc, name string
	features  []string   // read-not-drawn features it carries
	embedded  []embedTex // textures stored inside it
	dirs      []string   // where its textures are looked for
	textures  []string   // the external textures it names
}

func (p *parsed) add(m parsedModel) {
	p.mu.Lock()
	p.files = append(p.files, m)
	p.mu.Unlock()
}

// modelItems makes an item per model in each folder, reading it the way
// scenery.LoadModel does and parsing it.
func modelItems(game, class string, locs []string, texDirs func(loc string) []string, keep *parsed) []item {
	var items []item
	for _, loc := range locs {
		dir := filepath.Join(game, filepath.FromSlash(loc))
		for _, n := range models(dir) {
			loc, n := loc, n
			file := n
			if strings.EqualFold(filepath.Ext(n), ".npk") {
				file = strings.TrimSuffix(n, filepath.Ext(n)) + ".nif"
			}
			items = append(items, item{loc: loc, name: n, read: func() error {
				raw, err := scenery.LoadModel(dir, file)
				if err != nil {
					return err
				}
				f, err := nif.Parse(raw)
				if err != nil {
					return err
				}
				keep.add(parsedModel{class: class, loc: loc, name: n, dirs: texDirs(loc),
					textures: textureNames(f), embedded: embeddedTextures(f), features: features(f)})
				return nil
			}})
		}
	}
	return items
}

// --- the run -------------------------------------------------------------

// Run takes the census of the install at game.
func Run(game string, workers int) (*Report, error) {
	list, err := zone.List(game)
	if err != nil {
		return nil, fmt.Errorf("zones.dat: %w", err)
	}
	rel := func(p ...string) string { return filepath.Join(append([]string{game}, p...)...) }
	own := func(loc string) []string { return []string{filepath.Join(game, filepath.FromSlash(loc))} }
	withZones := func(loc string) []string {
		d := own(loc)
		if loc != "zones/Nifs" {
			d = append(d, rel("zones", "Nifs"))
		}
		return d
	}

	rep := &Report{}
	keep := &parsed{}

	rep.Classes = append(rep.Classes, runClass("scenery",
		"outdoor scenery models: trees, rocks, buildings, keeps",
		modelItems(game, "scenery", []string{"zones/Nifs", "frontiers/NIFS", "Newtowns/zones/Nifs"}, withZones, keep), workers))

	rep.Classes = append(rep.Classes, runClass("dungeon pieces",
		"the rooms and corridors dungeons are assembled from",
		modelItems(game, "dungeon pieces", []string{"zones/Dnifs", "frontiers/dnifs"}, own, keep), workers))

	var cityLocs []string
	for _, z := range list {
		if z.Kind != zone.City {
			continue
		}
		if dir, _, err := zone.Dir(game, z.Num); err == nil {
			if r, err := filepath.Rel(game, filepath.Join(dir, "nifs")); err == nil {
				cityLocs = append(cityLocs, filepath.ToSlash(r))
			}
		}
	}
	rep.Classes = append(rep.Classes, runClass("city blocks",
		"the blocks cities are assembled from, in each city zone's nifs folder",
		modelItems(game, "city blocks", cityLocs, own, keep), workers))

	var figs []item
	for _, n := range files(rel("figures"), ".nif") {
		path := rel("figures", n)
		figs = append(figs, item{loc: "figures", name: n, read: func() error {
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			f, err := nif.Parse(raw)
			if err != nil {
				return err
			}
			keep.add(parsedModel{class: "figures", features: features(f)})
			return nil
		}})
	}
	rep.Classes = append(rep.Classes, runClass("figures",
		"character and creature models", figs, workers))

	nodes, err := kfa.Nodes(game)
	if err != nil {
		return nil, err
	}
	var anims []item
	for _, n := range files(rel("anims"), ".kfa") {
		path := rel("anims", n)
		anims = append(anims, item{loc: "anims", name: n, read: func() error {
			_, err := kfa.Read(path, nodes)
			return err
		}})
	}
	rep.Classes = append(rep.Classes, runClass("animations",
		"recorded animations, each track resolved to a bone", anims, workers))

	rep.Classes = append(rep.Classes, runClass("textures",
		"textures the scenery, dungeon and city models name, decoded", textureItems(keep), workers))

	for _, c := range rep.Classes {
		c.ReadNotDrawn = readNotDrawn(keep, c.Name)
	}

	rep.Zones = zoneCensus(game, list, workers)
	return rep, nil
}

// textureItems makes an item per distinct texture the parsed models name,
// per folder.
func textureItems(keep *parsed) []item {
	type key struct{ loc, name string }
	seen := map[key][]string{}
	type embedKey struct{ loc, name string }
	embedded := map[embedKey]error{}
	for _, m := range keep.files {
		for _, e := range m.embedded {
			embedded[embedKey{m.loc, m.name + "#" + e.block}] = e.err
		}
		for _, t := range m.textures {
			k := key{m.loc, t}
			if _, ok := seen[k]; !ok {
				seen[k] = m.dirs
			}
		}
	}
	keys := make([]key, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].loc != keys[j].loc {
			return keys[i].loc < keys[j].loc
		}
		return keys[i].name < keys[j].name
	})
	var items []item
	for _, k := range keys {
		k, dirs := k, seen[k]
		items = append(items, item{loc: k.loc, name: k.name, read: func() error {
			_, err := scenery.ReadTexture(dirs, k.name)
			return err
		}})
	}
	// Embedded textures were decoded as their models were read.
	ekeys := make([]embedKey, 0, len(embedded))
	for k := range embedded {
		ekeys = append(ekeys, k)
	}
	sort.Slice(ekeys, func(i, j int) bool {
		if ekeys[i].loc != ekeys[j].loc {
			return ekeys[i].loc < ekeys[j].loc
		}
		return ekeys[i].name < ekeys[j].name
	})
	for _, k := range ekeys {
		err := embedded[k]
		items = append(items, item{loc: k.loc, name: k.name, read: func() error { return err }})
	}
	// Items are tallied in runs by folder, so keep each folder together.
	sort.SliceStable(items, func(i, j int) bool { return items[i].loc < items[j].loc })
	return items
}

// textureNames lists the external textures a model names.
func textureNames(f *nif.File) []string {
	var out []string
	for _, blk := range f.Blocks {
		if src, ok := blk.(*nif.SourceTexture); ok && src.UseExternal == 1 && src.FileName != "" {
			out = append(out, scenery.TexName(src.FileName))
		}
	}
	return out
}

// LocCount is a failure's count in one folder.
type LocCount struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// embedTex is a texture stored inside a model, and whether it decodes.
type embedTex struct {
	block string
	err   error
}

// embeddedTextures decodes each texture a model stores inside itself.
func embeddedTextures(f *nif.File) []embedTex {
	var out []embedTex
	for _, blk := range f.Blocks {
		src, ok := blk.(*nif.SourceTexture)
		if !ok || src.UseExternal != 0 {
			continue
		}
		e := embedTex{block: fmt.Sprint(src.PixelData)}
		if pd, ok := f.Block(src.PixelData).(*nif.PixelData); ok {
			_, e.err = scenery.DecodePixelData(f, pd)
		} else {
			e.err = fmt.Errorf("embedded texture without pixel data")
		}
		out = append(out, e)
	}
	return out
}
