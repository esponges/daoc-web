package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"daocweb/internal/gamedata"
)

func gamePath(t *testing.T) string {
	t.Helper()
	p := os.Getenv("DAOC_PATH")
	if p == "" {
		p = defaultGame
	}
	if _, err := os.Stat(p); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", p, err)
	}
	return p
}

// animnode.dat is the whole reason the animations are portable: a track names
// a bone by index into this list, not by name, so every humanoid model that
// uses Biped names can play every humanoid animation.
func TestAnimNodeRegistry(t *testing.T) {
	nodes, err := loadAnimNodes(gamePath(t))
	if err != nil {
		t.Fatalf("loadAnimNodes: %v", err)
	}
	if len(nodes) != 178 {
		t.Errorf("animnode.dat has %d entries, want 178", len(nodes))
	}
	// The indices this project relies on, spot-checked against the file.
	for idx, want := range map[int]string{
		0: "Bip01", 1: "Bip01 Pelvis", 2: "Bip01 Spine", 11: "Bip01 Head",
	} {
		if idx >= len(nodes) || nodes[idx] != want {
			got := "<out of range>"
			if idx < len(nodes) {
				got = nodes[idx]
			}
			t.Errorf("bone index %d is %q, want %q", idx, got, want)
		}
	}
	biped := 0
	for _, n := range nodes {
		if strings.HasPrefix(n, "Bip01") {
			biped++
		}
	}
	if biped < 100 {
		t.Errorf("only %d of %d registry entries are Biped names", biped, len(nodes))
	}
	t.Logf("%d bone slots, %d of them Biped", len(nodes), biped)
}

// The Norseman's locomotion set. The model is NVikingM.NIF and the animations
// are W_VM / R_VM / I_VM -- Viking Male -- which is the naming convention the
// scan turned up rather than an assumption.
func TestVikingLocomotionResolves(t *testing.T) {
	game := gamePath(t)
	nodes, err := loadAnimNodes(game)
	if err != nil {
		t.Fatalf("loadAnimNodes: %v", err)
	}
	skel, err := loadCharBones(filepath.Join("..", "..", "web", "data", "char", "norseman"))
	if err != nil {
		t.Skipf("character not converted; run charconv first (%v)", err)
	}

	for _, name := range []string{"W_VM", "R_VM", "I_VM"} {
		path, err := findCase(filepath.Join(game, "anims"), name+".kfa")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		a, err := readAnim(path, nodes)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(a.Tracks) != 61 {
			t.Errorf("%s has %d tracks, want 61", name, len(a.Tracks))
		}
		// Every track must land on a bone this model actually has. A track
		// that does not would mean the index lookup is wrong.
		unmapped := 0
		for _, tr := range a.Tracks {
			if !skel[strings.ToLower(tr.Bone)] {
				unmapped++
			}
		}
		if unmapped != 0 {
			t.Errorf("%s: %d of %d tracks name bones the Norseman does not have",
				name, unmapped, len(a.Tracks))
		}
		if a.Duration <= 0 || a.Duration > 10 {
			t.Errorf("%s lasts %gs, which is not a plausible cycle", name, a.Duration)
		}

		// A locomotion clip has to drive the legs, and with more than one key
		// -- a single key is a static pose.
		driven := map[string]int{}
		for _, tr := range a.Tracks {
			driven[tr.Bone] = len(tr.Rotations)
		}
		for _, bone := range gaitBones {
			if driven[bone] < 2 {
				t.Errorf("%s drives %s with %d rotation keys; want a real cycle",
					name, bone, driven[bone])
			}
		}
		t.Logf("%-5s %5.2fs, %d tracks, all on the model", name, a.Duration, len(a.Tracks))
	}
}

