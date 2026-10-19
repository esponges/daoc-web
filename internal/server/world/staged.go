package world

import (
	"database/sql"
	"fmt"
	"os"
)

// stagedTables are the tables a staged world store must have before the server
// will start on it.
var stagedTables = []string{"zone", "spawn", "npc_template", "item", "style", "startup", "merchant_stock"}

// SwapStaged puts a staged world store (path+".next", left by dbimport while
// the server held path) in place before the server opens it. The current store
// becomes path+".prev", for a one-step rollback. A staged store that does not
// open, lacks a table or has no spawns is left where it is, and the reason is
// returned with swapped false. No staged store is no error.
func SwapStaged(path string) (swapped bool, err error) {
	next := path + ".next"
	if _, err := os.Stat(next); err != nil {
		return false, nil
	}
	if err := checkStore(next); err != nil {
		return false, fmt.Errorf("staged world store %s not used: %w", next, err)
	}
	if _, err := os.Stat(path); err == nil {
		prev := path + ".prev"
		os.Remove(prev)
		if err := os.Rename(path, prev); err != nil {
			return false, fmt.Errorf("keeping the old world store: %w", err)
		}
	}
	if err := os.Rename(next, path); err != nil {
		return false, fmt.Errorf("swapping in the staged world store: %w", err)
	}
	return true, nil
}

func checkStore(path string) error {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	for _, t := range stagedTables {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, t).Scan(&n); err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("no %s table", t)
		}
	}
	var spawns int
	if err := db.QueryRow(`SELECT count(*) FROM spawn`).Scan(&spawns); err != nil {
		return err
	}
	if spawns == 0 {
		return fmt.Errorf("no spawns")
	}
	return nil
}
