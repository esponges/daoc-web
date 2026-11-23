package world

import (
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"daocweb/internal/server/rules"
)

// Every spell the Savagery line gives to level 50 can be cast, on the imported
// data: the Savage has a book, and the server does what each kind says.
func TestSavageCastsItsWholeLine(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	var bad []string
	total := 0
	for _, spec := range cat.TrainableSpecs(rules.ClassSavage) {
		for _, l := range cat.SpellsFor(rules.ClassSavage, spec) {
			sp := cat.Spells[l.Spell]
			if sp == nil || l.Level > 50 {
				continue
			}
			total++
			if !spellSupported(sp) {
				bad = append(bad, spec+" "+strconv.Itoa(l.Level)+" "+sp.Name+" ("+sp.Type+", "+sp.Target+")")
			}
		}
	}
	sort.Strings(bad)
	if total == 0 {
		t.Fatal("the Savage has no spells in the data")
	}
	for _, b := range bad {
		t.Errorf("the Savage cannot cast %s", b)
	}
}

func TestSavageBuffsCostHealthAndWork(t *testing.T) {
	cat, err := LoadCatalog(filepath.Join("..", "..", "..", "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	cast := func(r *rig, p *Player, sp *SpellDef, target combatant) {
		r.z.Cast(p, sp.ID, target.entityID())
		r.ticks(int(sp.CastTime*TickRate) + 10)
	}
	t.Run("a parry buff costs health and adds to the parry chance", func(t *testing.T) {
		r, p, n := realCaster(t, cat, rules.ClassSavage)
		sp := bestOfType(p, rules.SpellSavageParry)
		if sp == nil {
			t.Fatal("no parry buff in the Savage's book")
		}
		p.hp = float64(p.maxHP())
		power := p.power
		cost := p.powerCost(sp)
		want := p.hp - float64(cost)
		cast(r, p, sp, n)
		if p.hp > want+30 || p.hp < want-1 { // a little regeneration is allowed
			t.Errorf("health %v after a cast costing %d from %v, want about %v", p.hp, cost, float64(p.maxHP()), want)
		}
		if p.power != power {
			t.Errorf("power moved from %v to %v: the buff is paid in health", power, p.power)
		}
		if !hasKind(p, effSavPar) {
			t.Fatal("the parry buff left no effect")
		}
		if cb := p.combatant(cat); cb.ParryBonus <= 0 {
			t.Errorf("parry bonus %v after the buff", cb.ParryBonus)
		}
	})
	t.Run("a cost that would kill is refused", func(t *testing.T) {
		r, p, n := realCaster(t, cat, rules.ClassSavage)
		sp := bestOfType(p, rules.SpellSavageDPS)
		cost := p.powerCost(sp)
		p.hp = float64(cost)
		r.z.Cast(p, sp.ID, n.entityID())
		if p.casting != nil || hasKind(p, effDPS) || p.hp != float64(cost) {
			t.Errorf("cast with only %d health: casting %v, effect %v, health %v", cost, p.casting != nil, hasKind(p, effDPS), p.hp)
		}
	})
	t.Run("a resist buff guards the type it names", func(t *testing.T) {
		r, p, n := realCaster(t, cat, rules.ClassSavage)
		sp := bestOfType(p, rules.SpellSavageSlash)
		p.hp = float64(p.maxHP())
		cast(r, p, sp, n)
		cb := p.combatant(cat)
		if cb.Resists[rules.DamageSlash] <= 0 || cb.Resists[rules.DamageCrush] != 0 {
			t.Errorf("resists after a slash buff: %v", cb.Resists)
		}
	})
	t.Run("the speed buff shortens the swing", func(t *testing.T) {
		r, p, n := realCaster(t, cat, rules.ClassSavage)
		sp := bestOfType(p, rules.SpellSavageSpeed)
		p.hp = float64(p.maxHP())
		cast(r, p, sp, n)
		if p.haste() <= 0 {
			t.Errorf("haste %d after the combat speed buff", p.haste())
		}
	})
	t.Run("the endurance heal restores endurance for health", func(t *testing.T) {
		r, p, n := realCaster(t, cat, rules.ClassSavage)
		sp := bestOfType(p, rules.SpellSavageEndHeal)
		p.hp, p.end = float64(p.maxHP()), 10
		cast(r, p, sp, n)
		if p.end < 10+sp.Value*0.9 {
			t.Errorf("endurance %v after %s, want about %v", p.end, sp.Name, 10+sp.Value)
		}
		if p.hp >= float64(p.maxHP()) {
			t.Error("the endurance heal cost no health")
		}
	})
	t.Run("a taunt adds hate", func(t *testing.T) {
		r, p, n := realCaster(t, cat, rules.ClassSavage)
		sp := bestOfType(p, rules.SpellTaunt)
		if sp == nil {
			t.Fatal("no taunt in the Savage's book")
		}
		cast(r, p, sp, n)
		e := n.hate[p.entityID()]
		if e == nil || e.hate < rules.TauntHate(sp.Value) {
			t.Errorf("hate after %s: %+v, want at least %v", sp.Name, e, rules.TauntHate(sp.Value))
		}
	})
}