// The idle clip should barely move the legs while the walk swings them. This
// separates "we read the file" from "we read the right file".
func TestIdleIsStillerThanWalk(t *testing.T) {
	game := gamePath(t)
	nodes, err := loadAnimNodes(game)
	if err != nil {
		t.Fatalf("loadAnimNodes: %v", err)
	}
	spread := func(name string) float32 {
		path, err := findCase(filepath.Join(game, "anims"), name+".kfa")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		a, err := readAnim(path, nodes)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		// Peak-to-peak of the thigh quaternion's x component is a cheap proxy
		// for how far the leg swings; the forward kinematics are checked on
		// the JS side, where the skeleton lives.
		var lo, hi float32 = 1, -1
		for _, tr := range a.Tracks {
			if !strings.EqualFold(tr.Bone, "Bip01 L Thigh") {
				continue
			}
			for _, k := range tr.Rotations {
				if k.Value[1] < lo {
					lo = k.Value[1]
				}
				if k.Value[1] > hi {
					hi = k.Value[1]
				}
			}
		}
		return hi - lo
	}
	idle, walk := spread("I_VM"), spread("W_VM")
	if !(walk > idle*2) {
		t.Errorf("walk thigh swing %.4f is not clearly larger than idle's %.4f", walk, idle)
	}
	t.Logf("L thigh rotation spread: idle %.4f, walk %.4f", idle, walk)
}

// A spell's caster animations are numbered columns of anims.csv: Minor Dusk
// Dart's 41, 42 and 43 become cast41, cast42 and cast43 for a Norse Male, each
// a different clip, and a set that has none of a number simply lacks it.
func TestCastJobsFollowTheSpellTable(t *testing.T) {
	game := gamePath(t)
	tb, err := gamedataLoad(game)
	if err != nil {
		t.Fatal(err)
	}
	jobs := castJobs(tb, 28) // anim set 28: Norse Male
	byOut := map[string]job{}
	for _, j := range jobs {
		byOut[j.out] = j
		if !j.alias {
			t.Errorf("%s is not an alias job; a failed cast clip must not stop the run", j.out)
		}
	}
	files := map[string]bool{}
	for _, n := range []string{"cast41", "cast42", "cast43"} {
		j, ok := byOut[n]
		if !ok {
			t.Fatalf("no %s job for Norse Male", n)
		}
		files[strings.ToLower(j.src)] = true
	}
	if len(files) != 3 {
		t.Errorf("Bk-Hand start, mid and end are %v, want three different clips", files)
	}
	if len(castNumbers(tb)) < 40 {
		t.Errorf("only %d distinct caster animations in the spell table", len(castNumbers(tb)))
	}
}

// A style's animation is a numbered column of canims.csv, named by its row of
// styles.csv: Assault (style 150) animates as 60, "Style 1HS 4", and the
// one-handed crush styles 3 and 4 as 42 and 43, "Style 1HC 1" and "2".
func TestStyleJobsFollowTheStyleTable(t *testing.T) {
	game := gamePath(t)
	tb, err := gamedataLoad(game)
	if err != nil {
		t.Fatal(err)
	}
	for id, want := range map[int]int{150: 60, 3: 42, 4: 43} {
		if got := tb.StyleAnims[id]; got != want {
			t.Errorf("style %d animates as %d, want %d", id, got, want)
		}
	}
	jobs := styleJobs(tb, 28) // anim set 28: Norse Male
	byOut := map[string]job{}
	for _, j := range jobs {
		byOut[j.out] = j
		if !j.alias {
			t.Errorf("%s is not an alias job", j.out)
		}
	}
	for _, n := range []string{"style42", "style43", "style60"} {
		if _, ok := byOut[n]; !ok {
			t.Errorf("no %s job for Norse Male", n)
		}
	}
	if byOut["style42"].src == byOut["style43"].src {
		t.Errorf("Style 1HC 1 and 2 are the same clip %q", byOut["style42"].src)
	}
}

