// Package zone reads an outdoor zone's own data: its SECTOR.DAT settings, its
// heightmap, and the scenery its fixtures place.
//
// An outdoor zone is a folder of archives, zones/zoneNNN or, for the frontier
// zones, frontiers/zones/zoneNNN. datNNN.mpk holds the settings, the height
// rasters and two CSVs. nifs.csv maps a numeric id to a model file, and
// fixtures.csv places that id in the world:
//
//	ID,NIF #,Textual Name,X,Y,Z,A,Scale,...,3D Angle,3D Axis X,3D Axis Y,3D Axis Z
//	1,403,Pine,52736.00,19200.00,4960.00,225,150,...,2.356194,0,0,1
//
// The Z is absolute world height, not an offset: it matches the heightmap to
// the unit, which is a useful check on both. Rotation is taken from the
// axis-angle columns rather than the A column, because A is a clockwise
// heading while the axis-angle pair is a plain rotation and needs no
// convention guessed. Scale is a percentage.
package zone

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"daocweb/internal/mpak"
	"daocweb/internal/pcx"
	"daocweb/internal/sector"
)

// Units is the side of a zone in world units.
const Units = 65536

// Zone is an outdoor zone's folder and its open dat archive.
type Zone struct {
	Num int
	Dir string
	// Frontier is set for a zone under frontiers/, whose scenery models
	// live in frontiers/NIFS rather than zones/Nifs.
	Frontier bool
	Dat      *mpak.Archive
}

// Dir finds a zone's folder: zones/zoneNNN, else frontiers/zones/zoneNNN.
func Dir(game string, num int) (dir string, frontier bool, err error) {
	name := fmt.Sprintf("zone%03d", num)
	dir = filepath.Join(game, "zones", name)
	_, err = os.Stat(dir)
	if err == nil {
		return dir, false, nil
	}
	if fd := filepath.Join(game, "frontiers", "zones", name); statOK(fd) {
		return fd, true, nil
	}
	return "", false, fmt.Errorf("zone directory %s: %w", dir, err)
}

func statOK(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// Open finds a zone's folder and opens its dat archive.
func Open(game string, num int) (*Zone, error) {
	dir, frontier, err := Dir(game, num)
	if err != nil {
		return nil, err
	}
	dat, err := mpak.Open(filepath.Join(dir, fmt.Sprintf("dat%03d.mpk", num)))
	if err != nil {
		return nil, err
	}
	return &Zone{Num: num, Dir: dir, Frontier: frontier, Dat: dat}, nil
}

// Archive opens another of the zone's archives by kind: "lod", "ter", "tex".
func (z *Zone) Archive(kind string) (*mpak.Archive, error) {
	return mpak.Open(filepath.Join(z.Dir, fmt.Sprintf("%s%03d.mpk", kind, z.Num)))
}

// Entry reads a dat entry by base name, ignoring case.
func (z *Zone) Entry(name string) ([]byte, error) {
	return Entry(z.Dat, name)
}

// Entry reads an archive entry by base name, ignoring case.
func Entry(a *mpak.Archive, name string) ([]byte, error) {
	for _, e := range a.Entries {
		if strings.EqualFold(filepath.Base(e.Name), name) {
			return a.ReadEntry(e)
		}
	}
	return nil, fmt.Errorf("%s not found in archive", name)
}

// Sector parses the zone's SECTOR.DAT.
func (z *Zone) Sector() (*sector.File, error) {
	raw, err := z.Dat.Read("SECTOR.DAT")
	if err != nil {
		return nil, err
	}
	return sector.Parse(raw)
}

// Terrain is a zone's heightmap in world units.
type Terrain struct {
	Grid     int      // samples per side
	Heights  []uint16 // Grid*Grid, row-major
	Min, Max int
}

// Terrain builds the heightmap from terrain.pcx and offset.pcx, scaled by
// the sector's factors.
func (z *Zone) Terrain(sec *sector.File) (*Terrain, error) {
	sf, of := sec.ScaleFactor(), sec.OffsetFactor()
	ter, err := z.PCX("terrain.pcx")
	if err != nil {
		return nil, err
	}
	offs, err := z.PCX("offset.pcx")
	if err != nil {
		return nil, err
	}
	if ter.W != offs.W || ter.H != offs.H {
		return nil, fmt.Errorf("terrain %dx%d and offset %dx%d disagree", ter.W, ter.H, offs.W, offs.H)
	}
	grid := ter.W
	t := &Terrain{Grid: grid, Heights: make([]uint16, grid*grid), Min: math.MaxInt32}
	for y := 0; y < grid; y++ {
		for x := 0; x < grid; x++ {
			h := int(ter.At(x, y))*sf + int(offs.At(x, y))*of
			if h > 0xFFFF {
				return nil, fmt.Errorf("height %d at (%d,%d) overflows uint16", h, x, y)
			}
			t.Heights[y*grid+x] = uint16(h)
			if h < t.Min {
				t.Min = h
			}
			if h > t.Max {
				t.Max = h
			}
		}
	}
	return t, nil
}

// HeightAt is the terrain height at a world position, bilinear between
// samples, as the viewer puts the character on the ground.
func (t *Terrain) HeightAt(x, y float32) float32 {
	cell := float64(Units) / float64(t.Grid)
	gx, gy := float64(x)/cell, float64(y)/cell
	x0, y0 := int(math.Floor(gx)), int(math.Floor(gy))
	fx, fy := gx-float64(x0), gy-float64(y0)
	h := func(cx, cy int) float64 {
		cx = max(0, min(t.Grid-1, cx))
		cy = max(0, min(t.Grid-1, cy))
		return float64(t.Heights[cy*t.Grid+cx])
	}
	return float32((h(x0, y0)*(1-fx)+h(x0+1, y0)*fx)*(1-fy) + (h(x0, y0+1)*(1-fx)+h(x0+1, y0+1)*fx)*fy)
}

// PlaceOnGround sets the height of every OnGround fixture from the terrain.
func (t *Terrain) PlaceOnGround(fs []Fixture) {
	for i := range fs {
		if fs[i].OnGround {
			fs[i].Z = t.HeightAt(fs[i].X, fs[i].Y)
		}
	}
}

// PCX decodes one of the dat archive's rasters.
func (z *Zone) PCX(name string) (*pcx.Image, error) {
	raw, err := z.Dat.Read(name)
	if err != nil {
		return nil, err
	}
	im, err := pcx.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return im, nil
}

// ModelDef is a nifs.csv row: an id and the model file it names.
type ModelDef struct {
	ID   int
	Name string
	File string
}

// Fixture is a fixtures.csv row: a model placed in the world.
type Fixture struct {
	NifID int
	X     float32
	Y     float32
	Z     float32
	Yaw   float32
	Scale float32
	// OnGround is set for a fixture whose height the terrain gives: in the
	// older table these have a Z of 0 and the Ground flag. Callers set Z
	// from the heightmap; see Terrain.HeightAt.
	OnGround bool
}

// Models reads the zone's nifs.csv.
func (z *Zone) Models() (map[int]ModelDef, error) {
	b, err := z.Entry("nifs.csv")
	if err != nil {
		return nil, err
	}
	m, err := ParseNifs(b)
	if err != nil {
		return nil, fmt.Errorf("nifs.csv: %w", err)
	}
	return m, nil
}

// Fixtures reads the zone's fixtures.csv.
func (z *Zone) Fixtures() ([]Fixture, error) {
	b, err := z.Entry("fixtures.csv")
	if err != nil {
		return nil, err
	}
	f, err := ParseFixtures(b)
	if err != nil {
		return nil, fmt.Errorf("fixtures.csv: %w", err)
	}
	return f, nil
}

// Both CSVs carry two header rows: a banner and the column names.
func records(b []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(b))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 3 {
		return nil, fmt.Errorf("only %d rows", len(rows))
	}
	return rows[2:], nil
}

