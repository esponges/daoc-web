package main

// Ground detail: the layers the game paints the terrain with up close. What
// they are and how they are checked against the game is in internal/ground;
// this writes them out for the viewer.

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"daocweb/internal/ground"
	"daocweb/internal/zone"
)

type splatOut struct {
	Layers  string   `json:"layers"`  // JPEG, layerPx wide, one layer under another
	LayerPx int      `json:"layerPx"` // side of one layer
	Names   []string `json:"names"`   // layer names, in strip order
	Masks   string   `json:"masks"`   // PNG, one plane under another
	MaskPx  int      `json:"maskPx"`  // mask texels per sector side
	Planes  int      `json:"planes"`  // mask tiles per sector
	Sectors int      `json:"sectors"` // sectors per zone side
	// Slots[sy*Sectors+sx] lists the sector's layers in paint order, each
	// as [layer index, repeats across the zone].
	Slots  [][][2]float64 `json:"slots"`
	Layout string         `json:"layout"` // how mask tiles were found to sit
	Agree  float64        `json:"agree"`  // correlation with the game's own tiles
	// Blend is "weighted" where the layers are averaged by their mask
	// weights; left out where each is painted over the ones before.
	Blend string `json:"blend,omitempty"`
}

func buildSplat(game string, zoneNum, sectors int, outDir string) (*splatOut, error) {
	z, err := zone.Open(game, zoneNum)
	if err != nil {
		return nil, err
	}
	sp, err := ground.ReadSplat(game, z, sectors, logf)
	if err != nil {
		return nil, err
	}
	if err := writeJPEG(filepath.Join(outDir, "layers.jpg"), sp.Strip); err != nil {
		return nil, err
	}
	if err := writePNG(filepath.Join(outDir, "masks.png"), sp.Masks); err != nil {
		return nil, err
	}
	return &splatOut{
		Layers: "layers.jpg", LayerPx: ground.LayerPx, Names: sp.Names,
		Masks: "masks.png", MaskPx: ground.MaskPx, Planes: sp.Planes, Sectors: sp.Sectors,
		Slots: sp.Slots, Layout: sp.Layout, Agree: sp.Agree,
		Blend: blend(sp.Weighted),
	}, nil
}

func blend(weighted bool) string {
	if weighted {
		return "weighted"
	}
	return ""
}

// logf is how the shared readers report to the console.
func logf(format string, a ...any) { fmt.Printf(format, a...) }

func writeJPEG(path string, im image.Image) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	return jpeg.Encode(fh, im, &jpeg.Options{Quality: 90})
}
