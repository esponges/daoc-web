// Command propconv converts a zone's scenery -- the trees, fences, buildings
// and other props that fixtures.csv places -- into a single vertex buffer plus
// an instance list the viewer can draw.
//
// Reading the zone's CSVs is internal/zone's; baking each model is
// internal/scenery's. This command picks the models the zone places, bakes
// each once, and writes the shared buffers, the instance list and the
// textures.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"

	"daocweb/internal/scenery"
	"daocweb/internal/zone"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

type instanceOut struct {
	Model int        `json:"m"`
	Pos   [3]float32 `json:"p"`
	Yaw   float32    `json:"yaw"`
	Scale float32    `json:"s"`
}

type manifest struct {
	Zone        int             `json:"zone"`
	Stride      int             `json:"stride"`
	VertexCount int             `json:"vertexCount"`
	IndexCount  int             `json:"indexCount"`
	VertexBytes int             `json:"vertexBytes"`
	IndexBytes  int             `json:"indexBytes"`
	UpAxis      string          `json:"upAxis"`
	Models      []scenery.Model `json:"models"`
	Instances   []instanceOut   `json:"instances"`
	Skipped     []string        `json:"skipped"`
}

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	zone := flag.Int("zone", 100, "zone number")
	out := flag.String("out", "web/data", "output directory")
	listOnly := flag.Bool("list", false, "list the models a zone places and stop")
	flag.Parse()

	if err := run(*game, *zone, *out, *listOnly); err != nil {
		fmt.Fprintln(os.Stderr, "propconv:", err)
		os.Exit(1)
	}
}

func run(game string, zoneNum int, out string, listOnly bool) error {
	z, err := zone.Open(game, zoneNum)
	if err != nil {
		return err
	}
	models, err := z.Models()
	if err != nil {
		return err
	}
	fixtures, err := z.PlacedFixtures()
	if err != nil {
		return err
	}
	fmt.Printf("zone %d: %d model definitions, %d fixtures\n", zoneNum, len(models), len(fixtures))

	// Only convert what the zone actually places.
	used := map[int]int{}
	for _, f := range fixtures {
		used[f.NifID]++
	}
	ids := make([]int, 0, len(used))
	for id := range used {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	if listOnly {
		for _, id := range ids {
			m, ok := models[id]
			name := "<not in nifs.csv>"
			if ok {
				name = fmt.Sprintf("%-28s %s", m.Name, m.File)
			}
			fmt.Printf("  %4d  %5d placements  %s\n", id, used[id], name)
		}
		return nil
	}

	dir := filepath.Join(out, fmt.Sprintf("zone%03d", zoneNum), "props")
	if err := os.MkdirAll(filepath.Join(dir, "tex"), 0o755); err != nil {
		return err
	}

	dirs := scenery.ModelDirs(game, z.Frontier)
	var (
		verts    []scenery.Vertex
		indices  []uint32
		outModel []scenery.Model
		skipped  []string
		slot     = map[int]int{} // nif id -> index into outModel
		wantTex  = map[string]bool{}
		embedded = map[string]image.Image{} // textures carried inside models
	)
	for _, id := range ids {
		def, ok := models[id]
		if !ok {
			skipped = append(skipped, fmt.Sprintf("%d: no entry in nifs.csv (%d placements)", id, used[id]))
			continue
		}
		mo, err := scenery.ConvertModel(dirs, def.Name, def.File, &verts, &indices)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v (%d placements)", def.File, err, used[id]))
			continue
		}
		for _, g := range mo.Groups {
			if g.Texture != "" {
				wantTex[g.Texture] = true
			}
		}
		for _, p := range mo.Particles {
			wantTex[p.Texture] = true
		}
		for n, img := range mo.Embedded {
			embedded[n] = img
		}
		slot[id] = len(outModel)
		outModel = append(outModel, *mo)
	}

	var inst []instanceOut
	placed := 0
	for _, f := range fixtures {
		s, ok := slot[f.NifID]
		if !ok {
			continue
		}
		inst = append(inst, instanceOut{
			Model: s,
			Pos:   [3]float32{f.X, f.Y, f.Z},
			Yaw:   f.Yaw,
			Scale: f.Scale,
		})
		placed++
	}

	// Pack both arrays into one file so the viewer makes a single request.
	buf := &bytes.Buffer{}
	for _, v := range verts {
		writeF32(buf, v.P[0], v.P[1], v.P[2], v.N[0], v.N[1], v.N[2], v.UV[0], v.UV[1])
		buf.Write(v.C[:])
	}
	vertexBytes := buf.Len()
	for _, i := range indices {
		binary.Write(buf, binary.LittleEndian, i)
	}
	if err := os.WriteFile(filepath.Join(dir, "props.bin"), buf.Bytes(), 0o644); err != nil {
		return err
	}

	man := manifest{
		Zone:        zoneNum,
		Stride:      scenery.Stride,
		VertexCount: len(verts),
		IndexCount:  len(indices),
		VertexBytes: vertexBytes,
		IndexBytes:  buf.Len() - vertexBytes,
		UpAxis:      "Z",
		Models:      outModel,
		Instances:   inst,
		Skipped:     skipped,
	}
	js, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "props.json"), js, 0o644); err != nil {
		return err
	}

	// Embedded textures are written from the models; the rest are looked
	// up beside them.
	external := map[string]bool{}
	for n := range wantTex {
		if img, ok := embedded[n]; ok {
			if err := scenery.WritePNG(filepath.Join(dir, "tex", n+".png"), img); err != nil {
				return err
			}
			continue
		}
		external[n] = true
	}
	if len(embedded) > 0 {
		fmt.Printf("  %d embedded textures written\n", len(embedded))
	}
	if err := scenery.ExtractTextures(dirs, filepath.Join(dir, "tex"), external); err != nil {
		return err
	}

	fmt.Printf("\n%d models converted, %d skipped\n", len(outModel), len(skipped))
	for _, s := range skipped {
		fmt.Printf("  skipped %s\n", s)
	}
	fmt.Printf("%d of %d fixtures placed (%.1f%%)\n",
		placed, len(fixtures), 100*float64(placed)/float64(len(fixtures)))
	fmt.Printf("%d vertices, %d indices, %d textures -> %s\n",
		len(verts), len(indices), len(wantTex), dir)
	return nil
}

func writeF32(w io.Writer, vals ...float32) {
	var b [4]byte
	for _, v := range vals {
		binary.LittleEndian.PutUint32(b[:], math.Float32bits(v))
		w.Write(b[:])
	}
}
