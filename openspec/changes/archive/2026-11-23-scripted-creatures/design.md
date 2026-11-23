# Design: scripted-creatures

## Context

- **What the data gives.** The community spawn rows carry a class type and a brain *name* (for example `QueenMajorAddBrain`, `RainbowSpriteGreenBrain`), and nothing else about behaviour: no thresholds, timers, spells or add lists. Those live in C# scripts under GPL.
- **What the server already does.** Every spawn in `world.db` is placed as a creature from its template. Night and day presence already works: `importer.SpawnTiming` reads the class type and `world/timed.go` hides and shows the creature with the game clock.
- **What the census said.** It listed NightSpawn and DaySpawn as unimplemented, because it never asked `SpawnTiming`.
- **What is missing from the world.** Measured 2026-11-23 against a scratch import that keeps every model: 1,693 of 24,134 Midgard-region placements (about 7%) have no converted model, across 35 models. Converting all 35 with `charconv` fails: wisps and spirits (models 911, 912, 913, 966) have "no shapes selected" (effect models with no skinned mesh), and the skugga, frogs and crawlers fail with "no geometry or no skin" (composite figures). The biggest: morvalt skugga 137, wind wisp 119, mud frog 73, shivering presence 68, carrion crawler 68.

## Searched sources (2026-11-23)

Public searches for Myling, Queen Major, Stripe, Zrit-Zrit and Snarls found nothing on what they do. The reference server's scripts could be read for understanding, but they are GPL and a named boss's mechanics are the substance of the script, so adopting them would be copying in spirit.

## Decisions

**D1. No behaviour engine yet.** The original plan was a data table of adds, flee and call-for-help, timed abilities and a cycle, driven per kind. With no source for any row, the table would be empty and the engine dead code. It is deferred until a source exists for at least one creature, at which point the first row and the engine are written together.

**D2. Every kind gets a stated status in the census.**
- Present by time of day (NightSpawn, DaySpawn): implemented.
- Not classic (test dummies, teleporters of other servers, effect carriers, Atlas and Origin Fantasy services): marked, and the server already answers their interaction with a line.
- Keep guards: owned by `frontiers-and-keeps`. Crafting masters: owned by `crafting`.
- Enchanter and Recharger: whether 1.65 had them is unsettled and no source gives what they did, so they spawn as plain creatures with no service.
- Everything else: "behaviour unsourced", spawning as a plain creature.

**D3. The model gap is recorded, not hidden.** It is a converter problem (non-skinned effect models and composite figures), separate from creature behaviour, and goes in the README's known limitations with the measured count.

## Non-Goals

- Keep guards, crafting masters, guild NPCs (their own changes).
- Converting effect and composite models: a converter change of its own.
