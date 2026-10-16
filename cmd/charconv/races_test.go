package main

import (
	"testing"

	"daocweb/internal/gamedata"
)

func TestRaceRows(t *testing.T) {
	mon := map[int]gamedata.Monster{
		137: {ID: 137, Name: "Troll Male"},
		138: {ID: 138, Name: "Troll Male 2"}, // a head variant, not the base
		145: {ID: 145, Name: "Troll Female"},
		90:  {ID: 90, Name: "Troll Male Merchant 1"},
		50:  {ID: 50, Name: "Norse Male"},
		51:  {ID: 51, Name: "Norse Male"}, // the first row wins
		10:  {ID: 10, Name: "Briton Female"},
		7:   {ID: 7, Name: "Large Grey Wolf"},
		169: {ID: 169, Name: "Kobold Male 1"}, // some races number every row
		170: {ID: 170, Name: "Kobold Male 2"},
	}
	got := raceRows(mon)
	if got[6] != [2]int{137, 145} {
		t.Errorf("Troll = %v, want {137 145}", got[6])
	}
	if got[5] != [2]int{50, 0} {
		t.Errorf("Norse = %v, want a male only", got[5])
	}
	if got[1] != [2]int{0, 10} {
		t.Errorf("Briton = %v, want a female only", got[1])
	}
	if got[8] != [2]int{169, 0} {
		t.Errorf("Kobold = %v, want the first numbered row", got[8])
	}
	if _, ok := got[7]; ok {
		t.Errorf("Dwarf has no rows and should be absent")
	}
}

func TestBuildRacesByRealm(t *testing.T) {
	mon := map[int]gamedata.Monster{
		137: {ID: 137, Name: "Troll Male"},
		10:  {ID: 10, Name: "Briton Female"},
	}
	f, convert := buildRaces(mon, 2)
	if len(f.Races) != 1 || f.Races["6"].Male != "m137" || f.Races["6"].Female != "" {
		t.Errorf("Midgard file = %+v", f)
	}
	if len(convert) != 1 || convert[0] != 137 {
		t.Errorf("convert = %v", convert)
	}
	if all, _ := buildRaces(mon, 0); len(all.Races) != 2 {
		t.Errorf("all realms = %+v", all)
	}
}
