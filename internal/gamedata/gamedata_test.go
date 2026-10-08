package gamedata_test

import (
	"math"
	"os"
	"strings"
	"testing"

	"daocweb/internal/gamedata"
)

func load(t *testing.T) *gamedata.Tables {
	t.Helper()
	p := os.Getenv("DAOC_PATH")
	if p == "" {
		p = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`
	}
	if _, err := os.Stat(p); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", p, err)
	}
	tb, err := gamedata.Load(p)
	if err != nil {
		t.Fatal(err)
	}
	return tb
}

// The grey wolf resolves through every table to files that exist.
func TestGreyWolf(t *testing.T) {
	tb := load(t)
	m, ok := tb.Monsters[56]
	if !ok || m.Name != "Large Grey Wolf" {
		t.Fatalf("model 56 = %+v", m)
	}
	fig := tb.Figures[m.Figure]
	if fig.File != "wolf" || fig.AnimSet != 7 {
		t.Errorf("figure %+v, want wolf with anim set 7", fig)
	}
	if s := tb.Skins[m.Skins["Body"]]; s.DDS() != "wolf.dds" || s.Archive != 2 {
		t.Errorf("body skin %+v, want wolf.dds in skin002", s)
	}
	want := map[string]string{
		"walk": "wlf_walk.kfa", "run": "wlf_gllp.kfa",
		"idle": "wlf_idle.kfa", "death": "wlf_deth.kfa",
	}
	for role, file := range want {
		if an := mustClip(t, tb, fig, role); an.File != file {
			t.Errorf("%s = %s, want %s", role, an.File, file)
		}
	}
	if mustClip(t, tb, fig, "death").Loop {
		t.Error("death should play once, not loop")
	}
}

// The table is where the size difference between races lives: the troll is
// drawn 30% larger than authored, which no model file records.
func TestTrollScale(t *testing.T) {
	tb := load(t)
	m := tb.Monsters[137]
	if m.Name != "Troll Male" || m.Scale != 1.30 {
		t.Errorf("model 137 = %q scale %v, want Troll Male at 1.30", m.Name, m.Scale)
	}
	fig := tb.Figures[m.Figure]
	if fig.File != "ntrollm" {
		t.Errorf("figure %q, want ntrollm", fig.File)
	}
	// Its anim set names the same clips that were chosen by hand, by
	// scoring, before these tables were read.
	want := map[string]string{"walk": "troll_walk.kfa", "run": "troll_run.kfa", "idle": "troll_idle.kfa"}
	for role, file := range want {
		if an := mustClip(t, tb, fig, role); an.File != file {
			t.Errorf("%s = %s, want %s", role, an.File, file)
		}
	}
}

func mustClip(t *testing.T, tb *gamedata.Tables, fig gamedata.Figure, role string) gamedata.Anim {
	t.Helper()
	an, ok := tb.Clip(fig, role)
	if !ok {
		t.Fatalf("%s has no %s clip", fig.File, role)
	}
	return an
}

// Combat clips come from canims.csv, keyed by the same anim set. A creature's
// row names its own attacks; a player race's carries the humanoid ones.
func TestCombatClips(t *testing.T) {
	tb := load(t)
	want := []struct {
		set        int
		role, file string
		loop       bool
	}{
		{7, "att med", "wlf_alo.kfa", false},
		{7, "flinch", "wlf_hits.kfa", false},
		{7, "c-idle", "wlf_grwl.kfa", true},
		{54, "att med", "rat_atthi.kfa", false},
		{24, "h2h att m", "A_H_H2H_Med.kfa", false},
		{24, "c-idle", "ci_h_1s.kfa", true},
	}
	for _, w := range want {
		an, ok := tb.CombatClip(w.set, w.role)
		if !ok || !strings.EqualFold(an.File, w.file) || an.Loop != w.loop {
			t.Errorf("set %d %s = %+v, want %s loop=%v", w.set, w.role, an, w.file, w.loop)
		}
	}
	if fig := tb.Figures[237]; fig.Type != 1 {
		t.Errorf("ntrollm type %d, want 1 (player race)", fig.Type)
	}
	if fig := tb.Figures[212]; fig.Type != 0 {
		t.Errorf("wolf type %d, want 0 (creature)", fig.Type)
	}
}

// Idles are played slower than they were keyed; everything else at speed.
func TestPlaybackRate(t *testing.T) {
	tb := load(t)
	for file, want := range map[string]float64{
		"troll_idle.kfa": 2.0 / 15, "I_hm.kfa": 4.0 / 15, "wlf_idle.kfa": 12.0 / 15,
		"troll_walk.kfa": 1, "a_h_1s_med.kfa": 1,
	} {
		an, ok := tb.ByFile(file)
		if !ok || math.Abs(an.Rate()-want) > 1e-9 {
			t.Errorf("%s rate %v, want %v", file, an.Rate(), want)
		}
	}
}
