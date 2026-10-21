// Scratch: Slam's style rows and its style-spell rows.
package main

import (
	"fmt"

	"daocweb/internal/importer"
)

func main() {
	dir := "server/data/community/opendaoc-db-core/"
	ids := map[string]bool{}
	importer.ReadTable(dir+"style.sql", func(r *importer.Row) error {
		if r.Str("Name") == "Slam" {
			fmt.Printf("style Slam: ID=%s StyleID=%s ClassId=%s spec=%s\n", r.Str("ID"), r.Str("StyleID"), r.Str("ClassId"), r.Str("SpecKeyName"))
			ids[r.Str("ID")] = true
			ids[r.Str("StyleID")] = true
		}
		return nil
	})
	importer.ReadTable(dir+"stylexspell.sql", func(r *importer.Row) error {
		if ids[r.Str("StyleID")] {
			fmt.Printf("stylexspell: StyleID=%s ClassID=%s SpellID=%s Chance=%s\n", r.Str("StyleID"), r.Str("ClassID"), r.Str("SpellID"), r.Str("Chance"))
		}
		return nil
	})
}
