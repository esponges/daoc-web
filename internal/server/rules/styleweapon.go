package rules

import "strings"

// What a style needs in the hands, from the rules-melee capability ("A style
// needs its weapons"). The imported style table names a weapon object type, a
// shield, or one of two special requirements below. The checks are this
// project's own, written after reading how the classic reference server treats
// the two special codes (reference).

const (
	StyleTwoWeapons = 1000 // a weapon in each hand: Hand to Hand, Left Axe, Dual Wield, Celtic Dual
	StyleAnyWeapon  = 1001 // any weapon in the main hand: Critical Strike
)

// Held is what a character has in their hands, by object type: 0 for nothing.
type Held struct {
	Main, Left int
}

// StyleWeaponOK says whether a style of a line (its spec name) with weapon
// requirement req can be used with what is held.
func StyleWeaponOK(req int, spec string, h Held) bool {
	switch req {
	case ObjShield:
		return h.Left == ObjShield
	case StyleAnyWeapon:
		return h.Main != 0
	case StyleTwoWeapons:
		if h.Main == 0 || h.Left == 0 || h.Left == ObjShield {
			return false
		}
		switch spec {
		case "Hand to Hand":
			return h.Main == ObjHandToHand && h.Left == ObjHandToHand
		case "Left Axe":
			return h.Left == ObjLeftAxe
		}
		return true
	}
	return h.Main == req
}

// StyleWeaponNeeds says in words what a style's requirement asks for, for a
// style the character cannot use.
func StyleWeaponNeeds(req int, spec string) string {
	switch req {
	case ObjShield:
		return "a shield"
	case StyleAnyWeapon:
		return "a weapon"
	case StyleTwoWeapons:
		switch spec {
		case "Hand to Hand":
			return "a hand to hand weapon in each hand"
		case "Left Axe":
			return "a left axe in the left hand"
		}
		return "a weapon in each hand"
	}
	if spec != "" {
		return "a " + strings.ToLower(spec) + " weapon"
	}
	return "another weapon"
}

func init() {
	register(Rule{ID: "melee.style-weapon", Capability: "rules-melee", Requirement: "A style needs its weapons",
		Status: Reference, Sources: []string{"opendaoc", "dolsharp"},
		Scenarios: []string{"Claws in both hands", "One claw", "An axe and a shield", "A left axe", "Any weapon"}})
}
