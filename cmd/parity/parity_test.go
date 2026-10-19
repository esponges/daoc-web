package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// regressions lists what fresh has lost against the baseline: an item that was
// implemented and is not, a rule that is gone, or a rule whose status weakened
// (documented to reference or provisional, reference to provisional). New
// counts from a newer dump are not regressions.
func regressions(base, fresh *Report) []string {
	var out []string
	lists := func(r *Report) map[string][]Item {
		return map[string][]Item{"creature spell type": r.CreatureSpells, "class spell type": r.ClassSpells,
			"style feature": r.StyleFeatures, "ability": r.Abilities, "creature kind": r.CreatureKinds}
	}
	fl := lists(fresh)
	for kind, items := range lists(base) {
		now := map[string]bool{}
		for _, it := range fl[kind] {
			now[it.Name] = it.Implemented
		}
		for _, it := range items {
			if it.Implemented && !now[it.Name] {
				out = append(out, kind+" "+it.Name+" is no longer implemented")
			}
		}
	}
	rank := map[string]int{"documented": 0, "reference": 1, "provisional": 2}
	status := map[string]string{}
	for _, r := range fresh.Rules {
		status[r.ID] = r.Status
	}
	for _, r := range base.Rules {
		s, ok := status[r.ID]
		switch {
		case !ok:
			out = append(out, "rule "+r.ID+" is no longer registered")
		case rank[s] > rank[r.Status]:
			out = append(out, "rule "+r.ID+" went from "+r.Status+" to "+s)
		}
	}
	return out
}

func TestRegressionsAreNamed(t *testing.T) {
	base := &Report{Abilities: []Item{{Name: "Evade", Implemented: true}},
		Rules: []RuleRow{{ID: "melee.parry", Status: "documented"}, {ID: "melee.evade", Status: "documented"}}}
	fresh := &Report{Abilities: []Item{{Name: "Evade"}}, Rules: []RuleRow{{ID: "melee.evade", Status: "provisional"}}}
	got := strings.Join(regressions(base, fresh), "\n")
	for _, want := range []string{"ability Evade is no longer implemented", "rule melee.parry is no longer registered", "rule melee.evade went from documented to provisional"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if r := regressions(base, base); len(r) != 0 {
		t.Errorf("a report regressed against itself: %v", r)
	}
}

// The census against the committed baseline: skipped with -short (it reads the
// whole dump) or without the dump.
func TestParityDoesNotGoBackwards(t *testing.T) {
	if testing.Short() {
		t.Skip("reads the whole community database")
	}
	dump := filepath.Join("..", "..", "server", "data", "community", "opendaoc-db-core")
	if _, err := os.Stat(filepath.Join(dump, "mob.sql")); err != nil {
		t.Skip("no community database checked out in server/data/community")
	}
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "parity", "parity.json"))
	if err != nil {
		t.Fatalf("no baseline: run go run ./cmd/parity and commit docs/parity (%v)", err)
	}
	var base Report
	if err := json.Unmarshal(b, &base); err != nil {
		t.Fatal(err)
	}
	fresh, err := Run(dump)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range regressions(&base, fresh) {
		t.Error(r)
	}
}
