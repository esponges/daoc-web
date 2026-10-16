package main

import "testing"

func TestArmourLevel(t *testing.T) {
	cases := map[int]int{32: 0, 33: 1, 34: 2, 37: 2, 35: 3, 38: 3, 36: 4}
	for typ, want := range cases {
		if got, ok := armourLevel(typ); !ok || got != want {
			t.Errorf("armourLevel(%d) = %d, %v; want %d", typ, got, ok, want)
		}
	}
	if _, ok := armourLevel(10); ok {
		t.Error("a weapon type must not be armour")
	}
}

// skinsFor builds a made-up pskins set where each listed family has every part.
func skinsFor(families ...string) map[string]bool {
	m := map[string]bool{}
	for _, f := range families {
		for _, p := range armourParts {
			m[f+"_"+p] = true
		}
	}
	return m
}

func TestPickSet(t *testing.T) {
	skins := skinsFor("b_plt1", "b_plt3", "n_chn2", "n_lea1", "h_scl1", "h_ren1")
	skins["b_chn0_body1m"] = true // a family missing parts is not usable
	set, lv, ok := pickSet(skins, "b", 4)
	if !ok || lv != 4 || set.Body != "b_plt1_body1m" || set.LBody != "b_plt1_lbody1" {
		t.Errorf("Albion plate = %+v level %d ok %v", set, lv, ok)
	}
	// The lowest number that has every part wins.
	if set, _, _ := pickSet(skinsFor("b_plt3", "b_plt2"), "b", 4); set.Arms != "b_plt2_arm1" {
		t.Errorf("lowest number: %+v", set)
	}
	// Midgard has no plate: it falls back to its chain.
	if set, lv, ok := pickSet(skins, "n", 4); !ok || lv != 3 || set.Legs != "n_chn2_leg1" {
		t.Errorf("Midgard plate fell back to %+v level %d ok %v", set, lv, ok)
	}
	// Hibernia's fourth tier is scale and its third reinforced.
	if set, _, _ := pickSet(skins, "h", 3); set.Boots != "h_scl1_boot1" {
		t.Errorf("Hibernia level 3 = %+v", set)
	}
	if set, _, _ := pickSet(skins, "h", 2); set.Boots != "h_ren1_boot1" {
		t.Errorf("Hibernia level 2 = %+v", set)
	}
	// A realm with nothing at or below the level has no set.
	if _, _, ok := pickSet(skinsFor("b_plt1"), "n", 4); ok {
		t.Error("expected no Midgard set")
	}
}

func TestArmourPieces(t *testing.T) {
	// {item slot, model, object type, realm}: model 47 is Albion plate legs;
	// model 36 torso is used by leather items of two realms and one cloth.
	rows := [][4]int{
		{27, 47, 36, 1}, {27, 47, 36, 1},
		{25, 36, 33, 3}, {25, 36, 33, 1}, {25, 36, 33, 3}, {25, 36, 32, 2},
		{25, 99, 10, 1}, // not armour
		{22, 50, 36, 1}, // not worn by anyone wanted
	}
	wanted := map[[2]int]bool{{27, 47}: true, {25, 36}: true, {25, 99}: true}
	got := armourPieces(wanted, rows)
	if p := got["27:47"]; p.Level != 4 || p.Realm != 1 {
		t.Errorf("27:47 = %+v", p)
	}
	if p := got["25:36"]; p.Level != 1 || p.Realm != 3 {
		t.Errorf("25:36 = %+v, want leather of the commonest realm", p)
	}
	if _, ok := got["25:99"]; ok {
		t.Error("a non-armour item must not classify")
	}
	if _, ok := got["22:50"]; ok {
		t.Error("an unwanted number must not appear")
	}
}
