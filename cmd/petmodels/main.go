// petmodels lists the creature models the summon spells' templates name, and
// which of them are not yet converted into web/data/char, so the first-run
// recipe can convert them:
//
//	for m in $(go run ./cmd/petmodels -missing); do
//	  go run ./cmd/charconv -model $m -name m$m && go run ./cmd/animconv -char web/data/char/m$m
//	done
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

// summonTypes are the spell types that make a pet from a template.
const summonTypes = `'SummonMinion','SummonCommander','SummonSpiritFighter','SummonHunterPet'`

// models reads the models the templates of the summon spells name.
func models(db *sql.DB) ([]int, error) {
	rows, err := db.Query(`SELECT DISTINCT e.models FROM npc_template_ext e JOIN spell s ON s.life_drain=e.template WHERE s.type IN (` + summonTypes + `)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := map[int]bool{}
	for rows.Next() {
		var list string
		if err := rows.Scan(&list); err != nil {
			return nil, err
		}
		for _, f := range strings.FieldsFunc(list, func(r rune) bool { return r == ';' || r == '|' }) {
			if m, err := strconv.Atoi(strings.TrimSpace(f)); err == nil && m > 0 {
				seen[m] = true
			}
		}
	}
	var out []int
	for m := range seen {
		out = append(out, m)
	}
	sort.Ints(out)
	return out, rows.Err()
}

func main() {
	store := flag.String("db", filepath.Join("server", "data", "world.db"), "the world store")
	chars := flag.String("chars", filepath.Join("web", "data", "char"), "converted characters")
	missing := flag.Bool("missing", false, "list only the models not converted yet")
	flag.Parse()
	db, err := sql.Open("sqlite", "file:"+*store+"?mode=ro")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	ms, err := models(db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading the store (re-run dbimport if it predates pets):", err)
		os.Exit(1)
	}
	for _, m := range ms {
		_, statErr := os.Stat(filepath.Join(*chars, "m"+strconv.Itoa(m)))
		if *missing && statErr == nil {
			continue
		}
		fmt.Println(m)
	}
}
