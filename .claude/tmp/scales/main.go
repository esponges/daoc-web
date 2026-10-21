package main

import (
	"fmt"
	"sort"

	"daocweb/internal/zone"
)

func main() {
	game := `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`
	list, _ := zone.List(game)
	var all []float32
	for _, in := range list {
		z, err := zone.Open(game, in.Num)
		if err != nil {
			continue
		}
		fx, err := z.PlacedFixtures()
		if err != nil {
			continue
		}
		for _, f := range fx {
			all = append(all, f.Scale)
			if f.Scale > 8 {
				fmt.Printf("zone %d nif %d scale %.2f\n", in.Num, f.NifID, f.Scale)
			}
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	n := len(all)
	fmt.Println("n", n, "p50", all[n/2], "p99", all[n*99/100], "p999", all[n*999/1000], "max", all[n-1])
}
