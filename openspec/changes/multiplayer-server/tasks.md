# Tasks

## 1. Data spike and licence boundary

- [x] 1.1 Add `server/data/` to `.gitignore`, clone `OpenDAoC-Database` into `server/data/community/`, and list every table file with its size; verify `git status` shows nothing from the checkout and the list confirms or corrects the tables named in design.md (class and zone tables in particular)
- [x] 1.2 Write a throwaway reader that prints the columns and a few rows of `mob`, `npctemplate`, `regions`, `jumppoint` and `startuplocation`; verify the zone 100 (Vale of Mularn) region's spawns can be found and counted
- [x] 1.3 Check whether spawn model ids equal `monsters.csv` model numbers (for example wolf 56) and whether spawn coordinates map to zone coordinates; verify with a printed table of at least 20 Mularn spawns against `heights.u16` heights, and record the result and the transform in design.md, stopping here if neither lines up
- [x] 1.4 Add the licence and scope statement to the README's multiplayer section; verify it names the source repository, GPL-3.0, EA-owned content, "not redistributed" and private use

## 2. Importer

- [x] 2.1 Add `internal/importer` that parses MySQL `INSERT` statements from the per-table dumps (quoting, escapes, NULLs, multi-row inserts); verify with unit tests on small inline samples, including strings with quotes, commas and newlines
- [x] 2.2 Define the world store schema and add `cmd/dbimport` writing `server/data/world.db` from spawns, NPC templates and equipment, loot, items, spells, styles, specialisations and classes, races, regions, jump points, start locations and bind points; verify per-table rows read and kept are printed and a missing required table exits non-zero
- [x] 2.3 Match NPC models to converted figures and fit region coordinates to zone coordinates using the result of 1.3; verify the report lists unmatched models and unfitted zones with examples and that zone 100 keeps most of its spawns
- [x] 2.4 Drop spawns that fall outside the zone or far from the ground; verify a test over the imported zone 100 spawns checks every kept one is within the set distance of `heights.u16`
- [x] 2.5 Document the importer (how to fetch the dump, run it, what it writes) in the README; verify the documented commands run as written on a clean checkout

## 3. Server skeleton, accounts and characters

- [x] 3.1 Add `cmd/gameserver` that serves `web/` and a health endpoint, with config from flags and environment; verify `go run ./cmd/gameserver` serves the viewer at `localhost` and exits with a clear message when no admin password or Discord credentials are configured
- [x] 3.2 Add the player store in SQLite (pure-Go driver): accounts, characters, sessions; verify unit tests cover creating, loading and a restart round trip
- [x] 3.3 Add admin-password login with a bcrypt hash from the environment, signed session cookies and rate limiting of failures; verify tests for the right password, a wrong password, repeated failures and a missing cookie on the WebSocket upgrade
- [x] 3.4 Add Discord OAuth2 login (authorize, callback, account by Discord id); verify with a stubbed OAuth server in tests, and manually that a real Discord login reaches the character screen on `localhost`
  - Verified 2026-10-14: against a stand-in Discord in the tests, and with a real Discord application in a browser (authorize, callback, account created as a non-administrator, character selection).
- [x] 3.5 Add character create, list, select and delete with unique names ignoring case; verify tests for a duplicate name, another account's character and name length and characters
- [x] 3.6 Add a minimal login and character-select page to `web/`; verify in the browser that both login paths lead to selecting a Troll Warrior
- [x] 3.7 Document configuration (Discord application setup, admin password, secrets staying out of the repo) in the README; verify each documented variable is read by the server

## 4. Shared world and movement

- [x] 4.1 Define the protocol (version, hello, input, snapshot, correction, chat, error) in one document and Go types; verify a test round-trips every message and ignores an unknown type
- [x] 4.2 Add the hub and per-zone 20 Hz tick that loads `heights.u16`, `zone.json` and `props.json`; verify a test starts zone 100 and ticks with no players, with and without them
- [x] 4.3 Implement server movement from sequenced inputs with walk, run, strafe and back speeds on the same triangle-split ground height the viewer uses; verify tests for the 191 u/s run over one second, rejecting an over-speed report and clamping to the ground
- [x] 4.4 Add shared movement test vectors used by both a Go test and `web/net.test.mjs`; verify both produce the same positions for the same inputs
- [x] 4.5 Add `web/net.js`: socket, input sending, prediction, reconciliation and interpolation of remote entities; verify `node web/net.test.mjs` with simulated latency keeps the local error within tolerance and never snaps for small corrections
- [x] 4.6 Branch `web/viewer.js` into network mode (`?server=` or served by `gameserver`) while offline mode stays unchanged; verify the existing `node web/*.test.mjs` suites still pass and two browser tabs see each other walk in Mularn
- [x] 4.7 Add interest management by distance, say and broadcast chat with rate and length limits, and clean disconnect handling with saving; verify tests for out-of-range players, a flooding client and a closed tab
- [x] 4.8 Document the protocol, network mode and chat in the README; verify the described URL parameters work

