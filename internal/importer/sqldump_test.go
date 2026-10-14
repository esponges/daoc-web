package importer

import (
	"bufio"
	"strings"
	"testing"
)

const bs = string(rune(92)) // a backslash, as MySQL escapes quotes with

const sample = "/*!40101 SET NAMES utf8 */;\n" +
	"CREATE TABLE IF NOT EXISTS `t` (\n  `a` int,\n  `b` text\n);\n" +
	"REPLACE INTO `t` (`A`, `B`, `C`, `D`) VALUES\n" +
	"\t(1, 'plain', NULL, '50-58'),\n" +
	"\t(2, 'it''s, " + bs + "'quoted" + bs + "' (x)', 'multi\nline', -3.5),\n" +
	"\t(3, '', '', '');\n" +
	"INSERT INTO `u` (`X`) VALUES\n\t(7);\n"

func TestReadRows(t *testing.T) {
	var rows []*Row
	n, err := readRows(bufio.NewReader(strings.NewReader(sample)), func(r *Row) error {
		rows = append(rows, r)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 || len(rows) != 4 {
		t.Fatalf("rows = %d, want 4", n)
	}
	if rows[0].Int("A") != 1 || rows[0].Str("B") != "plain" || !rows[0].Nulls[2] {
		t.Errorf("row 0 = %+v", rows[0])
	}
	lo, hi := rows[0].Range("D")
	if lo != 50 || hi != 58 {
		t.Errorf("range = %d-%d", lo, hi)
	}
	if got := rows[1].Str("B"); got != "it's, 'quoted' (x)" {
		t.Errorf("quoting: %q", got)
	}
	if got := rows[1].Str("C"); got != "multi\nline" {
		t.Errorf("newline in string: %q", got)
	}
	if rows[1].Float("D") != -3.5 {
		t.Errorf("float = %v", rows[1].Float("D"))
	}
	if rows[3].Int("X") != 7 {
		t.Errorf("second statement: %+v", rows[3])
	}
	if rows[0].Has("Nope") || rows[0].Str("Nope") != "" {
		t.Error("absent column")
	}
}

func TestRangeNegativeAndPlain(t *testing.T) {
	for in, want := range map[string][2]int{"": {0, 0}, "-5": {-5, -5}, "7": {7, 7}, "10-11": {10, 11}, "abc": {0, 0}, "3.9": {3, 3}} {
		lo, hi := parseRange(in)
		if lo != want[0] || hi != want[1] {
			t.Errorf("parseRange(%q) = %d,%d want %v", in, lo, hi, want)
		}
	}
}

func TestBadRow(t *testing.T) {
	_, err := readRows(bufio.NewReader(strings.NewReader("REPLACE INTO `t` (`A`, `B`) VALUES\n\t(1);\n")), func(*Row) error { return nil })
	if err == nil {
		t.Error("a row with the wrong number of values must be an error")
	}
}
