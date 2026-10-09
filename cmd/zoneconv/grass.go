package main

// Grass: the small plants the game scatters over the ground near the camera.
// internal/ground reads the region's sprite table and the zone's grass
// maps; this writes them out for the viewer.

import (
	"fmt"
	"os"
	"path/filepath"

	"daocweb/internal/ground"
	"daocweb/internal/zone"
)

type grassOut struct {
	Table   string `json:"table"`
	Atlas   string `json:"atlas"`
	AtlasPx int    `json:"atlasPx"`
	Map     string `json:"map"` // per cell: group, then density, one byte each
	// CellsPerSide is how many map cells span the zone, where it is not the
	// heightmap's grid.
	CellsPerSide int             `json:"cellsPerSide,omitempty"`
	Sprites      []ground.Sprite `json:"sprites"`
}

func buildGrass(game string, zoneNum, grid int, outDir string) (*grassOut, error) {
	z, err := zone.Open(game, zoneNum)
	if err != nil {
		return nil, err
	}
	gr, err := ground.ReadGrass(game, z, grid, logf)
	if err != nil {
		return nil, err
	}
	out := &grassOut{Table: gr.Table, Atlas: "grass/atlas.png", Map: "grass/map.bin", Sprites: gr.Sprites}
	if gr.CellsPerSide != grid {
		out.CellsPerSide = gr.CellsPerSide
	}
	if err := os.MkdirAll(filepath.Join(outDir, "grass"), 0o755); err != nil {
		return nil, err
	}
	png, err := convertSkyTexture(gr.AtlasDir, gr.Atlas, filepath.Join(outDir, "grass"))
	if err != nil {
		return nil, err
	}
	if err := os.Rename(filepath.Join(outDir, "grass", png), filepath.Join(outDir, "grass", "atlas.png")); err != nil {
		return nil, err
	}
	out.AtlasPx = 512
	if err := os.WriteFile(filepath.Join(outDir, "grass", "map.bin"), gr.Cells, 0o644); err != nil {
		return nil, err
	}
	fmt.Printf("  grass: %s, %d sprites in %d groups, atlas %s\n", gr.Table, len(gr.Sprites), gr.Groups, gr.Atlas)
	return out, nil
}
