package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/importer"
	"daocweb/internal/server/rules"
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

// count says how many of items are implemented, as the README table writes it.
func count(items []Item) (done, total, uses int) {
	for _, it := range items {
		total++
		if it.Implemented {
			done++
			uses += it.Uses
		}
	}
	return
}

// The README's status table is written by hand from docs/parity/parity.json;
// this keeps it from going stale. It needs no dump, so -short runs it.
func TestReadmeAgreesWithBaseline(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "parity", "parity.json"))
	if err != nil {
		t.Fatal(err)
	}
	var base Report
	if err := json.Unmarshal(b, &base); err != nil {
		t.Fatal(err)
	}
	readme, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(readme)
	want := map[string][]Item{
		"Creature spell types": base.CreatureSpells, "Class spell types": base.ClassSpells,
		"Style features": base.StyleFeatures, "Abilities": base.Abilities, "Creature kinds": base.CreatureKinds,
	}
	for name, items := range want {
		done, total, _ := count(items)
		cell := fmt.Sprintf("| %s | %d of %d", name, done, total)
		if !strings.Contains(text, cell) {
			t.Errorf("README status table should have a row starting %q", cell)
		}
	}
	for _, it := range base.Content {
		if it.Implemented {
			continue
		}
		if !strings.Contains(text, "0 of "+fmt.Sprint(it.Uses)) {
			t.Errorf("README should list %s as 0 of %d", it.Name, it.Uses)
		}
	}
	rules := fmt.Sprintf("%d documented, %d reference, %d provisional",
		base.RulesByStatus["documented"], base.RulesByStatus["reference"], base.RulesByStatus["provisional"])
	if !strings.Contains(text, rules) {
		t.Errorf("README should say rules: %s", rules)
	}
}

// spellProblem says what is wrong with the amount the server would deal or
// restore for one imported spell, from a reference caster (level 20, 100 in the
// casting stat, spec and target at 20, the middle of the range, 720 health), or
// "" when it is as the rule gives. A type with no amount has no problem.
func spellProblem(id int, typ string, damage, value float64) string {
	const stat, level, health = 100, 20, 720
	switch typ {
	case rules.SpellDirectDamage, rules.SpellBolt, rules.SpellLifedrain:
		got := rules.PlayerSpellDamage(damage, stat, level, level, level, 0.5, 0)
		if damage > 0 && got > 0 && float64(got) <= damage*rules.PlayerSpellCapFactor {
			return ""
		}
		return fmt.Sprintf("spell %d (%s): damage %v gives %d, want 1 to %v", id, typ, damage, got, damage*rules.PlayerSpellCapFactor)
	case rules.SpellDamageOverTime:
		got := rules.PlayerDoTDamage(damage, stat, level, level, 0.5, 0)
		if damage > 0 && got > 0 && float64(got) <= damage*rules.PlayerSpellCapFactor {
			return ""
		}
		return fmt.Sprintf("spell %d (%s): damage %v gives a tick of %d, want 1 to %v", id, typ, damage, got, damage*rules.PlayerSpellCapFactor)
	case rules.SpellDamageAdd:
		got := rules.DamageAdd(damage, level, 3, 40, 0.5, 0)
		if damage != 0 && got > 0 && got <= 40+int(damage*3) {
			return ""
		}
		return fmt.Sprintf("spell %d (%s): damage %v gives %d on a 3 second swing and a blow of 40", id, typ, damage, got)
	case rules.SpellDDNoVariance:
		if damage > 0 {
			return ""
		}
		return fmt.Sprintf("spell %d (%s): damage %v gives nothing", id, typ, damage)
	case rules.SpellHeal:
		got := rules.PlayerHeal(value, level, level, 0.5)
		if value < 0 {
			got = rules.PercentHeal(value, health)
			if got > 0 && got <= health {
				return ""
			}
			return fmt.Sprintf("spell %d (%s): value %v gives %d, want 1 to %d", id, typ, value, got, health)
		}
		if got > 0 && float64(got) <= value*1.25 {
			return ""
		}
		return fmt.Sprintf("spell %d (%s): value %v gives %d, want 1 to %v", id, typ, value, got, value*1.25)
	}
	return ""
}

func TestSpellCheckFlagsAMismatch(t *testing.T) {
	if p := spellProblem(1, rules.SpellHeal, 0, 120); p != "" {
		t.Errorf("a plain heal is flagged: %s", p)
	}
	if p := spellProblem(2, rules.SpellHeal, 0, 0); !strings.Contains(p, "spell 2") || !strings.Contains(p, "value 0") {
		t.Errorf("a heal of nothing is not named: %q", p)
	}
	if p := spellProblem(3, rules.SpellBolt, -5, 0); !strings.Contains(p, "spell 3") {
		t.Errorf("a bolt of negative damage is not named: %q", p)
	}
	if p := spellProblem(4, rules.SpellHeal, 0, -40); p != "" {
		t.Errorf("a percent heal is flagged: %s", p)
	}
	if p := spellProblem(5, rules.SpellDamageOverTime, 10, 0); p != "" {
		t.Errorf("a damage over time is flagged: %s", p)
	}
	if p := spellProblem(6, rules.SpellDamageOverTime, 0, 0); !strings.Contains(p, "spell 6") {
		t.Errorf("a damage over time of nothing is not named: %q", p)
	}
	if p := spellProblem(7, rules.SpellDamageAdd, 0, 0); !strings.Contains(p, "spell 7") {
		t.Errorf("a damage add of nothing is not named: %q", p)
	}
	if p := spellProblem(8, rules.SpellDamageAdd, 4.5, 0); p != "" {
		t.Errorf("a damage add is flagged: %s", p)
	}
}

// Every spell of an implemented type in the community database has an amount
// the rules can give. Skipped with -short or without the dump.
func TestSpellAmountsAgreeWithTheTable(t *testing.T) {
	if testing.Short() {
		t.Skip("reads the spell table")
	}
	dump := filepath.Join("..", "..", "server", "data", "community", "opendaoc-db-core", "spell.sql")
	if _, err := os.Stat(dump); err != nil {
		t.Skip("no community database checked out in server/data/community")
	}
	n := 0
	if _, err := importer.ReadTable(dump, func(r *importer.Row) error {
		typ := r.Str("Type")
		if !rules.ImplementedPlayerSpell(typ) {
			return nil
		}
		n++
		if p := spellProblem(r.Int("SpellID"), typ, r.Float("Damage"), r.Float("Value")); p != "" {
			t.Error(p)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("checked %d spells", n)
}
