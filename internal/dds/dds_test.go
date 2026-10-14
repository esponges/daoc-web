package dds

import (
	"encoding/binary"
	"testing"
)

// A 2x1 RGB565 image: pure red and a dark grey, whose top bits are zero.
func TestDecode565(t *testing.T) {
	b := make([]byte, 128+4)
	copy(b, "DDS ")
	le := binary.LittleEndian
	le.PutUint32(b[12:], 1) // height
	le.PutUint32(b[16:], 2) // width
	le.PutUint32(b[80:], 0x40)
	le.PutUint32(b[88:], 16)
	le.PutUint32(b[92:], 0xF800)
	le.PutUint32(b[96:], 0x07E0)
	le.PutUint32(b[100:], 0x001F)
	le.PutUint16(b[128:], 0xF800)               // red
	le.PutUint16(b[130:], uint16(2<<11|4<<5|2)) // dark grey
	img, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if p := img.Pix[0:4]; p[0] != 255 || p[1] != 0 || p[2] != 0 || p[3] != 255 {
		t.Errorf("red decodes as %v", p)
	}
	// 2 of 31 is 16 of 255 when the five bits are repeated (00010 00010 -> 00010000).
	if p := img.Pix[4:8]; p[0] != 16 || p[1] != 16 || p[2] != 16 || p[3] != 255 {
		t.Errorf("dark grey decodes as %v", p)
	}
}
