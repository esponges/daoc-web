package world

import (
	"encoding/json"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
)

// lastDelve is the delveinfo a player was last sent, if any, and the text of
// the last refusal (err) sent with it. It drains the player's outbox.
func lastDelve(p *Player) (protocol.DelveInfo, bool) {
	d, ok, _ := delveReply(p)
	return d, ok
}

func delveReply(p *Player) (d protocol.DelveInfo, ok bool, refusal string) {
	types, raws := sent(p)
	for i := range types {
		switch types[i] {
		case protocol.TDelveInfo:
			d, ok = protocol.DelveInfo{}, true
			json.Unmarshal(raws[i], &d)
		case protocol.TErr:
			var e protocol.Err
			json.Unmarshal(raws[i], &e)
			refusal = e.Text
		}
	}
	return d, ok, refusal
}

func has2(lines []string, sub string) bool {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}

func TestDelveDescribesWhatThePlayerCanSee(t *testing.T) {
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, true)
	sent(p)

	// A worn weapon: its level, damage and speed, and gear is con coloured.
	r.z.Delve(p, &protocol.Delve{T: protocol.TDelve, From: "inv", Slot: SlotWeapon})
	d, ok := lastDelve(p)
	if !ok || d.Name != "bronze hammer" || d.Level != 2 || !d.Con {
		t.Fatalf("worn weapon: %+v ok %v", d, ok)
	}
	if !has2(d.Lines, "Level 2") || !has2(d.Lines, "1.8 DPS") || !has2(d.Lines, "3.1 s swing") || !has2(d.Lines, "Condition") {
		t.Errorf("weapon lines = %v", d.Lines)
	}

	// An empty slot says so.
	r.z.Delve(p, &protocol.Delve{T: protocol.TDelve, From: "inv", Slot: 7})
	if _, ok, msg := delveReply(p); ok || msg == "" {
		t.Errorf("an empty slot: details %v, refusal %q", ok, msg)
	}

	// A shop's item, only while that shop sells it.
	r.z.Delve(p, &protocol.Delve{T: protocol.TDelve, From: "shop", ID: "shield"})
	if _, ok, msg := delveReply(p); ok || msg == "" {
		t.Errorf("a shop item with no shop open: details %v, refusal %q", ok, msg)
	}
	r.z.Interact(p, merchantID)
	sent(p)
	r.z.Delve(p, &protocol.Delve{T: protocol.TDelve, From: "shop", ID: "shield"})
	if d, ok := lastDelve(p); !ok || d.Name != "round shield" || !d.Con {
		t.Errorf("shop item: %+v ok %v", d, ok)
	}
	r.z.Delve(p, &protocol.Delve{T: protocol.TDelve, From: "shop", ID: "free"}) // not in this shop's stock
	if _, ok, msg := delveReply(p); ok || msg == "" {
		t.Errorf("an item the shop does not sell: details %v, refusal %q", ok, msg)
	}

	// A spell the player does not have, and a thing that is not there.
	r.z.Delve(p, &protocol.Delve{T: protocol.TDelve, From: "spell", ID: "999999"})
	if _, ok, msg := delveReply(p); ok || msg == "" {
		t.Errorf("a spell not in the book: details %v, refusal %q", ok, msg)
	}
	r.z.Delve(p, &protocol.Delve{T: protocol.TDelve, From: "nowhere"})
	if _, ok, msg := delveReply(p); ok || msg == "" {
		t.Errorf("an unknown place: details %v, refusal %q", ok, msg)
	}
}

func TestItemViewCarriesWhatTheClientDrawsItBy(t *testing.T) {
	cases := []struct {
		it           *Item
		ot, ty, size int
	}{
		{&Item{ID: "a", Name: "great sword", ObjectType: 11, ItemType: 12}, 11, 12, 0},
		{&Item{ID: "b", Name: "composite bow", ObjectType: 15, ItemType: 13}, 15, 13, 0},
		{&Item{ID: "c", Name: "arrows", ObjectType: 43, ItemType: 40}, 43, 40, 0},
		{&Item{ID: "d", Name: "large shield", ObjectType: ObjShield, ItemType: 11, DamageType: 3}, ObjShield, 11, 3},
		{&Item{ID: "e", Name: "round shield", ObjectType: ObjShield, ItemType: 11, DamageType: 0}, ObjShield, 11, 0},
		{&Item{ID: "f", Name: "svarskodd boots", ObjectType: 34, ItemType: 23}, 34, 23, 0},
	}
	for _, c := range cases {
		v := itemView(0, c.it, 1)
		if v.ObjType != c.ot || v.ItemType != c.ty || v.Size != c.size {
			t.Errorf("%s: ot %d ty %d sz %d, want %d %d %d", c.it.Name, v.ObjType, v.ItemType, v.Size, c.ot, c.ty, c.size)
		}
	}
	b, _ := json.Marshal(itemView(0, cases[3].it, 1))
	for _, want := range []string{`"ot":42`, `"ty":11`, `"sz":3`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("%s lacks %s", b, want)
		}
	}
}

func TestDelveSpellFromTheBook(t *testing.T) {
	r := shopRig(t)
	p := r.player("Alice", 5, 30100, 30000, false)
	sp := &SpellDef{ID: 4242, Name: "Frost Bolt", Type: "DirectDamage", Target: "Enemy", Range: 1500, CastTime: 2.5, Damage: 40, DamageType: 11, Power: 10}
	p.book = []bookSpell{{Spec: "Runecarving", Level: 3, Spell: sp}}
	sent(p)
	r.z.Delve(p, &protocol.Delve{T: protocol.TDelve, From: "spell", ID: "4242"})
	d, ok := lastDelve(p)
	if !ok || d.Name != "Frost Bolt" || d.Con {
		t.Fatalf("spell: %+v ok %v", d, ok)
	}
	for _, want := range []string{"Level 3", "Range: 1500", "Cast time: 2.5 s", "Damage: 40 cold", "Power cost"} {
		if !has2(d.Lines, want) {
			t.Errorf("spell lines %v lack %q", d.Lines, want)
		}
	}
}
