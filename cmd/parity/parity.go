package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"daocweb/internal/importer"
	"daocweb/internal/server/rules"
)

// Midgard is what the census measures: the realm's regions and its classes as
// of 1.65, with the base classes they grow from.
var (
	midgardRegions = map[int]bool{100: true, 101: true, 151: true}
	midgardClasses = map[int]string{
		21: "Thane", 22: "Warrior", 23: "Shadowblade", 24: "Skald", 25: "Hunter", 26: "Healer",
		27: "Spiritmaster", 28: "Shaman", 29: "Runemaster", 30: "Bonedancer", 31: "Berserker", 32: "Savage",
		35: "Viking", 36: "Mystic", 37: "Seer", 38: "Midgard Rogue",
	}
)

// Item is one thing the data asks for, how much it is used, and whether the
// server implements it.
type Item struct {
	Name        string   `json:"name"`
	Uses        int      `json:"uses"`
	Classes     []string `json:"classes,omitempty"`
	Implemented bool     `json:"implemented"`
}

// RuleRow is one registered rule.
type RuleRow struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Report is the census.
type Report struct {
	Regions        []int          `json:"regions"`
	CreatureSpells []Item         `json:"creatureSpellTypes"` // uses: placements x spells
	ClassSpells    []Item         `json:"classSpellTypes"`    // uses: spells
	StyleFeatures  []Item         `json:"styleFeatures"`      // uses: styles
	Abilities      []Item         `json:"abilities"`          // uses: grants
	CreatureKinds  []Item         `json:"creatureKinds"`      // uses: placements
	Content        []Item         `json:"content"`            // quests, doors, keeps, relics, battlegrounds
	Rules          []RuleRow      `json:"rules"`
	RulesByStatus  map[string]int `json:"rulesByStatus"`
	Placements     int            `json:"placements"`
}

// implementedSpellTypes are the spell types the server can cast. None yet.
var implementedSpellTypes = map[string]bool{}

