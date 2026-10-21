package main

import (
	"fmt"
	"os"

	"daocweb/internal/mpak"
)

func main() {
	for _, p := range os.Args[1:] {
		a, err := mpak.Open(p)
		if err != nil {
			fmt.Println(p, err)
			continue
		}
		fmt.Println(p, len(a.Entries))
		for _, e := range a.Entries {
			fmt.Println("  ", e.Name)
		}
	}
}
