// Command worldconv converts every zone the census marks convertible -- the
// outdoor zones with zoneconv and propconv, the cities and dungeons with
// interiorconv -- and checks the result against the census.
//
//	go run ./cmd/census       # first, so the zone list is current
//	go run ./cmd/worldconv    # then convert them all into web/data
//
// An outdoor zone is convertible when its terrain reads. Its scenery is
// converted when its tables read; propconv must then place exactly as many
// fixtures as the census counted converting, since both bake through
// internal/scenery. A city or dungeon is convertible when its layout reads,
// and interiorconv must likewise write as many placements as the census
// counted converting.
// Everything written lands in web/data, which is not committed, along with
// zones.json, the list of converted zones the viewer's zone picker offers.
// -index rewrites that list from what is already converted, converting
// nothing; -only outdoor or -only interiors converts just those.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

// The census rows this needs; see cmd/census.
type census struct {
	Zones struct {
		Rows      []row `json:"rows"`
		Interiors []row `json:"interiors"`
	} `json:"zones"`
}

type row struct {
	Num        int    `json:"num"`
	Name       string `json:"name"`
	Kind       string `json:"kind"` // a city or dungeon's; empty outdoors
	Layout     string `json:"layout"`
	Frontier   bool   `json:"frontier"`
	Terrain    string `json:"terrain"`
	Scenery    string `json:"scenery"`
	Placements int    `json:"placements"`
	Converting int    `json:"converting"`
}

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	report := flag.String("census", filepath.Join("docs", "census", "census.json"), "the census to take zones from")
	out := flag.String("out", filepath.Join("web", "data"), "output directory")
	workers := flag.Int("j", runtime.NumCPU()/2+1, "zones converted in parallel")
	indexOnly := flag.Bool("index", false, "only rewrite zones.json from the zones already converted")
	only := flag.String("only", "", "convert only the \"outdoor\" zones or only the \"interiors\"")
	flag.Parse()
	if *only != "" && *only != "outdoor" && *only != "interiors" {
		fmt.Fprintln(os.Stderr, "worldconv: -only takes outdoor or interiors")
		os.Exit(2)
	}
	if *indexOnly {
		rows, err := readCensus(*report)
		if err == nil {
			err = writeIndex(*out, rows)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "worldconv:", err)
			os.Exit(1)
		}
		return
	}
	if err := run(*game, *report, *out, *workers, *only); err != nil {
		fmt.Fprintln(os.Stderr, "worldconv:", err)
		os.Exit(1)
	}
}

func readCensus(report string) ([]row, error) {
	raw, err := os.ReadFile(report)
	if err != nil {
		return nil, fmt.Errorf("%w (run the census first)", err)
	}
	var c census
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	return append(c.Zones.Rows, c.Zones.Interiors...), nil
}

// indexEntry is one zones.json entry: a converted zone, and how much of its
// scenery converted.
type indexEntry struct {
	Num        int    `json:"num"`
	Name       string `json:"name"`
	Frontier   bool   `json:"frontier,omitempty"`
	Kind       string `json:"kind,omitempty"` // "city" or "dungeon"; outdoor zones leave it out
	Placements int    `json:"placements"`
	Converting int    `json:"converting"`
}

// writeIndex lists every census zone whose zone.json is in out.
func writeIndex(out string, rows []row) error {
	var idx []indexEntry
	for _, r := range rows {
		if _, err := os.Stat(filepath.Join(out, fmt.Sprintf("zone%03d", r.Num), "zone.json")); err != nil {
			continue
		}
		idx = append(idx, indexEntry{r.Num, r.Name, r.Frontier, r.Kind, r.Placements, r.Converting})
	}
	sort.Slice(idx, func(i, j int) bool { return idx[i].Num < idx[j].Num })
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "zones.json"), append(b, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("zones.json lists %d converted zones\n", len(idx))
	return nil
}

