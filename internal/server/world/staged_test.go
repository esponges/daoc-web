package world

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"daocweb/internal/importer"
)

// writeStore makes a world store with the given number of spawns.
func writeStore(t *testing.T, path string, spawns int) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(importer.Schema); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < spawns; i++ {
		if _, err := db.Exec(`INSERT INTO spawn(zone,name) VALUES(100,'wolf')`); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSwapStaged(t *testing.T) {
	cases := []struct {
		name    string
		staged  func(t *testing.T, next string)
		swapped bool
		fails   bool
	}{
		{"valid", func(t *testing.T, next string) { writeStore(t, next, 3) }, true, false},
		{"no spawns", func(t *testing.T, next string) { writeStore(t, next, 0) }, false, true},
		{"truncated", func(t *testing.T, next string) {
			writeStore(t, next, 3)
			b, _ := os.ReadFile(next)
			os.WriteFile(next, b[:len(b)/3], 0o644)
		}, false, true},
		{"none", func(t *testing.T, next string) {}, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "world.db")
			writeStore(t, path, 1)
			players := filepath.Join(dir, "players.db")
			os.WriteFile(players, []byte("characters"), 0o644)
			c.staged(t, path+".next")
			old, _ := os.ReadFile(path)

			swapped, err := SwapStaged(path)
			if swapped != c.swapped || (err != nil) != c.fails {
				t.Fatalf("swapped %v err %v; want swapped %v, failure %v", swapped, err, c.swapped, c.fails)
			}
			cur, _ := os.ReadFile(path)
			if c.swapped {
				if prev, _ := os.ReadFile(path + ".prev"); !bytes.Equal(prev, old) {
					t.Error("the old store was not kept as .prev")
				}
				if _, err := os.Stat(path + ".next"); err == nil {
					t.Error("the staged store is still there")
				}
				db, _ := sql.Open("sqlite", path)
				var n int
				db.QueryRow(`SELECT count(*) FROM spawn`).Scan(&n)
				db.Close()
				if n != 3 {
					t.Errorf("store in place has %d spawns, want the staged 3", n)
				}
			} else {
				if !bytes.Equal(cur, old) {
					t.Error("the store in place changed")
				}
				if _, err := os.Stat(path + ".prev"); err == nil {
					t.Error("a .prev was made without a swap")
				}
			}
			if b, _ := os.ReadFile(players); string(b) != "characters" {
				t.Error("players.db was touched")
			}
		})
	}
}
