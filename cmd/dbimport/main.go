// dbimport reads a local checkout of the OpenDAoC community database and
// writes this project's own world store, server/data/world.db: spawns in
// zone coordinates, NPC templates, loot, items, styles, exits and starts.
//
// The checkout is GPL-3.0 and describes content owned by Electronic Arts, so
// it stays out of the repository (server/data/ is gitignored). Nothing here
// copies the emulator's code; see the README's Multiplayer section.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"daocweb/internal/gamedata"
	"daocweb/internal/importer"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

func main() {
	dump := flag.String("dump", filepath.Join("server", "data", "community", "opendaoc-db-core"), "the community database's table directory")
	out := flag.String("out", filepath.Join("server", "data", "world.db"), "world store to write")
	web := flag.String("web", filepath.Join("web", "data"), "converted zones, for the ground check and which zones exist")
	game := flag.String("game", defaultGame, "DAoC install directory, used only when no characters are converted: monsters.csv then decides which models can be drawn")
	chars := flag.String("chars", filepath.Join("web", "data", "char"), "converted characters; a spawn is kept only if its model was converted here (m<N>)")
	allModels := flag.Bool("all-models", false, "keep spawns of every model, converted or not (to list the models to convert; write to a scratch -out)")
	query := flag.String("q", "", "run this SQL against -out and print the rows, instead of importing")
	flag.Parse()
	if *query != "" {
		runQuery(*out, *query)
		return
	}

	if _, err := os.Stat(filepath.Join(*dump, "mob.sql")); err != nil {
		fmt.Fprintf(os.Stderr, "no community database at %s\n\n"+
			"Clone it (it is GPL-3.0 and not part of this repository):\n"+
			"  git clone https://github.com/OpenDAoC/OpenDAoC-Database server/data/community\n"+
			"then run this again, or pass -dump.\n", *dump)
		os.Exit(1)
	}
	var models map[int]bool
	if *allModels {
		models = nil
	} else if m := convertedModels(*chars); len(m) > 0 {
		models = m
		fmt.Printf("keeping spawns whose model is converted (%d models in %s)\n", len(m), *chars)
	} else if *game != "" {
		t, err := gamedata.Load(*game)
		if err != nil {
			fmt.Fprintf(os.Stderr, "reading %s: %v (pass -game \"\" to keep every model)\n", *game, err)
			os.Exit(1)
		}
		models = map[int]bool{}
		for id := range t.Monsters {
			models[id] = true
		}
	}
	rep, err := importer.Run(importer.Options{
		Dump: *dump, Out: *out, WebDir: *web, Models: models,
		Log: func(f string, a ...any) { fmt.Printf(f+"\n", a...) },
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "import:", err)
		os.Exit(1)
	}
	fmt.Printf("\nspawns dropped: %d outside any zone, %d far from the ground\n", rep.DroppedOutside, rep.DroppedGround)
	if rep.UnknownEquip > 0 {
		fmt.Printf("%d spawns name an equipment template with no rows; imported without gear\n", rep.UnknownEquip)
	}
	fmt.Printf("merchants: %d in the dump, %d kept (the rest lack a zone, ground or model)\n", rep.MerchantsSeen, rep.MerchantsKept)
	if n := len(rep.DroppedModel); n > 0 {
		type kv struct{ m, c int }
		var l []kv
		total := 0
		for m, c := range rep.DroppedModel {
			l = append(l, kv{m, c})
			total += c
		}
		sort.Slice(l, func(i, j int) bool { return l[i].c > l[j].c })
		fmt.Printf("%d spawns use %d models the client does not have; most common:", total, n)
		for i := 0; i < 8 && i < len(l); i++ {
			fmt.Printf(" model %d (%d)", l[i].m, l[i].c)
		}
		fmt.Println()
	}
	if len(rep.Missing) > 0 {
		fmt.Fprintf(os.Stderr, "required tables missing or empty: %v\n", rep.Missing)
		os.Exit(1)
	}
	fmt.Println("wrote", *out)
}

// convertedModels lists the model numbers that have a converted character
// (web/data/char/m<N>/char.json), which is what the viewer draws a creature
// with.
func convertedModels(dir string) map[int]bool {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	m := map[int]bool{}
	for _, e := range ents {
		var n int
		if _, err := fmt.Sscanf(e.Name(), "m%d", &n); err == nil && e.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, e.Name(), "char.json")); err == nil {
				m[n] = true
			}
		}
	}
	return m
}
