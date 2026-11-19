package world

import "daocweb/internal/server/rules"

// sighted is a Mover that can say whether two eye points see each other.
type sighted interface {
	Sees(a, b [3]float64) bool
}

// canSee says whether a body at feet a can see one at feet b. The scenery
// answers for walls, floors and trunks; an outdoor zone's heightmap answers
// for hills. A zone with neither (the tests' flat grounds, a zone without
// collision data) lets everything see everything.
func (z *Zone) canSee(ax, ay, az, bx, by, bz float64) bool {
	a, b := rules.Eye([3]float64{ax, ay, az}), rules.Eye([3]float64{bx, by, bz})
	if t, ok := z.G.(*Terrain); ok {
		if rules.SegmentUnderGround(a, b, func(x, y float64) float64 { return t.GroundAt(x, y, 0) }) {
			return false
		}
	}
	if s, ok := z.Mv.(sighted); ok {
		return s.Sees(a, b)
	}
	return true
}

// sees is canSee between two combatants.
func (z *Zone) sees(from, to combatant) bool {
	fx, fy := from.position()
	tx, ty := to.position()
	return z.canSee(fx, fy, from.footZ(), tx, ty, to.footZ())
}

// needsSight says whether a player's spell must see its target: not one on the
// caster, on the group, or an area around the caster.
func (z *Zone) needsSight(p *Player, sp *SpellDef, target combatant) bool {
	return rules.NeedsSight(target == combatant(p), sp.Target == "Group", isPBAoE(sp))
}
