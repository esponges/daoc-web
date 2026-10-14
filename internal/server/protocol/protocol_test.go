package protocol

import (
	"encoding/json"
	"reflect"
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
