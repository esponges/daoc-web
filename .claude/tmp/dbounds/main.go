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
	for _, a := range os.Args[1:] {
		n, _ := strconv.Atoi(a)
		l, err := interior.Read(game, n)
		if err != nil {
			fmt.Println(n, err)
			continue
		}
		dirs := interior.ModelDirs(game, l)
		type model struct {
			min, max [3]float32
			verts    [][3]float32
		}
		models := map[string]*model{}
		var boxes []interior.Box
		var verts [][3]float32
		for _, p := range l.Pieces {
			m, ok := models[p.Model]
			if !ok {
				var v []scenery.Vertex
				var ix []uint32
				mo, err := scenery.ConvertModel(dirs, p.Model, p.Model, &v, &ix)
				if err != nil {
					models[p.Model] = nil
					continue
				}
				m = &model{min: mo.Min, max: mo.Max}
				for _, x := range v {
					m.verts = append(m.verts, x.P)
				}
				models[p.Model] = m
			}
			if m == nil {
				continue
			}
			boxes = append(boxes, interior.Place(p, m.min, m.max))
			for _, v := range m.verts {
				verts = append(verts, p.Point(v))
			}
		}
		surf := interior.NewSurface(verts)
		_ = boxes
		fit := interior.FitInterior(l, surf)
		fmt.Println("zone", n, l.Kind, "origin", fit.Origin, "on surface", fit.OnSurface, "of", fit.Points, fit.Score())
	}
}
