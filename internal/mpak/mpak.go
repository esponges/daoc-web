// Package mpak reads Mythic package archives (.mpk / .npk) as shipped with
// Dark Age of Camelot.
//
// Layout, as determined by inspecting the shipped archives:
//
//	0x00  char[4]    "MPAK"
//	0x04  uint8      version (observed: 2)
//	0x05  byte[16]   header, obfuscated: real[i] = stored[i] ^ i
//	                   +0  uint32  checksum
//	                   +4  uint32  compressed size of the directory
//	                   +8  uint32  compressed size of the archive-name stream
//	                   +12 uint32  file count
//	0x15             zlib stream: the archive's own file name
//	0x15+nameSize    zlib stream: directory, count * 284 bytes
//	then             file data; one zlib stream per entry
//
// Each 284-byte directory entry is a 256-byte NUL-terminated name followed by
// seven uint32s. The name field is a reused buffer, so bytes after the
// terminator contain unrelated leftovers and must be ignored.
package mpak

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	magic      = "MPAK"
	headerSize = 21
	entrySize  = 284
	nameField  = 256
)

// Entry describes one file stored in an Archive.
type Entry struct {
	Name     string
	ModTime  uint32
	Flags    uint32
	UncOff   uint32
	UncSize  uint32
	CompOff  uint32
	CompSize uint32
	CRC      uint32
}

// Archive is an opened Mythic package.
type Archive struct {
	Name    string
	Entries []Entry
	data    []byte // file-data region; Entry.CompOff is relative to this
}

// Open reads and parses the archive at path.
func Open(path string) (*Archive, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	a, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return a, nil
}

// Parse parses an archive already held in memory.
func Parse(raw []byte) (*Archive, error) {
	if len(raw) < headerSize {
		return nil, errors.New("too short to be an MPAK")
	}
	if string(raw[:4]) != magic {
		return nil, fmt.Errorf("bad magic %q, want %q", raw[:4], magic)
	}

	// The 16 bytes at 0x05 are XORed with their own index.
	var hdr [16]byte
	for i := range hdr {
		hdr[i] = raw[5+i] ^ byte(i)
	}
	dirSize := binary.LittleEndian.Uint32(hdr[4:])
	nameSize := binary.LittleEndian.Uint32(hdr[8:])
	count := binary.LittleEndian.Uint32(hdr[12:])

	nameOff := uint32(headerSize)
	dirOff := nameOff + nameSize
	dataOff := dirOff + dirSize
	if uint64(dataOff) > uint64(len(raw)) {
		return nil, fmt.Errorf("directory extends past end of file (dataOff=%d, size=%d)", dataOff, len(raw))
	}

	name, err := inflate(raw[nameOff:])
	if err != nil {
		return nil, fmt.Errorf("archive name stream: %w", err)
	}
	dir, err := inflate(raw[dirOff:])
	if err != nil {
		return nil, fmt.Errorf("directory stream: %w", err)
	}
	if want := int(count) * entrySize; len(dir) < want {
		return nil, fmt.Errorf("directory is %d bytes, want %d for %d entries", len(dir), want, count)
	}

	a := &Archive{
		Name:    string(name),
		Entries: make([]Entry, 0, count),
		data:    raw[dataOff:],
	}
	for i := uint32(0); i < count; i++ {
		e := dir[i*entrySize : (i+1)*entrySize]
		nm := e[:nameField]
		if z := bytes.IndexByte(nm, 0); z >= 0 {
			nm = nm[:z]
		}
		u := e[nameField:]
		a.Entries = append(a.Entries, Entry{
			Name:     string(nm),
			ModTime:  binary.LittleEndian.Uint32(u[0:]),
			Flags:    binary.LittleEndian.Uint32(u[4:]),
			UncOff:   binary.LittleEndian.Uint32(u[8:]),
			UncSize:  binary.LittleEndian.Uint32(u[12:]),
			CompOff:  binary.LittleEndian.Uint32(u[16:]),
			CompSize: binary.LittleEndian.Uint32(u[20:]),
			CRC:      binary.LittleEndian.Uint32(u[24:]),
		})
	}
	return a, nil
}

// ReadEntry decompresses a single entry.
func (a *Archive) ReadEntry(e Entry) ([]byte, error) {
	lo, hi := uint64(e.CompOff), uint64(e.CompOff)+uint64(e.CompSize)
	if hi > uint64(len(a.data)) {
		return nil, fmt.Errorf("%s: data range [%d,%d) past end of archive", e.Name, lo, hi)
	}
	out, err := inflate(a.data[lo:hi])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", e.Name, err)
	}
	if uint32(len(out)) != e.UncSize {
		return nil, fmt.Errorf("%s: inflated to %d bytes, directory says %d", e.Name, len(out), e.UncSize)
	}
	return out, nil
}

// Read decompresses the named entry. The match is case-insensitive, because the
// archives are inconsistent about case (SECTOR.DAT vs terrain.pcx).
func (a *Archive) Read(name string) ([]byte, error) {
	for _, e := range a.Entries {
		if strings.EqualFold(e.Name, name) {
			return a.ReadEntry(e)
		}
	}
	return nil, fmt.Errorf("%q not found in %s", name, a.Name)
}

// Has reports whether the named entry exists.
func (a *Archive) Has(name string) bool {
	for _, e := range a.Entries {
		if strings.EqualFold(e.Name, name) {
			return true
		}
	}
	return false
}

// inflate decompresses one zlib stream from the front of b, ignoring whatever
// follows it. Archives pack streams back to back, so trailing bytes are normal.
func inflate(b []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	out, err := io.ReadAll(zr)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	return out, nil
}
