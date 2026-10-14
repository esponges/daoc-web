// Package importer reads the community database's MySQL dumps and writes this
// project's own world store. It parses the INSERT/REPLACE statements itself,
// so no database server is needed, and it keeps no data of its own.
package importer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Row is one table row: column name to its text. SQL NULL is the empty string
// and is also recorded in Nulls.
type Row struct {
	Cols  []string
	Vals  []string
	Nulls []bool
	idx   map[string]int
}

// Str is the column's text, "" if the column is absent or NULL.
func (r *Row) Str(col string) string {
	if i, ok := r.idx[col]; ok {
		return r.Vals[i]
	}
	return ""
}

// Has reports whether the dump has the column at all.
func (r *Row) Has(col string) bool { _, ok := r.idx[col]; return ok }

// Int is the column as an integer; a range such as "50-58" gives its low end
// and anything unreadable gives 0.
func (r *Row) Int(col string) int {
	lo, _ := r.Range(col)
	return lo
}

// Float is the column as a float, 0 if unreadable.
func (r *Row) Float(col string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(r.Str(col)), 64)
	return f
}

// Range reads "50" or "50-58" (the NPC template tables store some numbers as
// a text range) and returns low and high.
func (r *Row) Range(col string) (int, int) { return parseRange(r.Str(col)) }

func parseRange(s string) (int, int) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0
	}
	if i := strings.Index(s[1:], "-"); i >= 0 {
		lo, e1 := strconv.Atoi(strings.TrimSpace(s[:i+1]))
		hi, e2 := strconv.Atoi(strings.TrimSpace(s[i+2:]))
		if e1 == nil && e2 == nil {
			return lo, hi
		}
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v, v
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(f), int(f)
	}
	return 0, 0
}

// ReadTable streams every row of a dump file to fn. It stops on the first
// error fn returns. Columns come from the INSERT's own column list, so a
// column the importer does not know is simply never asked for.
func ReadTable(path string, fn func(*Row) error) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return readRows(bufio.NewReaderSize(f, 1<<20), fn)
}

func readRows(br *bufio.Reader, fn func(*Row) error) (int, error) {
	n := 0
	for {
		_, err := br.Peek(1)
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
		// Statements start at the beginning of a line; read until one is an
		// INSERT or REPLACE, skipping DDL and comments.
		stmt, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return n, err
		}
		t := strings.TrimSpace(stmt)
		if !(strings.HasPrefix(t, "INSERT INTO") || strings.HasPrefix(t, "REPLACE INTO") ||
			strings.HasPrefix(t, "INSERT IGNORE INTO")) {
			if err == io.EOF {
				return n, nil
			}
			continue
		}
		a := strings.Index(t, "(")
		b := strings.Index(t, ") VALUES")
		if a < 0 || b < a {
			return n, fmt.Errorf("importer: unreadable column list: %.80s", t)
		}
		var cols []string
		for _, c := range strings.Split(t[a+1:b], ",") {
			cols = append(cols, strings.Trim(strings.TrimSpace(c), "`"))
		}
		idx := make(map[string]int, len(cols))
		for i, c := range cols {
			idx[c] = i
		}
		if strings.TrimSpace(t[b+len(") VALUES"):]) != "" {
			return n, fmt.Errorf("importer: rows on the VALUES line are not supported: %.80s", t)
		}
		cnt, perr := readTuples(br, cols, idx, fn)
		n += cnt
		if perr != nil {
			return n, perr
		}
	}
}

// readTuples parses "(v, v, ...)," groups until the statement's ';'.
func readTuples(r *bufio.Reader, cols []string, idx map[string]int, fn func(*Row) error) (int, error) {
	n := 0
	for {
		c, err := skipSpace(r)
		if err != nil {
			return n, err
		}
		switch c {
		case ';':
			return n, nil
		case ',':
			continue
		case '(':
		default:
			return n, fmt.Errorf("importer: unexpected %q between rows", c)
		}
		row := &Row{Cols: cols, idx: idx}
		for {
			v, null, end, err := readValue(r)
			if err != nil {
				return n, err
			}
			row.Vals = append(row.Vals, v)
			row.Nulls = append(row.Nulls, null)
			if end {
				break
			}
		}
		if len(row.Vals) != len(cols) {
			return n, fmt.Errorf("importer: row has %d values, columns %d", len(row.Vals), len(cols))
		}
		n++
		if err := fn(row); err != nil {
			return n, err
		}
	}
}

func skipSpace(r *bufio.Reader) (byte, error) {
	for {
		c, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return c, nil
		}
	}
}

// readValue reads one value and the ',' or ')' after it.
func readValue(r *bufio.Reader) (val string, null, end bool, err error) {
	c, err := skipSpace(r)
	if err != nil {
		return
	}
	var sb strings.Builder
	if c == '\'' {
		for {
			c, err = r.ReadByte()
			if err != nil {
				return
			}
			switch c {
			case '\\':
				e, err2 := r.ReadByte()
				if err2 != nil {
					err = err2
					return
				}
				switch e {
				case 'n':
					sb.WriteByte('\n')
				case 'r':
					sb.WriteByte('\r')
				case 't':
					sb.WriteByte('\t')
				case '0':
					sb.WriteByte(0)
				default:
					sb.WriteByte(e)
				}
				continue
			case '\'':
				p, perr := r.Peek(1)
				if perr == nil && p[0] == '\'' {
					r.ReadByte()
					sb.WriteByte('\'')
					continue
				}
			default:
				sb.WriteByte(c)
				continue
			}
			break
		}
		c, err = skipSpace(r)
		if err != nil {
			return
		}
		return sb.String(), false, c == ')', nil
	}
	// A bare token: number, NULL, or an expression up to , or ).
	sb.WriteByte(c)
	for c != ',' && c != ')' {
		c, err = r.ReadByte()
		if err != nil {
			return
		}
		if c != ',' && c != ')' {
			sb.WriteByte(c)
		}
	}
	s := strings.TrimSpace(sb.String())
	if strings.EqualFold(s, "NULL") {
		return "", true, c == ')', nil
	}
	return s, false, c == ')', nil
}
