package nif

import (
	"errors"
	"fmt"
)

// Kind says why a file failed to read, so failures can be counted by cause
// rather than by message.
type Kind int

const (
	// Corrupt is malformed data: a non-finite float where geometry is
	// expected, an implausible count or string length, an index out of
	// range.
	Corrupt Kind = iota
	// Unsupported is content with no reader: a block type, or a format
	// version. Blocks carry no length, so nothing after one can be read.
	Unsupported
	// Overrun is a read past the end of the file.
	Overrun
	// FooterMismatch is a footer that does not end exactly at end-of-file:
	// some block above was read with the wrong layout.
	FooterMismatch
	// BadHeader is a file that is not a NIF at all.
	BadHeader
)

func (k Kind) String() string {
	switch k {
	case Corrupt:
		return "corrupt data"
	case Unsupported:
		return "unsupported"
	case Overrun:
		return "read past end"
	case FooterMismatch:
		return "footer mismatch"
	case BadHeader:
		return "bad header"
	}
	return fmt.Sprintf("kind %d", int(k))
}

// Error is every failure Parse returns. Its message is the detail alone; the
// block, type and offset are fields, so callers can group failures by Kind
// and Type without parsing text.
type Error struct {
	Kind   Kind
	Block  int    // index of the block being read, or -1 outside any block
	Type   string // that block's type, or the unsupported type itself
	Offset int    // byte offset of the block, or of the failure outside one
	Err    error  // the detail
}

func (e *Error) Error() string { return "nif: " + e.Err.Error() }

func (e *Error) Unwrap() error { return e.Err }

// kindError is a detail that knows its kind; detail errors without one are
// Corrupt.
type kindError struct {
	kind Kind
	msg  string
}

func (e *kindError) Error() string { return e.msg }

func errOf(k Kind, format string, args ...any) error {
	return &kindError{kind: k, msg: fmt.Sprintf(format, args...)}
}

// kindOf reports the kind a detail error carries, Corrupt by default.
func kindOf(err error) Kind {
	var ke *kindError
	if errors.As(err, &ke) {
		return ke.kind
	}
	return Corrupt
}

// blockError wraps a failure inside block i.
func blockError(i int, typ string, start int, err error) *Error {
	return &Error{
		Kind: kindOf(err), Block: i, Type: typ, Offset: start,
		Err: fmt.Errorf("block %d (%s) at %d: %w", i, typ, start, err),
	}
}

// fileError is a failure outside any block.
func fileError(k Kind, typ string, off int, format string, args ...any) *Error {
	return &Error{Kind: k, Block: -1, Type: typ, Offset: off, Err: fmt.Errorf(format, args...)}
}
