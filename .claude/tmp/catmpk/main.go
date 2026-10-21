package main

import (
	"fmt"
	"os"

	"daocweb/internal/mpak"
)

func main() {
	a, err := mpak.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	b, err := a.Read(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Stdout.Write(b)
}