// Run reads the dump and the registry.
func Run(dump string) (*Report, error) {
	read := func(table string, fn func(*importer.Row) error) error {
		_, err := importer.ReadTable(filepath.Join(dump, table+".sql"), fn)
		if err != nil {
			return fmt.Errorf("%s: %w", table, err)
		}
		return nil
	}
	rep := &Report{Regions: []int{100, 101, 151}, RulesByStatus: map[string]int{}}

	// Spells by id, and their types.
	spellType := map[string]string{}
	if err := read("spell", func(r *importer.Row) error {
		spellType[r.Str("SpellID")] = r.Str("Type")
		return nil
	}); err != nil {
		return nil, err
	}

	// Creatures: what they are, and the spells their templates cast.
	tplSpells := map[string][]string{}
	if err := read("npctemplate", func(r *importer.Row) error {
		if s := r.Str("Spells"); s != "" && s != "0" {
			tplSpells[r.Str("TemplateId")] = strings.Split(s, ";")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	kinds, creatureSpells := map[string]int{}, map[string]int{}
	if err := read("mob", func(r *importer.Row) error {
		if !midgardRegions[r.Int("Region")] {
			return nil
		}
		rep.Placements++
		kinds[r.Str("ClassType")]++
		for _, id := range tplSpells[r.Str("NPCTemplateID")] {
			if t := spellType[id]; t != "" {
				creatureSpells[t]++
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	for k, n := range creatureSpells {
		rep.CreatureSpells = append(rep.CreatureSpells, Item{Name: k, Uses: n, Implemented: implementedSpellTypes[k]})
	}
	for k, n := range kinds {
		rep.CreatureKinds = append(rep.CreatureKinds, Item{Name: k, Uses: n, Implemented: kindImplemented(k)})
	}

	// Classes: their specs and careers, the spell lines of those specs, and the
	// abilities they grant.
	classOf := map[string][]string{} // spec or career -> class names
	if err := read("classxspecialization", func(r *importer.Row) error {
		if name, ok := midgardClasses[r.Int("ClassID")]; ok {
			k := r.Str("SpecKeyName")
			classOf[k] = appendOnce(classOf[k], name)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	lineClasses := map[string][]string{}
	if err := read("spellline", func(r *importer.Row) error {
		if cs := classOf[r.Str("Spec")]; len(cs) > 0 {
			lineClasses[r.Str("KeyName")] = cs
		}
		return nil
	}); err != nil {
		return nil, err
	}
	classSpells, classSpellClasses := map[string]int{}, map[string][]string{}
	if err := read("linexspell", func(r *importer.Row) error {
		cs := lineClasses[r.Str("LineName")]
		t := spellType[r.Str("SpellID")]
		if len(cs) == 0 || t == "" {
			return nil
		}
		classSpells[t]++
		for _, c := range cs {
			classSpellClasses[t] = appendOnce(classSpellClasses[t], c)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	for k, n := range classSpells {
		rep.ClassSpells = append(rep.ClassSpells, Item{Name: k, Uses: n, Classes: sorted(classSpellClasses[k]), Implemented: implementedSpellTypes[k]})
	}
	abilities, abilityClasses := map[string]int{}, map[string][]string{}
	if err := read("specxability", func(r *importer.Row) error {
		cs := classOf[r.Str("Spec")]
		if len(cs) == 0 {
			return nil
		}
		if id := r.Int("ClassId"); id != 0 {
			name, ok := midgardClasses[id]
			if !ok {
				return nil
			}
			cs = []string{name}
		}
		k := r.Str("AbilityKey")
		abilities[k]++
		for _, c := range cs {
			abilityClasses[k] = appendOnce(abilityClasses[k], c)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	for k, n := range abilities {
		rep.Abilities = append(rep.Abilities, Item{Name: k, Uses: n, Classes: sorted(abilityClasses[k]), Implemented: rules.ActiveAbility(k)})
	}

	// Styles: which features Midgard's styles use.
	features := map[string]int{}
	procStyles := map[string]bool{}
	if err := read("stylexspell", func(r *importer.Row) error {
		procStyles[r.Str("ClassID")+"/"+r.Str("StyleID")] = true
		return nil
	}); err != nil {
		return nil, err
	}
	if err := read("style", func(r *importer.Row) error {
		if _, ok := midgardClasses[r.Int("ClassId")]; !ok {
			return nil
		}
		for _, f := range styleFeatures(r) {
			features[f]++
		}
		if procStyles[r.Str("ClassId")+"/"+r.Str("ID")] {
			features["proc spell"]++
		}
		return nil
	}); err != nil {
		return nil, err
	}
	for k, n := range features {
		rep.StyleFeatures = append(rep.StyleFeatures, Item{Name: k, Uses: n, Implemented: styleFeatureImplemented[k]})
	}

	// Content the world does not have yet.
	count := func(table string, keep func(*importer.Row) bool) (int, error) {
		n := 0
		err := read(table, func(r *importer.Row) error {
			if keep == nil || keep(r) {
				n++
			}
			return nil
		})
		return n, err
	}
	for _, c := range []struct {
		name, table string
		keep        func(*importer.Row) bool
	}{
		{"quests", "dataquest", func(r *importer.Row) bool { return midgardRegions[r.Int("StartRegionID")] }},
		{"keep doors", "door", nil},
		{"keeps", "keep", nil},
		{"relics", "relic", nil},
		{"battlegrounds", "battleground", nil},
	} {
		n, err := count(c.table, c.keep)
		if err != nil {
			return nil, err
		}
		rep.Content = append(rep.Content, Item{Name: c.name, Uses: n})
	}

	for _, r := range rules.Registry() {
		rep.Rules = append(rep.Rules, RuleRow{ID: r.ID, Status: r.Status.String()})
		rep.RulesByStatus[r.Status.String()]++
	}
	for _, l := range [][]Item{rep.CreatureSpells, rep.ClassSpells, rep.StyleFeatures, rep.Abilities, rep.CreatureKinds} {
		sortItems(l)
	}
	return rep, nil
}

// styleFeatures names what a style row asks for.
func styleFeatures(r *importer.Row) []string {
	var out []string
	switch r.Int("OpeningRequirementType") {
	case 1:
		switch r.Int("AttackResultRequirement") {
		case 3:
			out = append(out, "opening: after a parry")
		case 4:
			out = append(out, "opening: after a block")
		case 5:
			out = append(out, "opening: after an evade")
		default:
			out = append(out, "opening: after defence "+strconv.Itoa(r.Int("AttackResultRequirement")))
		}
	case 2:
		out = append(out, []string{"opening: from behind", "opening: from the side", "opening: from the front"}[min(2, max(0, r.Int("OpeningRequirementValue")))])
	default:
		switch r.Int("AttackResultRequirement") {
		case 0:
			out = append(out, "opening: anytime")
		case 1:
			out = append(out, "opening: after a miss")
		case 2:
			out = append(out, "opening: after a hit")
		case 7:
			out = append(out, "chain: after a style")
		default:
			out = append(out, "opening: after result "+strconv.Itoa(r.Int("AttackResultRequirement")))
		}
	}
	if r.Float("GrowthRate") > 0 {
		out = append(out, "growth damage")
	}
	if r.Int("BonusToHit") != 0 {
		out = append(out, "to-hit bonus")
	}
	if r.Int("BonusToDefense") != 0 {
		out = append(out, "defence bonus")
	}
	if r.Int("ArmorHitLocation") != 0 {
		out = append(out, "fixed hit location")
	}
	if r.Int("StealthRequirement") != 0 {
		out = append(out, "stealth opener")
	}
	return out
}

// styleFeatureImplemented is what the server does with styles. A fixed hit
// location only matters against players, which nothing fights yet.
var styleFeatureImplemented = map[string]bool{
	"opening: anytime": true, "opening: after a parry": true, "opening: after a block": true, "opening: after an evade": true,
	"opening: from behind": true, "opening: from the side": true, "opening: from the front": true,
	"opening: after a miss": true, "opening: after a hit": true, "chain: after a style": true,
	"growth damage": true, "to-hit bonus": true, "defence bonus": true,
}

// kindImplemented says whether the server gives a creature class type its
// behaviour: plain creatures, merchants, healers, and trainers of the classes
// it covers.
func kindImplemented(classType string) bool {
	if classType == "DOL.GS.GameNPC" || importer.IsMerchantClass(classType) || importer.IsHealerClass(classType) {
		return true
	}
	t := importer.TrainerClass(classType)
	for _, c := range rules.Classes {
		if t != "" && c.TaughtBy(t) {
			return true
		}
	}
	return false
}

func appendOnce(l []string, s string) []string {
	for _, x := range l {
		if x == s {
			return l
		}
	}
	return append(l, s)
}

func sorted(l []string) []string {
	out := append([]string(nil), l...)
	sort.Strings(out)
	return out
}

// sortItems puts the most used first, then by name.
func sortItems(l []Item) {
	sort.Slice(l, func(i, j int) bool {
		if l[i].Uses != l[j].Uses {
			return l[i].Uses > l[j].Uses
		}
		return l[i].Name < l[j].Name
	})
}
