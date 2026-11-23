package world

import (
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"daocweb/internal/server/rules"
)

// startZones are the open zones a Midgard character plays in first.
var startZones = map[int]bool{100: true, 101: true, 102: true, 151: true}

// The step lets a body fall when the ground drops away by more than
// rules.StepDown in one step. Walking and sprinting over ordinary terrain
// must not do that: only real cliffs may. This reads every converted outdoor
// zone and counts the heightmap's edges steep enough to make a sprinting body
// fall (skipped without converted zones).
func TestOrdinaryGroundDoesNotMakeAWalkerFall(t *testing.T) {
	web := filepath.Join("..", "..", "..", "web", "data")
	dirs, _ := filepath.Glob(filepath.Join(web, "zone[0-9][0-9][0-9]"))
	re := regexp.MustCompile(`zone(\d+)$`)
	checked := 0
	for _, d := range dirs {
		m := re.FindStringSubmatch(d)
		zone, _ := strconv.Atoi(m[1])
		tr, name, err := LoadTerrain(web, zone)
		if err != nil {
			continue // a city or dungeon, or not converted
		}
		checked++
		// One sprint step covers this many units; the ground may fall
		// StepDown over it before the body leaves it.
		step := rules.SprintSpeed(Run) * StepDT
		limit := rules.StepDown / step // the steepest slope that is still ground
		steep, edges := 0, 0
		for y := 0; y < tr.Grid-1; y++ {
			for x := 0; x < tr.Grid-1; x++ {
				h := tr.h(x, y)
				for _, o := range [][2]int{{1, 0}, {0, 1}} {
					edges++
					if math.Abs(tr.h(x+o[0], y+o[1])-h)/tr.Cell > limit {
						steep++
					}
				}
			}
		}
		share := float64(steep) / float64(edges)
		t.Logf("zone %d %s: %d of %d edges steeper than %.1f (%.3f%%)", zone, name, steep, edges, limit, 100*share)
		// The start zones of Midgard must be walkable; a mountain zone has
		// real cliffs, which is what falling is for.
		if startZones[zone] && share > 0.01 {
			t.Errorf("zone %d %s: %.1f%% of its ground is too steep to walk without falling", zone, name, 100*share)
		}
	}
	if checked == 0 {
		t.Skip("no converted outdoor zones")
	}
}
