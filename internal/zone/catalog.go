package zone

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"daocweb/internal/mpak"
)

// INI is the game's INI dialect: [sections] of key = value lines, both
// lower-cased on reading.
type INI map[string]map[string]string

// ParseINI reads the dialect: [sections], key = value, and ';' comments,
// which may also trail a line.
func ParseINI(b []byte) INI {
	out := INI{}
	sec := ""
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, ';'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sec = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if out[sec] == nil {
			out[sec] = map[string]string{}
		}
		out[sec][strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	return out
}

// Get returns a key's value, or "" when it is absent.
func (f INI) Get(sec, key string) string {
	return f[strings.ToLower(sec)][strings.ToLower(key)]
}

// Catalog reads zones/zones.mpk's zones.dat: every zone, and every region's
// settings -- skydome, grasscsv, grassmap and so on.
func Catalog(game string) (INI, error) {
	a, err := mpak.Open(filepath.Join(game, "zones", "zones.mpk"))
	if err != nil {
		return nil, err
	}
	raw, err := a.Read("zones.dat")
	if err != nil {
		return nil, err
	}
	return ParseINI(raw), nil
}

// RegionOf finds a zone's region and returns its number and settings.
func RegionOf(game string, num int) (int, map[string]string, error) {
	cat, err := Catalog(game)
	if err != nil {
		return 0, nil, err
	}
	region := cat.Get(fmt.Sprintf("zone%03d", num), "region")
	if region == "" {
		return 0, nil, fmt.Errorf("zones.dat has no region for zone %d", num)
	}
	n, _ := strconv.Atoi(region)
	reg := cat[fmt.Sprintf("region%03d", n)]
	if reg == nil {
		reg = map[string]string{}
	}
	return n, reg, nil
}

// Kind is a zone's declared type in zones.dat.
type Kind int

// The types zones.dat declares. A zone with no type line is terrain, like
// type 0: the Atlantis zones and a few others leave it out, and their
// folders hold the same heightmap and scenery archives.
const (
	Outdoor Kind = 0
	City    Kind = 1
	Dungeon Kind = 2
	Other   Kind = 4
)

func (k Kind) String() string {
	switch k {
	case Outdoor:
		return "outdoor"
	case City:
		return "city"
	case Dungeon:
		return "dungeon"
	}
	return "other"
}

// Info is one zone's zones.dat entry.
type Info struct {
	Num      int
	Name     string
	Kind     Kind
	Untyped  bool // no type line; counted as outdoor
	Region   int
	Frontier bool
}

// List reads every zone zones.dat declares, in number order.
func List(game string) ([]Info, error) {
	cat, err := Catalog(game)
	if err != nil {
		return nil, err
	}
	var out []Info
	for sec, kv := range cat {
		if !strings.HasPrefix(sec, "zone") {
			continue
		}
		n, err := strconv.Atoi(sec[4:])
		if err != nil {
			continue
		}
		in := Info{Num: n, Name: kv["name"], Frontier: kv["frontiers"] == "1"}
		in.Region, _ = strconv.Atoi(kv["region"])
		if t, ok := kv["type"]; ok {
			k, _ := strconv.Atoi(t)
			in.Kind = Kind(k)
			if k != 0 && k != 1 && k != 2 {
				in.Kind = Other
			}
		} else {
			in.Untyped = true
		}
		out = append(out, in)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Num < out[j].Num })
	return out, nil
}
