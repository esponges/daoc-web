// Command census measures how much of an installed copy of the game this
// pipeline can read and convert.
//
// It walks the install, reads every item of each asset class through the same
// code the converters use -- internal/nif, internal/kfa, internal/scenery,
// internal/zone -- and records which items pass and why the rest fail, grouped
// by cause. Outdoor zones are dry-run: terrain read and every placed model
// baked in memory, nothing written.
//
//	census -game "C:\...\Dark Age of Camelot" -out docs/census
//
// It writes census.json, the baseline the regression test compares against,
// and README.md, the same data for reading. Both hold names, counts and causes
// only, sorted, with no timestamps or paths outside the install, so a re-run
// on the same install and code gives the same bytes.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	out := flag.String("out", filepath.Join("docs", "census"), "directory for census.json and README.md")
	workers := flag.Int("j", runtime.NumCPU(), "items read in parallel")
	flag.Parse()

	if _, err := os.Stat(*game); err != nil {
		fmt.Fprintf(os.Stderr, "census: no install at %s: %v\n", *game, err)
		os.Exit(1)
	}
	rep, err := Run(*game, *workers)
	if err != nil {
		fmt.Fprintln(os.Stderr, "census:", err)
		os.Exit(1)
	}
	if err := write(rep, *out); err != nil {
		fmt.Fprintln(os.Stderr, "census:", err)
		os.Exit(1)
	}
	summary(os.Stdout, rep)
	fmt.Printf("\nwrote %s and %s\n", filepath.Join(*out, "census.json"), filepath.Join(*out, "README.md"))
}

func write(rep *Report, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	js, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "census.json"), append(js, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(markdown(rep)), 0o644)
}
