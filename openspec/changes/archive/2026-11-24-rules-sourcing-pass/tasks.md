# Tasks

## 1. Tooling

- [x] 1.1 Registry `Searched` field and the census count (D2). Update `parity_test`, `docs/parity` and the README.

## 2. The pass

- [x] 2.1 Combat: melee damage, spell damage and variance, resists, archery (draw, damage, range, abilities).
- [x] 2.2 Experience, hit points, group experience share.
- [x] 2.3 Health, power and endurance regeneration; resting; falling; movement and sprint.
- [x] 2.4 Class tables (`character.class-*`), concentration end, item condition wear and repair price, casting decision and heal.
- [ ] 2.5 (not done this pass) Reference rules: look for period sources for each.

## 3. Close

- [x] 3.1 Every touched rule: spec, worked examples, tests, registry, design log (D3). Sources in CLAUDE.md and the README.
- [x] 3.2 `go test -short ./...` and `cmd/parity`; README status table.

## Outcome

Tooling done (`Searched` in the registry, `rules/sourcing.go`, the census count). The pass found one change (sprint speed to 1.30) and recorded the searches for 44 rules; the web holds little 1.65 numeric data. Reference rules (2.5) were not searched this pass.
