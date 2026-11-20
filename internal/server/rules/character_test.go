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

func TestThaneAndSkald(t *testing.T) {
	thane, skald := Classes[ClassThane], Classes[ClassSkald]
	vector(t, cs, "Thane definition", func(t *testing.T) {
		if thane.ID != 21 || thane.Primary != STR || thane.Secondary != PIE || thane.Tertiary != CON ||
			thane.BaseHP != 720 || thane.WeaponSkillBase != 360 || thane.SpecMultiplier != 20 ||
			thane.Base != "Viking" || !thane.Casts || thane.Acuity != PIE {
			t.Errorf("Thane is %+v", thane)
		}
		for race, want := range map[int]bool{5: true, 6: true, 7: true, 8: false, 14: false} {
			if thane.MayBe(race) != want {
				t.Errorf("race %d may be a Thane: %v, want %v", race, thane.MayBe(race), want)
			}
		}
	})
	vector(t, cs, "Skald definition", func(t *testing.T) {
		if skald.ID != 24 || skald.Primary != CHA || skald.Secondary != STR || skald.Tertiary != CON ||
			skald.BaseHP != 760 || skald.WeaponSkillBase != 380 || skald.SpecMultiplier != 15 ||
			skald.Base != "Viking" || !skald.Casts || skald.Acuity != CHA {
			t.Errorf("Skald is %+v", skald)
		}
		for race, want := range map[int]bool{5: true, 6: true, 7: true, 8: true, 14: false} {
			if skald.MayBe(race) != want {
				t.Errorf("race %d may be a Skald: %v, want %v", race, skald.MayBe(race), want)
			}
		}
	})
	vector(t, cs, "Levels raise the class stats", func(t *testing.T) {
		// Level 6 gives a Skald one of each of charisma, strength and constitution;
		// level 7 only charisma.
		if g := LevelGains(skald, 6); g[CHA] != 1 || g[STR] != 1 || g[CON] != 1 || g[PIE] != 0 {
			t.Errorf("Skald gains at 6: %v", g)
		}
		a, b := LevelGains(skald, 6), LevelGains(skald, 7)
		if b[CHA]-a[CHA] != 1 || b[STR] != a[STR] || b[CON] != a[CON] {
			t.Errorf("level 6 %v, level 7 %v", a, b)
		}
	})
	t.Run("a Thane's pool rises with piety, a Skald's with charisma", func(t *testing.T) {
		if MaxPower(thane, 10, 80) <= MaxPower(thane, 10, 60) || MaxPower(skald, 10, 80) <= MaxPower(skald, 10, 60) {
			t.Error("the pool does not grow with the casting stat")
		}
	})
}

func TestSpiritmasterAndBonedancer(t *testing.T) {
	sm, bd := Classes[ClassSpiritmaster], Classes[ClassBonedancer]
	vector(t, cs, "Spiritmaster and Bonedancer numbers", func(t *testing.T) {
		for _, c := range []Class{sm, bd} {
			if c.Primary != PIE || c.Secondary != DEX || c.Tertiary != QUI ||
				c.BaseHP != 560 || c.WeaponSkillBase != 280 || c.SpecMultiplier != 10 || c.Base != "Mystic" {
				t.Errorf("%s is %+v", c.Name, c)
			}
		}
		if sm.ID != 27 || bd.ID != 30 {
			t.Errorf("class numbers %d and %d, want 27 and 30", sm.ID, bd.ID)
		}
		// The data's start table: Norseman and Kobold Spiritmasters; Troll,
		// Kobold and Valkyn Bonedancers.
		for race, want := range map[int]bool{5: true, 6: false, 7: false, 8: true, 14: false} {
			if sm.MayBe(race) != want {
				t.Errorf("race %d may be a Spiritmaster: %v, want %v", race, sm.MayBe(race), want)
			}
		}
		for race, want := range map[int]bool{5: false, 6: true, 7: false, 8: true, 14: true} {
			if bd.MayBe(race) != want {
				t.Errorf("race %d may be a Bonedancer: %v, want %v", race, bd.MayBe(race), want)
			}
		}
	})
	vector(t, cs, "Casting stat", func(t *testing.T) {
		for _, c := range []Class{sm, bd} {
			if !c.Casts || c.Acuity != PIE {
				t.Errorf("%s casts from %v, want piety", c.Name, c.Acuity)
			}
			if MaxPower(c, 10, 80) <= MaxPower(c, 10, 60) {
				t.Errorf("%s's pool does not grow with piety", c.Name)
			}
		}
	})
}

