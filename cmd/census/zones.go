package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"daocweb/internal/nif"
	"daocweb/internal/scenery"
	"daocweb/internal/zone"
)

// ZoneReport is the zones zones.dat declares, by type, and the outdoor
// zones' dry runs.
type ZoneReport struct {
	Declared int          `json:"declared"`
	Kinds    []*KindCount `json:"kinds"`
	Missing  []ZoneRef    `json:"missing"`
	Outdoor  OutdoorTotal `json:"outdoor"`
	Rows     []*ZoneRow   `json:"rows"`
}

// KindCount is one zone type: how many zones declare it, how many of those
// have their folder in this install, and how many leave the type out.
type KindCount struct {
	Kind     string `json:"kind"`
	Declared int    `json:"declared"`
	Present  int    `json:"present"`
	Untyped  int    `json:"untyped,omitempty"`
}

// ZoneRef names a zone.
type ZoneRef struct {
	Num  int    `json:"num"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// OutdoorTotal sums the outdoor dry runs.
type OutdoorTotal struct {
	Zones       int `json:"zones"`
	Terrain     int `json:"terrain"`     // zones whose terrain reads
	Clean       int `json:"clean"`       // zones whose terrain and every placement convert
	Placements  int `json:"placements"`  // fixtures placed across them
	Converting  int `json:"converting"`  // of those, placements whose model converts
	ModelsTried int `json:"modelsTried"` // distinct models baked
	ModelsBaked int `json:"modelsBaked"`
}

// ZoneRow is one outdoor zone's dry run.
type ZoneRow struct {
	Num      int    `json:"num"`
	Name     string `json:"name"`
	Frontier bool   `json:"frontier,omitempty"`
	Terrain  string `json:"terrain"` // "ok" or why not
	// Scenery says why the zone's nifs.csv or fixtures.csv could not be
	// read; empty when they were.
	Scenery    string `json:"scenery,omitempty"`
	Placements int    `json:"placements"`
	Converting int    `json:"converting"`
	// Failing lists the models that do not convert, each with how many
	// placements it costs and why.
	Failing []ModelFail `json:"failing,omitempty"`
}

// ModelFail is a placed model that does not convert.
type ModelFail struct {
	File       string `json:"file"`
	Placements int    `json:"placements"`
	Cause      string `json:"cause"`
}

// bakeCache bakes each distinct model once across all zones.
type bakeCache struct {
	mu   sync.Mutex
	done map[string]*bakeResult
}

type bakeResult struct {
	once sync.Once
	err  error
}

func (c *bakeCache) bake(dirs []string, file string) error {
	key := strings.Join(dirs, "|") + "|" + strings.ToLower(file)
	c.mu.Lock()
	r := c.done[key]
	if r == nil {
		r = &bakeResult{}
		c.done[key] = r
	}
	c.mu.Unlock()
	r.once.Do(func() {
		r.err = safely(func() error {
			raw, err := scenery.FindModel(dirs, file)
			if err != nil {
				return err
			}
			f, err := nif.Parse(raw)
			if err != nil {
				return err
			}
			var verts []scenery.Vertex
			var indices []uint32
			_, err = scenery.Bake(f, file, file, &verts, &indices)
			return err
		})
	})
	return r.err
}

// zoneCensus counts zones by type and dry-runs every outdoor zone present.
func zoneCensus(game string, list []zone.Info, workers int) ZoneReport {
	rep := ZoneReport{Declared: len(list)}
	kinds := map[string]*KindCount{}
	var outdoor []zone.Info
	for _, z := range list {
		k := z.Kind.String()
		kc := kinds[k]
		if kc == nil {
			kc = &KindCount{Kind: k}
			kinds[k] = kc
		}
		kc.Declared++
		if z.Untyped {
			kc.Untyped++
		}
		if present(game, z.Num) {
			kc.Present++
			if z.Kind == zone.Outdoor {
				outdoor = append(outdoor, z)
			}
		} else {
			rep.Missing = append(rep.Missing, ZoneRef{z.Num, z.Name, k})
		}
	}
	for _, k := range []string{"outdoor", "city", "dungeon", "other"} {
		if kc := kinds[k]; kc != nil {
			rep.Kinds = append(rep.Kinds, kc)
		}
	}

	cache := &bakeCache{done: map[string]*bakeResult{}}
	rows := make([]*ZoneRow, len(outdoor))
	parallel(len(outdoor), workers, func(i int) { rows[i] = dryRun(game, outdoor[i], cache) })
	rep.Rows = rows

	t := &rep.Outdoor
	for _, r := range rows {
		t.Zones++
		if r.Terrain == "ok" {
			t.Terrain++
			if r.Scenery == "" && r.Converting == r.Placements {
				t.Clean++
			}
		}
		t.Placements += r.Placements
		t.Converting += r.Converting
	}
	keys := make([]string, 0, len(cache.done))
	for k := range cache.done {
		keys = append(keys, k)
	}
	for _, k := range keys {
		t.ModelsTried++
		if cache.done[k].err == nil {
			t.ModelsBaked++
		}
	}
	return rep
}

// present is whether a zone's dat archive is in the install.
func present(game string, num int) bool {
	dir, _, err := zone.Dir(game, num)
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, fmt.Sprintf("dat%03d.mpk", num)))
	return err == nil
}

// dryRun reads an outdoor zone's terrain and bakes every model it places,
// as zoneconv and propconv would, without writing anything.
func dryRun(game string, info zone.Info, cache *bakeCache) *ZoneRow {
	row := &ZoneRow{Num: info.Num, Name: info.Name, Terrain: "ok"}
	z, err := zone.Open(game, info.Num)
	if err != nil {
		row.Terrain = cause(err)
		return row
	}
	row.Frontier = z.Frontier
	if err := safely(func() error {
		sec, err := z.Sector()
		if err != nil {
			return err
		}
		if _, err := z.Terrain(sec); err != nil {
			return err
		}
		// zoneconv colours the terrain from the LOD tiles, so a zone
		// without them does not convert.
		_, err = z.Archive("lod")
		return err
	}); err != nil {
		row.Terrain = cause(err)
	}

	var defs map[int]zone.ModelDef
	var fixtures []zone.Fixture
	err = safely(func() error {
		var err error
		if defs, err = z.Models(); err != nil {
			return err
		}
		fixtures, err = z.PlacedFixtures()
		return err
	})
	if err != nil {
		row.Scenery = cause(err)
		return row
	}
	dirs := scenery.ModelDirs(game, z.Frontier)
	fails := map[string]*ModelFail{}
	for _, f := range fixtures {
		row.Placements++
		def, ok := defs[f.NifID]
		var err error
		file := fmt.Sprintf("id %d", f.NifID)
		if !ok {
			err = fmt.Errorf("no entry in nifs.csv")
		} else {
			file = def.File
			err = cache.bake(dirs, def.File)
		}
		if err == nil {
			row.Converting++
			continue
		}
		mf := fails[file]
		if mf == nil {
			mf = &ModelFail{File: file, Cause: cause(err)}
			fails[file] = mf
		}
		mf.Placements++
	}
	for _, mf := range fails {
		row.Failing = append(row.Failing, *mf)
	}
	sort.Slice(row.Failing, func(i, j int) bool {
		a, b := row.Failing[i], row.Failing[j]
		if a.Placements != b.Placements {
			return a.Placements > b.Placements
		}
		return a.File < b.File
	})
	return row
}
