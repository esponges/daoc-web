package main

// The sky. zones/zones.mpk's zones.dat maps a zone to its region, and the
// region names a skydome file -- zone 100 is region 100, which names
// sky_midgard.dat. That file, in zones/sky/sky.mpk, is an INI that describes
// the whole sky: canopy colours from zenith down to the horizon, east and
// west, for dawn, day and dusk; two scrolling cloud layers and how opaque
// they are at each height; the sun and moon; the fog and the light the sky
// casts. This converts the clear day: the viewer has no clock or weather.

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"daocweb/internal/dds"
	"daocweb/internal/mpak"
)

type skyOut struct {
	Source string `json:"source"`
	// Canopy colours, 0..255 RGB: the zenith, then four rings down to the
	// horizon on the east side and on the west, then below the horizon.
	Zenith [3]int    `json:"zenith"`
	East   [4][3]int `json:"east"`
	West   [4][3]int `json:"west"`
	Nadir  [3]int    `json:"nadir"`
	// Cloud tint and opacity, RGBA, at the same heights.
	CloudZenith [4]int     `json:"cloudZenith"`
	CloudEast   [4][4]int  `json:"cloudEast"`
	CloudWest   [4][4]int  `json:"cloudWest"`
	Clouds      []cloudOut `json:"clouds"`
	Fog         [3]int     `json:"fog"`
	// The light the sky casts: ambient and direct, each a colour already
	// scaled by its amount, 0..1.
	Ambient [3]float64 `json:"ambient"`
	Sun     [3]float64 `json:"sun"`
	SunDisc *discOut   `json:"sunDisc,omitempty"`
}

type cloudOut struct {
	Texture string  `json:"texture"`
	Tile    float64 `json:"tile"`  // repeats across the dome
	Speed   float64 `json:"speed"` // scroll, texture widths a second x 1000
}

// discOut is a celestial object's look at its highest. Colours are RGBA
// 0..255 and drawn additively; scale is its half-size at distance.
type discOut struct {
	Texture   string  `json:"texture"`
	Shape     [4]int  `json:"shape"`
	Glow      [4]int  `json:"glow"`
	Scale     float64 `json:"scale"`
	GlowScale float64 `json:"glowScale"`
	Distance  float64 `json:"distance"`
}

func buildSky(game string, zoneNum int, outDir string) (*skyOut, error) {
	n, reg, err := regionOf(game, zoneNum)
	if err != nil {
		return nil, err
	}
	file := reg["skydome"]
	if file == "" {
		file = "sky_default.dat"
	}

	skyDir := filepath.Join(game, "zones", "sky")
	sarc, err := mpak.Open(filepath.Join(skyDir, "sky.mpk"))
	if err != nil {
		return nil, err
	}
	raw, err := sarc.Read(file)
	if err != nil {
		return nil, err
	}
	ini := parseINI(raw)

	out := &skyOut{Source: file}
	c := "canopy_color_clear"
	out.Zenith = rgb(ini.get(c, "day_zenith"))
	out.Nadir = rgb(ini.get(c, "day_nadir"))
	for i := 0; i < 4; i++ {
		out.East[i] = rgb(ini.get(c, fmt.Sprintf("day_east%d", i+1)))
		out.West[i] = rgb(ini.get(c, fmt.Sprintf("day_west%d", i+1)))
	}
	c = "cloud_color_clear"
	out.CloudZenith = rgba(ini.get(c, "day_zenith"))
	for i := 0; i < 4; i++ {
		out.CloudEast[i] = rgba(ini.get(c, fmt.Sprintf("day_east%d", i+1)))
		out.CloudWest[i] = rgba(ini.get(c, fmt.Sprintf("day_west%d", i+1)))
	}
	l := "lights_and_fog_clear"
	out.Fog = rgb(ini.get(l, "day_distance_fog"))
	out.Ambient = scaled(ini.get(l, "day_ambient_light"), ini.get(l, "day_ambient_light_amount"))
	out.Sun = scaled(ini.get(l, "day_dynamic_light"), ini.get(l, "day_dynamic_light_amount"))

	dir := filepath.Join(outDir, "sky")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	// The clear-weather layers: top, then bottom, drawn in that order.
	for _, k := range []string{"top", "bot"} {
		tex := ini.get("main", "cloud_"+k+"_texture")
		if tex == "" {
			continue
		}
		png, err := convertSkyTexture(skyDir, tex, dir)
		if err != nil {
			return nil, err
		}
		out.Clouds = append(out.Clouds, cloudOut{
			Texture: "sky/" + png,
			Tile:    num(ini.get("main", "cloud_"+k+"_tile"), 1),
			Speed:   num(ini.get("main", "cloud_"+k+"_speed"), 0),
		})
	}

	// The sun is whichever celestial object shows by day.
	for i := 1; i <= 8; i++ {
		s := fmt.Sprintf("celestial_object%02d", i)
		if ini.get(s, "day_object") != "true" {
			continue
		}
		tex := ini.get(s, "shape_texture")
		png, err := convertSkyTexture(skyDir, tex, dir)
		if err != nil {
			// sundisk.bmp has a .tga twin; the art is the same disc.
			alt := strings.TrimSuffix(tex, filepath.Ext(tex)) + ".tga"
			if png, err = convertSkyTexture(skyDir, alt, dir); err != nil {
				return nil, err
			}
		}
		out.SunDisc = &discOut{
			Texture:   "sky/" + png,
			Shape:     rgba(ini.get(s, "shape_color_apex")),
			Glow:      rgba(ini.get(s, "glow_color_apex")),
			Scale:     num(ini.get(s, "shape_scale"), 15),
			GlowScale: num(ini.get(s, "glow_scale"), 15),
			Distance:  num(ini.get(s, "distance"), 1000),
		}
		break
	}
	fmt.Printf("  sky: %s (zone %d -> region %d), %d cloud layers, fog %v\n", file, zoneNum, n, len(out.Clouds), out.Fog)
	return out, nil
}

