## 1. Sources

- [x] 1.1 Clone DOLSharp and OpenDAoC-Core into the scratch directory and read how damage over time, life drain, damage add, stat buffs, the base armour buff and the buff caps scale with the caster; record each formula's shape, project and file in `findings.md` (our own words, no code or identifiers); verify `findings.md` has an entry for each of `spell.damage-over-time`, life drain, damage add, `spell.buff` stat, `spell.buff` armour factor, and says where the two projects disagree.
- [x] 1.2 Run one search per formula for a published 1.65 account; verify `findings.md` records the query and result, including empty results.
- [x] 1.3 Settle the open questions in `design.md` (does damage add scale, does a weaker recast replace a stronger one, is the armour cap level-only); verify each is answered in `findings.md`.

## 2. Rules

- [x] 2.1 Add the DoT and life-drain scaling to `internal/server/rules/spells.go` as pure functions with the sources named, and update `spell.damage-over-time` in the registry; verify a table test per scenario in `rules-spell-effects` and `go test ./internal/server/rules` passes.
- [x] 2.2 Add the damage-add and buff strength and cap rules, and update `spell.buff` and `BuffAFCap` (status provisional if no source); verify the registry statuses changed only where a source was recorded.

## 3. Wiring

- [x] 3.1 Record the caster's scale on the effect when it lands in `addEffect`, use it in `tickEffectsOf`, and use the buff and damage-add rules in `buffBonuses` and `damageAdd`; verify creature DoTs and procs are unchanged by their existing tests.
- [x] 3.2 Add world tests at two spec levels for a Runemaster DoT, and a buff and a damage add, plus the recast and fixed-at-cast scenarios; verify `go test -short ./internal/server/world` passes.

## 4. The data check

- [x] 4.1 Extend `spellProblem` and `TestSpellAmountsAgreeWithTheTable` to the newly scaled types; verify the check reports no spell and skips without the dump.
- [x] 4.2 Plant a wrong value and check the failure names the spell id and both amounts, then remove the plant.

## 5. Close

- [x] 5.1 Regenerate `docs/parity`, update the README status counts and the rule-book text; verify `go test ./cmd/parity` passes.
- [x] 5.2 Cast a DoT and a buff in the running game at two spec levels on a scratch server and compare the tick and the stat shown with the rule; verify by driving the browser, then delete the scratch clones and data and check `git status` is clean of them.
