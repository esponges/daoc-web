// uiconv converts the game's interface data into what the browser loads.
//
//	go run ./cmd/uiconv -skin atlantis
//
// It reads <game>/ui/<skin>/ and every file ui/uimain.xml includes, and writes
// to web/data/ui/<skin>/: layout.json (the merged elements in file order),
// the textures and bitmap fonts they name as PNG, the TrueType fonts, and the
// icon sheets. Nothing is written outside web/data/, which is not committed.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"daocweb/internal/ui"
)

const defaultGame = `C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`

func main() {
	game := flag.String("game", defaultGame, "DAoC install directory")
	skin := flag.String("skin", "atlantis", "UI skin directory under ui/")
	out := flag.String("out", filepath.Join("web", "data", "ui"), "output directory")
	flag.Parse()
	if err := run(*game, *skin, *out); err != nil {
		fmt.Fprintln(os.Stderr, "uiconv:", err)
		os.Exit(1)
	}
}

// Layout is the file the browser loads.
type Layout struct {
	Skin     string            `json:"skin"`
	Textures []string          `json:"textures"` // files written as PNG, relative to the skin output
	Fonts    []string          `json:"fonts"`    // TrueType files copied alongside
	Sizes    map[string][2]int `json:"sizes"`    // width and height of each texture in Textures
	Icons    []IconSheet       `json:"icons"`    // icon sheets, ordered by first number
	// Bitmap fonts by the name layouts use: the sheet's PNG and its glyph table.
	Bitmap   map[string]*BitmapEntry `json:"bitmap"`
	Elements []*ui.Node              `json:"elements"`
}

// IconSheet is one icons/<kind>_<first>.bmp: icon n is on the sheet whose
// first = n - n%100, at column n%10, row n%100/10, each icon 32 pixels square.
type IconSheet struct {
	Kind  string `json:"kind"` // itm, wpn, spl, skl, cbt
	First int    `json:"first"`
	File  string `json:"file"`
}

// BitmapEntry is one bitmap font as the browser loads it.
type BitmapEntry struct {
	File string `json:"file"`
	*ui.BitmapFont
}

