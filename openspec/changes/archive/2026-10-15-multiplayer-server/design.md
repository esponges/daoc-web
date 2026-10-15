# Design

## Context

Today everything runs in the browser. `web/viewer.js` drives the player, `web/npc.js` the wandering, `web/combat.js` the fights; the Go side is offline converters plus `cmd/serve`, a static file server. Static facts the server needs are already reachable from Go: `internal/zone` (heightmap, water), `internal/scenery` (placements, models), `internal/interior` (dungeon geometry, exits), `internal/gamedata` (`monsters.csv` and friends). Motivation is in proposal.md.

The community source is the OpenDAoC database: one MySQL `.sql` file per table (`mob` ~30 MB, `itemtemplate` ~14 MB, `npctemplate`, `npcequipment`, `spell`, `style`, `stylexspell`, `linexspell`, `specialization`, `loottemplate`, `mobxloottemplate`, `path`, `pathpoints`, `jumppoint`, `regions`, `startuplocation`, `bindpoint` and more). The listing I saw was cut off at 105 files, so tables such as class and zone data are unconfirmed. It is GPL-3.0 and targets patch 1.65.

## Goals / Non-Goals

**Goals:**
- A server that owns the world and keeps the client honest enough for friends on a LAN.
- Content from imported data, not invented, with a clean licence boundary.
- Reuse the Go readers and the web animation/render code unchanged where possible.
- An offline viewer mode that keeps the existing tests running.

**Non-Goals:**
- Anti-cheat against a determined attacker, horizontal scaling, or more than a few dozen concurrent players.
- Matching the live game's balance, RvR, keeps, housing, crafting, guilds, or the live protocol.
- Public hosting hardening; that is a later change.

## Decisions

**Go server in this repo, one binary.** `cmd/gameserver` serves `web/`, the OAuth callbacks and the WebSocket. Alternatives: Node (second language and no reuse of `internal/zone`), or extending `cmd/serve` (it stays as the plain static server).

**WebSocket with JSON messages first.** Simple to debug in the browser, enough for tens of entities. Messages carry a type and a protocol version; unknown types are ignored. A binary format can replace the hot path (snapshots) later without changing behaviour. Alternative: WebRTC data channels, rejected as more setup for no gain on a LAN.

**Fixed-tick authoritative simulation, driven by inputs rather than positions.** One goroutine per active zone ticks at 20 Hz, applies queued inputs, steps NPCs and combat, and sends per-client snapshots of the entities within 14,000 units. The client does not report where it is. It reports what it holds down, as runs of fixed 1/60 s steps (`[count, fwd, strafe, turn, flags]`), and the server runs the same `Step` function on them. A player's steps are spent against real time, three per tick with a bank of twenty, so a client that sends ten seconds of steps at once moves one second's worth; the client's own predicted position is sent only for comparison and a disagreement sets a `fix` flag that makes the client snap instead of ease. The browser predicts with the same step (`web/move.js`), rewinds to the server's position at the last acknowledged step and replays the rest. Remote entities are drawn 120 ms in the past, interpolated. The rates (85 and 191 units a second, 2.8 rad/s turning) and the viewer's two-triangle `groundAt` are written once in Go and once in JS and checked by vectors the Go tests write (`testdata/move_vectors.json`, agreement to zero difference). Sprint is offline only: the server runs and walks.

**Zones are loaded from the same converted data.** The server reads `web/data/zone<N>/` (`heights.u16`, `zone.json`, `props.json`), not the install, so a converted tree is all it needs. Zones tick only with players in them.

**Persistence: SQLite through a pure-Go driver (`modernc.org/sqlite`).** No cgo keeps Windows builds simple. Accounts, characters, inventory and bind points in one file; the imported world in a second, read-only file produced by the importer, so a re-import never touches player data. Alternative: Postgres, deferred until public hosting.

**Auth: Discord OAuth2 code flow, plus an admin password.** The callback creates a signed session cookie; the WebSocket upgrade checks it. The admin password is stored as a bcrypt hash from an environment variable, with no default. Both paths end in the same session type, which carries an admin flag. Discord needs a registered application and HTTPS in production, but `http://localhost` works for development.