func run(game, report, out string, workers int, only string) error {
	rows, err := readCensus(report)
	if err != nil {
		return err
	}
	var zones []row
	for _, r := range rows {
		interior := r.Kind != ""
		if only == "outdoor" && interior || only == "interiors" && !interior {
			continue
		}
		if r.Terrain == "ok" || r.Layout == "ok" {
			zones = append(zones, r)
		}
	}
	sort.Slice(zones, func(i, j int) bool { return zones[i].Num < zones[j].Num })

	// Build the converters once rather than go-run them per zone.
	bin, err := os.MkdirTemp("", "worldconv")
	if err != nil {
		return err
	}
	defer os.RemoveAll(bin)
	exe := func(name string) string {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		return filepath.Join(bin, name)
	}
	for _, cmd := range []string{"zoneconv", "propconv", "interiorconv"} {
		b := exec.Command("go", "build", "-o", exe(cmd), "./cmd/"+cmd)
		b.Stdout, b.Stderr = os.Stdout, os.Stderr
		if err := b.Run(); err != nil {
			return fmt.Errorf("build %s: %w", cmd, err)
		}
	}

	problems := make([]string, len(zones))
	next := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				problems[i] = convert(game, out, exe, zones[i])
				status := "ok"
				if problems[i] != "" {
					status = problems[i]
				}
				fmt.Printf("zone %03d %-28s %s\n", zones[i].Num, zones[i].Name, status)
			}
		}()
	}
	for i := range zones {
		next <- i
	}
	close(next)
	wg.Wait()

	var bad []string
	for i, p := range problems {
		if p != "" {
			bad = append(bad, fmt.Sprintf("zone %03d: %s", zones[i].Num, p))
		}
	}
	fmt.Printf("\n%d zones converted, %d with problems\n", len(zones)-len(bad), len(bad))
	if err := writeIndex(out, rows); err != nil {
		return err
	}
	if len(bad) > 0 {
		return fmt.Errorf("\n  %s", strings.Join(bad, "\n  "))
	}
	return nil
}

// convert runs both converters on a zone and checks what they wrote. It
// returns what is wrong, or "".
func convert(game, out string, exe func(string) string, r row) string {
	num := fmt.Sprint(r.Num)
	dir := filepath.Join(out, fmt.Sprintf("zone%03d", r.Num))
	if r.Kind != "" {
		return convertInterior(game, out, exe, r)
	}
	if msg, err := exec.Command(exe("zoneconv"), "-game", game, "-zone", num, "-name", r.Name, "-out", out).CombinedOutput(); err != nil {
		return "zoneconv: " + lastLine(msg)
	}
	if _, err := os.Stat(filepath.Join(dir, "zone.json")); err != nil {
		return "no zone.json"
	}
	if r.Scenery != "" {
		return "" // no scenery tables to convert, as the census says
	}
	if msg, err := exec.Command(exe("propconv"), "-game", game, "-zone", num, "-out", out).CombinedOutput(); err != nil {
		return "propconv: " + lastLine(msg)
	}
	b, err := os.ReadFile(filepath.Join(dir, "props", "props.json"))
	if err != nil {
		return "no props.json"
	}
	var m struct {
		Instances []json.RawMessage `json:"instances"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return "props.json: " + err.Error()
	}
	if len(m.Instances) != r.Converting {
		return fmt.Sprintf("%d placements written, the census counts %d converting", len(m.Instances), r.Converting)
	}
	return ""
}

// convertInterior runs interiorconv on a city or dungeon and checks that it
// wrote as many placements as the census counts converting.
func convertInterior(game, out string, exe func(string) string, r row) string {
	dir := filepath.Join(out, fmt.Sprintf("zone%03d", r.Num))
	if msg, err := exec.Command(exe("interiorconv"), "-game", game, "-zone", fmt.Sprint(r.Num), "-name", r.Name, "-out", out).CombinedOutput(); err != nil {
		return "interiorconv: " + lastLine(msg)
	}
	b, err := os.ReadFile(filepath.Join(dir, "props", "props.json"))
	if err != nil {
		return "no props.json"
	}
	var m struct {
		Instances []json.RawMessage `json:"instances"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return "props.json: " + err.Error()
	}
	if _, err := os.Stat(filepath.Join(dir, "zone.json")); err != nil {
		return "no zone.json"
	}
	if len(m.Instances) != r.Converting {
		return fmt.Sprintf("%d placements written, the census counts %d converting", len(m.Instances), r.Converting)
	}
	return ""
}

func lastLine(b []byte) string {
	s := strings.TrimSpace(string(b))
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	}
	return s
}
