package main

import (
	"math"
	"sort"

	"daocweb/internal/interior"
	"daocweb/internal/zone"
)

// InteriorRow is one city or dungeon zone's dry run: its layout read, every
// model it places baked, and its transform to zone coordinates fitted, as
// interiorconv would, without writing anything.
type InteriorRow struct {
	Num        int    `json:"num"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Layout     string `json:"layout"` // "ok", or why the layout does not read
	Placements int    `json:"placements"`
	Converting int    `json:"converting"`
	Skipped    int    `json:"skipped,omitempty"` // layout rows naming no model
	// Failing lists the models that do not convert, as for outdoor zones.
	Failing []ModelFail `json:"failing,omitempty"`
	// Origin is the fitted transform's, with x mirrored; OnSurface is how
	// many of its Points (exits and lights) it puts on the geometry.
	Origin    [3]float64 `json:"origin"`
	OnSurface int        `json:"onSurface"`
	Points    int        `json:"points"`
}

// interiorCensus dry-runs every city and dungeon zone present.
func interiorCensus(game string, list []zone.Info, workers int) []*InteriorRow {
	var todo []zone.Info
	for _, z := range list {
		if (z.Kind == zone.City || z.Kind == zone.Dungeon) && present(game, z.Num) {
			todo = append(todo, z)
		}
	}
	rows := make([]*InteriorRow, len(todo))
	parallel(len(todo), workers, func(i int) { rows[i] = dryRunInterior(game, todo[i]) })
	sort.Slice(rows, func(i, j int) bool { return rows[i].Num < rows[j].Num })
	return rows
}

func dryRunInterior(game string, info zone.Info) *InteriorRow {
	row := &InteriorRow{Num: info.Num, Name: info.Name, Kind: info.Kind.String(), Layout: "ok"}
	var l *interior.Layout
	if err := safely(func() error {
		var err error
		l, err = interior.Read(game, info.Num)
		return err
	}); err != nil {
		row.Layout = scrub(game, cause(err))
		return row
	}
	var b *interior.Built
	if err := safely(func() error {
		b = interior.Build(game, l)
		return nil
	}); err != nil {
		row.Layout = scrub(game, cause(err))
		return row
	}
	row.Placements, row.Converting, row.Skipped = len(b.Placed), b.Converting(), b.Skipped
	for file, err := range b.Failing {
		row.Failing = append(row.Failing, ModelFail{File: file, Placements: b.Uses[file], Cause: scrub(game, cause(err))})
	}
	sort.Slice(row.Failing, func(i, j int) bool {
		a, c := row.Failing[i], row.Failing[j]
		if a.Placements != c.Placements {
			return a.Placements > c.Placements
		}
		return a.File < c.File
	})
	f := interior.FitInterior(l, b.Surface)
	for k := 0; k < 3; k++ {
		row.Origin[k] = math.Round(float64(f.Origin[k]))
	}
	row.OnSurface, row.Points = f.OnSurface, f.Points
	return row
}
