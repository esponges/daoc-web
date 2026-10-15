package ui

import "testing"

func TestParse(t *testing.T) {
	n, err := Parse([]byte("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>\n<Root_Element ID=\"x\">\n <WindowTemplate>\n  <Name>caf\xe9</Name>\n  <Width>350</Width>\n  <Position><X>5</X><Y>12</Y></Position>\n </WindowTemplate>\n <Empty/>\n</Root_Element>"))
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(n.Children))
	}
	w := n.Children[0]
	if w.Str("Name") != "café" {
		t.Errorf("Name = %q, want café", w.Str("Name"))
	}
	if w.Int("Width") != 350 || w.Child("Position").Int("Y") != 12 {
		t.Errorf("numbers read wrong: %+v", w)
	}
	if n.Children[1].Tag != "Empty" {
		t.Errorf("second child = %q", n.Children[1].Tag)
	}
}

func TestParseRejectsEmpty(t *testing.T) {
	if _, err := Parse([]byte("")); err == nil {
		t.Error("empty input parsed")
	}
}
