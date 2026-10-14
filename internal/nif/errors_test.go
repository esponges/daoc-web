package nif_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strings"
	"testing"

	"daocweb/internal/nif"
)

// file4 builds a NetImmerse 4.2.2.0 file: the header, then each block as
// its type name and body. No footer is added; the tests fail before it.
func file4(blocks ...[]byte) []byte {
	b := &bytes.Buffer{}
	b.WriteString("NetImmerse File Format, Version 4.2.2.0\n")
	put(b, uint32(0x04020200), uint32(len(blocks)))
	for _, blk := range blocks {
		b.Write(blk)
	}
	return b.Bytes()
}

func put(b *bytes.Buffer, vals ...any) {
	for _, v := range vals {
		binary.Write(b, binary.LittleEndian, v)
	}
}

func str(b *bytes.Buffer, s string) {
	put(b, uint32(len(s)))
	b.WriteString(s)
}

func parseErr(t *testing.T, raw []byte) *nif.Error {
	t.Helper()
	_, err := nif.Parse(raw)
	if err == nil {
		t.Fatal("parsed; want an error")
	}
	var ne *nif.Error
	if !errors.As(err, &ne) {
		t.Fatalf("error %T %q is not a *nif.Error", err, err)
	}
	return ne
}

// A block type with no reader stops the parse, naming the type and where it
// sits.
func TestUnsupportedBlockIsNamed(t *testing.T) {
	blk := &bytes.Buffer{}
	str(blk, "NiImaginaryBlock")
	put(blk, uint32(0), uint32(0))
	e := parseErr(t, file4(blk.Bytes()))
	if e.Kind != nif.Unsupported || e.Type != "NiImaginaryBlock" || e.Block != 0 {
		t.Errorf("kind %v, type %q, block %d; want unsupported NiImaginaryBlock, block 0", e.Kind, e.Type, e.Block)
	}
	if !strings.Contains(e.Error(), `unmodelled block type "NiImaginaryBlock"`) {
		t.Errorf("message %q does not name the block", e.Error())
	}
	if e.Offset <= 0 {
		t.Errorf("offset %d, want the block's position", e.Offset)
	}
}

// A NaN where a material's colour should be is corruption, not an
// unsupported type.
func TestNonFiniteFloatIsCorrupt(t *testing.T) {
	blk := &bytes.Buffer{}
	str(blk, "NiMaterialProperty")
	str(blk, "mat")                // name
	put(blk, int32(-1), int32(-1)) // extra data, controller
	put(blk, uint16(0))            // flags
	put(blk, float32(math.NaN()))  // ambient red
	put(blk, make([]float32, 13))  // the rest, never reached
	e := parseErr(t, file4(blk.Bytes()))
	if e.Kind != nif.Corrupt || e.Type != "NiMaterialProperty" {
		t.Errorf("kind %v, type %q; want corrupt data in NiMaterialProperty", e.Kind, e.Type)
	}
	if !strings.Contains(e.Error(), "non-finite float") {
		t.Errorf("message %q does not say what was wrong", e.Error())
	}
}

// Bytes left after the footer mean a block was read with the wrong layout.
func TestLeftoverBytesAreAFooterMismatch(t *testing.T) {
	raw := file4()
	raw = append(raw, 0, 0, 0, 0) // no roots
	raw = append(raw, 0xAB)       // and one byte too many
	e := parseErr(t, raw)
	if e.Kind != nif.FooterMismatch || e.Block != -1 {
		t.Errorf("kind %v, block %d; want a footer mismatch outside any block", e.Kind, e.Block)
	}
}

// A file that stops mid-block overran, which is distinct from bad data.
func TestTruncatedIsOverrun(t *testing.T) {
	blk := &bytes.Buffer{}
	str(blk, "NiMaterialProperty")
	str(blk, "mat")
	e := parseErr(t, file4(blk.Bytes()))
	if e.Kind != nif.Overrun {
		t.Errorf("kind %v, want read past end", e.Kind)
	}
}

// shapeData4 is a one-triangle NiTriShapeData at 4.2.2.0 whose first
// vertex x and first texture coordinate u are given.
func shapeData4(x, u float32) []byte {
	blk := &bytes.Buffer{}
	str(blk, "NiTriShapeData")
	put(blk, uint16(3), uint8(1)) // vertices, present
	put(blk, x, float32(0), float32(0), float32(1), float32(0), float32(0), float32(0), float32(1), float32(0))
	put(blk, uint8(0))                                       // no normals
	put(blk, float32(0), float32(0), float32(0), float32(1)) // centre, radius
	put(blk, uint8(0))                                       // no colours
	put(blk, uint16(1))                                      // one UV set
	put(blk, u, float32(0), float32(1), float32(0), float32(0), float32(1))
	put(blk, uint16(1), uint32(3), uint16(0), uint16(1), uint16(2)) // one triangle
	put(blk, uint16(0))                                             // no match groups
	raw := file4(blk.Bytes())
	return append(raw, 1, 0, 0, 0, 0, 0, 0, 0) // one root, block 0
}

// A NaN in a texture coordinate is junk the city blocks carry and is read
// as 0; a NaN in a vertex position is still corruption.
func TestNonFiniteUVIsZero(t *testing.T) {
	f, err := nif.Parse(shapeData4(0, float32(math.NaN())))
	if err != nil {
		t.Fatalf("NaN texture coordinate: %v", err)
	}
	d := f.Blocks[0].(*nif.ShapeData)
	if d.UV[0][0][0] != 0 || d.UV[0][1][0] != 1 {
		t.Errorf("UVs %v, want the NaN read as 0 and the rest kept", d.UV[0])
	}
	e := parseErr(t, shapeData4(float32(math.NaN()), 0))
	if e.Kind != nif.Corrupt || e.Type != "NiTriShapeData" {
		t.Errorf("NaN position: kind %v, type %q; want corrupt data in NiTriShapeData", e.Kind, e.Type)
	}
}
