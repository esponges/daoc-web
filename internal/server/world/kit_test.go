package world

import (
	"fmt"
	"strings"
	"testing"

	"daocweb/internal/server/store"
)

// kitRig is a rig whose catalog lists starting equipment, and a fresh
// Warrior with nothing yet. Logged lines are collected.
func kitRig(t *testing.T, kit ...string) (*rig, *store.Character, *[]string) {
	r := newRig(t)
	var logged []string
	r.w.Log = func(f string, a ...any) { logged = append(logged, fmt.Sprintf(f, a...)) }
	r.w.Cat.StartEquip = map[int][]string{22: kit}
	a, _ := r.st.AccountFor("d:kit", "kit", false)
	c := &store.Character{AccountID: a.ID, Name: "Newcomer", Race: 6, Class: 22, Level: 1, Zone: 100, X: 30000, Y: 30000, XPScheme: store.XPSchemeRules}
	if err := r.st.CreateCharacter(c); err != nil {
		t.Fatal(err)
	}
	return r, c, &logged
}

func slotsOf(t *testing.T, st *store.Store, id int64) map[int]string {
	t.Helper()
	items, err := st.Items(id)
	if err != nil {
		t.Fatal(err)
	}
	out := map[int]string{}
	for _, it := range items {
		out[it.Slot] = it.ItemID
	}
	return out
}

func TestStarterKitFromTable(t *testing.T) {
	r, c, logged := kitRig(t, "vest", "leggings", "sleeves", "boots", "gloves", "cap", "shield", "axe")
	if err := r.w.GiveStarterKit(c); err != nil {
		t.Fatal(err)
	}
	want := map[int]string{SlotTorso: "vest", SlotLegs: "leggings", SlotArms: "sleeves", SlotFeet: "boots",
		SlotHands: "gloves", SlotHead: "cap", SlotShield: "shield", SlotWeapon: "axe"}
	if got := slotsOf(t, r.st, c.ID); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("kit %v, want %v and nothing else", got, want)
	}
	if len(*logged) > 0 {
		t.Errorf("logged %v", *logged)
	}
}

// The first weapon by table order is worn, the second goes to the pack, as
// does armour the class cannot wear yet.
func TestStarterKitSecondWeaponInPack(t *testing.T) {
	r, c, _ := kitRig(t, "axe", "hammer", "hauberk")
	r.w.GiveStarterKit(c)
	got := slotsOf(t, r.st, c.ID)
	if got[SlotWeapon] != "axe" || got[0] != "hammer" || got[1] != "hauberk" || len(got) != 3 {
		t.Errorf("kit %v: want the axe worn, the hammer and the chain (too heavy at level 1) in the pack", got)
	}
}

func TestStarterKitUnknownItemLogged(t *testing.T) {
	r, c, logged := kitRig(t, "axe", "no_such_item")
	r.w.GiveStarterKit(c)
	if got := slotsOf(t, r.st, c.ID); len(got) != 1 || got[SlotWeapon] != "axe" {
		t.Errorf("kit %v, want only the axe", got)
	}
	if len(*logged) != 1 || !strings.Contains((*logged)[0], `"no_such_item"`) {
		t.Errorf("logged %v, want the missing id named", *logged)
	}
}

func TestStarterKitNothingListed(t *testing.T) {
	r, c, logged := kitRig(t)
	r.w.GiveStarterKit(c)
	if got := slotsOf(t, r.st, c.ID); len(got) != 0 {
		t.Errorf("kit %v, want nothing", got)
	}
	if len(*logged) != 1 || !strings.Contains((*logged)[0], "starts with nothing") {
		t.Errorf("logged %v", *logged)
	}
}

// A character made with the old kit keeps exactly what it has on joining.
func TestExistingCharacterKeepsOldKit(t *testing.T) {
	r, c, _ := kitRig(t, "vest", "axe")
	r.st.Tx(func(tx *store.Tx) error {
		tx.AddItem(c.ID, "hammer", SlotWeapon, 1)
		tx.AddItem(c.ID, "shield", SlotShield, 1)
		return nil
	})
	before := slotsOf(t, r.st, c.ID)
	a, _ := r.st.AccountFor("d:kit", "kit", false)
	r.z.join(newPlayer(a, c))
	if after := slotsOf(t, r.st, c.ID); fmt.Sprint(after) != fmt.Sprint(before) {
		t.Errorf("inventory %v became %v", before, after)
	}
}
