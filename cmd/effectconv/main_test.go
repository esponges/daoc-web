package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"daocweb/internal/gamedata"
)

// Plan follows a client effect to the effects it plays: Minor Dusk Dart's row
// names two caster effects and a target effect, and the plan lists exactly those.
func TestPlanFollowsASpellToItsEffects(t *testing.T) {
	tb := &gamedata.Tables{
		Spells: map[int]gamedata.SpellFX{
			2501: {ID: 2501, Name: "Minor Dusk Dart", Anim: [3]int{41, 42, 43}, Caster: []int{160, 161}, Target: 256},
			9:    {ID: 9, Name: "Bolt", Anim: [3]int{32, 33, 34}, Projectile: 77, Resist: 5},
		},
		Effects: map[int]gamedata.Effect{160: {ID: 160}, 161: {ID: 161}, 256: {ID: 256}, 77: {ID: 77},
			592: {ID: 592, Name: "arrow slash01 load"}, 593: {ID: 593, Name: "arrow slash01 travel"}, 594: {ID: 594, Name: "arrow slash01 strike"}},
	}
	idx, need := Plan(tb, []int{2501, 9, 4040})
	if got := idx.Spells["2501"]; !reflect.DeepEqual(got, SpellOut{Name: "Minor Dusk Dart", Anims: [3]int{41, 42, 43}, Caster: []int{160, 161}, Target: 256}) {
		t.Errorf("2501 plans as %+v", got)
	}
	if idx.Arrow == nil || idx.Arrow.Travel != 593 || idx.Arrow.Strike != 594 {
		t.Errorf("the bow shot plans as %+v, want the slashing arrow's travel 593 and strike 594", idx.Arrow)
	}
	if len(idx.Spells) != 2 {
		t.Errorf("%d spells planned, want 2: a client effect the table lacks is left out", len(idx.Spells))
	}
	want := map[int]bool{160: true, 161: true, 256: true, 77: true, 593: true, 594: true}
	if !reflect.DeepEqual(need, want) {
		t.Errorf("effects needed %v, want %v (effect 5 is not in the effects table)", need, want)
	}
}

func TestUndrawnTypesCountsTheAnimationsNotDrawn(t *testing.T) {
	got := UndrawnTypes([]string{"NiNode", "NiPathController", "NiTriShape", "NiPathController", "NiAlphaController"})
	want := []string{"NiAlphaController x1", "NiPathController x2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if UndrawnTypes([]string{"NiNode"}) != nil {
		t.Error("a plain model has nothing undrawn")
	}
}

// On the install, the effects Minor Dusk Dart plays resolve to models that exist.
func TestMinorDuskDartsEffectsHaveModels(t *testing.T) {
	game := os.Getenv("DAOC_PATH")
	if game == "" {
		game = defaultGame
	}
	if _, err := os.Stat(game); err != nil {
		t.Skipf("no DAoC install at %s", game)
	}
	tb, err := gamedata.Load(game)
	if err != nil {
		t.Fatal(err)
	}
	_, need := Plan(tb, []int{2501})
	if !need[160] || !need[161] || !need[256] {
		t.Fatalf("effects needed %v, want 160, 161 and 256", need)
	}
	for id := range need {
		e := tb.Effects[id]
		file, ok := tb.EffectNIFs[e.NIF]
		if !ok || e.Duration <= 0 || e.Node == "" {
			t.Errorf("effect %d %+v has no model file, duration or bone", id, e)
			continue
		}
		if onDisk(filepath.Join(game, "effects"), file) == "" {
			t.Errorf("effect %d: model %q is not in the effects directory", id, file)
		}
	}
}