**Licence boundary for imported data.** The importer (`cmd/dbimport`, `internal/importer`) is our own code, parses the SQL `INSERT` statements itself and writes our own schema. The checkout lives under a gitignored `server/data/community/`, outputs under `server/data/world.db`. No rows or dumps are committed, and combat and AI are written from documented mechanics and the imported numbers, not from OpenDAoC or DOLSharp source. This keeps the repo MIT. Alternative considered: relicensing the server GPL-3, which is simpler to defend but binds every future contributor; kept as the fallback if code reuse is ever wanted. Not legal advice. The data describes EA-owned content, so scope remains private and LAN.

**Matching rows to assets.** NPC templates carry a model id. Whether it equals `monsters.csv`'s model number is unverified and is the first thing the data spike checks. Coordinates likewise: region coordinates must be mapped to zone coordinates (the zone's region offset and the 65536-unit zone size). The spike derives the transform per zone from known landmarks, such as the heights of imported spawns against `heights.u16` or the position of Mularn's props, and the importer refuses zones it cannot fit, as `interiorconv` does for interiors.

**Combat rules are ours.** Swing timers, to-hit, damage and defences are implemented in `internal/server/combat` from documented DAoC mechanics, with the numbers (style requirements, endurance costs, damage tables, spell values) taken from the imported tables. The existing `web/combat.js` rules are the starting point, ported to Go with `combat.test.mjs` cases carried over as Go tests.

**Collision.** Three sources, one solver. Outdoors, `propconv` keeps each model's invisible `collidee` hull in `collision.bin`/`collision.json` beside `props.bin` (the visible output is byte-identical to before), and for a model with no hull, such as a tree, `nifs.csv`'s trunk `Radius` makes a cylinder; 43 of Mularn's 60 models have hulls, and the two pines account for most of the rest. Indoors there are no hulls to rely on, so the walls are the visible geometry's steep triangles and the floor is its upward-facing ones, found by casting down from a step above the feet as `web/interior.js` does. A walker is a circle of 32 units; what blocks stands between 60 and 170 units above its feet (lower is stepped over); a move is pushed out of whatever it would overlap, which slides it along a wall. Indoors a move that would leave the floor, onto a drop of more than 110 units or into a void, is refused the same way, since nothing falls. Go (`collision.go`) and JS (`web/collision.js`) are checked against shared vectors (`testdata/collision_vectors.json`). Creatures use the same test.

**Seamless land, separate zones.** The game's outdoor zones are one continuous region; here each is its own place. Walking off an edge enters the neighbour found by placing the player in region coordinates (`zones` table offsets in 8192-unit blocks) 600 units beyond the edge, and sets them down just inside it; the client reloads at the new zone. Exits (`zonepoint`) work the same way. A player who arrives standing in an exit, as everyone does coming through its twin, must step out of it before it triggers again.

**Clients, characters and models.** Creatures are drawn with `web/data/char/m<model number>`, converted by `charconv -model N -name mN`; `dbimport` keeps only spawns whose model is converted, so the server never sends one the client cannot draw. A model that converts but cannot be drawn (a skin needing more bones than the shader has) is dropped by the client with a warning instead of taking the frame down. Players are drawn as one of two archetypes, a Troll Warrior and a Norse Warrior (the race ids the start-location table uses), whose appearance is fixed by their outfit. Equipment changes the numbers, not the picture.

**Client network mode.** `web/net.js` handles the socket, prediction and interpolation, `web/remote.js` draws and animates server entities with the existing characters and animator, and `web/ui.js` is the chat, vitals, style bar, pack and loot panels. `viewer.js` enters network mode with `?play=<character id>`: it joins first, learns the zone from the server's welcome and reloads there if it differs. Without `?play=` the existing offline code runs unchanged, and `web/npc.js` and `web/combat.js` stay for offline use. A tab in the background sends a ping every 15 s so the server's idle timeout (120 s) does not take it for gone.

**Login.** Both paths issue the same signed cookie. The Discord callback logs each failure's reason (a missing state cookie, Discord's own error, a rejected token exchange), because a failed OAuth login is otherwise invisible from the server's side. The server creates its data directory, generates a session key into `server/data/session.key` if none is given, and refuses to start with no way to log in.