// regionOf finds a zone's region in zones/zones.mpk's zones.dat and returns
// its number and its settings -- skydome, grasscsv, grassmap and so on.
func regionOf(game string, zoneNum int) (int, map[string]string, error) {
	zarc, err := mpak.Open(filepath.Join(game, "zones", "zones.mpk"))
	if err != nil {
		return 0, nil, err
	}
	raw, err := zarc.Read("zones.dat")
	if err != nil {
		return 0, nil, err
	}
	zones := parseINI(raw)
	region := zones.get(fmt.Sprintf("zone%03d", zoneNum), "region")
	if region == "" {
		return 0, nil, fmt.Errorf("zones.dat has no region for zone %d", zoneNum)
	}
	n, _ := strconv.Atoi(region)
	reg := zones[fmt.Sprintf("region%03d", n)]
	if reg == nil {
		reg = map[string]string{}
	}
	return n, reg, nil
}

// convertSkyTexture writes one of zones/sky's images as PNG and returns
// its name.
func convertSkyTexture(dir, name, outDir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return "", err
	}
	var im image.Image
	switch strings.ToLower(filepath.Ext(name)) {
	case ".dds":
		im, err = dds.Decode(b)
	case ".tga":
		im, err = decodeTGA(b)
	default:
		return "", fmt.Errorf("%s: no decoder for this format", name)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	png := strings.TrimSuffix(strings.ToLower(name), filepath.Ext(name)) + ".png"
	return png, writePNG(filepath.Join(outDir, png), im)
}

// decodeTGA reads an uncompressed true-colour Targa, which is all zones/sky
// uses: 24 or 32 bits, BGR(A), rows bottom-up unless the descriptor flips
// them.
func decodeTGA(b []byte) (*image.NRGBA, error) {
	if len(b) < 18 {
		return nil, fmt.Errorf("tga: short header")
	}
	idLen, cmap, typ := int(b[0]), b[1], b[2]
	w := int(binary.LittleEndian.Uint16(b[12:]))
	h := int(binary.LittleEndian.Uint16(b[14:]))
	bpp, desc := int(b[16]), b[17]
	if typ != 2 || cmap != 0 || (bpp != 24 && bpp != 32) {
		return nil, fmt.Errorf("tga: type %d, %d bits not supported", typ, bpp)
	}
	px := b[18+idLen:]
	step := bpp / 8
	if len(px) < w*h*step {
		return nil, fmt.Errorf("tga: %d bytes of pixels for %dx%d", len(px), w, h)
	}
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	topDown := desc&0x20 != 0
	for y := 0; y < h; y++ {
		row := y
		if !topDown {
			row = h - 1 - y
		}
		for x := 0; x < w; x++ {
			s := px[(row*w+x)*step:]
			o := (y*w + x) * 4
			im.Pix[o], im.Pix[o+1], im.Pix[o+2], im.Pix[o+3] = s[2], s[1], s[0], 255
			if step == 4 {
				im.Pix[o+3] = s[3]
			}
		}
	}
	return im, nil
}

// --- INI ---

type iniFile map[string]map[string]string

// parseINI reads the game's INI dialect: [sections], key = value, and ';'
// comments, which may also trail a line.
func parseINI(b []byte) iniFile {
	out := iniFile{}
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

func (f iniFile) get(sec, key string) string {
	return f[strings.ToLower(sec)][strings.ToLower(key)]
}

func ints(s string) []int {
	var out []int
	for _, p := range strings.Split(s, ",") {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil {
			out = append(out, v)
		}
	}
	return out
}

func rgb(s string) [3]int {
	v := ints(s)
	var o [3]int
	copy(o[:], v)
	return o
}

func rgba(s string) [4]int {
	v := ints(s)
	o := [4]int{0, 0, 0, 255}
	copy(o[:], v)
	return o
}

func num(s string, def float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return def
	}
	return v
}

func scaled(color, amount string) [3]float64 {
	c, a := rgb(color), num(amount, 1)
	return [3]float64{float64(c[0]) / 255 * a, float64(c[1]) / 255 * a, float64(c[2]) / 255 * a}
}
