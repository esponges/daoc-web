package importer

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// A route is a path whose id is a ticket a stable master sells: its points come
// out in step order, in zone coordinates; a path nobody sells, a ticket with
// no path and a path that leaves every zone are not routes.
func TestImportHorseRoute(t *testing.T) {
	dir := minimalDump(t)
	dumpTable(t, dir, "zones", "`ZoneID`, `RegionID`, `OffsetX`, `OffsetY`, `Width`, `Height`, `Name`, `WaterLevel`, `Realm`",
		"1, 1, 2, 3, 1, 1, 'Test', 0, 2")
	dumpTable(t, dir, "mob", "`Name`, `Region`, `X`, `Y`, `Z`, `Model`, `ClassType`, `ItemsListTemplateID`",
		"'Stabler', 1, 16584, 24676, 0, 10, 'DOL.GS.GameStableMaster', 'ST'")
	dumpTable(t, dir, "merchantitem", "`ItemListID`, `ItemTemplateID`, `PageNumber`, `SlotPosition`",
		"'ST', 'HS_A_B', 0, 0", "'ST', 'hs_off', 0, 1", "'ST', 'junk', 0, 2")
	dumpTable(t, dir, "path", "`PathID`, `PathType`", "'HS_A_B', 1", "'HS_Unsold', 1", "'HS_OFF', 1")
	dumpTable(t, dir, "pathpoints", "`PathID`, `Step`, `X`, `Y`, `Z`, `MaxSpeed`, `WaitTime`",
		"'HS_A_B', 3, 16584, 24776, 30, 1000, 0",
		"'HS_A_B', 1, 16584, 24676, 10, 0, 0",
		"'HS_A_B', 2, 16684, 24676, 20, 5000, 2",
		"'HS_Unsold', 1, 16584, 24676, 10, 0, 0",
		"'HS_OFF', 1, 16584, 24676, 10, 0, 0",
		"'HS_OFF', 2, 900000, 900000, 0, 0, 0")
	out := filepath.Join(t.TempDir(), "world.db")
	rep, err := Run(Options{Dump: dir, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Routes != 1 || rep.RoutesDropped != 1 {
		t.Errorf("routes %d dropped %d, want 1 and 1", rep.Routes, rep.RoutesDropped)
	}
	db, err := sql.Open("sqlite", out)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var path string
	var start, n int
	if err := db.QueryRow(`SELECT path, start_zone, points FROM route`).Scan(&path, &start, &n); err != nil || path != "HS_A_B" || start != 1 || n != 3 {
		t.Fatalf("route %q start %d points %d (%v), want HS_A_B from zone 1 with 3", path, start, n, err)
	}
	rows, err := db.Query(`SELECT step, zone, x, y, z, speed, wait FROM route_point WHERE path='HS_A_B' ORDER BY step`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	want := [][7]float64{{1, 1, 200, 100, 10, 0, 0}, {2, 1, 300, 100, 20, 5000, 2}, {3, 1, 200, 200, 30, 1000, 0}}
	for i := 0; rows.Next(); i++ {
		var got [7]float64
		if err := rows.Scan(&got[0], &got[1], &got[2], &got[3], &got[4], &got[5], &got[6]); err != nil {
			t.Fatal(err)
		}
		if i >= len(want) || got != want[i] {
			t.Errorf("point %d = %v, want %v", i, got, want[min(i, len(want)-1)])
		}
	}
}
