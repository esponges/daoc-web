package protocol

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	cases := []interface{}{
		&Hello{T: THello, V: Version, Char: 7},
		&Input{T: TInput, Seq: 120, Steps: [][5]int{{30, 1, 0, 0, FlagRun}, {30, 1, 0, 1, 0}}, PX: 100.5, PY: -3},
		&Chat{T: TChat, Ch: "say", Text: "hello there"},
		&Target{T: TTarget, ID: 42},
		&Attack{T: TAttack, On: true},
		&Style{T: TStyle, Slot: 2},
		&Loot{T: TLoot, Corpse: 9, Index: -1},
		&Equip{T: TEquip, From: 3, To: 40},
		&Use{T: TUse},
		&GM{T: TGM, Line: "goto 103"},
		&Interact{T: TInteract, ID: 900},
		&Buy{T: TBuy, ID: "bronze_hammer"},
		&Sell{T: TSell, Slot: 4},
		&Train{T: TTrain, Spec: "Hammer", N: 2},
		&Restore{T: TRestore, NPC: 31},
	}
	for _, c := range cases {
		b, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		typ, got, err := Decode(b)
		if err != nil {
			t.Fatalf("%s: %v", b, err)
		}
		if want := reflect.ValueOf(c).Elem().FieldByName("T").String(); typ != want {
			t.Errorf("type %q, want %q", typ, want)
		}
		if !reflect.DeepEqual(got, c) {
			t.Errorf("round trip of %s:\n got  %+v\n want %+v", b, got, c)
		}
	}
}

func TestUnknownTypeIgnored(t *testing.T) {
	typ, v, err := Decode([]byte(`{"t":"from-the-future","x":1}`))
	if typ != "" || v != nil || err != nil {
		t.Errorf("unknown type: %q %v %v", typ, v, err)
	}
}

func TestBadJSON(t *testing.T) {
	if _, _, err := Decode([]byte(`not json`)); err == nil {
		t.Error("garbage must be an error")
	}
	if _, _, err := Decode([]byte(`{"t":"input","steps":"nope"}`)); err == nil {
		t.Error("a malformed body must be an error")
	}
}

func TestServerMessagesMarshal(t *testing.T) {
	s := Snap{T: TSnap, Tick: 5, Ack: 100, You: Self{X: 1, Y: 2, HP: 10}, Ents: []Entity{{ID: 1, K: "n", Name: "wolf", X: 3}}}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var back Snap
	if err := json.Unmarshal(b, &back); err != nil || !reflect.DeepEqual(back, s) {
		t.Errorf("snap round trip: %v\n%+v", err, back)
	}
}

func TestEntityEquipmentIsAdditive(t *testing.T) {
	geared := Entity{ID: 1, K: "n", Name: "guard", Equip: [][4]int{{25, 662, 17, 0}, {10, 311, 0, 0}}}
	b, err := json.Marshal(geared)
	if err != nil {
		t.Fatal(err)
	}
	var back Entity
	if err := json.Unmarshal(b, &back); err != nil || !reflect.DeepEqual(back, geared) {
		t.Errorf("round trip of %s: %+v, %v", b, back, err)
	}
	if b, _ := json.Marshal(Entity{ID: 2, K: "n", Name: "wolf"}); strings.Contains(string(b), `"eq"`) {
		t.Errorf("an ungeared entity must omit eq: %s", b)
	}
	// A client that predates the field decodes the same entity into its own shape.
	var old struct {
		ID   int64  `json:"id"`
		Name string `json:"nm"`
	}
	if err := json.Unmarshal(b, &old); err != nil || old.Name != "guard" {
		t.Errorf("old client: %+v, %v", old, err)
	}
}

func TestPlayerAppearanceIsAdditive(t *testing.T) {
	p := Entity{ID: 9, K: "p", Name: "Grimm", Char: "troll-warrior", Race: 6, Female: true, GearSet: true,
		Equip: [][4]int{{10, 311, 0, 0}}}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var back Entity
	if err := json.Unmarshal(b, &back); err != nil || !reflect.DeepEqual(back, p) {
		t.Errorf("round trip of %s: %+v, %v", b, back, err)
	}
	// A male with no race sent adds nothing to the entity.
	if b, _ := json.Marshal(Entity{ID: 1, K: "p", Name: "x"}); strings.Contains(string(b), `"rc"`) || strings.Contains(string(b), `"fm"`) || strings.Contains(string(b), `"gs"`) {
		t.Errorf("an entity without appearance must omit it: %s", b)
	}
	// The inventory always states what you wear, even when it is nothing.
	if b, _ := json.Marshal(Inv{T: TInv, Eq: [][4]int{}}); !strings.Contains(string(b), `"eq":[]`) {
		t.Errorf("inv must carry eq: %s", b)
	}
	// An older client decodes the same entity into its own shape.
	var old struct {
		ID   int64  `json:"id"`
		Name string `json:"nm"`
		Char string `json:"ch"`
	}
	if err := json.Unmarshal(b, &old); err != nil {
		t.Errorf("old client: %v", err)
	}
}