// A player body gets a swing set for every kind of weapon: a Norse Male (anim
// set 28) swings a two-hander with different clips from a one-hander, and both
// differ from his fists.
func TestWieldJobsGiveEachWeaponItsSwings(t *testing.T) {
	game := gamePath(t)
	tb, err := gamedataLoad(game)
	if err != nil {
		t.Fatal(err)
	}
	src := map[string]string{}
	for _, j := range wieldJobs(tb, 28) {
		src[j.out] = strings.ToLower(j.src)
		if !j.alias {
			t.Errorf("%s is not an alias job", j.out)
		}
	}
	for _, k := range []string{"1h", "2h", "pole", "staff", "dual", "h2h"} {
		if src["att_"+k+"1"] == "" {
			t.Errorf("no att_%s1 for Norse Male", k)
		}
	}
	if src["att_1h1"] == src["att_2h1"] || src["att_1h1"] == src["att_h2h1"] {
		t.Errorf("one-handed %q, two-handed %q and fists %q should differ", src["att_1h1"], src["att_2h1"], src["att_h2h1"])
	}
	if src["cidle_2h"] == "" {
		t.Error("no two-handed combat stance")
	}
}

func TestWithEntryReplacesBySameName(t *testing.T) {
	l := withEntry(nil, indexEntry{Name: "walk", File: "walk.json"})
	l = withEntry(l, indexEntry{Name: "cast41", File: "cast41.json"})
	l = withEntry(l, indexEntry{Name: "walk", File: "walk2.json"})
	if len(l) != 2 || l[0].File != "walk2.json" {
		t.Errorf("entries %+v, want walk replaced in place and cast41 kept", l)
	}
}

// Cast clips are written lazy, and two numbers that name the same file share it.
func TestCastClipsAreLazyAndShareFiles(t *testing.T) {
	game := gamePath(t)
	src := filepath.Join("..", "..", "web", "data", "char", "norseman", "char.json")
	b, err := os.ReadFile(src)
	if err != nil {
		t.Skipf("character not converted; run charconv first (%v)", err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "char.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	nodes, err := loadAnimNodes(game)
	if err != nil {
		t.Fatal(err)
	}
	skel, err := loadCharBones(dir)
	if err != nil {
		t.Fatal(err)
	}
	jobs := []job{
		{src: "W_VM", out: "cast1", rate: 1, alias: true},
		{src: "W_VM", out: "cast2", rate: 1, alias: true},
		{src: "no_such_clip", out: "cast3", rate: 1, alias: true},
	}
	if err := convert(game, dir, nodes, skel, jobs); err != nil {
		t.Fatalf("a missing cast clip must not stop the run: %v", err)
	}
	ents, _ := os.ReadDir(filepath.Join(dir, "anim"))
	if len(ents) != 2 { // cast1.json and index.json
		t.Errorf("%d files written, want the clip once and the index", len(ents))
	}
	idx, _ := os.ReadFile(filepath.Join(dir, "anim", "index.json"))
	if n := strings.Count(string(idx), `"lazy": true`); n != 2 {
		t.Errorf("%d lazy entries in %s, want 2", n, idx)
	}
	if strings.Contains(string(idx), "cast3") {
		t.Error("the clip that does not exist was listed")
	}
}

func TestPlayerFigureIsNotACreatureModel(t *testing.T) {
	data := t.TempDir()
	os.WriteFile(filepath.Join(data, "races.json"), []byte(`{"races":{"6":{"m":"m137","f":"m145"}}}`), 0o644)
	for dir, want := range map[string]bool{
		"norseman":      true,
		"troll-warrior": true,
		"mage":          true,
		"m137":          true, // a race base
		"m145":          true,
		"m100":          false, // a creature
		"m1051":         false, // not in this races.json
	} {
		if got := playerFigure(filepath.Join(data, "char", dir)); got != want {
			t.Errorf("playerFigure(%q) = %v, want %v", dir, got, want)
		}
	}
}

// Every player role names a column the tables have, and an emote has its own
// clip name.
func TestPlayerRolesNameRealColumns(t *testing.T) {
	game := gamePath(t)
	tb, err := gamedata.Load(game)
	if err != nil {
		t.Skipf("the game tables are not readable: %v", err)
	}
	set := 24 // the male troll's anim set, a player race
	for _, r := range playerRoles {
		if !r.player {
			t.Errorf("%s is a player role not marked so", r.out)
		}
		if _, ok := tb.SetClip(set, r.column); !ok {
			t.Errorf("anim set %d has no %q clip for %s", set, r.column, r.out)
		}
	}
}
