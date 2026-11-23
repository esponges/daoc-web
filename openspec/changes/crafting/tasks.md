# Tasks

## 1. Sources and data

- [ ] 1.1 Find 1.65 crafting sources (success, quality, time, skill gain, imbue points, overcharge, salvage). Add them to CLAUDE.md and the README.
- [ ] 1.2 List the station figures (forge, lathe, alchemy table) in Midgard's placements (D2).
- [ ] 1.3 Check the install for a crafting animation and sounds.
- [ ] 1.4 Importer: read `crafteditem` and `craftedxitem`, filter to 1.65 (D1), and import supply merchant lists. Re-import. Tests.

## 2. Rules

- [ ] 2.1 `rules/crafting.go` with table tests per `rules-crafting` scenario and registry entries.

## 3. Server

- [ ] 3.1 Store: profession and skills (D3).
- [ ] 3.2 `world/crafting.go`: join at a master, recipe list, timed craft with cancel, result, skill gain, salvage.
- [ ] 3.3 `world/spellcraft.go`: imbue and overcharge (D4). `world/alchemy.go`: potions, poisons, charges and procs.
- [ ] 3.4 Protocol additions, `docs/protocol.md` and a protocol test.

## 4. Client

- [ ] 4.1 The stock Craft, MiniCraft and Tradeskills windows, with fallbacks. The progress bar. The crafter's name and bonuses in item tooltips.
- [ ] 4.2 Check in the browser: join, buy materials, craft to 100, imbue an item, brew a potion. Check the fallback.

## 5. Census and docs

- [ ] 5.1 `cmd/parity`: a crafting row and the masters as implemented. README and CLAUDE.md.