## Risks / Trade-offs

- [Imported model ids or coordinates do not line up with the client assets] → They did (see the findings below); unmatched models are dropped and counted, and 692 of the dump's models have no converted character.
- [The dump's tables differ from what the first listing suggested] → `zones` and `zonepoint` exist and there is no class table, so the Warrior's definition and every hit point, damage and experience number are this project's (`rules.go`), with style requirements, endurance, weapons and creature levels from the data.
- [GPL-3 data or EA-owned content creates legal exposure] → Nothing imported is committed or redistributed, scope stays private; the README says so and re-checking is required before any public hosting.
- [Prediction and server movement diverge] → Shared test vectors, a server correction path, and a tolerance rather than exact equality.
- [A port of combat drifts from the offline JS] → Port the existing combat test scenarios as Go tests first.
- [Large imports slow startup] → The importer runs once; the server opens a prebuilt read-only SQLite file.
- [Discord OAuth is awkward for local development] → The admin password path works with no network.
- [Scope is large] → Phased so every phase leaves a working game; later phases (progression, collision, travel) can be cut or reordered without breaking earlier ones.

## Migration Plan

Additive. `cmd/serve` and the offline viewer keep working. `.gitignore` gains `server/data/`. No existing converted data changes, except that `propconv` additionally writes `collision.bin`; re-running `worldconv` regenerates it. Rollback is to stop using `gameserver`.

## Findings from the data spike (tasks 1.1-1.3)

Measured against the OpenDAoC database (110 table files) and the local install, on zone 100.

- **Tables.** The listing in Context was truncated. The dump also has `zones` (region, name, `OffsetX`, `OffsetY`, `Width`, `Height`) and `zonepoint` (zone exits: source and target region and position, 169 rows). There is **no `class` table**: classes are code in the emulator, so only `classxspecialization`, `specialization`, `specxability`, `linexspell`, `stylexspell` and `style` carry class data, and the Warrior's definition stays ours. `npctemplate` has 12,933 rows, `jumppoint` 16, `startuplocation` 183 and `mob` roughly 25,000 spawn rows (the dump is a 30 MB file). `Level`, `Size`, `Model` and `MaxSpeed` in `npctemplate` can be ranges stored as text, such as `'50-58'`.
- **Coordinates.** A spawn's `X`, `Y` are region coordinates. The zone's local position is `X - OffsetX*8192`, `Y - OffsetY*8192` (zone 100: `OffsetX` 92, `OffsetY` 82), with no axis flip, and local units match `heights.u16` cell units (256 per cell). `Z` is absolute height. 2,219 mobs in region 100 fall inside zone 100's box, and 2,202 of them have `Z` within 300 units of the derived terrain (mean absolute difference 39), so the transform is right and the importer need not search for one. The 17 outliers are high flyers and spawns the importer will drop.
- **Models.** `mob.Model` and `npctemplate.Model` are the client's `monsters.csv` ids: 56 is the Large Grey Wolf, 29 a snake, 99 a boar, 117 a spider, 101 a bear, 133 a lynx, 392 the Vendo Warrior, 1390 a Svartalf. The importer still reports models `charconv` cannot convert.
- **Consequence.** Neither hypothesis failed, so the plan proceeds. The importer takes the zone transform from the `zones` table rather than fitting it, but it keeps the height check as a guard.

## As built, and what was left out

The phases of the plan all landed. Differences from the first draft:

- Zone borders and a latch on exits were not in the draft; they came from playing it (the exits' arrival points are inside their twins' trigger radius).
- Collision hulls cover less than expected, so trunk radii and interior walls from visible geometry carry the rest.
- Imported interiors' zone coordinates come from the `zones` table's offsets (1 x 8192 for dungeons), which matched the converted dungeons' exits to within 170 units.
- Left out, deliberately: spells and style effects beyond damage (stuns, taunts), group play, merchants, trainers, quests, a second class, style icons, and the appearance of worn weapons. The README's Known limitations lists them.
- Verified in a browser, not only in tests: both logins (the real Discord one with a registered application), two tabs in one zone, combat with styles, loot, equipping, a dungeon entrance, and Myrkwood Forest, Gotar, Vale of Mularn and East Svealand loading.