func TestShadowblade(t *testing.T) {
	s := Classes[ClassShadowblade]
	vector(t, cs, "Shadowblade numbers", func(t *testing.T) {
		if s.ID != 23 || s.Primary != DEX || s.Secondary != QUI || s.Tertiary != STR || s.BaseHP != 760 ||
			s.WeaponSkillBase != 360 || s.SpecMultiplier != 22 || s.Base != "Rogue" || s.Casts {
			t.Errorf("Shadowblade is %+v", s)
		}
		for race, want := range map[int]bool{5: true, 6: false, 7: false, 8: true, 14: true} {
			if s.MayBe(race) != want {
				t.Errorf("race %d may be a Shadowblade: %v, want %v", race, s.MayBe(race), want)
			}
		}
	})
}

func TestHunter(t *testing.T) {
	h := Classes[ClassHunter]
	vector(t, cs, "Hunter numbers", func(t *testing.T) {
		if h.ID != 25 || h.Primary != DEX || h.Secondary != QUI || h.Tertiary != STR || h.BaseHP != 720 ||
			h.WeaponSkillBase != 380 || h.SpecMultiplier != 20 || h.Base != "Rogue" || !h.Casts || h.Acuity != DEX {
			t.Errorf("Hunter is %+v", h)
		}
		for race, want := range map[int]bool{5: true, 6: false, 7: true, 8: true, 14: true} {
			if h.MayBe(race) != want {
				t.Errorf("race %d may be a Hunter: %v, want %v", race, h.MayBe(race), want)
			}
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
		// Berserker, Savage, Shaman, Thane and Skald, and not the Healer or Runemaster.
		var got []string
		for _, id := range []int{ClassWarrior, ClassBerserker, ClassSavage, ClassHealer, ClassShaman, ClassRunemaster, ClassThane, ClassSkald} {
			if Classes[id].MayBe(6) {
				got = append(got, Classes[id].Name)
			}
		}
		if want := "Warrior Berserker Savage Shaman Thane Skald"; strings.Join(got, " ") != want {
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

const rp = "rules-power"

func TestPower(t *testing.T) {
	vector(t, rp, "A non-caster", func(t *testing.T) {
		if p := MaxPower(Classes[ClassWarrior], 30, 120); p != 0 {
			t.Errorf("a Warrior has %d power", p)
		}
	})
	vector(t, rp, "Levelling", func(t *testing.T) {
		healer := Classes[ClassHealer]
		a, b := MaxPower(healer, 1, 70), MaxPower(healer, 2, 70)
		if a != 25 || b != 30 {
			t.Errorf("level 1 and 2 pools %d and %d, want 25 and 30", a, b)
		}
		if got := MaxPower(healer, 10, 150); got != 150 {
			t.Errorf("level 10 with 150 piety: %d, want 150", got)
		}
		if got := MaxPower(healer, 1, 10); got != 5 {
			t.Errorf("the least pool is %d, want 5", got)
		}
	})
	vector(t, rp, "Resting", func(t *testing.T) {
		near(t, "level 1 tick", PowerRegen(1), 5.3636, 1e-3)
		near(t, "level 50 tick", PowerRegen(50), 23.1818, 1e-3)
	})
	vector(t, rp, "Interrupted cast", func(t *testing.T) {
		// Cost is taken only when a spell lands, which the world tests check; here, the cost
		// of a spell that names none is nothing.
		if c := SpellPowerCost(0, 90); c != 0 {
			t.Errorf("a spell with no power costs %d", c)
		}
	})
	vector(t, rp, "A fixed cost", func(t *testing.T) {
		if c := SpellPowerCost(3, 90); c != 3 {
			t.Errorf("a fixed 3 costs %d, want 3", c)
		}
	})
	vector(t, rp, "A share of the pool", func(t *testing.T) {
		if c := SpellPowerCost(-10, 90); c != 9 {
			t.Errorf("10%% of 90 is %d, want 9", c)
		}
		if c := SpellPowerCost(-1, 16); c != 1 {
			t.Errorf("the least share is %d, want 1", c)
		}
	})
}
