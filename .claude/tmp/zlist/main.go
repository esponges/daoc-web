package main

import (
	"fmt"
	"os"
	"strconv"

	"daocweb/internal/interior"
	"daocweb/internal/scenery"
)

func main() {
	game := `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`
	n, _ := strconv.Atoi(os.Args[1])
	l, _ := interior.Read(game, n)
	dirs := interior.ModelDirs(game, l)
	bad := map[string]int{}
	for _, p := range append(append([]interior.Placement{}, l.Pieces...), l.Props...) {
		var v []scenery.Vertex
		var ix []uint32
		if _, err := scenery.ConvertModel(dirs, p.Model, p.Model, &v, &ix); err != nil {
			bad[p.Model+": "+err.Error()]++
		}
	}
	fmt.Println(len(l.Pieces), "pieces", len(l.Props), "props", l.Skipped, "skipped; failing:", bad)
	zmin, zmax := float32(1e9), float32(-1e9)
	for _, p := range l.Pieces {
		zmin, zmax = min(zmin, p.Pos[2]), max(zmax, p.Pos[2])
		if p.Pos[2] < -20000 {
			fmt.Println("deep piece", p.Model, p.Pos)
		}
	}
	fmt.Println("z", zmin, zmax)
}
