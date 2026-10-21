// Scratch: the role columns of anims.csv and canims.csv in the install.
package main

import (
	"fmt"
	"sort"

	"daocweb/internal/gamedata"
)

func main() {
	t, err := gamedata.Load(`C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`)
	if err != nil {
		panic(err)
	}
	for name, sets := range map[string]map[int]map[string]int{"anims": t.AnimSets, "canims": t.CombatSets} {
		cols := map[string]bool{}
		for _, row := range sets {
			for c := range row {
				cols[c] = true
			}
		}
		var l []string
		for c := range cols {
			l = append(l, c)
		}
		sort.Strings(l)
		fmt.Println(name, l)
	}
}
