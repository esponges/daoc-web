// Package rules holds the game's rules: the formulas that turn levels, stats,
// items and dice into hit points, damage, experience and the rest. They are
// this project's own, written from the rule book (the rules-* capabilities in
// openspec/specs), which says for each one where it comes from.
//
// Nothing here knows about zones, clocks or connections, and nothing draws its
// own random numbers: a caller passes the rolls, so every worked example in the
// rule book is a table test.
package rules

import "sort"

// Status is how sure a rule is.
type Status int

const (
	// Documented: a public source states the rule.
	Documented Status = iota
	// Reference: no public source gives the number; it matches the classic
	// reference server, read only to understand it.
	Reference
	// Provisional: neither exists and the value is this project's choice.
	Provisional
)

func (s Status) String() string {
	switch s {
	case Documented:
		return "documented"
	case Reference:
		return "reference"
	}
	return "provisional"
}

// Rule is one entry of the registry: a mechanic the server implements.
type Rule struct {
	ID          string   // e.g. "melee.parry"
	Capability  string   // the rules-* spec, e.g. "rules-melee"
	Requirement string   // the requirement's name in that spec
	Status      Status   //
	Sources     []string // keys of Sources
	Scenarios   []string // scenario names in the spec that are table tests
}

// Sources are the places rules come from, by key.
var Sources = map[string]string{
	"disorder":      "disorder.dk, Dark Age of Camelot 1.65 mechanics pages",
	"herald":        "Camelot Herald wiki (melee damage, weapon skill, styles)",
	"grabbag":       "Mythic's Friday Grab Bag answers",
	"kb-xpclamp":    "Mythic knowledge base: experience clamp raised to 1.25 times a same-level kill",
	"silicondragon": "silicondragon.com 'XPs' tables: experience and constitution restore",
	"mythic-con":    "Mythic's description of con colour bands (level / 10 + 1)",
	"opendaoc":      "OpenDAoC-Core, read to understand (GPL; nothing copied)",
	"data":          "the imported community database",
	"project":       "this project's choice",
}

var registry []Rule

// register adds rules to the registry; each rules file calls it from init.
func register(rs ...Rule) { registry = append(registry, rs...) }

// Registry is every implemented rule, sorted by id.
func Registry() []Rule {
	out := append([]Rule(nil), registry...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Implemented says whether a rule id is registered.
func Implemented(id string) bool {
	for _, r := range registry {
		if r.ID == id {
			return true
		}
	}
	return false
}
