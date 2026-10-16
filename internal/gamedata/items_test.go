package gamedata

import "testing"

func TestItemKind(t *testing.T) {
	cases := map[string]string{
		"H_2h_CelticMace01":             KindTwoHand,
		"N_1h_warhammer01":              KindOneHand,
		"b_cr_mace01":                   KindMelee,
		"h_th_dagger01":                 KindMelee,
		"N_hth_BladedMoonClaw01":        KindMelee,
		"b_st_magestaff04":              KindTwoHand,
		"b_po_pike01":                   KindTwoHand,
		"H_bl_CelticMace01":             KindMelee,
		"H_sc_FirbolgGreatScythe01":     KindTwoHand,
		"B_bo_SaracenShortbow01_Energy": KindRanged,
		"B_Shield_KiteBuckler":          KindShield,
		"sh_rounda":                     KindShield,
		"M_Shield_Crescent":             KindShield,
		"Epic_wpn_palladin_slash_2H":    KindTwoHand,
		"Epic_wpn_palladin_slash_1H":    KindOneHand,
		"k_eviltentbig":                 KindUnknown,
		"int_floor_barrel_01":           KindUnknown,
		"statue_knight":                 KindUnknown, // "st" must be a whole token
		"in_flute":                      KindUnknown,
	}
	for nif, want := range cases {
		if got := ItemKind(nif); got != want {
			t.Errorf("ItemKind(%q) = %q, want %q", nif, got, want)
		}
	}
}

func TestKindFitsSlot(t *testing.T) {
	cases := []struct {
		kind string
		slot int
		want bool
	}{
		{KindTwoHand, 10, true}, // the database puts two-handers in the right hand
		{KindShield, 10, false}, // a shield in the right hand is wrong data
		{KindShield, 11, true},
		{KindTwoHand, 11, false},
		{KindRanged, 11, false}, // a bow in a hand slot
		{KindRanged, 13, true},
		{KindMelee, 12, true},
		{KindOneHand, 12, false},
		{KindUnknown, 10, false},
		{KindMelee, 25, false}, // not a held slot
	}
	for _, c := range cases {
		if got := KindFitsSlot(c.kind, c.slot); got != c.want {
			t.Errorf("KindFitsSlot(%q, %d) = %v, want %v", c.kind, c.slot, got, c.want)
		}
	}
}

func TestParseItems(t *testing.T) {
	rows := [][]string{
		{"311", "Hib 2H Mace", "H_2h_CelticMace01"},
		{"0", "zero id", "x"},
		{"5", "no model", ""},
		{"473", "Alb Shield 06", "B_Shield_KiteBuckler"},
	}
	got := ParseItems(rows)
	if len(got) != 2 || got[311].NIF != "H_2h_CelticMace01" || got[311].Kind != KindTwoHand || got[473].Kind != KindShield {
		t.Errorf("ParseItems = %+v", got)
	}
}

func TestParseTints(t *testing.T) {
	rows := [][]string{{"0", "White", "255", "255", "255"}, {"17", "Iron", "125", "126", "133", "1", "2", "3"}}
	got := ParseTints(rows)
	if got[0] != [3]int{255, 255, 255} || got[17] != [3]int{125, 126, 133} {
		t.Errorf("ParseTints = %v", got)
	}
}
