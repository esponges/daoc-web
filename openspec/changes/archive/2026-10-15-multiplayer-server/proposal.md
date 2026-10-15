# Proposal

## Why

daoc-web is a proof of concept: one player, in one browser tab, against NPCs this project invented. The client install holds every asset but none of the world's rules or population (spawns, stats, loot, classes, styles, persistence); those lived on the game server. To be a real game it needs a server that owns the world, and some source for the missing content. The OpenDAoC project publishes a community database (GPL-3.0) of exactly that content, which makes a server-authoritative multiplayer game feasible without inventing every number.

## What Changes

- A new Go game server (`cmd/gameserver`, `internal/server/...`) that serves `web/`, accepts WebSocket connections and owns the world state. It runs on the local machine or LAN first; public hosting is a later step.
- Login by Discord OAuth2, or by an admin password for local use and game-master tools. Accounts, characters and positions persist in SQLite.
- A shared world: several players in one zone see each other move and fight, with server-validated movement, client-side prediction for the local player, and chat.
- NPC population, wandering, aggro, leashing, combat and death/respawn move from `web/npc.js` and `web/combat.js` to the server, fed by imported spawn and template data instead of the hand-written `web/spawns/zone100.json`.
- A build-time importer (`cmd/dbimport`) reads a local, gitignored checkout of the OpenDAoC database dump and writes this project's own world store. No OpenDAoC data is committed and no OpenDAoC code is copied; the repository stays MIT.
- Server-side inventory, loot, experience and levels, then collision against the scenery hulls that `propconv` now discards, and zone travel through the exits already read from `ZONEJUMP.csv`.
- The offline viewer mode stays, so the existing `node web/*.test.mjs` suites keep working.
- **BREAKING**: the viewer's NPC and combat logic is no longer authoritative when connected to a server; `rpg-combat`'s "authored numbers only" stance is superseded wherever imported data supplies the numbers.

## Capabilities

### New Capabilities
- `player-accounts`: Discord OAuth2 and admin-password login, sessions, and character creation, selection and persistence.
- `world-sync`: the WebSocket protocol, server tick, movement validation and prediction, other players' visibility, and chat.
- `world-data-import`: the importer, the local-only data directory, and the licensing rules that keep community data out of the repository.
- `server-npcs`: imported spawns and templates, and their wandering, aggro, leashing and respawn, run by the server.
- `server-combat`: authoritative melee resolution, defence, death, and the style and spell data used by combat.
- `character-progression`: inventory, equipment, loot, experience and levels.
- `zone-travel`: scenery collision for characters and NPCs, and movement between zones, cities and dungeons through exits.

### Modified Capabilities
<!-- None: zone-fidelity covers what a converted zone draws, which this change does not alter. -->

## Impact

- New code: `cmd/gameserver`, `cmd/dbimport`, `internal/server/{auth,hub,world,store}`, `internal/importer`; new dependencies for WebSocket, SQLite (a pure-Go driver, no cgo) and OAuth2.
- Changed code: `web/viewer.js`, `web/npc.js`, `web/combat.js` gain a network mode; `cmd/propconv` and `internal/scenery` keep collision hulls; `.gitignore` gains the community data directory; `README.md` gains a multiplayer and licensing section.
- Overlaps `openspec/changes/rpg-combat`: this change takes its server-side rules and numbers from imported data; the client presentation work in `rpg-combat` still applies.
- Licensing: the OpenDAoC database and server are GPL-3.0 and describe content owned by Electronic Arts. Scope is private and LAN use; sharing with friends over the internet needs that re-checked first. This is not legal advice.
- Out of scope: realm-versus-realm, keeps, housing, crafting, guilds, group play and spells beyond what combat needs, and any contact with live DAoC servers or their network protocol.
