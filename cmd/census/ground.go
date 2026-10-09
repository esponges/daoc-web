package main

import (
	"errors"
	"math"
	"path/filepath"
	"strings"

	"daocweb/internal/ground"
	"daocweb/internal/zone"
)

// groundCensus fills in how a zone's ground will look: the colour atlas's
// layout, the ground detail and the grass, read by the same code zoneconv
// writes them with.
func groundCensus(game string, z *zone.Zone, sx, sy, grid int, row *ZoneRow) {
	quiet := func(string, ...any) {}
	why := func(err error) string { return scrub(game, cause(err)) }

	row.Layout = "no reference tiles"
	if sx == sy {
		if err := safely(func() error {
			ref, err := ground.ReadReference(z, sx)
			if err != nil {
				return err
			}
			lod, err := ground.ReadLOD(z, sx, sy)
			if err != nil {
				return err
			}
			c, err := ground.CheckLayout(lod, ref)
			if err != nil {
				return err
			}
			row.Layout = "as named"
			if c.Chosen != ground.Layouts()[0] {
				row.Layout = strings.Join(strings.Fields(c.Chosen.String()), " ")
			}
			row.LayoutAgree = round4(c.Agree)
			return nil
		}); err != nil {
			row.Layout = "no reference tiles: " + why(err)
		}
	}

	row.Detail = "ok"
	if sx != sy {
		row.Detail = "sectors not square"
	} else if err := safely(func() error {
		sp, err := ground.ReadSplat(game, z, sx, quiet)
		if err != nil {
			return err
		}
		row.DetailAgree = round4(sp.Agree)
		if sp.Weighted {
			row.DetailBlend = "weighted"
		}
		return nil
	}); err != nil {
		row.Detail = why(err)
	}

	row.Grass = "ok"
	if err := safely(func() error {
		_, err := ground.ReadGrass(game, z, grid, quiet)
		return err
	}); err != nil {
		row.Grass = why(err)
		if errors.Is(err, ground.ErrNoGrassMap) {
			row.Grass = "no grass map"
		}
	}
}

// scrub takes the install's path out of a reason, which the committed
// report must not carry.
func scrub(game, s string) string {
	for _, p := range []string{game, filepath.ToSlash(game)} {
		s = strings.ReplaceAll(s, p, "<install>")
	}
	return s
}

func round4(v float64) float64 { return math.Round(v*1e4) / 1e4 }
