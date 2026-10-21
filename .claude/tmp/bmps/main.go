package main

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strings"

	"daocweb/internal/mpak"
)

func main() {
	game := `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`
	var paths []string
	for _, g := range []string{"zones/zone*/tex*.mpk", "frontiers/zones/zone*/tex*.mpk"} {
		m, _ := filepath.Glob(filepath.Join(game, g))
		paths = append(paths, m...)
	}
	variants := map[string]int{}
	ext := map[string]int{}
	for _, p := range paths {
		a, err := mpak.Open(p)
		if err != nil {
			continue
		}
		for _, e := range a.Entries {
			ext[strings.ToLower(filepath.Ext(e.Name))]++
			if !strings.EqualFold(filepath.Ext(e.Name), ".bmp") {
				continue
			}
			b, err := a.ReadEntry(e)
			if err != nil {
				variants["read error"]++
				continue
			}
			le := binary.LittleEndian
			k := fmt.Sprintf("hdr=%d %dx%d bpp=%d comp=%d off=%d len=%d", le.Uint32(b[14:]), int32(le.Uint32(b[18:])), int32(le.Uint32(b[22:])), le.Uint16(b[28:]), le.Uint32(b[30:]), le.Uint32(b[10:]), len(b))
			variants[k]++
		}
	}
	fmt.Println(ext)
	for k, v := range variants {
		fmt.Println(v, k)
	}
}
