## Why

`verify-spell-numbers` made a player's direct damage and heals depend on the casting stat and the spec in the spell's line, and fixed the power cost. The rest of what a player casts still ignores the caster: a damage-over-time spell, a life drain, a damage add and a buff all do exactly the table's number whatever the caster's stat or spec, and the buff caps (`BuffStatCap`, `BuffAFCap`) are provisional, the armour one a guess of two a level. A Runemaster at spec 1 and at spec 50 get the same DoT; a Healer buff is the same at any level. These are the next spells a friend will notice, and the Skald and Thane classes (next) lean on buffs and damage add.

## What Changes

- Read DOLSharp and OpenDAoC-Core (GPL, understanding only, no code copied) for how damage over time, life drain, damage add, stat buffs and the base armour buff scale with the caster, and for the buff caps. Record each formula, its source file and the 1.65 search in `findings.md`, as `verify-spell-numbers` did.
- Player damage over time and life drain scale by the casting stat and spec, as the sources show; a creature's stay as they are. Where the sources say a spell type does not scale, the rule says so and a test shows it.
- Damage add and the buffs take their strength from the caster where the sources do, and the caps (stat, specialisation buffs, armour factor) come from a source or stay provisional with the search recorded.
- Registry and specs: `spell.damage-over-time` and `spell.buff` get the sources and statuses the findings support, with worked scenarios and table tests; `rules-spell-effects`, `player-spellcasting` and `rules-power` are the specs touched.
- The data check (`TestSpellAmountsAgreeWithTheTable`) covers the newly scaled types, so a spell whose amount leaves the range the rule gives is named.
- Not in scope: stat and resist debuffs, speed-decrease strength, the focus bonus on cost, and anything about creature spells. Debuff spell types are not implemented by the server yet, so they would be a change of their own.
- No protocol change: amounts reach the client as they do today.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `rules-spell-effects`: damage over time and life drain for a player scale with the caster; damage add strength.
- `player-spellcasting`: buffs take their value from the caster and obey the sourced caps.

## Impact

- `internal/server/rules/spells.go` (new functions and registry entries), `internal/server/world/effects.go` (the DoT tick and casting), `internal/server/world/pcasting.go` (damage add and buff values), `internal/server/world/sheet.go` (caps).
- Tests: rules table tests per scenario, world tests at two spec levels for a Runemaster DoT and a buff, the extended data check.
- `docs/parity` regenerated and the README status counts updated; the figures test fails if they disagree.
- Risk: the two community servers may disagree on a term (they did for the level difference in damage); the spec names the one followed.
