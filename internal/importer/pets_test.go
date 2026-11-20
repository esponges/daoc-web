package importer

import (
	"os"
	"path/filepath"
	"testing"
)

// Every Midgard summon spell in the community data names a template that is
// in the template table, and the import keeps the extra template columns.
func TestSummonSpellTemplatesResolve(t *testing.T) {
	d := filepath.Join("..", "..", "server", "data", "community", "opendaoc-db-core")
	if _, err := os.Stat(filepath.Join(d, "spell.sql")); err != nil {
		t.Skip("no community database checked out in server/data/community")
	}
	templates := map[int]bool{}
	if _, err := ReadTable(filepath.Join(d, "npctemplate.sql"), func(r *Row) error {
		templates[r.Int("TemplateId")] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	pkgs := map[string]bool{"Bonedancer": true, "Spiritmaster": true, "Summoning": true, "Beastcraft": true}
	seen, missing := 0, 0
	if _, err := ReadTable(filepath.Join(d, "spell.sql"), func(r *Row) error {
		switch r.Str("Type") {
		case "SummonMinion", "SummonCommander", "SummonSpiritFighter", "SummonHunterPet":
		default:
			return nil
		}
		if !pkgs[r.Str("PackageID")] {
			return nil
		}
		seen++
		if !templates[r.Int("LifeDrainReturn")] {
			missing++
			t.Errorf("%s (spell %d) names template %d, which is not in the data", r.Str("Name"), r.Int("SpellID"), r.Int("LifeDrainReturn"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if seen < 30 {
		t.Errorf("only %d summon spells of the Midgard packages found", seen)
	}
	_ = missing
}
