// Command parity measures how much of classic 1.65 Midgard the server
// implements: it reads the community database dump for what the realm's
// creatures, classes and styles ask for, and the rules registry for what the
// server does.
//
//	parity -dump server/data/community/opendaoc-db-core -out docs/parity
//
// It writes parity.json, the baseline the regression test compares against,
// and README.md, the same for reading. Both hold names, counts and statuses
// only, sorted, so a re-run on the same dump and code gives the same bytes.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dump := flag.String("dump", filepath.Join("server", "data", "community", "opendaoc-db-core"), "the community database's table directory")
	out := flag.String("out", filepath.Join("docs", "parity"), "directory for parity.json and README.md")
	flag.Parse()
	if _, err := os.Stat(filepath.Join(*dump, "mob.sql")); err != nil {
		fmt.Fprintf(os.Stderr, "parity: no community database at %s: %v\n", *dump, err)
		os.Exit(1)
	}
	rep, err := Run(*dump)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parity:", err)
		os.Exit(1)
	}
	if err := write(rep, *out); err != nil {
		fmt.Fprintln(os.Stderr, "parity:", err)
		os.Exit(1)
	}
	summary(os.Stdout, rep)
	fmt.Printf("\nwrote %s and %s\n", filepath.Join(*out, "parity.json"), filepath.Join(*out, "README.md"))
}

func write(rep *Report, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	js, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "parity.json"), append(js, '\n'), 0o644); err != nil {
		return err
	}
	var b strings.Builder
	readme(&b, rep)
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(b.String()), 0o644)
}

// done counts the implemented items of a list and their share of the uses.
func done(l []Item) (n, uses, total int) {
	for _, it := range l {
		total += it.Uses
		if it.Implemented {
			n++
			uses += it.Uses
		}
	}
	return
}

var sections = []struct {
	title, uses string
	get         func(*Report) []Item
}{
	{"Creature spell types", "placements x spells", func(r *Report) []Item { return r.CreatureSpells }},
	{"Class spell types", "spells", func(r *Report) []Item { return r.ClassSpells }},
	{"Style features", "styles", func(r *Report) []Item { return r.StyleFeatures }},
	{"Abilities", "grants", func(r *Report) []Item { return r.Abilities }},
	{"Creature kinds", "placements", func(r *Report) []Item { return r.CreatureKinds }},
}

func summary(w io.Writer, rep *Report) {
	fmt.Fprintf(w, "Midgard (regions 100, 101, 151): %d placements\n", rep.Placements)
	for _, s := range sections {
		l := s.get(rep)
		n, uses, total := done(l)
		fmt.Fprintf(w, "  %-22s %3d of %3d implemented (%d of %d %s)\n", s.title, n, len(l), uses, total, s.uses)
	}
	for _, c := range rep.Content {
		fmt.Fprintf(w, "  %-22s %d, none implemented\n", c.Name, c.Uses)
	}
	fmt.Fprintf(w, "  rules: %d documented, %d reference, %d provisional\n",
		rep.RulesByStatus["documented"], rep.RulesByStatus["reference"], rep.RulesByStatus["provisional"])
}

func readme(b *strings.Builder, rep *Report) {
	b.WriteString("# Parity census\n\n")
	b.WriteString("How much of classic 1.65 Midgard the server implements, measured against the community database by `go run ./cmd/parity`. ")
	b.WriteString("Counts and names only; the rules are in `openspec/specs/rules-*` and `internal/server/rules`.\n\n")
	b.WriteString("## Summary\n\n```\n")
	summary(b, rep)
	b.WriteString("```\n")
	for _, s := range sections {
		l := s.get(rep)
		fmt.Fprintf(b, "\n## %s\n\n| name | %s | implemented |\n| --- | --- | --- |\n", s.title, s.uses)
		for _, it := range l {
			mark := ""
			if it.Implemented {
				mark = "yes"
			}
			name := it.Name
			if len(it.Classes) > 0 {
				name += " (" + strings.Join(it.Classes, ", ") + ")"
			}
			fmt.Fprintf(b, "| %s | %d | %s |\n", name, it.Uses, mark)
		}
	}
	b.WriteString("\n## Rules\n\n| rule | status |\n| --- | --- |\n")
	for _, r := range rep.Rules {
		fmt.Fprintf(b, "| %s | %s |\n", r.ID, r.Status)
	}
}
