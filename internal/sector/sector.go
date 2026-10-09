// Package sector parses SECTOR.DAT, the per-zone INI file that carries the
// terrain height scaling, the zone's sector grid size and its water bodies.
package sector

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// File is a parsed SECTOR.DAT. Sections keep insertion order so that indexed
// keys (left00, right00, left01, ...) can be walked predictably.
type File struct {
	sections map[string]map[string]string
	order    []string
}

// Water is one body of water declared in the zone.
//
// The shoreline is stored as two matching chains, leftNN and rightNN, which
// together describe a ribbon: left[i] and right[i] are opposite banks at the
// same point along the body. They must stay separate, because the surface is
// triangulated as a strip between the two chains, not as one ring.
type Water struct {
	ID      string // section name, e.g. "river00"
	Name    string // friendly name, e.g. "Lake"
	Type    string // LAKE, RIVER, ...
	Height  int    // surface elevation in world units
	Color   int
	Texture string
	Left    [][2]int // one bank, in heightmap cells
	Right   [][2]int // the opposite bank, same length as Left
}

// Parse reads SECTOR.DAT content.
func Parse(b []byte) (*File, error) {
	f := &File{sections: map[string]map[string]string{}}
	cur := ""
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			cur = strings.ToLower(line[1 : len(line)-1])
			if _, ok := f.sections[cur]; !ok {
				f.sections[cur] = map[string]string{}
				f.order = append(f.order, cur)
			}
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || cur == "" {
			continue
		}
		f.sections[cur][strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(f.sections) == 0 {
		return nil, fmt.Errorf("sector: no sections found")
	}
	return f, nil
}

// Str returns a value, or def when absent.
func (f *File) Str(section, key, def string) string {
	if s, ok := f.sections[strings.ToLower(section)]; ok {
		if v, ok := s[strings.ToLower(key)]; ok {
			return v
		}
	}
	return def
}

// Int returns a value parsed as an integer, or def when absent or unparseable.
func (f *File) Int(section, key string, def int) int {
	v := f.Str(section, key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return n
}

// ScaleFactor and OffsetFactor convert the two 8-bit height rasters into world
// units: height = terrain*ScaleFactor + offset*OffsetFactor.
func (f *File) ScaleFactor() int  { return f.Int("terrain", "scalefactor", 8) }
func (f *File) OffsetFactor() int { return f.Int("terrain", "offsetfactor", 48) }

// SectorSize is the zone's sector grid, which also gives the LOD texture grid.
func (f *File) SectorSize() (x, y int) {
	x, y = f.Int("sectorsize", "sizex", 8), f.Int("sectorsize", "sizey", 8)
	// Six zones -- Avalon Isle, Raumarik, Uppland and others -- write 0 here,
	// though their LOD archives hold the usual 8x8 tiles and zones.dat
	// gives them a width and height of 8.
	if x <= 0 {
		x = 8
	}
	if y <= 0 {
		y = 8
	}
	return x, y
}

// Waters returns every water body declared by [waterdefs].
func (f *File) Waters() []Water {
	n := f.Int("waterdefs", "num", 0)
	out := make([]Water, 0, n)
	for _, name := range f.order {
		s := f.sections[name]
		if _, ok := s["height"]; !ok {
			continue
		}
		if _, ok := s["bankpoints"]; !ok {
			continue
		}
		w := Water{
			ID:      name,
			Name:    f.Str(name, "name", name),
			Type:    strings.ToUpper(f.Str(name, "type", "")),
			Height:  f.Int(name, "height", 0),
			Color:   f.Int(name, "color", 0),
			Texture: f.Str(name, "texture", ""),
		}
		pairs := f.Int(name, "bankpoints", 0)
		for i := 0; i < pairs; i++ {
			l, lok := f.point(name, fmt.Sprintf("left%02d", i))
			r, rok := f.point(name, fmt.Sprintf("right%02d", i))
			// Keep the chains index-aligned: a half pair would skew the strip.
			if lok && rok {
				w.Left = append(w.Left, l)
				w.Right = append(w.Right, r)
			}
		}
		out = append(out, w)
	}
	return out
}

// point parses an "x,y[,z]" value into heightmap cell coordinates.
func (f *File) point(section, key string) ([2]int, bool) {
	v := f.Str(section, key, "")
	if v == "" {
		return [2]int{}, false
	}
	parts := strings.Split(v, ",")
	if len(parts) < 2 {
		return [2]int{}, false
	}
	x, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	y, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return [2]int{}, false
	}
	return [2]int{x, y}, true
}