## 5. Server-run NPCs

- [x] 5.1 Load spawns, templates and paths from `world.db` into zones; verify zone 100 on the server holds the imported NPCs and the offline viewer still reads `web/spawns/zone100.json`
- [x] 5.2 Port wander, roam radius and water avoidance from `web/npc.js` to Go; verify a Go test that simulates ten minutes and reuses the checks of `npc.test.mjs` (no NPC outpaces its walk, strays from its radius or wets its feet)
- [x] 5.3 Add aggro, chase, leash, home and heal, and packmate assist; verify tests for an aggressive wolf approaching, a leashed NPC returning and a pack joining
- [x] 5.4 Add NPC snapshots and have the client draw them with the existing characters and animator; verify two tabs see the same wolf in the same place and state
- [x] 5.5 Add death, corpse time and respawn; verify a test that a killed NPC vanishes and respawns at home with full health

## 6. Server combat

- [x] 6.1 Port targeting, swing timers, to-hit and damage from `web/combat.js` to `internal/server/combat`, with the scenarios of `combat.test.mjs` as Go tests; verify they pass
- [x] 6.2 Add action requests (target, attack on/off, style) and results as protocol messages, and send swings, hits, flinches and deaths to nearby clients; verify a test that a client claiming a hit out of reach deals no damage
- [x] 6.3 Add block, parry and evade from class and weapon, with NPC defence; verify tests for each defence's chance and front-arc rules
- [x] 6.4 Add endurance with regeneration and imported styles with requirements (anytime, opening after block/parry/evade, positional, follow-up); verify tests for an opening, a failed requirement and not enough endurance
- [x] 6.5 Add the style hotbar and combat log in `web/`, using style icons, and the `rpg-combat` animation roles for styles, parry and evade; verify in the browser that a style swing plays for a second watcher
- [x] 6.6 Add player death, bind-point return and NPC disengage; verify a test that wolves go home when a player dies and the player stands at the bind point after the delay
- [x] 6.7 Document the combat model, stating that the rules are this project's own and which numbers come from imported data; verify the README matches the code constants

## 7. Inventory, loot and levels

- [x] 7.1 Add the item store from imported `itemtemplate` (names, slots, class limits, models) with inventory and equipment on characters; verify tests for equipping, class limits and a restart round trip
- [x] 7.2 Add equipment snapshots so other players see a worn hammer and shield; verify in the browser that a second tab sees the first's weapon
- [x] 7.3 Add loot tables for NPC kills with an exclusive window and atomic pickup saved with the character; verify tests for the drop roll, someone else's corpse and a restart mid-pickup
- [x] 7.4 Add experience, levels, hit point growth and style unlocks by level; verify tests for level thresholds, grey kills and unlocked styles
- [x] 7.5 Add inventory and character panels in `web/`; verify in the browser that loot taken from a wolf appears and can be equipped

## 8. Collision and zone travel

- [x] 8.1 Keep collision hulls in `propconv` and `internal/scenery`, writing `collision.bin` per zone beside `props.bin`, and leave the visible output byte-identical; verify `propconv -zone 100` writes the same `props.bin` and `props.json` as before and the new file's triangle count matches the models' hulls
- [x] 8.2 Add collision queries (a coarse grid and a capsule test) in Go and JS, with test vectors shared between them; verify tests for walking into a building, sliding along a wall and a model with no hull
- [x] 8.3 Apply collision to players and NPCs on the server and in prediction; verify two tabs agree on position against a Mularn building and `npc.test` style checks pass for NPCs
- [x] 8.4 Use the interior floor and wall triangles for dungeon and city collision and floor selection on the server; verify a test that walking at a Spindelhalla wall neither passes through it nor drops a level
- [x] 8.5 Wire zone exits from `ZONEJUMP.csv` and the imported jump points so walking into one moves the character to the target zone, refusing unconverted targets; verify a player enters Spindelhalla from the dungeon entrance and a second player there sees them
- [x] 8.6 Run zones only while they have players, saving and unloading empty ones; verify a test that the last player leaving stops the tick and NPCs reset on reload
- [x] 8.7 Update the README's Known limitations (collision, zone travel, NPC placement, combat) and the multiplayer section; verify no limitation listed is contradicted by the new behaviour

## 9. Integration

- [x] 9.1 Play a session with two browser tabs from login to leveling: log in by both methods, walk, chat, fight a wolf together, loot it, equip the drop, enter a dungeon; verify the whole session works with no console errors
  - Verified 2026-10-14 in a browser: administrator-password and real Discord logins, two tabs walking, chatting and fighting in Myrkwood Forest, loot and equipping, levelling by /gm, styles in a fight, a dungeon entrance, Gotar, Mularn and East Svealand loading, and a Discord player walking and refused a GM command.
- [x] 9.2 Run `go test ./...` and every `node web/*.test.mjs`; verify all pass, and that `git status` shows no community data or secrets tracked

## Workflow follow-up

- Re-check licensing and the EA content question before any public hosting.
- Archive the change after review.