func TestShopMessagesAndInteractionKindAreAdditive(t *testing.T) {
	shop := Shop{T: TShop, NPC: 900, Name: "Smith", Items: []InvItem{
		{Slot: 0, ID: "hammer", Name: "bronze hammer", Count: 1, Kind: "weapon", Line: "Hammer", Detail: "1.8 dps", Price: 1250},
		{Slot: 1, ID: "shield", Name: "round shield", Count: 1, Kind: "shield"},
	}}
	b, err := json.Marshal(shop)
	if err != nil {
		t.Fatal(err)
	}
	var back Shop
	if err := json.Unmarshal(b, &back); err != nil || !reflect.DeepEqual(back, shop) {
		t.Errorf("shop round trip of %s: %+v, %v", b, back, err)
	}
	// A price is only on shop rows: inventory and corpse rows are unchanged.
	if b, _ := json.Marshal(InvItem{Slot: 3, ID: "pelt", Name: "wolf pelt", Count: 1}); strings.Contains(string(b), `"pr"`) {
		t.Errorf("an item with no price must omit pr: %s", b)
	}
	if b, err := json.Marshal(ShopClose{T: TShopClose}); err != nil || string(b) != `{"t":"shopclose"}` {
		t.Errorf("shopclose = %s, %v", b, err)
	}

	merchant := Entity{ID: 900, K: "n", Name: "Smith", Kind: "merchant"}
	b, err = json.Marshal(merchant)
	if err != nil || !strings.Contains(string(b), `"ik":"merchant"`) {
		t.Errorf("merchant entity = %s, %v", b, err)
	}
	var back2 Entity
	if err := json.Unmarshal(b, &back2); err != nil || back2.Kind != "merchant" {
		t.Errorf("ik did not round trip: %+v, %v", back2, err)
	}
	if b, _ := json.Marshal(Entity{ID: 2, K: "n", Name: "wolf"}); strings.Contains(string(b), `"ik"`) {
		t.Errorf("an entity with no interaction must omit ik: %s", b)
	}
	// A client that predates the field decodes the same entity into its own shape.
	var old struct {
		ID   int64  `json:"id"`
		Name string `json:"nm"`
	}
	if err := json.Unmarshal(b, &old); err != nil || old.Name != "Smith" {
		t.Errorf("old client: %+v, %v", old, err)
	}
}

func TestRuleBookMessagesAreAdditive(t *testing.T) {
	if typ, v, err := Decode([]byte(`{"t":"pray"}`)); typ != TPray || v != nil || err != nil {
		t.Errorf("pray decodes as %q %v %v", typ, v, err)
	}
	con := -3
	b, _ := json.Marshal(Entity{ID: 1, K: "n", Con: &con})
	if !strings.Contains(string(b), `"cc":-3`) {
		t.Errorf("con colour missing: %s", b)
	}
	b, _ = json.Marshal(Entity{ID: 2, K: "n"})
	if strings.Contains(string(b), "cc") || strings.Contains(string(b), "own") {
		t.Errorf("an older client's entity grew fields: %s", b)
	}
	b, _ = json.Marshal(Stats{T: TStats, Class: "Warrior", Specs: []SpecRow{{Key: "Sword", Name: "Sword", Level: 3}}, Abilities: []Ability{{Name: "Evade", Level: 1, On: true}}})
	for _, want := range []string{`"t":"stats"`, `"cls":"Warrior"`, `"specs":[{"key":"Sword"`, `"ab":[{"nm":"Evade","lv":1,"on":true}]`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("stats lacks %s: %s", want, b)
		}
	}
	b, _ = json.Marshal(Healer{T: THealer, NPC: 3, Lost: 2, Price: 352, Total: 704})
	if !strings.Contains(string(b), `"total":704`) {
		t.Errorf("healer: %s", b)
	}
}

func TestEffectsMessage(t *testing.T) {
	b, _ := json.Marshal(Effects{T: TEffects, List: []EffectView{{Name: "Shroud of Doubt", Kind: "slow", Icon: 10082, Left: 4.5, Of: 5}}, Speed: 0.5})
	for _, want := range []string{`"t":"effects"`, `"list":[{"nm":"Shroud of Doubt","k":"slow","ic":10082,"left":4.5,"of":5}]`, `"speed":0.5`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("effects lacks %s: %s", want, b)
		}
	}
	if strings.Contains(string(b), "held") {
		t.Errorf("held is sent only while held: %s", b)
	}
	b, _ = json.Marshal(Event{K: "spell", From: 1, To: 2, Text: "Spirit Bolt", Amount: 24})
	if string(b) != `{"k":"spell","from":1,"to":2,"n":24,"x":"Spirit Bolt"}` {
		t.Errorf("spell event: %s", b)
	}
}

func TestCastEvent(t *testing.T) {
	b, _ := json.Marshal(Event{K: "cast", From: 5, To: 7, Text: "Simple Runebolt", Dur: 2.5})
	if string(b) != `{"k":"cast","from":5,"to":7,"x":"Simple Runebolt","dur":2.5}` {
		t.Errorf("cast event: %s", b)
	}
	b, _ = json.Marshal(Event{K: "hit", From: 1, To: 2, Amount: 3})
	if strings.Contains(string(b), "dur") {
		t.Errorf("dur appears on other events: %s", b)
	}
}
