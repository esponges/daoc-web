# Tasks

## 1. Sources

- [ ] 1.1 List the classic frontier zones per realm (D1). Find sources for realm point value, the rank table, the anti-farm rule, the rank skill bonus, immunity times and the 1.65 realm ability set.

## 2. Rules

- [ ] 2.1 `rules/realmpoints.go`: value, share, reduction, rank table; tests and registry.
- [ ] 2.2 Realm ability entries (amounts per level, costs, reuse); tests and registry.
- [ ] 2.3 PvP: hit location against players, immunity times; tests and registry.

## 3. Server

- [ ] 3.1 RvR areas and `mayHarm` for realms (D1). Every attack path is tested against an enemy player.
- [ ] 3.2 Damage credit and realm point award with group sharing (D2). Rank-up messages and the skill bonus. Store migration.
- [ ] 3.3 Realm ability buy, respec, passives and actives (D3). Import from `classxrealmability_atlas`, filtered to 1.65.
- [ ] 3.4 Death in RvR and the release point.
- [ ] 3.5 Admin: spawn a test character of a realm, set realm points.
- [ ] 3.6 Protocol additions, `docs/protocol.md` and a protocol test.

## 4. Client

- [ ] 4.1 The Realm training tab, realm ability icons and quickbar, the RP float text, enemy name colour, and the realm data on the character sheet.
- [ ] 4.2 Check in the browser with two realms in a frontier zone: a kill, RP shared in a group, and an ability used.

## 5. Census and docs

- [ ] 5.1 `cmd/parity`: realm abilities row; "fixed hit location" implemented. README and CLAUDE.md.
