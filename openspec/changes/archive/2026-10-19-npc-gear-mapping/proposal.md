## Why

`npc-equipment` gets each NPC's worn gear from the community database to the client (import, `eq` field on snapshots). Drawing it is blocked on one problem: the database names pieces by number, and the install is the only place that can turn a number into art. The spike in `npc-equipment` found that weapon numbers partly resolve through `items.csv`, that most armour numbers do not appear in `pskins.csv`, and that the same numbers hit unrelated rows in `skins.csv` and `mskins.csv` (each table has its own id space). The README already records that the install's item-to-armour mapping has not been found. Without a decision on this, NPCs stay plain.

## What Changes

- Draw weapons and shields on NPCs from `items.csv`, loading each item model once and sharing it, and only when the item's kind fits its slot (so a bow in a hand slot is rejected, not drawn).
- Run a time-boxed search for the real armour mapping in the install (client binaries, other archives, filename conventions) and record the result either way.
- If the search finds the mapping, resolve armour exactly from it. If not, draw armour by tier: pick the tier per slot from the NPC's realm and which slots it wears, dress it in the realm's texture family, and tint it with the real `Color` from `tints.csv`.
- Record in the README where each part comes from and which parts are exact versus approximate.
- **Scope extended:** players get the same treatment. The server sends a player's race, gender and worn equipment (additive), and the local and remote players draw their real gear through the same resolution path as NPCs, replacing the fixed baked outfit.
- **Race visuals:** players and NPCs draw the base model and texture for their race and gender. The race-to-model mapping is established from the install first (a spike), never by guessing an offset.
- **Validation:** once implemented, the result is checked in a real browser with claude-in-chrome, not only by node tests.
- The name `npc-gear-mapping` is now narrower than its scope; it is kept to avoid churn.

## Capabilities

### New Capabilities
- `player-appearance`: the server sends a player's race, gender and equipment, and the client draws the local and remote players from them.
- `race-visuals`: base model and texture follow race and gender for players and NPCs, with a plain fallback.
- `npc-gear-mapping`: how a database equipment piece is resolved to art, which resolutions are exact, and what happens when one cannot be resolved.

### Modified Capabilities
- `npc-appearance`: drawing now follows the resolution rules in `npc-gear-mapping` (the quiet-degrade requirement stays).

## Impact

- `cmd/charconv` / `cmd/worldconv`: convert held items once as shared models under `web/data/` instead of baking them into each character.
- `web/character.js`, `web/remote.js`: attach shared items to NPC sockets at runtime; armour tier textures per slot; tint.
- A generated, gitignored lookup in `web/data/` built from the install's `items.csv`, `tints.csv` and the armour search result. Nothing from the install or the database is committed (CLAUDE.md rules).
- Overlap with `npc-equipment`: its section 4 (converter and client drawing) is superseded by this change. `npc-equipment` should finish at import, server and protocol, and have its section 4 removed or marked as moved here. That edit is not made by this proposal.
- `internal/server/protocol`, `docs/protocol.md` (additive player fields plus a protocol test), `internal/server/world/player.go`, and the persistence of race and gender (a gender column; no creation UI); `web/viewer.js` for the local player.
- Risk: the database's player race ids may not map to the install's models, and converting every race and gender base is long, as it was for creatures.
- Risk: the armour search may find nothing. The tier fallback is the planned outcome in that case, and the result is labelled approximate.
- Out of scope: weapon effects (`Effect`), shield emblems, and faces and heads (a separate converter question).
- Out of scope for now: a character-creation wizard (race and gender choice, new login-page flow). Gender is stored and sent, and existing characters default to male, but players cannot yet choose a race or gender; the two fixed archetypes stay. A follow-up change should add the wizard.
