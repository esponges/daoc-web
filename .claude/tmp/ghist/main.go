package main

import (
	"fmt"
	"math"
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
	var boxes []interior.Box
	cache := map[string][2][3]float32{}
	box := func(p interior.Placement) (interior.Box, bool) {
		b, ok := cache[p.Model]
		if !ok {
			var v []scenery.Vertex
			var ix []uint32
			mo, err := scenery.ConvertModel(dirs, p.Model, p.Model, &v, &ix)
			if err != nil {
				return interior.Box{}, false
			}
			b = [2][3]float32{mo.Min, mo.Max}
			cache[p.Model] = b
		}
		return interior.Place(p, b[0], b[1]), true
	}
	for _, p := range l.Pieces {
		if b, ok := box(p); ok {
			boxes = append(boxes, b)
		}
	}
	t := interior.Transform{Origin: [3]float32{23808, 23808, 16128}, Turn: 2, MirrorY: true}
	dist := func(q [3]float32) (float64, int) {
		best, bi := math.Inf(1), -1
		for i, b := range boxes {
			d := 0.0
			for k := 0; k < 3; k++ {
				if q[k] < b.Min[k] {
					d = math.Max(d, float64(b.Min[k]-q[k]))
				} else if q[k] > b.Max[k] {
					d = math.Max(d, float64(q[k]-b.Max[k]))
				}
			}
			if d < best {
				best, bi = d, i
			}
		}
		return best, bi
	}
	hist := map[string]int{}
	for _, li := range l.Lights {
		q := t.Inverse(li.Pos)
		d, _ := dist(q)
		switch {
		case d <= 64:
			hist["inside"]++
		case d <= 300:
			hist["<300"]++
		case d <= 1000:
			hist["<1000"]++
		default:
			hist[">1000"]++
			fmt.Printf("  far light %v (layout %v) %.0f away, type %d\n", li.Pos, q, d, li.Type)
		}
	}
	for _, e := range l.Entrances {
		q := t.Inverse(e.Centre())
		d, _ := dist(q)
		fmt.Printf("  entrance %s layout %v %.0f away\n", e.Name, q, d)
	}
	near := hist["inside"] + hist["<300"] + hist["e-near"]
	all := near + hist["<1000"] + hist[">1000"] + hist["e-far"]
	fmt.Printf("zone %d: %d/%d within 300 = %.2f  %v\n", n, near, all, float64(near)/float64(all), hist)
}
