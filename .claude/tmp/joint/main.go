package main

import (
	"fmt"
	"math"

	"daocweb/internal/interior"
)

func main() {
	game := `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`
	for _, n := range []int{120, 125, 126, 127, 128, 129, 149, 150, 160, 161, 188, 189, 226} {
		l, _ := interior.Read(game, n)
		placed, _ := l.Placed()
		off, scaled := 0, 0
		worst := 0.0
		for _, p := range placed {
			// z column of the rotation part, normalised: how far model up
			// tips from layout up.
			sx := math.Sqrt(float64(p.M[0]*p.M[0] + p.M[4]*p.M[4] + p.M[8]*p.M[8]))
			zz := float64(p.M[10]) / sx
			tip := math.Acos(math.Max(-1, math.Min(1, math.Abs(zz))))
			if tip > 1e-3 {
				off++
				worst = math.Max(worst, tip)
			}
			if math.Abs(sx-1) > 1e-3 {
				scaled++
			}
		}
		fmt.Printf("zone %d: %d placed, %d tipped off vertical (worst %.1f deg), %d scaled\n", n, len(placed), off, worst*180/math.Pi, scaled)
	}
}
