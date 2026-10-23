## 1. Sources

- [x] 1.1 Clone DOLSharp and OpenDAoC-Core into the scratch directory and read their spell damage, variance, heal, buff cap and power code; record each formula's shape, project and file in `findings.md`; verify `findings.md` has an entry for each of `spell.damage`, `spell.buff`, `cast.heal`, `power.pool`, `power.regen`, `character.health-regen`.
- [x] 1.2 Run one search per formula for a published 1.65 account; verify `findings.md` records the query and result, including empty results.

## 2. Rules

- [x] 2.1 Rewrite `spell.damage` and `spell.heal` in `rules-spell-effects` from the confirmed formula, in our words, with the source named and worked examples; verify each numbered scenario has a table test and `cmd/parity` still passes.
- [x] 2.2 Do the same for `spell.buff` (in `player-spellcasting`), `cast.heal`, and the power and regeneration rules (in `rules-power`); verify the registry statuses changed only where a source was recorded.
- [x] 2.3 If the sources show scaling the server lacks (caster stat, spec level), add it with examples; verify a Runemaster and a Healer cast in a test at two spec levels. If it needs more than the rules package and the effects code, stop and split a separate change.

## 3. The data check

- [x] 3.1 Add a test that walks the imported spell table for implemented types and compares the server's amount, from the formula confirmed in group 2, with the table value; verify `go test ./internal/server/rules` passes with the dump and is skipped without it.
- [x] 3.2 Make the test fail on a planted wrong value; verify it names the spell id and both amounts, then remove the plant.

## 4. Close

- [x] 4.1 Regenerate `docs/parity` and update the README status counts; verify `go test ./cmd/parity` passes.
- [x] 4.2 Cast one damage spell and one heal in the running game and compare the numbers shown in chat with the check's expectation; verify by driving the browser.
- [x] 4.3 Delete the scratch clones; verify nothing from them is in `git status`.
