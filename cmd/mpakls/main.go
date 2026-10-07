// Command mpakls lists what is inside an MPAK/NPK archive, and can extract
// entries. Useful for finding where an asset actually lives, since DAoC
// scatters content across hundreds of these.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"daocweb/internal/mpak"
)

func main() {
	grep := flag.String("grep", "", "only entries whose name contains this (case-insensitive)")
	extract := flag.String("x", "", "extract matching entries into this directory")
	quiet := flag.Bool("q", false, "only print archives that have a match")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: mpakls [-grep substr] [-x outdir] [-q] <archive.mpk>...")
		os.Exit(2)
	}

	for _, path := range flag.Args() {
		a, err := mpak.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(path), err)
			continue
		}
		hits := a.Entries
		if *grep != "" {
			hits = nil
			for _, e := range a.Entries {
				if strings.Contains(strings.ToLower(e.Name), strings.ToLower(*grep)) {
					hits = append(hits, e)
				}
			}
		}
		if len(hits) == 0 && *quiet {
			continue
		}
		fmt.Printf("%s  (%d entries", filepath.Base(path), len(a.Entries))
		if *grep != "" {
			fmt.Printf(", %d matching", len(hits))
		}
		fmt.Println(")")
		for _, e := range hits {
			fmt.Printf("    %-40s %8d bytes\n", e.Name, e.UncSize)
			if *extract != "" {
				b, err := a.ReadEntry(e)
				if err != nil {
					fmt.Fprintf(os.Stderr, "      %v\n", err)
					continue
				}
				if err := os.MkdirAll(*extract, 0o755); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				dst := filepath.Join(*extract, filepath.Base(e.Name))
				if err := os.WriteFile(dst, b, 0o644); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}
	}
}
