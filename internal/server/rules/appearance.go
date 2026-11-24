package rules

import "fmt"

// Appearance: what a player chooses at creation besides race and gender.
// A face and a hair colour are numbers from 1; 0 is the race's default
// look. The pictures behind the numbers are the install's skins
// (charconv -appearance). Hair style is not offered: only one head shell
// per race is converted. Eyes are painted into the face skin, so there is
// no separate eye colour to choose. Status: provisional, the counts are
// what the install holds, not a published 1.65 list.

// Faces is how many faces every Midgard race and gender has.
const Faces = 6

// HairColours is how many hair colours a race may choose: the ten human
// colours, or the four the trolls have.
func HairColours(race int) int {
	if race == 6 { // Troll
		return 4
	}
	return 10
}

// ValidAppearance checks a creation choice; 0 leaves either at the default.
func ValidAppearance(race, face, hair int) error {
	if face < 0 || face > Faces {
		return fmt.Errorf("choose a face from 1 to %d", Faces)
	}
	if hair < 0 || hair > HairColours(race) {
		return fmt.Errorf("choose a hair colour from 1 to %d", HairColours(race))
	}
	return nil
}
