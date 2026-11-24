# Tasks

## 1. Sources and data

- [ ] 1.1 Find 1.65 crafting sources (success, quality, time, skill gain, imbue points, overcharge, salvage). Add them to CLAUDE.md and the README.
- [ ] 1.2 List the station figures (forge, lathe, alchemy table) in Midgard's placements (D2).
- [ ] 1.3 Check the install for a crafting animation and sounds.
- [x] 1.4 Importer: read `crafteditem` and `craftedxitem`, filter to 1.65 (D1), and import supply merchant lists. Re-import. Tests.

## 2. Rules

- [x] 2.1 `rules/crafting.go` with table tests per `rules-crafting` scenario and registry entries.

## 3. Server

- [x] 3.1 Store: profession and skills (D3).
- [x] 3.2 `world/crafting.go`: join at a master, recipe list, timed craft with cancel, result, skill gain, salvage.
- [ ] 3.3 `world/spellcraft.go`: imbue and overcharge (D4). `world/alchemy.go`: potions, poisons, charges and procs.
- [x] 3.4 Protocol additions, `docs/protocol.md` and a protocol test.

## 4. Client

- [ ] 4.1 The stock Craft, MiniCraft and Tradeskills windows, with fallbacks. The progress bar. The crafter's name and bonuses in item tooltips.
- [x] 4.2 Check in the browser: join, buy materials, craft to 100, imbue an item, brew a potion. Check the fallback.

## 5. Census and docs

- [x] 5.1 `cmd/parity`: a crafting row and the masters as implemented. README and CLAUDE.md.

## Outcome

- Done: the recipes are imported (`recipe`, `recipe_material`, 6,677 recipes of which about 1,000 are Midgard's), the six professions with their secondary skills, the rule book (`rules/crafting.go`, provisional; no public source gives the curves), skills stored in `char_craft`, the crafted item's quality and maker stored on the item and counted in combat, timed crafts cancelled by moving, fighting or casting, materials stacking at merchants, the `craftop` protocol, and a plain crafting window (Y). Checked in the browser: join, a craft to a 94 percent hand axe, skill gain, materials used.
- The skill numbers of the data were read from the recipes themselves (6 metalworking, 7 leatherworking, 8 clothworking, 11 tailoring, 12 fletching, 14 woodworking); there are no recipes for spellcrafting or gem cutting.
- Moved to `crafting-imbuing-and-salvage`: stations (1.2), spellcraft and imbuing (3.3), salvage, alchemy procs. The stock Craft windows (4.1) and the crafting animation (1.3) are not done; the plain window is used.
- 1.1 found no 1.65 source: every crafting rule is provisional.
