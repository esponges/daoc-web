package rules

import "testing"

func TestValidAppearance(t *testing.T) {
	for _, c := range []struct {
		name             string
		race, face, hair int
		ok               bool
	}{
		{"defaults", 5, 0, 0, true},
		{"the last face", 5, 6, 10, true},
		{"a seventh face", 5, 7, 1, false},
		{"a negative face", 5, -1, 1, false},
		{"a troll's last hair", 6, 1, 4, true},
		{"a troll cannot have a fifth colour", 6, 1, 5, false},
		{"an eleventh human colour", 7, 1, 11, false},
	} {
		if err := ValidAppearance(c.race, c.face, c.hair); (err == nil) != c.ok {
			t.Errorf("%s: %v", c.name, err)
		}
	}
}
