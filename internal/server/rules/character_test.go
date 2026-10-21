package rules

import (
	"strings"
	"testing"
)

const cs = "rules-character-stats"

var warrior = Classes[ClassWarrior]

// trollWarrior is a Troll Warrior with the default build at level 1.
func trollWarrior(t *testing.T) Stats {
	t.Helper()
	s, err := Create(Races[6], warrior, DefaultBuild(warrior))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCharacterStats(t *testing.T) {
	vector(t, cs, "The Midgard races", func(t *testing.T) {
		want := map[string]Stats{
			"Norseman": {70, 70, 50, 50, 60, 60, 60, 60},
			"Troll":    {100, 70, 35, 35, 60, 60, 60, 60},
			"Dwarf":    {60, 80, 50, 50, 60, 60, 60, 60},
			"Kobold":   {50, 50, 70, 70, 60, 60, 60, 60},
			"Valkyn":   {55, 45, 65, 75, 60, 60, 60, 60},
		}
		if len(Races) != len(want) {
			t.Fatalf("%d races, want %d", len(Races), len(want))
		}
		for _, r := range Races {
			if r.Base != want[r.Name] {
				t.Errorf("%s base %v, want %v", r.Name, r.Base, want[r.Name])
			}
		}
	})
	vector(t, cs, "Warrior definition", func(t *testing.T) {
		w := Classes[22]
		if w.Name != "Warrior" || w.Primary != STR || w.Secondary != CON || w.Tertiary != DEX ||
			w.BaseHP != 880 || w.WeaponSkillBase != 460 || w.SpecMultiplier != 20 {
			t.Errorf("warrior = %+v", w)
		}
		if !w.MayBe(8) {
			t.Error("a Kobold may be a Warrior")
		}
	})
	vector(t, cs, "Cost of a stat", func(t *testing.T) {
		if c := RaiseCost(18); c != 29 {
			t.Errorf("raising by 18 costs %d, want 29", c)
		}
	})
	vector(t, cs, "Default build", func(t *testing.T) {
		b := DefaultBuild(warrior)
		if b != (Stats{10, 10, 10}) {
			t.Errorf("default build %v", b)
		}
		if s := trollWarrior(t); s != (Stats{110, 80, 45, 35, 60, 60, 60, 60}) {
			t.Errorf("troll warrior %v", s)
		}
	})
	vector(t, cs, "Overspent build", func(t *testing.T) {
		if _, err := Create(Races[6], warrior, Stats{11, 10, 10}); err == nil {
			t.Error("a 31-point build was accepted")
		}
		if _, err := Create(Races[6], warrior, Stats{20, 10, 10, -10}); err == nil {
			t.Error("a build lowering a stat was accepted")
		}
	})
	vector(t, cs, "Troll Warrior at 50", func(t *testing.T) {
		m := Modified(trollWarrior(t), warrior, 50, Stats{}, 0)
		if m[STR] != 155 || m[CON] != 103 || m[DEX] != 60 || m[QUI] != 35 {
			t.Errorf("level 50 stats %v", m)
		}
	})
	vector(t, cs, "Level 7", func(t *testing.T) {
		a, b := LevelGains(warrior, 6), LevelGains(warrior, 7)
		if b[STR]-a[STR] != 1 || b[CON] != a[CON] || b[DEX] != a[DEX] {
			t.Errorf("level 6 %v, level 7 %v", a, b)
		}
	})
	vector(t, cs, "Over the cap", func(t *testing.T) {
		base := trollWarrior(t)
		m := Modified(base, warrior, 10, Stats{STR: 20}, 0)
		if got := m[STR] - Modified(base, warrior, 10, Stats{}, 0)[STR]; got != 15 {
			t.Errorf("+20 strength counts %d at level 10, want 15", got)
		}
	})
	vector(t, cs, "Level 1 Troll Warrior", func(t *testing.T) {
		if hp := MaxHP(warrior, 1, 80, 0); hp != 39 {
			t.Errorf("hp %d, want 39", hp)
		}
	})
	vector(t, cs, "Level 50 Troll Warrior", func(t *testing.T) {
		if hp := MaxHP(warrior, 50, 103, 0); hp != 1133 {
			t.Errorf("hp %d, want 1133", hp)
		}
	})
	vector(t, cs, "Low constitution", func(t *testing.T) {
		if hp := MaxHP(warrior, 50, 45, 0); hp != 856 {
			t.Errorf("hp %d, want 856", hp)
		}
	})
	vector(t, cs, "Item hit points", func(t *testing.T) {
		if d := MaxHP(warrior, 10, 83, 60) - MaxHP(warrior, 10, 83, 0); d != 40 {
			t.Errorf("+60 item hit points count %d at level 10, want 40", d)
		}
	})
	vector(t, cs, "After a style chain", func(t *testing.T) {
		if MaxEndurance != 100 || 40+10*EnduranceRegenOut < MaxEndurance {
			t.Error("40 endurance is not back to 100 within ten seconds out of combat")
		}
	})
	vector(t, cs, "Level 20 resting", func(t *testing.T) {
		if got := 10 * HealthRegen(20, false); got != 70 {
			t.Errorf("30 s of rest returns %d, want 70", got)
		}
		if HealthRegen(20, true) != 0 {
			t.Error("health returns in combat")
		}
	})
}

const mc = "midgard-classes"

func TestMidgardClasses(t *testing.T) {
	vector(t, mc, "A race that may not", func(t *testing.T) {
		healer := Classes[ClassHealer]
		if _, err := Create(Races[8], healer, DefaultBuild(healer)); err == nil {
			t.Error("a Kobold Healer was created")
		}
		if _, err := Create(Races[5], healer, DefaultBuild(healer)); err != nil {
			t.Errorf("a Norse Healer was refused: %v", err)
		}
	})
	vector(t, mc, "Class list", func(t *testing.T) {
		// The classes a Troll may be: the data's start table gives the Warrior,
		// Berserker, Savage and Shaman, and not the Healer or Runemaster.
		var got []string
		for _, id := range []int{ClassWarrior, ClassBerserker, ClassSavage, ClassHealer, ClassShaman, ClassRunemaster} {
			if Classes[id].MayBe(6) {
				got = append(got, Classes[id].Name)
			}
		}
		if want := "Warrior Berserker Savage Shaman"; strings.Join(got, " ") != want {
			t.Errorf("a Troll may be %v, want %s", got, want)
		}
	})
	t.Run("every class builds a legal default", func(t *testing.T) {
		for id, c := range Classes {
			for _, r := range c.Races {
				if _, err := Create(Races[r], c, DefaultBuild(c)); err != nil {
					t.Errorf("%s of race %d: %v", c.Name, r, err)
				}
			}
			if id != c.ID {
				t.Errorf("%s is keyed %d but numbered %d", c.Name, id, c.ID)
			}
		}
	})
}