// ParseNifs reads nifs.csv.
func ParseNifs(b []byte) (map[int]ModelDef, error) {
	rows, err := records(b)
	if err != nil {
		return nil, err
	}
	out := map[int]ModelDef{}
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(row[0]))
		if err != nil {
			continue
		}
		out[id] = ModelDef{ID: id, Name: strings.TrimSpace(row[1]), File: strings.TrimSpace(row[2])}
	}
	return out, nil
}

// ParseFixtures reads fixtures.csv.
func ParseFixtures(b []byte) ([]Fixture, error) {
	rows, err := records(b)
	if err != nil {
		return nil, err
	}
	var out []Fixture
	for _, row := range rows {
		// Two layouts. The newer has 19 columns, ending in an axis-angle
		// rotation. The older, which 54 of the outdoor zones use, stops at
		// 15: no axis-angle, and a Z of 0 with the Ground flag set for
		// anything standing on the terrain.
		if len(row) < 15 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(row[1]))
		if err != nil {
			continue
		}
		f := Fixture{NifID: id}
		f.X = atof(row[3])
		f.Y = atof(row[4])
		f.Z = atof(row[5])
		if len(row) >= 19 {
			// The A column is a clockwise heading in degrees; the
			// axis-angle columns say the same thing without a convention
			// to guess, so take the angle and let the Z axis sign carry
			// the direction.
			angle := atof(row[15])
			axisZ := atof(row[18])
			if axisZ < 0 {
				angle = -angle
			}
			f.Yaw = angle
		} else {
			// The heading alone. Where both are given, the axis-angle
			// yaw is minus the heading, wrapped to a half turn either
			// way: 959 of Mularn's 1006 fixtures agree to 0.005 radians,
			// and the rest are tilted as well as turned.
			f.Yaw = headingYaw(atof(row[6]))
			f.OnGround = atoi(row[11]) == 1 && f.Z == 0
		}
		f.Scale = atof(row[7]) / 100
		// Past maxScale it is a typo in the table, not a size: one oak in
		// West Svealand reads 175150 among neighbours of 150 to 450, and
		// drawn that big it walls off half the zone.
		if f.Scale <= 0 || f.Scale > maxScale {
			f.Scale = 1
		}
		out = append(out, f)
	}
	return out, nil
}

// maxScale is the largest placement scale taken at its word. Across every
// zone in the install, 99.9% of fixtures are at most 5x and the largest
// sensible one is 12x; the one above that is 1751x.
const maxScale = 50

// headingYaw turns a clockwise heading in degrees into a yaw in radians,
// in (-pi, pi].
func headingYaw(deg float32) float32 {
	y := -float64(deg) * math.Pi / 180
	y = math.Mod(y, 2*math.Pi)
	if y <= -math.Pi {
		y += 2 * math.Pi
	} else if y > math.Pi {
		y -= 2 * math.Pi
	}
	return float32(y)
}

func atoi(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

func atof(s string) float32 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 32)
	return float32(v)
}

// PlacedFixtures reads fixtures.csv and gives the fixtures standing on the
// terrain their height from it, reading the heightmap only if any need it.
func (z *Zone) PlacedFixtures() ([]Fixture, error) {
	fs, err := z.Fixtures()
	if err != nil {
		return nil, err
	}
	for _, f := range fs {
		if !f.OnGround {
			continue
		}
		sec, err := z.Sector()
		if err != nil {
			return nil, err
		}
		t, err := z.Terrain(sec)
		if err != nil {
			return nil, err
		}
		t.PlaceOnGround(fs)
		break
	}
	return fs, nil
}
