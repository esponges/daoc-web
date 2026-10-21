// Scratch: load a world store and print the start check.
package main

import (
	"fmt"
	"os"

	"daocweb/internal/server/world"
)

func main() {
	cat, err := world.LoadCatalog(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println("unsafe:", len(cat.Unsafe))
	s, ok := cat.StartFor(2, 6, 22)
	fmt.Println("troll warrior:", s, ok)
}
