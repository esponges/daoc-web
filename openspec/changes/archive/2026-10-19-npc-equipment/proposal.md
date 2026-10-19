## Why

Every NPC in the browser looks like the same plain, ungeared character: no armour, no weapon in hand, no cloak, no helm. The community database does describe their gear (`npcequipment`, about 52,000 rows, linked from `mob` and `npctemplate` by `EquipmentTemplateID`), but `internal/importer/world.go` never reads that table, so `spawn` has no equipment and nothing downstream can send or draw it. `world-data-import` already says the importer SHALL import NPC equipment; this change makes that true and carries it through to the screen.

## What Changes

- The importer reads `npcequipment` and each mob's `EquipmentTemplateID` (falling back to the NPC template's) into the world store, and prints rows read and kept for the new table.
- The server loads an NPC's equipment with its spawn and sends it in the snapshot as an additive field on NPC entities.
- The client draws worn gear on NPC meshes: armour pieces per slot, weapons and shields on the hand sockets, cloaks, and helms, with colour applied.
- A short investigation of faces and heads decides whether they are driven by equipment data (helm slot, `Extension`) or are part of the base model, and records the answer. Drawing faces is only in scope if the answer is cheap.
- No game data is committed. The imported table stays under `server/data/`, and the lookup from item model number to art is derived from the user's install.

## Capabilities

### New Capabilities
- `npc-appearance`: how an NPC's worn gear is chosen and drawn in the client, including the fallback when a piece cannot be resolved.

### Modified Capabilities
- `world-data-import`: the importer actually stores NPC equipment and reports its counts.
- `world-sync`: NPC entities in a snapshot carry their equipment.

## Impact

- `internal/importer/world.go` (new table and mob link), `cmd/dbimport`.
- `internal/server/world/catalog.go`, `npc.go`, `zone.go`; `internal/server/protocol/protocol.go` and `docs/protocol.md` (additive field, plus a protocol test).
- `cmd/charconv` / `cmd/worldconv` (per-creature conversion must expose swappable slots), `web/remote.js`, `web/net.js`, `web/character.js`.
- Risk: the README records that the install's own item-to-armour mapping has not been found, and CLAUDE.md warns that icon numbers in the imported data do not match this install's sheets. Item model numbers may be equally unreliable, so the mapping step is research and may only partly succeed; the design plans for a graceful partial result.
- Out of scope here (player gear moved to `npc-gear-mapping`): the guild label (already done), spell effects on weapons (`Effect`), and emblems on shields.
