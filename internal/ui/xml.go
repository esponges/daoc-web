// Package ui reads the game's own interface data: the XML window layouts,
// TGA and BMP textures and bitmap fonts under the install's ui/ folder.
package ui

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Node is one XML element. Layout files are a tree of small elements whose
// meaning depends on the tag, so the tree is kept generic and the web side
// reads the tags it knows. Children keep file order, which is draw order.
type Node struct {
	Tag      string  `json:"t"`
	Text     string  `json:"x,omitempty"`
	Children []*Node `json:"c,omitempty"`
}

// Child returns the first child with the tag, or nil.
func (n *Node) Child(tag string) *Node {
	if n == nil {
		return nil
	}
	for _, c := range n.Children {
		if c.Tag == tag {
			return c
		}
	}
	return nil
}

// Str returns the text of the first child with the tag.
func (n *Node) Str(tag string) string {
	if c := n.Child(tag); c != nil {
		return c.Text
	}
	return ""
}

// Int returns the integer text of the first child with the tag, or 0.
func (n *Node) Int(tag string) int {
	v, _ := strconv.Atoi(n.Str(tag))
	return v
}

// Parse reads one layout file. The files are ISO-8859-1 and the decoder has
// no built-in table for it, so bytes are widened to UTF-8 first.
func Parse(b []byte) (*Node, error) {
	b = latin1ToUTF8(b)
	d := xml.NewDecoder(bytes.NewReader(b))
	d.Strict = false
	d.CharsetReader = func(_ string, in io.Reader) (io.Reader, error) { return in, nil }
	var root *Node
	var stack []*Node
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ui: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			n := &Node{Tag: t.Name.Local}
			if len(stack) == 0 {
				root = n
			} else {
				p := stack[len(stack)-1]
				p.Children = append(p.Children, n)
			}
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) > 0 {
				if s := strings.TrimSpace(string(t)); s != "" {
					stack[len(stack)-1].Text += s
				}
			}
		}
	}
	if root == nil {
		return nil, fmt.Errorf("ui: no root element")
	}
	return root, nil
}

func latin1ToUTF8(b []byte) []byte {
	var out bytes.Buffer
	for _, c := range b {
		if c < 0x80 {
			out.WriteByte(c)
		} else {
			out.WriteRune(rune(c))
		}
	}
	// The declaration still says ISO-8859-1; the decoder is told not to care.
	return out.Bytes()
}

// Skin is every file of one skin merged: the elements of each included file
// in include order, which is how the client sees them.
type Skin struct {
	Root     string   // directory holding the skin's XML
	Files    []string // files read, in order
	Elements []*Node  // children of each file's root element, flattened
}

// LoadSkin reads every window file of a skin. The full list is the include
// list of <uiDir>/uimain.xml, whose names resolve inside the skin directory;
// the skin's own uimain.xml is a shorter list that also names the icon files.
// A missing include is returned in missing and does not stop the load.
func LoadSkin(uiDir, skin string) (*Skin, []string, error) {
	dir := filepath.Join(uiDir, skin)
	mainNode, err := readFile(filepath.Join(uiDir, "uimain.xml"))
	if err != nil {
		return nil, nil, err
	}
	skinMain, err := readFile(filepath.Join(dir, "uimain.xml"))
	if err != nil {
		return nil, nil, err
	}
	s := &Skin{Root: dir}
	var missing []string
	seen := map[string]bool{}
	var load func(name string)
	load = func(name string) {
		key := strings.ToLower(name)
		if seen[key] {
			return
		}
		seen[key] = true
		n, err := readFile(filepath.Join(dir, name))
		if err != nil {
			missing = append(missing, name)
			return
		}
		s.Files = append(s.Files, name)
		for _, c := range n.Children {
			if c.Tag == "Include" {
				load(c.Text)
				continue
			}
			s.Elements = append(s.Elements, c)
		}
	}
	// assets.xml names the fonts and textures every other file refers to.
	load("assets.xml")
	for _, m := range []*Node{mainNode, skinMain} {
		for _, c := range m.Children {
			if c.Tag == "Include" {
				load(c.Text)
			}
		}
	}
	return s, missing, nil
}

func readFile(p string) (*Node, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	n, err := Parse(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(p), err)
	}
	return n, nil
}
