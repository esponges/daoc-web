// Scratch: how many mobs does ReplaceMobValues change, and how?
package main

import (
	"fmt"
	"strconv"
	"strings"

	"daocweb/internal/importer"
)

type tmpl struct {
	levels  []int
	aggro   int
	replace bool
	name    string
}

func ints(s string) []int {
	var out []int
	for _, p := range strings.FieldsFunc(s, func(r rune) bool { return r == ';' || r == ',' }) {
		if v, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func main() {
	dir := "server/data/community/opendaoc-db-core/"
	t := map[int][]tmpl{}
	nRep := 0
	listLevels := 0
	importer.ReadTable(dir+"npctemplate.sql", func(r *importer.Row) error {
		x := tmpl{levels: ints(r.Str("Level")), aggro: r.Int("AggroLevel"), replace: r.Int("ReplaceMobValues") != 0, name: r.Str("Name")}
		if x.replace {
			nRep++
		}
		if strings.ContainsAny(r.Str("Level"), ";-") {
			listLevels++
		}
		t[r.Int("TemplateId")] = append(t[r.Int("TemplateId")], x)
		return nil
	})
	fmt.Println("templates", len(t), "replace", nRep, "level lists", listLevels)
	var mobs, withT, repl, lvlDiff, aggroDiff, mid, midRepl, midDiff int
	big := map[string]int{}
	importer.ReadTable(dir+"mob.sql", func(r *importer.Row) error {
		mobs++
		ts := t[r.Int("NPCTemplateID")]
		midg := r.Int("Region") == 100 || r.Int("Region") == 101 || r.Int("Region") == 151
		if midg {
			mid++
		}
		if len(ts) == 0 {
			return nil
		}
		withT++
		x := ts[0]
		if !x.replace {
			return nil
		}
		repl++
		if midg {
			midRepl++
		}
		lv := r.Int("Level")
		in := false
		lo, hi := 999, -1
		for _, l := range x.levels {
			if l == lv {
				in = true
			}
			lo, hi = min(lo, l), max(hi, l)
		}
		if len(x.levels) > 0 && !in {
			lvlDiff++
			if midg {
				midDiff++
			}
			if d := lv - hi; d >= 10 || lo-lv >= 10 {
				big[fmt.Sprintf("%s mob %d tmpl %d-%d", r.Str("Name"), lv, lo, hi)]++
			}
		}
		if x.aggro != r.Int("AggroLevel") {
			aggroDiff++
		}
		return nil
	})
	fmt.Println("mobs", mobs, "with template", withT, "replace", repl, "level outside template", lvlDiff, "aggro differs", aggroDiff)
	fmt.Println("midgard regions mobs", mid, "replace", midRepl, "level differs", midDiff)
	n := 0
	for k, v := range big {
		if n < 25 {
			fmt.Println(" ", v, k)
		}
		n++
	}
	fmt.Println("distinct big diffs", n)
}
