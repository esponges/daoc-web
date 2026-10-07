package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
