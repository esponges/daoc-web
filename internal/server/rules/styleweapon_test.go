package rules

import "testing"

func TestStyleWeapons(t *testing.T) {
	const c = "rules-melee"
	h2h := func(h Held) bool { return StyleWeaponOK(StyleTwoWeapons, "Hand to Hand", h) }
	vector(t, c, "Claws in both hands", func(t *testing.T) {
		if !h2h(Held{ObjHandToHand, ObjHandToHand}) {
			t.Error("a Savage with two claws cannot use a Hand to Hand style")
		}
	})
	vector(t, c, "One claw", func(t *testing.T) {
		if h2h(Held{ObjHandToHand, 0}) {
			t.Error("one claw was enough for a Hand to Hand style")
		}
		if got := StyleWeaponNeeds(StyleTwoWeapons, "Hand to Hand"); got != "a hand to hand weapon in each hand" {
			t.Errorf("needs %q", got)
		}
	})
	vector(t, c, "An axe and a shield", func(t *testing.T) {
		if h2h(Held{ObjAxe, ObjShield}) {
			t.Error("an axe and a shield allowed a Hand to Hand style")
		}
		if !StyleWeaponOK(ObjAxe, "Axe", Held{ObjAxe, ObjShield}) || !StyleWeaponOK(ObjShield, "Shields", Held{ObjAxe, ObjShield}) {
			t.Error("the axe and shield styles are refused")
		}
		if StyleWeaponOK(ObjSword, "Sword", Held{ObjAxe, ObjShield}) {
			t.Error("a sword style with an axe")
		}
	})
	vector(t, c, "A left axe", func(t *testing.T) {
		if !StyleWeaponOK(StyleTwoWeapons, "Left Axe", Held{ObjSword, ObjLeftAxe}) {
			t.Error("a sword and a left axe cannot use a Left Axe style")
		}
		if StyleWeaponOK(StyleTwoWeapons, "Left Axe", Held{ObjSword, ObjShield}) || StyleWeaponOK(StyleTwoWeapons, "Left Axe", Held{ObjSword, 0}) {
			t.Error("a Left Axe style without a left axe")
		}
	})
	vector(t, c, "Any weapon", func(t *testing.T) {
		if !StyleWeaponOK(StyleAnyWeapon, "Critical Strike", Held{ObjSword, 0}) {
			t.Error("a sword cannot use a Critical Strike style")
		}
		if StyleWeaponOK(StyleAnyWeapon, "Critical Strike", Held{}) {
			t.Error("empty hands used a Critical Strike style")
		}
	})
}