func run(game, skin, out string) error {
	uiDir := filepath.Join(game, "ui")
	s, missingIncludes, err := ui.LoadSkin(uiDir, skin)
	if err != nil {
		return err
	}
	skinOut := filepath.Join(out, skin)
	if err := os.MkdirAll(skinOut, 0o755); err != nil {
		return err
	}

	files := map[string]bool{}
	bitmapFonts := map[string]string{}
	ttf := map[string]bool{}
	tags := map[string]int{}
	var walk func(n *ui.Node)
	walk = func(n *ui.Node) {
		tags[n.Tag]++
		switch n.Tag {
		case "Texture", "Canvas":
			if f := n.Str("File"); f != "" {
				files[f] = true
			}
		case "Font":
			// A font definition in assets.xml has a File; a font reference
			// inside a template has colours and no File.
			if f := n.Str("File"); f != "" {
				files[f] = true
				bitmapFonts[n.Str("Name")] = f
			}
		case "TTFFont":
			if f := n.Str("File"); f != "" {
				ttf[f] = true
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, e := range s.Elements {
		walk(e)
	}

	lay := Layout{Skin: skin, Elements: s.Elements, Bitmap: map[string]*BitmapEntry{}, Sizes: map[string][2]int{}}
	var missingTex, failedTex, odd []string
	for _, f := range sortedKeys(files) {
		b, err := readInsensitive(uiDir, f)
		if err != nil {
			missingTex = append(missingTex, f)
			continue
		}
		img, err := ui.DecodeImage(f, b)
		if err != nil {
			failedTex = append(failedTex, fmt.Sprintf("%s (%v)", f, err))
			continue
		}
		if err := writePNG(filepath.Join(skinOut, pngName(f)), img); err != nil {
			return err
		}
		lay.Textures = append(lay.Textures, f)
		lay.Sizes[f] = [2]int{img.Bounds().Dx(), img.Bounds().Dy()}
		for name, ff := range bitmapFonts {
			if !strings.EqualFold(ff, f) {
				continue
			}
			bf, err := ui.ParseBitmapFont(img)
			if err != nil {
				failedTex = append(failedTex, fmt.Sprintf("%s (%v)", f, err))
				continue
			}
			if len(bf.Glyphs) < 223 {
				// Not the usual sheet layout; the browser falls back to another font.
				odd = append(odd, fmt.Sprintf("%s: %d glyphs in %s, left out", name, len(bf.Glyphs), f))
				continue
			}
			lay.Bitmap[name] = &BitmapEntry{File: f, BitmapFont: bf}
		}
	}
	var missingFonts []string
	for _, f := range sortedKeys(ttf) {
		b, err := readInsensitive(uiDir, f)
		if err != nil {
			missingFonts = append(missingFonts, f)
			continue
		}
		dst := filepath.Join(skinOut, strings.ToLower(f))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return err
		}
		lay.Fonts = append(lay.Fonts, strings.ToLower(f))
	}
	if lay.Icons, err = convertIcons(game, skinOut); err != nil {
		return err
	}
	if err := convertArt(game, skinOut, &lay); err != nil {
		return err
	}

	b, err := json.Marshal(lay)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(skinOut, "layout.json"), b, 0o644); err != nil {
		return err
	}

	windows := 0
	for _, e := range s.Elements {
		if e.Tag == "WindowTemplate" {
			windows++
		}
	}
	fmt.Printf("skin %s: %d files, %d elements at top level, %d windows\n", skin, len(s.Files), len(s.Elements), windows)
	fmt.Printf("textures: %d written, %d missing, %d failed; fonts: %d TrueType, %d bitmap; icon sheets: %d\n",
		len(lay.Textures), len(missingTex), len(failedTex), len(lay.Fonts), len(lay.Bitmap), len(lay.Icons))
	report("missing includes", missingIncludes)
	report("missing textures", missingTex)
	report("failed textures", failedTex)
	report("missing fonts", missingFonts)
	report("unusual font sheets", odd)
	var unknown []string
	for _, t := range sortedKeys(tags) {
		if !knownTags[t] {
			unknown = append(unknown, fmt.Sprintf("%s x%d", t, tags[t]))
		}
	}
	report("unknown elements", unknown)
	fmt.Printf("wrote %s (%d KB)\n", filepath.Join(skinOut, "layout.json"), len(b)/1024)
	return nil
}

// convertArt writes the classic skin's window art (data/skin1/*.bmp: the paper
// doll, vault, merchant and so on) as PNG under <skinOut>/skin1 and records
// each file and its size. Windows the install has no layout file for, such as
// the inventory, are drawn over this art.
func convertArt(game, skinOut string, lay *Layout) error {
	dir := filepath.Join(game, "data", "skin1")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, e := range ents {
		low := strings.ToLower(e.Name())
		if !strings.HasSuffix(low, ".bmp") && !strings.HasSuffix(low, ".tga") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		img, err := ui.DecodeImage(low, b)
		if err != nil {
			continue // 16-bit or compressed sheets nothing here uses
		}
		key := "skin1/" + strings.ReplaceAll(low, " ", "")
		if err := writePNG(filepath.Join(skinOut, pngName(key)), img); err != nil {
			return err
		}
		lay.Textures = append(lay.Textures, key)
		lay.Sizes[key] = [2]int{img.Bounds().Dx(), img.Bounds().Dy()}
	}
	return nil
}

var iconName = regexp.MustCompile(`^(itm|wpn|spl|skl|cbt)_(\d+)\.bmp$`)

// convertIcons writes the install's icon sheets to <skinOut>/icons as PNG.
func convertIcons(game, skinOut string) ([]IconSheet, error) {
	dir := filepath.Join(game, "icons")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil // an install without icons still has a UI
	}
	var out []IconSheet
	for _, e := range ents {
		m := iconName.FindStringSubmatch(strings.ToLower(e.Name()))
		if m == nil {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		img, err := ui.DecodeBMP(b)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if r := img.Bounds(); r.Dx() != 320 || r.Dy() != 320 {
			return nil, fmt.Errorf("%s is %dx%d, want 320x320", e.Name(), r.Dx(), r.Dy())
		}
		name := strings.TrimSuffix(strings.ToLower(e.Name()), ".bmp") + ".png"
		if err := writePNG(filepath.Join(skinOut, "icons", name), img); err != nil {
			return nil, err
		}
		first, _ := strconv.Atoi(m[2])
		out = append(out, IconSheet{Kind: m[1], First: first, File: "icons/" + name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].First < out[j].First })
	return out, nil
}

func writePNG(dst string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	fh, err := os.Create(dst)
	if err != nil {
		return err
	}
	err = png.Encode(fh, img)
	if cerr := fh.Close(); err == nil {
		err = cerr
	}
	return err
}

func report(what string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Printf("%s:\n", what)
	for i, it := range items {
		if i == 12 {
			fmt.Printf("  ... and %d more\n", len(items)-12)
			break
		}
		fmt.Println("  " + it)
	}
}

func sortedKeys[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// pngName maps a texture path to the PNG the browser loads, lower-cased
// because the install's own references differ from its file names in case.
func pngName(f string) string {
	f = strings.ToLower(filepath.ToSlash(f))
	return f[:strings.LastIndex(f, ".")] + ".png"
}

// readInsensitive reads rel under dir, matching each path element without
// regard to case: the layouts say "MyriadPro-Regular.ttf" for a file that may
// be stored in lower case, which Windows forgives and other systems do not.
func readInsensitive(dir, rel string) ([]byte, error) {
	cur := dir
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		ents, err := os.ReadDir(cur)
		if err != nil {
			return nil, err
		}
		found := ""
		for _, e := range ents {
			if strings.EqualFold(e.Name(), part) {
				found = e.Name()
				break
			}
		}
		if found == "" {
			return nil, os.ErrNotExist
		}
		cur = filepath.Join(cur, found)
	}
	return os.ReadFile(cur)
}

// knownTags are the element names the converter and the web renderer have
// been checked against. Anything else is reported so it is not lost silently.
var knownTags = map[string]bool{}

func init() {
	for _, t := range strings.Fields(`
	Root_Element XML Include Name File Texture Canvas Font TTFFont Height Hint Antialiased
	WindowTemplate ButtonTemplate ImageAreaTemplate BackgroundTemplate BackgroundHResizeTemplate
	BackgroundVResizeTemplate ChatControlTemplate CheckBoxTemplate ComboBoxTemplate
	CompassControlTemplate ControlAreaFullResizeImageTemplate DownButtonTemplate EditBoxTemplate
	ForegroundHResizeTemplate FullResizeImageTemplate GutterVResizeTemplate HeaderBackgroundTemplate
	HeaderHorizResizeTemplate HorizontalResizeButtonTemplate HorizontalResizeImageTemplate
	HorizontalSliderTemplate IconSetTemplate IconTemplate LabelButtonTemplate ListBoxHeaderTemplate
	ListBoxTemplate MenuTemplate ScrollbarTemplate StatusBarTemplate StatusIconTemplate TabButtonTemplate
	TabsTemplate TextAreaTemplate ThumbVResizeTemplate TreeControlTemplate UpButtonTemplate
	VerticalResizeImageTemplate VerticalScrollbarTemplate VerticalSliderTemplate VerticalStatusBarTemplate
	LabelDef ButtonDef InvisibleButtonDef ImageAreaDef StatusIconDef CheckBoxDef ScalarLabelDef IconDef
	HorizontalResizeImageDef HorizontalResizeButtonDef HorizontalResizeImageButtonDef FullResizeImageDef
	ChatControlDef Position Alignment Size TopLeft TextureName TemplateName ControlId Width Height width height
	X Y R G B A Color ColorNormal ColorPressed ColorHighlit Data MaxCharacters FontName Adapter AdapterName
	ColorAdapter OnClickEvent Onclickevent Label TabId TabControl ToolTipID ImageOffset TextCentered
	EndAligned HideLabel IconID GrowWidth GrowHeight OffsetRight LabelWidth LabelAdapterName
	ImageAreaTemplateName HRButtonTemplateName CloseButton MoveButton TitleWidth TitleHeight
	ResizeableWidth ResizeableHeight ResizeableTwoWayWidth ResizeableTwoWayHeight WindowId MinWidth
	MinHeight ContextTemplateName TabName LinePadding TextOffset BufferName TopHeight MiddleHeight
	BottomHeight LeftWidth MiddleWidth RightWidth TopMiddle TopRight MiddleLeft MiddleMiddle MiddleRight
	BottomLeft BottomMiddle BottomRight`) {
		knownTags[t] = true
	}
}
