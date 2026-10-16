## Why

Every NPC in the world is either a target or a corpse: the server cannot tell a merchant from a guard, and a right-click does nothing but turn the camera. The imported database has 1,218 merchants with stock lists, and the character already has money and a pack, so buying and selling is the first interaction that is both visible and cheap to make honest. Doing it also lays the foundation (an NPC's kind, an `interact` request, a range check) that trainers, healers and quest givers will need.

## What Changes

- **Import NPC kind and stock:** keep `Mob.ClassType` and `Mob.ItemsListTemplateID` on each spawn, and import `merchantitem` (list id, item, page, slot) into a new table. Fail loudly if `merchantitem` is missing or empty; print read and kept counts like the other tables.
- **Server:** classify each NPC by kind from its class type (`merchant` first; everything else is `none` for now). A new `interact` request names an NPC; the server checks the NPC exists in the player's zone, is alive, is a merchant, and is within interact range, then answers with a `shop` message carrying the stock. `buy` and `sell` requests are decided by the server: price, money, pack space and range are all checked again on each request.
- **Protocol (additive):** `interact`, `buy`, `sell` requests; `shop` message (stock pages, per-item price); a `shopclose` the server sends when the player walks out of range or the NPC dies. NPC entities gain an optional `ik` (interaction kind) so the client can show a cursor hint. Documented in `docs/protocol.md`, with protocol and server tests.
- **Client:** a right-click that moves less than the existing drag threshold is an interact; a right-drag stays the camera. Too far away, the client says so in chat and does not walk the character there. A merchant window opens in the stock UI frame, with the same fallback to plain panels when `web/data/ui/` is absent.
- **Selling:** the player may sell a pack item to the open merchant for a fraction of its price; worn items are not sold directly.
- **Out of scope:** quests, trainers, healers, teleporters, vaults, stable masters, buy-back, repair, stackable quantities beyond one, a walk-to-NPC auto-move, merchant stock that sells out or restocks, and per-realm or per-faction price changes.

## Capabilities

### New Capabilities
- `npc-interaction`: how a player asks to interact with an NPC, how the server decides what kind of NPC it is and whether the player is in reach, and how the client sends and shows it.
- `merchants`: the stock a merchant offers, the rules for buying and selling, and the merchant window.

### Modified Capabilities
- `world-data-import`: the importer now keeps each spawn's class type and stock list id and imports merchant stock.
- `world-sync`: new requests and messages for interaction and trade; entities may carry an interaction kind.

## Impact

- Code: `internal/importer/world.go` (+ test), `internal/server/world/{catalog,npc,inventory,zone,world}.go`, `internal/server/protocol/protocol.go` (+ test), `internal/server/gameserver/server.go`, `web/viewer.js` (right-click), `web/net.js`, a new `web/shop.js` (+ test), `docs/protocol.md`.
- Data: `server/data/world.db` must be re-imported (`go run ./cmd/dbimport`); a `world.db` without the new tables must make the server start with interaction disabled, not crash.
- Depends on: the existing money, pack and `inv` message; the stock-UI frame helpers in `web/uiwin.js`. No game data is committed; item pictures stay drawn by kind and weapon line (no icon-offset guessing).
