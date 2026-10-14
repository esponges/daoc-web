package importer

import (
	"database/sql"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// repo returns the paths of the checkout and converted zones, skipping the
// test when this machine has neither (they are never in the repository).
func repo(t *testing.T) (dump, web string) {
	t.Helper()
	dump = filepath.Join("..", "..", "server", "data", "community", "opendaoc-db-core")
	web = filepath.Join("..", "..", "web", "data")
	if _, err := os.Stat(filepath.Join(dump, "mob.sql")); err != nil {
		t.Skip("no community database checked out in server/data/community")
	}
	if _, err := os.Stat(filepath.Join(web, "zone100", "heights.u16")); err != nil {
		t.Skip("zone 100 is not converted")
	}
	return
}

func TestImportZone100(t *testing.T) {
	if testing.Short() {
		t.Skip("imports the whole dump")
	}
	dump, web := repo(t)
	out := filepath.Join(t.TempDir(), "world.db")
	rep, err := Run(Options{Dump: dump, Out: out, WebDir: web})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Missing) > 0 {
		t.Fatalf("missing tables: %v", rep.Missing)
	}
	db, err := sql.Open("sqlite", out)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var n int
	db.QueryRow(`SELECT count(*) FROM spawn WHERE zone=100`).Scan(&n)
	if n < 1500 {
		t.Errorf("zone 100 kept %d spawns, want over 1500", n)
	}
	// Every kept spawn is within MaxDz of the ground, and in the zone.
	hb, _ := os.ReadFile(filepath.Join(web, "zone100", "heights.u16"))
	rows, err := db.Query(`SELECT x,y,z FROM spawn WHERE zone=100`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	bad := 0
	for rows.Next() {
		var x, y, z float64
		rows.Scan(&x, &y, &z)
		if x < 0 || y < 0 || x >= 65536 || y >= 65536 {
			t.Fatalf("spawn outside the zone: %v,%v", x, y)
		}
		g := float64(binary.LittleEndian.Uint16(hb[(int(y/256)*256+int(x/256))*2:]))
		if d := z - g; d > 500 || d < -500 {
			bad++
		}
	}
	if bad > 0 {
		t.Errorf("%d spawns more than 500 units from the ground", bad)
	}
	// A wolf is model 56, which the client draws as its Large Grey Wolf.
	db.QueryRow(`SELECT count(*) FROM spawn WHERE zone=100 AND model=56`).Scan(&n)
	if n == 0 {
		t.Error("no model-56 spawns in Mularn")
	}
}

func TestMissingTableReported(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(Options{Dump: dir, Out: filepath.Join(t.TempDir(), "w.db")}); err == nil {
		t.Error("an empty dump directory must be an error")
	}
}
