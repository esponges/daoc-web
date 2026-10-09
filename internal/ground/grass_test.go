package ground

import (
	"errors"
	"testing"
)

// Gripklosa Mountains draws its grass maps at 512 against a 256 heightmap:
// cells come out at the maps' resolution.
func TestGrassAtTwiceTheGrid(t *testing.T) {
	game := gamePath(t)
	gr, err := ReadGrass(game, open(t, game, 152), 256, quiet)
	if err != nil {
		t.Fatal(err)
	}
	if gr.CellsPerSide != 512 || len(gr.Cells) != 512*512*2 {
		t.Fatalf("%d cells a side, %d bytes; want 512 and %d", gr.CellsPerSide, len(gr.Cells), 512*512*2)
	}
	groups := map[int]bool{}
	for _, s := range gr.Sprites {
		groups[s.Group] = true
	}
	grown := 0
	for i := 0; i < len(gr.Cells); i += 2 {
		if gr.Cells[i+1] > 0 && groups[int(gr.Cells[i])] {
			grown++
		}
	}
	if grown < 512*512/10 {
		t.Errorf("grass grows on %d of %d cells", grown, 512*512)
	}
}

// Uppland's data has no grass map: no grass, and not a failure.
func TestNoGrassMap(t *testing.T) {
	game := gamePath(t)
	_, err := ReadGrass(game, open(t, game, 111), 256, quiet)
	if !errors.Is(err, ErrNoGrassMap) {
		t.Fatalf("got %v, want ErrNoGrassMap", err)
	}
}
