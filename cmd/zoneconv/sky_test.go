package main

import (
	"os"
	"testing"
)

// TestMularnSky follows zone 100 to its region's skydome and checks what
// the Midgard sky file says reaches the manifest.
func TestMularnSky(t *testing.T) {
	game := os.Getenv("DAOC_PATH")
	if game == "" {
		game = defaultGamePath
	}
	if _, err := os.Stat(game); err != nil {
		t.Skipf("no DAoC install at %s (set DAOC_PATH to run); %v", game, err)
	}
	sk, err := buildSky(game, 100, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if sk.Source != "sky_midgard.dat" {
		t.Errorf("zone 100's sky is %q, want sky_midgard.dat (region 100)", sk.Source)
	}
	if sk.Zenith != [3]int{115, 159, 229} || sk.East[3] != [3]int{229, 229, 229} {
		t.Errorf("day canopy zenith %v, east horizon %v", sk.Zenith, sk.East[3])
	}
	if sk.Fog != [3]int{184, 201, 229} {
		t.Errorf("day fog %v, want 184,201,229", sk.Fog)
	}
	if len(sk.Clouds) != 2 || sk.Clouds[0].Tile != 4 || sk.Clouds[1].Speed != 12 {
		t.Errorf("cloud layers %+v", sk.Clouds)
	}
	// Clouds thin to nothing at the horizon.
	if sk.CloudEast[3][3] != 0 || sk.CloudWest[3][3] != 0 {
		t.Errorf("clouds at the horizon have opacity %d/%d, want 0", sk.CloudEast[3][3], sk.CloudWest[3][3])
	}
	if sk.SunDisc == nil {
		t.Error("no sun")
	}
}
