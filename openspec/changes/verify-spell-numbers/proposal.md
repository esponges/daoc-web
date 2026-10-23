## Why

Melee numbers trace to published 1.65 sources, but the spell numbers a player now
lives by do not: `spell.damage` (the 0.75 to 1.25 variance), `spell.buff` (the
stat and armour caps), `cast.heal` and the power pool and regeneration are
provisional, and `spell.damage` is the project's own choice. Spell damage also
takes no account of the caster: `SpellDamage` is the spell's base times a
variance times a resist, with no casting stat or spec level in it, and a heal is
the spell's flat `Value`. The tests prove the code matches the spec, not that the
spec matches 1.65.

## What Changes

- Read DOLSharp and OpenDAoC-Core (GPL: read to understand, copy nothing) for the
  spell damage, variance, spec and stat scaling, heal, buff and power formulas;
  search once for a published 1.65 source of each.
- Record what was found, with where it disagrees, in `findings.md`.
- Rewrite the affected rule specs in our own words with the source named, and move a
  rule's status only as far as its source supports (provisional to reference, or to
  documented where a 1.65 source exists).
- Add a check that computes, for every imported spell a player or creature can
  cast, what the server would deal or restore, and compares it with the spell
  table's own `Damage`/`Value` so a wiring mistake shows.
- Where a source shows scaling the server lacks (caster stat, spec level), add it
  with its worked examples. If that is large, split it into its own change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `rules-spell-effects`: player spell damage and heals scale with the casting stat and spec; spell rules carry a recorded source before a status moves, and the
  server's spell numbers are checked against the imported spell table.
- `rules-power`: the pool, regeneration and the reading of a spell's power cost follow the community servers.

## Impact

- `internal/server/rules` (`spells.go`, `casting.go`, `power.go`), their tests and the rule registry.
- A new test over the imported spell table (skips without `server/data/community`).
- `openspec/specs/rules-spell-effects`, `rules-power`, `player-spellcasting`; `docs/parity` regenerated.
- The two GPL repositories are cloned into a scratch directory outside the repo, never committed.
